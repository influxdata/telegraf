package modbus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBCD64(t *testing.T) {
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
			name:     "64-bit maximum",
			inType:   "BCD64",
			data:     []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99},
			expected: uint64(9999999999999999),
		},
		{
			name:     "64-bit native",
			inType:   "BCD64",
			data:     []byte{0x00, 0x00, 0x20, 0x26, 0x11, 0x25, 0x22, 0x30},
			expected: uint64(202611252230),
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
			name:    "invalid digit UINT64 with scale",
			inType:  "BCD64",
			outType: "UINT64",
			scale:   0.5,
			data:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
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
