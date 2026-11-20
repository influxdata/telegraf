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
			inType:   "FLOAT_UDEC_8E24M",
			data:     []byte{0xfd, 0x01, 0xe2, 0x40},
			expected: 123.456,
		},
		{
			name:     "signed example positive",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0xfd, 0x01, 0xe2, 0x40},
			expected: 123.456,
		},
		{
			name:     "signed example negative",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0xfd, 0xfe, 0x1d, 0xc0},
			expected: -123.456,
		},
		{
			name:     "unsigned mantissa with sign bit set",
			inType:   "FLOAT_UDEC_8E24M",
			data:     []byte{0xfd, 0xfe, 0x1d, 0xc0},
			expected: 16653.76,
		},
		{
			name:     "positive exponent",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0x02, 0x00, 0x00, 0x2a},
			expected: 4200,
		},
		{
			name:     "zero exponent",
			inType:   "FLOAT_UDEC_8E24M",
			data:     []byte{0x00, 0x00, 0x01, 0x00},
			expected: 256,
		},
		{
			name:     "negative exponent extreme",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0x80, 0x00, 0x00, 0x01},
			expected: 1e-128,
		},
		{
			name:     "signed minimum mantissa",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0x00, 0x80, 0x00, 0x00},
			expected: -8388608,
		},
		{
			name:      "byte order BADC",
			inType:    "FLOAT_DEC_8E24M",
			byteOrder: "BADC",
			data:      []byte{0x01, 0xfd, 0x40, 0xe2},
			expected:  123.456,
		},
		{
			name:      "byte order CDAB",
			inType:    "FLOAT_DEC_8E24M",
			byteOrder: "CDAB",
			data:      []byte{0xe2, 0x40, 0xfd, 0x01},
			expected:  123.456,
		},
		{
			name:      "byte order DCBA",
			inType:    "FLOAT_UDEC_8E24M",
			byteOrder: "DCBA",
			data:      []byte{0x40, 0xe2, 0x01, 0xfd},
			expected:  123.456,
		},
		{
			name:     "with scale",
			inType:   "FLOAT_DEC_8E24M",
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
	for _, in := range []string{"FLOAT_DEC_8E24M", "FLOAT_UDEC_8E24M"} {
		for _, scale := range []float64{0, 0.5} {
			_, err := determineConverter(in, "ABCD", "INT64", scale, 0, "")
			require.ErrorContains(t, err, "invalid output data-type")
		}
	}
}

func TestDecExp32FieldProperties(t *testing.T) {
	for _, in := range []string{"FLOAT_DEC_8E24M", "FLOAT_UDEC_8E24M"} {
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

func TestDecExp32RegisterStyleValidation(t *testing.T) {
	c := &configurationOriginal{}
	for _, in := range []string{"FLOAT_DEC_8E24M", "FLOAT_UDEC_8E24M"} {
		// The type requires exactly two registers
		require.NoError(t, validateFieldDefinitions([]fieldDefinition{
			{Name: "a", ByteOrder: "ABCD", DataType: in, Scale: 1, Address: []uint16{1, 2}},
		}, cHoldingRegisters))
		require.ErrorContains(t, validateFieldDefinitions([]fieldDefinition{
			{Name: "a", ByteOrder: "AB", DataType: in, Scale: 1, Address: []uint16{1}},
		}, cHoldingRegisters), "expecting 2 entries")
		require.ErrorContains(t, validateFieldDefinitions([]fieldDefinition{
			{Name: "a", ByteOrder: "ABCDEFGH", DataType: in, Scale: 1, Address: []uint16{1, 2, 3, 4}},
		}, cHoldingRegisters), "expecting 2 entries")

		out, err := c.normalizeOutputDatatype(in)
		require.NoError(t, err)
		require.Equal(t, "FLOAT64", out)
		typ, err := c.normalizeInputDatatype(in, 2)
		require.NoError(t, err)
		require.Equal(t, in, typ)
	}
}
