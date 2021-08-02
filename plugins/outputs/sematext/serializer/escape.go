// logic copied from plugins/serializers/influx/escape.go
package serializer

import "strings"

const (
	escapes     = "\t\n\f\r ,="
	nameEscapes = "\t\n\f\r ,"
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
)

// Escape a tagkey, tagvalue, or fieldkey
func escape(s string) string {
	if strings.ContainsAny(s, escapes) {
		return escaper.Replace(s)
	}
	return s
}

// Escape a measurement name
func nameEscape(s string) string {
	if strings.ContainsAny(s, nameEscapes) {
		return nameEscaper.Replace(s)
	}
	return s
}
