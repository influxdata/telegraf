package azure_monitor

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestAggregate(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *AzureMonitor
		metrics  []telegraf.Metric
		addTime  time.Time
		pushTime time.Time
		check    func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric)
	}{
		{
			name: "add metric outside window is dropped",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "/test",
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
			addTime:  time.Unix(3600, 0),
			pushTime: time.Unix(3600, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				require.Equal(t, int64(1), plugin.MetricOutsideWindow.Get())
				require.Len(t, metrics, 0)
			},
		},
		{
			name: "metric not sent until period expires",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "/test",
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
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(0, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				require.Len(t, metrics, 0)
			},
		},
		{
			name: "add strings as dimensions",
			plugin: &AzureMonitor{
				Region:              "test",
				ResourceID:          "/test",
				StringsAsDimensions: true,
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "localhost",
					},
					map[string]interface{}{
						"value":   42,
						"message": "howdy",
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				expected := []telegraf.Metric{
					testutil.MustMetric(
						"cpu-value",
						map[string]string{
							"host":    "localhost",
							"message": "howdy",
						},
						map[string]interface{}{
							"min":   42.0,
							"max":   42.0,
							"sum":   42.0,
							"count": 1,
						},
						time.Unix(0, 0),
					),
				}
				testutil.RequireMetricsEqual(t, expected, metrics)
			},
		},
		{
			name: "add metric to cache and push",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "/test",
				cache:      make(map[time.Time]map[uint64]*aggregate, 36),
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
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				expected := []telegraf.Metric{
					testutil.MustMetric(
						"cpu-value",
						map[string]string{},
						map[string]interface{}{
							"min":   42.0,
							"max":   42.0,
							"sum":   42.0,
							"count": 1,
						},
						time.Unix(0, 0),
					),
				}

				testutil.RequireMetricsEqual(t, expected, metrics)
			},
		},
		{
			name: "added metric are aggregated",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "/test",
				cache:      make(map[time.Time]map[uint64]*aggregate, 36),
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
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 84,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 2,
					},
					time.Unix(0, 0),
				),
			},
			addTime:  time.Unix(0, 0),
			pushTime: time.Unix(3600, 0),
			check: func(t *testing.T, plugin *AzureMonitor, metrics []telegraf.Metric) {
				expected := []telegraf.Metric{
					testutil.MustMetric(
						"cpu-value",
						map[string]string{},
						map[string]interface{}{
							"min":   2.0,
							"max":   84.0,
							"sum":   128.0,
							"count": 3,
						},
						time.Unix(0, 0),
					),
				}

				testutil.RequireMetricsEqual(t, expected, metrics)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Connect()
			require.NoError(t, err)

			// Reset globals
			tt.plugin.MetricOutsideWindow.Set(0)

			tt.plugin.timeFunc = func() time.Time { return tt.addTime }
			for _, m := range tt.metrics {
				tt.plugin.Add(m)
			}

			tt.plugin.timeFunc = func() time.Time { return tt.pushTime }
			metrics := tt.plugin.Push()
			tt.plugin.Reset()

			tt.check(t, tt.plugin, metrics)
		})
	}
}

func TestWrite(t *testing.T) {
	readBody := func(r *http.Request) ([]*azureMonitorMetric, error) {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(gz)

		azmetrics := make([]*azureMonitorMetric, 0)
		for scanner.Scan() {
			line := scanner.Text()
			var amm azureMonitorMetric
			err = json.Unmarshal([]byte(line), &amm)
			if err != nil {
				return nil, err
			}
			azmetrics = append(azmetrics, &amm)
		}

		return azmetrics, nil
	}

	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	url := "http://" + ts.Listener.Addr().String() + "/metrics"

	tests := []struct {
		name    string
		plugin  *AzureMonitor
		metrics []telegraf.Metric
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "if not an azure metric nothing is sent",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "/test",
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
				t.Fatal("should not call")
			},
		},
		{
			name: "single azure metric",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "/test",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu-value",
					map[string]string{},
					map[string]interface{}{
						"min":   float64(42),
						"max":   float64(42),
						"sum":   float64(42),
						"count": int64(1),
					},
					time.Unix(0, 0),
				),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				azmetrics, err := readBody(r)
				require.NoError(t, err)
				require.Len(t, azmetrics, 1)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "multiple azure metric",
			plugin: &AzureMonitor{
				Region:     "test",
				ResourceID: "/test",
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu-value",
					map[string]string{},
					map[string]interface{}{
						"min":   float64(42),
						"max":   float64(42),
						"sum":   float64(42),
						"count": int64(1),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"cpu-value",
					map[string]string{},
					map[string]interface{}{
						"min":   float64(42),
						"max":   float64(42),
						"sum":   float64(42),
						"count": int64(1),
					},
					time.Unix(60, 0),
				),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				azmetrics, err := readBody(r)
				require.NoError(t, err)
				require.Len(t, azmetrics, 2)
				w.WriteHeader(http.StatusOK)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.handler(t, w, r)
			})

			err := tt.plugin.Connect()
			require.NoError(t, err)

			// override real authorizer and write url
			tt.plugin.auth = autorest.NullAuthorizer{}
			tt.plugin.url = url

			err = tt.plugin.Write(tt.metrics)
			require.NoError(t, err)
		})
	}
}
