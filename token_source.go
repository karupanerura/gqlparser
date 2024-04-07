package gqlparser

type TokenSource interface {
	Next() bool
	Read() (Token, error)
	Unread(Token)
}
