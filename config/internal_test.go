package config

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
			expected: "Env variable ${NOT_SET} will be empty",
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
			name: "empty but set",
			setEnv: func(t *testing.T) {
				t.Setenv("EMPTY", "")
			},
			contents: "Contains ${EMPTY} nothing",
			expected: "Contains  nothing",
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
			actual, err := substituteEnvironment([]byte(tt.contents), false)
			if tt.wantErr {
				require.ErrorContains(t, err, tt.errSubstring)
				return
			}
			require.EqualValues(t, tt.expected, string(actual))
		})
	}
}

func TestEnvironmentSubstitutionOldBehavior(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		expected string
	}{
		{
			name:     "not defined no brackets",
			contents: `my-da$tabase`,
			expected: `my-da$tabase`,
		},
		{
			name:     "not defined brackets",
			contents: `my-da${ta}base`,
			expected: `my-da${ta}base`,
		},
		{
			name:     "not defined no brackets double dollar",
			contents: `my-da$$tabase`,
			expected: `my-da$$tabase`,
		},
		{
			name:     "not defined no brackets backslash",
			contents: `my-da\$tabase`,
			expected: `my-da\$tabase`,
		},
		{
			name:     "not defined brackets backslash",
			contents: `my-da\${ta}base`,
			expected: `my-da\${ta}base`,
		},
		{
			name:     "no brackets and suffix",
			contents: `my-da$VARbase`,
			expected: `my-da$VARbase`,
		},
		{
			name:     "no brackets",
			contents: `my-da$VAR`,
			expected: `my-dafoobar`,
		},
		{
			name:     "brackets",
			contents: `my-da${VAR}base`,
			expected: `my-dafoobarbase`,
		},
		{
			name:     "no brackets double dollar",
			contents: `my-da$$VAR`,
			expected: `my-da$foobar`,
		},
		{
			name:     "brackets double dollar",
			contents: `my-da$${VAR}`,
			expected: `my-da$foobar`,
		},
		{
			name:     "no brackets backslash",
			contents: `my-da\$VAR`,
			expected: `my-da\foobar`,
		},
		{
			name:     "brackets backslash",
			contents: `my-da\${VAR}base`,
			expected: `my-da\foobarbase`,
		},
		{
			name:     "fallback",
			contents: `my-da${ta:-omg}base`,
			expected: `my-daomgbase`,
		},
		{
			name:     "fallback env",
			contents: `my-da${ta:-${FALLBACK}}base`,
			expected: `my-dadefaultbase`,
		},
		{
			name:     "regex substitution",
			contents: `${1}`,
			expected: `${1}`,
		},
		{
			name:     "empty but set",
			contents: "Contains ${EMPTY} nothing",
			expected: "Contains  nothing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VAR", "foobar")
			t.Setenv("FALLBACK", "default")
			t.Setenv("EMPTY", "")
			actual, err := substituteEnvironment([]byte(tt.contents), true)
			require.NoError(t, err)
			require.EqualValues(t, tt.expected, string(actual))
		})
	}
}

func TestEnvironmentSubstitutionNewBehavior(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		expected string
	}{
		{
			name:     "not defined no brackets",
			contents: `my-da$tabase`,
			expected: `my-da$tabase`,
		},
		{
			name:     "not defined brackets",
			contents: `my-da${ta}base`,
			expected: `my-da${ta}base`,
		},
		{
			name:     "not defined no brackets double dollar",
			contents: `my-da$$tabase`,
			expected: `my-da$tabase`,
		},
		{
			name:     "not defined no brackets backslash",
			contents: `my-da\$tabase`,
			expected: `my-da\$tabase`,
		},
		{
			name:     "not defined brackets backslash",
			contents: `my-da\${ta}base`,
			expected: `my-da\${ta}base`,
		},
		{
			name:     "no brackets and suffix",
			contents: `my-da$VARbase`,
			expected: `my-da$VARbase`,
		},
		{
			name:     "no brackets",
			contents: `my-da$VAR`,
			expected: `my-dafoobar`,
		},
		{
			name:     "brackets",
			contents: `my-da${VAR}base`,
			expected: `my-dafoobarbase`,
		},
		{
			name:     "no brackets double dollar",
			contents: `my-da$$VAR`,
			expected: `my-da$VAR`,
		},
		{
			name:     "brackets double dollar",
			contents: `my-da$${VAR}`,
			expected: `my-da${VAR}`,
		},
		{
			name:     "no brackets backslash",
			contents: `my-da\$VAR`,
			expected: `my-da\foobar`,
		},
		{
			name:     "brackets backslash",
			contents: `my-da\${VAR}base`,
			expected: `my-da\foobarbase`,
		},
		{
			name:     "fallback",
			contents: `my-da${ta:-omg}base`,
			expected: `my-daomgbase`,
		},
		{
			name:     "fallback env",
			contents: `my-da${ta:-${FALLBACK}}base`,
			expected: `my-dadefaultbase`,
		},
		{
			name:     "regex substitution",
			contents: `${1}`,
			expected: `${1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VAR", "foobar")
			t.Setenv("FALLBACK", "default")
			actual, err := substituteEnvironment([]byte(tt.contents), false)
			require.NoError(t, err)
			require.EqualValues(t, tt.expected, string(actual))
		})
	}
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name     string
		setEnv   func(*testing.T)
		contents string
		expected string
		errmsg   string
	}{
		{
			name: "empty var name",
			contents: `
# Environment variables can be used anywhere in this config file, simply surround
# them with ${}. For strings the variable must be within quotes (ie, "${STR_VAR}"),
# for numbers and booleans they should be plain (ie, ${INT_VAR}, ${BOOL_VAR})Should output ${NOT_SET:-${FALLBACK}}
`,
			expected: "\n\n\n\n",
		},
		{
			name: "comment in command (issue #13643)",
			contents: `
			[[inputs.exec]]
			  commands = ["echo \"abc#def\""]
			`,
			expected: `
			[[inputs.exec]]
			  commands = ["echo \"abc#def\""]
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv != nil {
				tt.setEnv(t)
			}
			tbl, err := parseConfig([]byte(tt.contents))
			if tt.errmsg != "" {
				require.ErrorContains(t, err, tt.errmsg)
				return
			}

			require.NoError(t, err)
			if len(tt.expected) > 0 {
				require.EqualValues(t, tt.expected, string(tbl.Data))
			}
		})
	}
}

func TestRemoveComments(t *testing.T) {
	// Read expectation
	expected, err := os.ReadFile(filepath.Join("testdata", "envvar_comments_expected.toml"))
	require.NoError(t, err)

	// Read the file and remove the comments
	buf, err := os.ReadFile(filepath.Join("testdata", "envvar_comments.toml"))
	require.NoError(t, err)
	removed, err := removeComments(buf)
	require.NoError(t, err)
	lines := bytes.Split(removed, []byte{'\n'})
	for i, line := range lines {
		lines[i] = bytes.TrimRight(line, " \t")
	}
	actual := bytes.Join(lines, []byte{'\n'})

	// Do the comparison
	require.Equal(t, string(expected), string(actual))
}

func TestURLRetries3Fails(t *testing.T) {
	httpLoadConfigRetryInterval = 0 * time.Second
	responseCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		responseCounter++
	}))
	defer ts.Close()

	expected := fmt.Sprintf("error loading config file %s: failed to fetch HTTP config: 404 Not Found", ts.URL)

	c := NewConfig()
	err := c.LoadConfig(ts.URL)
	require.Error(t, err)
	require.Equal(t, expected, err.Error())
	require.Equal(t, 4, responseCounter)
}

func TestURLRetries3FailsThenPasses(t *testing.T) {
	httpLoadConfigRetryInterval = 0 * time.Second
	responseCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
