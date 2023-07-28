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
		{
			name:     "Tesla T4",
			filename: "tesla-t4.xml",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"nvidia_smi",
					map[string]string{
						"compute_mode": "Default",
						"index":        "0",
						"name":         "Tesla T4",
						"pstate":       "P0",
						"uuid":         "GPU-d37e67a5-91dd-3774-a5cb-99096249601a",
					},
					map[string]interface{}{
						"clocks_current_graphics":           585,
						"clocks_current_memory":             5000,
						"clocks_current_sm":                 585,
						"clocks_current_video":              810,
						"cuda_version":                      "11.7",
						"driver_version":                    "515.105.01",
						"encoder_stats_average_fps":         0,
						"encoder_stats_average_latency":     0,
						"encoder_stats_session_count":       0,
						"fbc_stats_average_fps":             0,
						"fbc_stats_average_latency":         0,
						"fbc_stats_session_count":           0,
						"power_draw":                        26.78,
						"memory_free":                       13939,
						"memory_total":                      15360,
						"memory_used":                       1032,
						"memory_reserved":                   388,
						"retired_pages_multiple_single_bit": 0,
						"retired_pages_double_bit":          0,
						"retired_pages_blacklist":           "No",
						"retired_pages_pending":             "No",
						"pcie_link_gen_current":             3,
						"pcie_link_width_current":           8,
						"temperature_gpu":                   40,
						"utilization_gpu":                   0,
						"utilization_memory":                0,
						"utilization_encoder":               0,
						"utilization_decoder":               0,
					},
					time.Unix(0, 0)),
			},
		},
		{
			name:     "A10G",
			filename: "a10g.xml",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"nvidia_smi",
					map[string]string{
						"compute_mode": "Default",
						"index":        "0",
						"name":         "NVIDIA A10G",
						"pstate":       "P8",
						"uuid":         "GPU-9a9a6c50-2a47-2f51-a902-b82c3b127e94",
					},
					map[string]interface{}{
						"clocks_current_graphics":       210,
						"clocks_current_memory":         405,
						"clocks_current_sm":             210,
						"clocks_current_video":          555,
						"cuda_version":                  "11.7",
						"driver_version":                "515.105.01",
						"encoder_stats_average_fps":     0,
						"encoder_stats_average_latency": 0,
						"encoder_stats_session_count":   0,
						"fbc_stats_average_fps":         0,
						"fbc_stats_average_latency":     0,
						"fbc_stats_session_count":       0,
						"fan_speed":                     0,
						"power_draw":                    25.58,
						"memory_free":                   22569,
						"memory_total":                  23028,
						"memory_used":                   22,
						"memory_reserved":               435,
						"remapped_rows_correctable":     0,
						"remapped_rows_uncorrectable":   0,
						"remapped_rows_pending":         "No",
						"remapped_rows_failure":         "No",
						"pcie_link_gen_current":         1,
						"pcie_link_width_current":       8,
						"temperature_gpu":               17,
						"utilization_gpu":               0,
						"utilization_memory":            0,
						"utilization_encoder":           0,
						"utilization_decoder":           0,
					},
					time.Unix(0, 0)),
			},
		},
		{
			name:     "RTC 3080 schema v12",
			filename: "rtx-3080-v12.xml",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"nvidia_smi",
					map[string]string{
						"compute_mode": "Default",
						"index":        "0",
						"name":         "NVIDIA GeForce RTX 3080",
						"arch":         "Ampere",
						"pstate":       "P8",
						"uuid":         "GPU-19d6d965-2acc-f646-00f8-4c76979aabb4",
					},
					map[string]interface{}{
						"clocks_current_graphics":       210,
						"clocks_current_memory":         405,
						"clocks_current_sm":             210,
						"clocks_current_video":          555,
						"cuda_version":                  "12.2",
						"driver_version":                "536.40",
						"encoder_stats_average_fps":     0,
						"encoder_stats_average_latency": 0,
						"encoder_stats_session_count":   0,
						"fbc_stats_average_fps":         0,
						"fbc_stats_average_latency":     0,
						"fbc_stats_session_count":       0,
						"fan_speed":                     0,
						"power_draw":                    22.78,
						"memory_free":                   8938,
						"memory_total":                  10240,
						"memory_used":                   1128,
						"memory_reserved":               173,
						"pcie_link_gen_current":         4,
						"pcie_link_width_current":       16,
						"temperature_gpu":               31,
						"utilization_gpu":               0,
						"utilization_memory":            37,
						"utilization_encoder":           0,
						"utilization_decoder":           0,
					},
					time.Unix(1689872450, 0)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			octets, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			require.NoError(t, err)

			plugin := &NvidiaSMI{Log: &testutil.Logger{}}

			var acc testutil.Accumulator
			require.NoError(t, plugin.parse(&acc, octets))
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}
