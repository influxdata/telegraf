package v2

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertGlobalStatus(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       sql.RawBytes
		expected    interface{}
		expectedErr error
	}{
		{
			name:        "default",
			key:         "ssl_ctx_verify_depth",
			value:       []byte("0"),
			expected:    uint64(0),
			expectedErr: nil,
		},
		{
			name:        "overflow int64",
			key:         "ssl_ctx_verify_depth",
			value:       []byte("18446744073709551615"),
			expected:    uint64(18446744073709551615),
			expectedErr: nil,
		},
		{
			name:        "defined variable but unset",
			key:         "ssl_ctx_verify_depth",
			value:       []byte(""),
			expected:    nil,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ConvertGlobalStatus(tt.key, tt.value)
			require.Equal(t, tt.expectedErr, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestCovertGlobalVariables(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       sql.RawBytes
		expected    interface{}
		expectedErr error
	}{
		{
			name:        "boolean type mysql<=5.6",
			key:         "gtid_mode",
			value:       []byte("ON"),
			expected:    int64(1),
			expectedErr: nil,
		},
		{
			name:        "enum type mysql>=5.7",
			key:         "gtid_mode",
			value:       []byte("ON_PERMISSIVE"),
			expected:    int64(1),
			expectedErr: nil,
		},
		{
			name:        "defined variable but unset",
			key:         "ssl_ctx_verify_depth",
			value:       []byte(""),
			expected:    nil,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ConvertGlobalVariables(tt.key, tt.value)
			require.Equal(t, tt.expectedErr, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestParseValue(t *testing.T) {
	testCases := []struct {
		rawByte sql.RawBytes
		output  interface{}
		err     string
	}{
		{sql.RawBytes("123"), int64(123), ""},
		{sql.RawBytes("abc"), "abc", ""},
		{sql.RawBytes("10.1"), 10.1, ""},
		{sql.RawBytes("ON"), 1, ""},
		{sql.RawBytes("OFF"), 0, ""},
		{sql.RawBytes("NO"), 0, ""},
		{sql.RawBytes("YES"), 1, ""},
		{sql.RawBytes("No"), 0, ""},
		{sql.RawBytes("Yes"), 1, ""},
		{sql.RawBytes("-794"), int64(-794), ""},
		{sql.RawBytes("2147483647"), int64(2147483647), ""},                       // max int32
		{sql.RawBytes("2147483648"), int64(2147483648), ""},                       // too big for int32
		{sql.RawBytes("9223372036854775807"), int64(9223372036854775807), ""},     // max int64
		{sql.RawBytes("9223372036854775808"), uint64(9223372036854775808), ""},    // too big for int64
		{sql.RawBytes("18446744073709551615"), uint64(18446744073709551615), ""},  // max uint64
		{sql.RawBytes("18446744073709551616"), float64(18446744073709552000), ""}, // too big for uint64
		{sql.RawBytes("18446744073709552333"), float64(18446744073709552000), ""}, // too big for uint64
		{sql.RawBytes(""), nil, "unconvertible value"},
	}
	for _, cases := range testCases {
		got, err := ParseValue(cases.rawByte)

		if err != nil && cases.err == "" {
			t.Errorf("for %q got unexpected error: %q", string(cases.rawByte), err.Error())
		} else if err != nil && !strings.HasPrefix(err.Error(), cases.err) {
			t.Errorf("for %q wanted error %q, got %q", string(cases.rawByte), cases.err, err.Error())
		} else if err == nil && cases.err != "" {
			t.Errorf("for %q did not get expected error: %s", string(cases.rawByte), cases.err)
		} else if got != cases.output {
			t.Errorf("for %q wanted %#v (%T), got %#v (%T)", string(cases.rawByte), cases.output, cases.output, got, got)
		}
	}
}
