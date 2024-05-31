package snmp

import (
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name       string
		conversion string
		ent        gosnmp.SnmpPDU
		expected   interface{}
		errmsg     string
	}{
		{
			name: "integer",
			ent: gosnmp.SnmpPDU{
				Type:  gosnmp.Integer,
				Value: int(2),
			},
			expected: 2,
		},
		{
			name: "gauge",
			ent: gosnmp.SnmpPDU{
				Type:  gosnmp.Gauge32,
				Value: uint(2),
			},
			expected: 2,
		},
		{
			name: "octet string with valid bytes",
			ent: gosnmp.SnmpPDU{
				Type:  gosnmp.OctetString,
				Value: []uint8{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x77, 0x6F, 0x72, 0x6C, 0x64},
			},
			expected: "Hello world",
		},
		{
			name: "octet string with invalid bytes",
			ent: gosnmp.SnmpPDU{
				Type:  gosnmp.OctetString,
				Value: []uint8{0x84, 0xc8, 0x7, 0xff, 0xfd, 0x38, 0x54, 0xc1},
			},
			expected: "84c807fffd3854c1",
		},
		{
			name:       "hextoint empty",
			conversion: "hextoint:BigEndian:uint64",
			ent:        gosnmp.SnmpPDU{},
		},
		{
			name:       "hextoint big endian uint64",
			conversion: "hextoint:BigEndian:uint64",
			ent: gosnmp.SnmpPDU{
				Type:  gosnmp.OctetString,
				Value: []uint8{0x84, 0xc8, 0x7, 0xff, 0xfd, 0x38, 0x54, 0xc1},
			},
			expected: uint64(0x84c807fffd3854c1),
		},
		{
			name:       "hextoint big endian invalid",
			conversion: "hextoint:BigEndian:invalid",
			ent:        gosnmp.SnmpPDU{Type: gosnmp.OctetString, Value: []uint8{}},
			errmsg:     "invalid bit value",
		},
		{
			name:       "hextoint little endian uint64",
			conversion: "hextoint:LittleEndian:uint64",
			ent: gosnmp.SnmpPDU{
				Type:  gosnmp.OctetString,
				Value: []uint8{0x84, 0xc8, 0x7, 0xff, 0xfd, 0x38, 0x54, 0xc1},
			},
			expected: uint64(0xc15438fdff07c884),
		},
		{
			name:       "hextoint little endian invalid",
			conversion: "hextoint:LittleEndian:invalid",
			ent:        gosnmp.SnmpPDU{Type: gosnmp.OctetString, Value: []uint8{}},
			errmsg:     "invalid bit value",
		},
		{
			name:       "hextoint invalid",
			conversion: "hextoint:invalid:uint64",
			ent:        gosnmp.SnmpPDU{Type: gosnmp.OctetString, Value: []uint8{}},
			errmsg:     "invalid Endian value",
		},
		{
			name:       "invalid",
			conversion: "invalid",
			ent:        gosnmp.SnmpPDU{},
			errmsg:     "invalid conversion type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Field{Conversion: tt.conversion}

			actual, err := f.Convert(tt.ent)

			if tt.errmsg != "" {
				require.ErrorContains(t, err, tt.errmsg)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expected, actual)
		})
	}
}
