package lookup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	plugin := &Processor{}
	require.ErrorContains(t, plugin.Init(), "missing 'files'")

	plugin = &Processor{
		Filenames: []string{"blah.json"},
	}
	require.ErrorContains(t, plugin.Init(), "missing 'key_template'")

	plugin = &Processor{
		Filenames:   []string{"blah.json"},
		KeyTemplate: "lala",
	}
	require.ErrorIs(t, plugin.Init(), os.ErrNotExist)

	plugin = &Processor{
		Filenames:   []string{"blah.json"},
		Fileformat:  "foo",
		KeyTemplate: "lala",
	}
	require.ErrorContains(t, plugin.Init(), "invalid format")
}

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	// Set up for file inputs
	processors.Add("lookup", func() telegraf.Processor {
		return &Processor{Log: testutil.Logger{}}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		testdataPath := filepath.Join("testcases", fname)
		configFilename := filepath.Join(testdataPath, "telegraf.conf")
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

			type unwrappable interface{ Unwrap() telegraf.Processor }
			proc := cfg.Processors[0].Processor.(unwrappable)
			plugin := proc.Unwrap().(*Processor)
			require.NoError(t, plugin.Init())

			// Process expected metrics and compare with resulting metrics
			actual := plugin.Apply(input...)
			testutil.RequireMetricsEqual(t, expected, actual)
		})
	}
}

func TestCasesTracking(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	// Set up for file inputs
	processors.Add("lookup", func() telegraf.Processor {
		return &Processor{Log: testutil.Logger{}}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		testdataPath := filepath.Join("testcases", fname)
		configFilename := filepath.Join(testdataPath, "telegraf.conf")
		inputFilename := filepath.Join(testdataPath, "input.influx")
		expectedFilename := filepath.Join(testdataPath, "expected.out")

		t.Run(fname, func(t *testing.T) {
			// Get parser to parse input and expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			inputRaw, err := testutil.ParseMetricsFromFile(inputFilename, parser)
			require.NoError(t, err)
			input := make([]telegraf.Metric, 0, len(inputRaw))
			notify := func(_ telegraf.DeliveryInfo) {}
			for _, m := range inputRaw {
				tm, _ := metric.WithTracking(m, notify)
				input = append(input, tm)
			}

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

			type unwrappableProcessor interface{ Unwrap() telegraf.Processor }
			proc := cfg.Processors[0].Processor.(unwrappableProcessor)
			plugin := proc.Unwrap().(*Processor)
			require.NoError(t, plugin.Init())

			// Process expected metrics and compare with resulting metrics
			actual := plugin.Apply(input...)
			testutil.RequireMetricsEqual(t, expected, actual)
		})
	}
}
