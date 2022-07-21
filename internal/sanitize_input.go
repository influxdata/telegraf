package internal

import "strings"

func SanitizeArgs(args []interface{}) []interface{} {
	var sanitizedArgs []interface{}
	for _, a := range args {
		switch t := a.(type) {
		case string:
			sanitizedArgs = append(sanitizedArgs, SanitizeInput(t))
		case []byte:
			sanitizedArgs = append(sanitizedArgs, SanitizeInput(string(t)))
		default:
			sanitizedArgs = append(sanitizedArgs, t)
		}
	}
	return sanitizedArgs
}

func SanitizeInput(input string) string {
	escaped := strings.Replace(input, "\n", " ", -1)
	return strings.Replace(escaped, "\r", "", -1)
}
