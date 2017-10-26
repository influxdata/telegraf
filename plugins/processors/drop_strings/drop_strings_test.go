package drop_strings

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestDropStrings(t *testing.T) {
	var metric, _ = metric.New("Test Metric",
		map[string]string{"state": "full"},
		map[string]interface{}{
			"integer": int64(23),
			"float":   float64(3.1415),
			"bool":    true,
			"string":  "should be dropped",
		},
		time.Now(),
	)

	dropStrings := DropStrings{}

	result := dropStrings.Apply(metric)[0]
	fields := result.Fields()

	assert.NotContains(t, fields, "string")

	assertFieldValue(t, int64(23), "integer", fields)
	assertFieldValue(t, float64(3.1415), "float", fields)
	assertFieldValue(t, true, "bool", fields)

	assert.Equal(t, "Test Metric", result.Name())
	assert.Equal(t, metric.Tags(), result.Tags())
	assert.Equal(t, metric.Time(), result.Time())
}

func assertFieldValue(t *testing.T, expected interface{}, field string, fields map[string]interface{}) {
	value, present := fields[field]
	assert.True(t, present, "value of field '"+field+"' was not present")
	assert.EqualValues(t, expected, value)
}

func TestDropEntireMetricIfOnlyStrings(t *testing.T) {
	var metric, _ = metric.New("Test Metric",
		map[string]string{"state": "full"},
		map[string]interface{}{
			"string1": "should be dropped",
			"string2": "should be dropped, also",
		},
		time.Now(),
	)

	dropStrings := DropStrings{}

	result := dropStrings.Apply(metric)

	assert.Len(t, result, 0, "No metric should be emitted.")
}
