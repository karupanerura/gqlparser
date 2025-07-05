package gqlparser

import (
	"errors"
	"fmt"
)

// tokenAcceptor represents a parser component that attempts to match
// a specific token pattern in the input stream.
//
// The accept method reads from the provided tokenReader and tries
// to consume one or more tokens according to its grammar rule.
//   - On success, accept returns nil and the reader remains positioned
//     after the consumed tokens.
//   - If the upcoming tokens do not match, accept returns ErrUnexpectedToken,
//     allowing callers to reset the reader and backtrack.
//   - Any other error is treated as a parsing failure and is returned directly.
type tokenAcceptor interface {
	// accept checks whether the next tokens in tr match this acceptor's pattern.
	// Returns nil on a match, ErrUnexpectedToken on a non-match, or another error on failure.
	accept(tokenReader) error
}

// tokenAcceptorFn adapts a plain function to implement tokenAcceptor.
type tokenAcceptorFn func(tokenReader) error

// accept calls the underlying function to perform matching.
// It satisfies the tokenAcceptor interface.
func (f tokenAcceptorFn) accept(tr tokenReader) error {
	return f(tr)
}

type nopAcceptorTyp struct{}

func (f nopAcceptorTyp) accept(tr tokenReader) error {
	return nil
}

// nopAcceptor is a no-op acceptor that always succeeds.
// It can be used as a placeholder or default acceptor in cases where
// no specific token matching is required.
// It is equivalent to a function that does nothing and returns nil.
// It is useful for cases where an acceptor is required but no specific
// behavior is needed.
var nopAcceptor nopAcceptorTyp

// notAcceptor returns an acceptor that inverts the result of the given one.
// If the inner acceptor returns ErrUnexpectedToken, notAcceptor succeeds;
// if the inner acceptor succeeds, notAcceptor fails with ErrUnexpectedToken;
// other errors are forwarded.
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
		return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken,
			rtr.history.tokens[0].GetContent(), rtr.history.tokens[0].GetPosition())
	})
}

// advanceAcceptor wraps an acceptor to ensure it always runs on a resettable reader.
// This is useful for acceptors that need to be run multiple times,
// such as in recursive grammar definitions or when the reader needs to be reset
// after each match attempt.
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

// deferAcceptor delays the construction of an acceptor until it is run,
// enabling recursive grammar definitions.
func deferAcceptor(getAcceptor func() tokenAcceptor) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		acceptor := getAcceptor()
		return acceptor.accept(tr)
	})
}

// tokenAcceptors is a sequence of acceptors to apply in order.
// Read tokens with each until one fails or all succeed.
type tokenAcceptors []tokenAcceptor

func (s tokenAcceptors) accept(tr tokenReader) error {
	for _, acceptor := range s {
		if err := acceptor.accept(tr); err != nil {
			return err
		}
	}
	return nil
}

// conditionalTokenAcceptor implements an if-then-else style control flow
// for token acceptors. It first applies the `ifAccept` acceptor on a
// resettable reader snapshot:
//   - If `ifAccept` returns nil, it succeeds, and the original reader is
//     advanced past the matched tokens; then `andThen` is applied to the
//     same reader to continue parsing in the "then" branch.
//   - If `ifAccept` returns ErrUnexpectedToken or ErrNoTokens, it resets
//     the reader to its original state and applies `orElse`, enabling
//     the "else" branch parser logic.
//   - Any other error from `ifAccept` is considered fatal, resets the
//     reader, and is returned immediately.
type conditionalTokenAcceptor struct {
	// ifAccept is the condition acceptor tested against the input stream.
	// It runs on a resettable reader snapshot.
	ifAccept tokenAcceptor

	// andThen is invoked when ifAccept succeeds, continuing parsing
	// in the "then" branch.
	andThen tokenAcceptor

	// orElse is invoked when ifAccept fails with ErrUnexpectedToken
	// or ErrNoTokens, handling the "else" branch.
	orElse tokenAcceptor
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

// acceptKeyword returns a tokenAcceptor that matches the next token
// against the specified keyword string.
// It reads one token from the reader and:
//   - returns nil if the token is a KeywordToken with Name equal to keyword,
//   - returns ErrNoTokens if there are no tokens remaining,
//   - returns ErrUnexpectedToken (with details) if the token does not match,
//   - forwards any other errors from the reader.
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

// acceptOperator returns a tokenAcceptor that matches the next token
// against the specified operator string.
// It reads one token from the reader and:
//   - returns nil if the token is an OperatorToken with Type equal to operator,
//   - returns ErrNoTokens if there are no tokens remaining,
//   - returns ErrUnexpectedToken (with details) if the token does not match,
//   - forwards any other errors from the reader.
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

// acceptSingleToken returns a tokenAcceptor that reads exactly one token from the reader
// and applies the provided function f if the token is of type T.
// It returns ErrNoTokens if there are no tokens, ErrUnexpectedToken if the token
// is not of the expected type, and forwards any other errors from the reader.
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

// acceptEitherToken returns a tokenAcceptor that reads one token from the reader
// and applies lf if the token is of type L, or rf if the token is of type R.
// It returns ErrNoTokens if there are no tokens remaining, ErrUnexpectedToken
// if the token is not of either type, and forwards any other errors from the reader.
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

// acceptTokenFromAny3 returns a tokenAcceptor that reads one token from the reader
// and applies lf if the token is of type L, cf if the token is of type C,
// or rf if the token is of type R.
// It returns ErrNoTokens if there are no tokens remaining, ErrUnexpectedToken
// if the token is not of any of the expected types, and forwards any other errors from the reader.
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

// acceptTokenFromAny4 returns a tokenAcceptor that reads one token from the reader and
// invokes lf if the token is of type L, clf if of type CL, crf if of type CR,
// or rf if of type R.
// Each handler receives the token and returns a tokenAcceptor to apply to the remaining
// stream or an error. If the handler returns a non-nil acceptor, it is run on the reader.
// If no handler matches, ErrUnexpectedToken is returned with details.
// ErrNoTokens is returned if there are no tokens, and other reader errors are forwarded.
func acceptTokenFromAny4[L Token, CL Token, CR Token, R Token](lf func(L) (tokenAcceptor, error), clf func(CL) (tokenAcceptor, error), crf func(CR) (tokenAcceptor, error), rf func(R) (tokenAcceptor, error)) tokenAcceptor {
	return tokenAcceptorFn(func(tr tokenReader) error {
		if token, err := tr.Read(); errors.Is(err, ErrEndOfToken) {
			return ErrNoTokens
		} else if err != nil {
			return err
		} else {
			var (
				acceptor tokenAcceptor
				err      error
			)
			switch t := token.(type) {
			case L:
				acceptor, err = lf(t)
			case CL:
				acceptor, err = clf(t)
			case CR:
				acceptor, err = crf(t)
			case R:
				acceptor, err = rf(t)
			default:
				return fmt.Errorf("%w: %s at %d", ErrUnexpectedToken, token.GetContent(), token.GetPosition())
			}

			if err != nil {
				return err
			}
			if acceptor != nil {
				if err := acceptor.accept(tr); err != nil {
					return err
				}
				return nil
			}
			return nil
		}
	})
}

// acceptWhitespaceToken is a tokenAcceptor that consumes exactly one
// WhitespaceToken from the reader. It returns nil if the next token is a
// WhitespaceToken, ErrNoTokens if there are no tokens remaining, or
// ErrUnexpectedToken if the next token is not WhitespaceToken.
var acceptWhitespaceToken = acceptSingleToken(func(*WhitespaceToken) error {
	return nil
})

// acceptWildcardToken is a tokenAcceptor that consumes exactly one
// WildcardToken from the reader. It returns nil if the next token is a
// WildcardToken, ErrNoTokens if there are no tokens remaining, or
// ErrUnexpectedToken if the next token is not WildcardToken.
var acceptWildcardToken = acceptSingleToken(func(*WildcardToken) error {
	return nil
})

// skipWhitespaceToken is a tokenAcceptor that attempts to consume one WhitespaceToken.
// If the next token is whitespace, it is consumed and the reader advances.
// If the next token is not whitespace, the reader is reset and no tokens are consumed.
// It always returns nil, never producing an error.
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
