package influx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var exptime = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

const (
	validInflux        = "cpu_load_short,cpu=cpu0 value=10 1257894000000000000"
	validInfluxNewline = "\ncpu_load_short,cpu=cpu0 value=10 1257894000000000000\n"
	invalidInflux      = "I don't think this is line protocol"
	invalidInflux2     = "{\"a\": 5, \"b\": {\"c\": 6}}"
)

const influxMulti = `
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

const influxMultiSomeInvalid = `
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu3, host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu4 , usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

func TestParseValidInflux(t *testing.T) {
	parser := InfluxParser{}

	metrics, err := parser.Parse([]byte(validInflux))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "cpu_load_short", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metrics[0].Tags())
	assert.Equal(t, exptime, metrics[0].Time())

	metrics, err = parser.Parse([]byte(validInfluxNewline))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "cpu_load_short", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metrics[0].Tags())
	assert.Equal(t, exptime, metrics[0].Time())
}

func TestParseLineValidInflux(t *testing.T) {
	parser := InfluxParser{}

	metric, err := parser.ParseLine(validInflux)
	assert.NoError(t, err)
	assert.Equal(t, "cpu_load_short", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metric.Tags())
	assert.Equal(t, exptime, metric.Time())

	metric, err = parser.ParseLine(validInfluxNewline)
	assert.NoError(t, err)
	assert.Equal(t, "cpu_load_short", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metric.Tags())
	assert.Equal(t, exptime, metric.Time())
}

func TestParseMultipleValid(t *testing.T) {
	parser := InfluxParser{}

	metrics, err := parser.Parse([]byte(influxMulti))
	assert.NoError(t, err)
	assert.Len(t, metrics, 7)

	for _, metric := range metrics {
		assert.Equal(t, "cpu", metric.Name())
		assert.Equal(t, map[string]string{
			"datacenter": "us-east",
			"host":       "foo",
		}, metrics[0].Tags())
		assert.Equal(t, map[string]interface{}{
			"usage_idle": float64(99),
			"usage_busy": float64(1),
		}, metrics[0].Fields())
	}
}

func TestParseSomeValid(t *testing.T) {
	parser := InfluxParser{}

	metrics, err := parser.Parse([]byte(influxMultiSomeInvalid))
	assert.Error(t, err)
	assert.Len(t, metrics, 4)

	for _, metric := range metrics {
		assert.Equal(t, "cpu", metric.Name())
		assert.Equal(t, map[string]string{
			"datacenter": "us-east",
			"host":       "foo",
		}, metrics[0].Tags())
		assert.Equal(t, map[string]interface{}{
			"usage_idle": float64(99),
			"usage_busy": float64(1),
		}, metrics[0].Fields())
	}
}

// Test that default tags are applied.
func TestParseDefaultTags(t *testing.T) {
	parser := InfluxParser{
		DefaultTags: map[string]string{
			"tag": "default",
		},
	}

	metrics, err := parser.Parse([]byte(influxMultiSomeInvalid))
	assert.Error(t, err)
	assert.Len(t, metrics, 4)

	for _, metric := range metrics {
		assert.Equal(t, "cpu", metric.Name())
		assert.Equal(t, map[string]string{
			"datacenter": "us-east",
			"host":       "foo",
			"tag":        "default",
		}, metrics[0].Tags())
		assert.Equal(t, map[string]interface{}{
			"usage_idle": float64(99),
			"usage_busy": float64(1),
		}, metrics[0].Fields())
	}
}

// Verify that metric tags will override default tags
func TestParseDefaultTagsOverride(t *testing.T) {
	parser := InfluxParser{
		DefaultTags: map[string]string{
			"host": "default",
		},
	}

	metrics, err := parser.Parse([]byte(influxMultiSomeInvalid))
	assert.Error(t, err)
	assert.Len(t, metrics, 4)

	for _, metric := range metrics {
		assert.Equal(t, "cpu", metric.Name())
		assert.Equal(t, map[string]string{
			"datacenter": "us-east",
			"host":       "foo",
		}, metrics[0].Tags())
		assert.Equal(t, map[string]interface{}{
			"usage_idle": float64(99),
			"usage_busy": float64(1),
		}, metrics[0].Fields())
	}
}

func TestParseInvalidInflux(t *testing.T) {
	parser := InfluxParser{}

	_, err := parser.Parse([]byte(invalidInflux))
	assert.Error(t, err)
	_, err = parser.Parse([]byte(invalidInflux2))
	assert.Error(t, err)
	_, err = parser.ParseLine(invalidInflux)
	assert.Error(t, err)
	_, err = parser.ParseLine(invalidInflux2)
	assert.Error(t, err)
}
