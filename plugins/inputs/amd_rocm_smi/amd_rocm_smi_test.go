package amd_rocm_smi

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
)

func TestErrorBehaviorDefault(t *testing.T) {
	// make sure we can't find rocm-smi in $PATH somewhere
	os.Unsetenv("PATH")
	plugin := &ROCmSMI{
		BinPath: "/random/non-existent/path",
		Log:     &testutil.Logger{},
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name: "amd_rocm_smi",
	})
	require.NoError(t, model.Init())

	var acc testutil.Accumulator
	var ferr *internal.FatalError
	require.False(t, errors.As(model.Start(&acc), &ferr))
	require.ErrorIs(t, model.Gather(&acc), internal.ErrNotConnected)
}

func TestErrorBehaviorError(t *testing.T) {
	// make sure we can't find rocm-smi in $PATH somewhere
	os.Unsetenv("PATH")
	plugin := &ROCmSMI{
		BinPath: "/random/non-existent/path",
		Log:     &testutil.Logger{},
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:                 "amd_rocm_smi",
		StartupErrorBehavior: "error",
	})
	require.NoError(t, model.Init())

	var acc testutil.Accumulator
	var ferr *internal.FatalError
	require.False(t, errors.As(model.Start(&acc), &ferr))
	require.ErrorIs(t, model.Gather(&acc), internal.ErrNotConnected)
}

func TestErrorBehaviorRetry(t *testing.T) {
	// make sure we can't find nvidia-smi in $PATH somewhere
	os.Unsetenv("PATH")
	plugin := &ROCmSMI{
		BinPath: "/random/non-existent/path",
		Log:     &testutil.Logger{},
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:                 "amd_rocm_smi",
		StartupErrorBehavior: "retry",
	})
	require.NoError(t, model.Init())

	var acc testutil.Accumulator
	var ferr *internal.FatalError
	require.False(t, errors.As(model.Start(&acc), &ferr))
	require.ErrorIs(t, model.Gather(&acc), internal.ErrNotConnected)
}

func TestErrorBehaviorIgnore(t *testing.T) {
	// make sure we can't find nvidia-smi in $PATH somewhere
	os.Unsetenv("PATH")
	plugin := &ROCmSMI{
		BinPath: "/random/non-existent/path",
		Log:     &testutil.Logger{},
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:                 "amd_rocm_smi",
		StartupErrorBehavior: "ignore",
	})
	require.NoError(t, model.Init())

	var acc testutil.Accumulator
	var ferr *internal.FatalError
	require.ErrorAs(t, model.Start(&acc), &ferr)
	require.ErrorIs(t, model.Gather(&acc), internal.ErrNotConnected)
}

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
						"card_model":                  "0xc1e",
						"card_vendor":                 "Advanced",
						"driver_version":              5925,
						"fan_speed":                   13,
						"memory_total":                int64(17163091968),
						"memory_used":                 int64(17776640),
						"memory_free":                 int64(17145315328),
						"temperature_sensor_edge":     39.0,
						"temperature_sensor_junction": 40.0,
						"temperature_sensor_memory":   92.0,
						"utilization_gpu":             0,
						"clocks_current_display":      600,
						"clocks_current_sm":           1269,
						"clocks_current_memory":       167,
						"clocks_current_system":       960,
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
						"card_model":                  "0x834",
						"card_series":                 "Radeon",
						"card_vendor":                 "Advanced",
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
						"clocks_current_display":      357,
						"clocks_current_fabric":       1080,
						"clocks_current_sm":           1725,
						"clocks_current_memory":       1000,
						"clocks_current_system":       971,
						"power_draw":                  26.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:     "mi100 + ROCm 571",
			filename: "mi100_rocm571.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "N/A",
						"name":          "card0",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              624,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6782976),
						"memory_free":                 int64(34336178176),
						"temperature_sensor_edge":     31.0,
						"temperature_sensor_junction": 34.0,
						"temperature_sensor_memory":   30.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  39.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "N/A",
						"name":          "card1",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              624,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6782976),
						"memory_free":                 int64(34336178176),
						"temperature_sensor_edge":     30.0,
						"temperature_sensor_junction": 33.0,
						"temperature_sensor_memory":   38.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  37.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "N/A",
						"name":          "card2",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              624,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6782976),
						"memory_free":                 int64(34336178176),
						"temperature_sensor_edge":     31.0,
						"temperature_sensor_junction": 34.0,
						"temperature_sensor_memory":   31.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  35.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "N/A",
						"name":          "card3",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              624,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6782976),
						"memory_free":                 int64(34336178176),
						"temperature_sensor_edge":     33.0,
						"temperature_sensor_junction": 35.0,
						"temperature_sensor_memory":   36.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  39.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "N/A",
						"name":          "card4",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              624,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6782976),
						"memory_free":                 int64(34336178176),
						"temperature_sensor_edge":     32.0,
						"temperature_sensor_junction": 34.0,
						"temperature_sensor_memory":   38.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  39.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "N/A",
						"name":          "card5",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              624,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6782976),
						"memory_free":                 int64(34336178176),
						"temperature_sensor_edge":     33.0,
						"temperature_sensor_junction": 35.0,
						"temperature_sensor_memory":   38.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  40.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:     "mi100 + ROCm 602",
			filename: "mi100_rocm602.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "0x79ccd55167a2124a",
						"name":          "card0",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              636,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6750208),
						"memory_free":                 int64(34336210944),
						"temperature_sensor_edge":     53.0,
						"temperature_sensor_junction": 55.0,
						"temperature_sensor_memory":   53.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  36.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "0x4edfb117a17a07d",
						"name":          "card1",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              636,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6750208),
						"memory_free":                 int64(34336210944),
						"temperature_sensor_edge":     55.0,
						"temperature_sensor_junction": 58.0,
						"temperature_sensor_memory":   54.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  44.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "0xd4a9ec48d03d261d",
						"name":          "card2",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              636,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6750208),
						"memory_free":                 int64(34336210944),
						"temperature_sensor_edge":     54.0,
						"temperature_sensor_junction": 57.0,
						"temperature_sensor_memory":   55.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  43.0,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x738c",
						"gpu_unique_id": "0x1b9dd972253c3736",
						"name":          "card3",
					},
					map[string]interface{}{
						"card_model":                  "0x0c34",
						"card_series":                 "Arcturus",
						"card_vendor":                 "Advanced",
						"driver_version":              636,
						"memory_total":                int64(34342961152),
						"memory_used":                 int64(6750208),
						"memory_free":                 int64(34336210944),
						"temperature_sensor_edge":     51.0,
						"temperature_sensor_junction": 53.0,
						"temperature_sensor_memory":   50.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_fabric":       1402,
						"clocks_current_sm":           300,
						"clocks_current_memory":       1200,
						"clocks_current_system":       1000,
						"power_draw":                  39.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:     "rx6700xt + ROCm 430",
			filename: "rx6700xt_rocm430.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x73df",
						"gpu_unique_id": "N/A",
						"name":          "card0",
					},
					map[string]interface{}{
						"card_model":                  "0x1002",
						"card_series":                 "0x1002",
						"card_vendor":                 "0x1002",
						"driver_version":              636,
						"memory_total":                int64(12868124672),
						"memory_used":                 int64(1622728704),
						"memory_free":                 int64(11245395968),
						"temperature_sensor_edge":     45.0,
						"temperature_sensor_junction": 47.0,
						"temperature_sensor_memory":   46.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_display":      480,
						"clocks_current_fabric":       1051,
						"clocks_current_sm":           500,
						"clocks_current_memory":       96,
						"clocks_current_system":       685,
						"power_draw":                  6.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:     "rx6700xt + ROCm 571",
			filename: "rx6700xt_rocm571.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x73df",
						"gpu_unique_id": "N/A",
						"name":          "card0",
					},
					map[string]interface{}{
						"card_model":                  "0x6601",
						"card_series":                 "Navi",
						"card_vendor":                 "Advanced",
						"driver_version":              636,
						"memory_total":                int64(12868124672),
						"memory_used":                 int64(1564491776),
						"memory_free":                 int64(11303632896),
						"temperature_sensor_edge":     45.0,
						"temperature_sensor_junction": 47.0,
						"temperature_sensor_memory":   46.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_display":      480,
						"clocks_current_fabric":       1051,
						"clocks_current_sm":           500,
						"clocks_current_memory":       96,
						"clocks_current_system":       685,
						"power_draw":                  6.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:     "rx6700xt + ROCm 602",
			filename: "rx6700xt_rocm602.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x73df",
						"gpu_unique_id": "N/A",
						"name":          "card0",
					},
					map[string]interface{}{
						"card_model":                  "0x6601",
						"card_series":                 "Navi",
						"card_vendor":                 "Advanced",
						"driver_version":              636,
						"memory_total":                int64(12868124672),
						"memory_used":                 int64(1572757504),
						"memory_free":                 int64(11295367168),
						"temperature_sensor_edge":     45.0,
						"temperature_sensor_junction": 47.0,
						"temperature_sensor_memory":   46.0,
						"utilization_gpu":             0,
						"utilization_memory":          0,
						"clocks_current_display":      480,
						"clocks_current_fabric":       1051,
						"clocks_current_sm":           500,
						"clocks_current_memory":       96,
						"clocks_current_system":       685,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:     "rx6700xt + ROCm 612",
			filename: "rx6700xt_rocm612.json",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"amd_rocm_smi",
					map[string]string{
						"gpu_id":        "0x73df",
						"gpu_unique_id": "N/A",
						"name":          "card0",
					},
					map[string]interface{}{
						"card_model":                  "0x73df",
						"card_series":                 "Navi",
						"card_vendor":                 "Advanced",
						"driver_version":              636,
						"memory_total":                int64(12868124672),
						"memory_used":                 int64(1572745216),
						"memory_free":                 int64(11295379456),
						"temperature_sensor_edge":     45.0,
						"temperature_sensor_junction": 47.0,
						"temperature_sensor_memory":   46.0,
						"utilization_gpu":             0,
						"utilization_memory":          12,
						"clocks_current_display":      480,
						"clocks_current_fabric":       1051,
						"clocks_current_sm":           0,
						"clocks_current_memory":       96,
						"clocks_current_system":       685,
						"power_draw":                  6.0,
					},
					time.Unix(0, 0),
				),
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

			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}
