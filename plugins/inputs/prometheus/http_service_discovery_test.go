package prometheus

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

func TestHttpSD(t *testing.T) {
	testcasePath := filepath.Join("testcases", "service_discovery")
	configFilename := filepath.Join(testcasePath, "telegraf.conf")
	expectedResult := filepath.Join(testcasePath, "http-services.json")

	// read expected result
	result, err := os.ReadFile(expectedResult)
	require.NoError(t, err)

	// Create a fake API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a GET request
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(result); err != nil {
			t.Errorf("Failed to write response: %v", err)
			return
		}
	}))
	defer server.Close()

	// Load the configuration
	cfg := config.NewConfig()
	require.NoError(t, cfg.LoadConfig(configFilename))
	require.Len(t, cfg.Inputs, 1)

	// Setup and start the plugin
	plugin := cfg.Inputs[0].Input.(*Prometheus)
	plugin.HTTPSDConfig.URL = server.URL
	require.NoError(t, plugin.Init())

	// refresh http services
	client := &http.Client{}
	defer client.CloseIdleConnections()
	require.NoError(t, plugin.refreshHTTPServices(server.URL, client))

	plugin.lock.Lock()
	defer plugin.lock.Unlock()
	// check we have 8 http services
	require.Len(t, plugin.httpServices, 8)
}
