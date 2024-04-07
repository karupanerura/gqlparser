package gqlparser_test

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/gqlparser"
)

type sliceTokenSource struct {
	s []gqlparser.Token
}

func (ts *sliceTokenSource) Read() (gqlparser.Token, error) {
	if len(ts.s) == 0 {
		return nil, gqlparser.ErrEndOfToken
	}

	tok := ts.s[0]
	ts.s = ts.s[1:]
	return tok, nil
}

func (ts *sliceTokenSource) Unread(tok gqlparser.Token) {
	ts.s = append([]gqlparser.Token{tok}, ts.s...)
}

func (ts *sliceTokenSource) Next() bool {
	return len(ts.s) != 0
}

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
				Properties: []gqlparser.Property{"a"},
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
				Properties: []gqlparser.Property{"a", "b", "c"},
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
				Properties: []gqlparser.Property{"a"},
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
				Properties: []gqlparser.Property{"a", "b", "c"},
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
				Properties: []gqlparser.Property{"a", "b", "c", "d"},
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
				DistinctOn: []gqlparser.Property{"a", "b"},
				Properties: []gqlparser.Property{"c", "d"},
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
				Where: &gqlparser.EitherComparatorCondition{Comparator: "=", Property: "col", Value: int64(1)},
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
				Where: &gqlparser.EitherComparatorCondition{Comparator: "=", Property: "col", Value: &gqlparser.IndexedBinding{Index: 1}},
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
				Where: &gqlparser.EitherComparatorCondition{Comparator: "=", Property: "col", Value: int64(1)},
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
				Properties: []gqlparser.Property{"c", "d"},
				DistinctOn: []gqlparser.Property{"a", "b"},
				Kind:       "Kind",
				Where: &gqlparser.OrCompoundCondition{
					Left: &gqlparser.BackwardComparatorCondition{Comparator: "HAS DESCENDANT", Property: "__key__", Value: &gqlparser.Key{Path: []*gqlparser.KeyPath{{Kind: "Kind", ID: 1}}}},
					Right: &gqlparser.AndCompoundCondition{
						Left: &gqlparser.ForwardComparatorCondition{Comparator: "HAS ANCESTOR", Property: "__key__", Value: &gqlparser.Key{Path: []*gqlparser.KeyPath{{Kind: "Kind", Name: "key1"}}}},
						Right: &gqlparser.OrCompoundCondition{
							Left:  &gqlparser.EitherComparatorCondition{Comparator: "=", Property: "col", Value: int64(1)},
							Right: &gqlparser.IsNullCondition{Property: "col"},
						},
					},
				},
				OrderBy: []gqlparser.OrderBy{
					{Property: "foo", Descending: true},
					{Property: "bar"},
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

			got, err := gqlparser.ParseQuery(&sliceTokenSource{tt.tokens})
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
