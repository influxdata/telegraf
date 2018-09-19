package influxdb_v2

import (
	"io"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func genURL(u string) *url.URL {
	URL, _ := url.Parse(u)
	return URL
}

func TestMakeWriteURL(t *testing.T) {
	tests := []struct {
		err bool
		url *url.URL
		act string
	}{
		{
			url: genURL("http://localhost:9999"),
			act: "http://localhost:9999/v2/write?bucket=telegraf&org=influx",
		},
		{
			url: genURL("unix://var/run/influxd.sock"),
			act: "http://127.0.0.1/v2/write?bucket=telegraf&org=influx",
		},
		{
			err: true,
			url: genURL("udp://localhost:9999"),
		},
	}

	for i := range tests {
		rURL, err := makeWriteURL(*tests[i].url, "influx", "telegraf")
		if !tests[i].err {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			t.Log(err)
		}
		if err == nil {
			require.Equal(t, tests[i].act, rURL)
		}
	}
}

func TestMakeWriteRequest(t *testing.T) {
	reader, _ := io.Pipe()
	cli := httpClient{
		WriteURL:        "http://localhost:9999/v2/write?bucket=telegraf&org=influx",
		ContentEncoding: "gzip",
		Headers:         map[string]string{"x": "y"},
	}
	_, err := cli.makeWriteRequest(reader)
	require.NoError(t, err)
}
