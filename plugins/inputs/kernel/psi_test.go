//go:build linux

package kernel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestPSIEnabledWrongDir(t *testing.T) {
	k := Kernel{
		psiDir:        "testdata/this_directory_does_not_exist/stub",
		ConfigCollect: []string{"psi"},
	}

	require.ErrorContains(t, k.Init(), "failed to initialize procfs on ")
}

func TestPSIStats(t *testing.T) {
	k := Kernel{
		psiDir:        "testdata/pressure",
		ConfigCollect: []string{"psi"},
	}
	require.NoError(t, k.Init())

	var acc testutil.Accumulator
	require.NoError(t, k.gatherPressure(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"pressure",
			map[string]string{
				"resource": "cpu",
				"type":     "some",
			},
			map[string]interface{}{
				"avg10":  float64(10),
				"avg60":  float64(60),
				"avg300": float64(300),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "cpu",
				"type":     "some",
			},
			map[string]interface{}{
				"total": uint64(114514),
			},
			time.Unix(0, 0),
			telegraf.Counter,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "memory",
				"type":     "some",
			},
			map[string]interface{}{
				"avg10":  float64(10),
				"avg60":  float64(60),
				"avg300": float64(300),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "memory",
				"type":     "some",
			},
			map[string]interface{}{
				"total": uint64(114514),
			},
			time.Unix(0, 0),
			telegraf.Counter,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "io",
				"type":     "some",
			},
			map[string]interface{}{
				"avg10":  float64(10),
				"avg60":  float64(60),
				"avg300": float64(300),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "io",
				"type":     "some",
			},
			map[string]interface{}{
				"total": uint64(114514),
			},
			time.Unix(0, 0),
			telegraf.Counter,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "memory",
				"type":     "full",
			},
			map[string]interface{}{
				"avg10":  float64(1),
				"avg60":  float64(6),
				"avg300": float64(30),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "memory",
				"type":     "full",
			},
			map[string]interface{}{
				"total": uint64(11451),
			},
			time.Unix(0, 0),
			telegraf.Counter,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "io",
				"type":     "full",
			},
			map[string]interface{}{
				"avg10":  float64(1),
				"avg60":  float64(6),
				"avg300": float64(30),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"pressure",
			map[string]string{
				"resource": "io",
				"type":     "full",
			},
			map[string]interface{}{
				"total": uint64(11451),
			},
			time.Unix(0, 0),
			telegraf.Counter,
		),
	}

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}
