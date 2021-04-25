package influxdb_v2

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

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
			act: "http://localhost:9999/api/v2/write?bucket=telegraf&org=influx",
		},
		{
			url: genURL("unix://var/run/influxd.sock"),
			act: "http://127.0.0.1/api/v2/write?bucket=telegraf&org=influx",
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

func TestExponentialBackoffCalculation(t *testing.T) {
	c := &httpClient{}
	tests := []struct {
		retryCount int
		expected   time.Duration
	}{
		{retryCount: 0, expected: 0},
		{retryCount: 1, expected: 0},
		{retryCount: 5, expected: 0},
		{retryCount: 10, expected: 2 * time.Second},
		{retryCount: 30, expected: 22 * time.Second},
		{retryCount: 40, expected: 40 * time.Second},
		{retryCount: 50, expected: 60 * time.Second},
		{retryCount: 100, expected: 60 * time.Second},
		{retryCount: 1000, expected: 60 * time.Second},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%d_retries", test.retryCount), func(t *testing.T) {
			c.retryCount = test.retryCount
			require.EqualValues(t, test.expected, c.getRetryDuration(http.Header{}))
		})
	}
}
