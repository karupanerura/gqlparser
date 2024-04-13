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
