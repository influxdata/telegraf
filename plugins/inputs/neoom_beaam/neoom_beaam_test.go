package neoom_beaam

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
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
	inputs.Add("neoom_beaam", func() telegraf.Input {
		return &NeoomBeaam{}
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
		inputFiles := filepath.Join(testcasePath, "*.json")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		t.Run(f.Name(), func(t *testing.T) {
			// Read the input data
			matches, err := filepath.Glob(inputFiles)
			require.NoError(t, err)
			require.NotEmpty(t, matches)
			sort.Strings(matches)
			endpoints := make(map[string][]byte, len(matches))
			for _, fn := range matches {
				buf, err := os.ReadFile(fn)
				require.NoError(t, err)
				key := strings.TrimSuffix(filepath.Base(fn), filepath.Ext(fn))
				if strings.HasPrefix(key, "thing_") {
					endpoints["/api/v1/things/"+strings.TrimPrefix(key, "thing_")+"/states"] = buf
				} else {
					endpoints["/api/v1/site/"+key] = buf
				}
			}

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

			// Create a fake API server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if msg, ok := endpoints[r.URL.Path]; ok {
					if _, err := w.Write(msg); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			// Load the configuration
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Setup and start the plugin
			plugin := cfg.Inputs[0].Input.(*NeoomBeaam)
			plugin.Address = server.URL
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Start(nil))
			defer plugin.Stop()

			// Gather the data
			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors)

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
		})
	}
}
