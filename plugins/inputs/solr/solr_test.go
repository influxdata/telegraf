package solr

import (
	"fmt"
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

func TestGatherStats(t *testing.T) {
	ts := createMockServer(t)
	solr := &Solr{
		Servers:     []string{ts.URL},
		HTTPTimeout: config.Duration(time.Second * 5),
	}
	require.NoError(t, solr.Init())

	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminMainCoreStatusExpected,
		map[string]string{"core": "main"})

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminCore1StatusExpected,
		map[string]string{"core": "core1"})

	acc.AssertContainsTaggedFields(t, "solr_core",
		solrCoreExpected,
		map[string]string{"core": "main", "handler": "searcher"})

	acc.AssertContainsTaggedFields(t, "solr_queryhandler",
		solrQueryHandlerExpected,
		map[string]string{"core": "main", "handler": "org.apache.solr.handler.component.SearchHandler"})

	acc.AssertContainsTaggedFields(t, "solr_updatehandler",
		solrUpdateHandlerExpected,
		map[string]string{"core": "main", "handler": "updateHandler"})

	acc.AssertContainsTaggedFields(t, "solr_cache",
		solrCacheExpected,
		map[string]string{"core": "main", "handler": "filterCache"})
}

func TestNoCoreDataHandling(t *testing.T) {
	ts := createMockNoCoreDataServer(t)
	solr := &Solr{
		Servers:     []string{ts.URL},
		HTTPTimeout: config.Duration(time.Second * 5),
	}
	require.NoError(t, solr.Init())

	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminMainCoreStatusExpected,
		map[string]string{"core": "main"})

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminCore1StatusExpected,
		map[string]string{"core": "core1"})

	acc.AssertDoesNotContainMeasurement(t, "solr_core")
	acc.AssertDoesNotContainMeasurement(t, "solr_queryhandler")
	acc.AssertDoesNotContainMeasurement(t, "solr_updatehandler")
	acc.AssertDoesNotContainMeasurement(t, "solr_handler")
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

func createMockServer(t *testing.T) *httptest.Server {
	statusResponse := readJSONAsString(t, "testdata/status_response.json")
	mBeansMainResponse := readJSONAsString(t, "testdata/m_beans_main_response.json")
	mBeansCore1Response := readJSONAsString(t, "testdata/m_beans_core1_response.json")

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/solr/admin/cores") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, statusResponse)
		} else if strings.Contains(r.URL.Path, "solr/main/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, mBeansMainResponse)
		} else if strings.Contains(r.URL.Path, "solr/core1/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, mBeansCore1Response)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "nope")
		}
	}))
}

func createMockNoCoreDataServer(t *testing.T) *httptest.Server {
	var nodata string
	statusResponse := readJSONAsString(t, "testdata/status_response.json")

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/solr/admin/cores") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, statusResponse)
		} else if strings.Contains(r.URL.Path, "solr/main/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, nodata)
		} else if strings.Contains(r.URL.Path, "solr/core1/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, nodata)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "nope")
		}
	}))
}

func readJSONAsString(t *testing.T, jsonFilePath string) string {
	data, err := os.ReadFile(jsonFilePath)
	require.NoErrorf(t, err, "could not read from JSON file %s", jsonFilePath)

	return string(data)
}
