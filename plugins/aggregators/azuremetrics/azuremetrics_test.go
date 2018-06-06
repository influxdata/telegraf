package azuremetrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	constants "github.com/influxdata/telegraf/utility"
)

var m1, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a": int64(1),
		"b": int64(1),
		"c": float64(2),
		"d": float64(2),
	},
	time.Now(),
)
var m2, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a":        int64(1),
		"b":        int64(3),
		"c":        float64(4),
		"d":        float64(6),
		"e":        float64(200),
		"ignoreme": "string",
		"andme":    true,
	},
	time.Now(),
)

// Test two metrics getting added.
func TestAzureMetricsAddingTwoMetrics(t *testing.T) {
	acc := testutil.Accumulator{}
	minmax := NewAzureMetrics()
	minmax.PeriodTag = "30s"
	minmax.Add(m1)
	minmax.Add(m2)
	minmax.Push(&acc)

	expectedFieldsInA := map[string]interface{}{
		constants.SAMPLE_COUNT: float64(2), //a
		constants.MAX_SAMPLE:   float64(1),
		constants.MIN_SAMPLE:   float64(1),
		constants.MEAN:         float64(1),
	}
	expectedTagsInA := map[string]string{
		"foo":    "bar",
		"period": "30s",
	}

	isMeasurementFound := false

	for _, metric := range acc.Metrics {
		if metric.Measurement == "a" {
			isMeasurementFound = true
			for key, val := range expectedFieldsInA {
				assert.Contains(t, metric.Fields, key)
				assert.EqualValues(t, metric.Fields[key], val)
			}
			for key, val := range expectedTagsInA {
				assert.Contains(t, metric.Tags, key)
				assert.EqualValues(t, metric.Tags[key], val)
			}
		}
	}
	assert.Equal(t, isMeasurementFound, true)
}
