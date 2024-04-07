package gqlparser

type tokenReader interface {
	Next() bool
	Read() (Token, error)
}

type tokenHistory struct {
	tokens []Token
}

type resettableTokenReader struct {
	source  TokenSource
	history *tokenHistory
	offset  int
}

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
