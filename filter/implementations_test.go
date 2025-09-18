package filter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePattern(t *testing.T) {
	// IMPORTANT: normalizePattern has a known limitation - it cannot distinguish
	// between literal separator characters and actual separators in the pattern.
	// For example, with separator '.', the pattern "foo.bar/baz.qux" will have
	// ALL dots replaced with '/', even if some were meant to be literal.
	// This is a trade-off for simplicity and performance.
	tests := []struct {
		name       string
		pattern    string
		separators []rune
		expected   string
	}{
		{
			name:       "no separators",
			pattern:    "foo.bar/baz",
			separators: []rune{},
			expected:   "foo.bar/baz",
		},
		{
			name:       "single dot separator",
			pattern:    "foo.bar.baz",
			separators: []rune{'.'},
			expected:   "foo/bar/baz",
		},
		{
			name:       "single comma separator",
			pattern:    "foo,bar,baz",
			separators: []rune{','},
			expected:   "foo/bar/baz",
		},
		{
			name:       "slash separator (no change)",
			pattern:    "foo/bar/baz",
			separators: []rune{'/'},
			expected:   "foo/bar/baz",
		},
		{
			name:       "multiple separators dot and comma",
			pattern:    "foo.bar,baz",
			separators: []rune{'.', ','},
			expected:   "foo/bar/baz",
		},
		{
			name:       "pattern with existing slashes and dot separator",
			pattern:    "foo.bar/baz.qux",
			separators: []rune{'.'},
			expected:   "foo/bar/baz/qux",
		},
		{
			name:       "glob pattern with separator",
			pattern:    "foo.*.bar",
			separators: []rune{'.'},
			expected:   "foo/*/bar",
		},
		{
			name:       "complex glob with multiple separators",
			pattern:    "foo.bar,baz.*.qux",
			separators: []rune{'.', ','},
			expected:   "foo/bar/baz/*/qux",
		},
		{
			name:       "double star pattern with dot separator",
			pattern:    "foo.**.bar",
			separators: []rune{'.'},
			expected:   "foo/**/bar",
		},
		{
			name:       "mixed literal and separator usage",
			pattern:    "foo.bar/lala.hoo",
			separators: []rune{'.'},
			expected:   "foo/bar/lala/hoo", // This shows the limitation - we can't distinguish literal dots from separator dots
		},
		{
			name:       "escaped characters (not actually escaped in our impl)",
			pattern:    `foo\.bar.baz`,
			separators: []rune{'.'},
			expected:   `foo\/bar/baz`, // Shows that escaping doesn't work as expected
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
			name:       "unicode separator",
			pattern:    "foo•bar•baz",
			separators: []rune{'•'},
			expected:   "foo/bar/baz",
		},
		{
			name:       "separator not in pattern",
			pattern:    "foo/bar/baz",
			separators: []rune{'.'},
			expected:   "foo/bar/baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePattern(tt.pattern, tt.separators)
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
			name:        "pattern with literal slash and dot separator",
			patterns:    []string{"path/to.*.file"},
			separators:  []rune{'.'},
			input:       "path/to.config.file",
			shouldMatch: true,
		},
		{
			name:        "no separators standard glob",
			patterns:    []string{"cpu.*"},
			separators:  []rune{},
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
		filter, err := newFilterGlob([]string{})
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

func BenchmarkNormalizePattern(b *testing.B) {
	pattern := "foo.bar.baz.qux.quux"
	separators := []rune{'.', ','}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizePattern(pattern, separators)
	}
}

func BenchmarkFilterGlobMatch(b *testing.B) {
	b.Run("without separators", func(b *testing.B) {
		filter, _ := newFilterGlob([]string{"cpu*", "mem*", "disk*"})
		input := "cpu.user.count"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = filter.Match(input)
		}
	})

	b.Run("with separators", func(b *testing.B) {
		filter, _ := newFilterGlobWithSeparators([]string{"cpu.*.count", "mem.*.used", "disk.*.free"}, []rune{'.'})
		input := "cpu.user.count"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = filter.Match(input)
		}
	})
}
