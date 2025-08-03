package gqlparser

import (
	"fmt"
	"time"
)

type conditionAST interface {
	toCondition() (Condition, error)
	toUnexpectedTokenError() error
}

type valueAST interface {
	value() any
	toUnexpectedTokenError() error
}

type forwardComparatorCondition struct {
	left   propertyAST
	op     *OperatorToken
	opType string
	right  valueAST
}

func (c *forwardComparatorCondition) toCondition() (Condition, error) {
	if _, isEitherOP := infixEitherOperatorInvertMap[c.opType]; isEitherOP {
		// not invert op to canonical
		comparator := EitherComparator(c.opType)
		return &EitherComparatorCondition{Comparator: comparator, Property: c.left.toProperty(), Value: c.right.value()}, nil
	}
	if c.opType == "IS" {
		if c.right.value() != nil {
			return nil, c.right.toUnexpectedTokenError()
		}
		return &IsNullCondition{Property: c.left.toProperty()}, nil
	}

	comparator := ForwardComparator(c.opType)
	if !comparator.Valid() {
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.op.GetContent(), c.op.GetPosition())
	}
	return &ForwardComparatorCondition{Comparator: comparator, Property: c.left.toProperty(), Value: c.right.value()}, nil
}

func (c *forwardComparatorCondition) toUnexpectedTokenError() error {
	return c.left.toUnexpectedTokenError()
}

type backwardComparatorCondition struct {
	left   valueAST
	op     *OperatorToken
	opType string
	right  propertyAST
}

func (c *backwardComparatorCondition) toCondition() (Condition, error) {
	if op, isEitherOP := infixEitherOperatorInvertMap[c.opType]; isEitherOP {
		// invert op to canonical
		comparator := EitherComparator(op)
		return &EitherComparatorCondition{Comparator: comparator, Property: c.right.toProperty(), Value: c.left.value()}, nil
	}

	comparator := BackwardComparator(c.opType)
	if !comparator.Valid() {
		return nil, fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.op.GetContent(), c.op.GetPosition())
	}
	return &BackwardComparatorCondition{Comparator: comparator, Property: c.right.toProperty(), Value: c.left.value()}, nil
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

type propertyAST interface {
	conditionAST
	toProperty() Property
}

type conditionField struct {
	sym *SymbolToken
	str *StringToken
}

var _ propertyAST = (*conditionField)(nil)

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

func (c *conditionField) toProperty() Property {
	return Property{Name: c.name()}
}

func (c *conditionField) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *conditionField) toUnexpectedTokenError() error {
	return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, c.token().GetContent(), c.token().GetPosition())
}

type conditionFieldAccess struct {
	left  propertyAST
	right propertyAST
}

var _ propertyAST = (*conditionFieldAccess)(nil)

func (c *conditionFieldAccess) toProperty() Property {
	left := c.left.toProperty()
	right := c.right.toProperty()

	last := &left
	for last.Child != nil {
		last = last.Child
	}
	last.Child = &right
	return left
}

func (c *conditionFieldAccess) toCondition() (Condition, error) {
	return nil, c.toUnexpectedTokenError()
}

func (c *conditionFieldAccess) toUnexpectedTokenError() error {
	return c.left.toUnexpectedTokenError()
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
	values       []valueAST
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
