package vault

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestVaultStats(t *testing.T) {
	var applyTests = []struct {
		name     string
		expected []telegraf.Metric
	}{
		{
			name: "Metrics",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"vault.raft.replication.appendEntries.logs",
					map[string]string{
						"peer_id": "clustnode-02",
					},
					map[string]interface{}{
						"count":  int(130),
						"rate":   float64(0.2),
						"sum":    int(2),
						"min":    int(0),
						"max":    int(1),
						"mean":   float64(0.015384615384615385),
						"stddev": float64(0.12355304447984486),
					},
					time.Unix(1638287340, 0),
					1,
				),
				testutil.MustMetric(
					"vault.core.unsealed",
					map[string]string{
						"cluster": "vault-cluster-23b671c7",
					},
					map[string]interface{}{
						"value": int(1),
					},
					time.Unix(1638287340, 0),
					2,
				),
				testutil.MustMetric(
					"vault.token.lookup",
					map[string]string{},
					map[string]interface{}{
						"count":  int(5135),
						"max":    float64(16.22449493408203),
						"mean":   float64(0.1698389152269865),
						"min":    float64(0.06690400093793869),
						"rate":   float64(87.21228296905755),
						"stddev": float64(0.24637634000854705),
						"sum":    float64(872.1228296905756),
					},
					time.Unix(1638287340, 0),
					1,
				),
			},
		},
	}

	for _, tt := range applyTests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI == "/v1/sys/metrics" {
					w.WriteHeader(http.StatusOK)
					responseKeyMetrics, err := os.ReadFile("testdata/response_key_metrics.json")
					require.NoError(t, err)
					_, err = fmt.Fprintln(w, string(responseKeyMetrics))
					require.NoError(t, err)
				}
			}))
			defer ts.Close()

			plugin := &Vault{
				URL:   ts.URL,
				Token: "s.CDDrgg5zPv5ssI0Z2P4qxJj2",
			}
			err := plugin.Init()
			require.NoError(t, err)

			acc := testutil.Accumulator{}
			err = plugin.Gather(&acc)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestRedirect(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"vault.raft.replication.appendEntries.logs",
			map[string]string{
				"peer_id": "clustnode-02",
			},
			map[string]interface{}{
				"count":  int(130),
				"rate":   float64(0.2),
				"sum":    int(2),
				"min":    int(0),
				"max":    int(1),
				"mean":   float64(0.015384615384615385),
				"stddev": float64(0.12355304447984486),
			},
			time.Unix(1638287340, 0),
			1,
		),
		testutil.MustMetric(
			"vault.core.unsealed",
			map[string]string{
				"cluster": "vault-cluster-23b671c7",
			},
			map[string]interface{}{
				"value": int(1),
			},
			time.Unix(1638287340, 0),
			2,
		),
		testutil.MustMetric(
			"vault.token.lookup",
			map[string]string{},
			map[string]interface{}{
				"count":  int(5135),
				"max":    float64(16.22449493408203),
				"mean":   float64(0.1698389152269865),
				"min":    float64(0.06690400093793869),
				"rate":   float64(87.21228296905755),
				"stddev": float64(0.24637634000854705),
				"sum":    float64(872.1228296905756),
			},
			time.Unix(1638287340, 0),
			1,
		),
	}

	response, err := os.ReadFile("testdata/response_key_metrics.json")
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/v1/sys/metrics":
			redirectURL := "http://" + r.Host + "/custom/metrics"
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		case "/custom/metrics":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(response)
		}
	}))
	defer server.Close()

	// Setup the plugin
	plugin := &Vault{
		URL:   server.URL,
		Token: "s.CDDrgg5zPv5ssI0Z2P4qxJj2",
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the docker container
	container := testutil.Container{
		Image:        "vault:1.13.3",
		ExposedPorts: []string{"8200"},
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID": "root",
		},
		HostConfigModifier: func(hc *dockercontainer.HostConfig) {
			hc.CapAdd = []string{"IPC_LOCK"}
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("Root Token: root"),
			wait.ForListeningPort(nat.Port("8200")),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Setup the plugin
	port := container.Ports["8200"]
	plugin := &Vault{
		URL:   "http://" + container.Address + ":" + port,
		Token: "root",
	}
	require.NoError(t, plugin.Init())

	// Collect the metrics and compare
	var acc testutil.Accumulator
	require.Eventually(t, func() bool {
		require.NoError(t, plugin.Gather(&acc))
		return len(acc.GetTelegrafMetrics()) > 50
	}, 5*time.Second, 100*time.Millisecond)
}
