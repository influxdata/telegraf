package influx

import "strings"

const (
	escapes            = " ,="
	nameEscapes        = " ,"
	stringFieldEscapes = `\"`
)

var (
	escaper = strings.NewReplacer(
		`,`, `\,`,
		`"`, `\"`, // ???
		` `, `\ `,
		`=`, `\=`,
	)

	nameEscaper = strings.NewReplacer(
		`,`, `\,`,
		` `, `\ `,
	)

	stringFieldEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

func escape(s string) string {
	if strings.ContainsAny(s, escapes) {
		return escaper.Replace(s)
	} else {
		return s
	}
}

func nameEscape(s string) string {
	if strings.ContainsAny(s, nameEscapes) {
		return nameEscaper.Replace(s)
	} else {
		return s
	}
}

func stringFieldEscape(s string) string {
	if strings.ContainsAny(s, stringFieldEscapes) {
		return stringFieldEscaper.Replace(s)
	} else {
		return s
	}
}
