package gqlparser

import (
	"github.com/karupanerura/runetrie"
)

// Kind represents the kind of a datastore entity in the GQL syntax.
type Kind string

// ProjectID represents the Google Cloud project ID in the GQL syntax.
type ProjectID string

// Property represents a property name of a datastore entity in the GQL syntax.
type Property struct {
	// Name is the name of the property.
	Name string

	// Child is an optional child property.
	// If set, it represents a nested property access.
	// For example, if Name is "foo" and Child is "bar", it represents "foo.bar".
	Child *Property
}

// String returns the string representation of the Property.
func (p *Property) String() string {
	if p.Child != nil {
		return p.Name + "." + p.Child.String()
	}
	return p.Name
}

// Cursor represents a datastore pagination cursor in the GQL syntax.
type Cursor string

// Syntax is an interface for a GQL syntax AST entity.
type Syntax interface {
	isSyntax()
}

// Key represents a datastore key in the syntax.
type Key struct {
	// ProjectID is the Google Cloud project ID.
	ProjectID ProjectID

	// Namespace is the namespace of the datastore.
	// It can be empty if the namespace is not specified.
	Namespace string

	// Path is a slice of KeyPath elements that represent the path to the entity.
	Path []*KeyPath
}

func (*Key) isSyntax() {}

// KeyPath represents a path element in a datastore key.
type KeyPath struct {
	// Kind is the kind of the entity.
	Kind Kind

	// ID is the numeric ID of the entity.
	ID int64

	// Name is the string name of the entity.
	Name string
}

// Query represents a GQL query syntax.
type Query struct {
	// Properties is a slice of properties to be selected in the projection query.
	Properties []Property

	// Distinct is a flag indicating whether to return distinct results.
	Distinct bool

	// DistinctOn is a slice of properties to be used for distinct results.
	DistinctOn []Property

	// Kind is the kind of the datastore entity to be queried.
	Kind Kind

	// Where is the condition to filter the results.
	Where Condition

	// OrderBy is a slice of order by clauses to sort the results.
	// Each OrderBy clause can specify ascending or descending order.
	OrderBy []OrderBy

	// Limit is the limit on the number of results to be returned.
	Limit *Limit

	// Offset is the number of results to skip before starting to collect the result set.
	Offset *Offset
}

func (*Query) isSyntax() {}

// OrderBy represents an order by clause in the query.
type OrderBy struct {
	// Descending indicates whether the order is descending.
	// If true, the order is descending. If false, the order is ascending.
	Descending bool

	// Property is the property to order by.
	Property Property
}

func (*OrderBy) isSyntax() {}

// Limit represents a limit clause in the query.
type Limit struct {
	// Position is the maximum number of results to return.
	// It is a positive integer.
	Position int64

	// Cursor is a cursor to start the results from.
	Cursor BindingVariable
}

func (*Limit) isSyntax() {}

// Offset represents an offset clause in the query.
type Offset struct {
	// Position is the number of results to skip before starting to collect the result set.
	// It is a positive integer.
	Position int64

	// Cursor is a cursor to start the results from.
	Cursor BindingVariable
}

func (*Offset) isSyntax() {}

// AggregationQuery represents a GQL aggregation query syntax.
// It extends the base Query syntax with aggregation capabilities.
type AggregationQuery struct {
	// Aggregations is a slice of aggregation clauses to be applied to the query.
	Aggregations []Aggregation

	// Query is the base query to be aggregated.
	Query
}

func (*AggregationQuery) isSyntax() {}

// Aggregation is an interface for different types of aggregation specifications in the GQL syntax.
type Aggregation interface {
	isAggregation()
}

// CountAggregation represents a count aggregation in the GQL syntax.
type CountAggregation struct {
	Alias string
}

func (*CountAggregation) isAggregation() {}
func (*CountAggregation) isSyntax()      {}

// CountUpToAggregation represents a count up to a specified limit in the GQL syntax.
type CountUpToAggregation struct {
	Limit int64
	Alias string
}

func (*CountUpToAggregation) isAggregation() {}
func (*CountUpToAggregation) isSyntax()      {}

// SumAggregation represents a sum aggregation in the GQL syntax.
type SumAggregation struct {
	Property Property
	Alias    string
}

func (*SumAggregation) isAggregation() {}
func (*SumAggregation) isSyntax()      {}

// AvgAggregation represents an average aggregation in the GQL syntax.
type AvgAggregation struct {
	Property Property
	Alias    string
}

func (*AvgAggregation) isAggregation() {}
func (*AvgAggregation) isSyntax()      {}

// CompoundCondition is an interface for compound conditions in the GQL syntax.
type CompoundCondition interface {
	Condition
	isCompoundCondition()
}

// AndCompoundCondition represents a compound condition that combines two conditions with an AND operator.
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
	return &AndCompoundCondition{
		Left:  c.Left.Normalize(),
		Right: c.Right.Normalize(),
	}
}

// OrCompoundCondition represents a compound condition that combines two conditions with an OR operator.
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
	return &OrCompoundCondition{
		Left:  c.Left.Normalize(),
		Right: c.Right.Normalize(),
	}
}

// Condition is an interface for WHERE conditions in the GQL syntax.
type Condition interface {
	isCondition()
	Bind(*BindingResolver) error
	Normalize() Condition
}

// IsNullCondition represents a condition that checks if a property is null.
type IsNullCondition struct {
	Property Property
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

// ForwardComparatorCondition represents a condition that uses a forward comparator.
type ForwardComparatorCondition struct {
	// Comparator is the forward comparator to be used in the condition.
	Comparator ForwardComparator

	// Property is the property to be compared.
	Property Property

	// Value is the value to be compared against.
	Value any
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

// ForwardComparator is an enum-like string type that represents forward comparison operators in the GQL syntax.
// These operators are used with properties on the left side of conditions.
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

// BackwardComparatorCondition represents a condition that uses a backward comparator.
type BackwardComparatorCondition struct {
	// Comparator is the backward comparator to be used in the condition.
	Comparator BackwardComparator

	// Property is the property to be compared.
	Property Property

	// Value is the value to be compared against.
	Value any
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

// BackwardComparator is an enum-like string type that represents backward comparison operators in the GQL syntax.
// These operators are used with properties on the right side of conditions.
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

// EitherComparatorCondition represents a condition that uses either a forward or backward comparator.
type EitherComparatorCondition struct {
	// Comparator is the either comparator to be used in the condition.
	Comparator EitherComparator

	// Property is the property to be compared.
	Property Property

	// Value is the value to be compared against.
	Value any
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

// EitherComparator is an enum-like string type that represents either forward or backward comparison operators in the GQL syntax.
// These operators can be used with properties on either side of conditions.
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
