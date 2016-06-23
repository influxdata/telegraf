package http_listener

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
	"net/http"
	"bytes"
)

const (
	testMsg = "cpu_load_short,host=server01 value=12.0 1422568543702900257\n"

	testMsgs = `cpu_load_short,host=server02 value=12.0 1422568543702900257
cpu_load_short,host=server03 value=12.0 1422568543702900257
cpu_load_short,host=server04 value=12.0 1422568543702900257
cpu_load_short,host=server05 value=12.0 1422568543702900257
cpu_load_short,host=server06 value=12.0 1422568543702900257
`
	badMsg = "blahblahblah: 42\n"

	emptyMsg = ""
)

func newTestHttpListener() (*HttpListener) {
	listener := &HttpListener{
		ServiceAddress: ":8186",
		ReadTimeout: "10",
		WriteTimeout: "10",
	}
	return listener
}

func TestWriteHTTP(t *testing.T) {
	listener := newTestHttpListener()
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	time.Sleep(time.Millisecond * 25)

	// post single message to listener
	var resp, err = http.Post("http://localhost:8186/write?db=mydb", "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	require.EqualValues(t, 204, resp.StatusCode)

	time.Sleep(time.Millisecond * 15)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)

	// post multiple message to listener
	resp, err = http.Post("http://localhost:8186/write?db=mydb", "", bytes.NewBuffer([]byte(testMsgs)))
	require.NoError(t, err)
	require.EqualValues(t, 204, resp.StatusCode)

	time.Sleep(time.Millisecond * 15)
	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}
}

func TestWriteHTTPInvalid(t *testing.T) {
	time.Sleep(time.Millisecond * 250)

	listener := newTestHttpListener()
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	time.Sleep(time.Millisecond * 25)

	// post single message to listener
	var resp, err = http.Post("http://localhost:8186/write?db=mydb", "", bytes.NewBuffer([]byte(badMsg)))
	require.NoError(t, err)
	require.EqualValues(t, 500, resp.StatusCode)
}

func TestWriteHTTPEmpty(t *testing.T) {
	time.Sleep(time.Millisecond * 250)

	listener := newTestHttpListener()
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	time.Sleep(time.Millisecond * 25)

	// post single message to listener
	var resp, err = http.Post("http://localhost:8186/write?db=mydb", "", bytes.NewBuffer([]byte(emptyMsg)))
	require.NoError(t, err)
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestQueryHTTP(t *testing.T) {
	time.Sleep(time.Millisecond * 250)

	listener := newTestHttpListener()
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	time.Sleep(time.Millisecond * 25)

	// post query to listener
	var resp, err = http.Post("http://localhost:8186/query?db=&q=CREATE+DATABASE+IF+NOT+EXISTS+%22mydb%22", "", nil)
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)
}

