package gqlparser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/karupanerura/runetrie"
)

// ErrEndOfToken is returned when the lexer reaches the end of the source string.
// It is used to indicate that there are no more tokens to read.
// This error is not returned by the lexer itself, but can be used by clients
// to check if they have reached the end of the token stream.
var ErrEndOfToken = errors.New("end of token")

// Lexer is a GQL lexer.
// It implements the TokenSource interface.
type Lexer struct {
	source   string
	position int
	buffer   []Token
}

var _ TokenSource = (*Lexer)(nil)

type keywordKind uint8

const (
	syntaxKeywordKind keywordKind = iota
	operatorKeywordKind
	orderKeywordKind
	booleanKeywordKind
)

var (
	keywordTrie    = runetrie.Must(runetrie.NewCaseInsensitiveTrie[string]())
	keywordKindMap = map[string]keywordKind{
		"SELECT":      syntaxKeywordKind,
		"FROM":        syntaxKeywordKind,
		"WHERE":       syntaxKeywordKind,
		"AGGREGATE":   syntaxKeywordKind,
		"OVER":        syntaxKeywordKind,
		"COUNT":       syntaxKeywordKind,
		"COUNT_UP_TO": syntaxKeywordKind,
		"SUM":         syntaxKeywordKind,
		"AVG":         syntaxKeywordKind,
		"AS":          syntaxKeywordKind,
		"DISTINCT":    syntaxKeywordKind,
		"ON":          syntaxKeywordKind,
		"ORDER":       syntaxKeywordKind,
		"BY":          syntaxKeywordKind,
		"LIMIT":       syntaxKeywordKind,
		"FIRST":       syntaxKeywordKind,
		"OFFSET":      syntaxKeywordKind,
		"KEY":         syntaxKeywordKind,
		"PROJECT":     syntaxKeywordKind,
		"NAMESPACE":   syntaxKeywordKind,
		"ARRAY":       syntaxKeywordKind,
		"BLOB":        syntaxKeywordKind,
		"DATETIME":    syntaxKeywordKind,
		"NULL":        syntaxKeywordKind,
		"AND":         operatorKeywordKind,
		"OR":          operatorKeywordKind,
		"IS":          operatorKeywordKind,
		"CONTAINS":    operatorKeywordKind,
		"HAS":         operatorKeywordKind,
		"ANCESTOR":    operatorKeywordKind,
		"IN":          operatorKeywordKind,
		"NOT":         operatorKeywordKind,
		"DESCENDANT":  operatorKeywordKind,
		"DESC":        orderKeywordKind,
		"ASC":         orderKeywordKind,
		"TRUE":        booleanKeywordKind,
		"FALSE":       booleanKeywordKind,
	}
)

func init() {
	for k := range keywordKindMap {
		_ = keywordTrie.Add(k)
	}
}

// NewLexer creates a new Lexer instance.
// It takes a source string as input and initializes the lexer with it.
func NewLexer(source string) *Lexer {
	return &Lexer{source: source}
}

// Next returns true if there are more tokens to read.
func (l *Lexer) Next() bool {
	return len(l.buffer) != 0 || l.position != len(l.source)
}

// Read reads the next token from the source string.
// It returns the token and an error if any occurs.
// If there are no more tokens to read, it returns ErrEndOfToken.
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

	case '(', ',', ')', '=', '.':
		t := &OperatorToken{Type: l.source[l.position : l.position+1], Position: l.position}
		l.position++
		return t, nil

	case '<', '>', '!':
		t, w := takeOperatorToken(l.source[l.position:], l.position)
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
		v, ok := keywordTrie.LongestMatchPrefixOf(l.source[l.position:])
		if !ok {
			return l.takeSymbolToken()
		}

		switch keywordKindMap[v] {
		case syntaxKeywordKind:
			t := &KeywordToken{Name: v, RawContent: l.source[l.position : l.position+len(v)], Position: l.position}
			l.position += len(v)
			return t, nil
		case operatorKeywordKind:
			t := &OperatorToken{Type: v, RawContent: l.source[l.position : l.position+len(v)], Position: l.position}
			l.position += len(v)
			return t, nil
		case orderKeywordKind:
			t := &OrderToken{Descending: v == "DESC", RawContent: l.source[l.position : l.position+len(v)], Position: l.position}
			l.position += len(v)
			return t, nil
		case booleanKeywordKind:
			t := &BooleanToken{Value: v == "TRUE", RawContent: l.source[l.position : l.position+len(v)], Position: l.position}
			l.position += len(v)
			return t, nil
		default:
			panic(fmt.Errorf("unknown keyword: %s", v))
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

// Unread un-reads the last read token.
// This allows for backtracking in the token stream.
// It should be called after Read() if you want to go back to the previous token.
func (l *Lexer) Unread(t Token) {
	l.buffer = append(l.buffer, t)
}

func isWhitespace(r byte) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
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

func takeOperatorToken(s string, pos int) (*OperatorToken, int) {
	if len(s) == 1 || s[1] != '=' {
		return &OperatorToken{Type: s[0:1], Position: pos}, 1
	}
	return &OperatorToken{Type: s[0:2], Position: pos}, 2
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
		for width < len(s) {
			r, size := utf8.DecodeRuneInString(s[width:])
			if r == utf8.RuneError || !isSymbol(r) {
				break
			}
			width += size
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
		return &OperatorToken{Type: "+", Position: pos}, 1, nil
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
	for width < len(s) {
		r, size := utf8.DecodeRuneInString(s[width:])
		if r == utf8.RuneError || !isSymbol(r) {
			break
		}
		width += size
	}
	if width == 0 {
		return nil, 0, fmt.Errorf("unexpected token: %c", s[width])
	}
	return &SymbolToken{Content: s[:width], Position: pos}, width, nil
}

// isSymbol matches the byte of a symbol.
// A symbol is a sequence of letters, digits, underscores, dollar signs, or unicode characters in the range from U+0080 to U+FFFF (inclusive),
// so long as the name does not begin with a digit.
// For example, foo, bar17, x_y, big$bux, __qux
func isSymbol(b rune) bool {
	return unicode.IsLetter(b) || unicode.IsDigit(b) || b == '_' || b == '$' || (b > unicode.MaxASCII && unicode.IsPrint(b))
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
