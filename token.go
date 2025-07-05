package gqlparser

import "strconv"

// privateSealed is just a private type. it's used to limit patterns like sealed class.
type privateSealed struct{}

// Token is an interface for representing tokens in the parser.
type Token interface {
	isToken() privateSealed

	// GetContent returns the content of the token.
	// The content is the string representation of the token.
	// This is used for debugging and error reporting.
	GetContent() string

	// GetPosition returns the position of the token in the source string.
	// The position is the index of the first character of the token in the source string.
	// This is used for debugging and error reporting.
	GetPosition() int
}

// StringToken represents a string token in the parser.
type StringToken struct {
	// Quote is the quote character used to delimit the string.
	// It can be either a single quote (') or a double quote (").
	// This is used to determine if the string is a single-quoted or double-quoted string.
	Quote byte

	// Content is the unquoted content of the string.
	// This is the string without the surrounding quotes.
	// It may contain escape sequences that need to be unescaped.
	// For example, if the string is "Hello, World!", the content would be Hello, World!.
	// The content is the actual value of the string token.
	Content string

	// RawContent is the raw string representation of the token.
	// This includes the surrounding quotes and any escape sequences.
	// For example, if the string is "Hello, World!", the raw content would be "Hello, World!".
	RawContent string

	// Position is the position of the token in the source string.
	// This is the index of the first character of the token in the source string.
	Position int
}

func (*StringToken) isToken() privateSealed { return privateSealed{} }
func (t *StringToken) GetContent() string   { return t.RawContent }
func (t *StringToken) GetPosition() int     { return t.Position }

// OperatorToken represents an operator token in the parser.
type OperatorToken struct {
	// Type is the type of the operator.
	Type string

	// RawContent is the raw string representation of the token.
	// This is an optional field. It is defined when different from the Type field.
	RawContent string

	// Position is the position of the token in the source string.
	Position int
}

func (*OperatorToken) isToken() privateSealed { return privateSealed{} }

func (t *OperatorToken) GetContent() string {
	if t.RawContent != "" {
		return t.RawContent
	}
	return t.Type
}

func (t *OperatorToken) GetPosition() int { return t.Position }

// WildcardToken represents a wildcard token in the parser.
type WildcardToken struct {
	// Position is the position of the token in the source string.
	// This is the index of the first character of the token in the source string.
	Position int
}

func (*WildcardToken) isToken() privateSealed { return privateSealed{} }
func (t *WildcardToken) GetContent() string   { return "*" }
func (t *WildcardToken) GetPosition() int     { return t.Position }

// BooleanToken represents a boolean token in the parser.
type BooleanToken struct {
	// Value is the boolean value of the token.
	Value bool

	// RawContent is the raw string representation of the token.
	RawContent string

	// Position is the position of the token in the source string.
	// This is the index of the first character of the token in the source string.
	Position int
}

func (*BooleanToken) isToken() privateSealed { return privateSealed{} }

func (t *BooleanToken) GetContent() string {
	return t.RawContent
}

func (t *BooleanToken) GetPosition() int { return t.Position }

// OrderToken represents an order token in the parser.
type OrderToken struct {
	// Descending indicates if the order is descending.
	// If true, the order is descending. If false, the order is ascending.
	Descending bool

	// RawContent is the raw string representation of the token.
	RawContent string

	// Position is the position of the token in the source string.
	Position int
}

func (*OrderToken) isToken() privateSealed { return privateSealed{} }

func (t *OrderToken) GetContent() string {
	return t.RawContent
}

func (t *OrderToken) GetPosition() int { return t.Position }

// SymbolToken represents a symbol token in the parser.
type SymbolToken struct {
	// Content is the content of the symbol token.
	// This is the string representation of the symbol.
	Content string

	// Position is the position of the token in the source string.
	Position int
}

func (*SymbolToken) isToken() privateSealed { return privateSealed{} }
func (t *SymbolToken) GetContent() string   { return t.Content }
func (t *SymbolToken) GetPosition() int     { return t.Position }

// KeywordToken represents a keyword token in the parser.
type KeywordToken struct {
	// Name is the canonicalized name of the keyword.
	Name string

	// RawContent is the raw string representation of the token.
	RawContent string

	// Position is the position of the token in the source string.
	Position int
}

func (*KeywordToken) isToken() privateSealed { return privateSealed{} }
func (t *KeywordToken) GetContent() string   { return t.RawContent }
func (t *KeywordToken) GetPosition() int     { return t.Position }

// NumericToken represents a numeric token in the parser.
type NumericToken struct {
	// Int64 is the integer value of the token.
	// This is used if the token represents an integer, or else it is zero.
	Int64 int64

	// Float64 is the floating-point value of the token.
	// This is used if the token represents a floating-point number, or else it is zero.
	Float64 float64

	// Floating indicates if the token represents a floating-point number.
	// If true, the token represents a floating-point number.
	// If false, the token represents an integer.
	Floating bool

	// RawContent is the raw string representation of the token.
	RawContent string

	// Position is the position of the token in the source string.
	Position int
}

func (*NumericToken) isToken() privateSealed { return privateSealed{} }

func (t *NumericToken) GetContent() string {
	return t.RawContent
}

func (t *NumericToken) GetPosition() int { return t.Position }

// BindingToken represents a binding token in the parser.
type BindingToken struct {
	// Index is the index of the binding token.
	// This is used for numeric binding tokens.
	Index int64

	// Name is the name of the binding token.
	// This is used for named binding tokens.
	Name string

	// Position is the position of the token in the source string.
	Position int
}

func (*BindingToken) isToken() privateSealed { return privateSealed{} }

func (t *BindingToken) GetContent() string {
	if t.Index != 0 {
		return "@" + strconv.FormatInt(t.Index, 10)
	}
	return "@" + t.Name
}

func (t *BindingToken) GetPosition() int { return t.Position }

// WhitespaceToken represents a whitespace token in the parser.
type WhitespaceToken struct {
	// Content is the actual whitespace of the token.
	Content string

	// Position is the position of the token in the source string.
	Position int
}

func (*WhitespaceToken) isToken() privateSealed { return privateSealed{} }
func (t *WhitespaceToken) GetContent() string   { return t.Content }
func (t *WhitespaceToken) GetPosition() int     { return t.Position }
