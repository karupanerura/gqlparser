package gqlparser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/karupanerura/runetrie"
)

var ErrEndOfToken = errors.New("end of token")

type Lexer struct {
	source   string
	position int
	buffer   []Token
}

var _ TokenSource = (*Lexer)(nil)

var (
	keywordTrie  = runetrie.Must(runetrie.NewCaseInsensitiveTrie[string]())
	operatorTrie = runetrie.Must(runetrie.NewCaseInsensitiveTrie[string]())
	orderTrie    = runetrie.Must(runetrie.NewCaseInsensitiveTrie[string]())
	booleanTrie  = runetrie.Must(runetrie.NewCaseInsensitiveTrie[string]())
)

func init() {
	_ = keywordTrie.Add(
		"SELECT",
		"FROM",
		"WHERE",
		"AGGREGATE",
		"OVER",
		"COUNT",
		"COUNT_UP_TO",
		"SUM",
		"AVG",
		"AS",
		"DISTINCT",
		"ON",
		"ORDER",
		"BY",
		"LIMIT",
		"FIRST",
		"OFFSET",
		"KEY",
		"PROJECT",
		"NAMESPACE",
		"ARRAY",
		"BLOB",
		"DATETIME",
		"NULL",
	)
	_ = operatorTrie.Add("AND", "OR", "IS", "CONTAINS", "HAS", "ANCESTOR", "IN", "NOT", "DESCENDANT")
	_ = orderTrie.Add("DESC", "ASC")
	_ = booleanTrie.Add("TRUE", "FALSE")
}

func NewLexer(source string) *Lexer {
	return &Lexer{source: source}
}

func (l *Lexer) Next() bool {
	return len(l.buffer) != 0 || l.position != len(l.source)
}

func (l *Lexer) Read() (Token, error) {
	if len(l.buffer) != 0 {
		token := l.buffer[len(l.buffer)-1]
		l.buffer = l.buffer[0 : len(l.buffer)-1]
		return token, nil
	}
	if l.position == len(l.source) {
		return nil, ErrEndOfToken
	}

	switch l.source[l.position] {
	case ' ', '\t', '\r', '\n': // isWhitespace
		// skip whitespace
		pos := l.position
		var ws strings.Builder
		for {
			ws.WriteByte(l.source[l.position])
			l.position++
			if l.position == len(l.source) || !isWhitespace(l.source[l.position]) {
				break
			}
		}
		return &WhitespaceToken{Content: ws.String(), Position: pos}, nil

	case '@':
		t, w, err := takeBindingToken(l.source[l.position:], l.position)
		if err != nil {
			return nil, err
		}
		l.position += w
		return t, nil

	case '`', '\'', '"':
		t, w, err := takeQuotedStringToken(l.source[l.position:], l.position)
		if err != nil {
			return nil, err
		}
		l.position += w
		return t, nil

	case '(', ',', ')', '=':
		t := &OperatorToken{Type: l.source[l.position : l.position+1], Position: l.position}
		l.position++
		return t, nil

	case '<', '>', '!':
		t, w, err := takeOperatorToken(l.source[l.position:], l.position)
		if err != nil {
			return nil, err
		}
		l.position += w
		return t, nil

	case '*':
		t := &WildcardToken{Position: l.position}
		l.position++
		return t, nil

	case '-', '+', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		t, w, err := takeNumericToken(l.source[l.position:], l.position)
		if err != nil {
			return nil, err
		}
		l.position += w
		return t, nil

	default:
		if v, ok := keywordTrie.LongestMatchPrefixOf(l.source[l.position:]); ok {
			t := &KeywordToken{Name: v, RawContent: l.source[l.position : l.position+len(v)], Position: l.position}
			l.position += len(v)
			return t, nil
		} else if v, ok := operatorTrie.LongestMatchPrefixOf(l.source[l.position:]); ok {
			t := &OperatorToken{Type: v, RawContent: l.source[l.position : l.position+len(v)], Position: l.position}
			l.position += len(v)
			return t, nil
		} else if v, ok := orderTrie.LongestMatchPrefixOf(l.source[l.position:]); ok {
			t := &OrderToken{Descending: v == "DESC", RawContent: l.source[l.position : l.position+len(v)], Position: l.position}
			l.position += len(v)
			return t, nil
		} else if v, ok := booleanTrie.LongestMatchPrefixOf(l.source[l.position:]); ok {
			t := &BooleanToken{Value: v == "TRUE", RawContent: l.source[l.position : l.position+len(v)], Position: l.position}
			l.position += len(v)
			return t, nil
		} else {
			return l.takeSymbolToken()
		}
	}
}

func (l *Lexer) takeSymbolToken() (Token, error) {
	t, w, err := takeSymbolToken(l.source[l.position:], l.position)
	if err != nil {
		return nil, err
	}
	l.position += w
	return t, nil
}

func (l *Lexer) Unread(t Token) {
	l.buffer = append(l.buffer, t)
}

func isWhitespace(r byte) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func ReadAllTokens(ts TokenSource) ([]Token, error) {
	var tokens []Token
	for ts.Next() {
		tok, err := ts.Read()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

func takeQuotedStringToken(s string, pos int) (*StringToken, int, error) {
	quote := s[0]
	begins := 1
	ends := 0
	needsUnescape := false
	for i := begins; i != len(s); i++ {
		if s[i] == quote {
			ends = i
			break
		}
		if s[i] == '\\' {
			i++
			if i == len(s) {
				return nil, 0, fmt.Errorf("unexpected token: \\")
			}
			needsUnescape = true
		}
	}
	if ends == 0 {
		return nil, 0, fmt.Errorf("unexpected token: %c", quote)
	}
	content := s[begins:ends]
	if needsUnescape {
		content = unquote(content)
	}
	return &StringToken{
		Quote:      quote,
		Content:    content,
		RawContent: s[0 : ends+1],
		Position:   pos,
	}, begins + ends, nil
}

func takeOperatorToken(s string, pos int) (*OperatorToken, int, error) {
	if len(s) == 1 || s[1] != '=' {
		return &OperatorToken{Type: s[0:1], Position: pos}, 1, nil
	}
	return &OperatorToken{Type: s[0:2], Position: pos}, 2, nil
}

func takeBindingToken(s string, pos int) (*BindingToken, int, error) {
	if len(s) == 1 {
		return nil, 0, fmt.Errorf("unexpected token: %c", s[0])
	}

	width := 1
	numeric := false
	switch s[width] {
	case '0':
		return nil, 0, fmt.Errorf("unexpected token: %s (invalid binding site)", s[0:width])
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		numeric = true
		for '0' <= s[width] && s[width] <= '9' {
			width++
			if width == len(s) {
				break
			}
		}
	default:
		for isSymbolByte(s[width]) {
			width++
			if width == len(s) {
				break
			}
		}
	}

	if numeric {
		n, err := strconv.ParseInt(s[1:width], 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("unexpected token: %s (%w)", s[0:width], err)
		}
		return &BindingToken{Index: n, Position: pos}, width, nil
	} else {
		return &BindingToken{Name: s[1:width], Position: pos}, width, nil
	}
}

func takeNumericToken(s string, pos int) (Token, int, error) {
	width := 0
	float := false
	for '0' <= s[width] && s[width] <= '9' || s[width] == '.' || s[width] == '-' || s[width] == '+' {
		if s[width] == '.' {
			float = true
		}

		width++
		if width == len(s) {
			break
		}
	}

	// it's a special case for a single '+' character
	if width == 1 && s[0] == '+' {
		return &OperatorToken{Type: "+", RawContent: "+", Position: pos}, 1, nil
	}

	if float {
		n, err := strconv.ParseFloat(s[:width], 64)
		if err != nil {
			return nil, 0, fmt.Errorf("unexpected token: %s (%w)", s[:width], err)
		}
		return &NumericToken{Float64: n, Floating: true, RawContent: s[:width], Position: pos}, width, nil
	} else {
		n, err := strconv.ParseInt(s[:width], 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("unexpected token: %s (%w)", s[:width], err)
		}
		return &NumericToken{Int64: n, Floating: false, RawContent: s[:width], Position: pos}, width, nil
	}
}

func takeSymbolToken(s string, pos int) (*SymbolToken, int, error) {
	width := 0
	for isSymbolByte(s[width]) {
		width++
		if width == len(s) {
			return &SymbolToken{Content: s[:width], Position: pos}, width, nil
		}
	}
	if width == 0 {
		return nil, 0, fmt.Errorf("unexpected token: %c", s[width])
	}
	for s[width] == '.' {
		width++
		if width == len(s) {
			return &SymbolToken{Content: s[:width], Position: pos}, width, nil
		}
		base := width
		for isSymbolFollowingByte(s[width]) {
			width++
			if width == len(s) {
				return &SymbolToken{Content: s[:width], Position: pos}, width, nil
			}
		}
		if width == base {
			return nil, 0, fmt.Errorf("unexpected token: %c", s[width])
		}
	}

	return &SymbolToken{Content: s[:width], Position: pos}, width, nil
}

// isSymbolByte matches the regular expression `[a-zA-Z0-9_$]`.
func isSymbolByte(b byte) bool {
	if b > unicode.MaxASCII {
		return false
	}
	return unicode.IsLetter(rune(b)) || unicode.IsDigit(rune(b)) || b == '_' || b == '$'
}

// isSymbolFollowingByte matches the regular expression `[0-9_$]`.
func isSymbolFollowingByte(b byte) bool {
	if b > unicode.MaxASCII {
		return false
	}
	return unicode.IsDigit(rune(b)) || b == '_' || b == '$'
}

var unquoteReplacer = strings.NewReplacer(
	"\\\\", "\\",
	"\\0", "\u0000", // NULL
	"\\b", "\b",
	"\\n", "\n",
	"\\r", "\r",
	"\\t", "\t",
	"\\Z", "\u001A", // SUBSTITUTE
	"\\'", "'",
	"\\\"", "\"",
	"\\`", "`",
	"\\%", "%",
	"\\_", "_",
)

func unquote(s string) string {
	return unquoteReplacer.Replace(s)
}
