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
		allowCompoundOP := false
		switch left.(type) {
		case conditionValuer:
			allowBackwardOP = true
		case *conditionField:
			allowForwardOP = true
		default:
			allowCompoundOP = true
		}

		var bp uint8
		var isEitherOP bool
		if allowBackwardOP || allowForwardOP {
			bp, isEitherOP = infixEitherOperatorBindingPowerMap[typ]
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
		case *OperatorToken:
			if v.Type != "+" && v.Type != "-" {
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
			}
			if err := acceptConditionValue(result).accept(tr); err != nil {
				// NOTE: ignore following errors because the prefix operator is a first unexpected token if error is occurred
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
			}
			if cv, ok := (*result).(*conditionValue); ok && cv.isNumericValue() {
				pc := &prefixCondition{op: v, right: cv}
				*result = pc
			} else {
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, tok.GetContent(), tok.GetPosition())
			}
			return nil
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
