package influx

import "strings"

const (
	escapes            = "\t\n\f\r ,="
	nameEscapes        = "\t\n\f\r ,"
	stringFieldEscapes = "\t\n\f\r\\\""
)

var (
	escaper = strings.NewReplacer(
		"\t", `\t`,
		"\n", `\n`,
		"\f", `\f`,
		"\r", `\r`,
		`,`, `\,`,
		` `, `\ `,
		`=`, `\=`,
	)

	nameEscaper = strings.NewReplacer(
		"\t", `\t`,
		"\n", `\n`,
		"\f", `\f`,
		"\r", `\r`,
		`,`, `\,`,
		` `, `\ `,
	)

	stringFieldEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

// Escape a tagkey, tagvalue, or fieldkey
func escape(s string) string {
	if strings.ContainsAny(s, escapes) {
		return escaper.Replace(s)
	} else {
		return s
	}
}

// Escape a measurement name
func nameEscape(s string) string {
	if strings.ContainsAny(s, nameEscapes) {
		return nameEscaper.Replace(s)
	} else {
		return s
	}
}

// Escape a string field
func stringFieldEscape(s string) string {
	if strings.ContainsAny(s, stringFieldEscapes) {
		return stringFieldEscaper.Replace(s)
	} else {
		return s
	}
}
