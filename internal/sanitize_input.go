package internal

import "strings"

func SanitizeArgs(args []interface{}) []interface{} {
	var sanitizedArgs []interface{}
	for _, a := range args {
		switch t := a.(type) {
		case string:
			sanitizedArgs = append(sanitizedArgs, sanitizeInput(t))
		case []byte:
			sanitizedArgs = append(sanitizedArgs, sanitizeInput(string(t)))
		default:
			sanitizedArgs = append(sanitizedArgs, t)
		}
	}
	return sanitizedArgs
}

func sanitizeInput(input string) string {
	escaped := strings.Replace(input, "\n", " ", -1)
	return strings.Replace(escaped, "\r", "", -1)
}
