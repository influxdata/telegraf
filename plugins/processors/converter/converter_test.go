package converter

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func Metric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

func TestConverter(t *testing.T) {
	tests := []struct {
		name      string
		converter *Converter
		input     telegraf.Metric
		expected  telegraf.Metric
	}{
		{
			name:      "empty",
			converter: &Converter{},
			input: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			),
		},
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
			input: Metric(
				metric.New(
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
			),
			expected: Metric(
				metric.New(
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
			),
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
			input: Metric(
				metric.New(
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
			),
			expected: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "from string field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a"},
					Integer:  []string{"b"},
					Unsigned: []string{"c"},
					Boolean:  []string{"d"},
					Float:    []string{"e"},
					Tag:      []string{"f"},
				},
			},
			input: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"a": "howdy",
						"b": "42",
						"c": "42",
						"d": "true",
						"e": "42.0",
						"f": "foo",
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"cpu",
					map[string]string{
						"f": "foo",
					},
					map[string]interface{}{
						"a": "howdy",
						"b": int64(42),
						"c": uint64(42),
						"d": true,
						"e": 42.0,
					},
					time.Unix(0, 0),
				),
			),
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
			input: Metric(
				metric.New(
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
			),
			expected: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "from integer field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a"},
					Integer:  []string{"b"},
					Unsigned: []string{"c", "negative_uint"},
					Boolean:  []string{"d"},
					Float:    []string{"e"},
					Tag:      []string{"f"},
				},
			},
			input: Metric(
				metric.New(
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
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
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
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "from unsigned field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a"},
					Integer:  []string{"b", "overflow_int"},
					Unsigned: []string{"c"},
					Boolean:  []string{"d"},
					Float:    []string{"e"},
					Tag:      []string{"f"},
				},
			},
			input: Metric(
				metric.New(
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
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
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
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "out of range for unsigned",
			converter: &Converter{
				Fields: &Conversion{
					Unsigned: []string{"a", "b"},
				},
			},
			input: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"a": int64(-42),
						"b": math.MaxFloat64,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"a": uint64(0),
						"b": uint64(math.MaxUint64),
					},
					time.Unix(0, 0),
				),
			),
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
			input: Metric(
				metric.New(
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
			),
			expected: Metric(
				metric.New(
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
			),
		},
		{
			name: "from float field",
			converter: &Converter{
				Fields: &Conversion{
					String:   []string{"a"},
					Integer:  []string{"b", "too_large_int", "too_small_int"},
					Unsigned: []string{"c", "negative_uint", "too_large_uint", "too_small_uint"},
					Boolean:  []string{"d"},
					Float:    []string{"e"},
					Tag:      []string{"f"},
				},
			},
			input: Metric(
				metric.New(
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
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
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
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "globbing",
			converter: &Converter{
				Fields: &Conversion{
					Integer: []string{"int_*"},
				},
			},
			input: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"int_a":   "1",
						"int_b":   "2",
						"float_a": 1.0,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"int_a":   int64(1),
						"int_b":   int64(2),
						"float_a": 1.0,
					},
					time.Unix(0, 0),
				),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.converter.Apply(tt.input)

			require.Equal(t, 1, len(metrics))
			require.Equal(t, tt.expected.Name(), metrics[0].Name())
			require.Equal(t, tt.expected.Tags(), metrics[0].Tags())
			require.Equal(t, tt.expected.Fields(), metrics[0].Fields())
			require.Equal(t, tt.expected.Time(), metrics[0].Time())
		})
	}
}
