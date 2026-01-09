package common

import (
	"strconv"
	"strings"
)

// SetTagIfUsed sets those tags whose value is different from empty string.
func SetTagIfUsed(m map[string]string, k, v string) {
	if v == "N/A" || v == "" || v == "Requested functionality has been deprecated" {
		return
	}
	m[k] = v
}

// SetIfUsed sets those fields whose value is different from empty string.
func SetIfUsed(t string, m map[string]any, k, v string) {
	if v == "N/A" || v == "" || v == "Requested functionality has been deprecated" {
		return
	}

	vals := strings.Fields(v)
	if len(vals) < 1 {
		return
	}
	val := vals[0]

	if k == "pcie_link_width_current" {
		val = strings.TrimSuffix(val, "x")
	}

	switch t {
	case "float":
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			m[k] = f
		}
	case "int":
		i, err := strconv.Atoi(val)
		if err == nil {
			m[k] = i
		}
	case "str":
		m[k] = val
	}
}
