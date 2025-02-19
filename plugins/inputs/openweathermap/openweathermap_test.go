package openweathermap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestFormatURL(t *testing.T) {
	n := &OpenWeatherMap{
		AppID:   "appid",
		Units:   "metric",
		Lang:    "de",
		BaseURL: "http://foo.com",
	}
	require.NoError(t, n.Init())

	require.Equal(t,
		"http://foo.com/data/2.5/forecast?APPID=appid&id=12345&lang=de&units=metric",
		n.formatURL("/data/2.5/forecast", "12345"))
}

func TestDefaultUnits(t *testing.T) {
	n := &OpenWeatherMap{}
	require.NoError(t, n.Init())

	require.Equal(t, "metric", n.Units)
}

func TestDefaultLang(t *testing.T) {
	n := &OpenWeatherMap{}
	require.NoError(t, n.Init())

	require.Equal(t, "en", n.Lang)
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("openweathermap", func() telegraf.Input {
		return &OpenWeatherMap{
			ResponseTimeout: config.Duration(5 * time.Second),
		}
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
			// Read the input data
			input, err := readInputData(testcasePath)
			require.NoError(t, err)

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

			// Start the test-server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Lookup the response
				key := strings.TrimPrefix(r.URL.Path, "/data/2.5/")
				if resp, found := input[key]; found {
					w.Header()["Content-Type"] = []string{"application/json"}
					if _, err := w.Write(resp); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						t.Error(err)
					}
					return
				}

				// Try to append the key and find the response
				ids := strings.Split(r.URL.Query().Get("id"), ",")
				if len(ids) > 1 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if len(ids) == 1 {
					key += "_" + ids[0]
					if resp, found := input[key]; found {
						w.Header()["Content-Type"] = []string{"application/json"}
						if _, err := w.Write(resp); err != nil {
							w.WriteHeader(http.StatusInternalServerError)
							t.Error(err)
						}
						return
					}
				}

				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Fake the reading
			plugin := cfg.Inputs[0].Input.(*OpenWeatherMap)
			plugin.BaseURL = server.URL
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))
			if len(acc.Errors) > 0 {
				var actualErrorMsgs []string
				for _, err := range acc.Errors {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
				require.ElementsMatch(t, actualErrorMsgs, expectedErrors)
			}

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
		})
	}
}

func readInputData(path string) (map[string][]byte, error) {
	pattern := filepath.Join(path, "response_*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Iterate over the response_*.json files and read the into the data map
	data := make(map[string][]byte, len(matches))
	for _, filename := range matches {
		resp, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("reading response %q failed: %w", filename, err)
		}
		key := filepath.Base(filename)
		key = strings.TrimPrefix(key, "response_")
		key = strings.TrimSuffix(key, ".json")
		data[key] = resp
	}
	return data, nil
}
