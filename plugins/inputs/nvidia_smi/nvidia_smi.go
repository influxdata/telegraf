package nvidia_smi

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/transport"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/xpath"

	generic "github.com/influxdata/telegraf/plugins/common/receive_parse"
)

func NewNvidiaSMI() *generic.ReceiveAndParse {
	return &generic.ReceiveAndParse{
		DescriptionText: "Pulls statistics from nvidia GPUs attached to the host",
		Receiver: &transport.Exec{
			BinPath: "/usr/bin/nvidia-smi",
			Timeout: config.Duration(5 * time.Second),
			BinArgs: []string{"-q", "-x"},
		},
		Parser: &xpath.Parser{
			Format: "xml",
			Configs: []xpath.Config{
				{
					MetricDefaultName: "nvidia_smi",
					IgnoreNaN:         true,
					Selection:         "//gpu",
					Tags: map[string]string{
						"index":        "count(./preceding-sibling::gpu)",
						"pstate":       "performance_state",
						"name":         "product_name",
						"uuid":         "uuid",
						"compute_mode": "compute_mode",
					},
					Fields: map[string]string{
						"driver_version": "../driver_version",
						"cuda_version":   "../cuda_version",
						"power_draw":     "number(substring-before(power_readings/power_draw, ' '))",
					},
					FieldsInt: map[string]string{
						"fan_speed":                     "substring-before(fan_speed, ' ')",
						"memory_total":                  "substring-before(fb_memory_usage/total, ' ')",
						"memory_used":                   "substring-before(fb_memory_usage/used, ' ')",
						"memory_free":                   "substring-before(fb_memory_usage/free, ' ')",
						"temperature_gpu":               "substring-before(temperature/gpu_temp, ' ')",
						"utilization_gpu":               "substring-before(utilization/gpu_util, ' ')",
						"utilization_memory":            "substring-before(utilization/memory_util, ' ')",
						"utilization_encoder":           "substring-before(utilization/encoder_util, ' ')",
						"utilization_decoder":           "substring-before(utilization/decoder_util, ' ')",
						"pcie_link_gen_current":         "pci/pci_gpu_link_info/pcie_gen/current_link_gen",
						"pcie_link_width_current":       "substring-before(pci/pci_gpu_link_info/link_widths/current_link_width, 'x')",
						"encoder_stats_session_count":   "encoder_stats/session_count",
						"encoder_stats_average_fps":     "encoder_stats/average_fps",
						"encoder_stats_average_latency": "encoder_stats/average_latency",
						"fbc_stats_session_count":       "fbc_stats/session_count",
						"fbc_stats_average_fps":         "fbc_stats/average_fps",
						"fbc_stats_average_latency":     "fbc_stats/average_latency",
						"clocks_current_graphics":       "substring-before(clocks/graphics_clock, ' ')",
						"clocks_current_sm":             "substring-before(clocks/sm_clock, ' ')",
						"clocks_current_memory":         "substring-before(clocks/mem_clock, ' ')",
						"clocks_current_video":          "substring-before(clocks/video_clock, ' ')",
					},
				},
			},
		},
	}
}

func init() {
	inputs.Add("nvidia_smi", func() telegraf.Input { return NewNvidiaSMI() })
}
