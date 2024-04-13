package gqlparser_test

import (
	"log"
	"runtime"
	"strings"

	"github.com/karupanerura/gqlparser"
)

type debugLogger interface {
	Logf(string, ...any)
}

type debugTokenSource struct {
	source gqlparser.TokenSource
	logger debugLogger
}

func (ts *debugTokenSource) getCaller() (file string, line int) {
	var rpc [16]uintptr
	n := runtime.Callers(3, rpc[:])
	if n == 0 {
		panic("cannot get caller")
	}

	frames := runtime.CallersFrames(rpc[:])
	for {
		frame, hasNext := frames.Next()
		if !strings.HasSuffix(frame.File, "/token_reader.go") && !strings.HasSuffix(frame.File, "/debug.go") {
			return frame.File, frame.Line
		}
		if !hasNext {
			return frame.File, frame.Line
		}
	}
}

func (ts *debugTokenSource) Next() bool {
	next := ts.source.Next()
	file, line := ts.getCaller()
	log.Printf("Next() = %v at %s line %d", next, file, line)
	return next
}

func (ts *debugTokenSource) Read() (gqlparser.Token, error) {
	file, line := ts.getCaller()
	token, err := ts.source.Read()
	if err != nil {
		ts.logger.Logf("Read() = error (%+v) at %s line %d", err, file, line)
		return nil, err
	}

	ts.logger.Logf("Read() = %+v at %s line %d", token, file, line)
	return token, nil
}

func (ts *debugTokenSource) Unread(token gqlparser.Token) {
	file, line := ts.getCaller()
	ts.logger.Logf("Unread(%+v) at %s line %d", token, file, line)
	ts.source.Unread(token)
}
