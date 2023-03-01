package avro

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testdata")
	require.NoError(t, err)
	// Make sure testdata contains data
	require.NotEmpty(t, folders)

	// Set up for file inputs
	inputs.Add("file", func() telegraf.Input {
		return &file.File{}
	})

	for _, f := range folders {
		fname := f.Name()
		testdataPath := filepath.Join("testdata", fname)
		configFilename := filepath.Join(testdataPath, "telegraf.conf")
		expectedFilename := filepath.Join(testdataPath, "expected.out")
		expectedErrorFilename := filepath.Join(testdataPath, "expected.err")

		t.Run(fname, func(t *testing.T) {
			// Get parser to parse expected output
			testdataParser := &influx.Parser{}
			require.NoError(t, testdataParser.Init())

			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, testdataParser)
				require.NoError(t, err)
			}

			// Read the expected errors if any
			var expectedErrors []string

			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Set up error catching
			var acc testutil.Accumulator
			var actualErrors []string
			var actual []telegraf.Metric

			// Configure the plugin
			cfg := config.NewConfig()
			err := cfg.LoadConfig(configFilename)
			require.NoError(t, err)

			for _, input := range cfg.Inputs {
				require.NoError(t, input.Init())

				if err := input.Gather(&acc); err != nil {
					actualErrors = append(actualErrors, err.Error())
				}
			}
			require.ElementsMatch(t, actualErrors, expectedErrors)
			actual = acc.GetTelegrafMetrics()
			// Process expected metrics and compare with resulting metrics
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}
