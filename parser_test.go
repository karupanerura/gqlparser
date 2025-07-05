package gqlparser_test

import (
	"encoding/binary"
	rand "math/rand/v2"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/gqlparser"
)

func TestParseQuery(t *testing.T) {
	// t.Parallel()

	tests := []struct {
		name    string
		tokens  []gqlparser.Token
		want    *gqlparser.Query
		wantErr bool
	}{
		{
			name: "QueryEverything",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.WildcardToken{},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
			},
			want: &gqlparser.Query{
				Kind: "Kind",
			},
		},
		{
			name: "QuerySingleColumn",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "a"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
			},
			want: &gqlparser.Query{
				Properties: []gqlparser.Property{{Name: "a"}},
				Kind:       "Kind",
			},
		},
		{
			name: "QueryMultipleSingleColumn",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "a"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.SymbolToken{Content: "b"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.SymbolToken{Content: "c"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
			},
			want: &gqlparser.Query{
				Properties: []gqlparser.Property{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				Kind:       "Kind",
			},
		},
		{
			name: "QueryEverythingWithDistinctSingleColumn",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "DISTINCT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "a"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
			},
			want: &gqlparser.Query{
				Distinct:   true,
				Properties: []gqlparser.Property{{Name: "a"}},
				Kind:       "Kind",
			},
		},
		{
			name: "QueryEverythingWithDistinctMultipleColumn",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "DISTINCT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "a"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.SymbolToken{Content: "b"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.SymbolToken{Content: "c"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
			},
			want: &gqlparser.Query{
				Distinct:   true,
				Properties: []gqlparser.Property{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				Kind:       "Kind",
			},
		},
		{
			name: "QueryEverythingWithDistinctMultipleColumnAndWhiteSpaces",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "DISTINCT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "a"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "b"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "c"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "d"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
			},
			want: &gqlparser.Query{
				Distinct:   true,
				Properties: []gqlparser.Property{{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}},
				Kind:       "Kind",
			},
		},
		{
			name: "QueryEverythingWithDistinctOnMultipleColumn",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "DISTINCT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "ON"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "("},
				&gqlparser.SymbolToken{Content: "a"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.SymbolToken{Content: "b"},
				&gqlparser.OperatorToken{Type: ")"},
				&gqlparser.SymbolToken{Content: "c"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.SymbolToken{Content: "d"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
			},
			want: &gqlparser.Query{
				DistinctOn: []gqlparser.Property{{Name: "a"}, {Name: "b"}},
				Properties: []gqlparser.Property{{Name: "c"}, {Name: "d"}},
				Kind:       "Kind",
			},
		},
		{
			name: "QueryWithSimpleWhere",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.WildcardToken{},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "WHERE"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "col"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "="},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.NumericToken{Int64: 1},
			},
			want: &gqlparser.Query{
				Kind:  "Kind",
				Where: &gqlparser.EitherComparatorCondition{Comparator: "=", Property: gqlparser.Property{Name: "col"}, Value: int64(1)},
			},
		},
		{
			name: "QueryWithSimpleWhereWithBinding",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.WildcardToken{},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "WHERE"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "col"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "="},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.BindingToken{Index: 1},
			},
			want: &gqlparser.Query{
				Kind:  "Kind",
				Where: &gqlparser.EitherComparatorCondition{Comparator: "=", Property: gqlparser.Property{Name: "col"}, Value: &gqlparser.IndexedBinding{Index: 1}},
			},
		},
		{
			name: "QueryWithSimpleWhereAndGrouping",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.WildcardToken{},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "WHERE"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "("},
				&gqlparser.SymbolToken{Content: "col"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "="},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.NumericToken{Int64: 1},
				&gqlparser.OperatorToken{Type: ")"},
			},
			want: &gqlparser.Query{
				Kind:  "Kind",
				Where: &gqlparser.EitherComparatorCondition{Comparator: "=", Property: gqlparser.Property{Name: "col"}, Value: int64(1)},
			},
		},
		{
			name: "ComplexQuery",
			tokens: []gqlparser.Token{
				&gqlparser.KeywordToken{Name: "SELECT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "DISTINCT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "ON"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "("},
				&gqlparser.SymbolToken{Content: "a"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "b"},
				&gqlparser.OperatorToken{Type: ")"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "c"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "d"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FROM"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "Kind"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "WHERE"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "KEY"},
				&gqlparser.OperatorToken{Type: "("},
				&gqlparser.SymbolToken{Content: "Kind"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.NumericToken{Int64: 1},
				&gqlparser.OperatorToken{Type: ")"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "HAS"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "DESCENDANT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "__key__"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "OR"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "__key__"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "HAS"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "ANCESTOR"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "KEY"},
				&gqlparser.OperatorToken{Type: "("},
				&gqlparser.SymbolToken{Content: "Kind"},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.StringToken{Quote: '"', Content: "key1"},
				&gqlparser.OperatorToken{Type: ")"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "AND"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "("},
				&gqlparser.SymbolToken{Content: "col"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "="},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.NumericToken{Int64: 1},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "OR"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "col"},
				&gqlparser.OperatorToken{Type: "."},
				&gqlparser.SymbolToken{Content: "child"},
				&gqlparser.OperatorToken{Type: "."},
				&gqlparser.SymbolToken{Content: "grand_child"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "IS"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "NULL"},
				&gqlparser.OperatorToken{Type: ")"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "ORDER"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "BY"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "foo"},
				&gqlparser.OperatorToken{Type: "."},
				&gqlparser.SymbolToken{Content: "bar"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OrderToken{Descending: true},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.SymbolToken{Content: "bar"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "LIMIT"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "FIRST"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "("},
				&gqlparser.NumericToken{Int64: 11},
				&gqlparser.OperatorToken{Type: ","},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.BindingToken{Index: 1},
				&gqlparser.OperatorToken{Type: ")"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.KeywordToken{Name: "OFFSET"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.BindingToken{Name: "foo"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.OperatorToken{Type: "+"},
				&gqlparser.WhitespaceToken{Content: " "},
				&gqlparser.NumericToken{Int64: 2},
			},
			want: &gqlparser.Query{
				Properties: []gqlparser.Property{{Name: "c"}, {Name: "d"}},
				DistinctOn: []gqlparser.Property{{Name: "a"}, {Name: "b"}},
				Kind:       "Kind",
				Where: &gqlparser.OrCompoundCondition{
					Left: &gqlparser.BackwardComparatorCondition{Comparator: "HAS DESCENDANT", Property: gqlparser.Property{Name: "__key__"}, Value: &gqlparser.Key{Path: []*gqlparser.KeyPath{{Kind: "Kind", ID: 1}}}},
					Right: &gqlparser.AndCompoundCondition{
						Left: &gqlparser.ForwardComparatorCondition{Comparator: "HAS ANCESTOR", Property: gqlparser.Property{Name: "__key__"}, Value: &gqlparser.Key{Path: []*gqlparser.KeyPath{{Kind: "Kind", Name: "key1"}}}},
						Right: &gqlparser.OrCompoundCondition{
							Left:  &gqlparser.EitherComparatorCondition{Comparator: "=", Property: gqlparser.Property{Name: "col"}, Value: int64(1)},
							Right: &gqlparser.IsNullCondition{Property: gqlparser.Property{Name: "col", Child: &gqlparser.Property{Name: "child", Child: &gqlparser.Property{Name: "grand_child"}}}},
						},
					},
				},
				OrderBy: []gqlparser.OrderBy{
					{Property: gqlparser.Property{Name: "foo", Child: &gqlparser.Property{Name: "bar"}}, Descending: true},
					{Property: gqlparser.Property{Name: "bar"}},
				},
				Limit: &gqlparser.Limit{
					Position: 11,
					Cursor:   &gqlparser.IndexedBinding{Index: 1},
				},
				Offset: &gqlparser.Offset{
					Position: 2,
					Cursor:   &gqlparser.NamedBinding{Name: "foo"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			normalizeTokens(tt.tokens)

			var query string
			for _, token := range tt.tokens {
				query += token.GetContent()
			}
			t.Log(query)

			got, err := gqlparser.ParseQuery(defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if df := cmp.Diff(tt.want, got); df != "" {
				t.Errorf("ParseQuery() = %+v, want %+v, diff = %s", got, tt.want, df)
			}
		})
	}
}

func normalizeTokens(tokens []gqlparser.Token) {
	pos := 0
	for _, token := range tokens {
		switch t := token.(type) {
		case *gqlparser.StringToken:
			t.Position = pos
			t.RawContent = string([]byte{t.Quote}) + t.Content + string([]byte{t.Quote})
			pos += len(t.GetContent())
		case *gqlparser.OperatorToken:
			t.Position = pos
			t.RawContent = t.Type
			pos += len(t.GetContent())
		case *gqlparser.WildcardToken:
			t.Position = pos
			pos += len(t.GetContent())
		case *gqlparser.BooleanToken:
			t.Position = pos
			if t.Value {
				t.RawContent = "TRUE"
			} else {
				t.RawContent = "FALSE"
			}
			pos += len(t.GetContent())
		case *gqlparser.OrderToken:
			t.Position = pos
			if t.Descending {
				t.RawContent = "DESC"
			} else {
				t.RawContent = "ASC"
			}
			pos += len(t.GetContent())
		case *gqlparser.SymbolToken:
			t.Position = pos
			pos += len(t.GetContent())
		case *gqlparser.KeywordToken:
			t.Position = pos
			t.RawContent = t.Name
			pos += len(t.GetContent())
		case *gqlparser.NumericToken:
			t.Position = pos
			if t.Floating {
				t.RawContent = strconv.FormatFloat(t.Float64, 'f', 10, 64)
			} else {
				t.RawContent = strconv.FormatInt(t.Int64, 10)
			}
			pos += len(t.GetContent())
		case *gqlparser.WhitespaceToken:
			t.Position = pos
			pos += len(t.GetContent())
		case *gqlparser.BindingToken:
			t.Position = pos
			pos += len(t.GetContent())
		}
	}
}

func FuzzParseQueryOrAggregationQuery(f *testing.F) {
	f.Fuzz(func(t *testing.T, s1, s2, s3, s4 string, i1, i2, i3 int64, f1, f2, f3 float64, length int, uint32seed uint32) {
		parts := []gqlparser.Token{
			&gqlparser.KeywordToken{Name: "SELECT"},
			&gqlparser.KeywordToken{Name: "FROM"},
			&gqlparser.KeywordToken{Name: "WHERE"},
			&gqlparser.KeywordToken{Name: "AGGREGATE"},
			&gqlparser.KeywordToken{Name: "OVER"},
			&gqlparser.KeywordToken{Name: "COUNT"},
			&gqlparser.KeywordToken{Name: "COUNT_UP_TO"},
			&gqlparser.KeywordToken{Name: "SUM"},
			&gqlparser.KeywordToken{Name: "AVG"},
			&gqlparser.KeywordToken{Name: "AS"},
			&gqlparser.KeywordToken{Name: "DISTINCT"},
			&gqlparser.KeywordToken{Name: "ON"},
			&gqlparser.KeywordToken{Name: "ORDER"},
			&gqlparser.KeywordToken{Name: "BY"},
			&gqlparser.KeywordToken{Name: "LIMIT"},
			&gqlparser.KeywordToken{Name: "FIRST"},
			&gqlparser.KeywordToken{Name: "OFFSET"},
			&gqlparser.KeywordToken{Name: "KEY"},
			&gqlparser.KeywordToken{Name: "PROJECT"},
			&gqlparser.KeywordToken{Name: "NAMESPACE"},
			&gqlparser.KeywordToken{Name: "ARRAY"},
			&gqlparser.KeywordToken{Name: "BLOB"},
			&gqlparser.KeywordToken{Name: "DATETIME"},
			&gqlparser.KeywordToken{Name: "NULL"},
			&gqlparser.WildcardToken{},
			&gqlparser.WhitespaceToken{Content: " "},
			&gqlparser.WhitespaceToken{Content: "\n"},
			&gqlparser.WhitespaceToken{Content: "\t"},
			&gqlparser.OperatorToken{Type: "AND"},
			&gqlparser.OperatorToken{Type: "OR"},
			&gqlparser.OperatorToken{Type: "="},
			&gqlparser.OperatorToken{Type: "!="},
			&gqlparser.OperatorToken{Type: ">"},
			&gqlparser.OperatorToken{Type: ">="},
			&gqlparser.OperatorToken{Type: "<"},
			&gqlparser.OperatorToken{Type: "<="},
			&gqlparser.OperatorToken{Type: "IN"},
			&gqlparser.OperatorToken{Type: "NOT"},
			&gqlparser.OperatorToken{Type: "CONTAINS"},
			&gqlparser.OperatorToken{Type: "HAS"},
			&gqlparser.OperatorToken{Type: "ANCESTOR"},
			&gqlparser.OperatorToken{Type: "DESCENDANT"},
			&gqlparser.OperatorToken{Type: "("},
			&gqlparser.OperatorToken{Type: ")"},
			&gqlparser.OperatorToken{Type: ","},
			&gqlparser.StringToken{Quote: '"', Content: s1},
			&gqlparser.StringToken{Quote: '\'', Content: s2},
			&gqlparser.StringToken{Quote: '`', Content: s3},
			&gqlparser.SymbolToken{Content: "symbol"},
			&gqlparser.NumericToken{Int64: i1},
			&gqlparser.NumericToken{Int64: i2},
			&gqlparser.NumericToken{Int64: i3},
			&gqlparser.NumericToken{Floating: true, Float64: f1},
			&gqlparser.NumericToken{Floating: true, Float64: f2},
			&gqlparser.NumericToken{Floating: true, Float64: f3},
			&gqlparser.BooleanToken{Value: true},
			&gqlparser.BooleanToken{Value: false},
			&gqlparser.BindingToken{Index: 1},
			&gqlparser.BindingToken{Name: "foo"},
			&gqlparser.OrderToken{Descending: false},
			&gqlparser.OrderToken{Descending: true},
		}

		var seed [32]byte
		binary.BigEndian.PutUint32(seed[:], uint32seed)

		r := rand.New(rand.NewChaCha8(seed))
		var tokens []gqlparser.Token
		for len(tokens) < length {
			tokens = append(tokens, parts[r.IntN(len(parts))])
		}
		normalizeTokens(tokens)

		_, _, _ = gqlparser.ParseQueryOrAggregationQuery(defaultTokenSourceFactory.NewSliceTokenSource(tokens))
		// should be no panics
	})
}

func FuzzParseAggregationQuery(f *testing.F) {
	f.Fuzz(func(t *testing.T, s1, s2, s3, s4 string, i1, i2, i3 int64, f1, f2, f3 float64, length int, uint32seed uint32) {
		parts := []gqlparser.Token{
			&gqlparser.KeywordToken{Name: "SELECT"},
			&gqlparser.KeywordToken{Name: "FROM"},
			&gqlparser.KeywordToken{Name: "WHERE"},
			&gqlparser.KeywordToken{Name: "AGGREGATE"},
			&gqlparser.KeywordToken{Name: "OVER"},
			&gqlparser.KeywordToken{Name: "COUNT"},
			&gqlparser.KeywordToken{Name: "COUNT_UP_TO"},
			&gqlparser.KeywordToken{Name: "SUM"},
			&gqlparser.KeywordToken{Name: "AVG"},
			&gqlparser.KeywordToken{Name: "AS"},
			&gqlparser.KeywordToken{Name: "KEY"},
			&gqlparser.KeywordToken{Name: "PROJECT"},
			&gqlparser.KeywordToken{Name: "NAMESPACE"},
			&gqlparser.KeywordToken{Name: "ARRAY"},
			&gqlparser.KeywordToken{Name: "BLOB"},
			&gqlparser.KeywordToken{Name: "DATETIME"},
			&gqlparser.KeywordToken{Name: "NULL"},
			&gqlparser.WildcardToken{},
			&gqlparser.WhitespaceToken{Content: " "},
			&gqlparser.WhitespaceToken{Content: "\n"},
			&gqlparser.WhitespaceToken{Content: "\t"},
			&gqlparser.OperatorToken{Type: "AND"},
			&gqlparser.OperatorToken{Type: "OR"},
			&gqlparser.OperatorToken{Type: "="},
			&gqlparser.OperatorToken{Type: "!="},
			&gqlparser.OperatorToken{Type: ">"},
			&gqlparser.OperatorToken{Type: ">="},
			&gqlparser.OperatorToken{Type: "<"},
			&gqlparser.OperatorToken{Type: "<="},
			&gqlparser.OperatorToken{Type: "IN"},
			&gqlparser.OperatorToken{Type: "NOT"},
			&gqlparser.OperatorToken{Type: "CONTAINS"},
			&gqlparser.OperatorToken{Type: "HAS"},
			&gqlparser.OperatorToken{Type: "ANCESTOR"},
			&gqlparser.OperatorToken{Type: "DESCENDANT"},
			&gqlparser.OperatorToken{Type: "("},
			&gqlparser.OperatorToken{Type: ")"},
			&gqlparser.OperatorToken{Type: ","},
			&gqlparser.StringToken{Quote: '"', Content: s1},
			&gqlparser.StringToken{Quote: '\'', Content: s2},
			&gqlparser.StringToken{Quote: '`', Content: s3},
			&gqlparser.SymbolToken{Content: "symbol"},
			&gqlparser.NumericToken{Int64: i1},
			&gqlparser.NumericToken{Int64: i2},
			&gqlparser.NumericToken{Int64: i3},
			&gqlparser.NumericToken{Floating: true, Float64: f1},
			&gqlparser.NumericToken{Floating: true, Float64: f2},
			&gqlparser.NumericToken{Floating: true, Float64: f3},
			&gqlparser.BooleanToken{Value: true},
			&gqlparser.BooleanToken{Value: false},
			&gqlparser.BindingToken{Index: 1},
			&gqlparser.BindingToken{Name: "foo"},
			&gqlparser.OrderToken{Descending: false},
			&gqlparser.OrderToken{Descending: true},
		}

		var seed [32]byte
		binary.BigEndian.PutUint32(seed[:], uint32seed)

		r := rand.New(rand.NewChaCha8(seed))
		var tokens []gqlparser.Token
		for len(tokens) < length {
			tokens = append(tokens, parts[r.IntN(len(parts))])
		}
		normalizeTokens(tokens)

		_, _ = gqlparser.ParseAggregationQuery(defaultTokenSourceFactory.NewSliceTokenSource(tokens))
		// should be no panics
	})
}

func FuzzParseQuery(f *testing.F) {
	f.Fuzz(func(t *testing.T, s1, s2, s3, s4 string, i1, i2, i3 int64, f1, f2, f3 float64, length int, uint32seed uint32) {
		parts := []gqlparser.Token{
			&gqlparser.KeywordToken{Name: "SELECT"},
			&gqlparser.KeywordToken{Name: "FROM"},
			&gqlparser.KeywordToken{Name: "WHERE"},
			&gqlparser.KeywordToken{Name: "DISTINCT"},
			&gqlparser.KeywordToken{Name: "ON"},
			&gqlparser.KeywordToken{Name: "ORDER"},
			&gqlparser.KeywordToken{Name: "BY"},
			&gqlparser.KeywordToken{Name: "LIMIT"},
			&gqlparser.KeywordToken{Name: "FIRST"},
			&gqlparser.KeywordToken{Name: "OFFSET"},
			&gqlparser.KeywordToken{Name: "KEY"},
			&gqlparser.KeywordToken{Name: "PROJECT"},
			&gqlparser.KeywordToken{Name: "NAMESPACE"},
			&gqlparser.KeywordToken{Name: "ARRAY"},
			&gqlparser.KeywordToken{Name: "BLOB"},
			&gqlparser.KeywordToken{Name: "DATETIME"},
			&gqlparser.KeywordToken{Name: "NULL"},
			&gqlparser.WildcardToken{},
			&gqlparser.WhitespaceToken{Content: " "},
			&gqlparser.WhitespaceToken{Content: "\n"},
			&gqlparser.WhitespaceToken{Content: "\t"},
			&gqlparser.OperatorToken{Type: "AND"},
			&gqlparser.OperatorToken{Type: "OR"},
			&gqlparser.OperatorToken{Type: "="},
			&gqlparser.OperatorToken{Type: "!="},
			&gqlparser.OperatorToken{Type: ">"},
			&gqlparser.OperatorToken{Type: ">="},
			&gqlparser.OperatorToken{Type: "<"},
			&gqlparser.OperatorToken{Type: "<="},
			&gqlparser.OperatorToken{Type: "IN"},
			&gqlparser.OperatorToken{Type: "NOT"},
			&gqlparser.OperatorToken{Type: "CONTAINS"},
			&gqlparser.OperatorToken{Type: "HAS"},
			&gqlparser.OperatorToken{Type: "ANCESTOR"},
			&gqlparser.OperatorToken{Type: "DESCENDANT"},
			&gqlparser.OperatorToken{Type: "("},
			&gqlparser.OperatorToken{Type: ")"},
			&gqlparser.OperatorToken{Type: ","},
			&gqlparser.StringToken{Quote: '"', Content: s1},
			&gqlparser.StringToken{Quote: '\'', Content: s2},
			&gqlparser.StringToken{Quote: '`', Content: s3},
			&gqlparser.SymbolToken{Content: "symbol"},
			&gqlparser.NumericToken{Int64: i1},
			&gqlparser.NumericToken{Int64: i2},
			&gqlparser.NumericToken{Int64: i3},
			&gqlparser.NumericToken{Floating: true, Float64: f1},
			&gqlparser.NumericToken{Floating: true, Float64: f2},
			&gqlparser.NumericToken{Floating: true, Float64: f3},
			&gqlparser.BooleanToken{Value: true},
			&gqlparser.BooleanToken{Value: false},
			&gqlparser.BindingToken{Index: 1},
			&gqlparser.BindingToken{Name: "foo"},
			&gqlparser.OrderToken{Descending: false},
			&gqlparser.OrderToken{Descending: true},
		}

		var seed [32]byte
		binary.BigEndian.PutUint32(seed[:], uint32seed)

		r := rand.New(rand.NewChaCha8(seed))
		var tokens []gqlparser.Token
		for len(tokens) < length {
			tokens = append(tokens, parts[r.IntN(len(parts))])
		}
		normalizeTokens(tokens)

		_, _ = gqlparser.ParseQuery(defaultTokenSourceFactory.NewSliceTokenSource(tokens))
		// should be no panics
	})
}

func FuzzParseCondition(f *testing.F) {
	f.Fuzz(func(t *testing.T, s1, s2, s3, s4 string, i1, i2, i3 int64, f1, f2, f3 float64, length int, uint32seed uint32) {
		parts := []gqlparser.Token{
			&gqlparser.KeywordToken{Name: "KEY"},
			&gqlparser.KeywordToken{Name: "PROJECT"},
			&gqlparser.KeywordToken{Name: "NAMESPACE"},
			&gqlparser.KeywordToken{Name: "ARRAY"},
			&gqlparser.KeywordToken{Name: "BLOB"},
			&gqlparser.KeywordToken{Name: "DATETIME"},
			&gqlparser.KeywordToken{Name: "NULL"},
			&gqlparser.WhitespaceToken{Content: " "},
			&gqlparser.WhitespaceToken{Content: "\n"},
			&gqlparser.WhitespaceToken{Content: "\t"},
			&gqlparser.OperatorToken{Type: "AND"},
			&gqlparser.OperatorToken{Type: "OR"},
			&gqlparser.OperatorToken{Type: "="},
			&gqlparser.OperatorToken{Type: "!="},
			&gqlparser.OperatorToken{Type: ">"},
			&gqlparser.OperatorToken{Type: ">="},
			&gqlparser.OperatorToken{Type: "<"},
			&gqlparser.OperatorToken{Type: "<="},
			&gqlparser.OperatorToken{Type: "IN"},
			&gqlparser.OperatorToken{Type: "NOT"},
			&gqlparser.OperatorToken{Type: "CONTAINS"},
			&gqlparser.OperatorToken{Type: "HAS"},
			&gqlparser.OperatorToken{Type: "ANCESTOR"},
			&gqlparser.OperatorToken{Type: "DESCENDANT"},
			&gqlparser.OperatorToken{Type: "("},
			&gqlparser.OperatorToken{Type: ")"},
			&gqlparser.OperatorToken{Type: ","},
			&gqlparser.StringToken{Quote: '"', Content: s1},
			&gqlparser.StringToken{Quote: '\'', Content: s2},
			&gqlparser.StringToken{Quote: '`', Content: s3},
			&gqlparser.NumericToken{Int64: i1},
			&gqlparser.NumericToken{Int64: i2},
			&gqlparser.NumericToken{Int64: i3},
			&gqlparser.NumericToken{Floating: true, Float64: f1},
			&gqlparser.NumericToken{Floating: true, Float64: f2},
			&gqlparser.NumericToken{Floating: true, Float64: f3},
			&gqlparser.BooleanToken{Value: true},
			&gqlparser.BooleanToken{Value: false},
			&gqlparser.BindingToken{Index: 1},
			&gqlparser.BindingToken{Name: "foo"},
		}

		var seed [32]byte
		binary.BigEndian.PutUint32(seed[:], uint32seed)

		r := rand.New(rand.NewChaCha8(seed))
		var tokens []gqlparser.Token
		for len(tokens) < length {
			tokens = append(tokens, parts[r.IntN(len(parts))])
		}
		normalizeTokens(tokens)

		_, _ = gqlparser.ParseCondition(defaultTokenSourceFactory.NewSliceTokenSource(tokens))
		// should be no panics
	})
}

func FuzzParseKey(f *testing.F) {
	f.Fuzz(func(t *testing.T, kind, name string, id int64, length int, uint32seed uint32) {
		parts := []gqlparser.Token{
			&gqlparser.KeywordToken{Name: "KEY"},
			&gqlparser.KeywordToken{Name: "PROJECT"},
			&gqlparser.KeywordToken{Name: "NAMESPACE"},
			&gqlparser.WhitespaceToken{Content: " "},
			&gqlparser.WhitespaceToken{Content: "\n"},
			&gqlparser.WhitespaceToken{Content: "\t"},
			&gqlparser.OperatorToken{Type: "("},
			&gqlparser.OperatorToken{Type: ")"},
			&gqlparser.OperatorToken{Type: ","},
			&gqlparser.SymbolToken{Content: "SymbolKind"},
			&gqlparser.StringToken{Quote: '`', Content: kind},
			&gqlparser.StringToken{Quote: '"', Content: name},
			&gqlparser.NumericToken{Int64: id},
		}

		var seed [32]byte
		binary.BigEndian.PutUint32(seed[:], uint32seed)

		r := rand.New(rand.NewChaCha8(seed))
		var tokens []gqlparser.Token
		for len(tokens) < length {
			tokens = append(tokens, parts[r.IntN(len(parts))])
		}
		normalizeTokens(tokens)

		_, _ = gqlparser.ParseKey(defaultTokenSourceFactory.NewSliceTokenSource(tokens))
		// should be no panics
	})
}
