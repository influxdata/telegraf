package sumologic

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/carbon2"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/prometheus"
)

func getMetric(t *testing.T) telegraf.Metric {
	m, err := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	require.NoError(t, err)
	return m
}

func getMetrics(t *testing.T, count int) []telegraf.Metric {
	var metrics = make([]telegraf.Metric, count)

	for i := 0; i < count; i++ {
		m, err := metric.New(
			fmt.Sprintf("cpu-%d", i),
			map[string]string{
				"ec2_instance": "aws-129038123",
				"image":        "aws-ami-1234567890",
			},
			map[string]interface{}{
				"idle":   5876876,
				"steal":  5876876,
				"system": 5876876,
				"user":   5876876,
				"temp":   70.0,
			},
			time.Unix(0, 0),
		)
		require.NoError(t, err)
		metrics[i] = m
	}
	return metrics
}

func TestInvalidMethod(t *testing.T) {
	plugin := &SumoLogic{
		URL:    "",
		Method: http.MethodGet,
	}

	err := plugin.Connect()
	require.Error(t, err)
}

func TestMethod(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name           string
		plugin         func() *SumoLogic
		expectedMethod string
		connectError   bool
	}{
		{
			name: "default method is POST",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				return s
			},
			expectedMethod: http.MethodPost,
		},
		{
			name: "put is okay",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.Method = http.MethodPut
				return s
			},
			expectedMethod: http.MethodPut,
		},
		{
			name: "get is invalid",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.Method = http.MethodGet
				return s
			},
			connectError: true,
		},
		{
			name: "method is case insensitive",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.Method = "poST"
				return s
			},
			expectedMethod: http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.expectedMethod, r.Method)
				w.WriteHeader(http.StatusOK)
			})

			serializer, err := carbon2.NewSerializer(string(carbon2.Carbon2FormatFieldSeparate))
			require.NoError(t, err)

			plugin := tt.plugin()
			plugin.SetSerializer(serializer)
			err = plugin.Connect()
			if tt.connectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			err = plugin.Write([]telegraf.Metric{getMetric(t)})
			require.NoError(t, err)
		})
	}
}

func TestStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	pluginFn := func() *SumoLogic {
		s := Default()
		s.URL = u.String()
		return s
	}

	tests := []struct {
		name       string
		plugin     *SumoLogic
		statusCode int
		errFunc    func(t *testing.T, err error)
	}{
		{
			name:       "success",
			plugin:     pluginFn(),
			statusCode: http.StatusOK,
			errFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:       "1xx status is an error",
			plugin:     pluginFn(),
			statusCode: http.StatusSwitchingProtocols,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:       "3xx status is an error",
			plugin:     pluginFn(),
			statusCode: http.StatusMultipleChoices,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:       "4xx status is an error",
			plugin:     pluginFn(),
			statusCode: http.StatusBadRequest,
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

			serializer, err := carbon2.NewSerializer(string(carbon2.Carbon2FormatFieldSeparate))
			require.NoError(t, err)

			tt.plugin.SetSerializer(serializer)
			err = tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write([]telegraf.Metric{getMetric(t)})
			tt.errFunc(t, err)
		})
	}
}

func TestContentType(t *testing.T) {
	tests := []struct {
		name         string
		plugin       func() *SumoLogic
		expectedBody []byte
	}{
		{
			name: "carbon2 (data format = field separate) is supported",
			plugin: func() *SumoLogic {
				s := Default()
				s.headers = map[string]string{
					contentTypeHeader: carbon2ContentType,
				}
				sr, err := carbon2.NewSerializer(string(carbon2.Carbon2FormatFieldSeparate))
				require.NoError(t, err)
				s.SetSerializer(sr)
				return s
			},
			expectedBody: []byte("metric=cpu field=value  42 0\n"),
		},
		{
			name: "carbon2 (data format = metric includes field) is supported",
			plugin: func() *SumoLogic {
				s := Default()
				s.headers = map[string]string{
					contentTypeHeader: carbon2ContentType,
				}
				sr, err := carbon2.NewSerializer(string(carbon2.Carbon2FormatMetricIncludesField))
				require.NoError(t, err)
				s.SetSerializer(sr)
				return s
			},
			expectedBody: []byte("metric=cpu_value  42 0\n"),
		},
		{
			name: "graphite is supported",
			plugin: func() *SumoLogic {
				s := Default()
				s.headers = map[string]string{
					contentTypeHeader: graphiteContentType,
				}
				s.SetSerializer(&graphite.GraphiteSerializer{})
				return s
			},
		},
		{
			name: "prometheus is supported",
			plugin: func() *SumoLogic {
				s := Default()
				s.headers = map[string]string{
					contentTypeHeader: prometheusContentType,
				}
				s.SetSerializer(&prometheus.Serializer{})
				return s
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gz, err := gzip.NewReader(r.Body)
				require.NoError(t, err)
				io.Copy(&body, gz)
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
			require.NoError(t, err)

			plugin := tt.plugin()
			plugin.URL = u.String()

			require.NoError(t, plugin.Connect())

			err = plugin.Write([]telegraf.Metric{getMetric(t)})
			require.NoError(t, err)

			if tt.expectedBody != nil {
				require.Equal(t, string(tt.expectedBody), body.String())
			}
		})
	}
}

func TestContentEncodingGzip(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name   string
		plugin func() *SumoLogic
	}{
		{
			name: "default content_encoding=gzip works",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				return s
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

				body, err := gzip.NewReader(r.Body)
				require.NoError(t, err)

				payload, err := ioutil.ReadAll(body)
				require.NoError(t, err)

				assert.Equal(t, string(payload), "metric=cpu field=value  42 0\n")

				w.WriteHeader(http.StatusNoContent)
			})

			serializer, err := carbon2.NewSerializer(string(carbon2.Carbon2FormatFieldSeparate))
			require.NoError(t, err)

			plugin := tt.plugin()

			plugin.SetSerializer(serializer)
			err = plugin.Connect()
			require.NoError(t, err)

			err = plugin.Write([]telegraf.Metric{getMetric(t)})
			require.NoError(t, err)
		})
	}
}

type TestHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)

func TestDefaultUserAgent(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	t.Run("default-user-agent", func(t *testing.T) {
		ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, internal.ProductToken(), r.Header.Get("User-Agent"))
			w.WriteHeader(http.StatusOK)
		})

		plugin := &SumoLogic{
			URL:               u.String(),
			Method:            defaultMethod,
			MaxRequstBodySize: Default().MaxRequstBodySize,
		}

		serializer, err := carbon2.NewSerializer(string(carbon2.Carbon2FormatFieldSeparate))
		require.NoError(t, err)

		plugin.SetSerializer(serializer)
		err = plugin.Connect()
		require.NoError(t, err)

		err = plugin.Write([]telegraf.Metric{getMetric(t)})
		require.NoError(t, err)
	})
}

func TestTOMLConfig(t *testing.T) {
	testcases := []struct {
		name          string
		configBytes   []byte
		expectedError bool
	}{
		{
			name: "carbon2 content type is supported",
			configBytes: []byte(`
[[outputs.sumologic]]
  url = "https://localhost:3000"
  data_format = "carbon2"
            `),
			expectedError: false,
		},
		{
			name: "graphite content type is supported",
			configBytes: []byte(`
[[outputs.sumologic]]
  url = "https://localhost:3000"
  data_format = "graphite"
            `),
			expectedError: false,
		},
		{
			name: "prometheus content type is supported",
			configBytes: []byte(`
[[outputs.sumologic]]
  url = "https://localhost:3000"
  data_format = "prometheus"
            `),
			expectedError: false,
		},
		{
			name: "setting extra headers is not possible",
			configBytes: []byte(`
[[outputs.sumologic]]
  url = "https://localhost:3000"
  data_format = "carbon2"
  [outputs.sumologic.headers]
    X-Sumo-Name = "dummy"       
    X-Sumo-Host = "dummy"
    X-Sumo-Category  = "dummy"
    X-Sumo-Dimensions = "dummy"
            `),
			expectedError: true,
		},
		{
			name: "full example from sample config is correct",
			configBytes: []byte(`
[[outputs.sumologic]]
  url = "https://localhost:3000"
  data_format = "carbon2"
  timeout = "5s"
  method = "POST"
  source_name = "name"
  source_host = "hosta"
  source_category = "category"
  dimensions = "dimensions"
            `),
			expectedError: false,
		},
		{
			name: "unknown key - sumo_metadata - in config fails",
			configBytes: []byte(`
[[outputs.sumologic]]
  url = "https://localhost:3000"
  data_format = "carbon2"
  timeout = "5s"
  method = "POST"
  source_name = "name"
  sumo_metadata = "metadata"
            `),
			expectedError: true,
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			c := config.NewConfig()

			if tt.expectedError {
				require.Error(t, c.LoadConfigData(tt.configBytes))
			} else {
				require.NoError(t, c.LoadConfigData(tt.configBytes))
			}
		})
	}
}

func TestMaxRequestBodySize(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	const count = 100

	testcases := []struct {
		name                     string
		plugin                   func() *SumoLogic
		metrics                  []telegraf.Metric
		expectedError            bool
		expectedRequestCount     int32
		expectedMetricLinesCount int32
	}{
		{
			name: "default max request body size is 1MB and doesn't split small enough metric slices",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				return s
			},
			metrics:                  []telegraf.Metric{getMetric(t)},
			expectedError:            false,
			expectedRequestCount:     1,
			expectedMetricLinesCount: 1,
		},
		{
			name: "default max request body size is 1MB and doesn't split small even medium sized metrics",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				return s
			},
			metrics:                  getMetrics(t, count),
			expectedError:            false,
			expectedRequestCount:     1,
			expectedMetricLinesCount: 500, // count (100) metrics, 5 lines per each (steal, idle, system, user, temp) = 500
		},
		{
			name: "when short by at least 1B the request is split",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				// getMetrics returns metrics that serialized (using carbon2),
				// uncompressed size is 43750B
				s.MaxRequstBodySize = 43_749
				return s
			},
			metrics:                  getMetrics(t, count),
			expectedError:            false,
			expectedRequestCount:     2,
			expectedMetricLinesCount: 500, // count (100) metrics, 5 lines per each (steal, idle, system, user, temp) = 500
		},
		{
			name: "max request body size properly splits requests - max 10_000",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.MaxRequstBodySize = 10_000
				return s
			},
			metrics:                  getMetrics(t, count),
			expectedError:            false,
			expectedRequestCount:     5,
			expectedMetricLinesCount: 500, // count (100) metrics, 5 lines per each (steal, idle, system, user, temp) = 500
		},
		{
			name: "max request body size properly splits requests - max 5_000",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.MaxRequstBodySize = 5_000
				return s
			},
			metrics:                  getMetrics(t, count),
			expectedError:            false,
			expectedRequestCount:     10,
			expectedMetricLinesCount: 500, // count (100) metrics, 5 lines per each (steal, idle, system, user, temp) = 500
		},
		{
			name: "max request body size properly splits requests - max 2_500",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.MaxRequstBodySize = 2_500
				return s
			},
			metrics:                  getMetrics(t, count),
			expectedError:            false,
			expectedRequestCount:     20,
			expectedMetricLinesCount: 500, // count (100) metrics, 5 lines per each (steal, idle, system, user, temp) = 500
		},
		{
			name: "max request body size properly splits requests - max 1_000",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.MaxRequstBodySize = 1_000
				return s
			},
			metrics:                  getMetrics(t, count),
			expectedError:            false,
			expectedRequestCount:     50,
			expectedMetricLinesCount: 500, // count (100) metrics, 5 lines per each (steal, idle, system, user, temp) = 500
		},
		{
			name: "max request body size properly splits requests - max 500",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.MaxRequstBodySize = 500
				return s
			},
			metrics:                  getMetrics(t, count),
			expectedError:            false,
			expectedRequestCount:     100,
			expectedMetricLinesCount: 500, // count (100) metrics, 5 lines per each (steal, idle, system, user, temp) = 500
		},
		{
			name: "max request body size properly splits requests - max 300",
			plugin: func() *SumoLogic {
				s := Default()
				s.URL = u.String()
				s.MaxRequstBodySize = 300
				return s
			},
			metrics:                  getMetrics(t, count),
			expectedError:            false,
			expectedRequestCount:     100,
			expectedMetricLinesCount: 500, // count (100) metrics, 5 lines per each (steal, idle, system, user, temp) = 500
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				requestCount int32
				linesCount   int32
			)
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requestCount, 1)

				if tt.expectedMetricLinesCount != 0 {
					atomic.AddInt32(&linesCount, int32(countLines(t, r.Body)))
				}

				w.WriteHeader(http.StatusOK)
			})

			serializer, err := carbon2.NewSerializer(string(carbon2.Carbon2FormatFieldSeparate))
			require.NoError(t, err)

			plugin := tt.plugin()
			plugin.SetSerializer(serializer)

			err = plugin.Connect()
			require.NoError(t, err)

			err = plugin.Write(tt.metrics)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedRequestCount, atomic.LoadInt32(&requestCount))
				require.Equal(t, tt.expectedMetricLinesCount, atomic.LoadInt32(&linesCount))
			}
		})
	}
}

func TestTryingToSendEmptyMetricsDoesntFail(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	metrics := make([]telegraf.Metric, 0)
	plugin := Default()
	plugin.URL = u.String()

	serializer, err := carbon2.NewSerializer(string(carbon2.Carbon2FormatFieldSeparate))
	require.NoError(t, err)
	plugin.SetSerializer(serializer)

	err = plugin.Connect()
	require.NoError(t, err)

	err = plugin.Write(metrics)
	require.NoError(t, err)
}

func countLines(t *testing.T, body io.Reader) int {
	// All requests coming from Sumo Logic output plugin are gzipped.
	gz, err := gzip.NewReader(body)
	require.NoError(t, err)

	var linesCount int
	for s := bufio.NewScanner(gz); s.Scan(); {
		linesCount++
	}

	return linesCount
}
