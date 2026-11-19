package modbus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecExp32(t *testing.T) {
	tests := []struct {
		name      string
		inType    string
		byteOrder string
		scale     float64
		data      []byte
		expected  float64
	}{
		{
			name:     "unsigned example",
			inType:   "UDECEXP32",
			data:     []byte{0xfd, 0x01, 0xe2, 0x40},
			expected: 123.456,
		},
		{
			name:     "signed example positive",
			inType:   "DECEXP32",
			data:     []byte{0xfd, 0x01, 0xe2, 0x40},
			expected: 123.456,
		},
		{
			name:     "signed example negative",
			inType:   "DECEXP32",
			data:     []byte{0xfd, 0xfe, 0x1d, 0xc0},
			expected: -123.456,
		},
		{
			name:     "unsigned mantissa with sign bit set",
			inType:   "UDECEXP32",
			data:     []byte{0xfd, 0xfe, 0x1d, 0xc0},
			expected: 16653.76,
		},
		{
			name:     "positive exponent",
			inType:   "DECEXP32",
			data:     []byte{0x02, 0x00, 0x00, 0x2a},
			expected: 4200,
		},
		{
			name:     "zero exponent",
			inType:   "UDECEXP32",
			data:     []byte{0x00, 0x00, 0x01, 0x00},
			expected: 256,
		},
		{
			name:     "negative exponent extreme",
			inType:   "DECEXP32",
			data:     []byte{0x80, 0x00, 0x00, 0x01},
			expected: 1e-128,
		},
		{
			name:     "signed minimum mantissa",
			inType:   "DECEXP32",
			data:     []byte{0x00, 0x80, 0x00, 0x00},
			expected: -8388608,
		},
		{
			name:      "byte order BADC",
			inType:    "DECEXP32",
			byteOrder: "BADC",
			data:      []byte{0x01, 0xfd, 0x40, 0xe2},
			expected:  123.456,
		},
		{
			name:      "byte order CDAB",
			inType:    "DECEXP32",
			byteOrder: "CDAB",
			data:      []byte{0xe2, 0x40, 0xfd, 0x01},
			expected:  123.456,
		},
		{
			name:      "byte order DCBA",
			inType:    "UDECEXP32",
			byteOrder: "DCBA",
			data:      []byte{0x40, 0xe2, 0x01, 0xfd},
			expected:  123.456,
		},
		{
			name:     "with scale",
			inType:   "DECEXP32",
			scale:    2,
			data:     []byte{0xfd, 0xfe, 0x1d, 0xc0},
			expected: -246.912,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := tt.byteOrder
			if order == "" {
				order = "ABCD"
			}
			for _, outType := range []string{"native", "FLOAT64"} {
				convert, err := determineConverter(tt.inType, order, outType, tt.scale, 0, "")
				require.NoError(t, err)
				require.InDelta(t, tt.expected, convert(tt.data), 1e-9*max(1, abs(tt.expected)))
				require.IsType(t, float64(0), convert(tt.data))
			}
		})
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func TestDecExp32InvalidOutputType(t *testing.T) {
	for _, in := range []string{"DECEXP32", "UDECEXP32"} {
		for _, scale := range []float64{0, 0.5} {
			_, err := determineConverter(in, "ABCD", "INT64", scale, 0, "")
			require.ErrorContains(t, err, "invalid output data-type")
		}
	}
}

func TestDecExp32FieldProperties(t *testing.T) {
	for _, in := range []string{"DECEXP32", "UDECEXP32"} {
		out, err := determineOutputDatatype(in)
		require.NoError(t, err)
		require.Equal(t, "FLOAT64", out)

		length, err := determineFieldLength(in, 0)
		require.NoError(t, err)
		require.Equal(t, uint16(2), length)

		pm := &configurationPerMetric{}
		out, err = pm.determineOutputDatatype(in)
		require.NoError(t, err)
		require.Equal(t, "FLOAT64", out)
		length, err = pm.determineFieldLength(in, 0)
		require.NoError(t, err)
		require.Equal(t, uint16(2), length)
	}
}
