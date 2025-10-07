package filter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePatternWithSlashEscape(t *testing.T) {
	// Test normalizePatternWithSlashEscape - preserves literal slashes when using non-slash separators
	// IMPORTANT: This has a known limitation - it cannot distinguish between literal separator
	// characters and actual separators. With separator '.', ALL dots become separators.
	tests := []struct {
		name       string
		pattern    string
		separators []rune
		expected   string
	}{
		{
			name:       "no separators",
			pattern:    "foo.bar/baz",
			separators: nil,
			expected:   "foo.bar/baz",
		},
		{
			name:       "single dot separator - preserve literal slash",
			pattern:    "foo.bar.baz",
			separators: []rune{'.'},
			expected:   "foo/bar/baz",
		},
		{
			name:       "single comma separator - preserve literal slash",
			pattern:    "foo,bar,baz",
			separators: []rune{','},
			expected:   "foo/bar/baz",
		},
		{
			name:       "multiple separators dot and comma - preserve literal slash",
			pattern:    "foo.bar,baz",
			separators: []rune{'.', ','},
			expected:   "foo/bar/baz",
		},
		{
			name:       "pattern with literal slashes and dot separator",
			pattern:    "foo.bar/baz.qux",
			separators: []rune{'.'},
			expected:   "foo/bar\uFFFDbaz/qux", // Slash preserved as U+FFFD, dots become /
		},
		{
			name:       "glob pattern with separator",
			pattern:    "foo.*.bar",
			separators: []rune{'.'},
			expected:   "foo/*/bar",
		},
		{
			name:       "complex glob with multiple separators and literal slash",
			pattern:    "foo.bar,baz/test.*.qux",
			separators: []rune{'.', ','},
			expected:   "foo/bar/baz\uFFFDtest/*/qux",
		},
		{
			name:       "double star pattern with dot separator",
			pattern:    "foo.**.bar",
			separators: []rune{'.'},
			expected:   "foo/**/bar",
		},
		{
			name:       "literal slash at start",
			pattern:    "/foo.bar.baz",
			separators: []rune{'.'},
			expected:   "\uFFFDfoo/bar/baz",
		},
		{
			name:       "literal slash at end",
			pattern:    "foo.bar.baz/",
			separators: []rune{'.'},
			expected:   "foo/bar/baz\uFFFD",
		},
		{
			name:       "multiple literal slashes",
			pattern:    "foo/bar/baz.qux",
			separators: []rune{'.'},
			expected:   "foo\uFFFDbar\uFFFDbaz/qux",
		},
		{
			name:       "empty pattern",
			pattern:    "",
			separators: []rune{'.'},
			expected:   "",
		},
		{
			name:       "pattern with only separators",
			pattern:    "...",
			separators: []rune{'.'},
			expected:   "///",
		},
		{
			name:       "unicode separator - preserve literal slash",
			pattern:    "foo•bar/test•baz",
			separators: []rune{'•'},
			expected:   "foo/bar\uFFFDtest/baz",
		},
		{
			name:       "separator not in pattern but slash preserved",
			pattern:    "foo/bar/baz",
			separators: []rune{'.'},
			expected:   "foo\uFFFDbar\uFFFDbaz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePatternWithSlashEscape(tt.pattern, tt.separators)
			require.Equal(t, tt.expected, result, "Pattern normalization failed for: %s", tt.pattern)
		})
	}
}

func TestNormalizePatternNoSlashEscape(t *testing.T) {
	// Test normalizePatternNoSlashEscape - treats slashes as separators
	tests := []struct {
		name       string
		pattern    string
		separators []rune
		expected   string
	}{
		{
			name:       "no separators",
			pattern:    "foo.bar/baz",
			separators: nil,
			expected:   "foo.bar/baz",
		},
		{
			name:       "slash separator only",
			pattern:    "foo/bar/baz",
			separators: []rune{'/'},
			expected:   "foo/bar/baz",
		},
		{
			name:       "slash in separators list - no escaping",
			pattern:    "foo.bar/baz.qux",
			separators: []rune{'.', '/'},
			expected:   "foo/bar/baz/qux",
		},
		{
			name:       "dot and slash separators with glob",
			pattern:    "foo.*/bar/baz.*",
			separators: []rune{'.', '/'},
			expected:   "foo/*/bar/baz/*",
		},
		{
			name:       "empty pattern",
			pattern:    "",
			separators: []rune{'.', '/'},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePatternNoSlashEscape(tt.pattern, tt.separators)
			require.Equal(t, tt.expected, result, "Pattern normalization failed for: %s", tt.pattern)
		})
	}
}

func TestFilterGlobWithSeparators(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		separators  []rune
		input       string
		shouldMatch bool
	}{
		{
			name:        "simple dot separator match",
			patterns:    []string{"cpu.*.count"},
			separators:  []rune{'.'},
			input:       "cpu.user.count",
			shouldMatch: true,
		},
		{
			name:        "simple dot separator no match",
			patterns:    []string{"cpu.*.count"},
			separators:  []rune{'.'},
			input:       "cpu.count",
			shouldMatch: false,
		},
		{
			name:        "multiple levels with dot separator",
			patterns:    []string{"cpu.*.*.count"},
			separators:  []rune{'.'},
			input:       "cpu.user.idle.count",
			shouldMatch: true,
		},
		{
			name:        "multiple separators",
			patterns:    []string{"cpu.user,count"},
			separators:  []rune{'.', ','},
			input:       "cpu.user,count",
			shouldMatch: true,
		},
		{
			name:        "glob with mixed separators in pattern",
			patterns:    []string{"metric.*.value,*"},
			separators:  []rune{'.', ','},
			input:       "metric.cpu.value,high",
			shouldMatch: true,
		},
		{
			name:        "no separators standard glob",
			patterns:    []string{"cpu.*"},
			separators:  nil,
			input:       "cpu.user",
			shouldMatch: true,
		},
		{
			name:        "separator in input but not in pattern",
			patterns:    []string{"cpu*"},
			separators:  []rune{'.'},
			input:       "cpu.user",
			shouldMatch: false, // When separator is defined, both pattern and input are normalized
		},
		// CRITICAL TEST CASES: Literal slash preservation with non-slash separators
		{
			name:        "literal slash in pattern should NOT match when using dot separator",
			patterns:    []string{"foo.bar/baz.qux"},
			separators:  []rune{'.'},
			input:       "foo.bar.baz.qux",
			shouldMatch: false, // Literal '/' in pattern doesn't match '.' in input
		},
		{
			name:        "literal slash in pattern SHOULD match literal slash in input",
			patterns:    []string{"foo.bar/baz.qux"},
			separators:  []rune{'.'},
			input:       "foo.bar/baz.qux",
			shouldMatch: true, // Literal '/' matches literal '/'
		},
		{
			name:        "literal slash in input should NOT match dot separator in pattern",
			patterns:    []string{"foo.bar.baz.qux"},
			separators:  []rune{'.'},
			input:       "foo.bar/baz.qux",
			shouldMatch: false, // '.' in pattern (separator) doesn't match '/' in input (literal)
		},
		{
			name:        "wildcard with literal slash - match",
			patterns:    []string{"foo/*/bar"},
			separators:  []rune{'.'},
			input:       "foo/anything/bar",
			shouldMatch: true,
		},
		{
			name:        "wildcard with literal slash - matches anything between slashes",
			patterns:    []string{"foo/*/bar"},
			separators:  []rune{'.'},
			input:       "foo/any/thing/bar",
			shouldMatch: true, // * matches "any/thing" as a single segment (slashes are literals)
		},
		{
			name:        "double star with dot separator and literal slash",
			patterns:    []string{"foo/**/bar.qux"},
			separators:  []rune{'.'},
			input:       "foo/x/y/z/bar.qux",
			shouldMatch: true, // ** crosses literal / boundaries
		},
		{
			name:        "complex pattern with mixed literal slashes and separators",
			patterns:    []string{"server/logs.*.error"},
			separators:  []rune{'.'},
			input:       "server/logs.2024.error",
			shouldMatch: true,
		},
		{
			name:        "slash as separator - slashes are treated as separators",
			patterns:    []string{"foo.bar/baz.qux"},
			separators:  []rune{'.', '/'},
			input:       "foo.bar/baz.qux",
			shouldMatch: true, // Both . and / are separators
		},
		{
			name:        "slash as separator - wildcard crosses slash boundary",
			patterns:    []string{"foo.*.qux"},
			separators:  []rune{'.', '/'},
			input:       "foo.bar/baz.qux",
			shouldMatch: false, // * doesn't match multiple segments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filter Filter
			var err error

			if len(tt.separators) == 0 {
				filter, err = newFilterGlob(tt.patterns)
			} else {
				filter, err = newFilterGlobWithSeparators(tt.patterns, tt.separators)
			}
			require.NoError(t, err, "Failed to create filter")

			result := filter.Match(tt.input)
			require.Equal(t, tt.shouldMatch, result, "Match result incorrect for input: %s", tt.input)
		})
	}
}

func TestFilterGlobEdgeCases(t *testing.T) {
	t.Run("invalid pattern", func(t *testing.T) {
		// Invalid patterns should be caught during creation
		_, err := newFilterGlobWithSeparators([]string{"[abc"}, []rune{'.'})
		require.Error(t, err, "Should error on invalid pattern")
	})

	t.Run("empty pattern list", func(t *testing.T) {
		filter, err := newFilterGlob(nil)
		require.NoError(t, err)
		require.False(t, filter.Match("anything"), "Empty pattern list should match nothing")
	})

	t.Run("pattern with bracket expressions and separators", func(t *testing.T) {
		filter, err := newFilterGlobWithSeparators([]string{"cpu.[ab].count"}, []rune{'.'})
		require.NoError(t, err)

		// These should match
		require.True(t, filter.Match("cpu.a.count"))
		require.True(t, filter.Match("cpu.b.count"))

		// These should not match
		require.False(t, filter.Match("cpu.c.count"))
		require.False(t, filter.Match("cpu.ab.count"))
	})

	t.Run("question mark with separators", func(t *testing.T) {
		filter, err := newFilterGlobWithSeparators([]string{"cpu.?.count"}, []rune{'.'})
		require.NoError(t, err)

		// Single character should match
		require.True(t, filter.Match("cpu.a.count"))
		require.True(t, filter.Match("cpu.1.count"))

		// Multiple characters should not match
		require.False(t, filter.Match("cpu.ab.count"))
		require.False(t, filter.Match("cpu..count"))
	})
}

func BenchmarkNormalizePatternWithSlashEscape(b *testing.B) {
	pattern := "foo.bar.baz.qux.quux"
	separators := []rune{'.', ','}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizePatternWithSlashEscape(pattern, separators)
	}
}

func BenchmarkNormalizePatternNoSlashEscape(b *testing.B) {
	pattern := "foo.bar/baz.qux/quux"
	separators := []rune{'.', '/'}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizePatternNoSlashEscape(pattern, separators)
	}
}

func BenchmarkFilterGlobMatch(b *testing.B) {
	b.Run("without separators", func(b *testing.B) {
		filter, err := newFilterGlob([]string{"cpu*", "mem*", "disk*"})
		require.NoError(b, err)
		input := "cpu.user.count"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = filter.Match(input)
		}
	})

	b.Run("with separators", func(b *testing.B) {
		filter, err := newFilterGlobWithSeparators([]string{"cpu.*.count", "mem.*.used", "disk.*.free"}, []rune{'.'})
		require.NoError(b, err)
		input := "cpu.user.count"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = filter.Match(input)
		}
	})
}
