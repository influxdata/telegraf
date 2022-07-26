package binary

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

var dummyEntry = Entry{
	Name:       "dummy",
	Type:       "uint8",
	Bits:       8,
	Assignment: "field",
}

func generateBinary(data []interface{}, order binary.ByteOrder) ([]byte, error) {
	var buf bytes.Buffer

	for _, x := range data {
		var err error
		switch v := x.(type) {
		case []byte:
			_, err = buf.Write(v)
		case string:
			_, err = buf.WriteString(v)
		default:
			err = binary.Write(&buf, order, x)
		}
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func determineEndianess(endianess string) binary.ByteOrder {
	switch endianess {
	case "le":
		return binary.LittleEndian
	case "be":
		return binary.BigEndian
	case "host":
		return hostEndianess
	}
	panic(fmt.Errorf("unknown endianess %q", endianess))
}

func TestInitInvalid(t *testing.T) {
	var tests = []struct {
		name      string
		metric    string
		config    []Config
		endianess string
		expected  string
	}{
		{
			name:      "wrong endianess",
			metric:    "binary",
			endianess: "garbage",
			expected:  `unknown endianess "garbage"`,
		},
		{
			name:      "empty configuration",
			metric:    "binary",
			endianess: "host",
			expected:  `no configuration given`,
		},
		{
			name:      "no metric name",
			config:    []Config{{}},
			endianess: "host",
			expected:  `config 0 invalid: no metric name given`,
		},
		{
			name:     "no field",
			config:   []Config{{}},
			metric:   "binary",
			expected: `config 0 invalid: no field defined`,
		},
		{
			name: "invalid entry",
			config: []Config{{
				Entries: []Entry{
					{
						Bits: 8,
					},
				},
			}},
			metric:   "binary",
			expected: `config 0 invalid: entry "" (0): missing name`,
		},
		{
			name: "multiple measurements",
			config: []Config{{
				Entries: []Entry{
					{
						Bits:       8,
						Assignment: "measurement",
					},
					{
						Bits:       8,
						Assignment: "measurement",
					},
				},
			}},
			metric:   "binary",
			expected: `config 0 invalid: multiple definitions of "measurement"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				Endianess:  tt.endianess,
				Log:        testutil.Logger{Name: "parsers.binary"},
				metricName: tt.metric,
			}

			parser.Configs = tt.config
			require.EqualError(t, parser.Init(), tt.expected)
		})
	}
}

func TestFilterInvalid(t *testing.T) {
	var tests = []struct {
		name     string
		filter   *Filter
		expected string
	}{
		{
			name:     "both length and length-min",
			filter:   &Filter{Length: 35, LengthMin: 33},
			expected: `config 0 invalid: length and length_min cannot be used together`,
		},
		{
			name:     "filter too long length",
			filter:   &Filter{Length: 3, Selection: []BinaryPart{{Offset: 16, Bits: 16}}},
			expected: `config 0 invalid: filter length (4) larger than constraint (3)`,
		},
		{
			name:     "filter invalid match",
			filter:   &Filter{Selection: []BinaryPart{{Offset: 16, Bits: 16, Match: "XYZ"}}},
			expected: `config 0 invalid: decoding match 0 failed: encoding/hex: invalid byte: U+0058 'X'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				Configs:    []Config{{Filter: tt.filter}},
				Log:        testutil.Logger{Name: "parsers.binary"},
				metricName: "binary",
			}
			require.EqualError(t, parser.Init(), tt.expected)
		})
	}
}

func TestFilterMatchInvalid(t *testing.T) {
	testdata := []byte{0x01, 0x02}

	var tests = []struct {
		name     string
		filter   *Filter
		expected string
	}{
		{
			name:     "filter length mismatch",
			filter:   &Filter{Selection: []BinaryPart{{Offset: 0, Bits: 8, Match: "0x0102"}}},
			expected: `no matching configuration`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				Configs:    []Config{{Filter: tt.filter, Entries: []Entry{{Name: "test", Type: "uint8"}}}},
				Log:        testutil.Logger{Name: "parsers.binary"},
				metricName: "binary",
			}
			require.NoError(t, parser.Init())
			_, err := parser.Parse(testdata)
			require.EqualError(t, err, tt.expected)
		})
	}
}

func TestFilterNoMatch(t *testing.T) {
	testdata := []interface{}{uint16(0x0102)}

	t.Run("no match error", func(t *testing.T) {
		parser := &Parser{
			Configs: []Config{
				{
					Filter:  &Filter{Length: 32},
					Entries: []Entry{dummyEntry},
				},
			},
			Log:        testutil.Logger{Name: "parsers.binary"},
			metricName: "binary",
		}
		require.NoError(t, parser.Init())

		data, err := generateBinary(testdata, hostEndianess)
		require.NoError(t, err)

		_, err = parser.Parse(data)
		require.EqualError(t, err, "no matching configuration")
	})

	t.Run("no match allow", func(t *testing.T) {
		parser := &Parser{
			AllowNoMatch: true,
			Configs: []Config{
				{
					Filter:  &Filter{Length: 32},
					Entries: []Entry{dummyEntry},
				},
			},
			Log:        testutil.Logger{Name: "parsers.binary"},
			metricName: "binary",
		}
		require.NoError(t, parser.Init())

		data, err := generateBinary(testdata, hostEndianess)
		require.NoError(t, err)

		metrics, err := parser.Parse(data)
		require.NoError(t, err)
		require.Empty(t, metrics)
	})
}

func TestFilterNone(t *testing.T) {
	testdata := []interface{}{
		uint64(0x01020304050607),
		uint64(0x08090A0B0C0D0E),
		uint64(0x0F101213141516),
		uint64(0x1718191A1B1C1D),
		uint64(0x1E1F2021222324),
	}

	var tests = []struct {
		name      string
		data      []interface{}
		filter    *Filter
		endianess string
	}{
		{
			name:      "no filter (BE)",
			data:      testdata,
			filter:    nil,
			endianess: "be",
		},
		{
			name:      "no filter (LE)",
			data:      testdata,
			filter:    nil,
			endianess: "le",
		},
		{
			name:      "no filter (host)",
			data:      testdata,
			filter:    nil,
			endianess: "host",
		},
		{
			name:      "empty filter (BE)",
			data:      testdata,
			filter:    &Filter{},
			endianess: "be",
		},
		{
			name:      "empty filter (LE)",
			data:      testdata,
			filter:    &Filter{},
			endianess: "le",
		},
		{
			name:      "empty filter (host)",
			data:      testdata,
			filter:    &Filter{},
			endianess: "host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				Endianess: tt.endianess,
				Configs: []Config{
					{
						Filter:  tt.filter,
						Entries: []Entry{dummyEntry},
					},
				},
				Log:        testutil.Logger{Name: "parsers.binary"},
				metricName: "binary",
			}
			require.NoError(t, parser.Init())

			order := determineEndianess(tt.endianess)
			data, err := generateBinary(tt.data, order)
			require.NoError(t, err)

			metrics, err := parser.Parse(data)
			require.NoError(t, err)
			require.NotEmpty(t, metrics)
		})
	}
}

func TestFilterLength(t *testing.T) {
	testdata := []interface{}{
		uint64(0x01020304050607),
		uint64(0x08090A0B0C0D0E),
		uint64(0x0F101213141516),
		uint64(0x1718191A1B1C1D),
		uint64(0x1E1F2021222324),
	}

	var tests = []struct {
		name     string
		data     []interface{}
		filter   *Filter
		expected bool
	}{
		{
			name:     "length match",
			data:     testdata,
			filter:   &Filter{Length: 40},
			expected: true,
		},
		{
			name:     "length no match too short",
			data:     testdata,
			filter:   &Filter{Length: 41},
			expected: false,
		},
		{
			name:     "length no match too long",
			data:     testdata,
			filter:   &Filter{Length: 39},
			expected: false,
		},
		{
			name:     "length min match",
			data:     testdata,
			filter:   &Filter{LengthMin: 40},
			expected: true,
		},
		{
			name:     "length min no match too short",
			data:     testdata,
			filter:   &Filter{LengthMin: 41},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				AllowNoMatch: true,
				Configs: []Config{
					{
						Filter:  tt.filter,
						Entries: []Entry{dummyEntry},
					},
				},
				Log:        testutil.Logger{Name: "parsers.binary"},
				metricName: "binary",
			}
			require.NoError(t, parser.Init())

			data, err := generateBinary(tt.data, hostEndianess)
			require.NoError(t, err)

			metrics, err := parser.Parse(data)
			require.NoError(t, err)
			if tt.expected {
				require.NotEmpty(t, metrics)
			} else {
				require.Empty(t, metrics)
			}
		})
	}
}

func TestFilterContent(t *testing.T) {
	testdata := [][]byte{
		{0x01, 0x02, 0x03, 0xA4, 0x05, 0x06, 0x07, 0x08},
		{0x01, 0xA2, 0x03, 0x04, 0x15, 0x01, 0x07, 0x08},
		{0xF1, 0xB1, 0x03, 0xA4, 0x25, 0x06, 0x07, 0x08},
		{0xF1, 0xC2, 0x03, 0x04, 0x35, 0x01, 0x07, 0x08},
		{0x42, 0xD1, 0x03, 0xA4, 0x25, 0x06, 0x42, 0x08},
		{0x42, 0xE2, 0x03, 0x04, 0x35, 0x01, 0x42, 0x08},
		{0x01, 0x00, 0x00, 0xA4},
	}
	var tests = []struct {
		name     string
		filter   *Filter
		expected int
	}{
		{
			name: "first byte",
			filter: &Filter{
				Selection: []BinaryPart{
					{
						Offset: 0,
						Bits:   8,
						Match:  "0xF1",
					},
				},
			},
			expected: 2,
		},
		{
			name: "last byte",
			filter: &Filter{
				Selection: []BinaryPart{
					{
						Offset: 7 * 8,
						Bits:   8,
						Match:  "0x08",
					},
				},
			},
			expected: 6,
		},
		{
			name: "none-byte boundary begin",
			filter: &Filter{
				Selection: []BinaryPart{
					{
						Offset: 12,
						Bits:   12,
						Match:  "0x0203",
					},
				},
			},
			expected: 4,
		},
		{
			name: "none-byte boundary end",
			filter: &Filter{
				Selection: []BinaryPart{
					{
						Offset: 16,
						Bits:   12,
						Match:  "0x003A",
					},
				},
			},
			expected: 3,
		},
		{
			name: "none-byte boundary end",
			filter: &Filter{
				Selection: []BinaryPart{
					{
						Offset: 36,
						Bits:   8,
						Match:  "0x50",
					},
				},
			},
			expected: 6,
		},
		{
			name: "multiple elements",
			filter: &Filter{
				Selection: []BinaryPart{
					{
						Offset: 4,
						Bits:   4,
						Match:  "0x01",
					},
					{
						Offset: 24,
						Bits:   8,
						Match:  "0xA4",
					},
				},
			},
			expected: 3,
		},
		{
			name: "multiple elements and length",
			filter: &Filter{
				Selection: []BinaryPart{
					{
						Offset: 4,
						Bits:   4,
						Match:  "0x01",
					},
					{
						Offset: 24,
						Bits:   8,
						Match:  "0xA4",
					},
				},
				Length: 4,
			},
			expected: 1,
		},
		{
			name: "multiple elements and length-min",
			filter: &Filter{
				Selection: []BinaryPart{
					{
						Offset: 4,
						Bits:   4,
						Match:  "0x01",
					},
					{
						Offset: 24,
						Bits:   8,
						Match:  "0xA4",
					},
				},
				LengthMin: 5,
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				AllowNoMatch: true,
				Configs: []Config{
					{
						Filter:  tt.filter,
						Entries: []Entry{dummyEntry},
					},
				},
				Log:        testutil.Logger{Name: "parsers.binary"},
				metricName: "binary",
			}
			require.NoError(t, parser.Init())

			var metrics []telegraf.Metric
			for _, data := range testdata {
				m, err := parser.Parse(data)
				require.NoError(t, err)
				metrics = append(metrics, m...)
			}
			require.Len(t, metrics, tt.expected)
		})
	}
}

func TestParseLineInvalid(t *testing.T) {
	var tests = []struct {
		name     string
		data     []interface{}
		configs  []Config
		expected string
	}{
		{
			name: "out-of-bounds",
			data: []interface{}{
				"2022-07-25T20:41:29+02:00", // time
				uint16(0x0102),              // address
				float64(42.123),             // value
			},
			configs: []Config{
				{
					Entries: []Entry{
						{
							Type:       "2006-01-02T15:04:05Z07:00",
							Assignment: "time",
							Timezone:   "UTC",
						},
						{
							Type: "uint32",
							Omit: true,
						},
						{
							Name: "value",
							Type: "float64",
						},
					},
				},
			},
			expected: `out-of-bounds @232 with 64 bits`,
		},
		{
			name: "multiple matches",
			data: []interface{}{
				"2022-07-25T20:41:29+02:00", // time
				uint16(0x0102),              // address
				float64(42.123),             // value
			},
			configs: []Config{
				{
					Entries: []Entry{
						{
							Type:       "2006-01-02T15:04:05Z07:00",
							Assignment: "time",
							Timezone:   "UTC",
						},
						{
							Type: "uint16",
							Omit: true,
						},
						{
							Name: "value",
							Type: "float64",
						},
					},
				},
				{
					Entries: []Entry{
						{
							Type:       "2006-01-02T15:04:05Z07:00",
							Assignment: "time",
							Timezone:   "UTC",
						},
						{
							Name:       "address",
							Type:       "uint16",
							Assignment: "tag",
						},
						{
							Name: "value",
							Type: "float64",
						},
					},
				},
			},
			expected: `cannot parse line with multiple (2) metrics`,
		},
	}

	for _, tt := range tests {
		for _, endianess := range []string{"be", "le", "host"} {
			name := fmt.Sprintf("%s (%s)", tt.name, endianess)
			t.Run(name, func(t *testing.T) {
				parser := &Parser{
					Endianess:  endianess,
					Configs:    tt.configs,
					Log:        testutil.Logger{Name: "parsers.binary"},
					metricName: "binary",
				}
				require.NoError(t, parser.Init())

				order := determineEndianess(endianess)
				data, err := generateBinary(tt.data, order)
				require.NoError(t, err)

				_, err = parser.ParseLine(string(data))
				require.EqualError(t, err, tt.expected)
			})
		}
	}
}

func TestParseLine(t *testing.T) {
	var tests = []struct {
		name     string
		data     []interface{}
		filter   *Filter
		entries  []Entry
		expected telegraf.Metric
	}{
		{
			name: "no match",
			data: []interface{}{
				"2022-07-25T20:41:29+02:00", // time
				uint16(0x0102),              // address
				float64(42.123),             // value
			},
			filter: &Filter{Length: 4},
			entries: []Entry{
				{
					Type:       "2006-01-02T15:04:05Z07:00",
					Assignment: "time",
					Timezone:   "UTC",
				},
				{
					Type: "uint16",
					Omit: true,
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
		},
		{
			name: "single match",
			data: []interface{}{
				"2022-07-25T20:41:29+02:00", // time
				uint16(0x0102),              // address
				float64(42.123),             // value
			},
			entries: []Entry{
				{
					Type:       "2006-01-02T15:04:05Z07:00",
					Assignment: "time",
					Timezone:   "UTC",
				},
				{
					Type: "uint16",
					Omit: true,
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: metric.New(
				"binary",
				map[string]string{},
				map[string]interface{}{"value": float64(42.123)},
				time.Unix(1658774489, 0),
			),
		},
	}

	for _, tt := range tests {
		for _, endianess := range []string{"be", "le", "host"} {
			name := fmt.Sprintf("%s (%s)", tt.name, endianess)
			t.Run(name, func(t *testing.T) {
				parser := &Parser{
					AllowNoMatch: true,
					Endianess:    endianess,
					Configs: []Config{{
						Filter:  tt.filter,
						Entries: tt.entries,
					}},
					Log:        testutil.Logger{Name: "parsers.binary"},
					metricName: "binary",
				}
				require.NoError(t, parser.Init())

				order := determineEndianess(endianess)
				data, err := generateBinary(tt.data, order)
				require.NoError(t, err)

				m, err := parser.ParseLine(string(data))
				require.NoError(t, err)

				testutil.RequireMetricEqual(t, tt.expected, m)
			})
		}
	}
}

func TestParseInvalid(t *testing.T) {
	var tests = []struct {
		name     string
		data     []interface{}
		entries  []Entry
		expected string
	}{
		{
			name: "message too short",
			data: []interface{}{uint64(0x0102030405060708)},
			entries: []Entry{
				{
					Name:       "command",
					Type:       "uint32",
					Assignment: "tag",
				},
				{
					Name:       "version",
					Type:       "uint32",
					Assignment: "tag",
				},
				{
					Name:       "address",
					Type:       "uint32",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: `out-of-bounds @64 with 32 bits`,
		},
		{
			name: "non-terminated string",
			data: []interface{}{
				uint16(0xAB42),       // address
				"testmetric",         // metric
				float64(42.23432243), // value
			},
			entries: []Entry{
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Type:       "string",
					Terminator: "null",
					Assignment: "measurement",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: `terminator not found for "measurement"`,
		},
		{
			name: "invalid time",
			data: []interface{}{
				"2022-07-25T18:41:XYZ", // time
				uint16(0x0102),         // address
				float64(42.123),        // value
			},
			entries: []Entry{
				{
					Type:       "2006-01-02T15:04:05Z",
					Assignment: "time",
				},
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: `time failed: parsing time "2022-07-25T18:41:XYZ" as "2006-01-02T15:04:05Z": cannot parse "XYZ" as "05"`,
		},
	}

	for _, tt := range tests {
		for _, endianess := range []string{"be", "le", "host"} {
			name := fmt.Sprintf("%s (%s)", tt.name, endianess)
			t.Run(name, func(t *testing.T) {
				parser := &Parser{
					Endianess:  endianess,
					Configs:    []Config{{Entries: tt.entries}},
					Log:        testutil.Logger{Name: "parsers.binary"},
					metricName: "binary",
				}
				require.NoError(t, parser.Init())

				order := determineEndianess(endianess)
				data, err := generateBinary(tt.data, order)
				require.NoError(t, err)

				_, err = parser.Parse(data)
				require.EqualError(t, err, tt.expected)
			})
		}
	}
}

func TestParse(t *testing.T) {
	timeBerlin, err := time.Parse(time.RFC3339, "2022-07-25T20:41:29+02:00")
	require.NoError(t, err)
	timeBerlinMilli, err := time.Parse(time.RFC3339Nano, "2022-07-25T20:41:29.123+02:00")
	require.NoError(t, err)

	var tests = []struct {
		name       string
		data       []interface{}
		entries    []Entry
		ignoreTime bool
		expected   []telegraf.Metric
	}{
		{
			name: "fixed numbers",
			data: []interface{}{
				uint16(0xAB42),             // command
				uint8(0x02),                // version
				uint32(0x010000FF),         // address
				uint64(0x0102030405060708), // serial-number
				int8(-25),                  // countdown as int32
				int16(-42),                 // overdue
				int32(-65535),              // batchleft
				int64(12345678),            // counter
				float32(3.1415),            // x
				float32(99.471),            // y
				float64(0.23432243),        // z
				uint8(0xFF),                // status
				uint8(0x0F),                // on/off bit-field
			},
			entries: []Entry{
				{
					Name:       "command",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name:       "version",
					Type:       "uint8",
					Assignment: "tag",
				},
				{
					Name:       "address",
					Type:       "uint32",
					Assignment: "tag",
				},
				{
					Name:       "serialnumber",
					Type:       "uint64",
					Assignment: "tag",
				},
				{
					Name: "countdown",
					Type: "int8",
				},
				{
					Name: "overdue",
					Type: "int16",
				},
				{
					Name: "batchleft",
					Type: "int32",
				},
				{
					Name: "counter",
					Type: "int64",
				},
				{
					Name: "x",
					Type: "float32",
				},
				{
					Name: "y",
					Type: "float32",
				},
				{
					Name: "z",
					Type: "float64",
				},
				{
					Name: "status",
					Type: "bool",
					Bits: 8,
				},
				{
					Name:       "error_part",
					Type:       "bool",
					Bits:       4,
					Assignment: "tag",
				},
				{
					Name:       "ok_part1",
					Type:       "bool",
					Bits:       1,
					Assignment: "tag",
				},
				{
					Name:       "ok_part2",
					Type:       "bool",
					Bits:       1,
					Assignment: "tag",
				},
				{
					Name:       "ok_part3",
					Type:       "bool",
					Bits:       1,
					Assignment: "tag",
				},
				{
					Name:       "ok_part4",
					Type:       "bool",
					Bits:       1,
					Assignment: "tag",
				},
			},
			ignoreTime: true,
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{
						"command":      "43842",
						"version":      "2",
						"address":      "16777471",
						"serialnumber": "72623859790382856",
						"error_part":   "false",
						"ok_part1":     "true",
						"ok_part2":     "true",
						"ok_part3":     "true",
						"ok_part4":     "true",
					},
					map[string]interface{}{
						"x":         float32(3.1415),
						"y":         float32(99.471),
						"z":         float64(0.23432243),
						"countdown": int8(-25),
						"overdue":   int16(-42),
						"batchleft": int32(-65535),
						"counter":   int64(12345678),

						"status": true,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "fixed length string",
			data: []interface{}{
				uint16(0xAB42),      // address
				"test",              // metric
				float64(0.23432243), // value
			},
			entries: []Entry{
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name:       "app",
					Type:       "string",
					Bits:       4 * 8,
					Assignment: "field",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			ignoreTime: true,
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{"address": "43842"},
					map[string]interface{}{
						"app":   "test",
						"value": float64(0.23432243),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "null-terminated string",
			data: []interface{}{
				uint16(0xAB42),                     // address
				append([]byte("testmetric"), 0x00), // metric
				float64(42.23432243),               // value
			},
			entries: []Entry{
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Type:       "string",
					Terminator: "null",
					Assignment: "measurement",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			ignoreTime: true,
			expected: []telegraf.Metric{
				metric.New(
					"testmetric",
					map[string]string{"address": "43842"},
					map[string]interface{}{"value": float64(42.23432243)},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "char-terminated string",
			data: []interface{}{
				uint16(0xAB42),                           // address
				append([]byte("testmetric"), 0x0A, 0x0B), // metric
				float64(42.23432243),                     // value
			},
			entries: []Entry{
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Type:       "string",
					Terminator: "0x0A0B",
					Assignment: "measurement",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			ignoreTime: true,
			expected: []telegraf.Metric{
				metric.New(
					"testmetric",
					map[string]string{"address": "43842"},
					map[string]interface{}{"value": float64(42.23432243)},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "time (unix/UTC)",
			data: []interface{}{
				uint64(1658774489), // time
				uint16(0x0102),     // address
				float64(42.123),    // value
			},
			entries: []Entry{
				{
					Type:       "unix",
					Assignment: "time",
				},
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{"address": "258"},
					map[string]interface{}{"value": float64(42.123)},
					time.Unix(1658774489, 0),
				),
			},
		},
		{
			name: "time (unix/Berlin)",
			data: []interface{}{
				uint64(1658774489), // time
				uint16(0x0102),     // address
				float64(42.123),    // value
			},
			entries: []Entry{
				{
					Type:       "unix",
					Assignment: "time",
					Timezone:   "Europe/Berlin",
				},
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{"address": "258"},
					map[string]interface{}{"value": float64(42.123)},
					timeBerlin,
				),
			},
		},
		{
			name: "time (unix_ms/UTC)",
			data: []interface{}{
				uint64(1658774489123), // time
				uint16(0x0102),        // address
				float64(42.123),       // value
			},
			entries: []Entry{
				{
					Type:       "unix_ms",
					Assignment: "time",
				},
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{"address": "258"},
					map[string]interface{}{"value": float64(42.123)},
					time.Unix(0, 1658774489123*1_000_000),
				),
			},
		},
		{
			name: "time (unix_ms/Berlin)",
			data: []interface{}{
				uint64(1658774489123), // time
				uint16(0x0102),        // address
				float64(42.123),       // value
			},
			entries: []Entry{
				{
					Type:       "unix_ms",
					Assignment: "time",
					Timezone:   "Europe/Berlin",
				},
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{"address": "258"},
					map[string]interface{}{"value": float64(42.123)},
					timeBerlinMilli,
				),
			},
		},
		{
			name: "time (RFC3339/UTC)",
			data: []interface{}{
				"2022-07-25T18:41:29Z", // time
				uint16(0x0102),         // address
				float64(42.123),        // value
			},
			entries: []Entry{
				{
					Type:       "2006-01-02T15:04:05Z",
					Assignment: "time",
				},
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{"address": "258"},
					map[string]interface{}{"value": float64(42.123)},
					time.Unix(1658774489, 0),
				),
			},
		},
		{
			name: "time (RFC3339/Berlin)",
			data: []interface{}{
				"2022-07-25T20:41:29+02:00", // time
				uint16(0x0102),              // address
				float64(42.123),             // value
			},
			entries: []Entry{
				{
					Type:       "2006-01-02T15:04:05Z07:00",
					Assignment: "time",
					Timezone:   "Europe/Berlin",
				},
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{"address": "258"},
					map[string]interface{}{"value": float64(42.123)},
					timeBerlin,
				),
			},
		},
		{
			name: "time (RFC3339/Berlin->UTC)",
			data: []interface{}{
				"2022-07-25T20:41:29+02:00", // time
				uint16(0x0102),              // address
				float64(42.123),             // value
			},
			entries: []Entry{
				{
					Type:       "2006-01-02T15:04:05Z07:00",
					Assignment: "time",
					Timezone:   "UTC",
				},
				{
					Name:       "address",
					Type:       "uint16",
					Assignment: "tag",
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{"address": "258"},
					map[string]interface{}{"value": float64(42.123)},
					time.Unix(1658774489, 0),
				),
			},
		},
		{
			name: "omit",
			data: []interface{}{
				"2022-07-25T20:41:29+02:00", // time
				uint16(0x0102),              // address
				float64(42.123),             // value
			},
			entries: []Entry{
				{
					Type:       "2006-01-02T15:04:05Z07:00",
					Assignment: "time",
					Timezone:   "UTC",
				},
				{
					Type: "uint16",
					Omit: true,
				},
				{
					Name: "value",
					Type: "float64",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"binary",
					map[string]string{},
					map[string]interface{}{"value": float64(42.123)},
					time.Unix(1658774489, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		for _, endianess := range []string{"be", "le", "host"} {
			name := fmt.Sprintf("%s (%s)", tt.name, endianess)
			t.Run(name, func(t *testing.T) {
				parser := &Parser{
					Endianess:  endianess,
					Configs:    []Config{{Entries: tt.entries}},
					Log:        testutil.Logger{Name: "parsers.binary"},
					metricName: "binary",
				}
				require.NoError(t, parser.Init())

				order := determineEndianess(endianess)
				data, err := generateBinary(tt.data, order)
				require.NoError(t, err)

				metrics, err := parser.Parse(data)
				require.NoError(t, err)

				var options []cmp.Option
				if tt.ignoreTime {
					options = append(options, testutil.IgnoreTime())
				}
				testutil.RequireMetricsEqual(t, tt.expected, metrics, options...)
			})
		}
	}
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := ioutil.ReadDir("testdata")
	require.NoError(t, err)
	// Make sure testdata contains data
	require.Greater(t, len(folders), 0)

	for _, f := range folders {
		t.Run(f.Name(), func(t *testing.T) {
			configFilename := filepath.Join("testdata", f.Name(), "telegraf.conf")
			expectedFilename := filepath.Join("testdata", f.Name(), "expected.out")

			// Process the telegraf config file for the test
			buf, err := os.ReadFile(configFilename)
			require.NoError(t, err)
			inputs.Add("file", func() telegraf.Input {
				return &file.File{}
			})
			cfg := config.NewConfig()
			err = cfg.LoadConfigData(buf)
			require.NoError(t, err)

			// Gather the metrics from the input file configure
			acc := testutil.Accumulator{}
			for _, input := range cfg.Inputs {
				require.NoError(t, input.Init())
				require.NoError(t, input.Gather(&acc))
			}

			// Process expected metrics and compare with resulting metrics
			expected, err := readMetricFile(expectedFilename)
			require.NoError(t, err)
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual)
		})
	}
}

func readMetricFile(path string) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric

	expectedFile, err := os.Open(path)
	if err != nil {
		return metrics, err
	}
	defer expectedFile.Close()

	parser := &influx.Parser{}
	if err := parser.Init(); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(expectedFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			m, err := parser.ParseLine(line)
			if err != nil {
				return nil, fmt.Errorf("unable to parse metric in %q failed: %v", line, err)
			}
			// The timezone needs to be UTC to match the timestamp test results
			m.SetTime(m.Time().UTC())
			metrics = append(metrics, m)
		}
	}

	return metrics, nil
}
