package value

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseValidValues(t *testing.T) {
	parser := ValueParser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metrics, err := parser.Parse([]byte("55"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "float",
	}
	metrics, err = parser.Parse([]byte("64"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "string",
	}
	metrics, err = parser.Parse([]byte("foobar"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	metrics, err = parser.Parse([]byte("true"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": true,
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

func TestParseMultipleValues(t *testing.T) {
	parser := ValueParser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metrics, err := parser.Parse([]byte(`55
45
223
12
999
`))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(999),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

func TestParseLineValidValues(t *testing.T) {
	parser := ValueParser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metric, err := parser.ParseLine("55")
	assert.NoError(t, err)
	assert.Equal(t, "value_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "float",
	}
	metric, err = parser.ParseLine("64")
	assert.NoError(t, err)
	assert.Equal(t, "value_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "string",
	}
	metric, err = parser.ParseLine("foobar")
	assert.NoError(t, err)
	assert.Equal(t, "value_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	metric, err = parser.ParseLine("true")
	assert.NoError(t, err)
	assert.Equal(t, "value_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": true,
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())
}

func TestParseInvalidValues(t *testing.T) {
	parser := ValueParser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metrics, err := parser.Parse([]byte("55.0"))
	assert.Error(t, err)
	assert.Len(t, metrics, 0)

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "float",
	}
	metrics, err = parser.Parse([]byte("foobar"))
	assert.Error(t, err)
	assert.Len(t, metrics, 0)

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	metrics, err = parser.Parse([]byte("213"))
	assert.Error(t, err)
	assert.Len(t, metrics, 0)
}

func TestParseLineInvalidValues(t *testing.T) {
	parser := ValueParser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	_, err := parser.ParseLine("55.0")
	assert.Error(t, err)

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "float",
	}
	_, err = parser.ParseLine("foobar")
	assert.Error(t, err)

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	_, err = parser.ParseLine("213")
	assert.Error(t, err)
}

func TestParseValidValuesDefaultTags(t *testing.T) {
	parser := ValueParser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err := parser.Parse([]byte("55"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "float",
	}
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err = parser.Parse([]byte("64"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "string",
	}
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err = parser.Parse([]byte("foobar"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = ValueParser{
		MetricName: "value_test",
		DataType:   "boolean",
	}
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err = parser.Parse([]byte("true"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": true,
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())
}

func TestParseValuesWithNullCharacter(t *testing.T) {
	parser := ValueParser{
		MetricName: "value_test",
		DataType:   "integer",
	}
	metrics, err := parser.Parse([]byte("55\x00"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}
