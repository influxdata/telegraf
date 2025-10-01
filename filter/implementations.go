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
	needsSlashEscape   bool // Pre-computed: true if we need to preserve literal slashes
}

func newFilterGlobWithSeparators(filters []string, separators []rune) (Filter, error) {
	// Validate all patterns
	for _, pattern := range filters {
		if _, err := doublestar.Match(pattern, ""); err != nil {
			return nil, err
		}
	}

	// Pre-compute whether we need to escape slashes
	// This is needed when '/' is NOT one of the separators
	needsSlashEscape := true
	for _, sep := range separators {
		if sep == '/' {
			needsSlashEscape = false
			break
		}
	}

	// Pre-compute normalized patterns
	normalizedPatterns := make([]string, len(filters))
	for i, pattern := range filters {
		normalizedPatterns[i] = normalizePattern(pattern, separators, needsSlashEscape)
	}

	return &filterGlobWithSeparators{
		normalizedPatterns: normalizedPatterns,
		separators:         separators,
		needsSlashEscape:   needsSlashEscape,
	}, nil
}

func (f *filterGlobWithSeparators) Match(s string) bool {
	// Normalize the input string once using pre-computed flags
	normalizedStr := normalizePattern(s, f.separators, f.needsSlashEscape)

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

// normalizePattern converts all separators to '/' for path matching while preserving
// literal slashes when they are not separators.
//
// This allows doublestar.PathMatch to treat custom separators as path separators.
//
// The normalization process:
//  1. If '/' is NOT a separator (needsSlashEscape=true), replace all literal '/' with
//     U+FFFD (replacement character �) to preserve them as literal characters
//  2. Replace all custom separators with '/' so doublestar treats them as path separators
//
// Example with separator '.':
//
//	Input:  "foo.bar/baz.qux"
//	Step 1: "foo.bar�baz.qux"  (preserve literal /)
//	Step 2: "foo/bar�baz/qux"  (convert dots to separators)
//	Result: Only dots are separators, slashes remain as literal �
//
// IMPORTANT LIMITATION: This function cannot distinguish between literal
// separator characters and actual separators. For example, with separator '.',
// ALL dots will be replaced with '/', even if some were meant to be literal.
// This matches the behavior of the original gobwas/glob implementation.
func normalizePattern(s string, separators []rune, needsSlashEscape bool) string {
	if len(separators) == 0 {
		return s
	}

	result := s

	// Step 1: Preserve literal slashes if '/' is not a separator
	// Replace them with U+FFFD (replacement character) which should not appear in normal strings
	if needsSlashEscape {
		result = strings.ReplaceAll(result, "/", "\uFFFD")
	}

	// Step 2: Replace all custom separators with '/'
	for _, sep := range separators {
		if sep != '/' {
			result = strings.ReplaceAll(result, string(sep), "/")
		}
	}

	return result
}
