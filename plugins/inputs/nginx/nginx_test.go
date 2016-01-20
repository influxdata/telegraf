package nginx

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
)

const sampleResponse = `
Active connections: 585
server accepts handled requests
 85340 85340 35085
Reading: 4 Writing: 135 Waiting: 446
`

// Verify that nginx tags are properly parsed based on the server
func TestNginxTags(t *testing.T) {
	urls := []string{"http://localhost/endpoint", "http://localhost:80/endpoint"}
	var addr *url.URL
	for _, url1 := range urls {
		addr, _ = url.Parse(url1)
		tagMap := getTags(addr)
		assert.Contains(t, tagMap["server"], "localhost")
	}
}

func TestNginxGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		if r.URL.Path == "/stub_status" {
			rsp = sampleResponse
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	n := &Nginx{
		Urls: []string{fmt.Sprintf("%s/stub_status", ts.URL)},
	}

	var acc testutil.Accumulator

	err := n.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"active":   uint64(585),
		"accepts":  uint64(85340),
		"handled":  uint64(85340),
		"requests": uint64(35085),
		"reading":  uint64(4),
		"writing":  uint64(135),
		"waiting":  uint64(446),
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
	acc.AssertContainsTaggedFields(t, "nginx", fields, tags)
}
