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

func TestBCD(t *testing.T) {
	tests := []struct {
		name      string
		inType    string
		outType   string
		byteOrder string
		scale     float64
		data      []byte
		expected  any
		// Expected error when creating the converter
		expectedErr string
	}{
		{
			name:     "16-bit native",
			inType:   "BCD16",
			data:     []byte{0x12, 0x34},
			expected: uint16(1234),
		},
		{
			name:      "16-bit little endian",
			inType:    "BCD16",
			byteOrder: "DCBA",
			data:      []byte{0x34, 0x12},
			expected:  uint16(1234),
		},
		{
			// Day, month and year of a date as in issue #15925
			name:     "32-bit native date",
			inType:   "BCD32",
			data:     []byte{0x00, 0x25, 0x11, 0x24},
			expected: uint32(251124),
		},
		{
			name:      "32-bit word swapped",
			inType:    "BCD32",
			byteOrder: "CDAB",
			data:      []byte{0x11, 0x24, 0x00, 0x25},
			expected:  uint32(251124),
		},
		{
			name:     "64-bit native",
			inType:   "BCD64",
			data:     []byte{0x00, 0x00, 0x20, 0x26, 0x11, 0x25, 0x22, 0x30},
			expected: uint64(202611252230),
		},
		{
			name:     "output INT64",
			inType:   "BCD16",
			outType:  "INT64",
			data:     []byte{0x12, 0x34},
			expected: int64(1234),
		},
		{
			name:     "output UINT64",
			inType:   "BCD32",
			outType:  "UINT64",
			data:     []byte{0x00, 0x25, 0x11, 0x24},
			expected: uint64(251124),
		},
		{
			name:     "output FLOAT64",
			inType:   "BCD16",
			outType:  "FLOAT64",
			data:     []byte{0x12, 0x34},
			expected: float64(1234),
		},
		{
			name:     "scale native",
			inType:   "BCD16",
			scale:    10,
			data:     []byte{0x01, 0x23},
			expected: uint16(1230),
		},
		{
			name:     "scale FLOAT64",
			inType:   "BCD16",
			outType:  "FLOAT64",
			scale:    0.1,
			data:     []byte{0x12, 0x34},
			expected: 123.4,
		},
		{
			name:     "scale INT64",
			inType:   "BCD32",
			outType:  "INT64",
			scale:    0.01,
			data:     []byte{0x00, 0x25, 0x11, 0x24},
			expected: int64(2511),
		},
		{
			name:     "scale UINT64",
			inType:   "BCD64",
			outType:  "UINT64",
			scale:    2,
			data:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x21},
			expected: uint64(42),
		},
		{
			name:    "invalid digit native",
			inType:  "BCD16",
			data:    []byte{0x12, 0xa4},
			outType: "native",
		},
		{
			name:    "invalid digit INT64",
			inType:  "BCD32",
			outType: "INT64",
			data:    []byte{0x00, 0x25, 0x1f, 0x24},
		},
		{
			name:    "invalid digit UINT64 with scale",
			inType:  "BCD64",
			outType: "UINT64",
			scale:   0.5,
			data:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:    "invalid digit FLOAT64 with scale",
			inType:  "BCD16",
			outType: "FLOAT64",
			scale:   0.1,
			data:    []byte{0xab, 0xcd},
		},
		{
			name:        "invalid output type",
			inType:      "BCD16",
			outType:     "STRING",
			expectedErr: "invalid output data-type",
		},
		{
			name:        "invalid output type with scale",
			inType:      "BCD32",
			outType:     "STRING",
			scale:       0.5,
			expectedErr: "invalid output data-type",
		},
		{
			name:        "invalid byte order",
			inType:      "BCD64",
			byteOrder:   "ABBA",
			expectedErr: "invalid byte-order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := tt.byteOrder
			if order == "" {
				order = "ABCD"
			}
			outType := tt.outType
			if outType == "" {
				outType = "native"
			}

			convert, err := determineConverter(tt.inType, order, outType, tt.scale, 0, "")
			if tt.expectedErr != "" {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)

			actual := convert(tt.data)
			if tt.expected == nil {
				// Not valid BCD: no value, so the field is left out of the metric
				require.Nil(t, actual)
				return
			}
			require.IsType(t, tt.expected, actual)
			if f, ok := tt.expected.(float64); ok {
				require.InEpsilon(t, f, actual, 1e-9)
				return
			}
			require.Equal(t, tt.expected, actual)
		})
	}
}
