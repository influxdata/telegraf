//go:build linux

package temp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestTemperatureInvalidMetricFormat(t *testing.T) {
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
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}
