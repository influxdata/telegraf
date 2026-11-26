package modbus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloats(t *testing.T) {
	tests := []struct {
		name      string
		inType    string
		outType   string
		byteOrder string
		scale     float64
		data      []byte
		expected  float64
		// Expected error when creating the converter
		expectedErr string
	}{
		{
			name:     "unsigned decade example",
			inType:   "FLOAT_UDEC_8E24M",
			data:     []byte{0xfd, 0x01, 0xe2, 0x40},
			expected: 123.456,
		},
		{
			name:     "signed decade example positive",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0xfd, 0x01, 0xe2, 0x40},
			expected: 123.456,
		},
		{
			name:     "signed decade example negative",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0xfd, 0xfe, 0x1d, 0xc0},
			expected: -123.456,
		},
		{
			name:     "unsigned decade mantissa with sign bit set",
			inType:   "FLOAT_UDEC_8E24M",
			data:     []byte{0xfd, 0xfe, 0x1d, 0xc0},
			expected: 16653.76,
		},
		{
			name:     "decade positive exponent",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0x02, 0x00, 0x00, 0x2a},
			expected: 4200,
		},
		{
			name:     "decade zero exponent",
			inType:   "FLOAT_UDEC_8E24M",
			data:     []byte{0x00, 0x00, 0x01, 0x00},
			expected: 256,
		},
		{
			name:     "decade negative exponent extreme",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0x80, 0x00, 0x00, 0x01},
			expected: 1e-128,
		},
		{
			name:     "decade signed minimum mantissa",
			inType:   "FLOAT_DEC_8E24M",
			data:     []byte{0x00, 0x80, 0x00, 0x00},
			expected: -8388608,
		},
		{
			name:      "decade byte order BADC",
			inType:    "FLOAT_DEC_8E24M",
			byteOrder: "BADC",
			data:      []byte{0x01, 0xfd, 0x40, 0xe2},
			expected:  123.456,
		},
		{
			name:      "decade byte order CDAB",
			inType:    "FLOAT_DEC_8E24M",
			byteOrder: "CDAB",
			data:      []byte{0xe2, 0x40, 0xfd, 0x01},
			expected:  123.456,
		},
		{
			name:      "decade byte order DCBA",
			inType:    "FLOAT_UDEC_8E24M",
			byteOrder: "DCBA",
			data:      []byte{0x40, 0xe2, 0x01, 0xfd},
			expected:  123.456,
		},
		{
			name:     "decade with scale",
			inType:   "FLOAT_DEC_8E24M",
			scale:    2,
			data:     []byte{0xfd, 0xfe, 0x1d, 0xc0},
			expected: -246.912,
		},
		{
			name:     "decade explicit float64 output",
			inType:   "FLOAT_DEC_8E24M",
			outType:  "FLOAT64",
			data:     []byte{0xfd, 0x01, 0xe2, 0x40},
			expected: 123.456,
		},
		{
			name:        "signed decade invalid output type",
			inType:      "FLOAT_DEC_8E24M",
			outType:     "INT64",
			expectedErr: "invalid output data-type",
		},
		{
			name:        "unsigned decade invalid output type with scale",
			inType:      "FLOAT_UDEC_8E24M",
			outType:     "INT64",
			scale:       0.5,
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
			require.IsType(t, float64(0), actual)
			require.InEpsilon(t, tt.expected, actual, 1e-9)
		})
	}
}

func TestFloatsFieldProperties(t *testing.T) {
	tests := []struct {
		name        string
		inType      string
		addresses   []uint16
		byteOrder   string
		expectedErr string
	}{
		{
			name:      "signed decade",
			inType:    "FLOAT_DEC_8E24M",
			addresses: []uint16{1, 2},
			byteOrder: "ABCD",
		},
		{
			name:      "unsigned decade",
			inType:    "FLOAT_UDEC_8E24M",
			addresses: []uint16{1, 2},
			byteOrder: "ABCD",
		},
		{
			name:        "decade single register",
			inType:      "FLOAT_DEC_8E24M",
			addresses:   []uint16{1},
			byteOrder:   "AB",
			expectedErr: "expecting 2 entries",
		},
		{
			name:        "decade four registers",
			inType:      "FLOAT_UDEC_8E24M",
			addresses:   []uint16{1, 2, 3, 4},
			byteOrder:   "ABCDEFGH",
			expectedErr: "expecting 2 entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check the field properties of the request and metric style
			out, err := determineOutputDatatype(tt.inType)
			require.NoError(t, err)
			require.Equal(t, "FLOAT64", out)

			length, err := determineFieldLength(tt.inType, 0)
			require.NoError(t, err)
			require.Equal(t, uint16(2), length)

			pm := &configurationPerMetric{}
			out, err = pm.determineOutputDatatype(tt.inType)
			require.NoError(t, err)
			require.Equal(t, "FLOAT64", out)
			length, err = pm.determineFieldLength(tt.inType, 0)
			require.NoError(t, err)
			require.Equal(t, uint16(2), length)

			// Check the register style
			c := &configurationOriginal{}
			out, err = c.normalizeOutputDatatype(tt.inType)
			require.NoError(t, err)
			require.Equal(t, "FLOAT64", out)

			err = validateFieldDefinitions([]fieldDefinition{
				{Name: "a", ByteOrder: tt.byteOrder, DataType: tt.inType, Scale: 1, Address: tt.addresses},
			}, cHoldingRegisters)
			if tt.expectedErr != "" {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
