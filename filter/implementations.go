package filter

import (
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type filterSingle struct {
	s string
}

func (f *filterSingle) Match(s string) bool {
	return f.s == s
}

type filterNoGlob struct {
	m map[string]struct{}
}

func newFilterNoGlob(filters []string) Filter {
	out := filterNoGlob{m: make(map[string]struct{})}
	for _, filter := range filters {
		out.m[filter] = struct{}{}
	}
	return &out
}

func (f *filterNoGlob) Match(s string) bool {
	_, ok := f.m[s]
	return ok
}

// filterGlob handles glob patterns WITHOUT separators
// This is optimized for the common case where no separators are specified
type filterGlob struct {
	patterns []string
}

func newFilterGlob(filters []string) (Filter, error) {
	// Validate all patterns
	for _, pattern := range filters {
		if _, err := doublestar.Match(pattern, ""); err != nil {
			return nil, err
		}
	}

	return &filterGlob{
		patterns: filters,
	}, nil
}

func (f *filterGlob) Match(s string) bool {
	for _, pattern := range f.patterns {
		matched, err := doublestar.Match(pattern, s)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// filterGlobWithSeparators handles glob patterns WITH separators
// This is a separate implementation to avoid the performance cost of checking
// for separators in the hot path
type filterGlobWithSeparators struct {
	normalizedPatterns []string
	separators         []rune
}

func newFilterGlobWithSeparators(filters []string, separators []rune) (Filter, error) {
	// Validate all patterns
	for _, pattern := range filters {
		if _, err := doublestar.Match(pattern, ""); err != nil {
			return nil, err
		}
	}

	// Pre-compute normalized patterns
	normalizedPatterns := make([]string, len(filters))
	for i, pattern := range filters {
		normalizedPatterns[i] = normalizePattern(pattern, separators)
	}

	return &filterGlobWithSeparators{
		normalizedPatterns: normalizedPatterns,
		separators:         separators,
	}, nil
}

func (f *filterGlobWithSeparators) Match(s string) bool {
	// Normalize the input string once
	normalizedStr := normalizePattern(s, f.separators)

	for _, pattern := range f.normalizedPatterns {
		// Use PathMatch which treats '/' as a separator
		matched, err := doublestar.PathMatch(pattern, normalizedStr)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// normalizePattern converts all separators to '/' for path matching.
// This allows doublestar.PathMatch to treat them as path separators.
//
// IMPORTANT LIMITATION: This function cannot distinguish between literal
// separator characters and actual separators. For example, with separator '.',
// the pattern "foo.bar/baz.qux" will have ALL dots replaced with '/',
// even if some dots were meant to be literal characters in the pattern.
//
// This is an acceptable trade-off because:
//  1. The original gobwas/glob had the same limitation
//  2. It maintains backward compatibility with existing Telegraf configs
//  3. Users can work around this by using glob patterns like "foo?bar" instead of "foo.bar"
//     when they need to match a literal separator character
func normalizePattern(s string, separators []rune) string {
	if len(separators) == 0 {
		return s
	}

	result := s
	for _, sep := range separators {
		if sep != '/' {
			result = strings.ReplaceAll(result, string(sep), "/")
		}
	}
	return result
}
