package wavefront

import (
	"bufio"
	"io"
)

// Lexical Point Scanner
type PointScanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *PointScanner {
	return &PointScanner{r: bufio.NewReader(r)}
}

// read reads the next rune from the buffered reader.
// Returns rune(0) if an error occurs (or io.EOF is returned).
func (s *PointScanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *PointScanner) unread() {
	_ = s.r.UnreadRune()
}

// Scan returns the next token and literal value.
func (s *PointScanner) Scan() (Token, string) {

	// Read the next rune
	ch := s.read()
	if isWhitespace(ch) {
		return WS, string(ch)
	} else if isLetter(ch) {
		return LETTER, string(ch)
	} else if isNumber(ch) {
		return NUMBER, string(ch)
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""
	case '\n':
		return NEWLINE, string(ch)
	case '.':
		return DOT, string(ch)
	case '-':
		return MINUS_SIGN, string(ch)
	case '_':
		return UNDERSCORE, string(ch)
	case '/':
		return SLASH, string(ch)
	case '\\':
		return BACKSLASH, string(ch)
	case ',':
		return COMMA, string(ch)
	case '"':
		return QUOTES, string(ch)
	case '=':
		return EQUALS, string(ch)
	}
	return ILLEGAL, string(ch)
}
