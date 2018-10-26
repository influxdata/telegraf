package influxdb_v2_test

import (
	"net/url"
	"testing"

	influxdb "github.com/influxdata/telegraf/plugins/outputs/influxdb_v2"
	"github.com/stretchr/testify/require"
)

func genURL(u string) *url.URL {
	URL, _ := url.Parse(u)
	return URL
}
func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		err bool
		cfg *influxdb.HTTPConfig
	}{
		{
			err: true,
			cfg: &influxdb.HTTPConfig{},
		},
		{
			err: true,
			cfg: &influxdb.HTTPConfig{
				URL: genURL("udp://localhost:9999"),
			},
		},
		{
			cfg: &influxdb.HTTPConfig{
				URL: genURL("unix://var/run/influxd.sock"),
			},
		},
	}

	for i := range tests {
		client, err := influxdb.NewHTTPClient(tests[i].cfg)
		if !tests[i].err {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			t.Log(err)
		}
		if err == nil {
			client.URL()
		}
	}
}
