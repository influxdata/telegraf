package nvidia_smi

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGatherValidXML(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected []telegraf.Metric
	}{
		{
			name:     "GeForce GTX 1070 Ti",
			filename: "gtx-1070-ti.xml",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"nvidia_smi",
					map[string]string{
						"name":         "GeForce GTX 1070 Ti",
						"compute_mode": "Default",
						"index":        "0",
						"pstate":       "P8",
						"uuid":         "GPU-f9ba66fc-a7f5-94c5-da19-019ef2f9c665",
					},
					map[string]interface{}{
						"clocks_current_graphics":       135,
						"clocks_current_memory":         405,
						"clocks_current_sm":             135,
						"clocks_current_video":          405,
						"encoder_stats_average_fps":     0,
						"encoder_stats_average_latency": 0,
						"encoder_stats_session_count":   0,
						"fan_speed":                     100,
						"memory_free":                   4054,
						"memory_total":                  4096,
						"memory_used":                   42,
						"pcie_link_gen_current":         1,
						"pcie_link_width_current":       16,
						"temperature_gpu":               39,
						"utilization_gpu":               0,
						"utilization_memory":            0,
					},
					time.Unix(0, 0)),
			},
		},
		{
			name:     "GeForce GTX 1660 Ti",
			filename: "gtx-1660-ti.xml",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"nvidia_smi",
					map[string]string{
						"compute_mode": "Default",
						"index":        "0",
						"name":         "Graphics Device",
						"pstate":       "P8",
						"uuid":         "GPU-304a277d-3545-63b8-3a36-dfde3c992989",
					},
					map[string]interface{}{
						"clocks_current_graphics":       300,
						"clocks_current_memory":         405,
						"clocks_current_sm":             300,
						"clocks_current_video":          540,
						"cuda_version":                  "10.1",
						"driver_version":                "418.43",
						"encoder_stats_average_fps":     0,
						"encoder_stats_average_latency": 0,
						"encoder_stats_session_count":   0,
						"fbc_stats_average_fps":         0,
						"fbc_stats_average_latency":     0,
						"fbc_stats_session_count":       0,
						"fan_speed":                     0,
						"memory_free":                   5912,
						"memory_total":                  5912,
						"memory_used":                   0,
						"pcie_link_gen_current":         1,
						"pcie_link_width_current":       16,
						"power_draw":                    8.93,
						"temperature_gpu":               40,
						"utilization_gpu":               0,
						"utilization_memory":            1,
						"utilization_encoder":           0,
						"utilization_decoder":           0,
					},
					time.Unix(0, 0)),
			},
		},
		{
			name:     "Quadro P400",
			filename: "quadro-p400.xml",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"nvidia_smi",
					map[string]string{
						"compute_mode": "Default",
						"index":        "0",
						"name":         "Quadro P400",
						"pstate":       "P8",
						"uuid":         "GPU-8f750be4-dfbc-23b9-b33f-da729a536494",
					},
					map[string]interface{}{
						"clocks_current_graphics":       139,
						"clocks_current_memory":         405,
						"clocks_current_sm":             139,
						"clocks_current_video":          544,
						"cuda_version":                  "10.1",
						"driver_version":                "418.43",
						"encoder_stats_average_fps":     0,
						"encoder_stats_average_latency": 0,
						"encoder_stats_session_count":   0,
						"fbc_stats_average_fps":         0,
						"fbc_stats_average_latency":     0,
						"fbc_stats_session_count":       0,
						"fan_speed":                     34,
						"memory_free":                   1998,
						"memory_total":                  1998,
						"memory_used":                   0,
						"pcie_link_gen_current":         1,
						"pcie_link_width_current":       16,
						"temperature_gpu":               33,
						"utilization_gpu":               0,
						"utilization_memory":            3,
						"utilization_encoder":           0,
						"utilization_decoder":           0,
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

			err = gatherNvidiaSMI(octets, &acc)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}
