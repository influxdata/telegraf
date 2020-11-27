package v2

import (
	"database/sql"
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
			expected:    int64(0),
			expectedErr: nil,
		},
		{
			name:        "overflow int64",
			key:         "ssl_ctx_verify_depth",
			value:       []byte("18446744073709551615"),
			expected:    int64(9223372036854775807),
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
