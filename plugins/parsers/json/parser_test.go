package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validJSON              = "{\"a\": 5, \"b\": {\"c\": 6}}"
	validJSONNewline       = "\n{\"d\": 7, \"b\": {\"d\": 8}}\n"
	validJSONArray         = "[{\"a\": 5, \"b\": {\"c\": 6}}]"
	validJSONArrayMultiple = "[{\"a\": 5, \"b\": {\"c\": 6}}, {\"a\": 7, \"b\": {\"c\": 8}}]"
	invalidJSON            = "I don't think this is JSON"
	invalidJSON2           = "{\"a\": 5, \"b\": \"c\": 6}}"
)

const validJSONTags = `
{
    "a": 5,
    "b": {
        "c": 6
    },
    "mytag": "foobar",
    "othertag": "baz"
}
`

const validJSONArrayTags = `
[
{
    "a": 5,
    "b": {
        "c": 6
    },
    "mytag": "foo",
    "othertag": "baz"
},
{
    "a": 7,
    "b": {
        "c": 8
    },
    "mytag": "bar",
    "othertag": "baz"
}
]
`

func TestParseValidJSON(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
	}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	// Test that newlines are fine
	metrics, err = parser.Parse([]byte(validJSONNewline))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"d":   float64(7),
		"b_d": float64(8),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	// Test that strings without TagKeys defined are ignored
	metrics, err = parser.Parse([]byte(validJSONTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

func TestParseLineValidJSON(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
	}

	// Most basic vanilla test
	metric, err := parser.ParseLine(validJSON)
	assert.NoError(t, err)
	assert.Equal(t, "json_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	// Test that newlines are fine
	metric, err = parser.ParseLine(validJSONNewline)
	assert.NoError(t, err)
	assert.Equal(t, "json_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"d":   float64(7),
		"b_d": float64(8),
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	// Test that strings without TagKeys defined are ignored
	metric, err = parser.ParseLine(validJSONTags)
	assert.NoError(t, err)
	assert.Equal(t, "json_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())
}

func TestParseInvalidJSON(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
	}

	_, err := parser.Parse([]byte(invalidJSON))
	assert.Error(t, err)
	_, err = parser.Parse([]byte(invalidJSON2))
	assert.Error(t, err)
	_, err = parser.ParseLine(invalidJSON)
	assert.Error(t, err)
}

func TestParseWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	metrics, err := parser.Parse([]byte(validJSONTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	// Test that single tag key is found and applied
	parser = JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
	}
	metrics, err = parser.Parse([]byte(validJSONTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"mytag": "foobar",
	}, metrics[0].Tags())

	// Test that both tag keys are found and applied
	parser = JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	metrics, err = parser.Parse([]byte(validJSONTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"mytag":    "foobar",
		"othertag": "baz",
	}, metrics[0].Tags())
}

func TestParseLineWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	metric, err := parser.ParseLine(validJSONTags)
	assert.NoError(t, err)
	assert.Equal(t, "json_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	assert.Equal(t, map[string]string{}, metric.Tags())

	// Test that single tag key is found and applied
	parser = JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
	}
	metric, err = parser.ParseLine(validJSONTags)
	assert.NoError(t, err)
	assert.Equal(t, "json_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"mytag": "foobar",
	}, metric.Tags())

	// Test that both tag keys are found and applied
	parser = JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	metric, err = parser.ParseLine(validJSONTags)
	assert.NoError(t, err)
	assert.Equal(t, "json_test", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"mytag":    "foobar",
		"othertag": "baz",
	}, metric.Tags())
}

func TestParseValidJSONDefaultTags(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
		DefaultTags: map[string]string{
			"t4g": "default",
		},
	}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"t4g": "default"}, metrics[0].Tags())

	// Test that tagkeys and default tags are applied
	metrics, err = parser.Parse([]byte(validJSONTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"t4g":   "default",
		"mytag": "foobar",
	}, metrics[0].Tags())
}

// Test that default tags are overridden by tag keys
func TestParseValidJSONDefaultTagsOverride(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
		DefaultTags: map[string]string{
			"mytag": "default",
		},
	}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"mytag": "default"}, metrics[0].Tags())

	// Test that tagkeys override default tags
	metrics, err = parser.Parse([]byte(validJSONTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"mytag": "foobar",
	}, metrics[0].Tags())
}

// Test that json arrays can be parsed
func TestParseValidJSONArray(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_array_test",
	}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validJSONArray))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "json_array_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	// Basic multiple datapoints
	metrics, err = parser.Parse([]byte(validJSONArrayMultiple))
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "json_array_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[1].Tags())
	assert.Equal(t, "json_array_test", metrics[1].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{}, metrics[1].Tags())
}

func TestParseArrayWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := JSONParser{
		MetricName: "json_array_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	metrics, err := parser.Parse([]byte(validJSONArrayTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "json_array_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())

	assert.Equal(t, "json_array_test", metrics[1].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{}, metrics[1].Tags())

	// Test that single tag key is found and applied
	parser = JSONParser{
		MetricName: "json_array_test",
		TagKeys:    []string{"mytag"},
	}
	metrics, err = parser.Parse([]byte(validJSONArrayTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "json_array_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"mytag": "foo",
	}, metrics[0].Tags())

	assert.Equal(t, "json_array_test", metrics[1].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{
		"mytag": "bar",
	}, metrics[1].Tags())

	// Test that both tag keys are found and applied
	parser = JSONParser{
		MetricName: "json_array_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	metrics, err = parser.Parse([]byte(validJSONArrayTags))
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "json_array_test", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"mytag":    "foo",
		"othertag": "baz",
	}, metrics[0].Tags())

	assert.Equal(t, "json_array_test", metrics[1].Name())
	assert.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{
		"mytag":    "bar",
		"othertag": "baz",
	}, metrics[1].Tags())
}
