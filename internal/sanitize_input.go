package internal

import "strings"

func SanitizeInput(input string) string {
	escaped := strings.Replace(input, "\n", " ", -1)
	return strings.Replace(escaped, "\r", "", -1)
}
