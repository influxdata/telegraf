package opcua

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompilePathPatternErrors(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		errMsg  string
	}{
		{"empty", "", "empty pattern"},
		{"only separator", "/", "empty segment"},
		{"leading separator", "/Objects", "empty segment"},
		{"trailing separator", "Objects/", "empty segment"},
		{"double separator", "Objects//Plant1", "empty segment"},
		{"unclosed bracket", "Objects/[Plant", "invalid segment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompilePathPattern(tt.pattern)
			require.ErrorContains(t, err, tt.errMsg)
		})
	}
}

func TestPathPatternMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		segments []string
		match    bool
	}{
		{"exact match", "Objects/Plant1", []string{"Objects", "Plant1"}, true},
		{"exact mismatch", "Objects/Plant1", []string{"Objects", "Plant2"}, false},
		{"shorter than pattern", "Objects/Plant1", []string{"Objects"}, false},
		{"longer than pattern", "Objects/Plant1", []string{"Objects", "Plant1", "MV01"}, false},

		{"single segment wildcard", "Objects/*", []string{"Objects", "Plant1"}, true},
		{"single segment wildcard rejects zero", "Objects/*", []string{"Objects"}, false},
		{"single segment wildcard rejects multi", "Objects/*", []string{"Objects", "Plant1", "Device1"}, false},

		{"prefix wildcard match", "MV*", []string{"MV01"}, true},
		{"prefix wildcard mismatch", "MV*", []string{"AB01"}, false},
		{"empty prefix wildcard", "MV*", []string{"MV"}, true},

		{"single char wildcard", "Plant?", []string{"Plant1"}, true},
		{"single char wildcard rejects multi", "Plant?", []string{"Plant10"}, false},
		{"single char wildcard rejects empty suffix", "Plant?", []string{"Plant"}, false},

		{"char class match low", "Plant[12]", []string{"Plant1"}, true},
		{"char class match high", "Plant[12]", []string{"Plant2"}, true},
		{"char class miss", "Plant[12]", []string{"Plant3"}, false},

		{"recursive at end matches one extra", "Objects/**", []string{"Objects", "Plant1"}, true},
		{"recursive at end matches zero extra", "Objects/**", []string{"Objects"}, true},
		{"recursive at end matches deep", "Objects/**", []string{"Objects", "Plant1", "Device1", "MV01"}, true},
		{"recursive at end mismatch root", "Objects/**", []string{"Plant1"}, false},

		{"recursive in middle zero", "Objects/**/Temperature", []string{"Objects", "Temperature"}, true},
		{"recursive in middle one", "Objects/**/Temperature", []string{"Objects", "Plant1", "Temperature"}, true},
		{"recursive in middle deep", "Objects/**/Temperature", []string{"Objects", "Plant1", "Device1", "Temperature"}, true},
		{"recursive in middle wrong leaf", "Objects/**/Temperature", []string{"Objects", "Plant1", "Pressure"}, false},

		{"only recursive matches anything", "**", []string{"Objects", "Plant1"}, true},
		{"only recursive matches empty", "**", nil, true},
		{"recursive then literal", "**/Temperature", []string{"Temperature"}, true},
		{"recursive then literal deep", "**/Temperature", []string{"Plant1", "Device1", "Temperature"}, true},

		{"compound match", "Objects/Plant1/*/MV*", []string{"Objects", "Plant1", "Device1", "MV01"}, true},
		{"compound wrong root", "Objects/Plant1/*/MV*", []string{"Objects", "Plant2", "Device1", "MV01"}, false},
		{"compound wrong leaf", "Objects/Plant1/*/MV*", []string{"Objects", "Plant1", "Device1", "Status"}, false},

		{"adjacent recursive collapse", "Objects/**/**/Temperature", []string{"Objects", "A", "B", "Temperature"}, true},
		{"adjacent recursive collapse zero", "Objects/**/**/Temperature", []string{"Objects", "Temperature"}, true},

		{"escaped wildcard literal", `Objects/MV\*`, []string{"Objects", "MV*"}, true},
		{"escaped wildcard rejects glob", `Objects/MV\*`, []string{"Objects", "MV01"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := CompilePathPattern(tt.pattern)
			require.NoError(t, err)
			require.Equal(t, tt.match, p.Match(tt.segments))
		})
	}
}

func TestPathPatternString(t *testing.T) {
	p, err := CompilePathPattern("Objects/Plant1/*/MV*")
	require.NoError(t, err)
	require.Equal(t, "Objects/Plant1/*/MV*", p.String())
}
