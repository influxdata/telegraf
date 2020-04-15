package k8s

import (
	"regexp"
	"strings"
)

const (
	qnameCharFmt           string = "[A-Za-z0-9]"
	qnameExtCharFmt        string = "[-A-Za-z0-9_./]"
	qualifiedNameFmt       string = "(" + qnameCharFmt + qnameExtCharFmt + "*)?" + qnameCharFmt
	qualifiedNameMaxLength int    = 63
	labelValueFmt          string = "(" + qualifiedNameFmt + ")?"
)

var labelValueRegexp = regexp.MustCompile("^" + labelValueFmt + "$")

// LabelSelector represents a Kubernetes label selector.
//
// Any values that don't conform to Kubernetes label value restrictions
// will be silently dropped.
//
//		l := new(k8s.LabelSelector)
//		l.Eq("component", "frontend")
//		l.In("type", "prod", "staging")
//
type LabelSelector struct {
	stmts []string
}

func (l *LabelSelector) Selector() Option {
	return QueryParam("labelSelector", l.String())
}

func (l *LabelSelector) String() string {
	return strings.Join(l.stmts, ",")
}

func validLabelValue(s string) bool {
	if len(s) > 63 || len(s) == 0 {
		return false
	}
	return labelValueRegexp.MatchString(s)
}

// Eq selects labels which have the key and the key has the provide value.
func (l *LabelSelector) Eq(key, val string) {
	if !validLabelValue(key) || !validLabelValue(val) {
		return
	}
	l.stmts = append(l.stmts, key+"="+val)
}

// NotEq selects labels where the key is present and has a different value
// than the value provided.
func (l *LabelSelector) NotEq(key, val string) {
	if !validLabelValue(key) || !validLabelValue(val) {
		return
	}
	l.stmts = append(l.stmts, key+"!="+val)
}

// In selects labels which have the key and the key has one of the provided values.
func (l *LabelSelector) In(key string, vals ...string) {
	if !validLabelValue(key) || len(vals) == 0 {
		return
	}
	for _, val := range vals {
		if !validLabelValue(val) {
			return
		}
	}
	l.stmts = append(l.stmts, key+" in ("+strings.Join(vals, ", ")+")")
}

// NotIn selects labels which have the key and the key is not one of the provided values.
func (l *LabelSelector) NotIn(key string, vals ...string) {
	if !validLabelValue(key) || len(vals) == 0 {
		return
	}
	for _, val := range vals {
		if !validLabelValue(val) {
			return
		}
	}
	l.stmts = append(l.stmts, key+" notin ("+strings.Join(vals, ", ")+")")
}
