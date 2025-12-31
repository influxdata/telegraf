package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetTagIfUsed(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected map[string]string
	}{
		{
			name:     "<performance_state>P8</performance_state>",
			key:      "pstate",
			value:    "P8",
			expected: map[string]string{"pstate": "P8"},
		},
		{
			name:     "<product_name>NVIDIA RTX PRO 6000 Blackwell Max-Q Workstation Edition</product_name>",
			key:      "name",
			value:    "NVIDIA RTX PRO 6000 Blackwell Max-Q Workstation Edition",
			expected: map[string]string{"name": "NVIDIA RTX PRO 6000 Blackwell Max-Q Workstation Edition"},
		},
		{
			name:     "<product_architecture>Blackwell</product_architecture>",
			key:      "arch",
			value:    "Blackwell",
			expected: map[string]string{"arch": "Blackwell"},
		},
		{
			name:     "<uuid>GPU-12345678-aaaa-bbbb-cccc-0123456789ab</uuid>",
			key:      "uuid",
			value:    "GPU-12345678-aaaa-bbbb-cccc-0123456789ab",
			expected: map[string]string{"uuid": "GPU-12345678-aaaa-bbbb-cccc-0123456789ab"},
		},
		{
			name:     "<compute_mode>Default</compute_mode>",
			key:      "compute_mode",
			value:    "Default",
			expected: map[string]string{"compute_mode": "Default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := make(map[string]string)
			SetTagIfUsed(m, tt.key, tt.value)
			require.Equal(t, tt.expected, m)
		})
	}
}

func TestSetIfUsed_Float(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected map[string]any
	}{
		{
			name:     "<power_draw>67.03 W</power_draw>",
			key:      "power_draw",
			value:    "67.03 W",
			expected: map[string]any{"power_draw": 67.03},
		},
		{
			name:     "<power_draw>N/A</power_draw>",
			key:      "power_draw",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<power_limit>300.00 W</power_limit>",
			key:      "power_limit",
			value:    "300.00 W",
			expected: map[string]any{"power_limit": 300.0},
		},
		{
			name:     "<current_power_limit>N/A</current_power_limit>",
			key:      "power_limit",
			value:    "N/A",
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := make(map[string]any)
			SetIfUsed("float", m, tt.key, tt.value)
			require.Equal(t, tt.expected, m)
		})
	}
}

func TestSetIfUsed_Int(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected map[string]any
	}{
		{
			name:     "<fan_speed>0 %</fan_speed>",
			key:      "fan_speed",
			value:    "0 %",
			expected: map[string]any{"fan_speed": 0},
		},
		{
			name:     "<fan_speed>30 %</fan_speed>",
			key:      "fan_speed",
			value:    "30 %",
			expected: map[string]any{"fan_speed": 30},
		},
		{
			name:     "<fan_speed>N/A</fan_speed>",
			key:      "fan_speed",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<total>97887 MiB</total>",
			key:      "memory_total",
			value:    "97887 MiB",
			expected: map[string]any{"memory_total": 97887},
		},
		{
			name:     "<reserved>637 MiB</reserved>",
			key:      "memory_reserved",
			value:    "637 MiB",
			expected: map[string]any{"memory_reserved": 637},
		},
		{
			name:     "<used>65720 MiB</used>",
			key:      "memory_used",
			value:    "65720 MiB",
			expected: map[string]any{"memory_used": 65720},
		},
		{
			name:     "<free>31531 MiB</free>",
			key:      "memory_free",
			value:    "31531 MiB",
			expected: map[string]any{"memory_free": 31531},
		},
		{
			name:     "<remapped_row_corr>0</remapped_row_corr>",
			key:      "remapped_rows_correctable",
			value:    "0",
			expected: map[string]any{"remapped_rows_correctable": 0},
		},
		{
			name:     "<remapped_row_unc>0</remapped_row_unc>",
			key:      "remapped_rows_uncorrectable",
			value:    "0",
			expected: map[string]any{"remapped_rows_uncorrectable": 0},
		},
		{
			name:     "<gpu_temp>24 C</gpu_temp>",
			key:      "temperature_gpu",
			value:    "24 C",
			expected: map[string]any{"temperature_gpu": 24},
		},
		{
			name:     "<gpu_util>32 %</gpu_util>",
			key:      "utilization_gpu",
			value:    "32 %",
			expected: map[string]any{"utilization_gpu": 32},
		},
		{
			name:     "<gpu_util>N/A</gpu_util>",
			key:      "utilization_gpu",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<memory_util>37 %</memory_util>",
			key:      "utilization_memory",
			value:    "37 %",
			expected: map[string]any{"utilization_memory": 37},
		},
		{
			name:     "<memory_util>N/A</memory_util>",
			key:      "utilization_memory",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<encoder_util>0 %</encoder_util>",
			key:      "utilization_encoder",
			value:    "0 %",
			expected: map[string]any{"utilization_encoder": 0},
		},
		{
			name:     "<encoder_util>N/A</encoder_util>",
			key:      "utilization_encoder",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<decoder_util>3 %</decoder_util>",
			key:      "utilization_decoder",
			value:    "3 %",
			expected: map[string]any{"utilization_decoder": 3},
		},
		{
			name:     "<decoder_util>N/A</decoder_util>",
			key:      "utilization_decoder",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<jpeg_util>32 %</jpeg_util>",
			key:      "utilization_jpeg",
			value:    "32 %",
			expected: map[string]any{"utilization_jpeg": 32},
		},
		{
			name:     "<jpeg_util>N/A</jpeg_util>",
			key:      "utilization_jpeg",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<ofa_util>0 %</ofa_util>",
			key:      "utilization_ofa",
			value:    "0 %",
			expected: map[string]any{"utilization_ofa": 0},
		},
		{
			name:     "<ofa_util>N/A</ofa_util>",
			key:      "utilization_ofa",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<current_link_gen>4</current_link_gen>",
			key:      "pcie_link_gen_current",
			value:    "4",
			expected: map[string]any{"pcie_link_gen_current": 4},
		},
		{
			name:     "<pcie_link_width_current>16x</pcie_link_width_current>",
			key:      "pcie_link_width_current",
			value:    "16x",
			expected: map[string]any{"pcie_link_width_current": 16},
		},
		{
			name:     "<session_count>0</session_count>",
			key:      "encoder_stats_session_count",
			value:    "0",
			expected: map[string]any{"encoder_stats_session_count": 0},
		},
		{
			name:     "<average_fps>0</average_fps>",
			key:      "encoder_stats_average_fps",
			value:    "0",
			expected: map[string]any{"encoder_stats_average_fps": 0},
		},
		{
			name:     "<average_latency>0</average_latency>",
			key:      "encoder_stats_average_latency",
			value:    "0",
			expected: map[string]any{"encoder_stats_average_latency": 0},
		},
		{
			name:     "<session_count>0</session_count>",
			key:      "fbc_stats_session_count",
			value:    "0",
			expected: map[string]any{"fbc_stats_session_count": 0},
		},
		{
			name:     "<average_fps>0</average_fps>",
			key:      "fbc_stats_average_fps",
			value:    "0",
			expected: map[string]any{"fbc_stats_average_fps": 0},
		},
		{
			name:     "<average_latency>0</average_latency>",
			key:      "fbc_stats_average_latency",
			value:    "0",
			expected: map[string]any{"fbc_stats_average_latency": 0},
		},
		{
			name:     "<graphics_clock>2100 MHz</graphics_clock>",
			key:      "clocks_current_graphics",
			value:    "2100 MHz",
			expected: map[string]any{"clocks_current_graphics": 2100},
		},
		{
			name:     "<graphics_clock>N/A</graphics_clock>",
			key:      "clocks_current_graphics",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<graphics_clock>Requested functionality has been deprecated</graphics_clock>",
			key:      "clocks_current_graphics",
			value:    "Requested functionality has been deprecated",
			expected: map[string]any{},
		},
		{
			name:     "<sm_clock>2100 MHz</sm_clock>",
			key:      "clocks_current_sm",
			value:    "2100 MHz",
			expected: map[string]any{"clocks_current_sm": 2100},
		},
		{
			name:     "<mem_clock>9751 MHz</mem_clock>",
			key:      "clocks_current_memory",
			value:    "9751 MHz",
			expected: map[string]any{"clocks_current_memory": 9751},
		},
		{
			name:     "<video_clock>1950 MHz</video_clock>",
			key:      "clocks_current_video",
			value:    "1950 MHz",
			expected: map[string]any{"clocks_current_video": 1950},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := make(map[string]any)
			SetIfUsed("int", m, tt.key, tt.value)
			require.Equal(t, tt.expected, m)
		})
	}
}

func TestSetIfUsed_String(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected map[string]any
	}{
		{
			name:     "<driver_version>590.44.01</driver_version>",
			key:      "driver_version",
			value:    "590.44.01",
			expected: map[string]any{"driver_version": "590.44.01"},
		},
		{
			name:     "<cuda_version>13.1</cuda_version>",
			key:      "cuda_version",
			value:    "13.1",
			expected: map[string]any{"cuda_version": "13.1"},
		},
		{
			name:     "<serial>1650522003820</serial>",
			key:      "serial",
			value:    "1650522003820",
			expected: map[string]any{"serial": "1650522003820"},
		},
		{
			name:     "<serial>N/A</serial>",
			key:      "serial",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<vbios_version>98.02.6A.00.03</vbios_version>",
			key:      "vbios_version",
			value:    "98.02.6A.00.03",
			expected: map[string]any{"vbios_version": "98.02.6A.00.03"},
		},
		{
			name:     "<display_active>Enabled</display_active>",
			key:      "display_active",
			value:    "Enabled",
			expected: map[string]any{"display_active": "Enabled"},
		},
		{
			name:     "<display_active>Disabled</display_active>",
			key:      "display_active",
			value:    "Disabled",
			expected: map[string]any{"display_active": "Disabled"},
		},
		{
			name:     "<display_mode>Disabled</display_mode>",
			key:      "display_mode",
			value:    "Disabled",
			expected: map[string]any{"display_mode": "Disabled"},
		},
		{
			name:     "<display_mode>Requested functionality has been deprecated</display_mode>",
			key:      "display_mode",
			value:    "Requested functionality has been deprecated",
			expected: map[string]any{},
		},
		{
			name:     "<current_ecc>Enabled</current_ecc>",
			key:      "current_ecc",
			value:    "Enabled",
			expected: map[string]any{"current_ecc": "Enabled"},
		},
		{
			name:     "<current_ecc>N/A</current_ecc>",
			key:      "current_ecc",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<pending_blacklist>No</pending_blacklist>",
			key:      "retired_pages_blacklist",
			value:    "No",
			expected: map[string]any{"retired_pages_blacklist": "No"},
		},
		{
			name:     "<pending_blacklist>N/A</pending_blacklist>",
			key:      "retired_pages_blacklist",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<pending_retirement>No</pending_retirement>",
			key:      "retired_pages_pending",
			value:    "No",
			expected: map[string]any{"retired_pages_pending": "No"},
		},
		{
			name:     "<pending_retirement>N/A</pending_retirement>",
			key:      "retired_pages_pending",
			value:    "N/A",
			expected: map[string]any{},
		},
		{
			name:     "<remapped_row_pending>No</remapped_row_pending>",
			key:      "remapped_rows_pending",
			value:    "No",
			expected: map[string]any{"remapped_rows_pending": "No"},
		},
		{
			name:     "<remapped_row_failure>No</remapped_row_failure>",
			key:      "remapped_rows_failure",
			value:    "No",
			expected: map[string]any{"remapped_rows_failure": "No"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := make(map[string]any)
			SetIfUsed("str", m, tt.key, tt.value)
			require.Equal(t, tt.expected, m)
		})
	}
}
