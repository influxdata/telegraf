package selfstat

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestCollectorRegisterIncrSet(t *testing.T) {
	defer maps.Clear(registry.stats)

	// Create a new collector with global tags
	collector := NewCollector(map[string]string{"global": "zoo"})

	// Register two statistics
	fields := []string{"field1", "field2"}
	tags := map[string]string{"test": "foo"}
	for _, f := range fields {
		collector.Register("test", f, tags)
	}
	require.Len(t, collector.statistics, len(fields))

	// Check for initial state of the registry
	require.Len(t, registry.stats, 1)
	stats := maps.Values(registry.stats)[0]
	require.Len(t, stats, len(fields))
	require.ElementsMatch(t, fields, maps.Keys(stats))

	// Check initial stats values
	for _, f := range fields {
		require.Equalf(t, int64(0), collector.Get("test", f, tags).Get(), "field %q has wrong value", f)
	}

	// Increase the statistics
	collector.Get("test", "field1", tags).Incr(10)
	collector.Get("test", "field2", tags).Incr(5)

	// Check the values
	expected := []telegraf.Metric{
		metric.New(
			"internal_test",
			map[string]string{"global": "zoo", "test": "foo"},
			map[string]interface{}{"field1": int64(10), "field2": int64(5)},
			time.Unix(0, 0),
		),
	}
	options := []cmp.Option{testutil.IgnoreTime(), testutil.SortMetrics()}
	testutil.RequireMetricsEqual(t, expected, Metrics(), options...)

	// Make sure that re-registering a field does not override the values
	collector.Register("test", "field1", tags)
	testutil.RequireMetricsEqual(t, expected, Metrics(), options...)

	// Make sure that registering with different tags creates a new metric
	collector.Register("test", "field1", map[string]string{"test": "bar"})
	collector.Get("test", "field1", map[string]string{"test": "bar"}).Set(42)

	expected = []telegraf.Metric{
		metric.New(
			"internal_test",
			map[string]string{"global": "zoo", "test": "foo"},
			map[string]interface{}{"field1": int64(10), "field2": int64(5)},
			time.Unix(0, 0),
		),
		metric.New(
			"internal_test",
			map[string]string{"global": "zoo", "test": "bar"},
			map[string]interface{}{"field1": int64(42)},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, Metrics(), options...)
}

func TestCollectorRegisterTimingIncrSet(t *testing.T) {
	defer maps.Clear(registry.stats)

	// Create a new collector with global tags
	collector := NewCollector(map[string]string{"global": "zoo"})

	// Register two statistics
	fields := []string{"field1_ns", "field2_ns"}
	tags := map[string]string{"test": "foo"}
	for _, f := range fields {
		collector.RegisterTiming("test", f, tags)
	}
	require.Len(t, collector.statistics, len(fields))

	// Check for initial state of the registry
	require.Len(t, registry.stats, 1)
	stats := maps.Values(registry.stats)[0]
	require.Len(t, stats, len(fields))
	require.ElementsMatch(t, fields, maps.Keys(stats))

	// Check initial stats values
	for _, f := range fields {
		require.Equalf(t, int64(0), collector.Get("test", f, tags).Get(), "field %q has wrong value", f)
	}

	// Increase the statistics
	collector.Get("test", "field1_ns", tags).Incr(10)
	collector.Get("test", "field2_ns", tags).Incr(5)

	// Check the values
	expected := []telegraf.Metric{
		metric.New(
			"internal_test",
			map[string]string{"global": "zoo", "test": "foo"},
			map[string]interface{}{"field1_ns": int64(10), "field2_ns": int64(5)},
			time.Unix(0, 0),
		),
	}
	options := []cmp.Option{testutil.IgnoreTime(), testutil.SortMetrics()}
	testutil.RequireMetricsEqual(t, expected, Metrics(), options...)

	// Make sure that re-registering a field does not override the values
	collector.Register("test", "field1_ns", tags)
	testutil.RequireMetricsEqual(t, expected, Metrics(), options...)

	// Make sure that registering with different tags creates a new metric
	collector.Register("test", "field1_ns", map[string]string{"test": "bar"})
	collector.Get("test", "field1_ns", map[string]string{"test": "bar"}).Set(42)

	expected = []telegraf.Metric{
		metric.New(
			"internal_test",
			map[string]string{"global": "zoo", "test": "foo"},
			map[string]interface{}{"field1_ns": int64(10), "field2_ns": int64(5)},
			time.Unix(0, 0),
		),
		metric.New(
			"internal_test",
			map[string]string{"global": "zoo", "test": "bar"},
			map[string]interface{}{"field1_ns": int64(42)},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, Metrics(), options...)
}
