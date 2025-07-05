package gqlparser

// TokenSource is an interface that defines methods for reading tokens from a source.
// It provides methods to check if there are more tokens, read the next token,
// and unread a token to allow for backtracking.
type TokenSource interface {
	// Next returns true if there are more tokens to read.
	// It should be called before calling Read() to check if there are tokens available.
	// It returns false if there are no more tokens.
	Next() bool

	// Read reads the next token from the source.
	// It returns the token and an error if there was a problem reading the token.
	Read() (Token, error)

	// Unread un-reads the last read token.
	// This allows for backtracking in the token stream.
	// It should be called after Read() if you want to go back to the previous token.
	// It returns an error if there was a problem un-reading the token.
	Unread(Token)
}

// ReadAllTokens reads all tokens from the provided TokenSource.
// It continues reading until there are no more tokens available.
func ReadAllTokens(ts TokenSource) ([]Token, error) {
	var tokens []Token
	for ts.Next() {
		tok, err := ts.Read()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}
