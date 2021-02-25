package valuecounter

import (
	"fmt"
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

func TestGreaterThanPredicate(t *testing.T) {
	c := config.NewConfig()
	err := c.LoadConfigData([]byte(`
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
`))
	assert.Nil(t, err)
	fmt.Print(c.Aggregators[0].Aggregator.Description())

	assert.Len(t, c.Aggregators, 1)
	assert.IsType(t, &ValueCounter{}, c.Aggregators[0].Aggregator)

	valueCounter := c.Aggregators[0].Aggregator.(*ValueCounter)
	assert.Len(t, valueCounter.Predicates, 1)

	err = valueCounter.Init()
	assert.Nil(t, err)

	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("path", "/"))
	valueCounter.Add(newMetric("path", "/"))

	var aggregatedMetrics []map[string]interface{}

	valueCounter.Push(TestTrackingAccumulator{Metrics: &aggregatedMetrics})

	assert.Len(t, aggregatedMetrics, 1)

	assert.Contains(t, aggregatedMetrics[0], "app_trafik")
	assert.Equal(t, 3, aggregatedMetrics[0]["app_trafik"].(int))

	assert.NotContains(t, "path_/", aggregatedMetrics[0])
}

func TestLessThanPredicate(t *testing.T) {
	c := config.NewConfig()
	err := c.LoadConfigData([]byte(`
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
`))
	assert.Nil(t, err)
	fmt.Print(c.Aggregators[0].Aggregator.Description())

	assert.Len(t, c.Aggregators, 1)
	assert.IsType(t, &ValueCounter{}, c.Aggregators[0].Aggregator)

	valueCounter := c.Aggregators[0].Aggregator.(*ValueCounter)
	assert.Len(t, valueCounter.Predicates, 1)

	err = valueCounter.Init()
	assert.Nil(t, err)

	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("path", "/"))
	valueCounter.Add(newMetric("path", "/"))

	var aggregatedMetrics []map[string]interface{}

	valueCounter.Push(TestTrackingAccumulator{Metrics: &aggregatedMetrics})

	assert.Len(t, aggregatedMetrics, 1)

	assert.Contains(t, aggregatedMetrics[0], "path_/")
	assert.Equal(t, 2, aggregatedMetrics[0]["path_/"].(int))

	assert.NotContains(t, aggregatedMetrics[0], "app_trafik")
}

func TestMultiplePredicates(t *testing.T) {
	c := config.NewConfig()
	err := c.LoadConfigData([]byte(`
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

  [[aggregators.valuecounter.predicate]]
    type = "greater_than"
    value = 1
`))
	assert.Nil(t, err)
	fmt.Print(c.Aggregators[0].Aggregator.Description())

	assert.Len(t, c.Aggregators, 1)
	assert.IsType(t, &ValueCounter{}, c.Aggregators[0].Aggregator)

	valueCounter := c.Aggregators[0].Aggregator.(*ValueCounter)
	assert.Len(t, valueCounter.Predicates, 2)

	err = valueCounter.Init()
	assert.Nil(t, err)

	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("app", "trafik"))
	valueCounter.Add(newMetric("path", "/"))
	valueCounter.Add(newMetric("path", "/"))
	valueCounter.Add(newMetric("host", "google"))

	var aggregatedMetrics []map[string]interface{}

	valueCounter.Push(TestTrackingAccumulator{Metrics: &aggregatedMetrics})

	assert.Len(t, aggregatedMetrics, 1)

	assert.Contains(t, aggregatedMetrics[0], "path_/")
	assert.Equal(t, 2, aggregatedMetrics[0]["path_/"].(int))

	assert.NotContains(t, aggregatedMetrics[0], "app_trafik")
	assert.NotContains(t, aggregatedMetrics[0], "host_google")

}
