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
	address, _ := url.Parse(u)
	return address
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
		{retryCount: 1, expected: 25 * time.Millisecond},
		{retryCount: 5, expected: 625 * time.Millisecond},
		{retryCount: 10, expected: 2500 * time.Millisecond},
		{retryCount: 30, expected: 22500 * time.Millisecond},
		{retryCount: 40, expected: 40 * time.Second},
		{retryCount: 50, expected: 60 * time.Second}, // max hit
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

func TestExponentialBackoffCalculationWithRetryAfter(t *testing.T) {
	c := &httpClient{}
	tests := []struct {
		retryCount int
		retryAfter string
		expected   time.Duration
	}{
		{retryCount: 0, retryAfter: "0", expected: 0},
		{retryCount: 0, retryAfter: "10", expected: 10 * time.Second},
		{retryCount: 0, retryAfter: "60", expected: 60 * time.Second},
		{retryCount: 0, retryAfter: "600", expected: 600 * time.Second},
		{retryCount: 0, retryAfter: "601", expected: 600 * time.Second}, // max hit
		{retryCount: 40, retryAfter: "39", expected: 40 * time.Second},  // retryCount wins
		{retryCount: 40, retryAfter: "41", expected: 41 * time.Second},  // retryAfter wins
		{retryCount: 100, retryAfter: "100", expected: 100 * time.Second},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%d_retries", test.retryCount), func(t *testing.T) {
			c.retryCount = test.retryCount
			hdr := http.Header{}
			hdr.Add("Retry-After", test.retryAfter)
			require.EqualValues(t, test.expected, c.getRetryDuration(hdr))
		})
	}
}
