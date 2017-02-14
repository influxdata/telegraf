package histogram

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

// NewTestHistogram creates new test histogram aggregation with specified config
func NewTestHistogram(cfg []config) telegraf.Aggregator {
	htm := &HistogramAggregator{Configs: cfg}
	htm.buckets = make(bucketsByMetrics)
	htm.resetCache()

	return htm
}

// firstMetric1 is the first test metric
var firstMetric1, _ = metric.New(
	"first_metric_name",
	map[string]string{"tag_name": "tag_value"},
	map[string]interface{}{
		"a": float64(15.3),
		"b": float64(40),
	},
	time.Now(),
)

// firstMetric1 is the first test metric with other value
var firstMetric2, _ = metric.New(
	"first_metric_name",
	map[string]string{"tag_name": "tag_value"},
	map[string]interface{}{
		"a": float64(15.9),
	},
	time.Now(),
)

// secondMetric is the second metric
var secondMetric, _ = metric.New(
	"second_metric_name",
	map[string]string{"tag_name": "tag_value"},
	map[string]interface{}{
		"a":        float64(105),
		"ignoreme": "string",
		"andme":    true,
	},
	time.Now(),
)

// BenchmarkApply runs benchmarks
func BenchmarkApply(b *testing.B) {
	histogram := NewHistogramAggregator()

	for n := 0; n < b.N; n++ {
		histogram.Add(firstMetric1)
		histogram.Add(firstMetric2)
		histogram.Add(secondMetric)
	}
}

// TestHistogramWithPeriodAndOneField tests metrics for one period and for one field
func TestHistogramWithPeriodAndOneField(t *testing.T) {
	var cfg []config
	cfg = append(cfg, config{Metric: "first_metric_name", Fields: []string{"a"}, Buckets: []float64{0.0, 10.0, 20.0, 30.0, 40.0}})
	histogram := NewTestHistogram(cfg)

	acc := &testutil.Accumulator{}

	histogram.Add(firstMetric1)
	histogram.Add(firstMetric2)
	histogram.Push(acc)

	if len(acc.Metrics) != 6 {
		assert.Fail(t, "Incorrect number of metrics")
	}

	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(0), "0.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(0), "10.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "20.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "30.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "40.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), bucketInf)
}

// TestHistogramWithPeriodAndAllFields tests two metrics for one period and for all fields
func TestHistogramWithPeriodAndAllFields(t *testing.T) {
	var cfg []config
	cfg = append(cfg, config{Metric: "first_metric_name", Buckets: []float64{0.0, 15.5, 20.0, 30.0, 40.0}})
	cfg = append(cfg, config{Metric: "second_metric_name", Buckets: []float64{0.0, 4.0, 10.0, 23.0, 30.0}})
	histogram := NewTestHistogram(cfg)

	acc := &testutil.Accumulator{}

	histogram.Add(firstMetric1)
	histogram.Add(firstMetric2)
	histogram.Add(secondMetric)
	histogram.Push(acc)

	if len(acc.Metrics) != 30 {
		assert.Fail(t, "Incorrect number of metrics")
	}

	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(0), "0.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(1), "15.5")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "20.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "30.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "40.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), bucketInf)

	assertContainsTaggedField(t, acc, "first_metric_name", "b_bucket", uint64(0), "0.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "b_bucket", uint64(0), "15.5")
	assertContainsTaggedField(t, acc, "first_metric_name", "b_bucket", uint64(0), "20.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "b_bucket", uint64(0), "30.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "b_bucket", uint64(1), "40.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "b_bucket", uint64(1), bucketInf)

	assertContainsTaggedField(t, acc, "second_metric_name", "a_bucket", uint64(0), "0.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "a_bucket", uint64(0), "4.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "a_bucket", uint64(0), "10.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "a_bucket", uint64(0), "23.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "a_bucket", uint64(0), "30.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "a_bucket", uint64(1), bucketInf)

	assertContainsTaggedField(t, acc, "second_metric_name", "ignoreme_bucket", uint64(0), "0.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "ignoreme_bucket", uint64(0), "4.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "ignoreme_bucket", uint64(0), "10.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "ignoreme_bucket", uint64(0), "23.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "ignoreme_bucket", uint64(0), "30.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "ignoreme_bucket", uint64(0), bucketInf)

	assertContainsTaggedField(t, acc, "second_metric_name", "andme_bucket", uint64(0), "0.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "andme_bucket", uint64(0), "4.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "andme_bucket", uint64(0), "10.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "andme_bucket", uint64(0), "23.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "andme_bucket", uint64(0), "30.0")
	assertContainsTaggedField(t, acc, "second_metric_name", "andme_bucket", uint64(0), bucketInf)
}

// TestHistogramDifferentPeriodsAndAllFields tests two metrics getting added with a push/reset in between (simulates
// getting added in different periods) for all fields
func TestHistogramDifferentPeriodsAndAllFields(t *testing.T) {

	var cfg []config
	cfg = append(cfg, config{Metric: "first_metric_name", Buckets: []float64{0.0, 10.0, 20.0, 30.0, 40.0}})
	histogram := NewTestHistogram(cfg)

	acc := &testutil.Accumulator{}
	histogram.Add(firstMetric1)
	histogram.Push(acc)

	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(0), "0.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(0), "10.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(1), "20.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(1), "30.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(1), "40.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(1), bucketInf)

	acc.ClearMetrics()
	histogram.Add(firstMetric2)
	histogram.Push(acc)

	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(0), "0.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(0), "10.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "20.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "30.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), "40.0")
	assertContainsTaggedField(t, acc, "first_metric_name", "a_bucket", uint64(2), bucketInf)
}

// TestWrongBucketsOrder tests the calling panic with incorrect order of buckets
func TestWrongBucketsOrder(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(
				t,
				"histogram buckets must be in increasing order: 90.00 >= 20.00, metrics: first_metric_name, field: a",
				fmt.Sprint(r),
			)
		}
	}()

	var cfg []config
	cfg = append(cfg, config{Metric: "first_metric_name", Buckets: []float64{0.0, 90.0, 20.0, 30.0, 40.0}})
	histogram := NewTestHistogram(cfg)
	histogram.Add(firstMetric2)
}

// assertContainsTaggedField is help functions to test histogram data
func assertContainsTaggedField(t *testing.T, acc *testutil.Accumulator, metricName string, field string, counts uint64, le string) {
	expectedFields := map[string]interface{}{}
	expectedFields[field] = counts

	acc.Lock()
	defer acc.Unlock()

	for _, metric := range acc.Metrics {
		if metric.Measurement != metricName {
			continue
		}

		if _, ok := metric.Fields[field]; !ok {
			continue
		}

		if metric.Tags[bucketTag] == le {
			if assert.Equal(t, expectedFields, metric.Fields) {
				return
			}

			assert.Fail(t, fmt.Sprintf("incorrect fields %v of metric %s", expectedFields, metricName))
		}
	}

	assert.Fail(t, fmt.Sprintf("unknown measurement %s with tags %v", metricName, []string{"tag_name", "le"}))
}
