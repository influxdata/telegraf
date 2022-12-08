package netflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeInt32(t *testing.T) {
	buf := []byte{0x82, 0xad, 0x80, 0x86}
	out, ok := decodeInt32(buf).(int64)
	require.True(t, ok)
	require.Equal(t, int64(-2102558586), out)
}

func TestDecodeUint(t *testing.T) {
	tests := []struct {
		name     string
		in       []byte
		expected uint64
	}{
		{
			name:     "uint8",
			in:       []byte{0x42},
			expected: 66,
		},
		{
			name:     "uint16",
			in:       []byte{0x0A, 0x42},
			expected: 2626,
		},
		{
			name:     "uint32",
			in:       []byte{0x82, 0xad, 0x80, 0x86},
			expected: 2192408710,
		},
		{
			name:     "uint64",
			in:       []byte{0x00, 0x00, 0x23, 0x42, 0x8f, 0xad, 0x80, 0x86},
			expected: 38768785326214,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, ok := decodeUint(tt.in).(uint64)
			require.True(t, ok)
			require.Equal(t, tt.expected, out)

		})
	}
}

func TestDecodeUintInvalid(t *testing.T) {
	require.Panics(t, func() { decodeUint([]byte{0x00, 0x00, 0x00}) })
}

func TestDecodeFloat64(t *testing.T) {
	buf := []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea}
	out, ok := decodeFloat64(buf).(float64)
	require.True(t, ok)
	require.Equal(t, float64(3.14159265359), out)
}

func TestDecodeBool(t *testing.T) {
	tests := []struct {
		name     string
		in       []byte
		expected interface{}
	}{
		{
			name:     "zero",
			in:       []byte{0x00},
			expected: uint8(0),
		},
		{
			name:     "true",
			in:       []byte{0x01},
			expected: true,
		},
		{
			name:     "false",
			in:       []byte{0x02},
			expected: false,
		},
		{
			name:     "other",
			in:       []byte{0x23},
			expected: uint8(35),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := decodeBool(tt.in)
			require.Equal(t, tt.expected, out)

		})
	}
}
