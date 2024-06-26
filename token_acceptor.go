package gqlparser

import (
	"errors"
	"fmt"
)

type tokenAcceptor interface {
	accept(tokenReader) error
}

type tokenAcceptorFn func(tokenReader) error

func (f tokenAcceptorFn) accept(tr tokenReader) error {
	return f(tr)
}

type nopAcceptorTyp struct{}

func (f nopAcceptorTyp) accept(tr tokenReader) error {
	return nil
}

var nopAcceptor nopAcceptorTyp

func notAcceptor(acceptor tokenAcceptor) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		rtr := asResettableTokenReader(tr)
		if err := acceptor.accept(rtr); errors.Is(err, ErrUnexpectedToken) {
			rtr.Reset()
			return nil
		} else if err != nil {
			rtr.Reset()
			return err
		}
		return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, rtr.history.tokens[0].GetContent(), rtr.history.tokens[0].GetPosition())
	})
}

func advanceAcceptor(acceptor tokenAcceptor) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		rtr := asResettableTokenReader(tr)
		defer rtr.Reset()
		if err := acceptor.accept(rtr); err != nil {
			return err
		}
		return nil
	})
}

func deferAcceptor(getAcceptor func() tokenAcceptor) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		acceptor := getAcceptor()
		return acceptor.accept(tr)
	})
}

type tokenAcceptors []tokenAcceptor

func (s tokenAcceptors) accept(tr tokenReader) error {
	for _, acceptor := range s {
		if err := acceptor.accept(tr); err != nil {
			return err
		}
	}
	return nil
}

type conditionalTokenAcceptor struct {
	ifAccept tokenAcceptor
	andThen  tokenAcceptor
	orElse   tokenAcceptor
}

func (acceptor *conditionalTokenAcceptor) accept(tr tokenReader) error {
	rtr := asResettableTokenReader(tr)
	if err := acceptor.ifAccept.accept(rtr); errors.Is(err, ErrUnexpectedToken) || errors.Is(err, ErrNoTokens) {
		rtr.Reset()
		return acceptor.orElse.accept(tr)
	} else if err != nil {
		rtr.Reset()
		return err
	}
	return acceptor.andThen.accept(tr)
}

func acceptKeyword(keywords ...string) tokenAcceptor {
	set := make(map[string]struct{}, len(keywords))
	for _, keyword := range keywords {
		set[keyword] = struct{}{}
	}

	return tokenAcceptorFn(func(tr tokenReader) error {
		if token, err := tr.Read(); errors.Is(err, ErrEndOfToken) {
			return ErrNoTokens
		} else if err != nil {
			return err
		} else if t, ok := token.(*KeywordToken); ok {
			if _, ok := set[t.Name]; !ok {
				return fmt.Errorf("%w: %s at %d (expect to be any of %q)", ErrUnexpectedToken, t.GetContent(), t.GetPosition(), keywords)
			}
			return nil
		} else {
			return fmt.Errorf("%w: %s at %d (expect to be any of %q)", ErrUnexpectedToken, token.GetContent(), token.GetPosition(), keywords)
		}
	})
}

func acceptOperator(operator string) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		if token, err := tr.Read(); errors.Is(err, ErrEndOfToken) {
			return ErrNoTokens
		} else if err != nil {
			return err
		} else if t, ok := token.(*OperatorToken); ok {
			if t.Type != operator {
				return fmt.Errorf("%w: %s at %d (expect to be %q)", ErrUnexpectedToken, t.Type, t.Position, operator)
			}
			return nil
		} else {
			return fmt.Errorf("%w: %s at %d (expect to be %q)", ErrUnexpectedToken, token.GetContent(), token.GetPosition(), operator)
		}
	})
}

func acceptSingleToken[T Token](f func(T) error) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		if token, err := tr.Read(); errors.Is(err, ErrEndOfToken) {
			return ErrNoTokens
		} else if err != nil {
			return err
		} else if t, ok := token.(T); ok {
			return f(t)
		} else {
			return fmt.Errorf("%w: %s at %d (expect to be %T)", ErrUnexpectedToken, token.GetContent(), token.GetPosition(), t)
		}
	})
}

func acceptEitherToken[L Token, R Token](lf func(L) error, rf func(R) error) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		if token, err := tr.Read(); errors.Is(err, ErrEndOfToken) {
			return ErrNoTokens
		} else if err != nil {
			return err
		} else {
			switch t := token.(type) {
			case L:
				return lf(t)
			case R:
				return rf(t)
			default:
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
			}
		}
	})
}

func acceptTokenFromAny3[L Token, C Token, R Token](lf func(L) error, cf func(C) error, rf func(R) error) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		if token, err := tr.Read(); errors.Is(err, ErrEndOfToken) {
			return ErrNoTokens
		} else if err != nil {
			return err
		} else {
			switch t := token.(type) {
			case L:
				return lf(t)
			case C:
				return cf(t)
			case R:
				return rf(t)
			default:
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
			}
		}
	})
}

var acceptWhitespaceToken = acceptSingleToken(func(*WhitespaceToken) error {
	return nil
})

var acceptWildcardToken = acceptSingleToken(func(*WildcardToken) error {
	return nil
})

var skipWhitespaceToken tokenAcceptorFn = func(tr tokenReader) error {
	rtr := asResettableTokenReader(tr)
	if token, err := rtr.Read(); errors.Is(err, ErrEndOfToken) {
		return nil
	} else if err != nil {
		return err
	} else if _, ok := token.(*WhitespaceToken); ok {
		return nil
	} else {
		rtr.Reset()
		return nil
	}
}
