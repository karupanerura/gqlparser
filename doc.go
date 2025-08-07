/*
Package gqlparser provides a parser for Google Cloud Datastore's GQL (Google Query Language).

GQL is a SQL-like query language for Firestore in Datastore mode.
This package provides lexical analysis, parsing, and AST generation.

# Features

- SELECT queries with projection, DISTINCT, WHERE, ORDER BY, LIMIT, OFFSET
- Aggregation queries (COUNT, COUNT_UP_TO, SUM, AVG)
- Complex conditions with AND/OR and comparison operators
- Special operators: IN, CONTAINS, HAS ANCESTOR, HAS DESCENDANT
- Property path traversal ("parent.child")
- Binding variables (@1, @name) for parameterized queries

# AST Types

- Query, AggregationQuery: Query representations
- Condition types: EitherComparatorCondition, ForwardComparatorCondition, etc.
- Key, Property, BindingVariable: Data types

See: https://cloud.google.com/datastore/docs/reference/gql_reference
*/
package gqlparser
