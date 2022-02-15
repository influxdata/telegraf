package amd_rocm_smi

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGatherValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected []telegraf.Metric
	}{
		{
			name:     "Vega 10 XT",
			filename: "vega-10-XT.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x6861",
						"gpu_unique_id": "0x2150e7d042a1124",
						"name":          "card0",
					},
					map[string]interface{}{
						"driver_version":              5925,
						"fan_speed":                   13,
						"memory_total":                int64(17163091968),
						"memory_used":                 int64(17776640),
						"memory_free":                 int64(17145315328),
						"temperature_sensor_edge":     39.0,
						"temperature_sensor_junction": 40.0,
						"temperature_sensor_memory":   92.0,
						"utilization_gpu":             0,
						"clocks_current_sm":           1269,
						"clocks_current_memory":       167,
						"power_draw":                  15.0,
					},
					time.Unix(0, 0)),
			},
		},
		{
			name:     "Vega 20 WKS GL-XE [Radeon Pro VII]",
			filename: "vega-20-WKS-GL-XE.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x66a1",
						"gpu_unique_id": "0x2f048617326b1ea",
						"name":          "card0",
					},
					map[string]interface{}{
						"driver_version":              5917,
						"fan_speed":                   0,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(10850304),
						"memory_free":                 int64(34332110848),
						"temperature_sensor_edge":     36.0,
						"temperature_sensor_junction": 38.0,
						"temperature_sensor_memory":   35.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_sm":           1725,
						"clocks_current_memory":       1000,
						"power_draw":                  26.0,
					},
					time.Unix(0, 0)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator
			octets, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			require.NoError(t, err)

			err = gatherROCmSMI(octets, &acc)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}
