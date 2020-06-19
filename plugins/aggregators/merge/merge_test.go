package seriesgrouper

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	plugin := &Merge{}

	err := plugin.Init()
	require.NoError(t, err)

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
	require.NoError(t, err)

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
	require.NoError(t, err)

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

	err := plugin.Init()
	require.NoError(t, err)

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
	require.NoError(t, err)

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
	require.NoError(t, err)

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

func TestReset(t *testing.T) {
	plugin := &Merge{}

	err := plugin.Init()
	require.NoError(t, err)

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
	require.NoError(t, err)

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
	require.NoError(t, err)

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
