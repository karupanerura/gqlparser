package gqlparser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/gqlparser"
)

func TestLexer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  string
		want    []gqlparser.Token
		wantErr bool
	}{
		{"Empty", "", nil, false},
		{"SingleKeyword", "SELECT", []gqlparser.Token{&gqlparser.KeywordToken{Name: "SELECT", RawContent: "SELECT"}}, false},
		{"SingleKeywordCaseInsensitive", "sElEcT", []gqlparser.Token{&gqlparser.KeywordToken{Name: "SELECT", RawContent: "sElEcT"}}, false},
		{
			name:   "EmptyBackQuoteString",
			source: "``",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '`', Content: "", RawContent: "``", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "BackQuoteString",
			source: "`Kind`",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '`', Content: "Kind", RawContent: "`Kind`", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "EscapedBackQuoteString",
			source: "`Kind\\`Kind\\``",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '`', Content: "Kind`Kind`", RawContent: "`Kind\\`Kind\\``", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "KeepBackQuoteString",
			source: "`Kind\\\"Kind\\\"`",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '`', Content: "Kind\"Kind\"", RawContent: "`Kind\\\"Kind\\\"`", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "EmptySingleQuoteString",
			source: "''",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '\'', Content: "", RawContent: "''", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "BackQuoteString",
			source: "'String'",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '\'', Content: "String", RawContent: "'String'", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "EscapedBackQuoteString",
			source: "'String\\'Quoted\\''",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '\'', Content: "String'Quoted'", RawContent: "'String\\'Quoted\\''", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "Integer",
			source: "123",
			want: []gqlparser.Token{
				&gqlparser.NumericToken{Int64: 123, RawContent: "123", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "IntegerWithPlusSign",
			source: "+123",
			want: []gqlparser.Token{
				&gqlparser.NumericToken{Int64: 123, RawContent: "+123", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "IntegerWithMinusSign",
			source: "-123",
			want: []gqlparser.Token{
				&gqlparser.NumericToken{Int64: -123, RawContent: "-123", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "EqualsCondition",
			source: "prop = 1",
			want: []gqlparser.Token{
				&gqlparser.SymbolToken{Content: "prop", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 4},
				&gqlparser.OperatorToken{Type: "=", Position: 5},
				&gqlparser.WhitespaceToken{Content: " ", Position: 6},
				&gqlparser.NumericToken{Int64: 1, RawContent: "1", Position: 7},
			},
			wantErr: false,
		},
		{
			name:   "LesserThanCondition",
			source: "prop < 1",
			want: []gqlparser.Token{
				&gqlparser.SymbolToken{Content: "prop", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 4},
				&gqlparser.OperatorToken{Type: "<", Position: 5},
				&gqlparser.WhitespaceToken{Content: " ", Position: 6},
				&gqlparser.NumericToken{Int64: 1, RawContent: "1", Position: 7},
			},
			wantErr: false,
		},
		{
			name:   "GraterThanCondition",
			source: "prop > 1",
			want: []gqlparser.Token{
				&gqlparser.SymbolToken{Content: "prop", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 4},
				&gqlparser.OperatorToken{Type: ">", Position: 5},
				&gqlparser.WhitespaceToken{Content: " ", Position: 6},
				&gqlparser.NumericToken{Int64: 1, RawContent: "1", Position: 7},
			},
			wantErr: false,
		},
		{
			name:   "LesserThanOrEqualsCondition",
			source: "prop <= 1",
			want: []gqlparser.Token{
				&gqlparser.SymbolToken{Content: "prop", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 4},
				&gqlparser.OperatorToken{Type: "<=", Position: 5},
				&gqlparser.WhitespaceToken{Content: " ", Position: 7},
				&gqlparser.NumericToken{Int64: 1, RawContent: "1", Position: 8},
			},
			wantErr: false,
		},
		{
			name:   "InCondition",
			source: "1 IN prop",
			want: []gqlparser.Token{
				&gqlparser.NumericToken{Int64: 1, RawContent: "1", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 1},
				&gqlparser.OperatorToken{Type: "IN", RawContent: "IN", Position: 2},
				&gqlparser.WhitespaceToken{Content: " ", Position: 4},
				&gqlparser.SymbolToken{Content: "prop", Position: 5},
			},
			wantErr: false,
		},
		{
			name:   "NotInCondition",
			source: "1 NOT IN prop",
			want: []gqlparser.Token{
				&gqlparser.NumericToken{Int64: 1, RawContent: "1", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 1},
				&gqlparser.OperatorToken{Type: "NOT", RawContent: "NOT", Position: 2},
				&gqlparser.WhitespaceToken{Content: " ", Position: 5},
				&gqlparser.OperatorToken{Type: "IN", RawContent: "IN", Position: 6},
				&gqlparser.WhitespaceToken{Content: " ", Position: 8},
				&gqlparser.SymbolToken{Content: "prop", Position: 9},
			},
			wantErr: false,
		},
		{
			name:   "GraterThanOrEqualsCondition",
			source: "prop >= 1",
			want: []gqlparser.Token{
				&gqlparser.SymbolToken{Content: "prop", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 4},
				&gqlparser.OperatorToken{Type: ">=", Position: 5},
				&gqlparser.WhitespaceToken{Content: " ", Position: 7},
				&gqlparser.NumericToken{Int64: 1, RawContent: "1", Position: 8},
			},
			wantErr: false,
		},
		{
			name:   "NestedPropertyCondition",
			source: "prop.subProp >= 1",
			want: []gqlparser.Token{
				&gqlparser.SymbolToken{Content: "prop", Position: 0},
				&gqlparser.OperatorToken{Type: ".", Position: 4},
				&gqlparser.SymbolToken{Content: "subProp", Position: 5},
				&gqlparser.WhitespaceToken{Content: " ", Position: 12},
				&gqlparser.OperatorToken{Type: ">=", Position: 13},
				&gqlparser.WhitespaceToken{Content: " ", Position: 15},
				&gqlparser.NumericToken{Int64: 1, RawContent: "1", Position: 16},
			},
			wantErr: false,
		},
		{
			name:   "BasicQuery",
			source: "SELECT * FROM Kind",
			want: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT", RawContent: "SELECT", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 6},
				&gqlparser.WildcardToken{Position: 7},
				&gqlparser.WhitespaceToken{Content: " ", Position: 8},
				&gqlparser.KeywordToken{Name: "FROM", RawContent: "FROM", Position: 9},
				&gqlparser.WhitespaceToken{Content: " ", Position: 13},
				&gqlparser.SymbolToken{Content: "Kind", Position: 14},
			},
			wantErr: false,
		},
		{
			name:   "ComplexQuery",
			source: "SELECT a, b, c FROM Kind",
			want: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT", RawContent: "SELECT", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 6},
				&gqlparser.SymbolToken{Content: "a", Position: 7},
				&gqlparser.OperatorToken{Type: ",", Position: 8},
				&gqlparser.WhitespaceToken{Content: " ", Position: 9},
				&gqlparser.SymbolToken{Content: "b", Position: 10},
				&gqlparser.OperatorToken{Type: ",", Position: 11},
				&gqlparser.WhitespaceToken{Content: " ", Position: 12},
				&gqlparser.SymbolToken{Content: "c", Position: 13},
				&gqlparser.WhitespaceToken{Content: " ", Position: 14},
				&gqlparser.KeywordToken{Name: "FROM", RawContent: "FROM", Position: 15},
				&gqlparser.WhitespaceToken{Content: " ", Position: 19},
				&gqlparser.SymbolToken{Content: "Kind", Position: 20},
			},
			wantErr: false,
		},
		{
			name:   "BindingTokenNumeric",
			source: "@123",
			want: []gqlparser.Token{
				&gqlparser.BindingToken{Index: 123, Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "BindingTokenNamed",
			source: "@paramName",
			want: []gqlparser.Token{
				&gqlparser.BindingToken{Name: "paramName", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "DoubleQuotedString",
			source: "\"Hello World\"",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '"', Content: "Hello World", RawContent: "\"Hello World\"", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "EscapedDoubleQuotedString",
			source: "\"Escaped\\\"Quote\"",
			want: []gqlparser.Token{
				&gqlparser.StringToken{Quote: '"', Content: "Escaped\"Quote", RawContent: "\"Escaped\\\"Quote\"", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "FloatingPointNumber",
			source: "3.14159",
			want: []gqlparser.Token{
				&gqlparser.NumericToken{Float64: 3.14159, Floating: true, RawContent: "3.14159", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "BooleanTrue",
			source: "TRUE",
			want: []gqlparser.Token{
				&gqlparser.BooleanToken{Value: true, RawContent: "TRUE", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "BooleanFalse",
			source: "FALSE",
			want: []gqlparser.Token{
				&gqlparser.BooleanToken{Value: false, RawContent: "FALSE", Position: 0},
			},
			wantErr: false,
		},
		{
			name:   "OrderByClause",
			source: "ORDER BY prop DESC",
			want: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "ORDER", RawContent: "ORDER", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 5},
				&gqlparser.KeywordToken{Name: "BY", RawContent: "BY", Position: 6},
				&gqlparser.WhitespaceToken{Content: " ", Position: 8},
				&gqlparser.SymbolToken{Content: "prop", Position: 9},
				&gqlparser.WhitespaceToken{Content: " ", Position: 13},
				&gqlparser.OrderToken{Descending: true, RawContent: "DESC", Position: 14},
			},
			wantErr: false,
		},
		{
			name:   "ComplexWhereClause",
			source: "WHERE prop1 = @1 AND prop2 > 10.5 OR prop3 IN ('a', 'b', 'c')",
			want: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "WHERE", RawContent: "WHERE", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 5},
				&gqlparser.SymbolToken{Content: "prop1", Position: 6},
				&gqlparser.WhitespaceToken{Content: " ", Position: 11},
				&gqlparser.OperatorToken{Type: "=", Position: 12},
				&gqlparser.WhitespaceToken{Content: " ", Position: 13},
				&gqlparser.BindingToken{Index: 1, Position: 14},
				&gqlparser.WhitespaceToken{Content: " ", Position: 16},
				&gqlparser.OperatorToken{Type: "AND", RawContent: "AND", Position: 17},
				&gqlparser.WhitespaceToken{Content: " ", Position: 20},
				&gqlparser.SymbolToken{Content: "prop2", Position: 21},
				&gqlparser.WhitespaceToken{Content: " ", Position: 26},
				&gqlparser.OperatorToken{Type: ">", Position: 27},
				&gqlparser.WhitespaceToken{Content: " ", Position: 28},
				&gqlparser.NumericToken{Float64: 10.5, Floating: true, RawContent: "10.5", Position: 29},
				&gqlparser.WhitespaceToken{Content: " ", Position: 33},
				&gqlparser.OperatorToken{Type: "OR", RawContent: "OR", Position: 34},
				&gqlparser.WhitespaceToken{Content: " ", Position: 36},
				&gqlparser.SymbolToken{Content: "prop3", Position: 37},
				&gqlparser.WhitespaceToken{Content: " ", Position: 42},
				&gqlparser.OperatorToken{Type: "IN", RawContent: "IN", Position: 43},
				&gqlparser.WhitespaceToken{Content: " ", Position: 45},
				&gqlparser.OperatorToken{Type: "(", Position: 46},
				&gqlparser.StringToken{Quote: '\'', Content: "a", RawContent: "'a'", Position: 47},
				&gqlparser.OperatorToken{Type: ",", Position: 50},
				&gqlparser.WhitespaceToken{Content: " ", Position: 51},
				&gqlparser.StringToken{Quote: '\'', Content: "b", RawContent: "'b'", Position: 52},
				&gqlparser.OperatorToken{Type: ",", Position: 55},
				&gqlparser.WhitespaceToken{Content: " ", Position: 56},
				&gqlparser.StringToken{Quote: '\'', Content: "c", RawContent: "'c'", Position: 57},
				&gqlparser.OperatorToken{Type: ")", Position: 60},
			},
			wantErr: false,
		},
		{
			name:   "DistinctClause",
			source: "SELECT DISTINCT prop FROM Kind",
			want: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT", RawContent: "SELECT", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 6},
				&gqlparser.KeywordToken{Name: "DISTINCT", RawContent: "DISTINCT", Position: 7},
				&gqlparser.WhitespaceToken{Content: " ", Position: 15},
				&gqlparser.SymbolToken{Content: "prop", Position: 16},
				&gqlparser.WhitespaceToken{Content: " ", Position: 20},
				&gqlparser.KeywordToken{Name: "FROM", RawContent: "FROM", Position: 21},
				&gqlparser.WhitespaceToken{Content: " ", Position: 25},
				&gqlparser.SymbolToken{Content: "Kind", Position: 26},
			},
			wantErr: false,
		},
		{
			name:   "AggregationQuery",
			source: "SELECT COUNT(*), SUM(value) FROM Kind",
			want: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT", RawContent: "SELECT", Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 6},
				&gqlparser.KeywordToken{Name: "COUNT", RawContent: "COUNT", Position: 7},
				&gqlparser.OperatorToken{Type: "(", Position: 12},
				&gqlparser.WildcardToken{Position: 13},
				&gqlparser.OperatorToken{Type: ")", Position: 14},
				&gqlparser.OperatorToken{Type: ",", Position: 15},
				&gqlparser.WhitespaceToken{Content: " ", Position: 16},
				&gqlparser.KeywordToken{Name: "SUM", RawContent: "SUM", Position: 17},
				&gqlparser.OperatorToken{Type: "(", Position: 20},
				&gqlparser.SymbolToken{Content: "value", Position: 21},
				&gqlparser.OperatorToken{Type: ")", Position: 26},
				&gqlparser.WhitespaceToken{Content: " ", Position: 27},
				&gqlparser.KeywordToken{Name: "FROM", RawContent: "FROM", Position: 28},
				&gqlparser.WhitespaceToken{Content: " ", Position: 32},
				&gqlparser.SymbolToken{Content: "Kind", Position: 33},
			},
			wantErr: false,
		},
		{
			name:   "ConsecutiveOperators",
			source: "prop!=10",
			want: []gqlparser.Token{
				&gqlparser.SymbolToken{Content: "prop", Position: 0},
				&gqlparser.OperatorToken{Type: "!=", Position: 4},
				&gqlparser.NumericToken{Int64: 10, RawContent: "10", Position: 6},
			},
			wantErr: false,
		},
		{
			name:   "MultipleLevelBindings",
			source: "@1 @param1 @2",
			want: []gqlparser.Token{
				&gqlparser.BindingToken{Index: 1, Position: 0},
				&gqlparser.WhitespaceToken{Content: " ", Position: 2},
				&gqlparser.BindingToken{Name: "param1", Position: 3},
				&gqlparser.WhitespaceToken{Content: " ", Position: 10},
				&gqlparser.BindingToken{Index: 2, Position: 11},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lexer := gqlparser.NewLexer(tt.source)
			got, err := gqlparser.ReadAllTokens(lexer)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tokenize() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if df := cmp.Diff(tt.want, got); df != "" {
				t.Errorf("Tokenize() = %+v, want %+v, diff = %s", got, tt.want, df)
			}
		})
	}
}

func FuzzLexer(f *testing.F) {
	f.Fuzz(func(t *testing.T, src string) {
		lexer := gqlparser.NewLexer(src)
		_, _ = gqlparser.ReadAllTokens(lexer)
		// should be no panics
	})
}
