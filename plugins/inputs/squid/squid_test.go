package squid

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleBody = `
sample_time = 1549933585.137495 (Tue, 12 Feb 2019 01:06:25 GMT)
client_http.requests = 37
client_http.hits = 0
client_http.errors = 3
client_http.kbytes_in = 5
client_http.kbytes_out = 207
client_http.hit_kbytes_out = 0
server.all.requests = 4
server.all.errors = 0
server.all.kbytes_in = 98
server.all.kbytes_out = 1
server.http.requests = 1
server.http.errors = 0
server.http.kbytes_in = 0
server.http.kbytes_out = 0
server.ftp.requests = 0
server.ftp.errors = 0
server.ftp.kbytes_in = 0
server.ftp.kbytes_out = 0
server.other.requests = 3
server.other.errors = 0
server.other.kbytes_in = 98
server.other.kbytes_out = 1
icp.pkts_sent = 0
icp.pkts_recv = 0
icp.queries_sent = 0
icp.replies_sent = 0
icp.queries_recv = 0
icp.replies_recv = 0
icp.query_timeouts = 0
icp.replies_queued = 0
icp.kbytes_sent = 0
icp.kbytes_recv = 0
icp.q_kbytes_sent = 0
icp.r_kbytes_sent = 0
icp.q_kbytes_recv = 0
icp.r_kbytes_recv = 0
icp.times_used = 0
cd.times_used = 0
cd.msgs_sent = 0
cd.msgs_recv = 0
cd.memory = 0
cd.local_memory = 0
cd.kbytes_sent = 0
cd.kbytes_recv = 0
unlink.requests = 0
page_faults = 0
select_loops = 5102
cpu_time = 0.291445
wall_time = 49.503122
swap.outs = 0
swap.ins = 0
swap.files_cleaned = 0
aborted_requests = 0`

func TestGatherTimeout(t *testing.T) {
	var acc testutil.Accumulator

	// dummy config to a url that should timeout
	s := &Squid{
		Url:             "http://localhost:1021",
		ResponseTimeout: internal.Duration{Duration: time.Millisecond * 100},
	}

	// this should return an error
	err := s.Gather(&acc)

	require.NoError(t, err)        // test that we did not return an error
	assert.Zero(t, acc.NFields())  // test that we didn't return any fields
	assert.NotEmpty(t, acc.Errors) // test that the accumulator has errors
}

func TestGatherFull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s := &Squid{
		Url:             fmt.Sprintf("http://%s:3128", testutil.GetLocalHost()),
		ResponseTimeout: internal.Duration{Duration: time.Second * 5},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)

	require.NoError(t, err)
	assert.Empty(t, acc.Errors, "accumulator had no errors")
	assert.True(t, acc.HasMeasurement("squid"), "Has a measurement called 'squid'")
	assert.Equal(t, s.Url, acc.TagValue("squid", "source"), "Has a tag value for squid equal to localhost")
	assert.True(t, acc.HasFloatField("squid", "client_http_requests"), "Has a float field called client_http.requests")
}

func TestParseBody(t *testing.T) {
	body := strings.NewReader(sampleBody)
	fields := parseBody(body)

	assert.Empty(t, fields["sample_time"], "ommitted sample_time field")
	assert.Equal(t, float64(37), fields["client_http_requests"], "Has field called 'client_http.requests'")
	assert.Equal(t, float64(49.503122), fields["wall_time"], "Has field called 'wall_time'")
}
