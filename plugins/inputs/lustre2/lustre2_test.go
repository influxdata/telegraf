//go:build linux

package lustre2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	inputs.Add("lustre2", func() telegraf.Input {
		return &Lustre2{}
	})

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		t.Run(f.Name(), func(t *testing.T) {
			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Read the expected output if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Load the configuration
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Setup and start the plugin
			plugin := cfg.Inputs[0].Input.(*Lustre2)
			plugin.rootdir = testcasePath

			// Gather the data
			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors)

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

func TestCacheStatsReportSamplesAndPages(t *testing.T) {
	rootdir := t.TempDir()
	statsDir := filepath.Join(rootdir, "proc", "fs", "lustre", "osd-ldiskfs", "OST0001")
	require.NoError(t, os.MkdirAll(statsDir, 0750))

	stats := strings.Join([]string{
		"cache_access              14035947725 samples [pages] 1 4096 4102574238162",
		"cache_hit                 10365450287 samples [pages] 1 4096 1561298774315",
		"cache_miss                3807445393 samples [pages] 1 4096 2541275463847",
	}, "\n")
	require.NoError(t, os.WriteFile(filepath.Join(statsDir, "stats"), []byte(stats), 0640))

	plugin := &Lustre2{
		OstProcfiles: []string{"/proc/fs/lustre/osd-ldiskfs/*/stats"},
		rootdir:      rootdir,
		Log:          testutil.Logger{},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 1)
	require.Equal(t, map[string]string{"name": "OST0001"}, metrics[0].Tags())
	require.Equal(t, map[string]interface{}{
		"cache_access":         uint64(14035947725),
		"cache_access_samples": uint64(14035947725),
		"cache_access_pages":   uint64(4102574238162),
		"cache_hit":            uint64(10365450287),
		"cache_hit_samples":    uint64(10365450287),
		"cache_hit_pages":      uint64(1561298774315),
		"cache_miss":           uint64(3807445393),
		"cache_miss_samples":   uint64(3807445393),
		"cache_miss_pages":     uint64(2541275463847),
	}, metrics[0].Fields())
}
