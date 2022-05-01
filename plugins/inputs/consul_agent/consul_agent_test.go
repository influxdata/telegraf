package consul_agent

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

func TestConsulStats(t *testing.T) {
	var applyTests = []struct {
		name     string
		expected []telegraf.Metric
	}{
		{
			name: "Metrics",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"consul.rpc.request",
					map[string]string{},
					map[string]interface{}{
						"count":  int(5),
						"max":    float64(1),
						"mean":   float64(1),
						"min":    float64(1),
						"rate":   float64(0.5),
						"stddev": float64(0),
						"sum":    float64(5),
					},
					time.Unix(1639218930, 0),
					1,
				),
				testutil.MustMetric(
					"consul.consul.members.clients",
					map[string]string{
						"datacenter": "dc1",
					},
					map[string]interface{}{
						"value": float64(0),
					},
					time.Unix(1639218930, 0),
					2,
				),
				testutil.MustMetric(
					"consul.api.http",
					map[string]string{
						"method": "GET",
						"path":   "v1_agent_self",
					},
					map[string]interface{}{
						"count":  int(1),
						"max":    float64(4.14815616607666),
						"mean":   float64(4.14815616607666),
						"min":    float64(4.14815616607666),
						"rate":   float64(0.414815616607666),
						"stddev": float64(0),
						"sum":    float64(4.14815616607666),
					},
					time.Unix(1639218930, 0),
					1,
				),
			},
		},
	}

	for _, tt := range applyTests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI == "/v1/agent/metrics" {
					w.WriteHeader(http.StatusOK)
					responseKeyMetrics, _ := ioutil.ReadFile("testdata/response_key_metrics.json")
					_, err := fmt.Fprintln(w, string(responseKeyMetrics))
					require.NoError(t, err)
				}
			}))
			defer ts.Close()

			plugin := &ConsulAgent{
				URL: ts.URL,
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
