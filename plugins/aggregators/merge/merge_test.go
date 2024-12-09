package merge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestSimple(t *testing.T) {
	plugin := &Merge{}
	require.NoError(t, plugin.Init())

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	)
	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	)

	var acc testutil.Accumulator
	plugin.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle":  42,
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestNanosecondPrecision(t *testing.T) {
	plugin := &Merge{}
	require.NoError(t, plugin.Init())

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 1),
		),
	)

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 1),
		),
	)

	var acc testutil.Accumulator
	acc.SetPrecision(time.Second)
	plugin.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle":  42,
				"time_guest": 42,
			},
			time.Unix(0, 1),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestNoRounding(t *testing.T) {
	plugin := &Merge{}
	require.NoError(t, plugin.Init())

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 23,
			},
			time.Unix(0, 1),
		),
	)

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 2),
		),
	)

	var acc testutil.Accumulator
	acc.SetPrecision(time.Second)
	plugin.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 23,
			},
			time.Unix(0, 1),
		),
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 2),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestWithRounding(t *testing.T) {
	plugin := &Merge{RoundTimestamp: config.Duration(10 * time.Nanosecond)}
	require.NoError(t, plugin.Init())

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 23,
			},
			time.Unix(0, 1),
		),
	)

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 2),
		),
	)

	var acc testutil.Accumulator
	acc.SetPrecision(time.Second)
	plugin.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle":  23,
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestReset(t *testing.T) {
	plugin := &Merge{}
	require.NoError(t, plugin.Init())

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	)

	var acc testutil.Accumulator
	plugin.Push(&acc)

	plugin.Reset()

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	)

	plugin.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func BenchmarkMergeOne(b *testing.B) {
	var merger Merge
	require.NoError(b, merger.Init())

	m := metric.New(
		"mymetric",
		map[string]string{
			"host":        "host.example.com",
			"mykey":       "myvalue",
			"another key": "another value",
		},
		map[string]interface{}{
			"f1": 1,
			"f2": 2,
			"f3": 3,
			"f4": 4,
			"f5": 5,
			"f6": 6,
			"f7": 7,
			"f8": 8,
		},
		time.Now(),
	)

	var acc testutil.NopAccumulator
	for n := 0; n < b.N; n++ {
		merger.Reset()
		merger.Add(m)
		merger.Push(&acc)
	}
}

func BenchmarkMergeTwo(b *testing.B) {
	var merger Merge
	require.NoError(b, merger.Init())

	now := time.Now()
	m1 := metric.New(
		"mymetric",
		map[string]string{
			"host":        "host.example.com",
			"mykey":       "myvalue",
			"another key": "another value",
		},
		map[string]interface{}{
			"f1": 1,
			"f2": 2,
			"f3": 3,
			"f4": 4,
			"f5": 5,
			"f6": 6,
			"f7": 7,
			"f8": 8,
		},
		now,
	)

	m2 := metric.New(
		"mymetric",
		map[string]string{
			"host":        "host.example.com",
			"mykey":       "myvalue",
			"another key": "another value",
		},
		map[string]interface{}{
			"f8":  8,
			"f9":  9,
			"f10": 10,
			"f11": 11,
			"f12": 12,
			"f13": 13,
			"f14": 14,
			"f15": 15,
			"f16": 16,
		},
		now,
	)

	var acc testutil.NopAccumulator
	for n := 0; n < b.N; n++ {
		merger.Reset()
		merger.Add(m1)
		merger.Add(m2)
		merger.Push(&acc)
	}
}
