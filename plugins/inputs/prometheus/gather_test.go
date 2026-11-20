package prometheus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
)

const exposition = `# TYPE requests_total counter
requests_total{url="from_label",address="from_label",static="from_label",path="/a"} 10
# TYPE temperature gauge
temperature{url="from_label",address="from_label",static="from_label"} 21.5
# TYPE latency summary
latency{url="from_label",quantile="0.5"} 0.2
latency_sum{url="from_label"} 12
latency_count{url="from_label"} 60
# TYPE size histogram
size_bucket{url="from_label",le="100"} 3
size_bucket{url="from_label",le="+Inf"} 5
size_sum{url="from_label"} 420
size_count{url="from_label"} 5
# TYPE build_info untyped
build_info{url="from_label",version="1.2.3"} 1
`

// TestGatherURLTagsAndTypes checks, through the agent accumulator, that the
// url tag, the address tag and the discovery tags override labels of the same
// name, with the discovery tags taking precedence over the address tag, that
// credentials never reach the url tag and that the value types set by the
// parser are kept.
func TestGatherURLTagsAndTypes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		if _, err := w.Write([]byte(exposition)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	for _, version := range []int{1, 2} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			p := &Prometheus{
				URLTag:        "url",
				MetricVersion: version,
				Log:           testutil.Logger{},
			}
			require.NoError(t, p.Init())

			target, err := url.Parse(ts.URL + "/metrics")
			require.NoError(t, err)
			target.User = url.UserPassword("user", "secret")
			original := *target
			u := urlAndAddress{
				url:         target,
				originalURL: &original,
				address:     "10.0.0.1",
				tags:        map[string]string{"static": "from_discovery", "address": "from_discovery"},
			}

			ch := make(chan telegraf.Metric, 100)
			acc := agent.NewAccumulator(models.NewRunningInput(p, &models.InputConfig{Name: "prometheus"}), ch)
			_, _, err = p.gatherURL(u, acc)
			require.NoError(t, err)
			close(ch)

			expectedURL := ts.URL + "/metrics"
			types := make(map[string]telegraf.ValueType)
			var count int
			for m := range ch {
				count++
				tags := m.Tags()
				require.Equal(t, expectedURL, tags["url"], "metric %s", m.Name())
				require.NotContains(t, tags["url"], "secret")
				require.Equal(t, "from_discovery", tags["address"], "metric %s", m.Name())
				require.Equal(t, "from_discovery", tags["static"], "metric %s", m.Name())
				name := m.Name()
				if version == 2 {
					for _, f := range m.FieldList() {
						name = f.Key
					}
				}
				types[strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, "_bucket"), "_sum"), "_count")] = m.Type()
			}
			require.NotZero(t, count)
			require.Equal(t, telegraf.Counter, types["requests_total"])
			require.Equal(t, telegraf.Gauge, types["temperature"])
			require.Equal(t, telegraf.Summary, types["latency"])
			require.Equal(t, telegraf.Histogram, types["size"])
			require.Equal(t, telegraf.Untyped, types["build_info"])
		})
	}
}
