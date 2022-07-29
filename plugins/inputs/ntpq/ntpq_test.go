package ntpq

import (
	"errors"
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

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("ntpq", func() telegraf.Input {
		return &NTPQ{}
	})

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		configFilename := filepath.Join("testcases", f.Name(), "telegraf.conf")
		inputFilename := filepath.Join("testcases", f.Name(), "input.txt")
		inputErrorFilename := filepath.Join("testcases", f.Name(), "input.err")
		expectedFilename := filepath.Join("testcases", f.Name(), "expected.out")
		expectedErrorFilename := filepath.Join("testcases", f.Name(), "expected.err")

		// Compare options
		options := []cmp.Option{
			testutil.IgnoreTime(),
			testutil.SortMetrics(),
		}

		t.Run(f.Name(), func(t *testing.T) {
			// Read the input data
			data, err := os.ReadFile(inputFilename)
			require.NoError(t, err)

			// Read the input error message if any
			var inputErr error
			if _, err := os.Stat(inputErrorFilename); err == nil {
				x, err := testutil.ParseLinesFromFile(inputErrorFilename)
				require.NoError(t, err)
				require.Len(t, x, 1)
				inputErr = errors.New(x[0])
			}

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Read the expected output if any
			var errorMsg string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				x, err := testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.Len(t, x, 1)
				errorMsg = x[0]
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Fake the reading
			plugin := cfg.Inputs[0].Input.(*NTPQ)
			plugin.runQ = func() ([]byte, error) {
				return data, inputErr
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			if errorMsg != "" {
				require.EqualError(t, plugin.Gather(&acc), errorMsg)
				return
			}

			// No error case
			require.NoError(t, plugin.Gather(&acc))
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}
