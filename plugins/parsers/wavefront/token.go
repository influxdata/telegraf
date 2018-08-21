package wavefront

type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	literal_beg
	LETTER // metric name, source/point tags
	NUMBER
	MINUS_SIGN
	UNDERSCORE
	DOT
	SLASH
	BACKSLASH
	COMMA
	literal_end

	// Misc characters
	QUOTES
	EQUALS
	NEWLINE
)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isNumber(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

var eof = rune(0)
