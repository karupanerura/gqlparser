package gqlparser_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/gqlparser"
)

type integrateTestCase struct {
	name    string
	source  string
	want    gqlparser.Syntax
	wantErr bool
}

var (
	queryTests = []integrateTestCase{
		{"Empty", "", nil, true},
		{
			name:   "SimpleQuery",
			source: "SELECT * FROM `Kind`",
			want: &gqlparser.Query{
				Kind: "Kind",
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithProperties",
			source: "SELECT `Name`, `Age` FROM `Kind`",
			want: &gqlparser.Query{
				Properties: []gqlparser.Property{"Name", "Age"},
				Kind:       "Kind",
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithWhere",
			source: "SELECT * FROM `Kind` WHERE `Name` = 'Alice'",
			want: &gqlparser.Query{
				Kind: "Kind",
				Where: &gqlparser.EitherComparatorCondition{
					Property:   "Name",
					Comparator: gqlparser.EqualsEitherComparator,
					Value:      "Alice",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithOrderBy",
			source: "SELECT * FROM `Kind` ORDER BY `Name` DESC",
			want: &gqlparser.Query{
				Kind: "Kind",
				OrderBy: []gqlparser.OrderBy{
					{
						Property:   "Name",
						Descending: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithLimit",
			source: "SELECT * FROM `Kind` LIMIT 10",
			want: &gqlparser.Query{
				Kind: "Kind",
				Limit: &gqlparser.Limit{
					Position: 10,
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithLimitFirst",
			source: "SELECT * FROM `Kind` LIMIT FIRST (12, @1)",
			want: &gqlparser.Query{
				Kind: "Kind",
				Limit: &gqlparser.Limit{
					Position: 12,
					Cursor:   &gqlparser.IndexedBinding{Index: 1},
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithOffset",
			source: "SELECT * FROM `Kind` OFFSET 10",
			want: &gqlparser.Query{
				Kind: "Kind",
				Offset: &gqlparser.Offset{
					Position: 10,
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithOffsets",
			source: "SELECT * FROM `Kind` OFFSET @1 + 2",
			want: &gqlparser.Query{
				Kind: "Kind",
				Offset: &gqlparser.Offset{
					Position: 2,
					Cursor:   &gqlparser.IndexedBinding{Index: 1},
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithOffsets",
			source: "SELECT * FROM `Kind` OFFSET @1 + +2",
			want: &gqlparser.Query{
				Kind: "Kind",
				Offset: &gqlparser.Offset{
					Position: 2,
					Cursor:   &gqlparser.IndexedBinding{Index: 1},
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleQueryWithLimitAndOffset",
			source: "SELECT * FROM `Kind` LIMIT 10 OFFSET 10",
			want: &gqlparser.Query{
				Kind: "Kind",
				Limit: &gqlparser.Limit{
					Position: 10,
				},
				Offset: &gqlparser.Offset{
					Position: 10,
				},
			},
			wantErr: false,
		},
	}
	aggregationQueryTests = []integrateTestCase{
		{"Empty", "", nil, true},
		{
			name:   "SimpleCountQuery",
			source: "SELECT COUNT(*) FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountAggregation{},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleCountQueryWithSymbolAlias",
			source: "SELECT COUNT(*) AS cnt FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountAggregation{Alias: "cnt"},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleCountQueryWithQuotedSymbolAlias",
			source: "SELECT COUNT(*) AS `count` FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountAggregation{Alias: "count"},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleCountUpToQuery",
			source: "SELECT COUNT_UP_TO(10) FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountUpToAggregation{
						Limit: 10,
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleCountUpToQueryWithSymbolAlias",
			source: "SELECT COUNT_UP_TO(10) AS c FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountUpToAggregation{
						Limit: 10,
						Alias: "c",
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleCountUpToQueryWithQuotedSymbolAlias",
			source: "SELECT COUNT_UP_TO(10) AS `count_up_to` FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountUpToAggregation{
						Limit: 10,
						Alias: "count_up_to",
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleSumQuery",
			source: "SELECT SUM(n) FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.SumAggregation{
						Property: "n",
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleSumQueryWithSymbolAlias",
			source: "SELECT SUM(n) AS s FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.SumAggregation{
						Property: "n",
						Alias:    "s",
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleSumQueryWithQuotedSymbolAlias",
			source: "SELECT SUM(n) AS `sum` FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.SumAggregation{
						Property: "n",
						Alias:    "sum",
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleAvgQuery",
			source: "SELECT AVG(n) FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.AvgAggregation{
						Property: "n",
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleAvgQueryWithSymbolAlias",
			source: "SELECT AVG(n) AS a FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.AvgAggregation{
						Property: "n",
						Alias:    "a",
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleAvgQueryWithQuotedSymbolAlias",
			source: "SELECT AVG(n) AS `avg` FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.AvgAggregation{
						Property: "n",
						Alias:    "avg",
					},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "MultipleAggregations",
			source: "SELECT AVG(n), SUM(n), COUNT_UP_TO(100), COUNT(*) FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.AvgAggregation{Property: "n"},
					&gqlparser.SumAggregation{Property: "n"},
					&gqlparser.CountUpToAggregation{Limit: 100},
					&gqlparser.CountAggregation{},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "MultipleAggregationsWithAliases",
			source: "SELECT AVG(n) AS `avg`, SUM(n) AS `sum`, COUNT_UP_TO(100) AS `count_up_to`, COUNT(*) AS `count` FROM `Kind`",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.AvgAggregation{Property: "n", Alias: "avg"},
					&gqlparser.SumAggregation{Property: "n", Alias: "sum"},
					&gqlparser.CountUpToAggregation{Limit: 100, Alias: "count_up_to"},
					&gqlparser.CountAggregation{Alias: "count"},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleCountWithWhereCondition",
			source: "SELECT COUNT(*) FROM `Kind` WHERE deleted = false",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountAggregation{},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
					Where: &gqlparser.EitherComparatorCondition{
						Property:   "deleted",
						Comparator: gqlparser.EqualsEitherComparator,
						Value:      false,
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleCountQueryWithAggregateSyntax",
			source: "AGGREGATE COUNT(*) OVER (SELECT * FROM `Kind`)",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountAggregation{},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleCountQueryWithAggregateSyntax",
			source: "AGGREGATE COUNT(*) AS `count` OVER (SELECT * FROM `Kind` WHERE deleted = false)",
			want: &gqlparser.AggregationQuery{
				Aggregations: []gqlparser.Aggregation{
					&gqlparser.CountAggregation{Alias: "count"},
				},
				Query: gqlparser.Query{
					Kind: "Kind",
					Where: &gqlparser.EitherComparatorCondition{
						Property:   "deleted",
						Comparator: gqlparser.EqualsEitherComparator,
						Value:      false,
					},
				},
			},
			wantErr: false,
		},
	}
)

func TestParseQuery_FromString(t *testing.T) {
	t.Parallel()

	tests := append([]integrateTestCase{}, queryTests...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := gqlparser.ParseQuery(gqlparser.NewLexer(tt.source))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseQuery() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseAggregationQuery_FromString(t *testing.T) {
	t.Parallel()

	tests := append([]integrateTestCase{}, aggregationQueryTests...)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := gqlparser.ParseAggregationQuery(gqlparser.NewLexer(tt.source))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAggregationQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseAggregationQuery() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseQueryOrAggregationQuery_FromString(t *testing.T) {
	t.Parallel()

	t.Run("Query", func(t *testing.T) {
		for _, tt := range queryTests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, aq, err := gqlparser.ParseQueryOrAggregationQuery(gqlparser.NewLexer(tt.source))
				if (err != nil) != tt.wantErr {
					t.Errorf("ParseQueryOrAggregationQuery() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err != nil {
					return
				}
				if aq != nil {
					t.Errorf("aggregation query should be nil but got: %v", aq)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("ParseQueryOrAggregationQuery() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("AggregationQuery", func(t *testing.T) {
		for _, tt := range aggregationQueryTests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				q, got, err := gqlparser.ParseQueryOrAggregationQuery(gqlparser.NewLexer(tt.source))
				if (err != nil) != tt.wantErr {
					t.Errorf("ParseQueryOrAggregationQuery() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err != nil {
					return
				}
				if q != nil {
					t.Errorf("query should be nil but got: %v", q)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("ParseQueryOrAggregationQuery() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}

func TestParseKey_FromString(t *testing.T) {
	t.Parallel()

	tests := []integrateTestCase{
		{"Empty", "", nil, true},
		{"EmptyKeyBody", "KEY()", nil, true},
		{
			name:   "SimpleNameKey",
			source: `KEY(Foo, 'bar')`,
			want: &gqlparser.Key{
				Path: []*gqlparser.KeyPath{
					{Kind: "Foo", Name: "bar"},
				},
			},
			wantErr: false,
		},
		{
			name:   "SimpleIDKey",
			source: `KEY(Foo, 123)`,
			want: &gqlparser.Key{
				Path: []*gqlparser.KeyPath{
					{Kind: "Foo", ID: 123},
				},
			},
			wantErr: false,
		},
		{
			name:   "WithProject",
			source: `KEY(PROJECT("baz"), Foo, 'bar')`,
			want: &gqlparser.Key{
				ProjectID: "baz",
				Path: []*gqlparser.KeyPath{
					{Kind: "Foo", Name: "bar"},
				},
			},
			wantErr: false,
		},
		{
			name:   "WithNamespace",
			source: `KEY(NAMESPACE("baz"), Foo, 123)`,
			want: &gqlparser.Key{
				Namespace: "baz",
				Path: []*gqlparser.KeyPath{
					{Kind: "Foo", ID: 123},
				},
			},
			wantErr: false,
		},
		{
			name:   "WithProjectAndNamespace",
			source: `KEY(PROJECT("foo"), NAMESPACE("bar"), Buz, 777)`,
			want: &gqlparser.Key{
				ProjectID: "foo",
				Namespace: "bar",
				Path: []*gqlparser.KeyPath{
					{Kind: "Buz", ID: 777},
				},
			},
			wantErr: false,
		},
		{
			name:   "Ancestor",
			source: `KEY(Parent, 1, Child, 9)`,
			want: &gqlparser.Key{
				Path: []*gqlparser.KeyPath{
					{Kind: "Parent", ID: 1},
					{Kind: "Child", ID: 9},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := gqlparser.ParseKey(gqlparser.NewLexer(tt.source))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseKey() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseCondition_FromString(t *testing.T) {
	t.Parallel()

	tests := []integrateTestCase{
		{"Empty", "", nil, true},
		{
			name:   "EqualsWithInteger",
			source: `a = 1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "EqualsByInteger",
			source: `1 = a`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithPlusInteger",
			source: `a = +1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "EqualsByPlusInteger",
			source: `+1 = a`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithMinusInteger",
			source: `a = -1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      int64(-1),
			},
			wantErr: false,
		},
		{
			name:   "EqualsByMinusInteger",
			source: `-1 = a`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      int64(-1),
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithFloatingNumber",
			source: `a = 0.5`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      float64(0.5),
			},
			wantErr: false,
		},
		{
			name:   "EqualsByFloatingNumber",
			source: `0.5 = a`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      float64(0.5),
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithPlusFloatingNumber",
			source: `a = +0.5`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      float64(0.5),
			},
			wantErr: false,
		},
		{
			name:   "EqualsByPlusFloatingNumber",
			source: `+0.5 = a`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      float64(0.5),
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithMinusFloatingNumber",
			source: `a = -0.5`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      float64(-0.5),
			},
			wantErr: false,
		},
		{
			name:   "EqualsByMinusFloatingNumber",
			source: `-0.5 = a`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      float64(-0.5),
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithString",
			source: `a = 'string'`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      "string",
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithTrue",
			source: `a = true`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      true,
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithFalse",
			source: `a = false`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      false,
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithIndexedBindingToken",
			source: `a = @1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      &gqlparser.IndexedBinding{Index: 1},
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithNamedBindingToken",
			source: `a = @foo`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      &gqlparser.NamedBinding{Name: "foo"},
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithArray",
			source: `a = ARRAY(1, 2, 3)`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      []any{int64(1), int64(2), int64(3)},
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithBlob",
			source: `a = BLOB("YmluYXJ5")`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      []byte("binary"),
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithDateTime",
			source: `a = DATETIME("2013-09-29T09:30:20.00002-08:00")`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value:      time.Date(2013, 9, 29, 9, 30, 20, 20000, time.FixedZone("PST", -8*60*60)),
			},
			wantErr: false,
		},
		{
			name:   "NotEquals",
			source: `a != 1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.NotEqualsEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "GreaterThan",
			source: `a > 1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.GreaterThanEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "GreaterThanOrEquals",
			source: `a >= 1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.GreaterThanOrEqualsThanEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "LesserThan",
			source: `a < 1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.LesserThanEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "LesserThanOrEquals",
			source: `a <= 1`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.LesserThanOrEqualsEitherComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "IsNull",
			source: `a IS NULL`,
			want: &gqlparser.IsNullCondition{
				Property: "a",
			},
			wantErr: false,
		},
		{
			name:   "Contains",
			source: `a CONTAINS 1`,
			want: &gqlparser.ForwardComparatorCondition{
				Comparator: gqlparser.ContainsForwardComparator,
				Property:   "a",
				Value:      int64(1),
			},
			wantErr: false,
		},
		{
			name:   "In",
			source: `a IN ARRAY(1, 2, 3)`,
			want: &gqlparser.ForwardComparatorCondition{
				Comparator: gqlparser.InForwardComparator,
				Property:   "a",
				Value:      []any{int64(1), int64(2), int64(3)},
			},
			wantErr: false,
		},
		{
			name:   "NotIn",
			source: `a NOT IN ARRAY(2, 3, 4)`,
			want: &gqlparser.ForwardComparatorCondition{
				Comparator: gqlparser.NotInForwardComparator,
				Property:   "a",
				Value:      []any{int64(2), int64(3), int64(4)},
			},
			wantErr: false,
		},
		{
			name:   "HasAncestor",
			source: `__key__ HAS ANCESTOR KEY(Parent, 1000)`,
			want: &gqlparser.ForwardComparatorCondition{
				Comparator: gqlparser.HasAncestorForwardComparator,
				Property:   "__key__",
				Value: &gqlparser.Key{
					Path: []*gqlparser.KeyPath{
						{Kind: "Parent", ID: 1000},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "InBackwardCondition",
			source: `ARRAY(KEY(Kind, 1), KEY(Kind, 2), KEY(Kind, 3)) IN __key__`,
			want: &gqlparser.BackwardComparatorCondition{
				Comparator: gqlparser.InBackwardComparator,
				Property:   "__key__",
				Value: []any{
					&gqlparser.Key{
						Path: []*gqlparser.KeyPath{
							{Kind: "Kind", ID: 1},
						},
					},
					&gqlparser.Key{
						Path: []*gqlparser.KeyPath{
							{Kind: "Kind", ID: 2},
						},
					},
					&gqlparser.Key{
						Path: []*gqlparser.KeyPath{
							{Kind: "Kind", ID: 3},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "HasDescendant",
			source: `KEY(Parent, 1000) HAS DESCENDANT __key__`,
			want: &gqlparser.BackwardComparatorCondition{
				Comparator: gqlparser.HasDescendantBackwardComparator,
				Property:   "__key__",
				Value: &gqlparser.Key{
					Path: []*gqlparser.KeyPath{
						{Kind: "Parent", ID: 1000},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "EqualsWithArrayWithComplexItems",
			source: `a = ARRAY(777, -0.25, "foo", true, NULL, @1, @foo, ARRAY(1, 2, 3), BLOB("YmluYXJ5"), DATETIME("2013-09-29T09:30:20.00002-08:00"), KEY(Kind, 1))`,
			want: &gqlparser.EitherComparatorCondition{
				Comparator: gqlparser.EqualsEitherComparator,
				Property:   "a",
				Value: []any{
					int64(777),
					float64(-0.25),
					"foo",
					true,
					nil,
					&gqlparser.IndexedBinding{Index: 1},
					&gqlparser.NamedBinding{Name: "foo"},
					[]any{int64(1), int64(2), int64(3)},
					[]byte("binary"),
					time.Date(2013, 9, 29, 9, 30, 20, 20000, time.FixedZone("PST", -8*60*60)),
					&gqlparser.Key{
						Path: []*gqlparser.KeyPath{
							{Kind: "Kind", ID: 1},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// got, err := gqlparser.ParseCondition(&debugTokenSource{source: gqlparser.NewLexer(tt.source), logger: t})
			got, err := gqlparser.ParseCondition(gqlparser.NewLexer(tt.source))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseCondition() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
