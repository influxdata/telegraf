package valuecounter

import (
	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// Create a valuecounter with config
func NewTestValueCounter(fields []string) telegraf.Aggregator {
	vc := &ValueCounter{
		Fields: fields,
	}
	vc.Reset()

	return vc
}

var m1, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"status": 200,
		"foobar": "bar",
	},
	time.Now(),
)

var m2, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"status":    "OK",
		"ignoreme":  "string",
		"andme":     true,
		"boolfield": false,
	},
	time.Now(),
)

func BenchmarkApply(b *testing.B) {
	vc := NewTestValueCounter([]string{"status"})

	for n := 0; n < b.N; n++ {
		vc.Add(m1)
		vc.Add(m2)
	}
}

// Test basic functionality
func TestBasic(t *testing.T) {
	vc := NewTestValueCounter([]string{"status"})
	acc := testutil.Accumulator{}

	vc.Add(m1)
	vc.Add(m2)
	vc.Add(m1)
	vc.Push(&acc)

	expectedFields := map[string]interface{}{
		"status_200": 2,
		"status_OK":  1,
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test with multiple fields to count
func TestMultipleFields(t *testing.T) {
	vc := NewTestValueCounter([]string{"status", "somefield", "boolfield"})
	acc := testutil.Accumulator{}

	vc.Add(m1)
	vc.Add(m2)
	vc.Add(m2)
	vc.Add(m1)
	vc.Push(&acc)

	expectedFields := map[string]interface{}{
		"status_200":      2,
		"status_OK":       2,
		"boolfield_false": 2,
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test with a reset between two runs
func TestWithReset(t *testing.T) {
	vc := NewTestValueCounter([]string{"status"})
	acc := testutil.Accumulator{}

	vc.Add(m1)
	vc.Add(m1)
	vc.Add(m2)
	vc.Push(&acc)

	expectedFields := map[string]interface{}{
		"status_200": 2,
		"status_OK":  1,
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)

	acc.ClearMetrics()
	vc.Reset()

	vc.Add(m2)
	vc.Add(m2)
	vc.Add(m1)
	vc.Push(&acc)

	expectedFields = map[string]interface{}{
		"status_200": 1,
		"status_OK":  2,
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

func newMetric(key string, value string) telegraf.Metric {
	var m1, _ = metric.New("m1",
		map[string]string{"foo": "bar"},
		map[string]interface{}{
			key: value,
		},
		time.Now())
	return m1
}

type TestTrackingAccumulator struct {
	telegraf.TrackingAccumulator
	Metrics *[]map[string]interface{}
}

func (t TestTrackingAccumulator) AddFields(name string, fields map[string]interface{}, tags map[string]string, time ...time.Time) {
	*t.Metrics = append(*t.Metrics, fields)
}

func _predicateTest(t *testing.T, configToml []byte, expectedNumberOfPredicates int, metrics []telegraf.Metric) []map[string]interface{} {
	c := config.NewConfig()
	err := c.LoadConfigData(configToml)
	assert.Nil(t, err)
	assert.Len(t, c.Aggregators, 1)
	assert.IsType(t, &ValueCounter{}, c.Aggregators[0].Aggregator)

	valueCounter := c.Aggregators[0].Aggregator.(*ValueCounter)
	assert.Len(t, valueCounter.Predicates, expectedNumberOfPredicates)

	err = valueCounter.Init()
	assert.Nil(t, err)

	for _, metric := range metrics {
		valueCounter.Add(metric)
	}

	var aggregatedMetrics []map[string]interface{}
	valueCounter.Push(TestTrackingAccumulator{Metrics: &aggregatedMetrics})
	assert.Len(t, aggregatedMetrics, 1)
	return aggregatedMetrics
}

func TestGreaterThanPredicate(t *testing.T) {
	toml := `
[[aggregators.valuecounter]]
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The fields for which the values will be counted
  fields = ["app", "path"]
  ## Only include fields for which the predicate is true
  [[aggregators.valuecounter.predicate]]
    type = "greater_than"
    value = 2
`
	metrics := []telegraf.Metric{}
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("path", "/"))
	metrics = append(metrics, newMetric("path", "/"))

	aggregatedMetrics := _predicateTest(t, []byte(toml), 1, metrics)

	assert.Contains(t, aggregatedMetrics[0], "app_trafik")
	assert.Equal(t, 3, aggregatedMetrics[0]["app_trafik"].(int))

	assert.NotContains(t, "path_/", aggregatedMetrics[0])
}

func TestLessThanPredicate(t *testing.T) {
	toml := `
[[aggregators.valuecounter]]
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The fields for which the values will be counted
  fields = ["app", "path"]
  ## Only include fields for which the predicate is true
  [[aggregators.valuecounter.predicate]]
    type = "less_than"
    value = 3
`
	metrics := []telegraf.Metric{}
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("path", "/"))
	metrics = append(metrics, newMetric("path", "/"))

	aggregatedMetrics := _predicateTest(t, []byte(toml), 1, metrics)

	assert.Len(t, aggregatedMetrics, 1)

	assert.Contains(t, aggregatedMetrics[0], "path_/")
	assert.Equal(t, 2, aggregatedMetrics[0]["path_/"].(int))

	assert.NotContains(t, aggregatedMetrics[0], "app_trafik")
}

func TestEqualToPredicate(t *testing.T) {
	toml := `
[[aggregators.valuecounter]]
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The fields for which the values will be counted
  fields = ["app", "path"]
  ## Only include fields for which the predicate is true
  [[aggregators.valuecounter.predicate]]
    type = "equal_to"
    value = 3
`
	metrics := []telegraf.Metric{}
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("path", "/"))
	metrics = append(metrics, newMetric("path", "/"))
	metrics = append(metrics, newMetric("path", "/robots.txt"))
	metrics = append(metrics, newMetric("path", "/robots.txt"))
	metrics = append(metrics, newMetric("path", "/robots.txt"))

	aggregatedMetrics := _predicateTest(t, []byte(toml), 1, metrics)

	assert.Len(t, aggregatedMetrics, 1)

	assert.Contains(t, aggregatedMetrics[0], "app_trafik")
	assert.Equal(t, 3, aggregatedMetrics[0]["app_trafik"].(int))
	assert.Contains(t, aggregatedMetrics[0], "path_/robots.txt")
	assert.Equal(t, 3, aggregatedMetrics[0]["path_/robots.txt"].(int))

	assert.NotContains(t, aggregatedMetrics[0], "path_/")
}

func TestNotEqualToPredicate(t *testing.T) {
	toml := `
[[aggregators.valuecounter]]
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The fields for which the values will be counted
  fields = ["app", "path"]
  ## Only include fields for which the predicate is true
  [[aggregators.valuecounter.predicate]]
    type = "not_equal_to"
    value = 3
`
	metrics := []telegraf.Metric{}
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("path", "/"))
	metrics = append(metrics, newMetric("path", "/"))
	metrics = append(metrics, newMetric("path", "/robots.txt"))
	metrics = append(metrics, newMetric("path", "/robots.txt"))
	metrics = append(metrics, newMetric("path", "/robots.txt"))

	aggregatedMetrics := _predicateTest(t, []byte(toml), 1, metrics)

	assert.Len(t, aggregatedMetrics, 1)

	assert.Contains(t, aggregatedMetrics[0], "path_/")
	assert.Equal(t, 2, aggregatedMetrics[0]["path_/"].(int))

	assert.NotContains(t, aggregatedMetrics[0], "path_/robots.txt")
	assert.NotContains(t, aggregatedMetrics[0], "app_trafik")
}

func TestMultiplePredicates(t *testing.T) {
	toml := `
[[aggregators.valuecounter]]
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The fields for which the values will be counted
  fields = ["app", "path"]
  ## Only include fields for which the predicate is true
  [[aggregators.valuecounter.predicate]]
    type = "less_than"
    value = 10

  [[aggregators.valuecounter.predicate]]
    type = "greater_than"
    value = 1

  [[aggregators.valuecounter.predicate]]
    type = "not_equal_to"
    value = 3

  [[aggregators.valuecounter.predicate]]
    type = "equal_to"
    value = 2`

	metrics := []telegraf.Metric{}
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("app", "trafik"))
	metrics = append(metrics, newMetric("path", "/"))
	metrics = append(metrics, newMetric("path", "/"))
	metrics = append(metrics, newMetric("host", "google"))

	aggregatedMetrics := _predicateTest(t, []byte(toml), 4, metrics)

	assert.Contains(t, aggregatedMetrics[0], "path_/")
	assert.Equal(t, 2, aggregatedMetrics[0]["path_/"].(int))

	assert.NotContains(t, aggregatedMetrics[0], "app_trafik")
	assert.NotContains(t, aggregatedMetrics[0], "host_google")

}
