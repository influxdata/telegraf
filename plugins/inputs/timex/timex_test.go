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

func TestMetricStructure(t *testing.T) {
	plugin := &Timex{}

	expected := []telegraf.Metric{
		metric.New(
			"timex",
			map[string]string{
				"status": "error",
			},
			map[string]interface{}{
				"offset_ns":                    int64(0),
				"frequency_offset_ppm":         float64(0),
				"maxerror_ns":                  int64(0),
				"estimated_error_ns":           int64(0),
				"status":                       int32(0),
				"loop_time_constant":           int64(0),
				"tick_ns":                      int64(0),
				"pps_frequency_ppm":            float64(0),
				"pps_jitter_ns":                int64(0),
				"pps_shift_sec":                int32(0),
				"pps_stability_ppm":            float64(0),
				"pps_jitter_total":             int64(0),
				"pps_calibration_total":        int64(0),
				"pps_error_total":              int64(0),
				"pps_stability_exceeded_total": int64(0),
				"tai_offset_sec":               int32(0),
				"synchronized":                 bool(false),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	// validate the status tag separately bacause the value could be different based on the system.
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsStructureEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreTags("status"))
	require.NotEmpty(t, actual[0].Tags()["status"])
}
