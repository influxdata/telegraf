package postgresql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

func TestURIParsing(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "short",
			uri:      `postgres://localhost`,
			expected: "host=localhost",
		},
		{
			name:     "with port",
			uri:      `postgres://localhost:5432`,
			expected: "host=localhost port=5432",
		},
		{
			name:     "with database",
			uri:      `postgres://localhost/mydb`,
			expected: "dbname=mydb host=localhost",
		},
		{
			name:     "with additional parameters",
			uri:      `postgres://localhost/mydb?application_name=pgxtest&search_path=myschema&connect_timeout=5`,
			expected: "application_name=pgxtest connect_timeout=5 dbname=mydb host=localhost search_path=myschema",
		},
		{
			name:     "with database setting in params",
			uri:      `postgres://localhost:5432/?database=mydb`,
			expected: "database=mydb host=localhost port=5432",
		},
		{
			name:     "with authentication",
			uri:      `postgres://jack:secret@localhost:5432/mydb?sslmode=prefer`,
			expected: "dbname=mydb host=localhost password=secret port=5432 sslmode=prefer user=jack",
		},
		{
			name:     "with spaces",
			uri:      `postgres://jack%20hunter:secret@localhost/mydb?application_name=pgx%20test`,
			expected: "application_name='pgx test' dbname=mydb host=localhost password=secret user='jack hunter'",
		},
		{
			name:     "with equal signs",
			uri:      `postgres://jack%20hunter:secret@localhost/mydb?application_name=pgx%3Dtest`,
			expected: "application_name='pgx=test' dbname=mydb host=localhost password=secret user='jack hunter'",
		},
		{
			name:     "multiple hosts",
			uri:      `postgres://jack:secret@foo:1,bar:2,baz:3/mydb?sslmode=disable`,
			expected: "dbname=mydb host=foo,bar,baz password=secret port=1,2,3 sslmode=disable user=jack",
		},
		{
			name:     "multiple hosts without ports",
			uri:      `postgres://jack:secret@foo,bar,baz/mydb?sslmode=disable`,
			expected: "dbname=mydb host=foo,bar,baz password=secret sslmode=disable user=jack",
		},
	}

	for _, tt := range tests {
		// Key value without spaces around equal sign
		t.Run(tt.name, func(t *testing.T) {
			actual, err := toKeyValue(tt.uri)
			require.NoError(t, err)
			require.Equalf(t, tt.expected, actual, "initial: %s", tt.uri)
		})
	}
}

func TestSanitizeAddressKeyValue(t *testing.T) {
	keys := []string{"password", "sslcert", "sslkey", "sslmode", "sslrootcert"}
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "simple text",
			value: `foo`,
		},
		{
			name:  "empty values",
			value: `''`,
		},
		{
			name:  "space in value",
			value: `'foo bar'`,
		},
		{
			name:  "equal sign in value",
			value: `'foo=bar'`,
		},
		{
			name:  "escaped quote",
			value: `'foo\'s bar'`,
		},
		{
			name:  "escaped quote no space",
			value: `\'foobar\'s\'`,
		},
		{
			name:  "escaped backslash",
			value: `'foo bar\\'`,
		},
		{
			name:  "escaped quote and backslash",
			value: `'foo\\\'s bar'`,
		},
		{
			name:  "two escaped backslashes",
			value: `'foo bar\\\\'`,
		},
		{
			name:  "multiple inline spaces",
			value: "'foo     \t bar'",
		},
		{
			name:  "leading space",
			value: `' foo bar'`,
		},
		{
			name:  "trailing space",
			value: `'foo bar '`,
		},
		{
			name:  "multiple equal signs",
			value: `'foo===bar'`,
		},
		{
			name:  "leading equal sign",
			value: `'=foo bar'`,
		},
		{
			name:  "trailing equal sign",
			value: `'foo bar='`,
		},
		{
			name:  "mix of equal signs and spaces",
			value: "'foo = a\t===\tbar'",
		},
	}

	for _, tt := range tests {
		// Key value without spaces around equal sign
		t.Run(tt.name, func(t *testing.T) {
			// Generate the DSN from the given keys and value
			parts := make([]string, 0, len(keys))
			for _, k := range keys {
				parts = append(parts, k+"="+tt.value)
			}
			dsn := strings.Join(parts, " canary=ok ")

			cfg := &Config{
				Address: config.NewSecret([]byte(dsn)),
			}

			expected := strings.Join(make([]string, len(keys)), "canary=ok ")
			expected = strings.TrimSpace(expected)
			actual, err := cfg.sanitizedAddress()
			require.NoError(t, err)
			require.Equalf(t, expected, actual, "initial: %s", dsn)
		})

		// Key value with spaces around equal sign
		t.Run("spaced "+tt.name, func(t *testing.T) {
			// Generate the DSN from the given keys and value
			parts := make([]string, 0, len(keys))
			for _, k := range keys {
				parts = append(parts, k+" = "+tt.value)
			}
			dsn := strings.Join(parts, " canary=ok ")

			cfg := &Config{
				Address: config.NewSecret([]byte(dsn)),
			}

			expected := strings.Join(make([]string, len(keys)), "canary=ok ")
			expected = strings.TrimSpace(expected)
			actual, err := cfg.sanitizedAddress()
			require.NoError(t, err)
			require.Equalf(t, expected, actual, "initial: %s", dsn)
		})
	}
}

func TestSanitizeAddressURI(t *testing.T) {
	keys := []string{"password", "sslcert", "sslkey", "sslmode", "sslrootcert"}
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "simple text",
			value: `foo`,
		},
		{
			name:  "empty values",
			value: ``,
		},
		{
			name:  "space in value",
			value: `foo bar`,
		},
		{
			name:  "equal sign in value",
			value: `foo=bar`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the DSN from the given keys and value
			value := strings.ReplaceAll(tt.value, "=", "%3D")
			value = strings.ReplaceAll(value, " ", "%20")
			parts := make([]string, 0, len(keys))
			for _, k := range keys {
				parts = append(parts, k+"="+value)
			}
			dsn := "postgresql://user:passwd@localhost:5432/db?" + strings.Join(parts, "&")

			cfg := &Config{
				Address: config.NewSecret([]byte(dsn)),
			}

			expected := "dbname=db host=localhost port=5432 user=user"
			actual, err := cfg.sanitizedAddress()
			require.NoError(t, err)
			require.Equalf(t, expected, actual, "initial: %s", dsn)
		})
	}
}
