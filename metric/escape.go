package metric

import (
	"strings"
)

var (
	// escaper is for escaping:
	//   - tag keys
	//   - tag values
	//   - field keys
	// see https://docs.influxdata.com/influxdb/v1.0/write_protocols/line_protocol_tutorial/#special-characters-and-keywords
	escaper   = strings.NewReplacer(`,`, `\,`, `"`, `\"`, ` `, `\ `, `=`, `\=`)
	unEscaper = strings.NewReplacer(`\,`, `,`, `\"`, `"`, `\ `, ` `, `\=`, `=`)

	// nameEscaper is for escaping measurement names only.
	// see https://docs.influxdata.com/influxdb/v1.0/write_protocols/line_protocol_tutorial/#special-characters-and-keywords
	nameEscaper   = strings.NewReplacer(`,`, `\,`, ` `, `\ `)
	nameUnEscaper = strings.NewReplacer(`\,`, `,`, `\ `, ` `)

	// stringFieldEscaper is for escaping string field values only.
	// see https://docs.influxdata.com/influxdb/v1.0/write_protocols/line_protocol_tutorial/#special-characters-and-keywords
	stringFieldEscaper   = strings.NewReplacer(`"`, `\"`)
	stringFieldUnEscaper = strings.NewReplacer(`\"`, `"`)
)

func escape(s string, t string) string {
	switch t {
	case "fieldkey", "tagkey", "tagval":
		return escaper.Replace(s)
	case "name":
		return nameEscaper.Replace(s)
	case "fieldval":
		return stringFieldEscaper.Replace(s)
	}
	return s
}

func unescape(s string, t string) string {
	switch t {
	case "fieldkey", "tagkey", "tagval":
		return unEscaper.Replace(s)
	case "name":
		return nameUnEscaper.Replace(s)
	case "fieldval":
		return stringFieldUnEscaper.Replace(s)
	}
	return s
}
