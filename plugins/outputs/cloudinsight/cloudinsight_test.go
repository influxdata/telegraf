package cloudinsight

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/samjegal/fincloud-sdk-for-go/services/cloudinsight"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	readBody := func(r *http.Request) ([]*cloudInsightMetric, error) {
		scanner := bufio.NewScanner(r.Body)

		cimetrics := make([]*cloudInsightMetric, 0)
		for scanner.Scan() {
			line := scanner.Text()
			var cim cloudInsightMetric
			err := json.Unmarshal([]byte(line), &cim)
			if err != nil {
				return nil, err
			}
			cimetrics = append(cimetrics, &cim)
		}

		return cimetrics, nil
	}

	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	tests := []struct {
		name    string
		plugin  *CloudInsight
		metrics []telegraf.Metric
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "single cloudinsight metric",
			plugin: &CloudInsight{
				Region:      "test",
				ProductName: "Custom/cpu",
				InstanceID:  "test",
				client: &cloudinsight.BaseClient{
					Client:  autorest.NewClientWithUserAgent(cloudinsight.UserAgent()),
					BaseURI: "http://" + ts.Listener.Addr().String(),
				},
				filter:   "cpu",
				timeFunc: time.Now,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				cimetrics, err := readBody(r)
				require.NoError(t, err)
				require.Len(t, cimetrics, 1)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "multiple cloudinsight metric",
			plugin: &CloudInsight{
				Region:      "test",
				ProductName: "Custom/cpu",
				InstanceID:  "test",
				client: &cloudinsight.BaseClient{
					Client:  autorest.NewClientWithUserAgent(cloudinsight.UserAgent()),
					BaseURI: "http://" + ts.Listener.Addr().String(),
				},
				filter:   "cpu",
				timeFunc: time.Now,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu-metric",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu-metric",
					map[string]string{},
					map[string]interface{}{
						"value": 38,
					},
					time.Unix(0, 0),
				),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				cimetrics, err := readBody(r)
				require.NoError(t, err)
				require.Len(t, cimetrics, 2)
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.handler(t, w, r)
			})

			err := tt.plugin.Write(tt.metrics)
			require.NoError(t, err)
		})
	}
}
