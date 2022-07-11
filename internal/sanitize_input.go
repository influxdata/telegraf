package internal

import "strings"

func SanitizeInputByte(input []byte) []byte {
	return []byte(SanitizeInput(string(input)))
}

func SanitizeInput(input string) string {
	escaped := strings.Replace(input, "\n", "", -1)
	return strings.Replace(escaped, "\r", "", -1)
}
