// This code is based on encoding/json and gorilla/schema

package encoding

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

var (
	Tags []string
)

const (
	TagName     = "gorethink"
	JSONTagName = "json"
	RefTagName  = "gorethink_ref"
)

// tagOptions is the string following a comma in a struct field's
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

func getTag(sf reflect.StructField) string {
	if Tags == nil {
		return sf.Tag.Get(TagName)
	}

	for _, tagName := range Tags {
		if tag := sf.Tag.Get(tagName); tag != "" {
			return tag
		}
	}

	return ""
}

func getRefTag(sf reflect.StructField) string {
	return sf.Tag.Get(RefTagName)
}

// parseTag splits a struct field's tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

func parseCompoundIndex(tag string) (string, int, bool) {
	lIdx := strings.Index(tag, "[")
	rIdx := strings.Index(tag, "]")
	if lIdx > 1 && rIdx > lIdx+1 {
		if elemIndex_, err := strconv.ParseInt(tag[lIdx+1:rIdx], 10, 64); err == nil {
			return tag[:lIdx], int(elemIndex_), true
		}
	}

	return tag, 0, false
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

// Contains returns whether checks that a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}
