package wavefront

type Token int

const (
	// Special tokens
	Illegal Token = iota
	EOF
	Ws

	// Literals
	literalBeg
	Letter // metric name, source/point tags
	Number
	MinusSign
	Underscore
	Dot
	Slash
	Backslash
	Comma
	Delta
	literalEnd

	// Misc characters
	Quotes
	Equals
	Newline
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

func isDelta(ch rune) bool {
	return ch == '\u2206' || ch == '\u0394'
}

var eof = rune(0)
