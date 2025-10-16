//go:build linux

package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestDefaultMetricFormat(t *testing.T) {
	plugin := &Timex{
		Log: &testutil.Logger{},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.GetTelegrafMetrics(), 1)
}

func TestMetricStructure(t *testing.T) {
	plugin := &Timex{
		Log: &testutil.Logger{},
	}

	expected := []telegraf.Metric{
		metric.New(
			"timex",
			map[string]string{},
			map[string]interface{}{
				"offset_ns":                    int64(0),
				"frequency_adjustment_ratio":   float64(0),
				"maxerror_ns":                  int64(0),
				"estimated_error_ns":           int64(0),
				"status":                       int64(0),
				"loop_time_constant":           int64(0),
				"tick_ns":                      int64(0),
				"pps_frequency_hertz":          float64(0),
				"pps_jitter_ns":                int64(0),
				"pps_shift_seconds":            int64(0),
				"pps_stability_hertz":          float64(0),
				"pps_jitter_total":             int64(0),
				"pps_calibration_total":        int64(0),
				"pps_error_total":              int64(0),
				"pps_stability_exceeded_total": int64(0),
				"tai_offset_seconds":           int64(0),
				"sync_status":                  bool(false),
			},
			time.Unix(0, 0),
			2,
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	actual := acc.GetTelegrafMetrics()

	testutil.RequireMetricsStructureEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreType())
}
