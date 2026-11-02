//go:build openbsd

package sensors

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	// The test binary itself stands in for sysctl, see TestMain
	exe, err := os.Executable()
	require.NoError(t, err)

	entries, err := os.ReadDir("testcases")
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testcasePath := filepath.Join("testcases", entry.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		inputFilename := filepath.Join(testcasePath, "input.txt")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		t.Run(entry.Name(), func(t *testing.T) {
			// Load the expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Load the expected error, if the testcase declares one
			var expectedError string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				errorLines, err := testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.Len(t, errorLines, 1)
				expectedError = errorLines[0]
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			plugin := cfg.Inputs[0].Input.(*Sensors)
			require.NoError(t, plugin.Init())

			// Replace sysctl by a mock process printing the testcase input
			t.Setenv("TELEGRAF_SENSORS_MOCK_OUTPUT", inputFilename)
			plugin.path = exe

			var acc testutil.Accumulator
			err = plugin.Gather(&acc)
			if expectedError != "" {
				require.ErrorContains(t, err, expectedError)
				return
			}
			require.NoError(t, err)
			require.Empty(t, acc.Errors)

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

// TestMain lets the test executable stand in for sysctl. When
// TELEGRAF_SENSORS_MOCK_OUTPUT is set the process prints the file it names
// instead of running the tests, so Gather exercises the real command path.
func TestMain(m *testing.M) {
	output := os.Getenv("TELEGRAF_SENSORS_MOCK_OUTPUT")
	if output == "" {
		os.Exit(m.Run())
	}

	if len(os.Args) != 2 || os.Args[1] != "hw.sensors" {
		fmt.Fprintf(os.Stderr, "unexpected arguments %v\n", os.Args[1:])
		os.Exit(1)
	}

	buf, err := os.ReadFile(output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(buf); err != nil {
		os.Exit(1)
	}
}
