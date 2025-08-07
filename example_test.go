package gqlparser_test

import (
	"errors"
	"fmt"
	"log"

	"github.com/karupanerura/gqlparser"
)

func ExampleParseQuery() {
	source := "SELECT * FROM Kind"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Properties: %d\n", len(query.Properties))
	// Output:
	// Kind: Kind
	// Properties: 0
}

func ExampleParseQuery_properties() {
	source := "SELECT name, age, email FROM User"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Properties:\n")
	for i, prop := range query.Properties {
		fmt.Printf("  %d: %s\n", i+1, prop.Name)
	}
	// Output:
	// Kind: User
	// Properties:
	//   1: name
	//   2: age
	//   3: email
}

func ExampleParseQuery_where() {
	source := "SELECT * FROM User WHERE age >= 18"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Has WHERE type: %T\n", query.Where)
	fmt.Printf("Has WHERE property: %s\n", query.Where.(*gqlparser.EitherComparatorCondition).Property.Name)
	fmt.Printf("Has WHERE comparator: %s\n", query.Where.(*gqlparser.EitherComparatorCondition).Comparator)
	fmt.Printf("Has WHERE value type: %T\n", query.Where.(*gqlparser.EitherComparatorCondition).Value)
	fmt.Printf("Has WHERE value: %v\n", query.Where.(*gqlparser.EitherComparatorCondition).Value)
	// Output:
	// Kind: User
	// Has WHERE type: *gqlparser.EitherComparatorCondition
	// Has WHERE property: age
	// Has WHERE comparator: >=
	// Has WHERE value type: int64
	// Has WHERE value: 18
}

func ExampleParseQuery_order_by() {
	source := "SELECT * FROM User ORDER BY name ASC, age DESC"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Order By clauses: %d\n", len(query.OrderBy))
	for i, orderBy := range query.OrderBy {
		direction := "ASC"
		if orderBy.Descending {
			direction = "DESC"
		}
		fmt.Printf("  %d: %s %s\n", i+1, orderBy.Property.Name, direction)
	}
	// Output:
	// Kind: User
	// Order By clauses: 2
	//   1: name ASC
	//   2: age DESC
}

func ExampleParseQuery_limit() {
	source := "SELECT * FROM User LIMIT 10"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Has Limit: %t\n", query.Limit != nil)
	if query.Limit != nil {
		fmt.Printf("Limit Position: %d\n", query.Limit.Position)
	}
	// Output:
	// Kind: User
	// Has Limit: true
	// Limit Position: 10
}

func ExampleParseQuery_distinct() {
	source := "SELECT DISTINCT name FROM User"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Distinct: %t\n", query.Distinct)
	fmt.Printf("Properties: %d\n", len(query.Properties))
	// Output:
	// Kind: User
	// Distinct: true
	// Properties: 1
}

func ExampleParseQuery_bindings() {
	source := "SELECT * FROM User WHERE age > @1 AND name = @username"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	err = query.Where.Bind(&gqlparser.BindingResolver{
		Indexed: []any{18},
		Named:   map[string]any{"username": "John"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Has WHERE type: %T\n", query.Where)
	fmt.Printf("Has WHERE AND left type: %T\n", query.Where.(*gqlparser.AndCompoundCondition).Left)
	fmt.Printf("Has WHERE AND left property: %s\n", query.Where.(*gqlparser.AndCompoundCondition).Left.(*gqlparser.EitherComparatorCondition).Property.Name)
	fmt.Printf("Has WHERE AND left comparator: %s\n", query.Where.(*gqlparser.AndCompoundCondition).Left.(*gqlparser.EitherComparatorCondition).Comparator)
	fmt.Printf("Has WHERE AND left value type: %T\n", query.Where.(*gqlparser.AndCompoundCondition).Left.(*gqlparser.EitherComparatorCondition).Value)
	fmt.Printf("Has WHERE AND left value: %v\n", query.Where.(*gqlparser.AndCompoundCondition).Left.(*gqlparser.EitherComparatorCondition).Value)
	fmt.Printf("Has WHERE AND right type: %T\n", query.Where.(*gqlparser.AndCompoundCondition).Right)
	fmt.Printf("Has WHERE AND right property: %s\n", query.Where.(*gqlparser.AndCompoundCondition).Right.(*gqlparser.EitherComparatorCondition).Property.Name)
	fmt.Printf("Has WHERE AND right comparator: %s\n", query.Where.(*gqlparser.AndCompoundCondition).Right.(*gqlparser.EitherComparatorCondition).Comparator)
	fmt.Printf("Has WHERE AND right value type: %T\n", query.Where.(*gqlparser.AndCompoundCondition).Right.(*gqlparser.EitherComparatorCondition).Value)
	fmt.Printf("Has WHERE AND right value: %v\n", query.Where.(*gqlparser.AndCompoundCondition).Right.(*gqlparser.EitherComparatorCondition).Value)
	// Output:
	// Kind: User
	// Has WHERE type: *gqlparser.AndCompoundCondition
	// Has WHERE AND left type: *gqlparser.EitherComparatorCondition
	// Has WHERE AND left property: age
	// Has WHERE AND left comparator: >
	// Has WHERE AND left value type: int
	// Has WHERE AND left value: 18
	// Has WHERE AND right type: *gqlparser.EitherComparatorCondition
	// Has WHERE AND right property: name
	// Has WHERE AND right comparator: =
	// Has WHERE AND right value type: string
	// Has WHERE AND right value: John
}

func ExampleParseQueryOrAggregationQuery() {
	source := "SELECT COUNT(*) FROM User"
	lexer := gqlparser.NewLexer(source)

	_, aggregationQuery, err := gqlparser.ParseQueryOrAggregationQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", aggregationQuery.Kind)
	fmt.Printf("Aggregations: %d\n", len(aggregationQuery.Aggregations))
	for i, agg := range aggregationQuery.Aggregations {
		fmt.Printf("  %d: %T\n", i+1, agg)
	}
	// Output:
	// Kind: User
	// Aggregations: 1
	//   1: *gqlparser.CountAggregation
}

func ExampleParseQueryOrAggregationQuery_multiple_functions() {
	source := "SELECT COUNT(*), SUM(age), AVG(score) FROM User"
	lexer := gqlparser.NewLexer(source)

	_, aggregationQuery, err := gqlparser.ParseQueryOrAggregationQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", aggregationQuery.Kind)
	fmt.Printf("Aggregations: %d\n", len(aggregationQuery.Aggregations))
	for i, agg := range aggregationQuery.Aggregations {
		fmt.Printf("  %d: %T\n", i+1, agg)
	}
	// Output:
	// Kind: User
	// Aggregations: 3
	//   1: *gqlparser.CountAggregation
	//   2: *gqlparser.SumAggregation
	//   3: *gqlparser.AvgAggregation
}

func ExampleParseQueryOrAggregationQuery_alias() {
	source := "SELECT COUNT(*) AS total, AVG(age) AS average_age FROM User"
	lexer := gqlparser.NewLexer(source)

	_, aggregationQuery, err := gqlparser.ParseQueryOrAggregationQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", aggregationQuery.Kind)
	fmt.Printf("Aggregations: %d\n", len(aggregationQuery.Aggregations))
	for i, agg := range aggregationQuery.Aggregations {
		fmt.Printf("  %d: %T\n", i+1, agg)
	}
	// Output:
	// Kind: User
	// Aggregations: 2
	//   1: *gqlparser.CountAggregation
	//   2: *gqlparser.AvgAggregation
}

func ExampleParseCondition() {
	source := "name = 'John' AND age > 25"
	lexer := gqlparser.NewLexer(source)

	condition, err := gqlparser.ParseCondition(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("condition type: %T\n", condition)
	fmt.Printf("Has condition AND left type: %T\n", condition.(*gqlparser.AndCompoundCondition).Left)
	fmt.Printf("Has condition AND left property: %s\n", condition.(*gqlparser.AndCompoundCondition).Left.(*gqlparser.EitherComparatorCondition).Property.Name)
	fmt.Printf("Has condition AND left comparator: %s\n", condition.(*gqlparser.AndCompoundCondition).Left.(*gqlparser.EitherComparatorCondition).Comparator)
	fmt.Printf("Has condition AND left value type: %T\n", condition.(*gqlparser.AndCompoundCondition).Left.(*gqlparser.EitherComparatorCondition).Value)
	fmt.Printf("Has condition AND left value: %v\n", condition.(*gqlparser.AndCompoundCondition).Left.(*gqlparser.EitherComparatorCondition).Value)
	fmt.Printf("Has condition AND right type: %T\n", condition.(*gqlparser.AndCompoundCondition).Right)
	fmt.Printf("Has condition AND right property: %s\n", condition.(*gqlparser.AndCompoundCondition).Right.(*gqlparser.EitherComparatorCondition).Property.Name)
	fmt.Printf("Has condition AND right comparator: %s\n", condition.(*gqlparser.AndCompoundCondition).Right.(*gqlparser.EitherComparatorCondition).Comparator)
	fmt.Printf("Has condition AND right value type: %T\n", condition.(*gqlparser.AndCompoundCondition).Right.(*gqlparser.EitherComparatorCondition).Value)
	fmt.Printf("Has condition AND right value: %v\n", condition.(*gqlparser.AndCompoundCondition).Right.(*gqlparser.EitherComparatorCondition).Value)
	// Output:
	// condition type: *gqlparser.AndCompoundCondition
	// Has condition AND left type: *gqlparser.EitherComparatorCondition
	// Has condition AND left property: name
	// Has condition AND left comparator: =
	// Has condition AND left value type: string
	// Has condition AND left value: John
	// Has condition AND right type: *gqlparser.EitherComparatorCondition
	// Has condition AND right property: age
	// Has condition AND right comparator: >
	// Has condition AND right value type: int64
	// Has condition AND right value: 25
}

func ExampleParseKey() {
	source := "KEY(Kind, 'key_name')"
	lexer := gqlparser.NewLexer(source)

	key, err := gqlparser.ParseKey(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Key paths: %d\n", len(key.Path))
	if len(key.Path) > 0 {
		fmt.Printf("Kind: %s\n", key.Path[0].Kind)
		fmt.Printf("Name: %s\n", key.Path[0].Name)
	}
	// Output:
	// Key paths: 1
	// Kind: Kind
	// Name: key_name
}

func ExampleParseKey_project() {
	source := "KEY(PROJECT('my-project'), Kind, 123)"
	lexer := gqlparser.NewLexer(source)

	key, err := gqlparser.ParseKey(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Project ID: %s\n", key.ProjectID)
	fmt.Printf("Key paths: %d\n", len(key.Path))
	if len(key.Path) > 0 {
		fmt.Printf("Kind: %s\n", key.Path[0].Kind)
		fmt.Printf("ID: %d\n", key.Path[0].ID)
	}
	// Output:
	// Project ID: my-project
	// Key paths: 1
	// Kind: Kind
	// ID: 123
}

func ExampleParseQuery_complex() {
	source := "SELECT DISTINCT name, age FROM User WHERE age >= 18 AND status = 'active' ORDER BY name ASC LIMIT 20 OFFSET 10"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Distinct: %t\n", query.Distinct)
	fmt.Printf("Properties: %d\n", len(query.Properties))
	fmt.Printf("Has WHERE: %t\n", query.Where != nil)
	fmt.Printf("Order By clauses: %d\n", len(query.OrderBy))
	fmt.Printf("Has Limit: %t\n", query.Limit != nil)
	fmt.Printf("Has Offset: %t\n", query.Offset != nil)
	// Output:
	// Kind: User
	// Distinct: true
	// Properties: 2
	// Has WHERE: true
	// Order By clauses: 1
	// Has Limit: true
	// Has Offset: true
}

func ExampleParseQuery_deep_property_path() {
	source := "SELECT user.profile.name FROM Document"
	lexer := gqlparser.NewLexer(source)

	query, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", query.Kind)
	fmt.Printf("Properties: %d\n", len(query.Properties))
	if len(query.Properties) > 0 {
		prop := &query.Properties[0]
		fmt.Printf("Property path: %s", prop.Name)
		for prop.Child != nil {
			prop = prop.Child
			fmt.Printf(".%s", prop.Name)
		}
		fmt.Println()
	}
	// Output:
	// Kind: Document
	// Properties: 1
	// Property path: user.profile.name
}

func ExampleNewLexer() {
	source := "SELECT name FROM User WHERE age > 25"
	lexer := gqlparser.NewLexer(source)

	tokenCount := 0
	for lexer.Next() {
		token, err := lexer.Read()
		if err != nil {
			log.Fatal(err)
		}
		if _, ok := token.(*gqlparser.WhitespaceToken); !ok {
			// Skip whitespace tokens for cleaner output
			tokenCount++
		}
	}

	fmt.Printf("Non-whitespace tokens: %d\n", tokenCount)
	// Output:
	// Non-whitespace tokens: 8
}

func ExampleParseQuery_errUnexpectedToken() {
	source := "INVALID QUERY SYNTAX"
	lexer := gqlparser.NewLexer(source)

	_, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		if errors.Is(err, gqlparser.ErrUnexpectedToken) {
			fmt.Printf("Parse error occurred: %v\n", err)
		} else {
			log.Fatal(err)
		}
	}
	// Output:
	// Parse error occurred: unexpected token: IN at 0 (expect to be any of ["SELECT"])
}

func ExampleParseQuery_errNoTokens() {
	source := "SELECT * FROM"
	lexer := gqlparser.NewLexer(source)

	_, err := gqlparser.ParseQuery(lexer)
	if err != nil {
		if errors.Is(err, gqlparser.ErrNoTokens) {
			fmt.Printf("Unexpected no tokens error occurred: %v\n", err)
		} else {
			log.Fatal(err)
		}
	}
	// Output:
	// Unexpected no tokens error occurred: no tokens
}

func ExampleParseQueryOrAggregationQuery_any() {
	// Regular query
	source1 := "SELECT * FROM User"
	lexer1 := gqlparser.NewLexer(source1)

	query, aggregationQuery, err := gqlparser.ParseQueryOrAggregationQuery(lexer1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Regular query - Query: %t, Aggregation: %t\n", query != nil, aggregationQuery != nil)

	// Aggregation query
	source2 := "SELECT COUNT(*) FROM User"
	lexer2 := gqlparser.NewLexer(source2)

	query, aggregationQuery, err = gqlparser.ParseQueryOrAggregationQuery(lexer2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Aggregation query - Query: %t, Aggregation: %t\n", query != nil, aggregationQuery != nil)
	// Output:
	// Regular query - Query: true, Aggregation: false
	// Aggregation query - Query: false, Aggregation: true
}
