package gqlparser

import (
	"errors"
	"fmt"
	"time"
)

var prefixOperatorBindingPowerSet = map[string]struct{}{
	"-": {},
	"+": {},
	"(": {},
}

var infixCompoundOperatorBindingPowerMap = map[string]uint8{
	"AND": 2,
	"OR":  1,
}

var infixEitherOperatorBindingPowerMap = map[string]uint8{
	"=":  3,
	"!=": 3,
	"<":  3,
	"<=": 3,
	">":  3,
	">=": 3,
}

var infixEitherOperatorInvertMap = map[string]string{
	"=":  "=",
	"!=": "!=",
	"<":  ">",
	"<=": ">=",
	">":  "<",
	">=": "<=",
}

var infixBackwardOperatorBindingPowerMap = map[string]uint8{
	"HAS DESCENDANT": 3,
	"IN":             3,
}

var infixForwardOperatorBindingPowerMap = map[string]uint8{
	"CONTAINS":     3,
	"HAS ANCESTOR": 3,
	"NOT IN":       3,
	"IN":           3,
	"IS":           3,
}

var specialOpMap = map[string]map[string]string{
	"NOT": {
		"IN": "NOT IN",
	},
	"HAS": {
		"ANCESTOR":   "HAS ANCESTOR",
		"DESCENDANT": "HAS DESCENDANT",
	},
}

func constructAST(tr tokenReader, minBP uint8) (conditionAST, error) {
	tok, err := tr.Read()
	if errors.Is(err, ErrEndOfToken) {
		return nil, ErrNoTokens
	} else if err != nil {
		return nil, err
	}

	var left conditionAST
	switch v := tok.(type) {
	case *SymbolToken:
		left = &conditionField{sym: v}
	case *BooleanToken:
		left = &conditionValue{b: v}
	case *StringToken:
		if v.Quote == '`' {
			left = &conditionField{str: v}
		} else {
			left = &conditionValue{s: v}
		}
	case *NumericToken:
		left = &conditionValue{n: v}
	case *BindingToken:
		left = &conditionValue{bind: v}
	case *OperatorToken:
		left, err = parsePrefixOP(tr, v)
		if err != nil {
			return nil, err
		}
	case *KeywordToken:
		switch v.Name {
		case "KEY":
			var key Key
			if err := acceptKeyBody(&key).accept(tr); err != nil {
				return nil, err
			}
			left = &conditionKey{keyKeyword: v, key: &key}
		case "ARRAY":
			var values []conditionValuer
			if err := acceptArrayBody(&values).accept(tr); err != nil {
				return nil, err
			}
			left = &conditionArray{arrayKeyword: v, values: values}
		case "BLOB":
			var b []byte
			if err := acceptBlobBody(&b).accept(tr); err != nil {
				return nil, err
			}
			left = &conditionBlob{blobKeyword: v, b: b}
		case "DATETIME":
			var t time.Time
			if err := acceptDateTimeBody(&t).accept(tr); err != nil {
				return nil, err
			}
			left = &conditionDateTime{dateTimeKeyword: v, t: t}
		case "NULL":
			left = &conditionValue{null: v}
		default:
			return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
		}
	default:
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
	}

	rtr := asResettableTokenReader(tr)
	for {
		if err := skipWhitespaceToken.accept(rtr); err != nil {
			return nil, err
		}

		tok, err := rtr.Read()
		if errors.Is(err, ErrEndOfToken) {
			return left, nil
		} else if err != nil {
			return nil, err
		}

		op, isOP := tok.(*OperatorToken)
		if !isOP {
			if _, isKeyword := tok.(*KeywordToken); isKeyword {
				rtr.Reset()
				return left, nil
			}
			return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
		}

		typ := op.Type
		if m, ok := specialOpMap[typ]; ok {
			if err := acceptWhitespaceToken.accept(rtr); err != nil {
				return nil, err
			}

			nextToken, err := rtr.Read()
			if errors.Is(err, ErrEndOfToken) {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
			} else if err != nil {
				return nil, err
			}

			nextOP, isOP := nextToken.(*OperatorToken)
			if !isOP {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, nextToken.GetContent(), nextToken.GetPosition())
			}

			typ = m[nextOP.Type]
			if typ == "" {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, nextToken.GetContent(), nextToken.GetPosition())
			}
		}

		if err := skipWhitespaceToken.accept(rtr); err != nil {
			return nil, err
		}

		allowBackwardOP := false
		allowForwardOP := false
		allowEitherOP := false
		allowCompoundOP := false
		switch left.(type) {
		case conditionValuer:
			allowBackwardOP = true
			allowEitherOP = true
		case *conditionField:
			allowForwardOP = true
			allowEitherOP = true
		default:
			allowCompoundOP = true
		}

		var bp uint8
		var isEitherOP bool
		if allowEitherOP {
			bp = infixEitherOperatorBindingPowerMap[typ]
			isEitherOP = true
		}
		if bp == 0 {
			if allowCompoundOP {
				bp = infixCompoundOperatorBindingPowerMap[typ]
			} else if allowForwardOP {
				bp = infixForwardOperatorBindingPowerMap[typ]
			} else if allowBackwardOP {
				bp = infixBackwardOperatorBindingPowerMap[typ]
			} else {
				panic("broken pattern")
			}
		}
		if bp == 0 || bp < minBP {
			rtr.Reset()
			return left, nil
		}

		right, err := constructAST(tr, bp+1)
		if errors.Is(err, ErrEndOfToken) {
			// ok: ignore it
		} else if err != nil {
			return nil, err
		}
		if right == nil {
			return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
		}

		if isEitherOP {
			if cv, isValue := left.(conditionValuer); isValue {
				fv, isField := right.(*conditionField)
				if !isField {
					return nil, right.toUnexpectedTokenError()
				}
				left = &backwardComparatorCondition{left: cv, op: op, opType: typ, right: fv}
			} else if fv, isField := left.(*conditionField); isField {
				cv, isValue := right.(conditionValuer)
				if !isValue {
					return nil, right.toUnexpectedTokenError()
				}
				left = &forwardComparatorCondition{left: fv, op: op, opType: typ, right: cv}
			} else {
				return nil, left.toUnexpectedTokenError()
			}
		} else if allowCompoundOP {
			left = &compoundComparatorCondition{left: left, op: op, right: right}
		} else if allowForwardOP {
			fv, isField := left.(*conditionField)
			if !isField {
				return nil, left.toUnexpectedTokenError()
			}

			cv, isValue := right.(conditionValuer)
			if !isValue {
				return nil, right.toUnexpectedTokenError()
			}
			left = &forwardComparatorCondition{left: fv, op: op, opType: typ, right: cv}
		} else if allowBackwardOP {
			cv, isValue := left.(conditionValuer)
			if !isValue {
				return nil, left.toUnexpectedTokenError()
			}

			fv, isField := right.(*conditionField)
			if !isField {
				return nil, right.toUnexpectedTokenError()
			}
			left = &backwardComparatorCondition{left: cv, op: op, opType: typ, right: fv}
		} else {
			panic("broken pattern")
		}

		rtr = asResettableTokenReader(tr) // new offset
		if err := skipWhitespaceToken.accept(rtr); err != nil {
			return nil, err
		}
	}
}

func parsePrefixOP(tr tokenReader, op *OperatorToken) (conditionAST, error) {
	if _, isPrefixOP := prefixOperatorBindingPowerSet[op.Type]; isPrefixOP {
		if op.Type == "(" {
			if err := skipWhitespaceToken.accept(tr); err != nil {
				return nil, err
			}

			children, err := constructAST(tr, 0)
			if errors.Is(err, ErrEndOfToken) {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, op.GetContent(), op.GetPosition())
			} else if err != nil {
				return nil, err
			}

			if err := skipWhitespaceToken.accept(tr); err != nil {
				return nil, err
			}

			nextToken, err := tr.Read()
			if errors.Is(err, ErrEndOfToken) {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, op.GetContent(), op.GetPosition())
			} else if err != nil {
				return nil, err
			}

			if t, isOp := nextToken.(*OperatorToken); !isOp {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, op.GetContent(), op.GetPosition())
			} else if t.Type != ")" {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, op.GetContent(), op.GetPosition())
			}

			return children, nil
		} else {
			nextToken, err := tr.Read()
			if errors.Is(err, ErrEndOfToken) {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, op.GetContent(), op.GetPosition())
			} else if err != nil {
				return nil, err
			}

			numTok, ok := nextToken.(*NumericToken)
			if !ok {
				return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, nextToken.GetContent(), nextToken.GetPosition())
			}

			return &prefixCondition{op: op, right: &conditionValue{n: numTok}}, nil
		}
	} else {
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, op.GetContent(), op.GetPosition())
	}
}

func acceptConditionValue(result *conditionValuer) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		tok, err := tr.Read()
		if errors.Is(err, ErrEndOfToken) {
			return ErrNoTokens
		} else if err != nil {
			return err
		}

		switch v := tok.(type) {
		case *BooleanToken:
			*result = &conditionValue{b: v}
			return nil
		case *StringToken:
			if v.Quote == '`' {
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
			} else {
				*result = &conditionValue{s: v}
				return nil
			}
		case *NumericToken:
			*result = &conditionValue{n: v}
			return nil
		case *BindingToken:
			*result = &conditionValue{bind: v}
			return nil
		case *KeywordToken:
			switch v.Name {
			case "KEY":
				var key Key
				if err := acceptKeyBody(&key).accept(tr); err != nil {
					return err
				}
				*result = &conditionKey{keyKeyword: v, key: &key}
				return nil
			case "ARRAY":
				var values []conditionValuer
				if err := acceptArrayBody(&values).accept(tr); err != nil {
					return err
				}
				*result = &conditionArray{arrayKeyword: v, values: values}
				return nil
			case "BLOB":
				var b []byte
				if err := acceptBlobBody(&b).accept(tr); err != nil {
					return err
				}
				*result = &conditionBlob{blobKeyword: v, b: b}
				return nil
			case "DATETIME":
				var t time.Time
				if err := acceptDateTimeBody(&t).accept(tr); err != nil {
					return err
				}
				*result = &conditionDateTime{dateTimeKeyword: v, t: t}
				return nil
			case "NULL":
				*result = &conditionValue{null: v}
				return nil
			default:
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
			}
		default:
			return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
		}
	})
}

type conditionAST interface {
	toCondition() (Condition, error)
	toUnexpectedTokenError() error
}

type conditionValuer interface {
	value() any
	toUnexpectedTokenError() error
}

type prefixCondition struct {
	op    *OperatorToken
	right *conditionValue
}

func (c *prefixCondition) value() any {
	switch c.op.Type {
	case "+":
		if c.right.n == nil {
			panic("invalid prefix condition")
		}
		return c.right.value()
	case "-":
		if c.right.n == nil {
			panic("invalid prefix condition")
		}
		switch v := c.right.value().(type) {
		case int64:
			return -v
		case float64:
			return -v
		default:
			panic(fmt.Sprintf("unknown right value type: %T", c.right.value()))
		}
	default:
		panic("unknown operator: " + c.op.Type)
	}
}

func (c *prefixCondition) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *prefixCondition) toUnexpectedTokenError() error {
	return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.op.GetContent(), c.op.GetPosition())
}

type forwardComparatorCondition struct {
	left   *conditionField
	op     *OperatorToken
	opType string
	right  conditionValuer
}

func (c *forwardComparatorCondition) toCondition() (Condition, error) {
	if _, isEitherOP := infixEitherOperatorInvertMap[c.opType]; isEitherOP {
		// not invert op to canonical
		comparator := EitherComparator(c.opType)
		if !comparator.Valid() {
			return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.op.GetContent(), c.op.GetPosition())
		}
		return &EitherComparatorCondition{Comparator: comparator, Property: c.left.name(), Value: c.right.value()}, nil
	}
	if c.opType == "IS" {
		if c.right.value() != nil {
			return nil, c.right.toUnexpectedTokenError()
		}
		return &IsNullCondition{Property: c.left.name()}, nil
	}

	comparator := ForwardComparator(c.opType)
	if !comparator.Valid() {
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.op.GetContent(), c.op.GetPosition())
	}
	return &ForwardComparatorCondition{Comparator: comparator, Property: c.left.name(), Value: c.right.value()}, nil
}

func (c *forwardComparatorCondition) toUnexpectedTokenError() error {
	return c.left.toUnexpectedTokenError()
}

type backwardComparatorCondition struct {
	left   conditionValuer
	op     *OperatorToken
	opType string
	right  *conditionField
}

func (c *backwardComparatorCondition) toCondition() (Condition, error) {
	if op, isEitherOP := infixEitherOperatorInvertMap[c.opType]; isEitherOP {
		// invert op to canonical
		comparator := EitherComparator(op)
		if !comparator.Valid() {
			return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.op.GetContent(), c.op.GetPosition())
		}
		return &EitherComparatorCondition{Comparator: comparator, Property: c.right.name(), Value: c.left.value()}, nil
	}

	comparator := BackwardComparator(c.opType)
	if !comparator.Valid() {
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.op.GetContent(), c.op.GetPosition())
	}
	return &BackwardComparatorCondition{Comparator: comparator, Property: c.right.name(), Value: c.left.value()}, nil
}

func (c *backwardComparatorCondition) toUnexpectedTokenError() error {
	return c.left.toUnexpectedTokenError()
}

type compoundComparatorCondition struct {
	left  conditionAST
	op    *OperatorToken
	right conditionAST
}

func (c *compoundComparatorCondition) toCondition() (Condition, error) {
	left, err := c.left.toCondition()
	if err != nil {
		return nil, err
	}

	right, err := c.right.toCondition()
	if err != nil {
		return nil, err
	}

	switch c.op.Type {
	case "AND":
		return &AndCompoundCondition{Left: left, Right: right}, nil
	case "OR":
		return &OrCompoundCondition{Left: left, Right: right}, nil
	default:
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.op.GetContent(), c.op.GetPosition())
	}
}

func (c *compoundComparatorCondition) toUnexpectedTokenError() error {
	return c.left.toUnexpectedTokenError()
}

type conditionField struct {
	sym *SymbolToken
	str *StringToken
}

func (c *conditionField) name() string {
	if c.sym != nil {
		return c.sym.Content
	}
	return c.str.Content
}

func (c *conditionField) token() Token {
	if c.sym != nil {
		return c.sym
	}
	return c.str
}

func (c *conditionField) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *conditionField) toUnexpectedTokenError() error {
	return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.token().GetContent(), c.token().GetPosition())
}

type conditionValue struct {
	b    *BooleanToken
	s    *StringToken
	n    *NumericToken
	null *KeywordToken
	bind *BindingToken
}

func (c *conditionValue) token() Token {
	if c.b != nil {
		return c.b
	}
	if c.s != nil {
		return c.s
	}
	if c.n != nil {
		return c.n
	}
	if c.null != nil {
		return c.null
	}
	if c.bind != nil {
		return c.bind
	}
	panic("every token is nil")
}

func (c *conditionValue) value() any {
	if c.b != nil {
		return c.b.Value
	}
	if c.s != nil {
		return c.s.Content
	}
	if c.n != nil {
		if c.n.Floating {
			return c.n.Float64
		}
		return c.n.Int64
	}
	if c.null != nil {
		return nil
	}
	if c.bind != nil {
		return parseBindingToken(c.bind)
	}
	panic("every token is nil")
}

func (c *conditionValue) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *conditionValue) toUnexpectedTokenError() error {
	return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.token().GetContent(), c.token().GetPosition())
}

type conditionKey struct {
	keyKeyword *KeywordToken
	key        *Key
}

func (c *conditionKey) value() any {
	return c.key
}

func (c *conditionKey) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *conditionKey) toUnexpectedTokenError() error {
	return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.keyKeyword.GetContent(), c.keyKeyword.GetPosition())
}

type conditionArray struct {
	arrayKeyword *KeywordToken
	values       []conditionValuer
}

func (c *conditionArray) value() any {
	values := make([]any, len(c.values))
	for i, v := range c.values {
		values[i] = v.value()
	}
	return values
}

func (c *conditionArray) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *conditionArray) toUnexpectedTokenError() error {
	return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.arrayKeyword.GetContent(), c.arrayKeyword.GetPosition())
}

type conditionBlob struct {
	blobKeyword *KeywordToken
	b           []byte
}

func (c *conditionBlob) value() any {
	return c.b
}

func (c *conditionBlob) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *conditionBlob) toUnexpectedTokenError() error {
	return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.blobKeyword.GetContent(), c.blobKeyword.GetPosition())
}

type conditionDateTime struct {
	dateTimeKeyword *KeywordToken
	t               time.Time
}

func (c *conditionDateTime) value() any {
	return c.t
}

func (c *conditionDateTime) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *conditionDateTime) toUnexpectedTokenError() error {
	return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.dateTimeKeyword.GetContent(), c.dateTimeKeyword.GetPosition())
}
