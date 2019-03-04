package testutil

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type metricDiff struct {
	Measurement string
	Tags        []*telegraf.Tag
	Fields      []*telegraf.Field
	Type        telegraf.ValueType
	Time        time.Time
}

func newMetricDiff(metric telegraf.Metric) *metricDiff {
	if metric == nil {
		return nil
	}

	m := &metricDiff{}
	m.Measurement = metric.Name()

	for _, tag := range metric.TagList() {
		m.Tags = append(m.Tags, tag)
	}
	sort.Slice(m.Tags, func(i, j int) bool {
		return m.Tags[i].Key < m.Tags[j].Key
	})

	for _, field := range metric.FieldList() {
		m.Fields = append(m.Fields, field)
	}
	sort.Slice(m.Fields, func(i, j int) bool {
		return m.Fields[i].Key < m.Fields[j].Key
	})

	m.Type = metric.Type()
	m.Time = metric.Time()
	return m
}

func MetricEqual(expected, actual telegraf.Metric) bool {
	var lhs, rhs *metricDiff
	if expected != nil {
		lhs = newMetricDiff(expected)
	}
	if actual != nil {
		rhs = newMetricDiff(actual)
	}

	return cmp.Equal(lhs, rhs)
}

func RequireMetricEqual(t *testing.T, expected, actual telegraf.Metric) {
	t.Helper()

	var lhs, rhs *metricDiff
	if expected != nil {
		lhs = newMetricDiff(expected)
	}
	if actual != nil {
		rhs = newMetricDiff(actual)
	}

	if diff := cmp.Diff(lhs, rhs); diff != "" {
		t.Fatalf("telegraf.Metric\n--- expected\n+++ actual\n%s", diff)
	}
}

func RequireMetricsEqual(t *testing.T, expected, actual []telegraf.Metric) {
	t.Helper()

	lhs := make([]*metricDiff, 0, len(expected))
	for _, m := range expected {
		lhs = append(lhs, newMetricDiff(m))
	}
	rhs := make([]*metricDiff, 0, len(actual))
	for _, m := range actual {
		rhs = append(rhs, newMetricDiff(m))
	}
	if diff := cmp.Diff(lhs, rhs); diff != "" {
		t.Fatalf("[]telegraf.Metric\n--- expected\n+++ actual\n%s", diff)
	}
}

// Metric creates a new metric or panics on error.
func MustMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	tm time.Time,
	tp ...telegraf.ValueType,
) telegraf.Metric {
	m, err := metric.New(name, tags, fields, tm, tp...)
	if err != nil {
		panic("MustMetric")
	}
	return m
}

func FromTestMetric(met *Metric) telegraf.Metric {
	m, err := metric.New(met.Measurement, met.Tags, met.Fields, met.Time)
	if err != nil {
		panic("MustMetric")
	}
	return m
}
