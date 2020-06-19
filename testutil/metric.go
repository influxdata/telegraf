package testutil

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func lessFunc(lhs, rhs *metricDiff) bool {
	if lhs.Measurement != rhs.Measurement {
		return lhs.Measurement < rhs.Measurement
	}

	for i := 0; ; i++ {
		if i >= len(lhs.Tags) && i >= len(rhs.Tags) {
			break
		} else if i >= len(lhs.Tags) {
			return true
		} else if i >= len(rhs.Tags) {
			return false
		}

		if lhs.Tags[i].Key != rhs.Tags[i].Key {
			return lhs.Tags[i].Key < rhs.Tags[i].Key
		}
		if lhs.Tags[i].Value != rhs.Tags[i].Value {
			return lhs.Tags[i].Value < rhs.Tags[i].Value
		}
	}

	for i := 0; ; i++ {
		if i >= len(lhs.Fields) && i >= len(rhs.Fields) {
			break
		} else if i >= len(lhs.Fields) {
			return true
		} else if i >= len(rhs.Fields) {
			return false
		}

		if lhs.Fields[i].Key != rhs.Fields[i].Key {
			return lhs.Fields[i].Key < rhs.Fields[i].Key
		}

		if lhs.Fields[i].Value != rhs.Fields[i].Value {
			ltype := reflect.TypeOf(lhs.Fields[i].Value)
			rtype := reflect.TypeOf(lhs.Fields[i].Value)

			if ltype.Kind() != rtype.Kind() {
				return ltype.Kind() < rtype.Kind()
			}

			switch v := lhs.Fields[i].Value.(type) {
			case int64:
				return v < lhs.Fields[i].Value.(int64)
			case uint64:
				return v < lhs.Fields[i].Value.(uint64)
			case float64:
				return v < lhs.Fields[i].Value.(float64)
			case string:
				return v < lhs.Fields[i].Value.(string)
			case bool:
				return !v
			default:
				panic("unknown type")
			}
		}
	}

	if lhs.Type != rhs.Type {
		return lhs.Type < rhs.Type
	}

	if lhs.Time.UnixNano() != rhs.Time.UnixNano() {
		return lhs.Time.UnixNano() < rhs.Time.UnixNano()
	}

	return false
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

// SortMetrics enables sorting metrics before comparison.
func SortMetrics() cmp.Option {
	return cmpopts.SortSlices(lessFunc)
}

// IgnoreTime disables comparison of timestamp.
func IgnoreTime() cmp.Option {
	return cmpopts.IgnoreFields(metricDiff{}, "Time")
}

// MetricEqual returns true if the metrics are equal.
func MetricEqual(expected, actual telegraf.Metric, opts ...cmp.Option) bool {
	var lhs, rhs *metricDiff
	if expected != nil {
		lhs = newMetricDiff(expected)
	}
	if actual != nil {
		rhs = newMetricDiff(actual)
	}

	opts = append(opts, cmpopts.EquateNaNs())
	return cmp.Equal(lhs, rhs, opts...)
}

// RequireMetricEqual halts the test with an error if the metrics are not
// equal.
func RequireMetricEqual(t *testing.T, expected, actual telegraf.Metric, opts ...cmp.Option) {
	t.Helper()

	var lhs, rhs *metricDiff
	if expected != nil {
		lhs = newMetricDiff(expected)
	}
	if actual != nil {
		rhs = newMetricDiff(actual)
	}

	opts = append(opts, cmpopts.EquateNaNs())
	if diff := cmp.Diff(lhs, rhs, opts...); diff != "" {
		t.Fatalf("telegraf.Metric\n--- expected\n+++ actual\n%s", diff)
	}
}

// RequireMetricsEqual halts the test with an error if the array of metrics
// are not equal.
func RequireMetricsEqual(t *testing.T, expected, actual []telegraf.Metric, opts ...cmp.Option) {
	t.Helper()

	lhs := make([]*metricDiff, 0, len(expected))
	for _, m := range expected {
		lhs = append(lhs, newMetricDiff(m))
	}
	rhs := make([]*metricDiff, 0, len(actual))
	for _, m := range actual {
		rhs = append(rhs, newMetricDiff(m))
	}

	opts = append(opts, cmpopts.EquateNaNs())
	if diff := cmp.Diff(lhs, rhs, opts...); diff != "" {
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
	m, err := metric.New(met.Measurement, met.Tags, met.Fields, met.Time, met.Type)
	if err != nil {
		panic("MustMetric")
	}
	return m
}
