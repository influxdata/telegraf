package cpu

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("CPU usage mocking works on Linux only!")
	}

	// Get all testcase directories
	testcases, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("cpu", func() telegraf.Input {
		return &CPU{
			PerCPU:   true,
			TotalCPU: true,
			ps:       psutil.NewSystemPS(),
		}
	})

	// Testing options
	opts := []cmp.Option{
		testutil.IgnoreTime(),
		testutil.IgnoreType(),
		testutil.SortMetrics(),
	}

	for _, testcase := range testcases {
		// Only handle folders
		if !testcase.IsDir() {
			continue
		}

		t.Run(testcase.Name(), func(t *testing.T) {
			testcaseDir := filepath.Join("testcases", testcase.Name())
			configFile := filepath.Join(testcaseDir, "telegraf.conf")

			// Load plugin from config
			conf := config.NewConfig()
			require.NoError(t, conf.LoadConfig(configFile))
			require.Len(t, conf.Inputs, 1)
			plugin, ok := conf.Inputs[0].Input.(*CPU)
			require.True(t, ok)

			// Create parser for loading the expected metrics
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			// Get all steps
			matches, err := filepath.Glob(filepath.Join(testcaseDir, "proc*"))
			require.NoError(t, err)
			slices.Sort(matches)

			// Make sure we cleanup the environment
			backup := os.Getenv("HOST_PROC")
			//nolint:usetesting // We need to reset the environment manually as we reset the environment multiple times
			t.Cleanup(func() { os.Setenv("HOST_PROC", backup) })

			// Iterate the different counter states and check the results
			for _, m := range matches {
				proc, err := filepath.Abs(m)
				require.NoError(t, err)

				// Point processing to mock file
				//nolint:usetesting // We manually cleanup the environment as it's unclear whether t.Setenv can handle
				// multiple calls within a test or not
				os.Setenv("HOST_PROC", proc)

				// Read the expected output if any
				expectedErrorFilename := filepath.Join(proc, "expected.err")
				var expectedError string
				if _, err := os.Stat(expectedErrorFilename); err == nil {
					expectedErrors, err := testutil.ParseLinesFromFile(expectedErrorFilename)
					require.NoError(t, err)
					require.Len(t, expectedErrors, 1)
					expectedError = expectedErrors[0]
				}

				// Load exected metrics
				var expected []telegraf.Metric
				if expectedError == "" {
					expected, err = testutil.ParseMetricsFromFile(filepath.Join(proc, "expected.out"), parser)
					require.NoError(t, err)
				}

				// Gather data and check for error or for the returned metrics
				var acc testutil.Accumulator
				err = plugin.Gather(&acc)
				if expectedError != "" {
					require.ErrorContains(t, err, expectedError)
					continue
				}

				// Check the outcome in no-error case
				require.NoError(t, err)
				t.Logf("checking %q", m)
				testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), opts...)
			}
		})
	}
}
