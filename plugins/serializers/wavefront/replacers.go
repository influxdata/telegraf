package wavefront

import "strings"

// catch many of the invalid chars that could appear in a metric or tag name
var sanitizedChars = strings.NewReplacer(
	"!", "-", "@", "-", "#", "-", "$", "-", "%", "-", "^", "-", "&", "-",
	"*", "-", "(", "-", ")", "-", "+", "-", "`", "-", "'", "-", "\"", "-",
	"[", "-", "]", "-", "{", "-", "}", "-", ":", "-", ";", "-", "<", "-",
	">", "-", ",", "-", "?", "-", "/", "-", "\\", "-", "|", "-", " ", "-",
	"=", "-",
)

// catch many of the invalid chars that could appear in a metric or tag name
var strictSanitizedChars = strings.NewReplacer(
	"!", "-", "@", "-", "#", "-", "$", "-", "%", "-", "^", "-", "&", "-",
	"*", "-", "(", "-", ")", "-", "+", "-", "`", "-", "'", "-", "\"", "-",
	"[", "-", "]", "-", "{", "-", "}", "-", ":", "-", ";", "-", "<", "-",
	">", "-", "?", "-", "\\", "-", "|", "-", " ", "-", "=", "-",
)

var tagValueReplacer = strings.NewReplacer("\"", "\\\"", "*", "-")

var pathReplacer = strings.NewReplacer("_", ".")

func Sanitize(strict bool, val string) string {
	if strict {
		return strictSanitizedChars.Replace(val)
	}
	return sanitizedChars.Replace(val)
}
