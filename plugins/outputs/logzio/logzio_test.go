package logzio

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const (
	logzioTestToken = "123456789"
)

func TestConnectWithDefaults(t *testing.T) {
	l := Logzio{
		Token: logzioTestToken,
	}

	err := l.Connect()
	require.NoError(t, err)
	require.Equal(t, l.Timeout.Duration, defaultLogzioRequestTimeout)
	require.Equal(t, l.URL, defaultLogzioURL)
}

func TestConnectWithoutToken(t *testing.T) {
	l := Logzio{}

	err := l.Connect()
	require.Error(t, err)
}

func TestRequestHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))

	l := Logzio{
		Token: logzioTestToken,
		URL:   ts.URL,
	}

	err := l.Connect()
	require.NoError(t, err)

	err = l.Write(testutil.MockMetrics())
	require.NoError(t, err)

}

func TestStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name       string
		plugin     *Logzio
		statusCode int
		errFunc    func(t *testing.T, err error)
	}{
		{
			name: "success",
			plugin: &Logzio{
				URL:   u.String(),
				Token: logzioTestToken,
			},
			statusCode: http.StatusOK,
			errFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "1xx status is an error",
			plugin: &Logzio{
				URL:   u.String(),
				Token: logzioTestToken,
			},
			statusCode: 103,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "3xx status is an error",
			plugin: &Logzio{
				URL:   u.String(),
				Token: logzioTestToken,
			},
			statusCode: http.StatusMultipleChoices,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "4xx status is an error",
			plugin: &Logzio{
				URL:   u.String(),
				Token: logzioTestToken,
			},
			statusCode: http.StatusMultipleChoices,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "5xx status is an error",
			plugin: &Logzio{
				URL:   u.String(),
				Token: logzioTestToken,
			},
			statusCode: http.StatusServiceUnavailable,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			err = tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write(testutil.MockMetrics())
			tt.errFunc(t, err)
		})
	}
}

func TestWrite(t *testing.T) {
	readBody := func(r *http.Request) ([]map[string]interface{}, error) {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(gz)

		metrics := make([]map[string]interface{}, 0)
		for scanner.Scan() {
			line := scanner.Text()
			var m map[string]interface{}
			err = json.Unmarshal([]byte(line), &m)
			if err != nil {
				return nil, err
			}
			metrics = append(metrics, m)
		}

		return metrics, nil
	}

	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name    string
		plugin  *Logzio
		metrics []telegraf.Metric
		errFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "single metric - no value type",
			plugin: &Logzio{
				URL:   u.String(),
				Token: logzioTestToken,
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
			errFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				metrics, err := readBody(r)
				require.NoError(t, err)
				require.Len(t, metrics, 1)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "multiple metric",
			plugin: &Logzio{
				URL:   u.String(),
				Token: logzioTestToken,
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
					telegraf.Histogram,
				),
			},
			errFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				metrics, err := readBody(r)
				require.NoError(t, err)
				require.Len(t, metrics, 2)
				w.WriteHeader(http.StatusOK)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.errFunc(t, w, r)
			})

			err := tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write(tt.metrics)
			require.NoError(t, err)
		})
	}
}
