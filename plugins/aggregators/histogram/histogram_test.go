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
		"c": float64(40),
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
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(0)}, "0")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(0)}, "10")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2)}, "20")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2)}, "30")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2)}, "40")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2)}, bucketInf)
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

	if len(acc.Metrics) != 12 {
		assert.Fail(t, "Incorrect number of metrics")
	}

	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(0), "b_bucket": int64(0), "c_bucket": int64(0)}, "0")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(1), "b_bucket": int64(0), "c_bucket": int64(0)}, "15.5")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2), "b_bucket": int64(0), "c_bucket": int64(0)}, "20")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2), "b_bucket": int64(0), "c_bucket": int64(0)}, "30")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2), "b_bucket": int64(1), "c_bucket": int64(1)}, "40")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2), "b_bucket": int64(1), "c_bucket": int64(1)}, bucketInf)

	assertContainsTaggedField(t, acc, "second_metric_name", map[string]interface{}{"a_bucket": int64(0), "ignoreme_bucket": int64(0), "andme_bucket": int64(0)}, "0")
	assertContainsTaggedField(t, acc, "second_metric_name", map[string]interface{}{"a_bucket": int64(0), "ignoreme_bucket": int64(0), "andme_bucket": int64(0)}, "4")
	assertContainsTaggedField(t, acc, "second_metric_name", map[string]interface{}{"a_bucket": int64(0), "ignoreme_bucket": int64(0), "andme_bucket": int64(0)}, "10")
	assertContainsTaggedField(t, acc, "second_metric_name", map[string]interface{}{"a_bucket": int64(0), "ignoreme_bucket": int64(0), "andme_bucket": int64(0)}, "23")
	assertContainsTaggedField(t, acc, "second_metric_name", map[string]interface{}{"a_bucket": int64(0), "ignoreme_bucket": int64(0), "andme_bucket": int64(0)}, "30")
	assertContainsTaggedField(t, acc, "second_metric_name", map[string]interface{}{"a_bucket": int64(1), "ignoreme_bucket": int64(0), "andme_bucket": int64(0)}, bucketInf)
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

	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(0), "b_bucket": int64(0)}, "0")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(0), "b_bucket": int64(0)}, "10")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(1), "b_bucket": int64(0)}, "20")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(1), "b_bucket": int64(0)}, "30")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(1), "b_bucket": int64(1)}, "40")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(1), "b_bucket": int64(1)}, bucketInf)

	acc.ClearMetrics()
	histogram.Add(firstMetric2)
	histogram.Push(acc)

	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(0), "b_bucket": int64(0), "c_bucket": int64(0)}, "0")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(0), "b_bucket": int64(0), "c_bucket": int64(0)}, "10")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2), "b_bucket": int64(0), "c_bucket": int64(0)}, "20")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2), "b_bucket": int64(0), "c_bucket": int64(0)}, "30")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2), "b_bucket": int64(1), "c_bucket": int64(1)}, "40")
	assertContainsTaggedField(t, acc, "first_metric_name", map[string]interface{}{"a_bucket": int64(2), "b_bucket": int64(1), "c_bucket": int64(1)}, bucketInf)
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
func assertContainsTaggedField(t *testing.T, acc *testutil.Accumulator, metricName string, fields map[string]interface{}, le string) {
	acc.Lock()
	defer acc.Unlock()

	for _, checkedMetric := range acc.Metrics {
		// check metric name
		if checkedMetric.Measurement != metricName {
			continue
		}

		// check "le" tag
		if checkedMetric.Tags[bucketTag] != le {
			continue
		}

		// check fields
		isFieldsIdentical := true
		for field := range fields {
			if _, ok := checkedMetric.Fields[field]; !ok {
				isFieldsIdentical = false
				break
			}
		}
		if !isFieldsIdentical {
			continue
		}

		// check fields with their counts
		if assert.Equal(t, fields, checkedMetric.Fields) {
			return
		}

		assert.Fail(t, fmt.Sprintf("incorrect fields %v of metric %s", fields, metricName))
	}

	assert.Fail(t, fmt.Sprintf("unknown measurement '%s' with tags: %v, fields: %v", metricName, map[string]string{"le": le}, fields))
}
