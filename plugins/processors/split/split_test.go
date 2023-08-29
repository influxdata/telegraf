package split

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestCases(t *testing.T) {
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)
	require.NotEmpty(t, folders)

	processors.Add("split", func() telegraf.Processor {
		return &Split{}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		testdataPath := filepath.Join("testcases", fname)
		configFilename := filepath.Join(testdataPath, "config.toml")
		inputFilename := filepath.Join(testdataPath, "input.influx")
		expectedFilename := filepath.Join(testdataPath, "expected.out")

		t.Run(fname, func(t *testing.T) {
			// Get parser to parse input and expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			input, err := testutil.ParseMetricsFromFile(inputFilename, parser)
			require.NoError(t, err)

			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Processors, 1, "wrong number of processors")

			proc := cfg.Processors[0].Processor.(processors.HasUnwrap)
			plugin := proc.Unwrap().(*Split)
			require.NoError(t, plugin.Init())

			// Process expected metrics and compare with resulting metrics
			actual := plugin.Apply(input...)

			testutil.RequireMetricsEqual(t, expected, actual)
		})
	}
}
