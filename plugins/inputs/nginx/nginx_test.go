package nginx

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

const nginxSampleResponse = `
Active connections: 585
server accepts handled requests
 85340 85340 35085
Reading: 4 Writing: 135 Waiting: 446
`
const tengineSampleResponse = `
Active connections: 403
server accepts handled requests request_time
 853 8533 3502 1546565864
Reading: 8 Writing: 125 Waiting: 946
`

// Verify that nginx tags are properly parsed based on the server
func TestNginxTags(t *testing.T) {
	urls := []string{"http://localhost/endpoint", "http://localhost:80/endpoint"}
	var addr *url.URL
	for _, url1 := range urls {
		addr, _ = url.Parse(url1)
		tagMap := getTags(addr)
		require.Contains(t, tagMap["server"], "localhost")
	}
}

func TestNginxGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		if r.URL.Path == "/stub_status" {
			rsp = nginxSampleResponse
		} else if r.URL.Path == "/tengine_status" {
			rsp = tengineSampleResponse
		} else {
			require.Fail(t, "Cannot handle request")
		}

		_, err := fmt.Fprintln(w, rsp)
		require.NoError(t, err)
	}))
	defer ts.Close()

	n := &Nginx{
		Urls: []string{fmt.Sprintf("%s/stub_status", ts.URL)},
	}

	nt := &Nginx{
		Urls: []string{fmt.Sprintf("%s/tengine_status", ts.URL)},
	}

	var accNginx testutil.Accumulator
	var accTengine testutil.Accumulator

	require.NoError(t, accNginx.GatherError(n.Gather))
	require.NoError(t, accTengine.GatherError(nt.Gather))

	fieldsNginx := map[string]interface{}{
		"active":   uint64(585),
		"accepts":  uint64(85340),
		"handled":  uint64(85340),
		"requests": uint64(35085),
		"reading":  uint64(4),
		"writing":  uint64(135),
		"waiting":  uint64(446),
	}

	fieldsTengine := map[string]interface{}{
		"active":   uint64(403),
		"accepts":  uint64(853),
		"handled":  uint64(8533),
		"requests": uint64(3502),
		"reading":  uint64(8),
		"writing":  uint64(125),
		"waiting":  uint64(946),
	}

	addr, err := url.Parse(ts.URL)
	require.NoError(t, err)

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
	accNginx.AssertContainsTaggedFields(t, "nginx", fieldsNginx, tags)
	accTengine.AssertContainsTaggedFields(t, "nginx", fieldsTengine, tags)
}
