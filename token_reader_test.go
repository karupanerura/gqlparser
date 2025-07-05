package gqlparser

import (
	"errors"
	"testing"
)

// emptyPureTokenReader implements tokenReader but is neither resettable nor TokenSource
type emptyPureTokenReader struct{}

func (m *emptyPureTokenReader) Next() bool           { return false }
func (m *emptyPureTokenReader) Read() (Token, error) { return nil, ErrEndOfToken }

func TestAsResettableTokenReader_WithTokenSource(t *testing.T) {
	// Create test tokens
	tokens := []Token{
		&StringToken{Content: "hello", RawContent: "\"hello\"", Position: 0},
		&OperatorToken{Type: "=", RawContent: "=", Position: 6},
		&StringToken{Content: "world", RawContent: "\"world\"", Position: 8},
	}

	// Create a token source
	source := defaultTokenSourceFactory.NewSliceTokenSource(tokens)

	// Wrap it as resettable
	reader := asResettableTokenReader(source)

	// Verify it's a resettableTokenReader
	if reader.source != source {
		t.Errorf("Expected source to be set correctly")
	}
	if reader.history == nil {
		t.Errorf("Expected history to be initialized")
	}
	if reader.offset != 0 {
		t.Errorf("Expected offset to be 0, got %d", reader.offset)
	}
}

func TestAsResettableTokenReader_WithResettableTokenReader(t *testing.T) {
	// Create test tokens
	tokens := []Token{
		&StringToken{Content: "hello", RawContent: "\"hello\"", Position: 0},
		&OperatorToken{Type: "=", RawContent: "=", Position: 6},
	}

	// Create a token source and wrap it once
	source := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
	firstReader := asResettableTokenReader(source)

	// Read one token to add to history
	firstReader.Read()

	// Wrap it again
	secondReader := asResettableTokenReader(firstReader)

	// Verify the second reader shares history but has different offset
	if secondReader.source != source {
		t.Errorf("Expected source to be the original source")
	}
	if secondReader.history != firstReader.history {
		t.Errorf("Expected history to be shared")
	}
	if secondReader.offset != len(firstReader.history.tokens) {
		t.Errorf("Expected offset to be %d, got %d", len(firstReader.history.tokens), secondReader.offset)
	}
}

func TestAsResettableTokenReader_WithUnsupportedType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for unsupported token reader type")
		}
	}()

	// Create a mock token reader that's neither resettable nor TokenSource
	mockReader := &emptyPureTokenReader{}
	asResettableTokenReader(mockReader)
}

func TestResettableTokenReader_Next(t *testing.T) {
	tokens := []Token{
		&StringToken{Content: "test", RawContent: "\"test\"", Position: 0},
	}

	source := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
	reader := asResettableTokenReader(source)

	// Should have next token
	if !reader.Next() {
		t.Errorf("Expected Next() to return true")
	}

	// Read the token
	reader.Read()

	// Should not have next token
	if reader.Next() {
		t.Errorf("Expected Next() to return false after reading all tokens")
	}
}

func TestResettableTokenReader_Read(t *testing.T) {
	tokens := []Token{
		&StringToken{Content: "hello", RawContent: "\"hello\"", Position: 0},
		&OperatorToken{Type: "=", RawContent: "=", Position: 6},
		&StringToken{Content: "world", RawContent: "\"world\"", Position: 8},
	}

	source := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
	reader := asResettableTokenReader(source)

	// Read first token
	token1, err := reader.Read()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if token1.GetContent() != "\"hello\"" {
		t.Errorf("Expected first token content to be '\"hello\"', got '%s'", token1.GetContent())
	}

	// Read second token
	token2, err := reader.Read()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if token2.GetContent() != "=" {
		t.Errorf("Expected second token content to be '=', got '%s'", token2.GetContent())
	}

	// Read third token
	token3, err := reader.Read()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if token3.GetContent() != "\"world\"" {
		t.Errorf("Expected third token content to be '\"world\"', got '%s'", token3.GetContent())
	}

	// Verify history was recorded
	if len(reader.history.tokens) != 3 {
		t.Errorf("Expected history to contain 3 tokens, got %d", len(reader.history.tokens))
	}
}

func TestResettableTokenReader_ReadError(t *testing.T) {
	testError := errors.New("test error")
	tokens := []Token{
		&StringToken{Content: "test", RawContent: "\"test\"", Position: 0},
	}

	source := defaultTokenSourceFactory.NewErrorTokenSource(tokens, map[int]error{0: testError})
	reader := asResettableTokenReader(source)

	// Should return error on first read
	_, err := reader.Read()
	if !errors.Is(err, testError) {
		t.Errorf("Expected error to be %v, got %v", testError, err)
	}

	// History should not be updated on error
	if len(reader.history.tokens) != 0 {
		t.Errorf("Expected history to be empty on error, got %d tokens", len(reader.history.tokens))
	}
}

func TestResettableTokenReader_Reset(t *testing.T) {
	tokens := []Token{
		&StringToken{Content: "hello", RawContent: "\"hello\"", Position: 0},
		&OperatorToken{Type: "=", RawContent: "=", Position: 6},
		&StringToken{Content: "world", RawContent: "\"world\"", Position: 8},
	}

	source := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
	reader := asResettableTokenReader(source)

	// Read all tokens
	token1, _ := reader.Read()
	token2, _ := reader.Read()
	token3, _ := reader.Read()

	// Verify we have 3 tokens in history
	if len(reader.history.tokens) != 3 {
		t.Errorf("Expected 3 tokens in history, got %d", len(reader.history.tokens))
	}

	// Reset should put all tokens back to source
	reader.Reset()

	// History should be empty (trimmed to offset 0)
	if len(reader.history.tokens) != 0 {
		t.Errorf("Expected history to be empty after reset, got %d tokens", len(reader.history.tokens))
	}

	// Should be able to read tokens again in same order
	newToken1, err := reader.Read()
	if err != nil {
		t.Errorf("Unexpected error after reset: %v", err)
	}
	if newToken1.GetContent() != token1.GetContent() {
		t.Errorf("Expected token after reset to match original, got '%s' vs '%s'",
			newToken1.GetContent(), token1.GetContent())
	}

	newToken2, err := reader.Read()
	if err != nil {
		t.Errorf("Unexpected error after reset: %v", err)
	}
	if newToken2.GetContent() != token2.GetContent() {
		t.Errorf("Expected token after reset to match original, got '%s' vs '%s'",
			newToken2.GetContent(), token2.GetContent())
	}

	newToken3, err := reader.Read()
	if err != nil {
		t.Errorf("Unexpected error after reset: %v", err)
	}
	if newToken3.GetContent() != token3.GetContent() {
		t.Errorf("Expected token after reset to match original, got '%s' vs '%s'",
			newToken3.GetContent(), token3.GetContent())
	}
}

func TestResettableTokenReader_ResetWithOffset(t *testing.T) {
	tokens := []Token{
		&StringToken{Content: "hello", RawContent: "\"hello\"", Position: 0},
		&OperatorToken{Type: "=", RawContent: "=", Position: 6},
		&StringToken{Content: "world", RawContent: "\"world\"", Position: 8},
		&OperatorToken{Type: "!", RawContent: "!", Position: 14},
	}

	source := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
	firstReader := asResettableTokenReader(source)

	// Read first two tokens with first reader
	firstReader.Read()
	firstReader.Read()

	// Create second reader (should have offset = 2)
	secondReader := asResettableTokenReader(firstReader)

	// Read remaining tokens with second reader
	token3, _ := secondReader.Read()
	token4, _ := secondReader.Read()

	// Reset second reader - should only reset tokens read after offset
	secondReader.Reset()

	// History should be trimmed to offset (2)
	if len(secondReader.history.tokens) != 2 {
		t.Errorf("Expected history to have 2 tokens after reset, got %d", len(secondReader.history.tokens))
	}

	// Should be able to read the last two tokens again
	newToken3, err := secondReader.Read()
	if err != nil {
		t.Errorf("Unexpected error after reset: %v", err)
	}
	if newToken3.GetContent() != token3.GetContent() {
		t.Errorf("Expected token after reset to match original, got '%s' vs '%s'",
			newToken3.GetContent(), token3.GetContent())
	}

	newToken4, err := secondReader.Read()
	if err != nil {
		t.Errorf("Unexpected error after reset: %v", err)
	}
	if newToken4.GetContent() != token4.GetContent() {
		t.Errorf("Expected token after reset to match original, got '%s' vs '%s'",
			newToken4.GetContent(), token4.GetContent())
	}
}

func TestResettableTokenReader_EmptySource(t *testing.T) {
	var tokens []Token
	source := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
	reader := asResettableTokenReader(source)

	// Next should return false for empty source
	if reader.Next() {
		t.Errorf("Expected Next() to return false for empty source")
	}

	// Read should return end of token error
	_, err := reader.Read()
	if !errors.Is(err, ErrEndOfToken) {
		t.Errorf("Expected ErrEndOfToken for empty source, got %v", err)
	}

	// Reset should work even with empty source
	reader.Reset()
	if len(reader.history.tokens) != 0 {
		t.Errorf("Expected empty history after reset, got %d tokens", len(reader.history.tokens))
	}
}

func TestResettableTokenReader_MultipleResets(t *testing.T) {
	tokens := []Token{
		&StringToken{Content: "test1", RawContent: "\"test1\"", Position: 0},
		&StringToken{Content: "test2", RawContent: "\"test2\"", Position: 7},
	}

	source := defaultTokenSourceFactory.NewSliceTokenSource(tokens)
	reader := asResettableTokenReader(source)

	// Read first token
	token1, _ := reader.Read()

	// Reset
	reader.Reset()

	// Read first token again
	newToken1, _ := reader.Read()
	if newToken1.GetContent() != token1.GetContent() {
		t.Errorf("Expected token to match after first reset")
	}

	// Read second token
	token2, _ := reader.Read()

	// Reset again
	reader.Reset()

	// Should be able to read both tokens again
	againToken1, _ := reader.Read()
	againToken2, _ := reader.Read()

	if againToken1.GetContent() != token1.GetContent() {
		t.Errorf("Expected first token to match after second reset")
	}
	if againToken2.GetContent() != token2.GetContent() {
		t.Errorf("Expected second token to match after second reset")
	}
}
