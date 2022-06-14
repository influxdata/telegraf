package vault

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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
					responseKeyMetrics, _ := ioutil.ReadFile("testdata/response_key_metrics.json")
					_, err := fmt.Fprintln(w, string(responseKeyMetrics))
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
