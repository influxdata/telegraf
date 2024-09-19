//go:build linux

package temp

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestDefaultMetricFormat(t *testing.T) {
	plugin := &Temperature{
		Log: &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.Equal(t, "v2", plugin.MetricFormat)
}

func TestInvalidMetricFormat(t *testing.T) {
	plugin := &Temperature{
		MetricFormat: "foo",
		Log:          &testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "invalid 'metric_format'")
}

func TestNameCollisions(t *testing.T) {
	t.Setenv("HOST_SYS", filepath.Join("testcases", "with_name", "sys"))

	plugin := &Temperature{Log: &testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.GetTelegrafMetrics(), 8)
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
			t.Setenv("HOST_SYS", filepath.Join(testcasePath, "sys"))

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
			t.Setenv("HOST_SYS", filepath.Join(testcasePath, "sys"))

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
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func TestRegression(t *testing.T) {
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

		// Test v1 metrics
		t.Run(f.Name()+"_v1", func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			actualFilename := filepath.Join(testcasePath, "expected_v1.out")

			// Read the expected output if any
			var actual []telegraf.Metric
			if _, err := os.Stat(actualFilename); err == nil {
				var err error
				actual, err = testutil.ParseMetricsFromFile(actualFilename, parser)
				require.NoError(t, err)
			}

			// Remove potential device-tags
			for i := range actual {
				actual[i].RemoveTag("device")
			}

			// Use the <v1.22.4 code to compare against
			temps, err := sensorsTemperaturesOld(filepath.Join(testcasePath, "sys"))
			require.NoError(t, err)

			var acc testutil.Accumulator
			for _, temp := range temps {
				tags := map[string]string{
					"sensor": temp.SensorKey,
				}
				fields := map[string]interface{}{
					"temp": temp.Temperature,
				}
				acc.AddFields("temp", fields, tags)
			}

			expected := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})

		// Test v2 metrics
		t.Run(f.Name()+"_v2", func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			actualFilename := filepath.Join(testcasePath, "expected_v2.out")

			// Read the expected output if any
			var actual []telegraf.Metric
			if _, err := os.Stat(actualFilename); err == nil {
				var err error
				actual, err = testutil.ParseMetricsFromFile(actualFilename, parser)
				require.NoError(t, err)
			}

			// Remove potential device-tags
			for i := range actual {
				actual[i].RemoveTag("device")
			}

			// Prepare the environment
			t.Setenv("HOST_SYS", filepath.Join(testcasePath, "sys"))

			// Use the v1.28.x code to compare against
			var acc testutil.Accumulator
			temps, err := host.SensorsTemperatures()
			require.NoError(t, err)
			for _, temp := range temps {
				tags := map[string]string{
					"sensor": temp.SensorKey,
				}
				fields := map[string]interface{}{
					"temp": temp.Temperature,
				}
				acc.AddFields("temp", fields, tags)
			}

			expected := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func sensorsTemperaturesOld(syspath string) ([]host.TemperatureStat, error) {
	files, err := filepath.Glob(syspath + "/class/hwmon/hwmon*/temp*_*")
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		// CentOS has an intermediate /device directory:
		// https://github.com/giampaolo/psutil/issues/971
		files, err = filepath.Glob(syspath + "/class/hwmon/hwmon*/device/temp*_*")
		if err != nil {
			return nil, err
		}
	}

	if len(files) == 0 { // handle distributions without hwmon, like raspbian #391, parse legacy thermal_zone files
		files, err = filepath.Glob(syspath + "/class/thermal/thermal_zone*/")
		if err != nil {
			return nil, err
		}

		temperatures := make([]host.TemperatureStat, 0, len(files))
		for _, file := range files {
			// Get the name of the temperature you are reading
			name, err := os.ReadFile(filepath.Join(file, "type"))
			if err != nil {
				fmt.Println(err)
				continue
			}
			// Get the temperature reading
			current, err := os.ReadFile(filepath.Join(file, "temp"))
			if err != nil {
				fmt.Println(err)
				continue
			}
			temperature, err := strconv.ParseInt(strings.TrimSpace(string(current)), 10, 64)
			if err != nil {
				fmt.Println(err)
				continue
			}

			temperatures = append(temperatures, host.TemperatureStat{
				SensorKey:   strings.TrimSpace(string(name)),
				Temperature: float64(temperature) / 1000.0,
			})
		}
		return temperatures, nil
	}

	// example directory
	// device/           temp1_crit_alarm  temp2_crit_alarm  temp3_crit_alarm  temp4_crit_alarm  temp5_crit_alarm  temp6_crit_alarm  temp7_crit_alarm
	// name              temp1_input       temp2_input       temp3_input       temp4_input       temp5_input       temp6_input       temp7_input
	// power/            temp1_label       temp2_label       temp3_label       temp4_label       temp5_label       temp6_label       temp7_label
	// subsystem/        temp1_max         temp2_max         temp3_max         temp4_max         temp5_max         temp6_max         temp7_max
	// temp1_crit        temp2_crit        temp3_crit        temp4_crit        temp5_crit        temp6_crit        temp7_crit        uevent
	temperatures := make([]host.TemperatureStat, 0, len(files))
	for _, file := range files {
		filename := strings.Split(filepath.Base(file), "_")
		if filename[1] == "label" {
			// Do not try to read the temperature of the label file
			continue
		}

		// Get the label of the temperature you are reading
		var label string
		//nolint:errcheck // skip on error
		c, _ := os.ReadFile(filepath.Join(filepath.Dir(file), filename[0]+"_label"))
		if c != nil {
			// format the label from "Core 0" to "core0_"
			label = strings.Join(strings.Split(strings.TrimSpace(strings.ToLower(string(c))), " "), "") + "_"
		}

		// Get the name of the temperature you are reading
		name, err := os.ReadFile(filepath.Join(filepath.Dir(file), "name"))
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Get the temperature reading
		current, err := os.ReadFile(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		temperature, err := strconv.ParseFloat(strings.TrimSpace(string(current)), 64)
		if err != nil {
			fmt.Println(err)
			continue
		}

		tempName := strings.TrimSpace(strings.ToLower(strings.Join(filename[1:], "")))
		temperatures = append(temperatures, host.TemperatureStat{
			SensorKey:   fmt.Sprintf("%s_%s%s", strings.TrimSpace(string(name)), label, tempName),
			Temperature: temperature / 1000.0,
		})
	}
	return temperatures, nil
}
