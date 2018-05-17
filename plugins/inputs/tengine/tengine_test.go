package tengine

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

const tengineSampleResponse = `35.190.79.28,784,1511,2,2,1,0,1,0,0,0,0,0,0,1,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0
check.proxyradar.com,535,771,1,1,0,0,1,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0
tenka-prod-api.txwy.tw,2992518616,14187906760,32562,3180654,3157774,0,0,22880,0,246095380,2666852,246095109,2666852,3157774,0,0,0,0,0,0,0,16384,6496,0,0,0,0,0,22880`

// Verify that tengine tags are properly parsed based on the server
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
		rsp = tengineSampleResponse
		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	n := &Tengine{
		Urls: []string{fmt.Sprintf("%s/stub_status", ts.URL)},
	}

	var acc_tengine testutil.Accumulator

	err_tengine := acc_tengine.GatherError(n.Gather)

	require.NoError(t, err_tengine)

	fields_tengine := map[string]interface{}{
		"bytes_in": uint64(185),
		"bytes_out": uint64(185),
		"conn_total": uint64(185),
		"req_total": uint64(185),
		"http_2xx": uint64(185),
		"http_3xx": uint64(185),
		"http_4xx": uint64(185),
		"http_5xx": uint64(185),
		"http_other_status": uint64(185),
		"rt": uint64(185),
		"ups_req": uint64(185),
		"ups_rt": uint64(185),
		"ups_tries": uint64(185),
		"http_200": uint64(185),
		"http_206": uint64(185),
		"http_302": uint64(185),
		"http_304": uint64(185),
		"http_403": uint64(185),
		"http_404": uint64(185),
		"http_416": uint64(185),
		"http_499": uint64(185),
		"http_500": uint64(185),
		"http_502": uint64(185),
		"http_503": uint64(185),
		"http_504": uint64(185),
		"http_508": uint64(185),
		"http_other_detail_status": uint64(185),
		"http_ups_4xx": uint64(185),
		"http_ups_5xx": uint64(185),
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
	acc_tengine.AssertContainsTaggedFields(t, "tengine", fields_tengine, tags)
}
