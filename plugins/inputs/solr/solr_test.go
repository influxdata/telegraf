package solr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

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
			expectedFilename := filepath.Join(testdataPath, "expected.out")

			// Load the expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Create a HTTP server that delivers all files in the test-directory
			handler, err := createRequestHandler(testdataPath)
			require.NoError(t, err)
			server := httptest.NewServer(handler)
			require.NotNil(t, server)
			defer server.Close()

			plugin := &Solr{
				Servers:     []string{server.URL},
				HTTPTimeout: config.Duration(time.Second * 5),
			}
			//require.NoError(t, plugin.Init())

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
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
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

func TestSolr3GatherStats(t *testing.T) {
	ts := createMockSolr3Server(t)
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
	var acc testutil.Accumulator
	require.NoError(t, solr.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminMainCoreStatusExpected,
		map[string]string{"core": "main"})

	acc.AssertContainsTaggedFields(t, "solr_admin",
		solrAdminCore1StatusExpected,
		map[string]string{"core": "core1"})

	acc.AssertContainsTaggedFields(t, "solr_core",
		solr3CoreExpected,
		map[string]string{"core": "main", "handler": "searcher"})

	acc.AssertContainsTaggedFields(t, "solr_queryhandler",
		solr3QueryHandlerExpected,
		map[string]string{"core": "main", "handler": "org.apache.solr.handler.component.SearchHandler"})

	acc.AssertContainsTaggedFields(t, "solr_updatehandler",
		solr3UpdateHandlerExpected,
		map[string]string{"core": "main", "handler": "updateHandler"})

	acc.AssertContainsTaggedFields(t, "solr_cache",
		solr3CacheExpected,
		map[string]string{"core": "main", "handler": "filterCache"})
}
func TestNoCoreDataHandling(t *testing.T) {
	ts := createMockNoCoreDataServer(t)
	solr := NewSolr()
	solr.Servers = []string{ts.URL}
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

func createRequestHandler(path string) (http.HandlerFunc, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/solr/") {
			w.WriteHeader(http.StatusNotFound)
		}

		// Construct the path
		file := strings.TrimPrefix(r.URL.Path, "/solr/")
		file = filepath.Join(abspath, filepath.FromSlash(file))
		file, err = filepath.Abs(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		file += ".json"

		// Prevent directory traversal
		if !strings.HasPrefix(file, abspath) {
			w.WriteHeader(http.StatusBadRequest)
		}

		http.ServeFile(w, r, file)
	}, nil
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

func createMockSolr3Server(t *testing.T) *httptest.Server {
	data := readJSONAsString(t, "testdata/m_beans_solr3_main_response.json")
	statusResponse := readJSONAsString(t, "testdata/status_response.json")

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/solr/admin/cores") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, statusResponse)
		} else if strings.Contains(r.URL.Path, "solr/main/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, data)
		} else if strings.Contains(r.URL.Path, "solr/core1/admin") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, data)
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
