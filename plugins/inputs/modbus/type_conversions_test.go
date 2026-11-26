package modbus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeBCD(t *testing.T) {
	tests := []struct {
		name     string
		raw      uint64
		expected uint64
		valid    bool
	}{
		{name: "zero", raw: 0x0, expected: 0, valid: true},
		{name: "16-bit", raw: 0x1234, expected: 1234, valid: true},
		{name: "16-bit maximum", raw: 0x9999, expected: 9999, valid: true},
		{name: "32-bit maximum", raw: 0x99999999, expected: 99999999, valid: true},
		{name: "64-bit maximum", raw: 0x9999999999999999, expected: 9999999999999999, valid: true},
		{name: "leading zeros", raw: 0x00000042, expected: 42, valid: true},
		{name: "invalid digit in the middle", raw: 0x12a4},
		{name: "invalid most significant digit", raw: 0xf000},
		{name: "invalid least significant digit", raw: 0x000b},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, valid := decodeBCD(tt.raw)
			require.Equal(t, tt.valid, valid)
			if tt.valid {
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}
