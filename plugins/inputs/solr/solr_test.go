package solr

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	// Register the plugin
	inputs.Add("gnmi", func() telegraf.Input {
		return &Solr{HTTPTimeout: config.Duration(5 * time.Second)}
	})

	options := []cmp.Option{
		testutil.IgnoreTime(),
		testutil.SortMetrics(),
	}

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		t.Run(fname, func(t *testing.T) {
			testdataPath := filepath.Join("testcases", fname)
			configFilename := filepath.Join(testdataPath, "telegraf.conf")
			expectedFilename := filepath.Join(testdataPath, "expected.out")

			// Load the expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Load the pages for the test case
			pages, err := loadPages(testdataPath)
			require.NoError(t, err)

			// Create a HTTP server that delivers all files in the test-directory
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, "/solr/") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				path := strings.TrimPrefix(r.URL.Path, "/solr/")
				page, found := pages[path]
				if !found {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.Write(page)
			}))
			require.NotNil(t, server)
			defer server.Close()

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Setup the plugin
			plugin := cfg.Inputs[0].Input.(*Solr)
			plugin.Servers = []string{server.URL}
			require.NoError(t, plugin.Init())

			// Gather data and compare results
			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func loadPages(path string) (map[string][]byte, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	responses := make(map[string][]byte)
	err = filepath.Walk(abspath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			// Ignore directories and files not matching the expectations
			return nil
		}
		relpath, err := filepath.Rel(abspath, path)
		if err != nil {
			return err
		}
		relpath = strings.TrimSuffix(relpath, ".json")
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		responses[relpath] = data

		return nil
	})

	return responses, err
}
