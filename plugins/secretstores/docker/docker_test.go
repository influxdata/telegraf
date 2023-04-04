package docker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSampleConfig(t *testing.T) {
	plugin := &Docker{}
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *Docker
		expected string
	}{
		{
			name:     "invalid id",
			plugin:   &Docker{},
			expected: "id missing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.ErrorContains(t, err, tt.expected)
		})
	}
}
