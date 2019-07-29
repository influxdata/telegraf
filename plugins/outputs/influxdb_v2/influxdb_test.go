package influxdb_v2_test

import (
	"testing"

	"github.com/influxdata/telegraf/internal/tls"
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
	require.Equal(t, "http://localhost:9999", output.URLs[0])
}
func TestConnect(t *testing.T) {
	tests := []struct {
		err bool
		out influxdb.InfluxDB
	}{
		{
			out: influxdb.InfluxDB{
				URLs:      []string{"http://localhost:1234"},
				HTTPProxy: "http://localhost:9999",
				HTTPHeaders: map[string]string{
					"x": "y",
				},
			},
		},
		{
			err: true,
			out: influxdb.InfluxDB{
				URLs:      []string{"!@#$qwert"},
				HTTPProxy: "http://localhost:9999",
				HTTPHeaders: map[string]string{
					"x": "y",
				},
			},
		},
		{
			err: true,
			out: influxdb.InfluxDB{
				URLs:      []string{"http://localhost:1234"},
				HTTPProxy: "!@#$%^&*()_+",
				HTTPHeaders: map[string]string{
					"x": "y",
				},
			},
		},
		{
			err: true,
			out: influxdb.InfluxDB{
				URLs:      []string{"!@#$%^&*()_+"},
				HTTPProxy: "http://localhost:9999",
				HTTPHeaders: map[string]string{
					"x": "y",
				},
			},
		},
		{
			err: true,
			out: influxdb.InfluxDB{
				URLs:      []string{":::@#$qwert"},
				HTTPProxy: "http://localhost:9999",
				HTTPHeaders: map[string]string{
					"x": "y",
				},
			},
		},
		{
			err: true,
			out: influxdb.InfluxDB{
				URLs: []string{"https://localhost:8080"},
				ClientConfig: tls.ClientConfig{
					TLSCA: "thing",
				},
			},
		},
	}

	for i := range tests {
		err := tests[i].out.Connect()
		if !tests[i].err {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			t.Log(err)
		}
	}
}

func TestUnused(t *testing.T) {
	thing := influxdb.InfluxDB{}
	thing.Close()
	thing.Description()
	thing.SampleConfig()
	outputs.Outputs["influxdb_v2"]()
}
