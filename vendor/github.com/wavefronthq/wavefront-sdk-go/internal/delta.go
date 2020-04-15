package internal

import "strings"

const (
	deltaPrefix    = "\u2206"
	altDeltaPrefix = "\u0394"
)

func HasDeltaPrefix(name string) bool {
	return strings.HasPrefix(name, deltaPrefix) || strings.HasPrefix(name, altDeltaPrefix)
}

// Gets a delta counter name prefixed with ∆.
func DeltaCounterName(name string) string {
	if HasDeltaPrefix(name) {
		return name
	}
	return deltaPrefix + name
}
