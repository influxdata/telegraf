package nomad

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

func TestNomadStats(t *testing.T) {
	var applyTests = []struct {
		name     string
		expected []telegraf.Metric
	}{
		{
			name: "Metrics",
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"nomad.nomad.rpc.query",
					map[string]string{
						"host": "node1",
					},
					map[string]interface{}{
						"count": int(7),
						"max":   float64(1),
						"min":   float64(1),
						"mean":  float64(1),
						"rate":  float64(0.7),
						"sum":   float64(7),
						"sumsq": float64(0),
					},
					time.Unix(1636843140, 0),
					1,
				),
				testutil.MustMetric(
					"nomad.client.allocated.cpu",
					map[string]string{
						"node_scheduling_eligibility": "eligible",
						"host":                        "node1",
						"node_id":                     "2bbff078-8473-a9de-6c5e-42b4e053e12f",
						"datacenter":                  "dc1",
						"node_class":                  "none",
						"node_status":                 "ready",
					},
					map[string]interface{}{
						"value": float32(500),
					},
					time.Unix(1636843140, 0),
					2,
				),
				testutil.MustMetric(
					"nomad.memberlist.gossip",
					map[string]string{
						"host": "node1",
					},
					map[string]interface{}{
						"count":  int(20),
						"max":    float64(0.03747599944472313),
						"mean":   float64(0.013159099989570678),
						"min":    float64(0.003459000028669834),
						"rate":   float64(0.026318199979141355),
						"stddev": float64(0.009523742715522742),
						"sum":    float64(0.26318199979141355),
						"sumsq":  float64(0),
					},
					time.Unix(1636843140, 0),
					1,
				),
			},
		},
	}

	for _, tt := range applyTests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RequestURI == "/v1/metrics" {
					w.WriteHeader(http.StatusOK)
					responseKeyMetrics, _ := ioutil.ReadFile("testdata/response_key_metrics.json")
					_, err := fmt.Fprintln(w, string(responseKeyMetrics))
					require.NoError(t, err)
				}
			}))
			defer ts.Close()

			plugin := &Nomad{
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
