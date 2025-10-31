package xpath

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

const invalidXML = `
<?xml version="1.0"?>
	<Device_1>This one has to fail due to missing end-tag
`

const singleMetricValuesXML = `
<?xml version="1.0"?>
	<Device_1>
		<Name>Device TestDevice1</Name>
		<State>ok</State>
		<Timestamp_unix>1577923199</Timestamp_unix>
		<Timestamp_unix_ms>1577923199128</Timestamp_unix_ms>
		<Timestamp_unix_us>1577923199128256</Timestamp_unix_us>
		<Timestamp_unix_ns>1577923199128256512</Timestamp_unix_ns>
		<Timestamp_iso>2020-01-01T23:59:59Z</Timestamp_iso>
		<value_int>98247</value_int>
		<value_float>98695.81</value_float>
		<value_bool>true</value_bool>
		<value_string>this is a test</value_string>
		<value_position>42;23</value_position>
	</Device_1>
`
const singleMetricAttributesXML = `
<?xml version="1.0"?>
	<Device_1>
		<Name value="Device TestDevice1"/>
		<State _="ok"/>
		<Timestamp_unix value="1577923199"/>
		<Timestamp_iso value="2020-01-01T23:59:59Z"/>
		<attr_int _="12345"/>
		<attr_float _="12345.678"/>
		<attr_bool _="true"/>
		<attr_bool_numeric _="1"/>
		<attr_string _="this is a test"/>
	</Device_1>
`
const singleMetricMultiValuesXML = `
<?xml version="1.0"?>
	<Timestamp value="1577923199"/>
	<Device>
		<Value>1</Value>
		<Value>2</Value>
		<Value>3</Value>
		<Value>4</Value>
		<Value>5</Value>
		<Value>6</Value>
	</Device>
`
const multipleNodesXML = `
<?xml version="1.0"?>
	<Timestamp value="1577923199"/>
	<Device name="Device 1">
		<Value mode="0">42.0</Value>
		<Active>1</Active>
		<State>ok</State>
	</Device>
	<Device name="Device 2">
		<Value mode="1">42.1</Value>
		<Active>0</Active>
		<State>ok</State>
	</Device>
	<Device name="Device 3">
		<Value mode="2">42.2</Value>
		<Active>1</Active>
		<State>ok</State>
	</Device>
	<Device name="Device 4">
		<Value mode="3">42.3</Value>
		<Active>0</Active>
		<State>failed</State>
	</Device>
	<Device name="Device 5">
		<Value mode="4">42.4</Value>
		<Active>1</Active>
		<State>failed</State>
	</Device>
`

const metricNameQueryXML = `
<?xml version="1.0"?>
	<Device_1>
		<Timestamp_unix>1577923199</Timestamp_unix>
		<Metric state="ok"/>
	</Device_1>
`

func TestParseInvalidXML(t *testing.T) {
	var tests = []struct {
		name          string
		input         string
		configs       []Config
		defaultTags   map[string]string
		expectedError string
	}{
		{
			name:  "invalid XML (missing close tag)",
			input: invalidXML,
			configs: []Config{
				{
					MetricQuery: "test",
					Timestamp:   "/Device_1/Timestamp_unix",
				},
			},
			defaultTags:   map[string]string{},
			expectedError: "XML syntax error on line 4: unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "xml",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			_, err := parser.ParseLine(tt.input)
			require.Error(t, err)
			require.Equal(t, tt.expectedError, err.Error())
		})
	}
}

func TestInvalidTypeQueriesFail(t *testing.T) {
	var tests = []struct {
		name          string
		input         string
		configs       []Config
		defaultTags   map[string]string
		expectedError string
	}{
		{
			name:  "invalid field (int) type",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					FieldsInt: map[string]string{
						"a": "/Device_1/value_string",
					},
				},
			},
			defaultTags:   map[string]string{},
			expectedError: `failed to parse field (int) "a": strconv.ParseInt: parsing "this is a test": invalid syntax`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "xml",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			_, err := parser.ParseLine(tt.input)
			require.Error(t, err)
			require.Equal(t, tt.expectedError, err.Error())
		})
	}
}

func TestInvalidTypeQueries(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		configs     []Config
		defaultTags map[string]string
		expected    telegraf.Metric
	}{
		{
			name:  "invalid field type (number)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"a": "number(/Device_1/value_string)",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": math.NaN(),
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "invalid field type (boolean)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"a": "boolean(/Device_1/value_string)",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": true,
				},
				time.Unix(1577923199, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			actual, err := parser.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual)
		})
	}
}

func TestParseTimestamps(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		configs     []Config
		defaultTags map[string]string
		expected    telegraf.Metric
	}{
		{
			name:  "parse timestamp (no fmt)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse timestamp (unix)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp:    "/Device_1/Timestamp_unix",
					TimestampFmt: "unix",
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse timestamp (unix_ms)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp:    "/Device_1/Timestamp_unix_ms",
					TimestampFmt: "unix_ms",
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(0, int64(1577923199128*1e6)),
			),
		},
		{
			name:  "parse timestamp (unix_us)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp:    "/Device_1/Timestamp_unix_us",
					TimestampFmt: "unix_us",
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(0, int64(1577923199128256*1e3)),
			),
		},
		{
			name:  "parse timestamp (unix_us)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp:    "/Device_1/Timestamp_unix_ns",
					TimestampFmt: "unix_ns",
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(0, int64(1577923199128256512)),
			),
		},
		{
			name:  "parse timestamp (RFC3339)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp:    "/Device_1/Timestamp_iso",
					TimestampFmt: "2006-01-02T15:04:05Z",
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(1577923199, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			actual, err := parser.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual)
		})
	}
}

func TestParseSingleValues(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		configs     []Config
		defaultTags map[string]string
		expected    telegraf.Metric
	}{
		{
			name:  "parse scalar values as string fields",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"a": "/Device_1/value_int",
						"b": "/Device_1/value_float",
						"c": "/Device_1/value_bool",
						"d": "/Device_1/value_string",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": "98247",
					"b": "98695.81",
					"c": "true",
					"d": "this is a test",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse scalar values as typed fields (w/o int)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"a": "number(Device_1/value_int)",
						"b": "number(/Device_1/value_float)",
						"c": "boolean(/Device_1/value_bool)",
						"d": "/Device_1/value_string",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 98247.0,
					"b": 98695.81,
					"c": true,
					"d": "this is a test",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse values as typed fields (w/ int)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"b": "number(/Device_1/value_float)",
						"c": "boolean(/Device_1/value_bool)",
						"d": "/Device_1/value_string",
					},
					FieldsInt: map[string]string{
						"a": "/Device_1/value_int",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 98247,
					"b": 98695.81,
					"c": true,
					"d": "this is a test",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse substring values",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"x": "substring-before(/Device_1/value_position, ';')",
						"y": "substring-after(/Device_1/value_position, ';')",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"x": "42",
					"y": "23",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse substring values (typed)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"x": "number(substring-before(/Device_1/value_position, ';'))",
						"y": "number(substring-after(/Device_1/value_position, ';'))",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"x": 42.0,
					"y": 23.0,
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse substring values (typed int)",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					FieldsInt: map[string]string{
						"x": "substring-before(/Device_1/value_position, ';')",
						"y": "substring-after(/Device_1/value_position, ';')",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"x": 42,
					"y": 23,
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse tags",
			input: singleMetricValuesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix",
					Tags: map[string]string{
						"state": "/Device_1/State",
						"name":  "substring-after(/Device_1/Name, ' ')",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{
					"state": "ok",
					"name":  "TestDevice1",
				},
				map[string]interface{}{},
				time.Unix(1577923199, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			actual, err := parser.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual)
		})
	}
}

func TestParseSingleAttributes(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		configs     []Config
		defaultTags map[string]string
		expected    telegraf.Metric
	}{
		{
			name:  "parse attr timestamp (unix)",
			input: singleMetricAttributesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix/@value",
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse attr timestamp (RFC3339)",
			input: singleMetricAttributesXML,
			configs: []Config{
				{
					Timestamp:    "/Device_1/Timestamp_iso/@value",
					TimestampFmt: "2006-01-02T15:04:05Z",
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse attr as string fields",
			input: singleMetricAttributesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix/@value",
					Fields: map[string]string{
						"a": "/Device_1/attr_int/@_",
						"b": "/Device_1/attr_float/@_",
						"c": "/Device_1/attr_bool/@_",
						"d": "/Device_1/attr_string/@_",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": "12345",
					"b": "12345.678",
					"c": "true",
					"d": "this is a test",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse attr as typed fields (w/o int)",
			input: singleMetricAttributesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix/@value",
					Fields: map[string]string{
						"a": "number(/Device_1/attr_int/@_)",
						"b": "number(/Device_1/attr_float/@_)",
						"c": "boolean(/Device_1/attr_bool/@_)",
						"d": "/Device_1/attr_string/@_",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 12345.0,
					"b": 12345.678,
					"c": true,
					"d": "this is a test",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse attr as typed fields (w/ int)",
			input: singleMetricAttributesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix/@value",
					Fields: map[string]string{
						"b": "number(/Device_1/attr_float/@_)",
						"c": "boolean(/Device_1/attr_bool/@_)",
						"d": "/Device_1/attr_string/@_",
					},
					FieldsInt: map[string]string{
						"a": "/Device_1/attr_int/@_",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 12345,
					"b": 12345.678,
					"c": true,
					"d": "this is a test",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse attr substring",
			input: singleMetricAttributesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix/@value",
					Fields: map[string]string{
						"name": "substring-after(/Device_1/Name/@value, ' ')",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"name": "TestDevice1",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse attr tags",
			input: singleMetricAttributesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix/@value",
					Tags: map[string]string{
						"state": "/Device_1/State/@_",
						"name":  "substring-after(/Device_1/Name/@value, ' ')",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{
					"state": "ok",
					"name":  "TestDevice1",
				},
				map[string]interface{}{},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse attr bool",
			input: singleMetricAttributesXML,
			configs: []Config{
				{
					Timestamp: "/Device_1/Timestamp_unix/@value",
					Fields: map[string]string{
						"a": "/Device_1/attr_bool_numeric/@_ = 1",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": true,
				},
				time.Unix(1577923199, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			actual, err := parser.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual)
		})
	}
}

func TestParseMultiValues(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		configs     []Config
		defaultTags map[string]string
		expected    telegraf.Metric
	}{
		{
			name:  "select values (float)",
			input: singleMetricMultiValuesXML,
			configs: []Config{
				{
					Timestamp: "/Timestamp/@value",
					Fields: map[string]string{
						"a": "number(/Device/Value[1])",
						"b": "number(/Device/Value[2])",
						"c": "number(/Device/Value[3])",
						"d": "number(/Device/Value[4])",
						"e": "number(/Device/Value[5])",
						"f": "number(/Device/Value[6])",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1.0,
					"b": 2.0,
					"c": 3.0,
					"d": 4.0,
					"e": 5.0,
					"f": 6.0,
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "select values (int)",
			input: singleMetricMultiValuesXML,
			configs: []Config{
				{
					Timestamp: "/Timestamp/@value",
					FieldsInt: map[string]string{
						"a": "/Device/Value[1]",
						"b": "/Device/Value[2]",
						"c": "/Device/Value[3]",
						"d": "/Device/Value[4]",
						"e": "/Device/Value[5]",
						"f": "/Device/Value[6]",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"b": 2,
					"c": 3,
					"d": 4,
					"e": 5,
					"f": 6,
				},
				time.Unix(1577923199, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			actual, err := parser.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual)
		})
	}
}

func TestParseMultiNodes(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		configs     []Config
		defaultTags map[string]string
		expected    []telegraf.Metric
	}{
		{
			name:  "select all devices",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "/Device",
					Timestamp: "/Timestamp/@value",
					Fields: map[string]string{
						"value":  "number(Value)",
						"active": "Active = 1",
					},
					FieldsInt: map[string]string{
						"mode": "Value/@mode",
					},
					Tags: map[string]string{
						"name":  "@name",
						"state": "State",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test",
					map[string]string{
						"name":  "Device 1",
						"state": "ok",
					},
					map[string]interface{}{
						"value":  42.0,
						"active": true,
						"mode":   0,
					},
					time.Unix(1577923199, 0),
				),
				testutil.MustMetric(
					"test",
					map[string]string{
						"name":  "Device 2",
						"state": "ok",
					},
					map[string]interface{}{
						"value":  42.1,
						"active": false,
						"mode":   1,
					},
					time.Unix(1577923199, 0),
				),
				testutil.MustMetric(
					"test",
					map[string]string{
						"name":  "Device 3",
						"state": "ok",
					},
					map[string]interface{}{
						"value":  42.2,
						"active": true,
						"mode":   2,
					},
					time.Unix(1577923199, 0),
				),
				testutil.MustMetric(
					"test",
					map[string]string{
						"name":  "Device 4",
						"state": "failed",
					},
					map[string]interface{}{
						"value":  42.3,
						"active": false,
						"mode":   3,
					},
					time.Unix(1577923199, 0),
				),
				testutil.MustMetric(
					"test",
					map[string]string{
						"name":  "Device 5",
						"state": "failed",
					},
					map[string]interface{}{
						"value":  42.4,
						"active": true,
						"mode":   4,
					},
					time.Unix(1577923199, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			actual, err := parser.Parse([]byte(tt.input))
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestParseMetricQuery(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		configs     []Config
		defaultTags map[string]string
		expected    telegraf.Metric
	}{
		{
			name:  "parse metric name query",
			input: metricNameQueryXML,
			configs: []Config{
				{
					MetricQuery: "name(/Device_1/Metric/@*[1])",
					Timestamp:   "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"value": "/Device_1/Metric/@*[1]",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"state",
				map[string]string{},
				map[string]interface{}{
					"value": "ok",
				},
				time.Unix(1577923199, 0),
			),
		},
		{
			name:  "parse metric name constant",
			input: metricNameQueryXML,
			configs: []Config{
				{
					MetricQuery: "'the_metric'",
					Timestamp:   "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"value": "/Device_1/Metric/@*[1]",
					},
				},
			},
			defaultTags: map[string]string{},
			expected: testutil.MustMetric(
				"the_metric",
				map[string]string{},
				map[string]interface{}{
					"value": "ok",
				},
				time.Unix(1577923199, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       tt.defaultTags,
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			actual, err := parser.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual)
		})
	}
}

func TestParseErrors(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		configs  []Config
		expected string
	}{
		{
			name:  "string metric name query",
			input: metricNameQueryXML,
			configs: []Config{
				{
					MetricQuery: "arbitrary",
					Timestamp:   "/Device_1/Timestamp_unix",
					Fields: map[string]string{
						"value": "/Device_1/Metric/@*[1]",
					},
				},
			},
			expected: "failed to query metric name: query result is of type <nil> not 'string'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       map[string]string{},
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			_, err := parser.ParseLine(tt.input)
			require.Error(t, err)
			require.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestEmptySelection(t *testing.T) {
	var tests = []struct {
		name    string
		input   string
		configs []Config
	}{
		{
			name:  "empty path",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "/Device/NonExisting",
					Fields:    map[string]string{"value": "number(Value)"},
					FieldsInt: map[string]string{"mode": "Value/@mode"},
					Tags:      map[string]string{},
				},
			},
		},
		{
			name:  "empty pattern",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "//NonExisting",
					Fields:    map[string]string{"value": "number(Value)"},
					FieldsInt: map[string]string{"mode": "Value/@mode"},
					Tags:      map[string]string{},
				},
			},
		},
		{
			name:  "empty axis",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "/Device/child::NonExisting",
					Fields:    map[string]string{"value": "number(Value)"},
					FieldsInt: map[string]string{"mode": "Value/@mode"},
					Tags:      map[string]string{},
				},
			},
		},
		{
			name:  "empty predicate",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "/Device[@NonExisting=true]",
					Fields:    map[string]string{"value": "number(Value)"},
					FieldsInt: map[string]string{"mode": "Value/@mode"},
					Tags:      map[string]string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName: "test",
				Configs:           tt.configs,
				DefaultTags:       map[string]string{},
				Log:               testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			_, err := parser.Parse([]byte(tt.input))
			require.Error(t, err)
			require.Equal(t, "cannot parse with empty selection node", err.Error())
		})
	}
}

func TestEmptySelectionAllowed(t *testing.T) {
	var tests = []struct {
		name    string
		input   string
		configs []Config
	}{
		{
			name:  "empty path",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "/Device/NonExisting",
					Fields:    map[string]string{"value": "number(Value)"},
					FieldsInt: map[string]string{"mode": "Value/@mode"},
					Tags:      map[string]string{},
				},
			},
		},
		{
			name:  "empty pattern",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "//NonExisting",
					Fields:    map[string]string{"value": "number(Value)"},
					FieldsInt: map[string]string{"mode": "Value/@mode"},
					Tags:      map[string]string{},
				},
			},
		},
		{
			name:  "empty axis",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "/Device/child::NonExisting",
					Fields:    map[string]string{"value": "number(Value)"},
					FieldsInt: map[string]string{"mode": "Value/@mode"},
					Tags:      map[string]string{},
				},
			},
		},
		{
			name:  "empty predicate",
			input: multipleNodesXML,
			configs: []Config{
				{
					Selection: "/Device[@NonExisting=true]",
					Fields:    map[string]string{"value": "number(Value)"},
					FieldsInt: map[string]string{"mode": "Value/@mode"},
					Tags:      map[string]string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &Parser{
				DefaultMetricName:   "xml",
				Configs:             tt.configs,
				AllowEmptySelection: true,
				DefaultTags:         map[string]string{},
				Log:                 testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())

			_, err := parser.Parse([]byte(tt.input))
			require.NoError(t, err)
		})
	}
}

func TestTestCases(t *testing.T) {
	var tests = []struct {
		name     string
		filename string
	}{
		{
			name:     "explicit basic",
			filename: "testcases/multisensor_explicit_basic.conf",
		},
		{
			name:     "explicit batch",
			filename: "testcases/multisensor_explicit_batch.conf",
		},
		{
			name:     "field selection batch",
			filename: "testcases/multisensor_selection_batch.conf",
		},
		{
			name:     "earthquakes quakeml",
			filename: "testcases/earthquakes.conf",
		},
		{
			name:     "openweathermap forecast (xml)",
			filename: "testcases/openweathermap_xml.conf",
		},
		{
			name:     "openweathermap forecast (json)",
			filename: "testcases/openweathermap_json.conf",
		},
		{
			name:     "addressbook tutorial (protobuf)",
			filename: "testcases/addressbook.conf",
		},
		{
			name:     "message-pack",
			filename: "testcases/tracker_msgpack.conf",
		},
		{
			name:     "field and tag batch (json)",
			filename: "testcases/field_tag_batch.conf",
		},
	}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.FromSlash(tt.filename)
			cfg, header, err := loadTestConfiguration(filename)
			require.NoError(t, err)

			// Load the xml-content
			input, err := testutil.ParseRawLinesFrom(header, "File:")
			require.NoError(t, err)
			require.Len(t, input, 1)

			filefields := strings.Fields(input[0])
			require.NotEmpty(t, filefields)
			datafile := filepath.FromSlash(filefields[0])
			fileformat := ""
			if len(filefields) > 1 {
				fileformat = filefields[1]
			}

			// Load the protocol buffer information if required
			var pbmsgdef, pbmsgtype string
			if fileformat == "xpath_protobuf" {
				input, err := testutil.ParseRawLinesFrom(header, "Protobuf:")
				require.NoError(t, err)
				require.Len(t, input, 1)

				protofields := strings.Fields(input[0])
				require.Len(t, protofields, 2)
				pbmsgdef = protofields[0]
				pbmsgtype = protofields[1]
			}

			content, err := os.ReadFile(datafile)
			require.NoError(t, err)

			// Get the expectations
			//nolint:errcheck // these may not be set by the testcase, in which case it would error correctly
			expectedOutputs, _ := testutil.ParseMetricsFrom(header, "Expected Output:", parser)

			//nolint:errcheck // these may not be set by the testcase, in which case it would error correctly
			expectedErrors, _ := testutil.ParseRawLinesFrom(header, "Expected Error:")

			// Setup the parser and run it.
			metricName := "xml"
			if fileformat != "" {
				metricName = fileformat
			}
			parser := &Parser{
				DefaultMetricName:    metricName,
				Format:               fileformat,
				ProtobufMessageFiles: []string{pbmsgdef},
				ProtobufMessageType:  pbmsgtype,
				Configs:              []Config{*cfg},
				Log:                  testutil.Logger{Name: "parsers.xml"},
			}
			require.NoError(t, parser.Init())
			outputs, err := parser.Parse(content)
			if len(expectedErrors) == 0 {
				require.NoError(t, err)
			}

			// If no timestamp is given we cannot test it. So use the one of the output
			if cfg.Timestamp == "" {
				testutil.RequireMetricsEqual(t, expectedOutputs, outputs, testutil.IgnoreTime())
			} else {
				testutil.RequireMetricsEqual(t, expectedOutputs, outputs)
			}
		})
	}
}

func TestProtobufImporting(t *testing.T) {
	// Setup the parser and run it.
	parser := &Parser{
		DefaultMetricName:    "xpath_protobuf",
		Format:               "xpath_protobuf",
		ProtobufMessageFiles: []string{"person.proto"},
		ProtobufMessageType:  "importtest.Person",
		ProtobufImportPaths:  []string{"testcases/protos"},
		Log:                  testutil.Logger{Name: "parsers.protobuf"},
	}
	require.NoError(t, parser.Init())
}

func TestMultipleConfigs(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure the folder contains data
	require.NotEmpty(t, folders)

	// Register the wrapper plugin
	inputs.Add("file", func() telegraf.Input {
		return &file.File{}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() || f.Name() == "protos" {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		t.Run(f.Name(), func(t *testing.T) {
			// Prepare the influx parser for expectations
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			parser.SetTimeFunc(func() time.Time { return time.Time{} })

			// Compare options
			options := []cmp.Option{testutil.SortMetrics()}

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}
			if len(expected) > 0 && expected[0].Time().IsZero() {
				options = append(options, testutil.IgnoreTime())
			}

			// Read the expected output if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.NotEmpty(t, cfg.Inputs)

			// Gather the metrics from the input file configure
			var acc testutil.Accumulator
			var errs []error
			for _, input := range cfg.Inputs {
				require.NoError(t, input.Init())
				err := input.Gather(&acc)
				if err != nil {
					errs = append(errs, err)
				}
			}

			// Check for errors if we expect any
			if len(expectedErrors) > 0 {
				require.Len(t, errs, len(expectedErrors))
				for i, err := range errs {
					require.ErrorContains(t, err, expectedErrors[i])
				}
			} else {
				require.Empty(t, errs)
			}

			// Process expected metrics and compare with resulting metrics
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func loadTestConfiguration(filename string) (*Config, []string, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	header := make([]string, 0)
	for _, line := range strings.Split(string(buf), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			header = append(header, line)
		}
	}
	cfg := Config{}
	err = toml.Unmarshal(buf, &cfg)
	return &cfg, header, err
}

var benchmarkExpectedMetrics = []telegraf.Metric{
	metric.New(
		"benchmark",
		map[string]string{
			"tags_host":     "myhost",
			"tags_platform": "python",
			"tags_sdkver":   "3.11.5",
		},
		map[string]interface{}{
			"value": 5.0,
		},
		time.Unix(1577923199, 0),
	),
	metric.New(
		"benchmark",
		map[string]string{
			"tags_host":     "myhost",
			"tags_platform": "python",
			"tags_sdkver":   "3.11.4",
		},
		map[string]interface{}{
			"value": 4.0,
		},
		time.Unix(1577923199, 0),
	),
}

const benchmarkDataXML = `
<?xml version="1.0"?>
	<Timestamp value="1577923199"/>
	<Benchmark>
		<tags_host>myhost</tags_host>
		<tags_sdkver>3.11.5</tags_sdkver>
		<tags_platform>python</tags_platform>
		<value>5</value>
	</Benchmark>
	<Benchmark>
		<tags_host>myhost</tags_host>
		<tags_sdkver>3.11.4</tags_sdkver>
		<tags_platform>python</tags_platform>
		<value>4</value>
	</Benchmark>
`

var benchmarkConfigXML = Config{
	Selection: "/Benchmark",
	Tags: map[string]string{
		"tags_host":     "tags_host",
		"tags_sdkver":   "tags_sdkver",
		"tags_platform": "tags_platform",
	},
	Fields: map[string]string{
		"value": "number(value)",
	},
	Timestamp:    "/Timestamp/@value",
	TimestampFmt: "unix",
}

func TestBenchmarkDataXML(t *testing.T) {
	plugin := &Parser{
		DefaultMetricName: "benchmark",
		Format:            "xml",
		Configs:           []Config{benchmarkConfigXML},
		Log:               testutil.Logger{Name: "parsers.xpath"},
	}
	require.NoError(t, plugin.Init())

	actual, err := plugin.Parse([]byte(benchmarkDataXML))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, benchmarkExpectedMetrics, actual)
}

func BenchmarkParsingXML(b *testing.B) {
	plugin := &Parser{
		DefaultMetricName: "benchmark",
		Format:            "xml",
		Configs:           []Config{benchmarkConfigXML},
		Log:               testutil.Logger{Name: "parsers.xpath", Quiet: true},
	}
	require.NoError(b, plugin.Init())

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse([]byte(benchmarkDataXML))
	}
}

const benchmarkDataJSON = `
{
	"timestamp": 1577923199,
	"data": [
		{
			"tags_host": "myhost",
			"tags_sdkver": "3.11.5",
			"tags_platform": "python",
			"value": 5.0
		},
		{
			"tags_host": "myhost",
			"tags_sdkver": "3.11.4",
			"tags_platform": "python",
			"value": 4.0
		}
	]
}
`

var benchmarkConfigJSON = Config{
	Selection: "data/*",
	Tags: map[string]string{
		"tags_host":     "tags_host",
		"tags_sdkver":   "tags_sdkver",
		"tags_platform": "tags_platform",
	},
	Fields: map[string]string{
		"value": "number(value)",
	},
	Timestamp:    "//timestamp",
	TimestampFmt: "unix",
}

func TestBenchmarkDataJSON(t *testing.T) {
	plugin := &Parser{
		DefaultMetricName: "benchmark",
		Format:            "xpath_json",
		Configs:           []Config{benchmarkConfigJSON},
		Log:               testutil.Logger{Name: "parsers.xpath"},
	}
	require.NoError(t, plugin.Init())

	actual, err := plugin.Parse([]byte(benchmarkDataJSON))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, benchmarkExpectedMetrics, actual)
}

func BenchmarkParsingJSON(b *testing.B) {
	plugin := &Parser{
		DefaultMetricName: "benchmark",
		Format:            "xpath_json",
		Configs:           []Config{benchmarkConfigJSON},
		Log:               testutil.Logger{Name: "parsers.xpath", Quiet: true},
	}
	require.NoError(b, plugin.Init())

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse([]byte(benchmarkDataJSON))
	}
}

func BenchmarkParsingProtobuf(b *testing.B) {
	plugin := &Parser{
		DefaultMetricName:    "benchmark",
		Format:               "xpath_protobuf",
		ProtobufMessageFiles: []string{"benchmark.proto"},
		ProtobufMessageType:  "benchmark.BenchmarkData",
		ProtobufImportPaths:  []string{".", "./testcases/protobuf_benchmark"},
		NativeTypes:          true,
		Configs: []Config{
			{
				Selection:    "//data",
				Timestamp:    "timestamp",
				TimestampFmt: "unix_ns",
				Tags: map[string]string{
					"source":        "source",
					"tags_sdkver":   "tags_sdkver",
					"tags_platform": "tags_platform",
				},
				Fields: map[string]string{
					"value": "value",
				},
			},
		},
		Log: testutil.Logger{Name: "parsers.xpath", Quiet: true},
	}
	require.NoError(b, plugin.Init())

	benchmarkData, err := os.ReadFile(filepath.Join("testcases", "protobuf_benchmark", "message.bin"))
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(benchmarkData)
	}
}

var benchmarkDataMsgPack = [][]byte{
	{
		0xdf, 0x00, 0x00, 0x00, 0x05, 0xa9, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0xce,
		0x62, 0x90, 0x98, 0x9d, 0xa5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x05, 0xa6, 0x73, 0x6f, 0x75, 0x72,
		0x63, 0x65, 0xa6, 0x6d, 0x79, 0x68, 0x6f, 0x73, 0x74, 0xad, 0x74, 0x61, 0x67, 0x73, 0x5f, 0x70,
		0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0xa6, 0x70, 0x79, 0x74, 0x68, 0x6f, 0x6e, 0xab, 0x74,
		0x61, 0x67, 0x73, 0x5f, 0x73, 0x64, 0x6b, 0x76, 0x65, 0x72, 0xa6, 0x33, 0x2e, 0x31, 0x31, 0x2e,
		0x35,
	},
	{
		0x85, 0xA6, 0x73, 0x6F, 0x75, 0x72, 0x63, 0x65, 0xA6, 0x6D, 0x79, 0x68, 0x6F, 0x73, 0x74, 0xAB,
		0x74, 0x61, 0x67, 0x73, 0x5F, 0x73, 0x64, 0x6B, 0x76, 0x65, 0x72, 0xA6, 0x33, 0x2E, 0x31, 0x31,
		0x2E, 0x34, 0xAD, 0x74, 0x61, 0x67, 0x73, 0x5F, 0x70, 0x6C, 0x61, 0x74, 0x66, 0x6F, 0x72, 0x6D,
		0xA6, 0x70, 0x79, 0x74, 0x68, 0x6F, 0x6E, 0xA5, 0x76, 0x61, 0x6C, 0x75, 0x65, 0x04, 0xA9, 0x74,
		0x69, 0x6D, 0x65, 0x73, 0x74, 0x61, 0x6D, 0x70, 0xCE, 0x62, 0x90, 0x98, 0x9D,
	},
}

func TestBenchmarkDataMsgPack(t *testing.T) {
	plugin := &Parser{
		DefaultMetricName: "benchmark",
		Format:            "xpath_msgpack",
		Configs: []Config{
			{
				Tags: map[string]string{
					"source":        "source",
					"tags_sdkver":   "tags_sdkver",
					"tags_platform": "tags_platform",
				},
				Fields: map[string]string{
					"value": "number(value)",
				},
				Timestamp:    "timestamp",
				TimestampFmt: "unix",
			},
		},
		Log: testutil.Logger{Name: "parsers.xpath", Quiet: true},
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"benchmark",
			map[string]string{
				"source":        "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.5",
			},
			map[string]interface{}{
				"value": 5.0,
			},
			time.Unix(1653643421, 0),
		),
		metric.New(
			"benchmark",
			map[string]string{
				"source":        "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.4",
			},
			map[string]interface{}{
				"value": 4.0,
			},
			time.Unix(1653643421, 0),
		),
	}

	actual := make([]telegraf.Metric, 0, 2)
	for _, msg := range benchmarkDataMsgPack {
		m, err := plugin.Parse(msg)
		require.NoError(t, err)
		actual = append(actual, m...)
	}
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func BenchmarkParsingMsgPack(b *testing.B) {
	plugin := &Parser{
		DefaultMetricName: "benchmark",
		Format:            "xpath_msgpack",
		Configs: []Config{
			{
				Tags: map[string]string{
					"source":        "source",
					"tags_sdkver":   "tags_sdkver",
					"tags_platform": "tags_platform",
				},
				Fields: map[string]string{
					"value": "number(value)",
				},
				Timestamp:    "timestamp",
				TimestampFmt: "unix",
			},
		},
		Log: testutil.Logger{Name: "parsers.xpath", Quiet: true},
	}
	require.NoError(b, plugin.Init())

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(benchmarkDataMsgPack[n%2])
	}
}

func BenchmarkParsingCBOR(b *testing.B) {
	plugin := &Parser{
		DefaultMetricName: "benchmark",
		Format:            "xpath_cbor",
		NativeTypes:       true,
		Configs: []Config{
			{
				Selection:    "//data",
				Timestamp:    "timestamp",
				TimestampFmt: "unix_ns",
				Tags: map[string]string{
					"source":        "source",
					"tags_sdkver":   "tags_sdkver",
					"tags_platform": "tags_platform",
				},
				Fields: map[string]string{
					"value": "value",
				},
			},
		},
		Log: testutil.Logger{Name: "parsers.xpath", Quiet: true},
	}
	require.NoError(b, plugin.Init())

	benchmarkData, err := os.ReadFile(filepath.Join("testcases", "cbor_benchmark", "message.bin"))
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(benchmarkData)
	}
}
