package rfc5424

// unsafeUTF8DecimalCodePointsToInt converts a slice containing
// a series of UTF-8 decimal code points into their integer rapresentation.
//
// It assumes input code points are in the range 48-57.
// Returns a pointer since an empty slice is equal to nil and not to the zero value of the codomain (ie., `int`).
func unsafeUTF8DecimalCodePointsToInt(chars []uint8) int {
	out := 0
	ord := 1
	for i := len(chars) - 1; i >= 0; i-- {
		curchar := int(chars[i])
		out += (curchar - '0') * ord
		ord *= 10
	}
	return out
}

// escape adds a backslash to \, ], " characters
func escape(value string) string {
	res := ""
	for i, c := range value {
		if c == 92 || c == 93 || c == 34 {
			res += `\`
		}
		res += string(value[i])
	}

	return res
}

// rmchars remove byte at given positions starting from offset
func rmchars(data []byte, positions []int, offset int) []byte {
	// We need a copy here to not modify original data
	cp := append([]byte(nil), data...)
	for i, pos := range positions {
		at := pos - i - offset
		cp = append(cp[:at], cp[(at+1):]...)
	}
	return cp
}
