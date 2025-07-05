package gqlparser

// tokenReader is an internal interface used by the parser to read tokens.
// Next reports whether there are more tokens available.
// Read returns the next Token and any error encountered.
type tokenReader interface {
	Next() bool
	Read() (Token, error)
}

type tokenHistory struct {
	tokens []Token
}

// resettableTokenReader wraps a TokenSource to support lookahead and backtracking.
// It records all tokens read from the underlying source and can reset the source
// back to a previously marked offset, effectively "unreading" any tokens read since.
type resettableTokenReader struct {
	source  TokenSource
	history *tokenHistory
	offset  int
}

// asResettableTokenReader wraps a tokenReader into a resettableTokenReader,
// enabling lookahead and backtracking by recording tokens as they are read.
// If tr is already a resettableTokenReader, it returns a new reader with its
// offset set to the current history length, so subsequent reads are recorded
// from that point. If tr implements TokenSource, it wraps it with an empty history.
// Panics if tr is neither a resettableTokenReader nor TokenSource.
func asResettableTokenReader(tr tokenReader) *resettableTokenReader {
	switch v := tr.(type) {
	case *resettableTokenReader:
		return &resettableTokenReader{source: v.source, history: v.history, offset: len(v.history.tokens)}
	case TokenSource:
		return &resettableTokenReader{source: v, history: &tokenHistory{}}
	default:
		panic("unknown token reader")
	}
}

func (tr *resettableTokenReader) Next() bool {
	return tr.source.Next()
}

func (tr *resettableTokenReader) Read() (Token, error) {
	token, err := tr.source.Read()
	if err != nil {
		return nil, err
	}

	tr.history.tokens = append(tr.history.tokens, token)
	return token, nil
}

func (tr *resettableTokenReader) Reset() {
	for i := len(tr.history.tokens) - 1; i >= tr.offset; i-- {
		tr.source.Unread(tr.history.tokens[i])
	}
	tr.history.tokens = tr.history.tokens[:tr.offset]
}
