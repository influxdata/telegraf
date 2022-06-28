package value

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseValidValues(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metrics, err := parser.Parse([]byte("55"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "float",
	}
	metrics, err = parser.Parse([]byte("64"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "string",
	}
	metrics, err = parser.Parse([]byte("foobar"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	metrics, err = parser.Parse([]byte("true"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": true,
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())
}

func TestParseMultipleValues(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metrics, err := parser.Parse([]byte(`55
45
223
12
999
`))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": int64(999),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())
}

func TestParseCustomFieldName(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	parser.FieldName = "penguin"
	metrics, err := parser.Parse([]byte(`55`))

	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{
		"penguin": int64(55),
	}, metrics[0].Fields())
}

func TestParseLineValidValues(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metric, err := parser.ParseLine("55")
	require.NoError(t, err)
	require.Equal(t, "value_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "float",
	}
	metric, err = parser.ParseLine("64")
	require.NoError(t, err)
	require.Equal(t, "value_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "string",
	}
	metric, err = parser.ParseLine("foobar")
	require.NoError(t, err)
	require.Equal(t, "value_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	metric, err = parser.ParseLine("true")
	require.NoError(t, err)
	require.Equal(t, "value_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"value": true,
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())
}

func TestParseInvalidValues(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metrics, err := parser.Parse([]byte("55.0"))
	require.Error(t, err)
	require.Len(t, metrics, 0)

	parser = Parser{
		MetricName: "value_test",
		DataType:   "float",
	}
	metrics, err = parser.Parse([]byte("foobar"))
	require.Error(t, err)
	require.Len(t, metrics, 0)

	parser = Parser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	metrics, err = parser.Parse([]byte("213"))
	require.Error(t, err)
	require.Len(t, metrics, 0)
}

func TestParseLineInvalidValues(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	_, err := parser.ParseLine("55.0")
	require.Error(t, err)

	parser = Parser{
		MetricName: "value_test",
		DataType:   "float",
	}
	_, err = parser.ParseLine("foobar")
	require.Error(t, err)

	parser = Parser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	_, err = parser.ParseLine("213")
	require.Error(t, err)
}

func TestParseValidValuesDefaultTags(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err := parser.Parse([]byte("55"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "float",
	}
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err = parser.Parse([]byte("64"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "string",
	}
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err = parser.Parse([]byte("foobar"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = Parser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err = parser.Parse([]byte("true"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": true,
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())
}

func TestParseValuesWithNullCharacter(t *testing.T) {
	parser := Parser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metrics, err := parser.Parse([]byte("55\x00"))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "value_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())
}
