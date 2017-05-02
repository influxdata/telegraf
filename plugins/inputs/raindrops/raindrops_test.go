package raindrops

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"time"
)

const sampleResponse = `
calling: 100
writing: 200
0.0.0.0:8080 active: 1
0.0.0.0:8080 queued: 2
0.0.0.0:8081 active: 3
0.0.0.0:8081 queued: 4
127.0.0.1:8082 active: 5
127.0.0.1:8082 queued: 6
0.0.0.0:8083 active: 7
0.0.0.0:8083 queued: 8
0.0.0.0:8084 active: 9
0.0.0.0:8084 queued: 10
0.0.0.0:3000 active: 11
0.0.0.0:3000 queued: 12
/tmp/listen.me active: 13
/tmp/listen.me queued: 14
`

// Verify that raindrops tags are properly parsed based on the server
func TestRaindropsTags(t *testing.T) {
	urls := []string{"http://localhost/_raindrops", "http://localhost:80/_raindrops"}
	var addr *url.URL
	r := &Raindrops{}
	for _, url1 := range urls {
		addr, _ = url.Parse(url1)
		tagMap := r.getTags(addr)
		assert.Contains(t, tagMap["server"], "localhost")
	}
}

func TestRaindropsGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		if r.URL.Path == "/_raindrops" {
			rsp = sampleResponse
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	n := &Raindrops{
		Urls: []string{fmt.Sprintf("%s/_raindrops", ts.URL)},
		http_client: &http.Client{Transport: &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
		}},
	}

	var acc testutil.Accumulator

	err := acc.GatherError(n.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"calling": uint64(100),
		"writing": uint64(200),
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

	tags := map[string]string{"server": host, "port": port}
	acc.AssertContainsTaggedFields(t, "raindrops", fields, tags)

	tags = map[string]string{
		"port": "8081",
		"ip":   "0.0.0.0",
	}
	fields = map[string]interface{}{
		"active": uint64(3),
		"queued": uint64(4),
	}
	acc.AssertContainsTaggedFields(t, "raindrops_listen", fields, tags)
}
