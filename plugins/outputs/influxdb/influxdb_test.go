package influxdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs/influxdb"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	URLF            func() string
	DatabaseF       func() string
	WriteF          func(context.Context, []telegraf.Metric) error
	CreateDatabaseF func(ctx context.Context) error
}

func (c *MockClient) URL() string {
	return c.URLF()
}

func (c *MockClient) Database() string {
	return c.DatabaseF()
}

func (c *MockClient) Write(ctx context.Context, metrics []telegraf.Metric) error {
	return c.WriteF(ctx, metrics)
}

func (c *MockClient) CreateDatabase(ctx context.Context) error {
	return c.CreateDatabaseF(ctx)
}

func TestDeprecatedURLSupport(t *testing.T) {
	var actual *influxdb.UDPConfig
	output := influxdb.InfluxDB{
		URL: "udp://localhost:8086",

		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{}, nil
		},
	}
	err := output.Connect()
	require.NoError(t, err)
	require.Equal(t, "udp://localhost:8086", actual.URL.String())
}

func TestDefaultURL(t *testing.T) {
	var actual *influxdb.HTTPConfig
	output := influxdb.InfluxDB{
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{
				CreateDatabaseF: func(ctx context.Context) error {
					return nil
				},
			}, nil
		},
	}
	err := output.Connect()
	require.NoError(t, err)
	require.Equal(t, "http://localhost:8086", actual.URL.String())
}

func TestConnectUDPConfig(t *testing.T) {
	var actual *influxdb.UDPConfig

	output := influxdb.InfluxDB{
		URLs:       []string{"udp://localhost:8086"},
		UDPPayload: 42,

		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{}, nil
		},
	}
	err := output.Connect()
	require.NoError(t, err)

	require.Equal(t, "udp://localhost:8086", actual.URL.String())
	require.Equal(t, 42, actual.MaxPacketSize)
	require.NotNil(t, actual.Serializer)
}

func TestConnectHTTPConfig(t *testing.T) {
	var actual *influxdb.HTTPConfig

	output := influxdb.InfluxDB{
		URLs:             []string{"http://localhost:8089"},
		Database:         "telegraf",
		RetentionPolicy:  "default",
		WriteConsistency: "any",
		Timeout:          internal.Duration{Duration: 5 * time.Second},
		Username:         "guy",
		Password:         "smiley",
		UserAgent:        "telegraf",
		HTTPProxy:        "http://localhost:8089",
		HTTPHeaders: map[string]string{
			"x": "y",
		},
		ContentEncoding:    "gzip",
		InsecureSkipVerify: true,

		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{
				CreateDatabaseF: func(ctx context.Context) error {
					return nil
				},
			}, nil
		},
	}
	err := output.Connect()
	require.NoError(t, err)

	require.Equal(t, output.URLs[0], actual.URL.String())
	require.Equal(t, output.UserAgent, actual.UserAgent)
	require.Equal(t, output.Timeout.Duration, actual.Timeout)
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
