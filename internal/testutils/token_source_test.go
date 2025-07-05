package testutils_test

import (
	"errors"
	"testing"

	"github.com/karupanerura/gqlparser"
	"github.com/karupanerura/gqlparser/internal/testutils"
)

// Mock token type for testing
type mockToken struct {
	content  string
	position int
}

func (m *mockToken) GetContent() string { return m.content }
func (m *mockToken) GetPosition() int   { return m.position }

// Test errors
var (
	testEndOfTokenError = errors.New("test end of token")
	testFirstReadError  = errors.New("test first read error")
	testCustomError     = errors.New("test custom error")
)

// TokenSource interface for testing (matches the gqlparser.TokenSource interface)
type TokenSource[T any] interface {
	Next() bool
	Read() (T, error)
	Unread(T)
}

func TestNewTestTokenSourceFactory(t *testing.T) {
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testEndOfTokenError)

	if factory == nil {
		t.Fatal("NewTestTokenSourceFactory returned nil")
	}

	// Test that factory can create different types of token sources
	tokens := []*mockToken{
		{content: "test1", position: 0},
		{content: "test2", position: 5},
	}

	// Test slice token source creation
	sliceSource := factory.NewSliceTokenSource(tokens)
	if sliceSource == nil {
		t.Error("NewSliceTokenSource returned nil")
	}

	// Test error token source creation
	errorSource := factory.NewErrorTokenSource(tokens, map[int]error{0: testFirstReadError})
	if errorSource == nil {
		t.Error("NewErrorTokenSource returned nil")
	}

	// Test error token source with end error creation
	endErrorSource := factory.NewErrorTokenSource(tokens, nil)
	if endErrorSource == nil {
		t.Error("NewErrorTokenSource returned nil")
	}
}

func TestSliceTokenSource(t *testing.T) {
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testEndOfTokenError)

	tests := []struct {
		name          string
		tokens        []*mockToken
		expectedReads int
	}{
		{
			name:          "empty token source",
			tokens:        []*mockToken{},
			expectedReads: 0,
		},
		{
			name: "single token",
			tokens: []*mockToken{
				{content: "single", position: 0},
			},
			expectedReads: 1,
		},
		{
			name: "multiple tokens",
			tokens: []*mockToken{
				{content: "first", position: 0},
				{content: "second", position: 5},
				{content: "third", position: 12},
			},
			expectedReads: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := factory.NewSliceTokenSource(tt.tokens)

			// Test Next() method
			for i := 0; i < tt.expectedReads; i++ {
				if !source.Next() {
					t.Errorf("Expected Next() to return true for token %d", i)
				}

				token, err := source.Read()
				if err != nil {
					t.Errorf("Unexpected error reading token %d: %v", i, err)
				}

				if token.GetContent() != tt.tokens[i].content {
					t.Errorf("Expected token content %s, got %s", tt.tokens[i].content, token.GetContent())
				}

				if token.GetPosition() != tt.tokens[i].position {
					t.Errorf("Expected token position %d, got %d", tt.tokens[i].position, token.GetPosition())
				}
			}

			// Test that Next() returns false after all tokens are read
			if source.Next() {
				t.Error("Expected Next() to return false after all tokens are read")
			}

			// Test that Read() returns error after all tokens are read
			_, err := source.Read()
			if !errors.Is(err, testEndOfTokenError) {
				t.Errorf("Expected end of token error, got %v", err)
			}
		})
	}
}

func TestSliceTokenSourceUnread(t *testing.T) {
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testEndOfTokenError)
	tokens := []*mockToken{
		{content: "first", position: 0},
		{content: "second", position: 5},
	}

	source := factory.NewSliceTokenSource(tokens)

	// Read first token
	if !source.Next() {
		t.Fatal("Expected Next() to return true")
	}

	token1, err := source.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Read second token
	if !source.Next() {
		t.Fatal("Expected Next() to return true")
	}

	token2, err := source.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Unread the second token
	source.Unread(token2)

	// Should be able to read the second token again
	if !source.Next() {
		t.Error("Expected Next() to return true after unread")
	}

	token2Again, err := source.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token2Again.GetContent() != token2.GetContent() {
		t.Errorf("Expected unread token content %s, got %s", token2.GetContent(), token2Again.GetContent())
	}

	// Unread both tokens
	source.Unread(token2Again)
	source.Unread(token1)

	// Should be able to read from the beginning again
	if !source.Next() {
		t.Error("Expected Next() to return true after unread")
	}

	firstAgain, err := source.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if firstAgain.GetContent() != token1.GetContent() {
		t.Errorf("Expected first token content %s, got %s", token1.GetContent(), firstAgain.GetContent())
	}
}

func TestErrorTokenSource(t *testing.T) {
	// Test with default factory
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testEndOfTokenError)

	// Test with custom error factory
	customFactory := testutils.NewTestTokenSourceFactory[*mockToken](testCustomError)

	tests := []struct {
		name             string
		factory          *testutils.TestTokenSourceFactory[*mockToken]
		tokens           []*mockToken
		errorMap         map[int]error
		expectedBehavior func(t *testing.T, source TokenSource[*mockToken])
	}{
		{
			name:     "error on first read",
			factory:  factory,
			tokens:   []*mockToken{{content: "test", position: 0}},
			errorMap: map[int]error{0: testFirstReadError},
			expectedBehavior: func(t *testing.T, source TokenSource[*mockToken]) {
				// Should return true initially to allow Read() to be called
				if !source.Next() {
					t.Error("Expected Next() to return true initially")
				}

				// Should return the error on Read()
				_, err := source.Read()
				if !errors.Is(err, testFirstReadError) {
					t.Errorf("Expected first read error, got %v", err)
				}

				// Should return false after error
				if source.Next() {
					t.Error("Expected Next() to return false after error")
				}
			},
		},
		{
			name:     "error at end of tokens",
			factory:  customFactory,
			tokens:   []*mockToken{{content: "test", position: 0}},
			errorMap: nil,
			expectedBehavior: func(t *testing.T, source TokenSource[*mockToken]) {
				// Should be able to read the token
				if !source.Next() {
					t.Error("Expected Next() to return true")
				}

				token, err := source.Read()
				if err != nil {
					t.Errorf("Unexpected error reading token: %v", err)
				}

				if token.GetContent() != "test" {
					t.Errorf("Expected token content 'test', got %s", token.GetContent())
				}

				// Should return false after reading all tokens
				if source.Next() {
					t.Error("Expected Next() to return false after reading all tokens")
				}

				// Should return custom error when trying to read past end
				_, err = source.Read()
				if !errors.Is(err, testCustomError) {
					t.Errorf("Expected custom end error, got %v", err)
				}
			},
		},
		{
			name:     "no tokens with end error",
			factory:  customFactory,
			tokens:   []*mockToken{},
			errorMap: nil,
			expectedBehavior: func(t *testing.T, source TokenSource[*mockToken]) {
				// Should return false for empty source
				if source.Next() {
					t.Error("Expected Next() to return false for empty source")
				}

				// Should return custom error immediately
				_, err := source.Read()
				if !errors.Is(err, testCustomError) {
					t.Errorf("Expected custom end error, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.factory.NewErrorTokenSource(tt.tokens, tt.errorMap)
			tt.expectedBehavior(t, source)
		})
	}
}

func TestErrorTokenSourceWithEndError(t *testing.T) {
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testCustomError)
	tokens := []*mockToken{
		{content: "first", position: 0},
		{content: "second", position: 5},
	}

	source := factory.NewErrorTokenSource(tokens, nil)

	// Should be able to read all tokens normally
	for i, expectedToken := range tokens {
		if !source.Next() {
			t.Errorf("Expected Next() to return true for token %d", i)
		}

		token, err := source.Read()
		if err != nil {
			t.Errorf("Unexpected error reading token %d: %v", i, err)
		}

		if token.GetContent() != expectedToken.content {
			t.Errorf("Expected token content %s, got %s", expectedToken.content, token.GetContent())
		}
	}

	// Should return false after all tokens are read
	if source.Next() {
		t.Error("Expected Next() to return false after all tokens are read")
	}

	// Should return custom error when trying to read past end
	_, err := source.Read()
	if !errors.Is(err, testCustomError) {
		t.Errorf("Expected custom end error, got %v", err)
	}
}

func TestErrorTokenSourceUnread(t *testing.T) {
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testCustomError)
	tokens := []*mockToken{
		{content: "first", position: 0},
		{content: "second", position: 5},
	}

	source := factory.NewErrorTokenSource(tokens, nil)

	// Read first token
	if !source.Next() {
		t.Fatal("Expected Next() to return true")
	}

	token1, err := source.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Read second token
	if !source.Next() {
		t.Fatal("Expected Next() to return true")
	}

	token2, err := source.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Unread the second token
	source.Unread(token2)

	// Should be able to read the second token again
	if !source.Next() {
		t.Error("Expected Next() to return true after unread")
	}

	token2Again, err := source.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token2Again.GetContent() != token2.GetContent() {
		t.Errorf("Expected unread token content %s, got %s", token2.GetContent(), token2Again.GetContent())
	}

	// Unread first token (should handle position correctly)
	source.Unread(token1)

	// Should be able to read from the beginning again
	if !source.Next() {
		t.Error("Expected Next() to return true after unread")
	}

	firstAgain, err := source.Read()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if firstAgain.GetContent() != token1.GetContent() {
		t.Errorf("Expected first token content %s, got %s", token1.GetContent(), firstAgain.GetContent())
	}
}

func TestFactoryWithDifferentTokenTypes(t *testing.T) {
	// Test with gqlparser.Token interface
	gqlFactory := testutils.NewTestTokenSourceFactory[gqlparser.Token](gqlparser.ErrEndOfToken)

	gqlTokens := []gqlparser.Token{
		&gqlparser.StringToken{Content: "hello", RawContent: `"hello"`, Position: 0},
		&gqlparser.OperatorToken{Type: "=", RawContent: "=", Position: 7},
	}

	gqlSource := gqlFactory.NewSliceTokenSource(gqlTokens)

	// Test reading gql tokens
	if !gqlSource.Next() {
		t.Error("Expected Next() to return true for gql tokens")
	}

	token, err := gqlSource.Read()
	if err != nil {
		t.Errorf("Unexpected error reading gql token: %v", err)
	}

	if token.GetContent() != `"hello"` {
		t.Errorf("Expected gql token content '\"hello\"', got %s", token.GetContent())
	}

	// Test with custom mock token type
	mockFactory := testutils.NewTestTokenSourceFactory[*mockToken](testEndOfTokenError)

	mockTokens := []*mockToken{
		{content: "mock1", position: 0},
		{content: "mock2", position: 5},
	}

	mockSource := mockFactory.NewSliceTokenSource(mockTokens)

	// Test reading mock tokens
	if !mockSource.Next() {
		t.Error("Expected Next() to return true for mock tokens")
	}

	mockToken, err := mockSource.Read()
	if err != nil {
		t.Errorf("Unexpected error reading mock token: %v", err)
	}

	if mockToken.GetContent() != "mock1" {
		t.Errorf("Expected mock token content 'mock1', got %s", mockToken.GetContent())
	}
}

func TestFactoryErrorHandling(t *testing.T) {
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testEndOfTokenError)

	// Test NewErrorTokenSource (should use factory's default end error)
	tokens := []*mockToken{{content: "test", position: 0}}
	source := factory.NewErrorTokenSource(tokens, nil)

	// Read the token
	if !source.Next() {
		t.Error("Expected Next() to return true")
	}

	_, err := source.Read()
	if err != nil {
		t.Errorf("Unexpected error reading token: %v", err)
	}

	// Should return factory's default error
	if source.Next() {
		t.Error("Expected Next() to return false after reading all tokens")
	}

	_, err = source.Read()
	if !errors.Is(err, testEndOfTokenError) {
		t.Errorf("Expected factory's default end error, got %v", err)
	}

	// Test NewErrorTokenSource (should use factory's default end error)
	source2 := factory.NewErrorTokenSource(tokens, nil)

	// Read the token
	if !source2.Next() {
		t.Error("Expected Next() to return true")
	}

	_, err = source2.Read()
	if err != nil {
		t.Errorf("Unexpected error reading token: %v", err)
	}

	// Should return factory's default error
	if source2.Next() {
		t.Error("Expected Next() to return false after reading all tokens")
	}

	_, err = source2.Read()
	if !errors.Is(err, testEndOfTokenError) {
		t.Errorf("Expected factory's default end error, got %v", err)
	}
}

func TestEdgeCases(t *testing.T) {
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testEndOfTokenError)

	t.Run("multiple unreads", func(t *testing.T) {
		tokens := []*mockToken{
			{content: "first", position: 0},
			{content: "second", position: 5},
		}

		source := factory.NewSliceTokenSource(tokens)

		// Read both tokens
		source.Next()
		token1, _ := source.Read()
		source.Next()
		token2, _ := source.Read()

		// Unread both in reverse order
		source.Unread(token2)
		source.Unread(token1)

		// Should be able to read from beginning
		if !source.Next() {
			t.Error("Expected Next() to return true after multiple unreads")
		}

		readToken, err := source.Read()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if readToken.GetContent() != "first" {
			t.Errorf("Expected first token after unreads, got %s", readToken.GetContent())
		}
	})

	t.Run("unread on empty source", func(t *testing.T) {
		source := factory.NewSliceTokenSource([]*mockToken{})

		// Should handle unread gracefully even with empty source
		dummyToken := &mockToken{content: "dummy", position: 0}
		source.Unread(dummyToken)

		if !source.Next() {
			t.Error("Expected Next() to return true after unread on empty source")
		}

		readToken, err := source.Read()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if readToken.GetContent() != "dummy" {
			t.Errorf("Expected dummy token, got %s", readToken.GetContent())
		}
	})

	t.Run("error source unread with position zero", func(t *testing.T) {
		factory := testutils.NewTestTokenSourceFactory[*mockToken](testCustomError)
		tokens := []*mockToken{{content: "test", position: 0}}
		source := factory.NewErrorTokenSource(tokens, nil)

		// Try to unread when position is 0
		dummyToken := &mockToken{content: "dummy", position: 0}
		source.Unread(dummyToken) // Should handle gracefully

		// Should still work normally
		if !source.Next() {
			t.Error("Expected Next() to return true")
		}

		token, err := source.Read()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if token.GetContent() != "test" {
			t.Errorf("Expected original token, got %s", token.GetContent())
		}
	})
}

func TestIntegrationWithRealGQLParserTypes(t *testing.T) {
	// Test integration with real gqlparser types
	factory := testutils.NewTestTokenSourceFactory[gqlparser.Token](gqlparser.ErrEndOfToken)

	// Create various token types
	tokens := []gqlparser.Token{
		&gqlparser.StringToken{
			Quote:      '"',
			Content:    "hello",
			RawContent: `"hello"`,
			Position:   0,
		},
		&gqlparser.OperatorToken{
			Type:       "=",
			RawContent: "=",
			Position:   7,
		},
		&gqlparser.NumericToken{
			Int64:      42,
			Floating:   false,
			RawContent: "42",
			Position:   9,
		},
		&gqlparser.BooleanToken{
			Value:      true,
			RawContent: "true",
			Position:   12,
		},
		&gqlparser.SymbolToken{
			Content:  "variable",
			Position: 17,
		},
		&gqlparser.KeywordToken{
			Name:       "SELECT",
			RawContent: "select",
			Position:   26,
		},
		&gqlparser.BindingToken{
			Index:    1,
			Position: 33,
		},
		&gqlparser.BindingToken{
			Name:     "param",
			Position: 36,
		},
		&gqlparser.WildcardToken{
			Position: 42,
		},
		&gqlparser.OrderToken{
			Descending: true,
			RawContent: "DESC",
			Position:   44,
		},
		&gqlparser.WhitespaceToken{
			Content:  " ",
			Position: 49,
		},
	}

	t.Run("slice source with real tokens", func(t *testing.T) {
		source := factory.NewSliceTokenSource(tokens)

		readTokens := []gqlparser.Token{}
		for source.Next() {
			token, err := source.Read()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			readTokens = append(readTokens, token)
		}

		if len(readTokens) != len(tokens) {
			t.Errorf("Expected %d tokens, got %d", len(tokens), len(readTokens))
		}

		for i, token := range readTokens {
			if token.GetContent() != tokens[i].GetContent() {
				t.Errorf("Token %d: expected content %s, got %s", i, tokens[i].GetContent(), token.GetContent())
			}
			if token.GetPosition() != tokens[i].GetPosition() {
				t.Errorf("Token %d: expected position %d, got %d", i, tokens[i].GetPosition(), token.GetPosition())
			}
		}
	})

	t.Run("error source with real tokens", func(t *testing.T) {
		customError := errors.New("custom read error")
		source := factory.NewErrorTokenSource(tokens[:3], map[int]error{0: customError})

		// Should error on first read
		if !source.Next() {
			t.Error("Expected Next() to return true")
		}

		_, err := source.Read()
		if !errors.Is(err, customError) {
			t.Errorf("Expected custom error, got %v", err)
		}
	})

	t.Run("error source with end error and real tokens", func(t *testing.T) {
		endError := errors.New("custom end error")
		customFactory := testutils.NewTestTokenSourceFactory[gqlparser.Token](endError)
		source := customFactory.NewErrorTokenSource(tokens[:2], nil)

		// Should read tokens normally
		for i := 0; i < 2; i++ {
			if !source.Next() {
				t.Errorf("Expected Next() to return true for token %d", i)
			}

			token, err := source.Read()
			if err != nil {
				t.Errorf("Unexpected error reading token %d: %v", i, err)
			}

			if token.GetContent() != tokens[i].GetContent() {
				t.Errorf("Token %d: expected content %s, got %s", i, tokens[i].GetContent(), token.GetContent())
			}
		}

		// Should return custom error at end
		if source.Next() {
			t.Error("Expected Next() to return false after reading all tokens")
		}

		_, err := source.Read()
		if !errors.Is(err, endError) {
			t.Errorf("Expected custom end error, got %v", err)
		}
	})
}

func TestUnreadBehaviorConsistency(t *testing.T) {
	factory := testutils.NewTestTokenSourceFactory[*mockToken](testEndOfTokenError)
	tokens := []*mockToken{
		{content: "first", position: 0},
		{content: "second", position: 5},
		{content: "third", position: 10},
	}

	t.Run("slice source unread consistency", func(t *testing.T) {
		source := factory.NewSliceTokenSource(tokens)

		// Read all tokens
		var readTokens []*mockToken
		for source.Next() {
			token, err := source.Read()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			readTokens = append(readTokens, token)
		}

		// Unread all tokens in reverse order
		for i := len(readTokens) - 1; i >= 0; i-- {
			source.Unread(readTokens[i])
		}

		// Read all tokens again and verify they're the same
		var readTokensAgain []*mockToken
		for source.Next() {
			token, err := source.Read()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			readTokensAgain = append(readTokensAgain, token)
		}

		if len(readTokens) != len(readTokensAgain) {
			t.Errorf("Expected %d tokens after unread, got %d", len(readTokens), len(readTokensAgain))
		}

		for i, token := range readTokensAgain {
			if token.GetContent() != readTokens[i].GetContent() {
				t.Errorf("Token %d after unread: expected content %s, got %s", i, readTokens[i].GetContent(), token.GetContent())
			}
		}
	})

	t.Run("error source unread consistency", func(t *testing.T) {
		customFactory := testutils.NewTestTokenSourceFactory[*mockToken](testCustomError)
		source := customFactory.NewErrorTokenSource(tokens, nil)

		// Read all tokens
		var readTokens []*mockToken
		for source.Next() {
			token, err := source.Read()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			readTokens = append(readTokens, token)
		}

		// Unread all tokens in reverse order
		for i := len(readTokens) - 1; i >= 0; i-- {
			source.Unread(readTokens[i])
		}

		// Read all tokens again and verify they're the same
		var readTokensAgain []*mockToken
		for source.Next() {
			token, err := source.Read()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			readTokensAgain = append(readTokensAgain, token)
		}

		if len(readTokens) != len(readTokensAgain) {
			t.Errorf("Expected %d tokens after unread, got %d", len(readTokens), len(readTokensAgain))
		}

		for i, token := range readTokensAgain {
			if token.GetContent() != readTokens[i].GetContent() {
				t.Errorf("Token %d after unread: expected content %s, got %s", i, readTokens[i].GetContent(), token.GetContent())
			}
		}
	})
}
