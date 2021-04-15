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

// Scan returns the next token and literal value.
func (s *PointScanner) Scan() (Token, string) {
	// Read the next rune
	ch := s.read()
	if isWhitespace(ch) {
		return Ws, string(ch)
	} else if isLetter(ch) {
		return Letter, string(ch)
	} else if isNumber(ch) {
		return Number, string(ch)
	} else if isDelta(ch) {
		return Delta, string(ch)
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""
	case '\n':
		return Newline, string(ch)
	case '.':
		return Dot, string(ch)
	case '-':
		return MinusSign, string(ch)
	case '_':
		return Underscore, string(ch)
	case '/':
		return Slash, string(ch)
	case '\\':
		return Backslash, string(ch)
	case ',':
		return Comma, string(ch)
	case '"':
		return Quotes, string(ch)
	case '=':
		return Equals, string(ch)
	}
	return Illegal, string(ch)
}
