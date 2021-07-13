package value

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseValidValues(t *testing.T) {
	parser := NewValueParser("value_test", "integer", "", nil)
	metrics, err := parser.Parse([]byte("55"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = NewValueParser("value_test", "float", "", nil)
	metrics, err = parser.Parse([]byte("64"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = NewValueParser("value_test", "string", "", nil)
	metrics, err = parser.Parse([]byte("foobar"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	parser = NewValueParser("value_test", "boolean", "", nil)
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
	parser := NewValueParser("value_test", "integer", "", nil)
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

func TestParseCustomFieldName(t *testing.T) {
	parser := NewValueParser("value_test", "integer", "", nil)
	parser.FieldName = "penguin"
	metrics, err := parser.Parse([]byte(`55`))

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"penguin": int64(55),
	}, metrics[0].Fields())
}

func TestParseLineValidValues(t *testing.T) {
	parser := NewValueParser("value_test", "integer", "", nil)
	metric, err := parser.ParseLine("55")
	assert.NoError(t, err)
	assert.Equal(t, "value_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	parser = NewValueParser("value_test", "float", "", nil)
	metric, err = parser.ParseLine("64")
	assert.NoError(t, err)
	assert.Equal(t, "value_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	parser = NewValueParser("value_test", "string", "", nil)
	metric, err = parser.ParseLine("foobar")
	assert.NoError(t, err)
	assert.Equal(t, "value_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	parser = NewValueParser("value_test", "boolean", "", nil)
	metric, err = parser.ParseLine("true")
	assert.NoError(t, err)
	assert.Equal(t, "value_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": true,
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())
}

func TestParseInvalidValues(t *testing.T) {
	parser := NewValueParser("value_test", "integer", "", nil)
	metrics, err := parser.Parse([]byte("55.0"))
	assert.Error(t, err)
	assert.Len(t, metrics, 0)

	parser = NewValueParser("value_test", "float", "", nil)
	metrics, err = parser.Parse([]byte("foobar"))
	assert.Error(t, err)
	assert.Len(t, metrics, 0)

	parser = NewValueParser("value_test", "boolean", "", nil)
	metrics, err = parser.Parse([]byte("213"))
	assert.Error(t, err)
	assert.Len(t, metrics, 0)
}

func TestParseLineInvalidValues(t *testing.T) {
	parser := NewValueParser("value_test", "integer", "", nil)
	_, err := parser.ParseLine("55.0")
	assert.Error(t, err)

	parser = NewValueParser("value_test", "float", "", nil)
	_, err = parser.ParseLine("foobar")
	assert.Error(t, err)

	parser = NewValueParser("value_test", "boolean", "", nil)
	_, err = parser.ParseLine("213")
	assert.Error(t, err)
}

func TestParseValidValuesDefaultTags(t *testing.T) {
	parser := NewValueParser("value_test", "integer", "", nil)
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err := parser.Parse([]byte("55"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = NewValueParser("value_test", "float", "", nil)
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err = parser.Parse([]byte("64"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(64),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = NewValueParser("value_test", "string", "", nil)
	parser.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err = parser.Parse([]byte("foobar"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": "foobar",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"test": "tag"}, metrics[0].Tags())

	parser = NewValueParser("value_test", "boolean", "", nil)
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
	parser := NewValueParser("value_test", "integer", "", nil)
	metrics, err := parser.Parse([]byte("55\x00"))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "value_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": int64(55),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}
