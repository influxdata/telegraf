package converter

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConverter(t *testing.T) {
	tests := []struct {
		name      string
		converter *Converter
		input     telegraf.Metric
		expected  []telegraf.Metric
	}{
		{
			name: "from tag",
			converter: &Converter{
				Tags: &Conversion{
					String:   []string{"string"},
					Integer:  []string{"int"},
					Unsigned: []string{"uint"},
					Boolean:  []string{"bool"},
					Float:    []string{"float"},
					Tag:      []string{"tag"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{
					"float":  "42",
					"int":    "42",
					"uint":   "42",
					"bool":   "true",
					"string": "howdy",
					"tag":    "tag",
				},
				map[string]interface{}{},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"tag": "tag",
					},
					map[string]interface{}{
						"float":  42.0,
						"int":    int64(42),
						"uint":   uint64(42),
						"bool":   true,
						"string": "howdy",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "from tag unconvertible",
			converter: &Converter{
				Tags: &Conversion{
					Integer:  []string{"int"},
					Unsigned: []string{"uint"},
					Boolean:  []string{"bool"},
					Float:    []string{"float"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{
					"float": "a",
					"int":   "b",
					"uint":  "c",
					"bool":  "maybe",
				},
				map[string]interface{}{},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "from string field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a"},
					Integer:  []string{"b", "b1", "b2", "b3"},
					Unsigned: []string{"c", "c1", "c2", "c3"},
					Boolean:  []string{"d"},
					Float:    []string{"e", "g"},
					Tag:      []string{"f"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"a":  "howdy",
					"b":  "42",
					"b1": "42.2",
					"b2": "42.5",
					"b3": "0x2A",
					"c":  "42",
					"c1": "42.2",
					"c2": "42.5",
					"c3": "0x2A",
					"d":  "true",
					"e":  "42.0",
					"f":  "foo",
					"g":  "foo",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"f": "foo",
					},
					map[string]interface{}{
						"a":  "howdy",
						"b":  int64(42),
						"b1": int64(42),
						"b2": int64(43),
						"b3": int64(42),
						"c":  uint64(42),
						"c1": uint64(42),
						"c2": uint64(43),
						"c3": uint64(42),
						"d":  true,
						"e":  42.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "from string field unconvertible",
			converter: &Converter{
				Fields: &Conversion{
					Integer:  []string{"a"},
					Unsigned: []string{"b"},
					Boolean:  []string{"c"},
					Float:    []string{"d"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"a": "a",
					"b": "b",
					"c": "c",
					"d": "d",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "from integer field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a"},
					Integer:  []string{"b"},
					Unsigned: []string{"c", "negative_uint"},
					Boolean:  []string{"d", "bool_zero"},
					Float:    []string{"e"},
					Tag:      []string{"f"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"a":             int64(42),
					"b":             int64(42),
					"c":             int64(42),
					"d":             int64(42),
					"e":             int64(42),
					"f":             int64(42),
					"negative_uint": int64(-42),
					"bool_zero":     int64(0),
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"f": "42",
					},
					map[string]interface{}{
						"a":             "42",
						"b":             int64(42),
						"c":             uint64(42),
						"d":             true,
						"e":             42.0,
						"negative_uint": uint64(0),
						"bool_zero":     false,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "from unsigned field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a"},
					Integer:  []string{"b", "overflow_int"},
					Unsigned: []string{"c"},
					Boolean:  []string{"d", "bool_zero"},
					Float:    []string{"e"},
					Tag:      []string{"f"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"a":            uint64(42),
					"b":            uint64(42),
					"c":            uint64(42),
					"d":            uint64(42),
					"e":            uint64(42),
					"f":            uint64(42),
					"overflow_int": uint64(math.MaxUint64),
					"bool_zero":    uint64(0),
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"f": "42",
					},
					map[string]interface{}{
						"a":            "42",
						"b":            int64(42),
						"c":            uint64(42),
						"d":            true,
						"e":            42.0,
						"overflow_int": int64(math.MaxInt64),
						"bool_zero":    false,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "out of range for unsigned",
			converter: &Converter{
				Fields: &Conversion{
					Unsigned: []string{"a", "b"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"a": int64(-42),
					"b": math.MaxFloat64,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"a": uint64(0),
						"b": uint64(math.MaxUint64),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "boolean field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a", "af"},
					Integer:  []string{"b", "bf"},
					Unsigned: []string{"c", "cf"},
					Boolean:  []string{"d", "df"},
					Float:    []string{"e", "ef"},
					Tag:      []string{"f", "ff"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"a":  true,
					"b":  true,
					"c":  true,
					"d":  true,
					"e":  true,
					"f":  true,
					"af": false,
					"bf": false,
					"cf": false,
					"df": false,
					"ef": false,
					"ff": false,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"f":  "true",
						"ff": "false",
					},
					map[string]interface{}{
						"a":  "true",
						"af": "false",
						"b":  int64(1),
						"bf": int64(0),
						"c":  uint64(1),
						"cf": uint64(0),
						"d":  true,
						"df": false,
						"e":  1.0,
						"ef": 0.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "from float field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a"},
					Integer:  []string{"b", "too_large_int", "too_small_int"},
					Unsigned: []string{"c", "negative_uint", "too_large_uint", "too_small_uint"},
					Boolean:  []string{"d", "bool_zero"},
					Float:    []string{"e"},
					Tag:      []string{"f"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"a":              42.0,
					"b":              42.0,
					"c":              42.0,
					"d":              42.0,
					"e":              42.0,
					"f":              42.0,
					"too_large_int":  math.MaxFloat64,
					"too_large_uint": math.MaxFloat64,
					"too_small_int":  -math.MaxFloat64,
					"too_small_uint": -math.MaxFloat64,
					"negative_uint":  -42.0,
					"bool_zero":      0.0,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"f": "42",
					},
					map[string]interface{}{
						"a":              "42",
						"b":              int64(42),
						"c":              uint64(42),
						"d":              true,
						"e":              42.0,
						"too_large_int":  int64(math.MaxInt64),
						"too_large_uint": uint64(math.MaxUint64),
						"too_small_int":  int64(math.MinInt64),
						"too_small_uint": uint64(0),
						"negative_uint":  uint64(0),
						"bool_zero":      false,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "globbing",
			converter: &Converter{
				Fields: &Conversion{
					Integer: []string{"int_*"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"int_a":   "1",
					"int_b":   "2",
					"float_a": 1.0,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"int_a":   int64(1),
						"int_b":   int64(2),
						"float_a": 1.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "from string field hexidecimal",
			converter: &Converter{
				Fields: &Conversion{
					Integer:  []string{"a"},
					Unsigned: []string{"b"},
					Float:    []string{"c"},
				},
			},
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"a": "0x11826c",
					"b": "0x11826c",
					"c": "0x2139d19bb1c580ebe0",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"a": int64(1147500),
						"b": uint64(1147500),
						"c": float64(612908836750534700000),
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.converter.Log = testutil.Logger{}

			err := tt.converter.Init()
			require.NoError(t, err)
			actual := tt.converter.Apply(tt.input)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestMeasurement(t *testing.T) {
	tests := []struct {
		name      string
		converter *Converter
		input     telegraf.Metric
		expected  []telegraf.Metric
	}{
		{
			name: "measurement from tag",
			converter: &Converter{
				Tags: &Conversion{
					Measurement: []string{"filepath"},
				},
			},
			input: testutil.MustMetric(
				"file",
				map[string]string{
					"filepath": "/var/log/syslog",
				},
				map[string]interface{}{
					"msg": "howdy",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"/var/log/syslog",
					map[string]string{},
					map[string]interface{}{
						"msg": "howdy",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "measurement from field",
			converter: &Converter{
				Fields: &Conversion{
					Measurement: []string{"topic"},
				},
			},
			input: testutil.MustMetric(
				"file",
				map[string]string{},
				map[string]interface{}{
					"v":     1,
					"topic": "telegraf",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"telegraf",
					map[string]string{},
					map[string]interface{}{
						"v": 1,
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.converter.Log = testutil.Logger{}
			err := tt.converter.Init()
			require.NoError(t, err)

			actual := tt.converter.Apply(tt.input)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestEmptyConfigInitError(t *testing.T) {
	converter := &Converter{
		Log: testutil.Logger{},
	}
	err := converter.Init()
	require.Error(t, err)
}
