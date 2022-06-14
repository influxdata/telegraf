package temp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/testutil"
)

func TestTemperature(t *testing.T) {
	var mps system.MockPS
	var err error
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	ts := host.TemperatureStat{
		SensorKey: "coretemp_sensor1",
		Critical:  60.5,
	}

	mps.On("Temperature").Return([]host.TemperatureStat{ts}, nil)

	err = (&Temperature{ps: &mps}).Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"temp": float64(60.5),
	}

	expectedTags := map[string]string{
		"sensor": "coretemp_sensor1_crit",
	}
	acc.AssertContainsTaggedFields(t, "temp", expectedFields, expectedTags)
}

func TestTemperatureOutputInvalid(t *testing.T) {
	plugin := &Temperature{Scheme: "garbage", ps: system.NewSystemPS()}
	require.Error(t, plugin.Init(), "alal")
}

func TestTemperatureOutputMeasurement(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []telegraf.Metric
	}{
		{
			name: "general",
			path: filepath.Join("testdata", "general"),
			expected: []telegraf.Metric{
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_composite_crit"},
					map[string]interface{}{"temp": 84.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_composite_max"},
					map[string]interface{}{"temp": 81.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_composite_input"},
					map[string]interface{}{"temp": 35.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_sensor_1_crit"},
					map[string]interface{}{"temp": 0.0},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_sensor_1_max"},
					map[string]interface{}{"temp": 65261.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_sensor_1_input"},
					map[string]interface{}{"temp": 35.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_sensor_2_crit"},
					map[string]interface{}{"temp": 0.0},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_sensor_2_max"},
					map[string]interface{}{"temp": 65261.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_sensor_2_input"},
					map[string]interface{}{"temp": 38.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "k10temp_tctl_crit"},
					map[string]interface{}{"temp": 0.0},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "k10temp_tctl_max"},
					map[string]interface{}{"temp": 0.0},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "k10temp_tctl_input"},
					map[string]interface{}{"temp": 33.25},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "k10temp_tccd1_crit"},
					map[string]interface{}{"temp": 0.0},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "k10temp_tccd1_max"},
					map[string]interface{}{"temp": 0.0},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "k10temp_tccd1_input"},
					map[string]interface{}{"temp": 33.25},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		os.Setenv("HOST_SYS", filepath.Join(tt.path, "sys"))

		plugin := &Temperature{Scheme: "measurement", ps: system.NewSystemPS()}
		require.NoError(t, plugin.Init())

		var acc testutil.Accumulator
		require.NoError(t, plugin.Gather(&acc))
		testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
	}
}

func TestTemperatureOutputField(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []telegraf.Metric
	}{
		{
			name: "general",
			path: filepath.Join("testdata", "general"),
			expected: []telegraf.Metric{
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_composite"},
					map[string]interface{}{"crit": 84.85, "high": 81.85, "temp": 35.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_sensor_1"},
					map[string]interface{}{"crit": 0.0, "high": 65261.85, "temp": 35.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "nvme_sensor_2"},
					map[string]interface{}{"crit": 0.0, "high": 65261.85, "temp": 38.85},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "k10temp_tctl"},
					map[string]interface{}{"crit": 0.0, "high": 0.0, "temp": 33.25},
					time.Unix(0, 0),
				),
				metric.New(
					"temp",
					map[string]string{"sensor": "k10temp_tccd1"},
					map[string]interface{}{"crit": 0.0, "high": 0.0, "temp": 33.25},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		os.Setenv("HOST_SYS", filepath.Join(tt.path, "sys"))

		plugin := &Temperature{Scheme: "field", ps: system.NewSystemPS()}
		require.NoError(t, plugin.Init())

		var acc testutil.Accumulator
		require.NoError(t, plugin.Gather(&acc))
		testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
	}
}
