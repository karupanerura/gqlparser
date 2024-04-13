package gqlparser

import (
	"github.com/karupanerura/runetrie"
)

type Kind string

type ProjectID string

type Property string

type Cursor string

type Syntax interface {
	isSyntax()
}

type Key struct {
	ProjectID ProjectID
	Namespace string
	Path      []*KeyPath
}

func (*Key) isSyntax() {}

type KeyPath struct {
	Kind Kind
	ID   int64
	Name string
}

type Query struct {
	Properties []Property
	Distinct   bool
	DistinctOn []Property
	Kind       Kind
	Where      Condition
	OrderBy    []OrderBy
	Limit      *Limit
	Offset     *Offset
}

func (*Query) isSyntax() {}

type OrderBy struct {
	Descending bool
	Property   Property
}

func (*OrderBy) isSyntax() {}

type Limit struct {
	Position int64
	Cursor   BindingVariable
}

func (*Limit) isSyntax() {}

type Offset struct {
	Position int64
	Cursor   BindingVariable
}

func (*Offset) isSyntax() {}

type AggregationQuery struct {
	Aggregations []Aggregation
	Query
}

func (*AggregationQuery) isSyntax() {}

type Aggregation interface {
	isAggregation()
}

type CountAggregation struct {
	Alias string
}

func (*CountAggregation) isAggregation() {}
func (*CountAggregation) isSyntax()      {}

type CountUpToAggregation struct {
	Limit int64
	Alias string
}

func (*CountUpToAggregation) isAggregation() {}
func (*CountUpToAggregation) isSyntax()      {}

type SumAggregation struct {
	Property string
	Alias    string
}

func (*SumAggregation) isAggregation() {}
func (*SumAggregation) isSyntax()      {}

type AvgAggregation struct {
	Property string
	Alias    string
}

func (*AvgAggregation) isAggregation() {}
func (*AvgAggregation) isSyntax()      {}

type CompoundCondition interface {
	Condition
	isCompoundCondition()
}

type AndCompoundCondition struct {
	Left  Condition
	Right Condition
}

func (*AndCompoundCondition) isCompoundCondition() {}
func (*AndCompoundCondition) isCondition()         {}
func (*AndCompoundCondition) isSyntax()            {}

func (c *AndCompoundCondition) Bind(br *BindingResolver) error {
	if err := c.Left.Bind(br); err != nil {
		return err
	}
	if err := c.Right.Bind(br); err != nil {
		return err
	}
	return nil
}

func (c *AndCompoundCondition) Normalize() Condition {
	return c
}

type OrCompoundCondition struct {
	Left  Condition
	Right Condition
}

func (*OrCompoundCondition) isCompoundCondition() {}
func (*OrCompoundCondition) isCondition()         {}
func (*OrCompoundCondition) isSyntax()            {}

func (c *OrCompoundCondition) Bind(br *BindingResolver) error {
	if err := c.Left.Bind(br); err != nil {
		return err
	}
	if err := c.Right.Bind(br); err != nil {
		return err
	}
	return nil
}

func (c *OrCompoundCondition) Normalize() Condition {
	return c
}

type Condition interface {
	isCondition()
	Bind(*BindingResolver) error
	Normalize() Condition
}

type IsNullCondition struct {
	Property string
}

func (*IsNullCondition) isCondition()                   {}
func (*IsNullCondition) isSyntax()                      {}
func (*IsNullCondition) Bind(br *BindingResolver) error { return nil }

func (c *IsNullCondition) Normalize() Condition {
	return &EitherComparatorCondition{
		Comparator: EqualsEitherComparator,
		Property:   c.Property,
		Value:      nil,
	}
}

type ForwardComparatorCondition struct {
	Comparator ForwardComparator
	Property   string
	Value      any
}

func (*ForwardComparatorCondition) isCondition() {}
func (*ForwardComparatorCondition) isSyntax()    {}

func (c *ForwardComparatorCondition) Bind(br *BindingResolver) error {
	if bv, ok := c.Value.(BindingVariable); ok {
		if v, err := br.Resolve(bv); err != nil {
			return err
		} else {
			c.Value = v
		}
	}
	return nil
}

func (c *ForwardComparatorCondition) Normalize() Condition {
	switch c.Comparator {
	case ContainsForwardComparator:
		return &EitherComparatorCondition{
			Comparator: EqualsEitherComparator,
			Property:   c.Property,
			Value:      c.Value,
		}
	default:
		return c
	}
}

type ForwardComparator string

const (
	ContainsForwardComparator    ForwardComparator = "CONTAINS"
	HasAncestorForwardComparator ForwardComparator = "HAS ANCESTOR"
	InForwardComparator          ForwardComparator = "IN"
	NotInForwardComparator       ForwardComparator = "NOT IN"
)

var forwardComparatorTrie = runetrie.NewTrie(
	ContainsForwardComparator,
	HasAncestorForwardComparator,
	InForwardComparator,
	NotInForwardComparator,
)

func (c ForwardComparator) Valid() bool {
	return forwardComparatorTrie.MatchAny(c)
}

type BackwardComparatorCondition struct {
	Comparator BackwardComparator
	Property   string
	Value      any
}

func (*BackwardComparatorCondition) isCondition() {}
func (*BackwardComparatorCondition) isSyntax()    {}

func (c *BackwardComparatorCondition) Bind(br *BindingResolver) error {
	if bv, ok := c.Value.(BindingVariable); ok {
		if v, err := br.Resolve(bv); err != nil {
			return err
		} else {
			c.Value = v
		}
	}
	return nil
}

func (c *BackwardComparatorCondition) Normalize() Condition {
	switch c.Comparator {
	case InBackwardComparator:
		return &EitherComparatorCondition{
			Comparator: EqualsEitherComparator,
			Property:   c.Property,
			Value:      c.Value,
		}
	case HasDescendantBackwardComparator:
		return &ForwardComparatorCondition{
			Comparator: HasAncestorForwardComparator,
			Property:   c.Property,
			Value:      c.Value,
		}
	default:
		return c
	}
}

type BackwardComparator string

const (
	InBackwardComparator            BackwardComparator = "IN"
	HasDescendantBackwardComparator BackwardComparator = "HAS DESCENDANT"
)

var backwardComparatorTrie = runetrie.NewTrie(
	InBackwardComparator,
	HasDescendantBackwardComparator,
)

func (c BackwardComparator) Valid() bool {
	return backwardComparatorTrie.MatchAny(c)
}

type EitherComparatorCondition struct {
	Comparator EitherComparator
	Property   string
	Value      any
}

func (*EitherComparatorCondition) isCondition() {}
func (*EitherComparatorCondition) isSyntax()    {}

func (c *EitherComparatorCondition) Bind(br *BindingResolver) error {
	if bv, ok := c.Value.(BindingVariable); ok {
		if v, err := br.Resolve(bv); err != nil {
			return err
		} else {
			c.Value = v
		}
	}
	return nil
}

func (c *EitherComparatorCondition) Normalize() Condition {
	return c
}

type EitherComparator string

const (
	EqualsEitherComparator                  EitherComparator = "="
	NotEqualsEitherComparator               EitherComparator = "!="
	GreaterThanEitherComparator             EitherComparator = ">"
	GreaterThanOrEqualsThanEitherComparator EitherComparator = ">="
	LesserThanEitherComparator              EitherComparator = "<"
	LesserThanOrEqualsEitherComparator      EitherComparator = "<="
)

var eitherComparatorTrie = runetrie.NewTrie(
	EqualsEitherComparator,
	NotEqualsEitherComparator,
	GreaterThanEitherComparator,
	GreaterThanOrEqualsThanEitherComparator,
	LesserThanEitherComparator,
	LesserThanOrEqualsEitherComparator,
)

func (c EitherComparator) Valid() bool {
	return eitherComparatorTrie.MatchAny(c)
}
