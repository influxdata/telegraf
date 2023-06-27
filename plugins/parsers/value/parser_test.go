package value

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestParseValidValues(t *testing.T) {
	tests := []struct {
		name     string
		dtype    string
		input    []byte
		expected interface{}
	}{
		{
			name:     "integer",
			dtype:    "integer",
			input:    []byte("55"),
			expected: int64(55),
		},
		{
			name:     "float",
			dtype:    "float",
			input:    []byte("64"),
			expected: float64(64),
		},
		{
			name:     "string",
			dtype:    "string",
			input:    []byte("foobar"),
			expected: "foobar",
		},
		{
			name:     "boolean",
			dtype:    "boolean",
			input:    []byte("true"),
			expected: true,
		},
		{
			name:     "multiple integers",
			dtype:    "integer",
			input:    []byte(`55 45 223 12 999`),
			expected: int64(999),
		},
		{
			name:     "auto integer",
			dtype:    "auto_integer",
			input:    []byte("55"),
			expected: int64(55),
		},
		{
			name:     "auto integer with string",
			dtype:    "auto_integer",
			input:    []byte("foobar"),
			expected: "foobar",
		},
		{
			name:     "auto integer with float",
			dtype:    "auto_integer",
			input:    []byte("55.0"),
			expected: "55.0",
		},
		{
			name:     "auto float",
			dtype:    "auto_float",
			input:    []byte("64.2"),
			expected: float64(64.2),
		},
		{
			name:     "auto float with string",
			dtype:    "auto_float",
			input:    []byte("foobar"),
			expected: "foobar",
		},
		{
			name:     "auto float with integer",
			dtype:    "auto_float",
			input:    []byte("64"),
			expected: float64(64),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := metric.New(
				"value_test",
				map[string]string{},
				map[string]interface{}{"value": tt.expected},
				time.Unix(0, 0),
			)

			plugin := Parser{
				MetricName: "value_test",
				DataType:   tt.dtype,
			}
			require.NoError(t, plugin.Init())
			actual, err := plugin.Parse(tt.input)
			require.NoError(t, err)
			require.Len(t, actual, 1)
			testutil.RequireMetricEqual(t, expected, actual[0], testutil.IgnoreTime())
		})
	}
}

func TestParseLineValidValues(t *testing.T) {
	tests := []struct {
		name     string
		dtype    string
		input    string
		expected interface{}
	}{
		{
			name:     "integer",
			dtype:    "integer",
			input:    "55",
			expected: int64(55),
		},
		{
			name:     "float",
			dtype:    "float",
			input:    "64",
			expected: float64(64),
		},
		{
			name:     "string",
			dtype:    "string",
			input:    "foobar",
			expected: "foobar",
		},
		{
			name:     "boolean",
			dtype:    "boolean",
			input:    "true",
			expected: true,
		},
		{
			name:     "multiple integers",
			dtype:    "integer",
			input:    `55 45 223 12 999`,
			expected: int64(999),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := metric.New(
				"value_test",
				map[string]string{},
				map[string]interface{}{"value": tt.expected},
				time.Unix(0, 0),
			)

			plugin := Parser{
				MetricName: "value_test",
				DataType:   tt.dtype,
			}
			require.NoError(t, plugin.Init())
			actual, err := plugin.ParseLine(tt.input)
			require.NoError(t, err)
			testutil.RequireMetricEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestParseCustomFieldName(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
		FieldName:  "penguin",
	}
	require.NoError(t, parser.Init())

	metrics, err := parser.Parse([]byte(`55`))
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"penguin": int64(55)}, metrics[0].Fields())
}

func TestParseInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		dtype string
		input []byte
	}{
		{
			name:  "integer",
			dtype: "integer",
			input: []byte("55.0"),
		},
		{
			name:  "float",
			dtype: "float",
			input: []byte("foobar"),
		},
		{
			name:  "boolean",
			dtype: "boolean",
			input: []byte("213"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Parser{
				MetricName: "value_test",
				DataType:   tt.dtype,
			}
			require.NoError(t, plugin.Init())
			actual, err := plugin.Parse(tt.input)
			require.ErrorContains(t, err, "invalid syntax")
			require.Empty(t, actual)
		})
	}
}

func TestParseLineInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		dtype string
		input string
	}{
		{
			name:  "integer",
			dtype: "integer",
			input: "55.0",
		},
		{
			name:  "float",
			dtype: "float",
			input: "foobar",
		},
		{
			name:  "boolean",
			dtype: "boolean",
			input: "213",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Parser{
				MetricName: "value_test",
				DataType:   tt.dtype,
			}
			require.NoError(t, plugin.Init())
			actual, err := plugin.ParseLine(tt.input)
			require.ErrorContains(t, err, "invalid syntax")
			require.Empty(t, actual)
		})
	}
}

func TestParseValidValuesDefaultTags(t *testing.T) {
	expected := metric.New(
		"value_test",
		map[string]string{"test": "tag"},
		map[string]interface{}{"value": int64(55)},
		time.Unix(0, 0),
	)

	plugin := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	require.NoError(t, plugin.Init())
	plugin.SetDefaultTags(map[string]string{"test": "tag"})

	actual, err := plugin.Parse([]byte("55"))
	require.NoError(t, err)
	require.Len(t, actual, 1)

	testutil.RequireMetricEqual(t, expected, actual[0], testutil.IgnoreTime())
}

func TestParseValuesWithNullCharacter(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	require.NoError(t, parser.Init())
	metrics, err := parser.Parse([]byte("55\x00"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())
}

func TestInvalidDatatype(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "foo",
	}
	require.ErrorContains(t, parser.Init(), "unknown datatype")
}
