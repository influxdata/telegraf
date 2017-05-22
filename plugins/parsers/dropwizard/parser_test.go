package dropwizard

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
	"time"
)

// validEmptyJSON is a valid dropwizard json document, but without any metrics
const validEmptyJSON = `
{
	"version": 		"3.0.0",
	"counters" :	{},
	"meters" :		{},
	"gauges" :		{},
	"histograms" :	{},
	"timers" :		{}
}
`

func TestParseValidEmptyJSON(t *testing.T) {
	parser := Parser{}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validEmptyJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)
}

// validCounterJSON is a valid dropwizard json document containing one counter
const validCounterJSON = `
{
	"version": 		"3.0.0",
	"counters" : 	{ 
		"measurement" : {
			"count" : 1
		}
	},
	"meters" : 		{},
	"gauges" : 		{},
	"histograms" : 	{},
	"timers" : 		{}
}
`

func TestParseValidCounterJSON(t *testing.T) {
	parser := Parser{}

	metrics, err := parser.Parse([]byte(validCounterJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "counter.measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

// validEmbeddedCounterJSON is a valid json document containing separate fields for dropwizard metrics, tags and time override.
const validEmbeddedCounterJSON = `
{
	"time" : "2017-02-22T14:33:03.662+02:00",
	"tags" : {
		"tag1" : "green",
		"tag2" : "yellow"
	},
	"metrics" : {
		"counters" : 	{ 
			"measurement" : {
				"count" : 1
			}
		},
		"meters" : 		{},
		"gauges" : 		{},
		"histograms" : 	{},
		"timers" : 		{}
	}
}
`

func TestParseValidEmbeddedCounterJSON(t *testing.T) {
	timeFormat := "2006-01-02T15:04:05Z07:00"
	metricTime, _ := time.Parse(timeFormat, "2017-02-22T15:33:03.662+03:00")
	parser := Parser{
		MetricRegistryPath: "metrics",
		TagsPath:           "tags",
		TimePath:           "time",
	}

	metrics, err := parser.Parse([]byte(validEmbeddedCounterJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "counter.measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"tag1": "green", "tag2": "yellow"}, metrics[0].Tags())
	assert.True(t, metricTime.Equal(metrics[0].Time()), fmt.Sprintf("%s should be equal to %s", metrics[0].Time(), metricTime))

	// now test json tags through TagPathsMap
	parser2 := Parser{
		MetricRegistryPath: "metrics",
		TagPathsMap:        map[string]string{"tag1": "tags.tag1"},
		TimePath:           "time",
	}
	metrics2, err2 := parser2.Parse([]byte(validEmbeddedCounterJSON))
	assert.NoError(t, err2)
	assert.Equal(t, map[string]string{"tag1": "green"}, metrics2[0].Tags())
}

// validMeterJSON is a valid dropwizard json document containing two meters, with the second meter containing one tag
const validMeterJSON = `
{
	"version": 		"3.0.0",
	"counters" : 	{},
	"meters" : 		{ 
		"measurement1" : {
			"count" : 1,
			"m15_rate" : 1.0,
			"m1_rate" : 1.0,
			"m5_rate" : 1.0,
			"mean_rate" : 1.0,
			"units" : "events/second"
		},
		"measurement2,key=value" : {
			"count" : 2,
			"m15_rate" : 2.0,
			"m1_rate" : 2.0,
			"m5_rate" : 2.0,
			"mean_rate" : 2.0,
			"units" : "events/second"
		}
	},
	"gauges" : 		{},
	"histograms" : 	{},
	"timers" : 		{}
}
`

func TestParseValidMeterJSON(t *testing.T) {
	parser := Parser{}

	metrics, err := parser.Parse([]byte(validMeterJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
	assert.Equal(t, "meter.measurement1", metrics[0].Name())
	assert.Equal(t, "meter.measurement2", metrics[1].Name())
	assert.Equal(t, map[string]interface{}{
		"count":     float64(1),
		"m15_rate":  float64(1),
		"m1_rate":   float64(1),
		"m5_rate":   float64(1),
		"mean_rate": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]interface{}{
		"count":     float64(2),
		"m15_rate":  float64(2),
		"m1_rate":   float64(2),
		"m5_rate":   float64(2),
		"mean_rate": float64(2),
	}, metrics[1].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
	assert.Equal(t, map[string]string{"key": "value"}, metrics[1].Tags())
}

// validGaugeJSON is a valid dropwizard json document containing one gauge
const validGaugeJSON = `
{
	"version": 		"3.0.0",
	"counters" : 	{},
	"meters" : 		{},
	"gauges" : 		{
		"measurement" : {
			"value" : 0
		}
	},
	"histograms" : 	{},
	"timers" : 		{}
}
`

func TestParseValidGaugeJSON(t *testing.T) {
	parser := Parser{}

	metrics, err := parser.Parse([]byte(validGaugeJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "gauge.measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(0),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

// validHistogramJSON is a valid dropwizard json document containing one histogram
const validHistogramJSON = `
{
	"version": 		"3.0.0",
	"counters" : 	{},
	"meters" : 		{},
	"gauges" : 		{},
	"histograms" : 	{
		"measurement" : {
			"count" : 1,
			"max" : 2,
			"mean" : 3,
			"min" : 4,
			"p50" : 5,
			"p75" : 6,
			"p95" : 7,
			"p98" : 8,
			"p99" : 9,
			"p999" : 10,
			"stddev" : 11
		}
	},
	"timers" : 		{}
}
`

func TestParseValidHistogramJSON(t *testing.T) {
	parser := Parser{}

	metrics, err := parser.Parse([]byte(validHistogramJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "histogram.measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count":  float64(1),
		"max":    float64(2),
		"mean":   float64(3),
		"min":    float64(4),
		"p50":    float64(5),
		"p75":    float64(6),
		"p95":    float64(7),
		"p98":    float64(8),
		"p99":    float64(9),
		"p999":   float64(10),
		"stddev": float64(11),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

// validTimerJSON is a valid dropwizard json document containing one timer
const validTimerJSON = `
{
	"version": 		"3.0.0",
	"counters" : 	{},
	"meters" : 		{},
	"gauges" : 		{},
	"histograms" : 	{},
	"timers" : 		{
		"measurement" : {
			"count" : 1,
			"max" : 2,
			"mean" : 3,
			"min" : 4,
			"p50" : 5,
			"p75" : 6,
			"p95" : 7,
			"p98" : 8,
			"p99" : 9,
			"p999" : 10,
			"stddev" : 11,
			"m15_rate" : 12,
			"m1_rate" : 13,
			"m5_rate" : 14,
			"mean_rate" : 15,
			"duration_units" : "seconds",
			"rate_units" : "calls/second"
		}
	}
}
`

func TestParseValidTimerJSON(t *testing.T) {
	parser := Parser{}

	metrics, err := parser.Parse([]byte(validTimerJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "timer.measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count":     float64(1),
		"max":       float64(2),
		"mean":      float64(3),
		"min":       float64(4),
		"p50":       float64(5),
		"p75":       float64(6),
		"p95":       float64(7),
		"p98":       float64(8),
		"p99":       float64(9),
		"p999":      float64(10),
		"stddev":    float64(11),
		"m15_rate":  float64(12),
		"m1_rate":   float64(13),
		"m5_rate":   float64(14),
		"mean_rate": float64(15),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{}, metrics[0].Tags())
}

// validAllJSON is a valid dropwizard json document containing one metric of each type
const validAllJSON = `
{
	"version": 		"3.0.0",
	"counters" : 	{
		"measurement" : {"count" : 1}
	},
	"meters" : 		{
		"measurement" : {"count" : 1}
	},
	"gauges" : 		{
		"measurement" : {"value" : 1}
	},
	"histograms" : 	{
		"measurement" : {"count" : 1}
	},
	"timers" : 		{
		"measurement" : {"count" : 1}
	}
}
`

func TestParseValidAllJSON(t *testing.T) {
	parser := Parser{}

	metrics, err := parser.Parse([]byte(validAllJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 5)
}

func TestTagParsingProblems(t *testing.T) {
	// giving a wrong path results in empty tags
	parser1 := Parser{MetricRegistryPath: "metrics", TagsPath: "tags1"}
	metrics1, err1 := parser1.Parse([]byte(validEmbeddedCounterJSON))
	assert.NoError(t, err1)
	assert.Len(t, metrics1, 1)
	assert.Equal(t, map[string]string{}, metrics1[0].Tags())

	// giving a wrong TagsPath falls back to TagPathsMap
	parser2 := Parser{
		MetricRegistryPath: "metrics",
		TagsPath:           "tags1",
		TagPathsMap:        map[string]string{"tag1": "tags.tag1"},
	}
	metrics2, err2 := parser2.Parse([]byte(validEmbeddedCounterJSON))
	assert.NoError(t, err2)
	assert.Len(t, metrics2, 1)
	assert.Equal(t, map[string]string{"tag1": "green"}, metrics2[0].Tags())
}
