package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEnvironmentSubstitution(t *testing.T) {
	tests := []struct {
		name         string
		setEnv       func(*testing.T)
		contents     string
		expected     string
		wantErr      bool
		errSubstring string
	}{
		{
			name: "Legacy with ${} and without {}",
			setEnv: func(t *testing.T) {
				t.Setenv("TEST_ENV1", "VALUE1")
				t.Setenv("TEST_ENV2", "VALUE2")
			},
			contents: "A string with ${TEST_ENV1}, $TEST_ENV2 and $TEST_ENV1 as repeated",
			expected: "A string with VALUE1, VALUE2 and VALUE1 as repeated",
		},
		{
			name:     "Env not set",
			contents: "Env variable ${NOT_SET} will be empty",
			expected: "Env variable  will be empty", // Two spaces present
		},
		{
			name:     "Env not set, fallback to default",
			contents: "Env variable ${THIS_IS_ABSENT:-Fallback}",
			expected: "Env variable Fallback",
		},
		{
			name: "No fallback",
			setEnv: func(t *testing.T) {
				t.Setenv("MY_ENV1", "VALUE1")
			},
			contents: "Env variable ${MY_ENV1:-Fallback}",
			expected: "Env variable VALUE1",
		},
		{
			name: "Mix and match",
			setEnv: func(t *testing.T) {
				t.Setenv("MY_VAR", "VALUE")
				t.Setenv("MY_VAR2", "VALUE2")
			},
			contents: "Env var ${MY_VAR} is set, with $MY_VAR syntax and default on this ${MY_VAR1:-Substituted}, no default on this ${MY_VAR2:-NoDefault}",
			expected: "Env var VALUE is set, with VALUE syntax and default on this Substituted, no default on this VALUE2",
		},
		{
			name:     "Default has special chars",
			contents: `Not recommended but supported ${MY_VAR:-Default with special chars Supported#$\"}`,
			expected: `Not recommended but supported Default with special chars Supported#$\"`, // values are escaped
		},
		{
			name:         "unset error",
			contents:     "Contains ${THIS_IS_NOT_SET?unset-error}",
			wantErr:      true,
			errSubstring: "unset-error",
		},
		{
			name: "env empty error",
			setEnv: func(t *testing.T) {
				t.Setenv("ENV_EMPTY", "")
			},
			contents:     "Contains ${ENV_EMPTY:?empty-error}",
			wantErr:      true,
			errSubstring: "empty-error",
		},
		{
			name: "Fallback as env variable",
			setEnv: func(t *testing.T) {
				t.Setenv("FALLBACK", "my-fallback")
			},
			contents: "Should output ${NOT_SET:-${FALLBACK}}",
			expected: "Should output my-fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv != nil {
				tt.setEnv(t)
			}
			actual, err := substituteEnvironment([]byte(tt.contents))
			if tt.wantErr {
				require.ErrorContains(t, err, tt.errSubstring)
				return
			}
			require.EqualValues(t, tt.expected, string(actual))
		})
	}
}

func TestURLRetries3Fails(t *testing.T) {
	httpLoadConfigRetryInterval = 0 * time.Second
	responseCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		responseCounter++
	}))
	defer ts.Close()

	expected := fmt.Sprintf("error loading config file %s: retry 3 of 3 failed to retrieve remote config: 404 Not Found", ts.URL)

	c := NewConfig()
	err := c.LoadConfig(ts.URL)
	require.Error(t, err)
	require.Equal(t, expected, err.Error())
	require.Equal(t, 4, responseCounter)
}

func TestURLRetries3FailsThenPasses(t *testing.T) {
	httpLoadConfigRetryInterval = 0 * time.Second
	responseCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if responseCounter <= 2 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		responseCounter++
	}))
	defer ts.Close()

	c := NewConfig()
	require.NoError(t, c.LoadConfig(ts.URL))
	require.Equal(t, 4, responseCounter)
}

func TestConfig_getDefaultConfigPathFromEnvURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewConfig()
	t.Setenv("TELEGRAF_CONFIG_PATH", ts.URL)
	configPath, err := getDefaultConfigPath()
	require.NoError(t, err)
	require.Equal(t, []string{ts.URL}, configPath)
	err = c.LoadConfig("")
	require.NoError(t, err)
}
