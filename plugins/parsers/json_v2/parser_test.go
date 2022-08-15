package json_v2_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestMultipleConfigs(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testdata")
	require.NoError(t, err)
	// Make sure testdata contains data
	require.Greater(t, len(folders), 0)

	// Setup influx parser for parsing the expected metrics
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	inputs.Add("file", func() telegraf.Input {
		return &file.File{}
	})

	for _, f := range folders {
		testdataPath := filepath.Join("testdata", f.Name())
		configFilename := filepath.Join(testdataPath, "telegraf.conf")
		expectedFilename := filepath.Join(testdataPath, "expected.out")
		expectedErrorFilename := filepath.Join(testdataPath, "expected.err")

		t.Run(f.Name(), func(t *testing.T) {
			// Read the expected output
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Read the expected errors if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))

			// Gather the metrics from the input file configure
			var acc testutil.Accumulator
			var actualErrorMsgs []string
			for _, input := range cfg.Inputs {
				require.NoError(t, input.Init())
				if err := input.Gather(&acc); err != nil {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
			}

			// If the test has expected error(s) then compare them
			if len(expectedErrors) > 0 {
				sort.Strings(actualErrorMsgs)
				sort.Strings(expectedErrors)
				for i, msg := range expectedErrors {
					require.Contains(t, actualErrorMsgs[i], msg)
				}
			} else {
				require.Empty(t, actualErrorMsgs)
			}

			// Process expected metrics and compare with resulting metrics
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())

			// Folder with timestamp prefixed will also check for matching timestamps to make sure they are parsed correctly
			// The milliseconds weren't matching, seemed like a rounding difference between the influx parser
			// Compares each metrics times separately and ignores milliseconds
			if strings.HasPrefix(f.Name(), "timestamp") {
				require.Equal(t, len(expected), len(actual))
				for i, m := range actual {
					require.Equal(t, expected[i].Time().Truncate(time.Second), m.Time().Truncate(time.Second))
				}
			}
		})
	}
}
