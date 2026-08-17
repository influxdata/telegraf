package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetIfUsed(t *testing.T) {
	tests := []struct {
		name     string
		datatype string
		key      string
		value    string
		expected map[string]interface{}
	}{
		{
			name:     "unit is stripped",
			datatype: "int",
			key:      "memory_total",
			value:    "20475 MiB",
			expected: map[string]interface{}{"memory_total": int64(20475)},
		},
		{
			name:     "value exceeding 32 bit",
			datatype: "int",
			key:      "clocks_event_reasons_counters_sw_power_cap",
			value:    "4251825415 us",
			expected: map[string]interface{}{"clocks_event_reasons_counters_sw_power_cap": int64(4251825415)},
		},
		{
			name:     "negative value",
			datatype: "int",
			key:      "temperature_max_tlimit_threshold",
			value:    "-7 C",
			expected: map[string]interface{}{"temperature_max_tlimit_threshold": int64(-7)},
		},
		{
			name:     "link width suffix is stripped",
			datatype: "int",
			key:      "pcie_link_width_current",
			value:    "16x",
			expected: map[string]interface{}{"pcie_link_width_current": int64(16)},
		},
		{
			name:     "float",
			datatype: "float",
			key:      "power_draw",
			value:    "8.44 W",
			expected: map[string]interface{}{"power_draw": 8.44},
		},
		{
			name:     "string",
			datatype: "str",
			key:      "driver_version",
			value:    "595.84",
			expected: map[string]interface{}{"driver_version": "595.84"},
		},
		{
			name:     "multi-word value is preserved",
			datatype: "str",
			key:      "clocks_event_reason_hw_slowdown",
			value:    "Not Active",
			expected: map[string]interface{}{"clocks_event_reason_hw_slowdown": "Not Active"},
		},
		{
			name:     "surrounding whitespace is trimmed",
			datatype: "str",
			key:      "display_mode",
			value:    "  Enabled\n",
			expected: map[string]interface{}{"display_mode": "Enabled"},
		},
		{
			name:     "unsupported value is skipped",
			datatype: "int",
			key:      "memory_temp",
			value:    "N/A",
			expected: map[string]interface{}{},
		},
		{
			name:     "empty value is skipped",
			datatype: "int",
			key:      "fan_speed",
			value:    "",
			expected: map[string]interface{}{},
		},
		{
			name:     "unparsable value is skipped",
			datatype: "int",
			key:      "fan_speed",
			value:    "Unknown Error",
			expected: map[string]interface{}{},
		},
		{
			name:     "deprecated value is skipped",
			datatype: "str",
			key:      "display_mode",
			value:    "Requested functionality has been deprecated",
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := make(map[string]interface{})
			SetIfUsed(tt.datatype, actual, tt.key, tt.value)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestSetTagIfUsed(t *testing.T) {
	actual := make(map[string]string)
	SetTagIfUsed(actual, "name", "NVIDIA RTX 4000 SFF Ada Generation")
	SetTagIfUsed(actual, "arch", "")
	require.Equal(t, map[string]string{"name": "NVIDIA RTX 4000 SFF Ada Generation"}, actual)
}
