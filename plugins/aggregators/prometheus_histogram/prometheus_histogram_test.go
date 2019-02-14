package prometheus_histogram_test

import (
	"fmt"
	"github.com/influxdata/telegraf/plugins/aggregators/prometheus_histogram"
	"github.com/prometheus/client_golang/prometheus"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func newTestHistogram(cfg []*prometheus_histogram.Config) telegraf.Aggregator {
	return &prometheus_histogram.PrometheusHistogramAggregator{Configs: cfg, Registry: prometheus.NewRegistry()}
}

func newTestGaugeMetric(name string, value float64) telegraf.Metric {
	metric, _ := metric.New(
		name,
		map[string]string{"tag_name": "tag_value"},
		map[string]interface{}{
			"gauge": value,
		},
		time.Now(),
	)

	return metric
}

// BenchmarkApply runs benchmarks
func BenchmarkApply(b *testing.B) {
	histogram := prometheus_histogram.NewPrometheusHistogramAggregator()

	for n := 0; n < b.N; n++ {
		histogram.Add(newTestGaugeMetric("first", 15.3))
		histogram.Add(newTestGaugeMetric("first", 30.9))
		histogram.Add(newTestGaugeMetric("second", 105))
	}
}

func TestHistogramWithTwoPeriods(t *testing.T) {
	cfg := []*prometheus_histogram.Config{{
		MeasurementName: "metric",
		Unit:            "seconds",
		Buckets:         []float64{0.0, 10.0, 20.0, 30.0, 40.0},
	}}
	histogram := newTestHistogram(cfg)
	acc := &testutil.Accumulator{}

	histogram.Add(newTestGaugeMetric("metric", 15.3))
	histogram.Add(newTestGaugeMetric("metric", 30.9))
	histogram.Push(acc)

	acc.AssertContainsFields(t, "metric_seconds", map[string]interface{}{
		"0":     0.0,
		"10":    0.0,
		"20":    1.0,
		"30":    1.0,
		"40":    2.0,
		"count": 2.0,
		"sum":   46.2,
	})
}

func TestHistogramWithInvalidGaugeField(t *testing.T) {
	cfg := []*prometheus_histogram.Config{{
		MeasurementName: "metric",
		Unit:            "seconds",
		Buckets:         []float64{0.0, 10.0, 20.0, 30.0, 40.0},
	}}
	histogram := newTestHistogram(cfg)
	acc := &testutil.Accumulator{}

	invalidMetric, err := metric.New(
		"metric",
		map[string]string{"tag_name": "tag_value"},
		map[string]interface{}{
			"gauge": "foo bar",
		},
		time.Now(),
	)

	assert.Nil(t, err, "Expected err not to have occurred but was %+v", err)

	histogram.Add(newTestGaugeMetric("metric", 15.3))
	histogram.Add(invalidMetric)
	histogram.Push(acc)

	acc.AssertContainsFields(t, "metric_seconds", map[string]interface{}{
		"0":     0.0,
		"10":    0.0,
		"20":    1.0,
		"30":    1.0,
		"40":    1.0,
		"count": 1.0,
		"sum":   15.3,
	})
}

func TestHistogramWithTwoSeriesWithMultipleMetrics(t *testing.T) {
	cfg := []*prometheus_histogram.Config{
		{MeasurementName: "first", Unit: "seconds", Buckets: []float64{0.0, 15.5, 20.0, 30.0, 40.0}},
		{MeasurementName: "second", Unit: "seconds", Buckets: []float64{0.0, 40.0, 50.0, 100.0, 150.0}},
	}
	histogram := newTestHistogram(cfg)
	acc := &testutil.Accumulator{}

	histogram.Add(newTestGaugeMetric("first", 15.3))
	histogram.Add(newTestGaugeMetric("first", 30.9))
	histogram.Add(newTestGaugeMetric("second", 105))
	histogram.Push(acc)

	acc.AssertContainsFields(t, "first_seconds", map[string]interface{}{
		"0":     0.0,
		"15.5":  1.0,
		"20":    1.0,
		"30":    1.0,
		"40":    2.0,
		"count": 2.0,
		"sum":   46.2,
	})

	acc.AssertContainsFields(t, "second_seconds", map[string]interface{}{
		"0":     0.0,
		"40":    0.0,
		"50":    0.0,
		"100":   0.0,
		"150":   1.0,
		"count": 1.0,
		"sum":   105.0,
	})
}

func TestHistogramDifferentPeriodsAndAllFields(t *testing.T) {
	cfg := []*prometheus_histogram.Config{{
		MeasurementName: "metric",
		Unit:            "seconds",
		Buckets:         []float64{0.0, 10.0, 20.0, 30.0, 40.0},
	}}
	histogram := newTestHistogram(cfg)

	acc := &testutil.Accumulator{}
	histogram.Add(newTestGaugeMetric("metric", 15.3))
	histogram.Push(acc)

	expectedFieldsForFirstMetric := map[string]interface{}{
		"0":     0.0,
		"10":    0.0,
		"20":    1.0,
		"30":    1.0,
		"40":    1.0,
		"count": 1.0,
		"sum":   15.3,
	}
	acc.AssertContainsFields(t, "metric_seconds", expectedFieldsForFirstMetric)

	acc.ClearMetrics()
	histogram.Add(newTestGaugeMetric("metric", 30.9))
	histogram.Push(acc)

	expectedFieldsForFirstMetric = map[string]interface{}{
		"0":     0.0,
		"10":    0.0,
		"20":    1.0,
		"30":    1.0,
		"40":    2.0,
		"count": 2.0,
		"sum":   46.2,
	}
	acc.AssertContainsFields(t, "metric_seconds", expectedFieldsForFirstMetric)
}

func TestWrongBucketsOrder(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(
				t,
				"histogram buckets must be in increasing order: 90.000000 >= 20.000000",
				fmt.Sprint(r),
			)
		}
	}()

	cfg := []*prometheus_histogram.Config{{
		MeasurementName: "metric",
		Buckets:         []float64{0.0, 90.0, 20.0, 30.0, 40.0},
	}}
	histogram := newTestHistogram(cfg)
	histogram.Add(newTestGaugeMetric("metric", 30.9))
}
