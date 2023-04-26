package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	// Set up for file inputs
	secretstores.Add("http", func(id string) telegraf.SecretStore {
		return &HTTP{Log: testutil.Logger{}}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		t.Run(fname, func(t *testing.T) {
			testdataPath := filepath.Join("testcases", fname)
			configFilename := filepath.Join(testdataPath, "telegraf.conf")
			inputFilename := filepath.Join(testdataPath, "secrets.json")
			expectedFilename := filepath.Join(testdataPath, "expected.json")

			// Read the input data
			input, err := os.ReadFile(inputFilename)
			require.NoError(t, err)

			// Read the expected output data
			buf, err := os.ReadFile(expectedFilename)
			require.NoError(t, err)
			var expected map[string]string
			require.NoError(t, json.Unmarshal(buf, &expected))

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.NotEmpty(t, cfg.SecretStores)

			var id string
			var plugin telegraf.SecretStore
			actual := make(map[string]string, len(expected))
			for id, plugin = range cfg.SecretStores {
				// Setup dummy server and redirect the plugin's URL to that dummy
				httpPlugin, ok := plugin.(*HTTP)
				require.True(t, ok)

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/secrets" {
						_, _ = w.Write([]byte(input))
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer server.Close()

				us, err := url.Parse(server.URL)
				require.NoError(t, err)

				u, err := url.Parse(httpPlugin.URL)
				require.NoError(t, err)
				u.Host = us.Host
				httpPlugin.URL = u.String()
				require.NoError(t, httpPlugin.download())

				// Retrieve the secrets from the plugin
				keys, err := plugin.List()
				require.NoError(t, err)

				for _, k := range keys {
					v, err := plugin.Get(k)
					require.NoError(t, err)
					actual[id+"."+k] = string(v)
				}
			}
			require.EqualValues(t, expected, actual)
		})
	}
}
