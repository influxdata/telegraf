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

// filterGlob handles both single and multiple glob patterns with optional separators
type filterGlob struct {
	patterns           []string
	normalizedPatterns []string // Pre-computed normalized patterns for performance
	separators         []rune
	hasSeparators      bool
}

func newFilterGlob(filters []string, separators ...rune) (Filter, error) {
	// Validate all patterns
	for _, pattern := range filters {
		if _, err := doublestar.Match(pattern, ""); err != nil {
			return nil, err
		}
	}

	filter := &filterGlob{
		patterns:      filters,
		separators:    separators,
		hasSeparators: len(separators) > 0,
	}

	// Pre-compute normalized patterns if separators are present
	if filter.hasSeparators {
		filter.normalizedPatterns = make([]string, len(filters))
		for i, pattern := range filters {
			filter.normalizedPatterns[i] = normalizePattern(pattern, separators)
		}
	}

	return filter, nil
}

func (f *filterGlob) Match(s string) bool {
	// Pre-normalize the input string once if we have separators
	var normalizedStr string
	if f.hasSeparators {
		normalizedStr = normalizePattern(s, f.separators)
	}

	for i, pattern := range f.patterns {
		var matched bool
		var err error

		if f.hasSeparators {
			// Use pre-computed normalized pattern
			matched, err = doublestar.PathMatch(f.normalizedPatterns[i], normalizedStr)
		} else {
			// Standard glob matching without path semantics
			matched, err = doublestar.Match(pattern, s)
		}

		// Continue on error but don't match
		if err != nil {
			continue
		}

		if matched {
			return true
		}
	}
	return false
}

// normalizePattern converts all separators to '/' for path matching
// This allows doublestar.PathMatch to treat them as path separators
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
