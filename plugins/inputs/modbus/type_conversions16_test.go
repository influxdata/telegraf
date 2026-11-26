package modbus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBCD16(t *testing.T) {
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
			name:     "output INT64",
			inType:   "BCD16",
			outType:  "INT64",
			data:     []byte{0x12, 0x34},
			expected: int64(1234),
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
			name:    "invalid digit native",
			inType:  "BCD16",
			data:    []byte{0x12, 0xa4},
			outType: "native",
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
