package testutils

import (
	"log"
	"runtime"
	"strings"
)

// tokenSource is an alias interface for gqlparser.TokenSource to avoid cyclic dependencies.
type tokenSource[T any] interface {
	// Next checks if there are more tokens to read.
	Next() bool

	// Read reads the next token from the source.
	// Returns a zero value of type T and an error if there are no more tokens.
	Read() (T, error)

	// Unread puts the token back into the source for re-reading.
	Unread(token T)
}

// TestTokenSourceFactory is a factory for creating various types of token sources for testing.
// It provides methods to create both successful and error-prone token sources with different configurations.
// T must be a type that implements the gqlparser.Token interface.
type TestTokenSourceFactory[T any] struct {
	errEndOfToken error
}

// NewTestTokenSourceFactory creates a new factory with the default end-of-token error.
func NewTestTokenSourceFactory[T any](errEndOfToken error) *TestTokenSourceFactory[T] {
	return &TestTokenSourceFactory[T]{
		errEndOfToken: errEndOfToken,
	}
}

// NewSliceTokenSource creates a new sliceTokenSource with the given tokens.
// It uses the factory's errEndOfToken setting.
func (f *TestTokenSourceFactory[T]) NewSliceTokenSource(tokens []T) *sliceTokenSource[T] {
	return &sliceTokenSource[T]{
		s:             tokens,
		errEndOfToken: f.errEndOfToken,
	}
}

// NewErrorTokenSource creates a new errorTokenSource with the given tokens and error settings.
// It always uses the factory's errEndOfToken setting.
func (f *TestTokenSourceFactory[T]) NewErrorTokenSource(tokens []T, errorMap map[int]error) *errorTokenSource[T] {
	return &errorTokenSource[T]{
		tokens:        tokens,
		pos:           0,
		errorMap:      errorMap,
		errEndOfToken: f.errEndOfToken,
	}
}

// sliceTokenSource is a simple implementation of TokenSource that reads from a slice of tokens.
// It implements the TokenSource interface and allows reading tokens sequentially.
// It also supports unreading tokens, which allows for backtracking in the token stream.
// T must be the gqlparser.Token interface to implement the TokenSource interface.
//
// Usage examples:
//   - defaultTokenSourceFactory.NewSliceTokenSource(tokens)
//   - Custom factory: NewTestTokenSourceFactory[CustomTokenType]().NewSliceTokenSource(customTokens)
type sliceTokenSource[T any] struct {
	s             []T
	errEndOfToken error
}

func (ts *sliceTokenSource[T]) Read() (T, error) {
	if len(ts.s) == 0 {
		var zero T // Create a zero value of type T
		return zero, ts.errEndOfToken
	}

	tok := ts.s[0]
	ts.s = ts.s[1:]
	return tok, nil
}

func (ts *sliceTokenSource[T]) Unread(tok T) {
	ts.s = append([]T{tok}, ts.s...)
}

func (ts *sliceTokenSource[T]) Next() bool {
	return len(ts.s) != 0
}

// errorTokenSource is a mock TokenSource that returns an error
// T must be the gqlparser.Token interface to implement the TokenSource interface.
//
// Usage examples:
//   - defaultTokenSourceFactory.NewErrorTokenSource(tokens, errorMap)
//   - defaultTokenSourceFactory.NewErrorTokenSourceWithMultipleErrors(tokens, errorMap)
//   - Custom factory: NewTestTokenSourceFactory[CustomTokenType]().NewErrorTokenSource(customTokens, errorMap)
type errorTokenSource[T any] struct {
	tokens        []T
	pos           int
	errorMap      map[int]error // Map of position to error
	errEndOfToken error
	unreadTokens  []T // Stack of unread tokens
}

func (ts *errorTokenSource[T]) Next() bool {
	// If we have unread tokens, we can read them
	if len(ts.unreadTokens) > 0 {
		return true
	}

	// If we have an error at current position, we should still indicate there's a token to read
	// so that Read() gets called and can return the error
	if _, hasError := ts.errorMap[ts.pos]; hasError {
		return true
	}
	return ts.pos < len(ts.tokens)
}

func (ts *errorTokenSource[T]) Read() (T, error) {
	// If we have unread tokens, return them first
	if len(ts.unreadTokens) > 0 {
		token := ts.unreadTokens[len(ts.unreadTokens)-1]
		ts.unreadTokens = ts.unreadTokens[:len(ts.unreadTokens)-1]
		return token, nil
	}

	// Check if there's an error at the current position
	if err, hasError := ts.errorMap[ts.pos]; hasError {
		ts.pos++   // Increment position after encountering error
		var zero T // Create a zero value of type T
		return zero, err
	}

	if ts.pos >= len(ts.tokens) {
		var zero T // Create a zero value of type T
		return zero, ts.errEndOfToken
	}
	token := ts.tokens[ts.pos]
	ts.pos++
	return token, nil
}

func (ts *errorTokenSource[T]) Unread(token T) {
	// Only accept unread if we've actually read at least one token
	if ts.pos > 0 || len(ts.unreadTokens) > 0 {
		ts.unreadTokens = append(ts.unreadTokens, token)
	}
	// If pos is 0 and no unread tokens, we ignore the unread (handle gracefully)
}

// SetErrorAtPosition sets an error to be returned when reading at the specified position.
// This allows for dynamic error configuration after the errorTokenSource is created.
func (ts *errorTokenSource[T]) SetErrorAtPosition(position int, err error) {
	if ts.errorMap == nil {
		ts.errorMap = make(map[int]error)
	}
	ts.errorMap[position] = err
}

// ClearErrorAtPosition removes the error at the specified position.
func (ts *errorTokenSource[T]) ClearErrorAtPosition(position int) {
	if ts.errorMap != nil {
		delete(ts.errorMap, position)
	}
}

// ClearAllErrors removes all position-specific errors.
func (ts *errorTokenSource[T]) ClearAllErrors() {
	ts.errorMap = make(map[int]error)
}

// DebugLogger is an interface for logging debug messages.
type DebugLogger interface {
	Logf(string, ...any)
}

// DebugTokenSource is a wrapper around TokenSource that adds debug logging
// for the Next, Read, and Unread methods.
type DebugTokenSource[T any] struct {
	Source tokenSource[T]
	Logger DebugLogger
}

func (ts *DebugTokenSource[T]) getCaller() (file string, line int) {
	var rpc [16]uintptr
	n := runtime.Callers(3, rpc[:])
	if n == 0 {
		panic("cannot get caller")
	}

	frames := runtime.CallersFrames(rpc[:])
	for {
		frame, hasNext := frames.Next()
		if !strings.HasSuffix(frame.File, "/token_reader.go") && !strings.HasSuffix(frame.File, "/token_source.go") {
			return frame.File, frame.Line
		}
		if !hasNext {
			return frame.File, frame.Line
		}
	}
}

func (ts *DebugTokenSource[T]) Next() bool {
	next := ts.Source.Next()
	file, line := ts.getCaller()
	log.Printf("Next() = %v at %s line %d", next, file, line)
	return next
}

func (ts *DebugTokenSource[T]) Read() (T, error) {
	file, line := ts.getCaller()
	token, err := ts.Source.Read()
	if err != nil {
		ts.Logger.Logf("Read() = error (%+v) at %s line %d", err, file, line)
		var zero T
		return zero, err
	}

	ts.Logger.Logf("Read() = %+v at %s line %d", token, file, line)
	return token, nil
}

func (ts *DebugTokenSource[T]) Unread(token T) {
	file, line := ts.getCaller()
	ts.Logger.Logf("Unread(%+v) at %s line %d", token, file, line)
	ts.Source.Unread(token)
}
