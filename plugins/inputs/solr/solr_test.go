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

	"github.com/docker/go-connections/nat"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

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
				_, _ = w.Write(page)
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
			require.NoError(t, plugin.Start(&acc))
			require.NoError(t, plugin.Gather(&acc))
			plugin.Stop()

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get all integration test files in testcases
	resultFiles, err := filepath.Glob(filepath.Join("testcases", "*.result"))
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, resultFiles)

	options := []cmp.Option{
		testutil.IgnoreTime(),
		testutil.SortMetrics(),
	}

	const servicePort = "8983"

	for _, f := range resultFiles {
		fname := strings.TrimSuffix(filepath.Base(f), ".result")
		t.Run(fname, func(t *testing.T) {
			expectedFilename := filepath.Join("testcases", fname+".result")

			// Load the expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Determine container version for the integration test
			// The version number is the last element in the filename separated
			// by a dash and prefixed with a 'v'.
			image := "solr"
			parts := strings.Split(fname, "-")
			if len(parts) > 1 {
				version := parts[len(parts)-1]
				require.True(t, strings.HasPrefix(version, "v"))
				image += ":" + strings.TrimPrefix(version, "v")
			}

			// Start the container
			container := testutil.Container{
				Image:        image,
				ExposedPorts: []string{servicePort},
				Cmd:          []string{"solr-precreate", "main"},
				WaitingFor: wait.ForAll(
					wait.ForListeningPort(nat.Port(servicePort)),
					wait.ForLog("Registered new searcher"),
				),
			}
			require.NoError(t, container.Start(), "failed to start container")
			defer container.Terminate()

			server := []string{fmt.Sprintf("http://%s:%s", container.Address, container.Ports[servicePort])}

			// Setup the plugin
			plugin := &Solr{
				Servers:     server,
				HTTPTimeout: config.Duration(5 * time.Second),
				Log:         &testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Gather data and compare results
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			require.NoError(t, plugin.Gather(&acc))
			plugin.Stop()

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsStructureEqual(t, expected, actual, options...)
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
		responses[filepath.ToSlash(relpath)] = data

		return nil
	})

	return responses, err
}
