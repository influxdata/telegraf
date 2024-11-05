package influxdb_v2_test

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	influxdb "github.com/influxdata/telegraf/plugins/outputs/influxdb_v2"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSampleConfig(t *testing.T) {
	plugin := influxdb.InfluxDB{}
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestPluginRegistered(t *testing.T) {
	require.Contains(t, outputs.Outputs, "influxdb_v2")
}

func TestCloseWithoutConnect(t *testing.T) {
	plugin := influxdb.InfluxDB{}
	require.NoError(t, plugin.Close())
}

func TestDefaultURL(t *testing.T) {
	plugin := influxdb.InfluxDB{}
	require.NoError(t, plugin.Init())
	require.Len(t, plugin.URLs, 1)
	require.Equal(t, "http://localhost:8086", plugin.URLs[0])
}

func TestInit(t *testing.T) {
	tests := []*influxdb.InfluxDB{
		{
			URLs: []string{"https://localhost:8080"},
			ClientConfig: tls.ClientConfig{
				TLSCA: "thing",
			},
		},
	}

	for _, plugin := range tests {
		t.Run(plugin.URLs[0], func(t *testing.T) {
			require.Error(t, plugin.Init())
		})
	}
}

func TestConnectFail(t *testing.T) {
	tests := []*influxdb.InfluxDB{
		{
			URLs:      []string{"!@#$qwert"},
			HTTPProxy: "http://localhost:8086",
			HTTPHeaders: map[string]string{
				"x": "y",
			},
		},

		{

			URLs:      []string{"http://localhost:1234"},
			HTTPProxy: "!@#$%^&*()_+",
			HTTPHeaders: map[string]string{
				"x": "y",
			},
		},

		{

			URLs:      []string{"!@#$%^&*()_+"},
			HTTPProxy: "http://localhost:8086",
			HTTPHeaders: map[string]string{
				"x": "y",
			},
		},

		{

			URLs:      []string{":::@#$qwert"},
			HTTPProxy: "http://localhost:8086",
			HTTPHeaders: map[string]string{
				"x": "y",
			},
		},
	}

	for _, plugin := range tests {
		t.Run(plugin.URLs[0], func(t *testing.T) {
			require.NoError(t, plugin.Init())
			require.Error(t, plugin.Connect())
		})
	}
}

func TestConnect(t *testing.T) {
	tests := []*influxdb.InfluxDB{
		{
			URLs:      []string{"http://localhost:1234"},
			HTTPProxy: "http://localhost:8086",
			HTTPHeaders: map[string]string{
				"x": "y",
			},
		},
	}

	for _, plugin := range tests {
		t.Run(plugin.URLs[0], func(t *testing.T) {
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
		})
	}
}

func TestInfluxDBLocalAddress(t *testing.T) {
	t.Log("Starting server")
	server, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer server.Close()

	output := influxdb.InfluxDB{LocalAddr: "localhost"}
	require.NoError(t, output.Connect())
	require.NoError(t, output.Close())
}

func TestWrite(t *testing.T) {
	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v2/write":
				require.NoError(t, r.ParseForm())
				require.Equal(t, []string{"foobar"}, r.Form["bucket"])

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), "cpu value=42.123")

				w.WriteHeader(http.StatusNoContent)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:             []string{"http://" + ts.Listener.Addr().String()},
		Bucket:           "telegraf",
		BucketTag:        "bucket",
		ExcludeBucketTag: true,
		ContentEncoding:  "identity",
		PingTimeout:      config.Duration(15 * time.Second),
		ReadIdleTimeout:  config.Duration(30 * time.Second),
		Log:              &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Test writing
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"bucket": "foobar",
			},
			map[string]interface{}{
				"value": 42.123,
			},
			time.Unix(0, 0),
		),
	}
	require.NoError(t, plugin.Write(metrics))
	require.NoError(t, plugin.Write(metrics))
}

func TestWriteBucketTagWorksOnRetry(t *testing.T) {
	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v2/write":
				require.NoError(t, r.ParseForm())
				require.Equal(t, []string{"foo"}, r.Form["bucket"])

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), "cpu value=42")

				w.WriteHeader(http.StatusNoContent)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:             []string{"http://" + ts.Listener.Addr().String()},
		Bucket:           "telegraf",
		BucketTag:        "bucket",
		ExcludeBucketTag: true,
		ContentEncoding:  "identity",
		Log:              &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Send the metrics which should be succeed if sent twice
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"bucket": "foo",
			},
			map[string]interface{}{
				"value": 42.0,
			},
			time.Unix(0, 0),
		),
	}
	require.NoError(t, plugin.Write(metrics))
	require.NoError(t, plugin.Write(metrics))
}

func TestTooLargeWriteRetry(t *testing.T) {
	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v2/write":
				require.NoError(t, r.ParseForm())

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				// Ensure metric body size is small
				if len(body) > 16 {
					w.WriteHeader(http.StatusRequestEntityTooLarge)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}

				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:             []string{"http://" + ts.Listener.Addr().String()},
		Bucket:           "telegraf",
		BucketTag:        "bucket",
		ExcludeBucketTag: true,
		ContentEncoding:  "identity",
		Log:              &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Together the metric batch size is too big, split up, we get success
	metrics := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{
				"bucket": "foo",
			},
			map[string]interface{}{
				"value": 42.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"cpu",
			map[string]string{
				"bucket": "bar",
			},
			map[string]interface{}{
				"value": 99.0,
			},
			time.Unix(0, 0),
		),
	}
	require.NoError(t, plugin.Write(metrics))

	// These metrics are too big, even after splitting in half, expect error
	hugeMetrics := []telegraf.Metric{
		metric.New(
			"reallyLargeMetric",
			map[string]string{
				"bucket": "foobar",
			},
			map[string]interface{}{
				"value": 123.456,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"evenBiggerMetric",
			map[string]string{
				"bucket": "fizzbuzzbang",
			},
			map[string]interface{}{
				"value": 999.999,
			},
			time.Unix(0, 0),
		),
	}
	require.Error(t, plugin.Write(hugeMetrics))
}
