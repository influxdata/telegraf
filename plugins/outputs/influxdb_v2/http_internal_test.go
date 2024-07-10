package influxdb_v2

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func genURL(u string) *url.URL {
	//nolint:errcheck // known test urls
	address, _ := url.Parse(u)
	return address
}

func TestMakeWriteURL(t *testing.T) {
	tests := []struct {
		err bool
		url *url.URL
		act string
		bkt string
		org string
	}{
		{
			url: genURL("http://localhost:9999"),
			act: "http://localhost:9999/api/v2/write?bucket=telegraf0&org=influx0",
			bkt: "telegraf0",
			org: "influx0",
		},
		{
			url: genURL("http://localhost:9999?id=abc"),
			act: "http://localhost:9999/api/v2/write?bucket=telegraf1&id=abc&org=influx1",
			bkt: "telegraf1",
			org: "influx1",
		},
		{
			url: genURL("unix://var/run/influxd.sock"),
			act: "http://127.0.0.1/api/v2/write?bucket=telegraf2&org=influx2",
			bkt: "telegraf2",
			org: "influx2",
		},
		{
			err: true,
			url: genURL("udp://localhost:9999"),
		},
	}

	for i := range tests {
		rURL, params, err := prepareWriteURL(*tests[i].url, tests[i].org)
		if !tests[i].err {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			t.Log(err)
		}
		if err == nil {
			for j := 0; j < 2; j++ {
				require.Equal(t, tests[i].act, makeWriteURL(*rURL, params, tests[i].bkt))
			}
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

var (
	bucket = "bkt"
	org    = "org"
	//nolint:errcheck // known test urls
	loc, params, _ = prepareWriteURL(*genURL("http://localhost:8086"), org)
)

// goos: linux
// goarch: amd64
// pkg: github.com/influxdata/telegraf/plugins/outputs/influxdb_v2
// cpu: 11th Gen Intel(R) Core(TM) i7-11850H @ 2.50GHz
// BenchmarkOldMakeWriteURL
// BenchmarkOldMakeWriteURL-16    	 1556631	       683.2 ns/op	     424 B/op	      14 allocs/op
// PASS
// ok  	github.com/influxdata/telegraf/plugins/outputs/influxdb_v2	1.851s
func BenchmarkOldMakeWriteURL(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Skip error for benchmarking
		oldMakeWriteURL(*loc, org, bucket)
	}
}

// goos: linux
// goarch: amd64
// pkg: github.com/influxdata/telegraf/plugins/outputs/influxdb_v2
// cpu: 11th Gen Intel(R) Core(TM) i7-11850H @ 2.50GHz
// BenchmarkNewMakeWriteURL
// BenchmarkNewMakeWriteURL-16    	 2084415	       496.5 ns/op	     280 B/op	       9 allocs/op
// PASS
// ok  	github.com/influxdata/telegraf/plugins/outputs/influxdb_v2	1.626s
func BenchmarkNewMakeWriteURL(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		makeWriteURL(*loc, params, bucket)
	}
}

func oldMakeWriteURL(loc url.URL, org, bucket string) (string, error) {
	params := url.Values{}
	params.Set("bucket", bucket)
	params.Set("org", org)

	switch loc.Scheme {
	case "unix":
		loc.Scheme = "http"
		loc.Host = "127.0.0.1"
		loc.Path = "/api/v2/write"
	case "http", "https":
		loc.Path = path.Join(loc.Path, "/api/v2/write")
	default:
		return "", fmt.Errorf("unsupported scheme: %q", loc.Scheme)
	}
	loc.RawQuery = params.Encode()
	return loc.String(), nil
}
