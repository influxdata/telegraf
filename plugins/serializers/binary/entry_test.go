package binary

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal"
)

func TestSerialization(t *testing.T) {
	tests := []struct {
		name     string
		entry    *Entry
		input    interface{}
		expected map[binary.ByteOrder][]byte
		overflow bool
	}{
		{
			name:  "positive int serialization",
			entry: &Entry{Name: "test", DataFormat: "int32"},
			input: 1,
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x00, 0x00, 0x00, 0x01},
				binary.LittleEndian: {0x01, 0x00, 0x00, 0x00},
			},
		},
		{
			name:  "negative int serialization",
			entry: &Entry{Name: "test", DataFormat: "int32"},
			input: -1,
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0xff, 0xff, 0xff, 0xff},
				binary.LittleEndian: {0xff, 0xff, 0xff, 0xff},
			},
		},
		{
			name:  "negative int serialization | uint32 representation",
			entry: &Entry{Name: "test", DataFormat: "uint32"},
			input: -1,
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0xff, 0xff, 0xff, 0xff},
				binary.LittleEndian: {0xff, 0xff, 0xff, 0xff},
			},
			overflow: true,
		},
		{
			name:  "uint to int serialization",
			entry: &Entry{Name: "test", DataFormat: "uint8"},
			input: uint(1),
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x01},
				binary.LittleEndian: {0x01},
			},
		},
		{
			name:  "string serialization",
			entry: &Entry{Name: "test", DataFormat: "string", StringLength: 4},
			input: "test",
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x74, 0x65, 0x73, 0x00},
				binary.LittleEndian: {0x74, 0x65, 0x73, 0x00},
			},
		},
		{
			name:  "string serialization with terminator",
			entry: &Entry{Name: "test", DataFormat: "string", StringLength: 5, StringTerminator: "null"},
			input: "test",
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x74, 0x65, 0x73, 0x74, 0x00},
				binary.LittleEndian: {0x74, 0x65, 0x73, 0x74, 0x00},
			},
		},
		{
			name:  "string serialization with hex terminator",
			entry: &Entry{Name: "test", DataFormat: "string", StringLength: 5, StringTerminator: "0x01"},
			input: "test",
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x74, 0x65, 0x73, 0x74, 0x01},
				binary.LittleEndian: {0x74, 0x65, 0x73, 0x74, 0x01},
			},
		},
		{
			name:  "time serialization",
			entry: &Entry{ReadFrom: "time", DataFormat: "uint64"},
			input: time.Date(2024, time.January, 6, 19, 44, 10, 0, time.UTC),
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x00, 0x00, 0x00, 0x00, 0x65, 0x99, 0xad, 0x8a},
				binary.LittleEndian: {0x8a, 0xad, 0x99, 0x65, 0x00, 0x00, 0x00, 0x00},
			},
		},
		{
			name:  "float32 serialization",
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: float32(3.1415),
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x40, 0x49, 0x0e, 0x56},
				binary.LittleEndian: {0x56, 0x0e, 0x49, 0x40},
			},
		},
		{
			name:  "float32 serialization | float64 representation",
			entry: &Entry{Name: "test", DataFormat: "float64"},
			input: float32(3.1415),
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x40, 0x09, 0x21, 0xCA, 0xC0, 0x00, 0x00, 0x00},
				binary.LittleEndian: {0x00, 0x00, 0x00, 0xC0, 0xCA, 0x21, 0x09, 0x40},
			},
		},
		{
			name:  "float64 serialization",
			entry: &Entry{Name: "test", DataFormat: "float64"},
			input: 3.1415,
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x40, 0x09, 0x21, 0xCA, 0xC0, 0x83, 0x12, 0x6F},
				binary.LittleEndian: {0x6F, 0x12, 0x83, 0xC0, 0xCA, 0x21, 0x09, 0x40},
			},
		},
		{
			name:  "float64 serialization | float32 representation",
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: 3.1415,
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x40, 0x49, 0x0e, 0x56},
				binary.LittleEndian: {0x56, 0x0e, 0x49, 0x40},
			},
		},
		{
			name:  "float64 serialization | int64 representation",
			entry: &Entry{Name: "test", DataFormat: "int64"},
			input: 3.1415,
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
				binary.LittleEndian: {0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
		},
		{
			name:  "float64 serialization | uint8 representation",
			entry: &Entry{Name: "test", DataFormat: "uint8"},
			input: 3.1415,
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian: {0x03}, binary.LittleEndian: {0x03},
			},
		},
		{
			name:  "uint serialization | float32 representation",
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: uint(1),
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x3f, 0x80, 0x00, 0x00},
				binary.LittleEndian: {0x00, 0x00, 0x80, 0x3f},
			},
		},
		{
			name:  "uint serialization | float64 representation",
			entry: &Entry{Name: "test", DataFormat: "float64"},
			input: uint(1),
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				binary.LittleEndian: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x3f},
			},
		},
		{
			name:  "int serialization | float32 representation",
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: -101,
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0xc2, 0xca, 0x00, 0x00},
				binary.LittleEndian: {0x00, 0x00, 0xca, 0xc2},
			},
		},
		{
			name:  "string serialization | float32 representation",
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: "-101.25",
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0xc2, 0xca, 0x80, 0x00},
				binary.LittleEndian: {0x00, 0x80, 0xca, 0xc2},
			},
		},
		{
			name:  "string serialization | int32 representation",
			entry: &Entry{Name: "test", DataFormat: "int32"},
			input: "1",
			expected: map[binary.ByteOrder][]byte{
				binary.BigEndian:    {0x00, 0x00, 0x00, 0x01},
				binary.LittleEndian: {0x01, 0x00, 0x00, 0x00},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, tc.entry.fillDefaults())

			for endianness, expected := range tc.expected {
				value, err := tc.entry.serializeValue(tc.input, endianness)
				if tc.overflow {
					require.ErrorIs(t, err, internal.ErrOutOfRange)
				} else {
					require.NoError(t, err)
				}
				require.Equal(t, expected, value)
			}
		})
	}
}

func TestNoNameSerialization(t *testing.T) {
	e := &Entry{}
	require.ErrorContains(t, e.fillDefaults(), "missing name")
}

func BenchmarkSerialization(b *testing.B) {
	entries := []struct {
		entry *Entry
		input interface{}
	}{
		{
			entry: &Entry{Name: "test", DataFormat: "int32"},
			input: 1,
		},
		{
			entry: &Entry{Name: "test", DataFormat: "int32"},
			input: -1,
		},
		{
			entry: &Entry{Name: "test", DataFormat: "uint8"},
			input: uint(1),
		},
		{
			entry: &Entry{Name: "test", DataFormat: "string", StringLength: 4},
			input: "test",
		},
		{
			entry: &Entry{Name: "test", DataFormat: "string", StringLength: 5, StringTerminator: "null"},
			input: "test",
		},
		{
			entry: &Entry{Name: "test", DataFormat: "string", StringLength: 5, StringTerminator: "0x01"},
			input: "test",
		},
		{
			entry: &Entry{ReadFrom: "time", DataFormat: "uint64"},
			input: time.Date(2024, time.January, 6, 19, 44, 10, 0, time.UTC),
		},
		{
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: float32(3.1415),
		},
		{
			entry: &Entry{Name: "test", DataFormat: "float64"},
			input: float32(3.1415),
		},
		{
			entry: &Entry{Name: "test", DataFormat: "float64"},
			input: 3.1415,
		},
		{
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: 3.1415,
		},
		{
			entry: &Entry{Name: "test", DataFormat: "int64"},
			input: 3.1415,
		},
		{
			entry: &Entry{Name: "test", DataFormat: "uint8"},
			input: 3.1415,
		},
		{
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: uint(1),
		},
		{
			entry: &Entry{Name: "test", DataFormat: "float64"},
			input: uint(1),
		},
		{
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: -101,
		},
		{
			entry: &Entry{Name: "test", DataFormat: "float32"},
			input: "-101.25",
		},
		{
			entry: &Entry{Name: "test", DataFormat: "int32"},
			input: "1",
		},
	}

	for _, tc := range entries {
		require.NoError(b, tc.entry.fillDefaults())
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range entries {
			_, err := tc.entry.serializeValue(tc.input, binary.BigEndian)
			require.NoError(b, err)
		}
	}
}
