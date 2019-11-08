package nvidia_smi

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var payload = []byte(`<?xml version="1.0" ?>
<!DOCTYPE nvidia_smi_log SYSTEM "nvsmi_device_v10.dtd">
<nvidia_smi_log>
        <gpu id="00000000:01:00.0">
                <product_name>GeForce GTX 1070 Ti</product_name>
                <uuid>GPU-f9ba66fc-a7f5-94c5-da19-019ef2f9c665</uuid>
                <pci>
                        <pci_gpu_link_info>
                                <pcie_gen>
                                        <current_link_gen>1</current_link_gen>
                                </pcie_gen>
                                <link_widths>
                                        <current_link_width>16x</current_link_width>
                                </link_widths>
                        </pci_gpu_link_info>
                </pci>
                <fan_speed>100 %</fan_speed>
                <performance_state>P8</performance_state>
                <fb_memory_usage>
                        <total>4096 MiB</total>
                        <used>42 MiB</used>
                        <free>4054 MiB</free>
                </fb_memory_usage>
                <compute_mode>Default</compute_mode>
                <utilization>
                        <gpu_util>0 %</gpu_util>
                        <memory_util>0 %</memory_util>
                </utilization>
                <encoder_stats>
                        <session_count>0</session_count>
                        <average_fps>0</average_fps>
                        <average_latency>0</average_latency>
                </encoder_stats>
                <temperature>
                        <gpu_temp>39 C</gpu_temp>
                </temperature>
                <power_readings>
                        <power_draw>N/A</power_draw>
                </power_readings>
                <clocks>
                        <graphics_clock>135 MHz</graphics_clock>
                        <sm_clock>135 MHz</sm_clock>
                        <mem_clock>405 MHz</mem_clock>
                        <video_clock>405 MHz</video_clock>
                </clocks>
        </gpu>
</nvidia_smi_log>`)

func TestGatherSMI(t *testing.T) {
	var expectedMetric = struct {
		tags   map[string]string
		fields map[string]interface{}
	}{
		tags: map[string]string{
			"name":         "GeForce GTX 1070 Ti",
			"compute_mode": "Default",
			"index":        "0",
			"pstate":       "P8",
			"uuid":         "GPU-f9ba66fc-a7f5-94c5-da19-019ef2f9c665",
		},
		fields: map[string]interface{}{
			"fan_speed":                     100,
			"memory_free":                   4054,
			"memory_used":                   42,
			"memory_total":                  4096,
			"temperature_gpu":               39,
			"utilization_gpu":               0,
			"utilization_memory":            0,
			"pcie_link_gen_current":         1,
			"pcie_link_width_current":       16,
			"encoder_stats_session_count":   0,
			"encoder_stats_average_fps":     0,
			"encoder_stats_average_latency": 0,
			"clocks_current_graphics":       135,
			"clocks_current_sm":             135,
			"clocks_current_memory":         405,
			"clocks_current_video":          405,
		},
	}

	acc := &testutil.Accumulator{}

	gatherNvidiaSMI(payload, acc)
	fmt.Println()

	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, expectedMetric.fields, acc.Metrics[0].Fields)
	require.Equal(t, expectedMetric.tags, acc.Metrics[0].Tags)
}
