package nginx

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleResponse = `
Active connections: 585 
server accepts handled requests
 85340 85340 35085 
Reading: 4 Writing: 135 Waiting: 446 
`

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

	metrics := []struct {
		name  string
		value uint64
	}{
		{"active", 585},
		{"accepts", 85340},
		{"handled", 85340},
		{"requests", 35085},
		{"reading", 4},
		{"writing", 135},
		{"waiting", 446},
	}
	addr, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}
	host, _, _ := net.SplitHostPort(addr.Host)
	tags := map[string]string{"server": host}

	for _, m := range metrics {
		assert.NoError(t, acc.ValidateTaggedValue(m.name, m.value, tags))
	}
}
