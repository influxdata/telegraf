package procfilter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const eof = rune(0)
const eol = rune('\n')

type tokenType int

// Token Types
const (
	tTEOF     = iota
	tTString  // " or ' delimited
	tTRegexp  // string followd by a r
	tTComment // # .... eol
	tTWhitespace
	tTNumber
	tTIdentifier // [_.A-Za-z0-9]+
	tTLeftPar    // (
	tTRightPar   // )
	tTLeftArrow  // <-
	tTEqual      // =
	tTComma      // ,
	tTBang       // !
	tTIllegal
)

// Scanner represents a lexical scanner.
type Scanner struct {
	r       *bufio.Reader
	pos     int
	prevPos int // previous line last scanned position
	line    int
	last    rune
	stPos   int // current token start pos
	stLine  int // current toketn start line
}

// NewScanner returns a new instance of Scanner.
func newScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value.
func (s *Scanner) scan() (tok tokenType, lit string) {
	s.stPos = s.pos
	s.stLine = s.line
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdentifier()
	} else if isDigit(ch) {
		s.unread()
		return s.scanNumber()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return tTEOF, ""
	case '#':
		return s.scanComment()
	case '"':
		return s.scanString('"')
	case '\'':
		return s.scanString('\'')
	case '(':
		return tTLeftPar, string(ch)
	case ')':
		return tTRightPar, string(ch)
	case ',':
		return tTComma, string(ch)
	case '=':
		return tTEqual, string(ch)
	case '!':
		return tTBang, string(ch)
	case '<':
		ch := s.read()
		if ch == '-' {
			return tTLeftArrow, "<-"
		}
		return tTIllegal, string(ch)
	}

	return tTIllegal, string(ch)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok tokenType, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return tTWhitespace, buf.String()
}

// scanComment consumes one comment (until end of line)
func (s *Scanner) scanComment() (tok tokenType, lit string) {
	var buf bytes.Buffer

	for {
		if ch := s.read(); ch == eof || ch == eol {
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return tTComment, buf.String()
}

// scanString consumes one string and its optional r suffix
func (s *Scanner) scanString(d rune) (tok tokenType, lit string) {
	var buf bytes.Buffer

	for {
		ch := s.read()
		if ch == d {
			// TODO add handling of \ ?
			nch := s.read()
			if nch == 'r' {
				// the r suffix denotes a regexp
				return tTRegexp, buf.String()
			} else {
				s.unread()
			}
			return tTString, buf.String()
		} else if ch == eof || ch == eol {
			return tTIllegal, buf.String()
		} else {
			buf.WriteRune(ch)
		}
	}

	return tTIllegal, buf.String() // unreachable code
}

// scanIdent consumes the current rune and all contiguous ident runes ([._a-zA-Z0-9]).
func (s *Scanner) scanIdentifier() (tok tokenType, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOL/EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof || ch == eol {
			s.unread()
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' && ch != '.' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular identifier.
	id := buf.String()
	fmt.Sprintf("id:%q", id)
	return tTIdentifier, id
}

// scanIdent consumes a number.
func (s *Scanner) scanNumber() (tok tokenType, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); isDigit(ch) {
			_, _ = buf.WriteRune(ch)
		} else {
			s.unread()
			return tTNumber, buf.String()
		}
	}
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	//fmt.Printf("(%d:%d)'%q' ", s.line, s.pos, ch)
	s.last = ch
	s.pos++
	if ch == eol {
		s.line++
		s.prevPos = s.pos // store in case of unread of the eol
		s.pos = 0
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
	if s.last == eol {
		s.line--
		s.pos = s.prevPos
	} else {
		s.pos--
	}
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }

// posInfo returns a string describing the scanner position
func (s *Scanner) posInfo() string {
	return fmt.Sprintf("line %d char %d", s.stLine+1, s.stPos+1) // humans count lines and chars starting from 1
}
