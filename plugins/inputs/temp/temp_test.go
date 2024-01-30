//go:build linux

package temp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestTemperatureInvaldiMetricFormat(t *testing.T) {
	plugin := &Temperature{
		MetricFormat: "foo",
		Log:          &testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "invalid 'metric_format'")
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
	testutil.PrintMetrics(actual)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("temp", func() telegraf.Input {
		return &Temperature{}
	})

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		// Compare options
		options := []cmp.Option{
			testutil.IgnoreTime(),
			testutil.SortMetrics(),
		}

		// Test v1
		t.Run(f.Name()+"_v1", func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			expectedFilename := filepath.Join(testcasePath, "expected_v1.out")

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Prepare the environment
			require.NoError(t, os.Setenv("HOST_SYS", filepath.Join(testcasePath, "sys")))

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Temperature)
			plugin.MetricFormat = "v1"
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})

		// Test v2
		t.Run(f.Name()+"_v2", func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			expectedFilename := filepath.Join(testcasePath, "expected_v2.out")

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Prepare the environment
			require.NoError(t, os.Setenv("HOST_SYS", filepath.Join(testcasePath, "sys")))

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Temperature)
			plugin.MetricFormat = "v2"
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.PrintMetrics(acc.GetTelegrafMetrics())
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}
