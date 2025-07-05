package gqlparser_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/gqlparser"
)

func TestReadAllTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		tokens []gqlparser.Token
	}{
		{"Empty", nil},
		{"Single", []gqlparser.Token{
			&gqlparser.SymbolToken{Content: "foo"},
		}},
		{"Multiple", []gqlparser.Token{
			&gqlparser.SymbolToken{Content: "a"},
			&gqlparser.WhitespaceToken{Content: " "},
			&gqlparser.SymbolToken{Content: "b"},
		}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := defaultTokenSourceFactory.NewSliceTokenSource(append([]gqlparser.Token(nil), tt.tokens...))
			got, err := gqlparser.ReadAllTokens(ts)
			if err != nil {
				t.Fatalf("ReadAllTokens error: %v", err)
			}
			if df := cmp.Diff(tt.tokens, got); df != "" {
				t.Errorf("ReadAllTokens mismatch (-want +got):\n%s", df)
			}
		})
	}
}

func TestReadAllTokensError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tokenSource gqlparser.TokenSource
		wantErr     bool
		expectedErr error
	}{
		{
			name: "ErrorOnFirstRead",
			tokenSource: defaultTokenSourceFactory.NewErrorTokenSource(
				[]gqlparser.Token{
					&gqlparser.SymbolToken{Content: "foo"},
				},
				map[int]error{0: errors.New("read error")},
			),
			wantErr:     true,
			expectedErr: errors.New("read error"),
		},
		{
			name: "ErrorAfterSomeTokens",
			tokenSource: defaultTokenSourceFactory.NewErrorTokenSource(
				[]gqlparser.Token{
					&gqlparser.SymbolToken{Content: "a"},
					&gqlparser.WhitespaceToken{Content: " "},
				},
				map[int]error{0: errors.New("network timeout")},
			),
			wantErr:     true,
			expectedErr: errors.New("network timeout"),
		},
		{
			name: "EmptySourceWithError",
			tokenSource: defaultTokenSourceFactory.NewErrorTokenSource(
				[]gqlparser.Token{},
				map[int]error{0: errors.New("empty source error")},
			),
			wantErr:     true,
			expectedErr: errors.New("empty source error"),
		},
		{
			name: "IOError",
			tokenSource: defaultTokenSourceFactory.NewErrorTokenSource(
				[]gqlparser.Token{
					&gqlparser.SymbolToken{Content: "start"},
				},
				map[int]error{0: errors.New("I/O error")},
			),
			wantErr:     true,
			expectedErr: errors.New("I/O error"),
		},
		{
			name: "UnexpectedError",
			tokenSource: defaultTokenSourceFactory.NewErrorTokenSource(
				[]gqlparser.Token{},
				map[int]error{0: errors.New("unexpected internal error")},
			),
			wantErr:     true,
			expectedErr: errors.New("unexpected internal error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := gqlparser.ReadAllTokens(tt.tokenSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadAllTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.expectedErr != nil {
				if err.Error() != tt.expectedErr.Error() {
					t.Errorf("ReadAllTokens() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}
			if !tt.wantErr && got != nil {
				t.Errorf("ReadAllTokens() should return nil tokens when error expected")
			}
		})
	}
}
