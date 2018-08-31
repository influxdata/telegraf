package testutil

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

// MustEqual requires a and b to be identical.
func MustEqual(t *testing.T, got telegraf.Metric, want Metric) {
	require.Equal(t, want.Measurement, got.Name())
	require.Equal(t, want.Fields, got.Fields())
	require.Equal(t, want.Tags, got.Tags())
	require.Equal(t, want.Time, got.Time())
}

// MustMetric creates a new metric or panics on error.
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
