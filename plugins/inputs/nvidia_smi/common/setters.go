package common

import (
	"strconv"
	"strings"
)

// SetTagIfUsed sets those tags whose value is different from empty string.
func SetTagIfUsed(m map[string]string, k, v string) {
	if v != "" {
		m[k] = v
	}
}

// SetIfUsed sets those fields whose value is different from empty string.
func SetIfUsed(t string, m map[string]interface{}, k, v string) {
	v = strings.TrimSpace(v)
	if v == "" || v == "N/A" {
		return
	}

	switch t {
	case "float":
		if f, err := strconv.ParseFloat(number(v), 64); err == nil {
			m[k] = f
		}
	case "int":
		if i, err := strconv.ParseInt(number(v), 10, 64); err == nil {
			m[k] = i
		}
	case "str":
		m[k] = v
	}
}

// number strips the unit from a measurement such as "20475 MiB" or "16x".
func number(v string) string {
	return strings.TrimSuffix(strings.Fields(v)[0], "x")
}
