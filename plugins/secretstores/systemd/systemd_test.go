//go:build linux

package systemd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func getSystemdVersionMin() (int, error) {
	return systemdMinimumVersion, nil
}

func TestSampleConfig(t *testing.T) {
	plugin := &Systemd{}
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestMinimumVersion(t *testing.T) {
	getSystemdVersion = func() (int, error) { return 123, nil }

	plugin := &Systemd{Log: testutil.Logger{}}
	require.ErrorContains(t, plugin.Init(), "below minimum version")
}

func TestEmptyPath(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin

	plugin := &Systemd{Log: testutil.Logger{}}
	require.ErrorContains(t, plugin.Init(), "'path' required without CREDENTIALS_DIRECTORY")
}

func TestEmptyCredentialsDirectoryWarning(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin

	logger := &testutil.CaptureLogger{}
	plugin := &Systemd{
		Path: "testdata",
		Log:  logger}
	require.NoError(t, plugin.Init())

	actual := logger.Warnings()
	require.Len(t, actual, 1)
	require.Contains(t, actual[0], "CREDENTIALS_DIRECTORY environment variable undefined")
}

func TestPathNonExistentExplicit(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin
	t.Setenv("CREDENTIALS_DIRECTORY", "testdata")

	plugin := &Systemd{
		Path: "non/existent/path",
		Log:  testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "accessing credentials directory")
}

func TestPathNonExistentImplicit(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin
	t.Setenv("CREDENTIALS_DIRECTORY", "non/existent/path")

	plugin := &Systemd{
		Log: testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "accessing credentials directory")
}

func TestInit(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin
	t.Setenv("CREDENTIALS_DIRECTORY", "testdata")

	plugin := &Systemd{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())
}

func TestSetNotAvailable(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin
	t.Setenv("CREDENTIALS_DIRECTORY", "testdata")

	plugin := &Systemd{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	// Try to Store the secrets, which this plugin should not let
	require.ErrorContains(t, plugin.Set("foo", "bar"), "secret-store does not support creating secrets")
}

func TestListGet(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin
	t.Setenv("CREDENTIALS_DIRECTORY", "testdata")

	// secret files name and their content to compare under the `testdata` directory
	secrets := map[string]string{
		"secret-file-1": "IWontTell",
		"secret_file_2": "SuperDuperSecret!23",
		"secretFile":    "foobar",
	}

	// Initialize the plugin
	plugin := &Systemd{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	// List the Secrets
	keys, err := plugin.List()
	require.NoError(t, err)
	require.Len(t, keys, len(secrets))
	// check if the returned array from List() is the same
	// as the name of secret files
	for secretFileName := range secrets {
		require.Contains(t, keys, secretFileName)
	}

	// Get the secrets
	for _, k := range keys {
		value, err := plugin.Get(k)
		require.NoError(t, err)
		v, found := secrets[k]
		require.Truef(t, found, "unexpected secret requested that was not found: %q", k)
		require.Equal(t, v, string(value))
	}
}

func TestResolver(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin
	t.Setenv("CREDENTIALS_DIRECTORY", "testdata")

	// Secret Value Name to Resolve
	secretFileName := "secret-file-1"
	// Secret Value to Resolve To
	secretVal := "IWontTell"

	// Initialize the plugin
	plugin := &Systemd{Log: testutil.Logger{}}
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
	getSystemdVersion = getSystemdVersionMin
	t.Setenv("CREDENTIALS_DIRECTORY", "testdata")

	// Initialize the plugin
	plugin := &Systemd{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	// Get the resolver
	resolver, err := plugin.GetResolver("foo")
	require.NoError(t, err)
	require.NotNil(t, resolver)
	_, _, err = resolver()
	require.ErrorContains(t, err, "cannot read the secret's value:")
}

func TestGetNonExistent(t *testing.T) {
	getSystemdVersion = getSystemdVersionMin
	t.Setenv("CREDENTIALS_DIRECTORY", "testdata")

	// Initialize the plugin
	plugin := &Systemd{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	// Get the resolver
	_, err := plugin.Get("foo")
	require.ErrorContains(t, err, "cannot read the secret's value:")
}
