package azure_loganalytics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func getMetric(name string) telegraf.Metric {
	m, err := metric.New(
		name,
		map[string]string{
			"host": "test-machine",
			"env":  "development",
		},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	if err != nil {
		panic(err)
	}
	return m
}

func TestMissingCustomerId(t *testing.T) {
	plugin := &AzLogAnalytics{}

	err := plugin.Connect()
	require.Error(t, err)
}

func TestMissingSharedKey(t *testing.T) {
	plugin := &AzLogAnalytics{
		CustomerID: "dummy_id",
	}

	err := plugin.Connect()
	require.Error(t, err)
}

func TestInvalidNamespacePrefix(t *testing.T) {
	plugin := &AzLogAnalytics{
		CustomerID:      "dummy_id",
		SharedKey:       "ZHVtbXlfa2V5",
		NamespacePrefix: "Testing_",
	}

	err := plugin.Connect()
	require.Error(t, err)
}

func TestInvalidNamespacePrefixLength(t *testing.T) {
	plugin := &AzLogAnalytics{
		CustomerID:      "dummy_id",
		SharedKey:       "ZHVtbXlfa2V5",
		NamespacePrefix: "ThisIsALongStringToTestTheMaxLength",
	}

	err := plugin.Connect()
	require.Error(t, err)
}

func TestValid(t *testing.T) {
	plugin := &AzLogAnalytics{
		CustomerID: "dummy_id",
		SharedKey:  "ZHVtbXlfa2V5",
	}

	err := plugin.Connect()
	require.NoError(t, err)
}

func TestStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	url, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name       string
		plugin     *AzLogAnalytics
		statusCode int
		errFunc    func(t *testing.T, err error)
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			errFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:       "400 status is an error",
			statusCode: http.StatusBadRequest,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:       "403 status is an error",
			statusCode: http.StatusForbidden,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:       "404 status is an error",
			statusCode: http.StatusNotFound,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:       "429 status is an error",
			statusCode: http.StatusTooManyRequests,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:       "500 status is an error",
			statusCode: http.StatusInternalServerError,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:       "503 status is an error",
			statusCode: http.StatusServiceUnavailable,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	plugin := &AzLogAnalytics{
		CustomerID: "dummy_id",
		SharedKey:  "ZHVtbXlfa2V5",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			err = plugin.Connect()
			require.NoError(t, err)

			// override write url
			plugin.serviceURL = url.String()

			err = plugin.Write([]telegraf.Metric{getMetric("cpu")})
			tt.errFunc(t, err)
		})
	}
}

func TestWrite(t *testing.T) {
	readBody := func(r *http.Request) ([]map[string]interface{}, error) {
		buffer, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		var amm []map[string]interface{}
		err = json.Unmarshal(buffer, &amm)
		if err != nil {
			return nil, err
		}

		return amm, nil
	}

	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	url, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name    string
		metrics []telegraf.Metric
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "metric with different names sent in different requests",
			metrics: []telegraf.Metric{
				getMetric("cpu"),
				getMetric("mem"),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				azmetrics, err := readBody(r)
				require.NoError(t, err)
				require.Len(t, azmetrics, 1)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "single azure metric",
			metrics: []telegraf.Metric{
				getMetric("cpu"),
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
			metrics: []telegraf.Metric{
				getMetric("cpu"),
				getMetric("cpu"),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				azmetrics, err := readBody(r)
				require.NoError(t, err)
				require.Len(t, azmetrics, 2)
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	plugin := &AzLogAnalytics{
		CustomerID: "dummy_id",
		SharedKey:  "ZHVtbXlfa2V5",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.handler(t, w, r)
			})

			err := plugin.Connect()
			require.NoError(t, err)

			// override write url
			plugin.serviceURL = url.String()

			err = plugin.Write(tt.metrics)
			require.NoError(t, err)
		})
	}
}
