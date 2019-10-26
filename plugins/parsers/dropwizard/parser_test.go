package dropwizard

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var TimeFunc = func() time.Time {
	return time.Unix(0, 0)
}

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
	parser := NewParser()

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
	parser := NewParser()

	metrics, err := parser.Parse([]byte(validCounterJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"metric_type": "counter"}, metrics[0].Tags())
}

// validEmbeddedCounterJSON is a valid json document containing separate fields for dropwizard metrics, tags and time override.
const validEmbeddedCounterJSON = `
{
	"time" : "2017-02-22T14:33:03.662+02:00",
	"tags" : {
		"tag1" : "green",
		"tag2" : "yellow",
		"tag3 space,comma=equals" : "red ,="
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
	parser := NewParser()
	parser.MetricRegistryPath = "metrics"
	parser.TagsPath = "tags"
	parser.TimePath = "time"

	metrics, err := parser.Parse([]byte(validEmbeddedCounterJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count": float64(1),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"metric_type":             "counter",
		"tag1":                    "green",
		"tag2":                    "yellow",
		"tag3 space,comma=equals": "red ,=",
	}, metrics[0].Tags())
	assert.True(t, metricTime.Equal(metrics[0].Time()), fmt.Sprintf("%s should be equal to %s", metrics[0].Time(), metricTime))

	// now test json tags through TagPathsMap
	parser2 := NewParser()
	parser2.MetricRegistryPath = "metrics"
	parser2.TagPathsMap = map[string]string{"tag1": "tags.tag1"}
	parser2.TimePath = "time"
	metrics2, err2 := parser2.Parse([]byte(validEmbeddedCounterJSON))
	assert.NoError(t, err2)
	assert.Equal(t, map[string]string{"metric_type": "counter", "tag1": "green"}, metrics2[0].Tags())
}

// validMeterJSON1 is a valid dropwizard json document containing one meter
const validMeterJSON1 = `
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
		}
	},
	"gauges" : 		{},
	"histograms" : 	{},
	"timers" : 		{}
}
`

func TestParseValidMeterJSON1(t *testing.T) {
	parser := NewParser()

	metrics, err := parser.Parse([]byte(validMeterJSON1))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "measurement1", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count":     float64(1),
		"m15_rate":  float64(1),
		"m1_rate":   float64(1),
		"m5_rate":   float64(1),
		"mean_rate": float64(1),
		"units":     "events/second",
	}, metrics[0].Fields())

	assert.Equal(t, map[string]string{"metric_type": "meter"}, metrics[0].Tags())
}

// validMeterJSON2 is a valid dropwizard json document containing one meter with one tag
const validMeterJSON2 = `
{
	"version": 		"3.0.0",
	"counters" : 	{},
	"meters" : 		{ 
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

func TestParseValidMeterJSON2(t *testing.T) {
	parser := NewParser()

	metrics, err := parser.Parse([]byte(validMeterJSON2))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "measurement2", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count":     float64(2),
		"m15_rate":  float64(2),
		"m1_rate":   float64(2),
		"m5_rate":   float64(2),
		"mean_rate": float64(2),
		"units":     "events/second",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"metric_type": "meter", "key": "value"}, metrics[0].Tags())
}

// validGaugeJSON is a valid dropwizard json document containing one gauge
const validGaugeJSON = `
{
	"version": 		"3.0.0",
	"counters" : 	{},
	"meters" : 		{},
	"gauges" : 		{
		"measurement" : {
			"value" : true
		}
	},
	"histograms" : 	{},
	"timers" : 		{}
}
`

func TestParseValidGaugeJSON(t *testing.T) {
	parser := NewParser()

	metrics, err := parser.Parse([]byte(validGaugeJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": true,
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"metric_type": "gauge"}, metrics[0].Tags())
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
	parser := NewParser()

	metrics, err := parser.Parse([]byte(validHistogramJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "measurement", metrics[0].Name())
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
	assert.Equal(t, map[string]string{"metric_type": "histogram"}, metrics[0].Tags())
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
	parser := NewParser()

	metrics, err := parser.Parse([]byte(validTimerJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "measurement", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"count":          float64(1),
		"max":            float64(2),
		"mean":           float64(3),
		"min":            float64(4),
		"p50":            float64(5),
		"p75":            float64(6),
		"p95":            float64(7),
		"p98":            float64(8),
		"p99":            float64(9),
		"p999":           float64(10),
		"stddev":         float64(11),
		"m15_rate":       float64(12),
		"m1_rate":        float64(13),
		"m5_rate":        float64(14),
		"mean_rate":      float64(15),
		"duration_units": "seconds",
		"rate_units":     "calls/second",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"metric_type": "timer"}, metrics[0].Tags())
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
	parser := NewParser()

	metrics, err := parser.Parse([]byte(validAllJSON))
	assert.NoError(t, err)
	assert.Len(t, metrics, 5)
}

func TestTagParsingProblems(t *testing.T) {
	// giving a wrong path results in empty tags
	parser1 := NewParser()
	parser1.MetricRegistryPath = "metrics"
	parser1.TagsPath = "tags1"
	metrics1, err1 := parser1.Parse([]byte(validEmbeddedCounterJSON))
	assert.NoError(t, err1)
	assert.Len(t, metrics1, 1)
	assert.Equal(t, map[string]string{"metric_type": "counter"}, metrics1[0].Tags())

	// giving a wrong TagsPath falls back to TagPathsMap
	parser2 := NewParser()
	parser2.MetricRegistryPath = "metrics"
	parser2.TagsPath = "tags1"
	parser2.TagPathsMap = map[string]string{"tag1": "tags.tag1"}
	metrics2, err2 := parser2.Parse([]byte(validEmbeddedCounterJSON))
	assert.NoError(t, err2)
	assert.Len(t, metrics2, 1)
	assert.Equal(t, map[string]string{"metric_type": "counter", "tag1": "green"}, metrics2[0].Tags())
}

// sampleTemplateJSON is a sample json document containing metrics to be tested against the templating engine.
const sampleTemplateJSON = `
{
	"version": 		"3.0.0",
	"counters" :	{},
	"meters" :		{},
	"gauges" :		{
		"vm.memory.heap.committed" 		: { "value" : 1 },
		"vm.memory.heap.init" 			: { "value" : 2 },
		"vm.memory.heap.max" 			: { "value" : 3 },
		"vm.memory.heap.usage" 			: { "value" : 4 },
		"vm.memory.heap.used" 			: { "value" : 5 },
		"vm.memory.non-heap.committed" 	: { "value" : 6 },
		"vm.memory.non-heap.init" 		: { "value" : 7 },
		"vm.memory.non-heap.max" 		: { "value" : 8 },
		"vm.memory.non-heap.usage" 		: { "value" : 9 },
		"vm.memory.non-heap.used" 		: { "value" : 10 }
	},
	"histograms" :	{
		"jenkins.job.building.duration" : {
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
	"timers" :		{}
}
`

func TestParseSampleTemplateJSON(t *testing.T) {
	parser := NewParser()
	err := parser.SetTemplates("_", []string{
		"jenkins.* measurement.metric.metric.field",
		"vm.* measurement.measurement.pool.field",
	})
	require.NoError(t, err)

	metrics, err := parser.Parse([]byte(sampleTemplateJSON))
	require.NoError(t, err)

	require.Len(t, metrics, 11)

	jenkinsMetric := search(metrics, "jenkins", nil, "")
	require.NotNil(t, jenkinsMetric, "the metrics should contain a jenkins measurement")
	require.Equal(t, map[string]interface{}{
		"duration_count":  float64(1),
		"duration_max":    float64(2),
		"duration_mean":   float64(3),
		"duration_min":    float64(4),
		"duration_p50":    float64(5),
		"duration_p75":    float64(6),
		"duration_p95":    float64(7),
		"duration_p98":    float64(8),
		"duration_p99":    float64(9),
		"duration_p999":   float64(10),
		"duration_stddev": float64(11),
	}, jenkinsMetric.Fields())
	require.Equal(t, map[string]string{"metric_type": "histogram", "metric": "job_building"}, jenkinsMetric.Tags())

	vmMemoryHeapCommitted := search(metrics, "vm_memory", map[string]string{"pool": "heap"}, "committed_value")
	require.NotNil(t, vmMemoryHeapCommitted)
	require.Equal(t, map[string]interface{}{
		"committed_value": float64(1),
	}, vmMemoryHeapCommitted.Fields())
	require.Equal(t, map[string]string{"metric_type": "gauge", "pool": "heap"}, vmMemoryHeapCommitted.Tags())

	vmMemoryNonHeapCommitted := search(metrics, "vm_memory", map[string]string{"pool": "non-heap"}, "committed_value")
	require.NotNil(t, vmMemoryNonHeapCommitted)
	require.Equal(t, map[string]interface{}{
		"committed_value": float64(6),
	}, vmMemoryNonHeapCommitted.Fields())
	require.Equal(t, map[string]string{"metric_type": "gauge", "pool": "non-heap"}, vmMemoryNonHeapCommitted.Tags())
}

func search(metrics []telegraf.Metric, name string, tags map[string]string, fieldName string) telegraf.Metric {
	for _, v := range metrics {
		if v.Name() == name && containsAll(v.Tags(), tags) {
			if len(fieldName) == 0 {
				return v
			}
			if _, ok := v.Fields()[fieldName]; ok {
				return v
			}
		}
	}
	return nil
}

func containsAll(t1 map[string]string, t2 map[string]string) bool {
	for k, v := range t2 {
		if foundValue, ok := t1[k]; !ok || v != foundValue {
			return false
		}
	}
	return true
}

func Metric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

func NoError(t *testing.T, err error) {
	require.NoError(t, err)
}

func TestDropWizard(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		metrics []telegraf.Metric
		errFunc func(t *testing.T, err error)
	}{
		{
			name:  "minimal",
			input: []byte(`{"version": "3.0.0", "counters": {"cpu": {"value": 42}}}`),
			metrics: []telegraf.Metric{
				Metric(
					metric.New(
						"cpu",
						map[string]string{
							"metric_type": "counter",
						},
						map[string]interface{}{
							"value": 42.0,
						},
						TimeFunc(),
					),
				),
			},
			errFunc: NoError,
		},
		{
			name:  "name with space unescaped",
			input: []byte(`{"version": "3.0.0", "counters": {"hello world": {"value": 42}}}`),
			metrics: []telegraf.Metric{
				Metric(
					metric.New(
						"hello world",
						map[string]string{
							"metric_type": "counter",
						},
						map[string]interface{}{
							"value": 42.0,
						},
						TimeFunc(),
					),
				),
			},
			errFunc: NoError,
		},
		{
			name:  "name with space single slash escaped is not valid JSON",
			input: []byte(`{"version": "3.0.0", "counters": {"hello\ world": {"value": 42}}}`),
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:  "name with space double slash escape",
			input: []byte(`{"version": "3.0.0", "counters": {"hello\\ world": {"value": 42}}}`),
			metrics: []telegraf.Metric{
				Metric(
					metric.New(
						"hello world",
						map[string]string{
							"metric_type": "counter",
						},
						map[string]interface{}{
							"value": 42.0,
						},
						TimeFunc(),
					),
				),
			},
			errFunc: NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			parser.SetTimeFunc(TimeFunc)
			metrics, err := parser.Parse(tt.input)
			tt.errFunc(t, err)

			require.Equal(t, len(tt.metrics), len(metrics))
			for i, expected := range tt.metrics {
				require.Equal(t, expected.Name(), metrics[i].Name())
				require.Equal(t, expected.Tags(), metrics[i].Tags())
				require.Equal(t, expected.Fields(), metrics[i].Fields())
				require.Equal(t, expected.Time(), metrics[i].Time())
			}
		})
	}
}
