//go:build darwin || linux || windows
// +build darwin linux windows

package os

import (
	"testing"

	"github.com/influxdata/telegraf/internal/choice"
	"github.com/stretchr/testify/require"
)

func TestSampleConfig(t *testing.T) {
	plugin := &OS{}
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *OS
		expected string
	}{
		{
			name:     "invalid id",
			plugin:   &OS{},
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

func TestResolverInvalid(t *testing.T) {
	plugin := &OS{ID: "test"}
	require.NoError(t, plugin.Init())

	// Make sure the key does not exist and try to read that key
	testKey := "foobar secret key"
	keys, err := plugin.List()
	require.NoError(t, err)
	for choice.Contains(testKey, keys) {
		testKey += "x"
	}
	// Get the resolver
	resolver, err := plugin.GetResolver(testKey)
	require.NoError(t, err)
	require.NotNil(t, resolver)
	_, _, err = resolver()
	require.Error(t, err)
}

func TestGetNonExisting(t *testing.T) {
	plugin := &OS{ID: "test"}
	require.NoError(t, plugin.Init())

	// Make sure the key does not exist and try to read that key
	testKey := "foobar secret key"
	keys, err := plugin.List()
	require.NoError(t, err)
	for choice.Contains(testKey, keys) {
		testKey += "x"
	}
	_, err = plugin.Get(testKey)
	require.EqualError(t, err, "The specified item could not be found in the keyring")
}
