package gqlparser

import (
	"errors"
	"testing"
)

func TestTokenAcceptorFn(t *testing.T) {
	tests := []struct {
		name     string
		fn       tokenAcceptorFn
		tokens   []Token
		wantErr  error
		consumed int
	}{
		{
			name: "successful acceptance",
			fn: tokenAcceptorFn(func(tr tokenReader) error {
				_, err := tr.Read()
				return err
			}),
			tokens:   []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:  nil,
			consumed: 1,
		},
		{
			name: "failing acceptance",
			fn: tokenAcceptorFn(func(tr tokenReader) error {
				return ErrUnexpectedToken
			}),
			tokens:   []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:  ErrUnexpectedToken,
			consumed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)

			// Count initial tokens
			initialCount := len(tt.tokens)

			err := tt.fn.accept(tr)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}

			// Count remaining tokens
			remainingCount := 0
			for tr.Next() {
				_, _ = tr.Read()
				remainingCount++
			}

			actualConsumed := initialCount - remainingCount
			if actualConsumed != tt.consumed {
				t.Errorf("expected %d tokens consumed, got %d", tt.consumed, actualConsumed)
			}
		})
	}
}

func TestNopAcceptor(t *testing.T) {
	tests := []struct {
		name   string
		tokens []Token
	}{
		{
			name:   "empty tokens",
			tokens: []Token{},
		},
		{
			name:   "single token",
			tokens: []Token{&KeywordToken{Name: "test", Position: 0}},
		},
		{
			name: "multiple tokens",
			tokens: []Token{
				&KeywordToken{Name: "test1", Position: 0},
				&KeywordToken{Name: "test2", Position: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := nopAcceptor.accept(tr)

			if err != nil {
				t.Errorf("nopAcceptor should always return nil, got %v", err)
			}
		})
	}
}

func TestNotAcceptor(t *testing.T) {
	successAcceptor := tokenAcceptorFn(func(tr tokenReader) error {
		// A realistic success acceptor that actually reads a token
		_, err := tr.Read()
		return err
	})

	failAcceptor := tokenAcceptorFn(func(tr tokenReader) error {
		return ErrUnexpectedToken
	})

	errorAcceptor := tokenAcceptorFn(func(tr tokenReader) error {
		return errors.New("some other error")
	})

	tests := []struct {
		name      string
		acceptor  tokenAcceptor
		tokens    []Token
		wantErr   error
		shouldErr bool
	}{
		{
			name:      "invert success to failure",
			acceptor:  notAcceptor(successAcceptor),
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
		{
			name:      "invert failure to success",
			acceptor:  notAcceptor(failAcceptor),
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name:      "forward other errors",
			acceptor:  notAcceptor(errorAcceptor),
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := tt.acceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestDeferAcceptor(t *testing.T) {
	var recursiveAcceptor tokenAcceptor

	recursiveAcceptor = deferAcceptor(func() tokenAcceptor {
		return tokenAcceptorFn(func(tr tokenReader) error {
			token, err := tr.Read()
			if err != nil {
				return err
			}
			if kw, ok := token.(*KeywordToken); ok && kw.Name == "recursive" {
				return recursiveAcceptor.accept(tr)
			}
			return nil
		})
	})

	tests := []struct {
		name      string
		tokens    []Token
		wantErr   error
		shouldErr bool
	}{
		{
			name:      "simple deferred call",
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name: "recursive call",
			tokens: []Token{
				&KeywordToken{Name: "recursive", Position: 0},
				&KeywordToken{Name: "test", Position: 9},
			},
			wantErr:   nil,
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := recursiveAcceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
		})
	}
}

func TestTokenAcceptors(t *testing.T) {
	acceptor1 := tokenAcceptorFn(func(tr tokenReader) error {
		token, err := tr.Read()
		if err != nil {
			return err
		}
		if kw, ok := token.(*KeywordToken); ok && kw.Name == "first" {
			return nil
		}
		return ErrUnexpectedToken
	})

	acceptor2 := tokenAcceptorFn(func(tr tokenReader) error {
		token, err := tr.Read()
		if err != nil {
			return err
		}
		if kw, ok := token.(*KeywordToken); ok && kw.Name == "second" {
			return nil
		}
		return ErrUnexpectedToken
	})

	tests := []struct {
		name      string
		acceptors tokenAcceptors
		tokens    []Token
		wantErr   error
		shouldErr bool
	}{
		{
			name:      "empty acceptors",
			acceptors: tokenAcceptors{},
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name:      "single acceptor success",
			acceptors: tokenAcceptors{acceptor1},
			tokens:    []Token{&KeywordToken{Name: "first", Position: 0}},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name:      "single acceptor failure",
			acceptors: tokenAcceptors{acceptor1},
			tokens:    []Token{&KeywordToken{Name: "wrong", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
		{
			name:      "multiple acceptors success",
			acceptors: tokenAcceptors{acceptor1, acceptor2},
			tokens: []Token{
				&KeywordToken{Name: "first", Position: 0},
				&KeywordToken{Name: "second", Position: 6},
			},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name:      "multiple acceptors first fails",
			acceptors: tokenAcceptors{acceptor1, acceptor2},
			tokens: []Token{
				&KeywordToken{Name: "wrong", Position: 0},
				&KeywordToken{Name: "second", Position: 6},
			},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
		{
			name:      "multiple acceptors second fails",
			acceptors: tokenAcceptors{acceptor1, acceptor2},
			tokens: []Token{
				&KeywordToken{Name: "first", Position: 0},
				&KeywordToken{Name: "wrong", Position: 6},
			},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := tt.acceptors.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestConditionalTokenAcceptor(t *testing.T) {
	ifKeywordAcceptor := tokenAcceptorFn(func(tr tokenReader) error {
		token, err := tr.Read()
		if err != nil {
			return err
		}
		if kw, ok := token.(*KeywordToken); ok && kw.Name == "if" {
			return nil
		}
		return ErrUnexpectedToken
	})

	thenAcceptor := tokenAcceptorFn(func(tr tokenReader) error {
		token, err := tr.Read()
		if err != nil {
			return err
		}
		if kw, ok := token.(*KeywordToken); ok && kw.Name == "then" {
			return nil
		}
		return ErrUnexpectedToken
	})

	elseAcceptor := tokenAcceptorFn(func(tr tokenReader) error {
		token, err := tr.Read()
		if err != nil {
			return err
		}
		if kw, ok := token.(*KeywordToken); ok && kw.Name == "else" {
			return nil
		}
		return ErrUnexpectedToken
	})

	tests := []struct {
		name      string
		acceptor  *conditionalTokenAcceptor
		tokens    []Token
		wantErr   error
		shouldErr bool
	}{
		{
			name: "if condition succeeds, then branch executes",
			acceptor: &conditionalTokenAcceptor{
				ifAccept: ifKeywordAcceptor,
				andThen:  thenAcceptor,
				orElse:   elseAcceptor,
			},
			tokens: []Token{
				&KeywordToken{Name: "if", Position: 0},
				&KeywordToken{Name: "then", Position: 3},
			},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name: "if condition fails, else branch executes",
			acceptor: &conditionalTokenAcceptor{
				ifAccept: ifKeywordAcceptor,
				andThen:  thenAcceptor,
				orElse:   elseAcceptor,
			},
			tokens: []Token{
				&KeywordToken{Name: "else", Position: 0},
			},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name: "if condition fails with ErrNoTokens, else branch executes",
			acceptor: &conditionalTokenAcceptor{
				ifAccept: tokenAcceptorFn(func(tr tokenReader) error { return ErrNoTokens }),
				andThen:  thenAcceptor,
				orElse:   elseAcceptor,
			},
			tokens: []Token{
				&KeywordToken{Name: "else", Position: 0},
			},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name: "if condition succeeds but then branch fails",
			acceptor: &conditionalTokenAcceptor{
				ifAccept: ifKeywordAcceptor,
				andThen:  thenAcceptor,
				orElse:   elseAcceptor,
			},
			tokens: []Token{
				&KeywordToken{Name: "if", Position: 0},
				&KeywordToken{Name: "wrong", Position: 3},
			},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
		{
			name: "if condition has fatal error",
			acceptor: &conditionalTokenAcceptor{
				ifAccept: tokenAcceptorFn(func(tr tokenReader) error { return errors.New("fatal error") }),
				andThen:  thenAcceptor,
				orElse:   elseAcceptor,
			},
			tokens: []Token{
				&KeywordToken{Name: "test", Position: 0},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := tt.acceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAcceptKeyword(t *testing.T) {
	tests := []struct {
		name      string
		keyword   string
		tokens    []Token
		wantErr   error
		shouldErr bool
		consumed  int
	}{
		{
			name:      "matching keyword",
			keyword:   "SELECT",
			tokens:    []Token{&KeywordToken{Name: "SELECT", RawContent: "SELECT", Position: 0}},
			wantErr:   nil,
			shouldErr: false,
			consumed:  1,
		},
		{
			name:      "non-matching keyword",
			keyword:   "SELECT",
			tokens:    []Token{&KeywordToken{Name: "INSERT", RawContent: "INSERT", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
			consumed:  1,
		},
		{
			name:      "non-keyword token",
			keyword:   "SELECT",
			tokens:    []Token{&StringToken{Content: "SELECT", RawContent: "\"SELECT\"", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
			consumed:  1,
		},
		{
			name:      "no tokens",
			keyword:   "SELECT",
			tokens:    []Token{},
			wantErr:   ErrNoTokens,
			shouldErr: true,
			consumed:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acceptor := acceptKeyword(tt.keyword)
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)

			// Count initial tokens
			initialCount := len(tt.tokens)

			err := acceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}

			// Count remaining tokens
			remainingCount := 0
			for tr.Next() {
				_, _ = tr.Read()
				remainingCount++
			}

			actualConsumed := initialCount - remainingCount
			if actualConsumed != tt.consumed {
				t.Errorf("expected %d tokens consumed, got %d", tt.consumed, actualConsumed)
			}
		})
	}
}

func TestAcceptOperator(t *testing.T) {
	tests := []struct {
		name      string
		operator  string
		tokens    []Token
		wantErr   error
		shouldErr bool
		consumed  int
	}{
		{
			name:      "matching operator",
			operator:  "=",
			tokens:    []Token{&OperatorToken{Type: "=", RawContent: "=", Position: 0}},
			wantErr:   nil,
			shouldErr: false,
			consumed:  1,
		},
		{
			name:      "non-matching operator",
			operator:  "=",
			tokens:    []Token{&OperatorToken{Type: "!=", RawContent: "!=", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
			consumed:  1,
		},
		{
			name:      "non-operator token",
			operator:  "=",
			tokens:    []Token{&KeywordToken{Name: "=", RawContent: "=", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
			consumed:  1,
		},
		{
			name:      "no tokens",
			operator:  "=",
			tokens:    []Token{},
			wantErr:   ErrNoTokens,
			shouldErr: true,
			consumed:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acceptor := acceptOperator(tt.operator)
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)

			initialCount := len(tt.tokens)
			err := acceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}

			remainingCount := 0
			for tr.Next() {
				_, _ = tr.Read()
				remainingCount++
			}

			actualConsumed := initialCount - remainingCount
			if actualConsumed != tt.consumed {
				t.Errorf("expected %d tokens consumed, got %d", tt.consumed, actualConsumed)
			}
		})
	}
}

func TestAcceptSingleToken(t *testing.T) {
	successHandler := func(*KeywordToken) error { return nil }
	failHandler := func(*KeywordToken) error { return ErrUnexpectedToken }

	tests := []struct {
		name      string
		handler   func(*KeywordToken) error
		tokens    []Token
		wantErr   error
		shouldErr bool
	}{
		{
			name:      "matching token type and successful handler",
			handler:   successHandler,
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name:      "matching token type but failing handler",
			handler:   failHandler,
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
		{
			name:      "non-matching token type",
			handler:   successHandler,
			tokens:    []Token{&StringToken{Content: "test", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
		{
			name:      "no tokens",
			handler:   successHandler,
			tokens:    []Token{},
			wantErr:   ErrNoTokens,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			acceptor := acceptSingleToken(tt.handler)
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := acceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAcceptEitherToken(t *testing.T) {
	keywordHandler := func(*KeywordToken) error { return nil }
	stringHandler := func(*StringToken) error { return nil }
	failKeywordHandler := func(*KeywordToken) error { return ErrUnexpectedToken }

	tests := []struct {
		name         string
		leftHandler  func(*KeywordToken) error
		rightHandler func(*StringToken) error
		tokens       []Token
		wantErr      error
		shouldErr    bool
	}{
		{
			name:         "left token type matches",
			leftHandler:  keywordHandler,
			rightHandler: stringHandler,
			tokens:       []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:      nil,
			shouldErr:    false,
		},
		{
			name:         "right token type matches",
			leftHandler:  keywordHandler,
			rightHandler: stringHandler,
			tokens:       []Token{&StringToken{Content: "test", Position: 0}},
			wantErr:      nil,
			shouldErr:    false,
		},
		{
			name:         "left handler fails",
			leftHandler:  failKeywordHandler,
			rightHandler: stringHandler,
			tokens:       []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:      ErrUnexpectedToken,
			shouldErr:    true,
		},
		{
			name:         "neither token type matches",
			leftHandler:  keywordHandler,
			rightHandler: stringHandler,
			tokens:       []Token{&NumericToken{Int64: 42, Position: 0}},
			wantErr:      ErrUnexpectedToken,
			shouldErr:    true,
		},
		{
			name:         "no tokens",
			leftHandler:  keywordHandler,
			rightHandler: stringHandler,
			tokens:       []Token{},
			wantErr:      ErrNoTokens,
			shouldErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acceptor := acceptEitherToken(tt.leftHandler, tt.rightHandler)
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := acceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAcceptTokenFromAny3(t *testing.T) {
	keywordHandler := func(*KeywordToken) error { return nil }
	stringHandler := func(*StringToken) error { return nil }
	numericHandler := func(*NumericToken) error { return nil }

	tests := []struct {
		name          string
		leftHandler   func(*KeywordToken) error
		centerHandler func(*StringToken) error
		rightHandler  func(*NumericToken) error
		tokens        []Token
		wantErr       error
		shouldErr     bool
	}{
		{
			name:          "left token matches",
			leftHandler:   keywordHandler,
			centerHandler: stringHandler,
			rightHandler:  numericHandler,
			tokens:        []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:       nil,
			shouldErr:     false,
		},
		{
			name:          "center token matches",
			leftHandler:   keywordHandler,
			centerHandler: stringHandler,
			rightHandler:  numericHandler,
			tokens:        []Token{&StringToken{Content: "test", Position: 0}},
			wantErr:       nil,
			shouldErr:     false,
		},
		{
			name:          "right token matches",
			leftHandler:   keywordHandler,
			centerHandler: stringHandler,
			rightHandler:  numericHandler,
			tokens:        []Token{&NumericToken{Int64: 42, Position: 0}},
			wantErr:       nil,
			shouldErr:     false,
		},
		{
			name:          "no token matches",
			leftHandler:   keywordHandler,
			centerHandler: stringHandler,
			rightHandler:  numericHandler,
			tokens:        []Token{&BooleanToken{Value: true, Position: 0}},
			wantErr:       ErrUnexpectedToken,
			shouldErr:     true,
		},
		{
			name:          "no tokens",
			leftHandler:   keywordHandler,
			centerHandler: stringHandler,
			rightHandler:  numericHandler,
			tokens:        []Token{},
			wantErr:       ErrNoTokens,
			shouldErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acceptor := acceptTokenFromAny3(tt.leftHandler, tt.centerHandler, tt.rightHandler)
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := acceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAcceptTokenFromAny4(t *testing.T) {
	keywordHandler := func(*KeywordToken) (tokenAcceptor, error) { return nopAcceptor, nil }
	stringHandler := func(*StringToken) (tokenAcceptor, error) { return nopAcceptor, nil }
	numericHandler := func(*NumericToken) (tokenAcceptor, error) { return nopAcceptor, nil }
	booleanHandler := func(*BooleanToken) (tokenAcceptor, error) { return nopAcceptor, nil }

	errorHandler := func(*KeywordToken) (tokenAcceptor, error) { return nil, errors.New("handler error") }
	nilAcceptorHandler := func(*KeywordToken) (tokenAcceptor, error) { return nil, nil }

	tests := []struct {
		name         string
		leftHandler  func(*KeywordToken) (tokenAcceptor, error)
		clHandler    func(*StringToken) (tokenAcceptor, error)
		crHandler    func(*NumericToken) (tokenAcceptor, error)
		rightHandler func(*BooleanToken) (tokenAcceptor, error)
		tokens       []Token
		wantErr      error
		shouldErr    bool
	}{
		{
			name:         "left token matches",
			leftHandler:  keywordHandler,
			clHandler:    stringHandler,
			crHandler:    numericHandler,
			rightHandler: booleanHandler,
			tokens:       []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:      nil,
			shouldErr:    false,
		},
		{
			name:         "handler returns error",
			leftHandler:  errorHandler,
			clHandler:    stringHandler,
			crHandler:    numericHandler,
			rightHandler: booleanHandler,
			tokens:       []Token{&KeywordToken{Name: "test", Position: 0}},
			shouldErr:    true,
		},
		{
			name:         "handler returns nil acceptor",
			leftHandler:  nilAcceptorHandler,
			clHandler:    stringHandler,
			crHandler:    numericHandler,
			rightHandler: booleanHandler,
			tokens:       []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:      nil,
			shouldErr:    false,
		},
		{
			name:         "no token matches",
			leftHandler:  keywordHandler,
			clHandler:    stringHandler,
			crHandler:    numericHandler,
			rightHandler: booleanHandler,
			tokens:       []Token{&OperatorToken{Type: "=", Position: 0}},
			wantErr:      ErrUnexpectedToken,
			shouldErr:    true,
		},
		{
			name:         "no tokens",
			leftHandler:  keywordHandler,
			clHandler:    stringHandler,
			crHandler:    numericHandler,
			rightHandler: booleanHandler,
			tokens:       []Token{},
			wantErr:      ErrNoTokens,
			shouldErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acceptor := acceptTokenFromAny4(tt.leftHandler, tt.clHandler, tt.crHandler, tt.rightHandler)
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := acceptor.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAcceptWhitespaceToken(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []Token
		wantErr   error
		shouldErr bool
	}{
		{
			name:      "whitespace token",
			tokens:    []Token{&WhitespaceToken{Content: " ", Position: 0}},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name:      "non-whitespace token",
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
		{
			name:      "no tokens",
			tokens:    []Token{},
			wantErr:   ErrNoTokens,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := acceptWhitespaceToken.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAcceptWildcardToken(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []Token
		wantErr   error
		shouldErr bool
	}{
		{
			name:      "wildcard token",
			tokens:    []Token{&WildcardToken{Position: 0}},
			wantErr:   nil,
			shouldErr: false,
		},
		{
			name:      "non-wildcard token",
			tokens:    []Token{&KeywordToken{Name: "test", Position: 0}},
			wantErr:   ErrUnexpectedToken,
			shouldErr: true,
		},
		{
			name:      "no tokens",
			tokens:    []Token{},
			wantErr:   ErrNoTokens,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			err := acceptWildcardToken.accept(tr)

			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestSkipWhitespaceToken(t *testing.T) {
	tests := []struct {
		name   string
		tokens []Token
	}{
		{
			name:   "whitespace token - consumed",
			tokens: []Token{&WhitespaceToken{Content: " ", Position: 0}},
		},
		{
			name:   "non-whitespace token - not consumed",
			tokens: []Token{&KeywordToken{Name: "test", Position: 0}},
		},
		{
			name:   "no tokens",
			tokens: []Token{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewSliceTokenSource(tt.tokens)
			originalNext := tr.Next()

			err := skipWhitespaceToken.accept(tr)

			// skipWhitespaceToken should never return an error
			if err != nil {
				t.Errorf("skipWhitespaceToken should never return error, got %v", err)
			}

			// Check if token was consumed correctly
			newNext := tr.Next()
			if len(tt.tokens) > 0 {
				if _, ok := tt.tokens[0].(*WhitespaceToken); ok {
					// Whitespace should be consumed
					if originalNext && newNext {
						t.Error("whitespace token should have been consumed")
					}
				} else {
					// Non-whitespace should not be consumed
					if originalNext != newNext {
						t.Error("non-whitespace token should not have been consumed")
					}
				}
			}
		})
	}
}

// Test with reader errors
func TestAcceptorsWithReaderErrors(t *testing.T) {
	readerError := errors.New("reader error")

	tests := []struct {
		name     string
		acceptor tokenAcceptor
	}{
		{
			name:     "acceptKeyword with reader error",
			acceptor: acceptKeyword("test"),
		},
		{
			name:     "acceptOperator with reader error",
			acceptor: acceptOperator("="),
		},
		{
			name:     "acceptWhitespaceToken with reader error",
			acceptor: acceptWhitespaceToken,
		},
		{
			name:     "acceptWildcardToken with reader error",
			acceptor: acceptWildcardToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := defaultTokenSourceFactory.NewErrorTokenSource([]Token{}, map[int]error{0: readerError})
			err := tt.acceptor.accept(tr)

			if !errors.Is(err, readerError) {
				t.Errorf("expected reader error %v, got %v", readerError, err)
			}
		})
	}
}
