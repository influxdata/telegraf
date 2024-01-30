//go:build linux

package temp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestTemperatureInvaldiMetricFormat(t *testing.T) {
	plugin := &Temperature{
		MetricFormat: "foo",
		Log:          &testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "invalid 'metric_format'")
}

func TestTemperatureMetricV1(t *testing.T) {
	expected := []telegraf.Metric{
		// hwmon0 / temp1
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_composite_alarm"},
			map[string]interface{}{"active": false},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_composite_crit"},
			map[string]interface{}{"temp": 84.85},
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
			map[string]string{"sensor": "nvme_composite_max"},
			map[string]interface{}{"temp": 81.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_composite_min"},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		// hwmon0 / temp2
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_sensor1_input"},
			map[string]interface{}{"temp": 35.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_sensor1_max"},
			map[string]interface{}{"temp": 65261.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_sensor1_min"},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		// hwmon0 / temp3
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_sensor2_input"},
			map[string]interface{}{"temp": 38.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_sensor2_max"},
			map[string]interface{}{"temp": 65261.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_sensor2_min"},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		// hwmon1 / temp1
		metric.New(
			"temp",
			map[string]string{"sensor": "k10temp_tctl_input"},
			map[string]interface{}{"temp": 33.25},
			time.Unix(0, 0),
		),
		// hwmon1 / temp3
		metric.New(
			"temp",
			map[string]string{"sensor": "k10temp_tccd1_input"},
			map[string]interface{}{"temp": 33.25},
			time.Unix(0, 0),
		),
	}

	require.NoError(t, os.Setenv("HOST_SYS", filepath.Join("testcases", "general", "sys")))

	plugin := &Temperature{
		MetricFormat: "v1",
		Log:          &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestTemperature(t *testing.T) {
	expected := []telegraf.Metric{
		// hwmon0 / temp1
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_composite_alarm"},
			map[string]interface{}{"active": false},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_composite_crit"},
			map[string]interface{}{"temp": 84.85},
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
			map[string]string{"sensor": "nvme_composite_max"},
			map[string]interface{}{"temp": 81.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_composite_min"},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		// hwmon0 / temp2
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_sensor_1_input"},
			map[string]interface{}{"temp": 35.85},
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
			map[string]string{"sensor": "nvme_sensor_1_min"},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		// hwmon0 / temp3
		metric.New(
			"temp",
			map[string]string{"sensor": "nvme_sensor_2_input"},
			map[string]interface{}{"temp": 38.85},
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
			map[string]string{"sensor": "nvme_sensor_2_min"},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		// hwmon1 / temp1
		metric.New(
			"temp",
			map[string]string{"sensor": "k10temp_tctl_input"},
			map[string]interface{}{"temp": 33.25},
			time.Unix(0, 0),
		),
		// hwmon1 / temp3
		metric.New(
			"temp",
			map[string]string{"sensor": "k10temp_tccd1_input"},
			map[string]interface{}{"temp": 33.25},
			time.Unix(0, 0),
		),
	}

	require.NoError(t, os.Setenv("HOST_SYS", filepath.Join("testcases", "general", "sys")))
	plugin := &Temperature{Log: &testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestTemperatureNameCollisions(t *testing.T) {
	require.NoError(t, os.Setenv("HOST_SYS", filepath.Join("testcases", "with_name", "sys")))
	plugin := &Temperature{Log: &testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.GetTelegrafMetrics(), 24)
}

func TestTemperatureWithDeviceTag(t *testing.T) {
	expected := []telegraf.Metric{
		// hwmon0 / temp1
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_input",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": 32.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_alarm",
				"device": "nvme0",
			},
			map[string]interface{}{"active": false},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_crit",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": 84.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_min",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_max",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": 81.85},
			time.Unix(0, 0),
		),
		// hwmon0 / temp2
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_1_input",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": 32.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_1_min",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_1_max",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": 65261.85},
			time.Unix(0, 0),
		),
		// hwmon0 / temp3
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_2_input",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": 36.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_2_min",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_2_max",
				"device": "nvme0",
			},
			map[string]interface{}{"temp": 65261.85},
			time.Unix(0, 0),
		),
		// hwmon1 / temp1
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_input",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": 35.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_alarm",
				"device": "nvme1",
			},
			map[string]interface{}{"active": false},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_crit",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": 84.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_min",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_composite_max",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": 81.85},
			time.Unix(0, 0),
		),
		// hwmon1 / temp2
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_1_input",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": 35.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_1_min",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_1_max",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": 65261.85},
			time.Unix(0, 0),
		),
		// hwmon1 / temp3
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_2_input",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": 37.85},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_2_min",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": -273.15},
			time.Unix(0, 0),
		),
		metric.New(
			"temp",
			map[string]string{
				"sensor": "nvme_sensor_2_max",
				"device": "nvme1",
			},
			map[string]interface{}{"temp": 65261.85},
			time.Unix(0, 0),
		),
		// hwmon2 / temp1
		metric.New(
			"temp",
			map[string]string{
				"sensor": "k10temp_tctl_input",
				"device": "0000:00:18.3",
			},
			map[string]interface{}{"temp": 31.875},
			time.Unix(0, 0),
		),
		// hwmon2 / temp3
		metric.New(
			"temp",
			map[string]string{
				"sensor": "k10temp_tccd1_input",
				"device": "0000:00:18.3",
			},
			map[string]interface{}{"temp": 30.75},
			time.Unix(0, 0),
		),
	}

	require.NoError(t, os.Setenv("HOST_SYS", filepath.Join("testcases", "with_name", "sys")))
	plugin := &Temperature{
		DeviceTag: true,
		Log:       &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}
