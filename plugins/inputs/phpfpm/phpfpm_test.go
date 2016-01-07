package phpfpm

import (
	"fmt"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
)

func TestPhpFpmGeneratesMetrics(t *testing.T) {
	//We create a fake server to return test data
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, outputSample)
	}))
	defer ts.Close()

	//Now we tested again above server, with our authentication data
	r := &phpfpm{
		Urls: []string{ts.URL},
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"url":  ts.Listener.Addr().String(),
		"pool": "www",
	}

	fields := map[string]interface{}{
		"accepted_conn":        int64(3),
		"listen_queue":         int64(1),
		"max_listen_queue":     int64(0),
		"listen_queue_len":     int64(0),
		"idle_processes":       int64(1),
		"active_processes":     int64(1),
		"total_processes":      int64(2),
		"max_active_processes": int64(1),
		"max_children_reached": int64(2),
		"slow_requests":        int64(1),
	}

	acc.AssertContainsTaggedFields(t, "phpfpm", fields, tags)
}

//When not passing server config, we default to localhost
//We just want to make sure we did request stat from localhost
func TestHaproxyDefaultGetFromLocalhost(t *testing.T) {
	r := &phpfpm{}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "127.0.0.1/status")
}

const outputSample = `
pool:                 www
process manager:      dynamic
start time:           11/Oct/2015:23:38:51 +0000
start since:          1991
accepted conn:        3
listen queue:         1
max listen queue:     0
listen queue len:     0
idle processes:       1
active processes:     1
total processes:      2
max active processes: 1
max children reached: 2
slow requests:        1
`
