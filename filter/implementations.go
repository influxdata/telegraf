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

// filterGlobWithSeparators handles glob patterns WITH separators where '/' is also a separator.
// This is a separate implementation optimized for the case where slashes don't need escaping.
type filterGlobWithSeparators struct {
	normalizedPatterns []string
	separators         []rune
}

// filterGlobWithSeparatorsAndSlashEscape handles glob patterns WITH separators where '/' is NOT a separator.
// This requires escaping literal slashes to preserve them during path matching.
type filterGlobWithSeparatorsAndSlashEscape struct {
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

	// Remove '/' from separators to avoid no-op replacements in the hot path
	// since we're converting all separators to '/' anyway
	filteredSeparators := make([]rune, 0, len(separators))
	for _, sep := range separators {
		if sep != '/' {
			filteredSeparators = append(filteredSeparators, sep)
		}
	}

	// Pre-compute normalized patterns without slash escaping
	normalizedPatterns := make([]string, len(filters))
	for i, pattern := range filters {
		normalizedPatterns[i] = normalizePatternNoSlashEscape(pattern, filteredSeparators)
	}

	return &filterGlobWithSeparators{
		normalizedPatterns: normalizedPatterns,
		separators:         filteredSeparators,
	}, nil
}

func newFilterGlobWithSeparatorsAndSlashEscape(filters []string, separators []rune) (Filter, error) {
	// Validate all patterns
	for _, pattern := range filters {
		if _, err := doublestar.Match(pattern, ""); err != nil {
			return nil, err
		}
	}

	// Pre-compute normalized patterns with slash escaping
	normalizedPatterns := make([]string, len(filters))
	for i, pattern := range filters {
		normalizedPatterns[i] = normalizePatternWithSlashEscape(pattern, separators)
	}

	return &filterGlobWithSeparatorsAndSlashEscape{
		normalizedPatterns: normalizedPatterns,
		separators:         separators,
	}, nil
}

func (f *filterGlobWithSeparators) Match(s string) bool {
	// Normalize the input string without slash escaping
	normalizedStr := normalizePatternNoSlashEscape(s, f.separators)

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

func (f *filterGlobWithSeparatorsAndSlashEscape) Match(s string) bool {
	// Normalize the input string with slash escaping
	normalizedStr := normalizePatternWithSlashEscape(s, f.separators)

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

// normalizePatternWithSlashEscape converts all separators to '/' while preserving literal slashes.
// Used when '/' is NOT in the separators list.
//
// This allows doublestar.PathMatch to treat custom separators as path separators while
// keeping literal slashes as regular characters.
//
// The normalization process:
//  1. Replace all literal '/' with U+FFFD (replacement character �) to preserve them
//  2. Replace all custom separators with '/' so doublestar treats them as path separators
//
// Example with separator '.':
//
//	Input:  "foo.bar/baz.qux"
//	Step 1: "foo.bar�baz.qux"  (preserve literal /)
//	Step 2: "foo/bar�baz/qux"  (convert dots to separators)
//	Result: Only dots are path separators, slashes remain as literals (�)
//
// IMPORTANT LIMITATIONS:
//  1. Cannot distinguish between literal separator characters and actual separators.
//     For example, with separator '.', ALL dots will be replaced with '/', even if
//     some were meant to be literal. This matches gobwas/glob behavior.
//  2. If input legitimately contains U+FFFD (�), it could be incorrectly matched.
//     This is acceptable because U+FFFD is a replacement character for invalid/undecodable
//     Unicode and should not appear in normal metric names.
func normalizePatternWithSlashEscape(s string, separators []rune) string {
	// Early return for edge case (should not happen in normal usage)
	if len(separators) == 0 {
		return s
	}

	// Step 1: Preserve literal slashes by replacing with U+FFFD
	result := strings.ReplaceAll(s, "/", "\uFFFD")

	// Step 2: Replace all custom separators with '/'
	return normalizePatternNoSlashEscape(result, separators)
}

// normalizePatternNoSlashEscape converts all separators to '/' without escaping.
// Used when '/' IS in the separators list (or equals the only separator).
//
// This allows doublestar.PathMatch to treat custom separators as path separators.
//
// Example with separators '.' and '/':
//
//	Input:  "foo.bar/baz.qux"
//	Result: "foo/bar/baz/qux"  (both . and / become separators)
//
// IMPORTANT LIMITATION: Cannot distinguish between literal separator characters
// and actual separators. With separator '.', ALL dots will be replaced with '/',
// even if some were meant to be literal. This matches gobwas/glob behavior.
func normalizePatternNoSlashEscape(s string, separators []rune) string {
	// Early return for edge case (should not happen in normal usage)
	if len(separators) == 0 {
		return s
	}

	result := s

	// Replace all custom separators with '/'
	for _, sep := range separators {
		result = strings.ReplaceAll(result, string(sep), "/")
	}

	return result
}
