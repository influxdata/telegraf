package docker

import (
	"os"
	"path/filepath"
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
		{
			name:     "invalid path",
			plugin:   &Docker{ID: "test"},
			expected: "path missing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestSetListGet(t *testing.T) {
	secrets := map[string]string{
		"secret-file-1": "I won't tell",
		"secret_file_2": "secret",
		"secretFile":    "foobar",
	}

	// Create a temporary directory we can use to store the secrets
	testdir, err := os.MkdirTemp("", "docker-*")
	require.NoError(t, err)
	defer os.RemoveAll(testdir)

	// Initialize the plugin
	plugin := &Docker{
		ID:   "test",
		Path: testdir,
	}
	require.NoError(t, plugin.Init())

	// Try to Store the secrets, which this plugin should not let
	for k, v := range secrets {
		require.ErrorContains(t, plugin.Set(k, v), "secret-store does not support creating secrets")
	}

	// Generate the secrets files under the temporary directory
	for fileName, secretContent := range secrets {
		fname := filepath.Join(testdir, fileName)
		err := os.WriteFile(fname, []byte(secretContent), 0644)
		require.NoError(t, err)
	}

	// List the Secrets
	keys, err := plugin.List()
	require.NoError(t, err)
	require.Len(t, keys, len(secrets))
	for _, k := range keys {
		_, found := secrets[k]
		require.True(t, found)
	}

	// Get the secrets
	require.Len(t, keys, len(secrets))
	for _, k := range keys {
		value, err := plugin.Get(k)
		require.NoError(t, err)
		v, found := secrets[k]
		require.True(t, found)
		require.Equal(t, v, string(value))
	}
}

func TestResolver(t *testing.T) {
	secretFileKey := "secret-file"
	secretVal := "I won't tell"

	// Create a temporary directory we can use to store the secrets
	testdir, err := os.MkdirTemp("", "docker-*")
	require.NoError(t, err)
	defer os.RemoveAll(testdir)

	// Initialize the plugin
	plugin := &Docker{
		ID:   "test",
		Path: testdir,
	}
	require.NoError(t, plugin.Init())

	// simulate what docker does by generating the secret file
	fname := filepath.Join(testdir, secretFileKey)
	err = os.WriteFile(fname, []byte(secretVal), 0644)
	require.NoError(t, err)

	// Get the resolver
	resolver, err := plugin.GetResolver(secretFileKey)
	require.NoError(t, err)
	require.NotNil(t, resolver)
	s, dynamic, err := resolver()
	require.NoError(t, err)
	require.False(t, dynamic)
	require.Equal(t, secretVal, string(s))
}

func TestResolverInvalid(t *testing.T) {
	secretFileKey := "secret_file_1"
	secretVal := "I won't tell"

	// Create a temporary directory we can use to store the secrets
	testdir, err := os.MkdirTemp("", "docker-*")
	require.NoError(t, err)
	defer os.RemoveAll(testdir)

	// Initialize the plugin
	plugin := &Docker{
		ID:   "test",
		Path: testdir,
	}
	require.NoError(t, plugin.Init())
	// simulate what docker does by generating the secret file
	fname := filepath.Join(testdir, secretFileKey)
	err = os.WriteFile(fname, []byte(secretVal), 0644)
	require.NoError(t, err)

	// Get the resolver
	resolver, err := plugin.GetResolver("foo")
	require.NoError(t, err)
	require.NotNil(t, resolver)
	_, _, err = resolver()
	require.Error(t, err)
}

func TestGetNonExistant(t *testing.T) {
	secretFileKey := "secretFile"
	secretVal := "I won't tell"

	// Create a temporary directory we can use to store the secrets
	testdir, err := os.MkdirTemp("", "docker-*")
	require.NoError(t, err)
	defer os.RemoveAll(testdir)

	// Initialize the plugin
	plugin := &Docker{
		ID:   "test",
		Path: testdir,
	}
	require.NoError(t, plugin.Init())
	// simulate what docker does by generating the secret file
	fname := filepath.Join(testdir, secretFileKey)
	err = os.WriteFile(fname, []byte(secretVal), 0644)
	require.NoError(t, err)

	// Get the resolver
	_, err = plugin.Get("foo")
	require.EqualError(t, err, "cannot find the secrets file under the directory mentioned in path parameter")
}
