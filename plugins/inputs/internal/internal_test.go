package internal

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

func TestSelfPlugin(t *testing.T) {
	s := Internal{
		CollectMemstats: true,
	}
	acc := &testutil.Accumulator{}

	require.NoError(t, s.Gather(acc))
	require.True(t, acc.HasMeasurement("internal_memstats"))

	// test that a registered stat is incremented
	stat := selfstat.Register("mytest", "test", map[string]string{"test": "foo"})
	defer selfstat.Unregister("mytest", "test", map[string]string{"test": "foo"})
	stat.Incr(1)
	stat.Incr(2)
	require.NoError(t, s.Gather(acc))

	acc.AssertContainsTaggedFields(t, "internal_mytest",
		map[string]interface{}{
			"test": int64(3),
		},
		map[string]string{
			"test":    "foo",
			"version": "unknown",
		},
	)
	acc.ClearMetrics()

	// test that a registered stat is set properly
	stat.Set(101)
	require.NoError(t, s.Gather(acc))
	acc.AssertContainsTaggedFields(t, "internal_mytest",
		map[string]interface{}{
			"test": int64(101),
		},
		map[string]string{
			"test":    "foo",
			"version": "unknown",
		},
	)
	acc.ClearMetrics()

	// test that regular and timing stats can share the same measurement, and
	// that timings are set properly.
	timing := selfstat.RegisterTiming("mytest", "test_ns", map[string]string{"test": "foo"})
	defer selfstat.Unregister("mytest", "test_ns", map[string]string{"test": "foo"})
	timing.Incr(100)
	timing.Incr(200)
	require.NoError(t, s.Gather(acc))
	acc.AssertContainsTaggedFields(t, "internal_mytest",
		map[string]interface{}{
			"test":    int64(101),
			"test_ns": int64(150),
		},
		map[string]string{
			"test":    "foo",
			"version": "unknown",
		},
	)
}

func TestNoMemStat(t *testing.T) {
	s := Internal{
		CollectMemstats: false,
		CollectGostats:  false,
	}
	acc := &testutil.Accumulator{}

	require.NoError(t, s.Gather(acc))
	require.False(t, acc.HasMeasurement("internal_memstats"))
	require.False(t, acc.HasMeasurement("internal_gostats"))
}

func TestGostats(t *testing.T) {
	s := Internal{
		CollectMemstats: false,
		CollectGostats:  true,
	}
	acc := &testutil.Accumulator{}

	require.NoError(t, s.Gather(acc))
	require.False(t, acc.HasMeasurement("internal_memstats"))
	require.True(t, acc.HasMeasurement("internal_gostats"))

	var actual *testutil.Metric
	for _, m := range acc.Metrics {
		if m.Measurement == "internal_gostats" {
			actual = m
			break
		}
	}

	require.NotNil(t, actual)
	require.Equal(t, "internal_gostats", actual.Measurement)
	require.Len(t, actual.Tags, 1)
	require.Contains(t, actual.Tags, "go_version")

	for name, value := range actual.Fields {
		switch value.(type) {
		case int64, uint64, float64:
		default:
			require.Failf(t, "Wrong type of field", "Field %s is of non-numeric type %T", name, value)
		}
	}
}

func TestPerInstance(t *testing.T) {
	// Setup plugin statistics to gather with different plugin IDs
	for i := range 3 {
		selfstat.Register(
			"mytest",
			"calls",
			map[string]string{"test": "foo", "_id": "id-" + strconv.Itoa(i)},
		).Incr(int64(100 + i))
		selfstat.Register(
			"mytest",
			"writes",
			map[string]string{"test": "foo", "_id": "id-" + strconv.Itoa(i)},
		).Incr(3 * int64(100+i))
	}
	defer func() {
		for i := range 3 {
			selfstat.Unregister(
				"mytest",
				"calls",
				map[string]string{"test": "foo", "_id": "id-" + strconv.Itoa(i)},
			)
			selfstat.Unregister(
				"mytest",
				"writes",
				map[string]string{"test": "foo", "_id": "id-" + strconv.Itoa(i)},
			)
		}
	}()

	// Setup the internal plugin to collect statistics per plugin _instance_
	plugin := Internal{
		PerInstance: true,
	}

	// Collect
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	// Check the resulting metrics
	expected := []telegraf.Metric{
		metric.New(
			"internal_mytest",
			map[string]string{"_id": "id-0", "test": "foo", "version": "unknown"},
			map[string]interface{}{"calls": 100, "writes": 300},
			time.Unix(0, 0),
		),
		metric.New(
			"internal_mytest",
			map[string]string{"_id": "id-1", "test": "foo", "version": "unknown"},
			map[string]interface{}{"calls": 101, "writes": 303},
			time.Unix(0, 0),
		),
		metric.New(
			"internal_mytest",
			map[string]string{"_id": "id-2", "test": "foo", "version": "unknown"},
			map[string]interface{}{"calls": 102, "writes": 306},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestAccumulatedPerType(t *testing.T) {
	// Setup plugin statistics to gather with different plugin IDs
	for i := range 2 {
		selfstat.Register(
			"mytest",
			"calls",
			map[string]string{"test": "foo", "_id": "id-" + strconv.Itoa(i)},
		).Incr(int64(10 + i))
		selfstat.Register(
			"mytest",
			"writes",
			map[string]string{"test": "foo", "_id": "id-" + strconv.Itoa(i)},
		).Incr(3 * int64(10+i))
	}
	for i := range 2 {
		selfstat.Register(
			"mytest",
			"calls",
			map[string]string{"test": "bar", "_id": "id-" + strconv.Itoa(10+i)},
		).Incr(int64(100 + i))
		selfstat.Register(
			"mytest",
			"writes",
			map[string]string{"test": "bar", "_id": "id-" + strconv.Itoa(10+i)},
		).Incr(2 * int64(100+i))
	}
	defer func() {
		for i := range 2 {
			selfstat.Unregister(
				"mytest",
				"calls",
				map[string]string{"test": "foo", "_id": "id-" + strconv.Itoa(i)},
			)
			selfstat.Unregister(
				"mytest",
				"writes",
				map[string]string{"test": "foo", "_id": "id-" + strconv.Itoa(i)},
			)
		}
		for i := range 2 {
			selfstat.Unregister(
				"mytest",
				"calls",
				map[string]string{"test": "bar", "_id": "id-" + strconv.Itoa(10+i)},
			)
			selfstat.Unregister(
				"mytest",
				"writes",
				map[string]string{"test": "bar", "_id": "id-" + strconv.Itoa(10+i)},
			)
		}
	}()

	// Setup the internal plugin to collect statistics per plugin _type_ not instance
	plugin := Internal{}

	// Collect
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	// Check the resulting metrics
	expected := []telegraf.Metric{
		metric.New(
			"internal_mytest",
			map[string]string{"test": "foo", "version": "unknown"},
			map[string]interface{}{"calls": 21, "writes": 63},
			time.Unix(0, 0),
		),
		metric.New(
			"internal_mytest",
			map[string]string{"test": "bar", "version": "unknown"},
			map[string]interface{}{"calls": 201, "writes": 402},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}
