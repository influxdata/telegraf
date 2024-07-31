package influxdb_v2_test

import (
	"net"
	"testing"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	influxdb "github.com/influxdata/telegraf/plugins/outputs/influxdb_v2"
	"github.com/stretchr/testify/require"
)

func TestDefaultURL(t *testing.T) {
	output := influxdb.InfluxDB{}
	err := output.Connect()
	require.NoError(t, err)
	if len(output.URLs) < 1 {
		t.Fatal("Default URL failed to get set")
	}
	require.Equal(t, "http://localhost:8086", output.URLs[0])
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

func TestUnused(_ *testing.T) {
	thing := influxdb.InfluxDB{}
	thing.Close()
	thing.SampleConfig()
	outputs.Outputs["influxdb_v2"]()
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
