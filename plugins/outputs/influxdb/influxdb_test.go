package influxdb_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs/influxdb"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

type MockClient struct {
	URLF            func() string
	WriteF          func() error
	CreateDatabaseF func() error
	DatabaseF       func() string
	CloseF          func()

	log telegraf.Logger
}

func (c *MockClient) URL() string {
	return c.URLF()
}

func (c *MockClient) Write(context.Context, []telegraf.Metric) error {
	return c.WriteF()
}

func (c *MockClient) CreateDatabase(context.Context, string) error {
	return c.CreateDatabaseF()
}

func (c *MockClient) Database() string {
	return c.DatabaseF()
}

func (c *MockClient) Close() {
	if c.CloseF != nil {
		c.CloseF()
	}
}

func (c *MockClient) SetLogger(log telegraf.Logger) {
	c.log = log
}

func TestDeprecatedURLSupport(t *testing.T) {
	var actual *influxdb.UDPConfig
	output := influxdb.InfluxDB{
		URLs: []string{"udp://localhost:8089"},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{}, nil
		},
		Log:        testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
	}
	defer output.Statistics.UnregisterAll()

	require.NoError(t, output.Init())
	require.NoError(t, output.Connect())
	defer output.Close()

	require.Equal(t, "udp://localhost:8089", actual.URL.String())
}

func TestDefaultURL(t *testing.T) {
	var actual *influxdb.HTTPConfig
	output := influxdb.InfluxDB{
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{
				DatabaseF: func() string {
					return "telegraf"
				},
				CreateDatabaseF: func() error {
					return nil
				},
			}, nil
		},
		Log:        testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
	}
	defer output.Statistics.UnregisterAll()

	require.NoError(t, output.Init())
	require.NoError(t, output.Connect())
	defer output.Close()

	require.Equal(t, "http://localhost:8086", actual.URL.String())
}

func TestConnectUDPConfig(t *testing.T) {
	var actual *influxdb.UDPConfig

	output := influxdb.InfluxDB{
		URLs:       []string{"udp://localhost:8089"},
		UDPPayload: config.Size(42),

		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{}, nil
		},
		Log:        testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
	}
	defer output.Statistics.UnregisterAll()

	require.NoError(t, output.Init())
	require.NoError(t, output.Connect())
	defer output.Close()

	require.Equal(t, "udp://localhost:8089", actual.URL.String())
	require.Equal(t, 42, actual.MaxPayloadSize)
	require.NotNil(t, actual.Serializer)
}

func TestConnectHTTPConfig(t *testing.T) {
	var actual *influxdb.HTTPConfig

	output := influxdb.InfluxDB{
		URLs:             []string{"http://localhost:8086"},
		Database:         "telegraf",
		RetentionPolicy:  "default",
		WriteConsistency: "any",
		Timeout:          config.Duration(5 * time.Second),
		Username:         config.NewSecret([]byte("guy")),
		Password:         config.NewSecret([]byte("smiley")),
		UserAgent:        "telegraf",
		HTTPProxy:        "http://localhost:8086",
		HTTPHeaders: map[string]string{
			"x": "y",
		},
		ContentEncoding: "gzip",
		ClientConfig: tls.ClientConfig{
			InsecureSkipVerify: true,
		},

		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{
				DatabaseF: func() string {
					return "telegraf"
				},
				CreateDatabaseF: func() error {
					return nil
				},
			}, nil
		},
		Log:        testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
	}
	defer output.Statistics.UnregisterAll()

	require.NoError(t, output.Init())
	require.NoError(t, output.Connect())
	defer output.Close()

	require.Equal(t, output.URLs[0], actual.URL.String())
	require.Equal(t, output.UserAgent, actual.UserAgent)
	require.Equal(t, time.Duration(output.Timeout), actual.Timeout)
	require.Equal(t, output.Username, actual.Username)
	require.Equal(t, output.Password, actual.Password)
	require.Equal(t, output.HTTPProxy, actual.Proxy.String())
	require.Equal(t, output.HTTPHeaders, actual.Headers)
	require.Equal(t, output.ContentEncoding, actual.ContentEncoding)
	require.Equal(t, output.Database, actual.Database)
	require.Equal(t, output.RetentionPolicy, actual.RetentionPolicy)
	require.Equal(t, output.WriteConsistency, actual.Consistency)
	require.NotNil(t, actual.TLSConfig)
	require.NotNil(t, actual.Serializer)

	require.Equal(t, output.Database, actual.Database)
}

func TestWriteRecreateDatabaseIfDatabaseNotFound(t *testing.T) {
	output := influxdb.InfluxDB{
		URLs: []string{"http://localhost:8086"},
		CreateHTTPClientF: func(*influxdb.HTTPConfig) (influxdb.Client, error) {
			return &MockClient{
				DatabaseF: func() string {
					return "telegraf"
				},
				CreateDatabaseF: func() error {
					return nil
				},
				WriteF: func() error {
					return &influxdb.DatabaseNotFoundError{
						APIError: influxdb.APIError{
							StatusCode:  http.StatusNotFound,
							Title:       "404 Not Found",
							Description: `database not found "telegraf"`,
						},
					}
				},
				URLF: func() string {
					return "http://localhost:8086"
				},
			}, nil
		},
		Log:        testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
	}
	defer output.Statistics.UnregisterAll()

	require.NoError(t, output.Init())
	require.NoError(t, output.Connect())
	defer output.Close()

	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	metrics := []telegraf.Metric{m}

	// We only have one URL, so we expect an error
	require.Error(t, output.Write(metrics))
}

func TestInfluxDBLocalAddress(t *testing.T) {
	output := influxdb.InfluxDB{
		URLs:      []string{"http://localhost:8086"},
		LocalAddr: "localhost",

		CreateHTTPClientF: func(_ *influxdb.HTTPConfig) (influxdb.Client, error) {
			return &MockClient{
				DatabaseF: func() string {
					return "telegraf"
				},
				CreateDatabaseF: func() error {
					return nil
				},
			}, nil
		},
		Log:        testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
	}
	defer output.Statistics.UnregisterAll()

	require.NoError(t, output.Init())
	require.NoError(t, output.Connect())
	output.Close()
}

func TestBytesWrittenHTTP(t *testing.T) {
	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		ContentEncoding:      "none",
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Check that we start with a zero counter
	stat := plugin.Statistics.Get("write", "bytes_written", nil)
	require.Zero(t, stat.Get())

	// Write data
	input := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(42),
			},
			time.Unix(0, 0),
		),
	}
	require.NoError(t, plugin.Write(input))

	require.Equal(t, int64(28), stat.Get())
}

func TestBytesWrittenHTTPGzip(t *testing.T) {
	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		ContentEncoding:      "gzip",
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Check that we start with a zero counter
	stat := plugin.Statistics.Get("write", "bytes_written", nil)
	require.Zero(t, stat.Get())

	// Write data
	input := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(42),
			},
			time.Unix(0, 0),
		),
	}
	require.NoError(t, plugin.Write(input))

	require.Equal(t, int64(52), stat.Get())
}

func TestBytesWrittenUDP(t *testing.T) {
	// Setup a test server
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer conn.Close()

	addr := conn.LocalAddr()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{addr.Network() + "://" + addr.String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Check that we start with a zero counter
	stat := plugin.Statistics.Get("write", "bytes_written", nil)
	require.Zero(t, stat.Get())

	// Write data
	input := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(42),
			},
			time.Unix(0, 0),
		),
	}
	require.NoError(t, plugin.Write(input))

	require.Equal(t, int64(28), stat.Get())
}

func BenchmarkWrite1k(b *testing.B) {
	batchsize := 1000

	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		ContentEncoding:      "gzip",
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(b, plugin.Init())
	require.NoError(b, plugin.Connect())
	defer plugin.Close()

	metrics := make([]telegraf.Metric, 0, batchsize)
	for i := range batchsize {
		metrics = append(metrics, metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(i),
			},
			time.Unix(0, 0),
		))
	}

	// Benchmark the writing
	b.ResetTimer()
	for b.Loop() {
		require.NoError(b, plugin.Write(metrics))
	}
	b.ReportMetric(float64(batchsize*b.N)/b.Elapsed().Seconds(), "metrics/s")
}

func BenchmarkWrite5k(b *testing.B) {
	batchsize := 5000

	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		ContentEncoding:      "gzip",
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(b, plugin.Init())
	require.NoError(b, plugin.Connect())
	defer plugin.Close()

	metrics := make([]telegraf.Metric, 0, batchsize)
	for i := range batchsize {
		metrics = append(metrics, metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(i),
			},
			time.Unix(0, 0),
		))
	}

	// Benchmark the writing
	b.ResetTimer()
	for b.Loop() {
		require.NoError(b, plugin.Write(metrics))
	}
	b.ReportMetric(float64(batchsize*b.N)/b.Elapsed().Seconds(), "metrics/s")
}

func BenchmarkWrite10k(b *testing.B) {
	batchsize := 10000

	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		ContentEncoding:      "gzip",
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(b, plugin.Init())
	require.NoError(b, plugin.Connect())
	defer plugin.Close()

	metrics := make([]telegraf.Metric, 0, batchsize)
	for i := range batchsize {
		metrics = append(metrics, metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(i),
			},
			time.Unix(0, 0),
		))
	}

	// Benchmark the writing
	b.ResetTimer()
	for b.Loop() {
		require.NoError(b, plugin.Write(metrics))
	}
	b.ReportMetric(float64(batchsize*b.N)/b.Elapsed().Seconds(), "metrics/s")
}

func BenchmarkWrite25k(b *testing.B) {
	batchsize := 25000

	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		ContentEncoding:      "gzip",
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(b, plugin.Init())
	require.NoError(b, plugin.Connect())
	defer plugin.Close()

	metrics := make([]telegraf.Metric, 0, batchsize)
	for i := range batchsize {
		metrics = append(metrics, metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(i),
			},
			time.Unix(0, 0),
		))
	}

	// Benchmark the writing
	b.ResetTimer()
	for b.Loop() {
		require.NoError(b, plugin.Write(metrics))
	}
	b.ReportMetric(float64(batchsize*b.N)/b.Elapsed().Seconds(), "metrics/s")
}

func BenchmarkWrite50k(b *testing.B) {
	batchsize := 50000

	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		ContentEncoding:      "gzip",
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(b, plugin.Init())
	require.NoError(b, plugin.Connect())
	defer plugin.Close()

	metrics := make([]telegraf.Metric, 0, batchsize)
	for i := range batchsize {
		metrics = append(metrics, metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(i),
			},
			time.Unix(0, 0),
		))
	}

	// Benchmark the writing
	b.ResetTimer()
	for b.Loop() {
		require.NoError(b, plugin.Write(metrics))
	}
	b.ReportMetric(float64(batchsize*b.N)/b.Elapsed().Seconds(), "metrics/s")
}

func BenchmarkWrite100k(b *testing.B) {
	batchsize := 100000

	// Setup a test server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	defer ts.Close()

	// Setup plugin and connect
	plugin := &influxdb.InfluxDB{
		URLs:       []string{"http://" + ts.Listener.Addr().String()},
		Username:   config.NewSecret([]byte("user")),
		Password:   config.NewSecret([]byte("secret")),
		Database:   "my_database",
		Timeout:    config.Duration(time.Second * 5),
		Log:        &testutil.Logger{},
		Statistics: selfstat.NewCollector(nil),
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			return influxdb.NewHTTPClient(*config)
		},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			return influxdb.NewUDPClient(*config)
		},
		ContentEncoding:      "gzip",
		SkipDatabaseCreation: true,
	}
	defer plugin.Statistics.UnregisterAll()

	require.NoError(b, plugin.Init())
	require.NoError(b, plugin.Connect())
	defer plugin.Close()

	metrics := make([]telegraf.Metric, 0, batchsize)
	for i := range batchsize {
		metrics = append(metrics, metric.New(
			"cpu",
			map[string]string{
				"database": "foo",
			},
			map[string]interface{}{
				"value": float64(i),
			},
			time.Unix(0, 0),
		))
	}

	// Benchmark the writing
	b.ResetTimer()
	for b.Loop() {
		require.NoError(b, plugin.Write(metrics))
	}
	b.ReportMetric(float64(batchsize*b.N)/b.Elapsed().Seconds(), "metrics/s")
}
