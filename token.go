package gqlparser

import "strconv"

// privateSealed is just a private type. it's used to limit patterns like sealed class.
type privateSealed struct{}

type Token interface {
	isToken() privateSealed
	GetContent() string
	GetPosition() int
}

type StringToken struct {
	Quote      byte
	Content    string
	RawContent string
	Position   int
}

func (*StringToken) isToken() privateSealed { return privateSealed{} }
func (t *StringToken) GetContent() string   { return t.RawContent }
func (t *StringToken) GetPosition() int     { return t.Position }

type OperatorToken struct {
	Type       string
	RawContent string
	Position   int
}

func (*OperatorToken) isToken() privateSealed { return privateSealed{} }

func (t *OperatorToken) GetContent() string {
	if t.RawContent != "" {
		return t.RawContent
	}
	return t.Type
}

func (t *OperatorToken) GetPosition() int { return t.Position }

type WildcardToken struct {
	Position int
}

func (*WildcardToken) isToken() privateSealed { return privateSealed{} }
func (t *WildcardToken) GetContent() string   { return "*" }
func (t *WildcardToken) GetPosition() int     { return t.Position }

type BooleanToken struct {
	Value      bool
	RawContent string
	Position   int
}

func (*BooleanToken) isToken() privateSealed { return privateSealed{} }

func (t *BooleanToken) GetContent() string {
	return t.RawContent
}

func (t *BooleanToken) GetPosition() int { return t.Position }

type OrderToken struct {
	Descending bool
	RawContent string
	Position   int
}

func (*OrderToken) isToken() privateSealed { return privateSealed{} }

func (t *OrderToken) GetContent() string {
	return t.RawContent
}

func (t *OrderToken) GetPosition() int { return t.Position }

type SymbolToken struct {
	Content  string
	Position int
}

func (*SymbolToken) isToken() privateSealed { return privateSealed{} }
func (t *SymbolToken) GetContent() string   { return t.Content }
func (t *SymbolToken) GetPosition() int     { return t.Position }

type KeywordToken struct {
	Name       string
	RawContent string
	Position   int
}

func (*KeywordToken) isToken() privateSealed { return privateSealed{} }
func (t *KeywordToken) GetContent() string   { return t.RawContent }
func (t *KeywordToken) GetPosition() int     { return t.Position }

type NumericToken struct {
	Int64      int64
	Float64    float64
	Floating   bool
	RawContent string
	Position   int
}

func (*NumericToken) isToken() privateSealed { return privateSealed{} }

func (t *NumericToken) GetContent() string {
	return t.RawContent
}

func (t *NumericToken) GetPosition() int { return t.Position }

type BindingToken struct {
	Index    int64
	Name     string
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

type WhitespaceToken struct {
	Content  string
	Position int
}

func (*WhitespaceToken) isToken() privateSealed { return privateSealed{} }
func (t *WhitespaceToken) GetContent() string   { return t.Content }
func (t *WhitespaceToken) GetPosition() int     { return t.Position }
