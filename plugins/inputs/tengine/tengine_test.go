package tengine

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

const tengineSampleResponse = `127.0.0.1,784,1511,2,2,1,0,1,0,0,0,0,0,0,1,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0`

// Verify that tengine tags are properly parsed based on the server
func TestTengineTags(t *testing.T) {
	urls := []string{"http://localhost/us", "http://localhost:80/us"}
	var addr *url.URL
	for _, url1 := range urls {
		addr, _ = url.Parse(url1)
		tagMap := getTags(addr, "127.0.0.1")
		require.Contains(t, tagMap["server"], "localhost")
	}
}

func TestTengineGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, tengineSampleResponse)
		require.NoError(t, err)
	}))
	defer ts.Close()

	n := &Tengine{
		Urls: []string{fmt.Sprintf("%s/us", ts.URL)},
	}

	var accTengine testutil.Accumulator

	errTengine := accTengine.GatherError(n.Gather)

	require.NoError(t, errTengine)

	fieldsTengine := map[string]interface{}{
		"bytes_in":                 uint64(784),
		"bytes_out":                uint64(1511),
		"conn_total":               uint64(2),
		"req_total":                uint64(2),
		"http_2xx":                 uint64(1),
		"http_3xx":                 uint64(0),
		"http_4xx":                 uint64(1),
		"http_5xx":                 uint64(0),
		"http_other_status":        uint64(0),
		"rt":                       uint64(0),
		"ups_req":                  uint64(0),
		"ups_rt":                   uint64(0),
		"ups_tries":                uint64(0),
		"http_200":                 uint64(1),
		"http_206":                 uint64(0),
		"http_302":                 uint64(0),
		"http_304":                 uint64(0),
		"http_403":                 uint64(0),
		"http_404":                 uint64(1),
		"http_416":                 uint64(0),
		"http_499":                 uint64(0),
		"http_500":                 uint64(0),
		"http_502":                 uint64(0),
		"http_503":                 uint64(0),
		"http_504":                 uint64(0),
		"http_508":                 uint64(0),
		"http_other_detail_status": uint64(0),
		"http_ups_4xx":             uint64(0),
		"http_ups_5xx":             uint64(0),
	}

	addr, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	host, port, err := net.SplitHostPort(addr.Host)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}
	tags := map[string]string{"server": host, "port": port, "server_name": "127.0.0.1"}
	accTengine.AssertContainsTaggedFields(t, "tengine", fieldsTengine, tags)
}
