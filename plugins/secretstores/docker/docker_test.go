package docker

import (
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestPathNonExistant(t *testing.T) {
	plugin := &Docker{
		ID:   "non_existent_path_test",
		Path: "non/existent/path",
	}
	err := plugin.Init()
	require.ErrorContains(t, err, "directory non/existent/path does not exist")
}

func TestSetNotAvailable(t *testing.T) {
	testdir, err := filepath.Abs("testdata")
	require.NoError(t, err, "testdata cannot be found")

	plugin := &Docker{
		ID:   "set_path_test",
		Path: testdir,
	}
	err = plugin.Init()
	require.NoError(t, err)

	// Try to Store the secrets, which this plugin should not let
	secret := map[string]string{
		"secret-file-1": "TryToSetThis",
	}
	for k, v := range secret {
		require.ErrorContains(t, plugin.Set(k, v), "secret-store does not support creating secrets")
	}
}

func TestListGet(t *testing.T) {
	// secret files name and their content to compare under the `testdata` directory
	secrets := map[string]string{
		"secret-file-1": "IWontTell",
		"secret_file_2": "SuperDuperSecret!23",
		"secretFile":    "foobar",
	}

	testdir, err := filepath.Abs("testdata")
	require.NoError(t, err, "testdata cannot be found")

	// Initialize the plugin
	plugin := &Docker{
		ID:   "test_list_get",
		Path: testdir,
	}
	require.NoError(t, plugin.Init())

	// List the Secrets
	keys, err := plugin.List()
	require.NoError(t, err)
	require.Len(t, keys, len(secrets))
	for _, k := range keys {
		_, found := secrets[k]
		require.True(t, found)
	}

	// Get the secrets
	for _, k := range keys {
		value, err := plugin.Get(k)
		require.NoError(t, err)
		v, found := secrets[k]
		require.True(t, found)
		require.Equal(t, v, string(value))
	}
}

func TestResolver(t *testing.T) {
	// Secret Value Name to Resolve
	secretFileName := "secret-file-1"
	// Secret Value to Resolve To
	secretVal := "IWontTell"

	testdir, err := filepath.Abs("testdata")
	require.NoError(t, err, "testdata cannot be found")

	// Initialize the plugin
	plugin := &Docker{
		ID:   "test_resolver",
		Path: testdir,
	}
	require.NoError(t, plugin.Init())

	// Get the resolver
	resolver, err := plugin.GetResolver(secretFileName)
	require.NoError(t, err)
	require.NotNil(t, resolver)
	s, dynamic, err := resolver()
	require.NoError(t, err)
	require.False(t, dynamic)
	require.Equal(t, secretVal, string(s))
}

func TestResolverInvalid(t *testing.T) {
	testdir, err := filepath.Abs("testdata")
	require.NoError(t, err, "testdata cannot be found")

	// Initialize the plugin
	plugin := &Docker{
		ID:   "test_invalid_resolver",
		Path: testdir,
	}
	require.NoError(t, plugin.Init())

	// Get the resolver
	resolver, err := plugin.GetResolver("foo")
	require.NoError(t, err)
	require.NotNil(t, resolver)
	_, _, err = resolver()
	require.ErrorContains(t, err, "cannot read the secret's value under the directory:")
}

func TestGetNonExistant(t *testing.T) {
	testdir, err := filepath.Abs("testdata")
	require.NoError(t, err, "testdata cannot be found")

	// Initialize the plugin
	plugin := &Docker{
		ID:   "test_nonexistent_get",
		Path: testdir,
	}
	require.NoError(t, plugin.Init())

	// Get the resolver
	_, err = plugin.Get("foo")
	require.ErrorContains(t, err, "cannot read the secret's value under the directory")
}
