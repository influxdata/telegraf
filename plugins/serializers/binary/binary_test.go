package binary

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/metric"
)

func TestMetricSerialization(t *testing.T) {
	m := metric.New(
		"modbus",
		map[string]string{
			"tag_1": "ABC",
			"tag_2": "1.63",
		},
		map[string]interface{}{
			"addr_2":     7,
			"addr_3":     17001,
			"addr_4_5":   617001,
			"addr_6_7":   423.1700134277344,
			"addr_16_20": "A_B_C_D_E_",
			"addr_3_sc":  1700.1000000000001,
		},
		time.Unix(1703018620, 0),
	)

	tests := []struct {
		name     string
		entries  []*Entry
		expected map[string]string
	}{
		{
			name: "complex metric serialization",
			entries: []*Entry{
				{
					ReadFrom:   "field",
					Name:       "addr_3",
					DataFormat: "int16",
				},
				{
					ReadFrom:   "field",
					Name:       "addr_2",
					DataFormat: "int16",
				},
				{
					ReadFrom:   "field",
					Name:       "addr_4_5",
					DataFormat: "int32",
				},
				{
					ReadFrom:   "field",
					Name:       "addr_6_7",
					DataFormat: "float32",
				},
				{
					ReadFrom:         "field",
					Name:             "addr_16_20",
					DataFormat:       "string",
					StringTerminator: "null",
					StringLength:     11,
				},
				{
					ReadFrom:   "field",
					Name:       "addr_3_sc",
					DataFormat: "float64",
				},
				{
					ReadFrom:   "time",
					DataFormat: "int32",
					TimeFormat: "unix",
				},
				{
					ReadFrom:         "name",
					DataFormat:       "string",
					StringTerminator: "null",
					StringLength:     20,
				},
				{
					ReadFrom:     "tag",
					Name:         "tag_1",
					DataFormat:   "string",
					StringLength: 4,
				},
				{
					ReadFrom:   "tag",
					Name:       "tag_2",
					DataFormat: "float32",
				},
			},
			expected: map[string]string{
				"little": "69420700296a0900c395d343415f425f435f445f455f006766666666909a407c" +
					"0082656d6f64627573000000000000000000000000000041424300d7a3d03f",
				"big": "4269000700096a2943d395c3415f425f435f445f455f00409a90666666666765" +
					"82007c6d6f646275730000000000000000000000000000414243003fd0a3d7",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for endianness, expected := range tc.expected {
				serializer := &Serializer{
					Entries:    tc.entries,
					Endianness: endianness,
				}

				require.NoError(t, serializer.Init())

				serialized, err := serializer.Serialize(m)
				actual := hex.EncodeToString(serialized)

				require.NoError(t, err)
				require.Equal(t, expected, actual)
			}
		})
	}
}
