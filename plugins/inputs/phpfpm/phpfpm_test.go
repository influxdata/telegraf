package phpfpm

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type statServer struct{}

// We create a fake server to return test data
func (s statServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprint(len(outputSample)))
	fmt.Fprint(w, outputSample)
}

func TestPhpFpmGeneratesMetrics_From_Http(t *testing.T) {
	sv := statServer{}
	ts := httptest.NewServer(sv)
	defer ts.Close()

	r := &phpfpm{
		Urls: []string{ts.URL},
	}

	err := r.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(r.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"pool": "www",
		"url":  ts.URL,
	}

	fields := map[string]interface{}{
		"start_since":          int64(1991),
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

func TestPhpFpmGeneratesMetrics_From_Fcgi(t *testing.T) {
	// Let OS find an available port
	tcp, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("Cannot initialize test server")
	}
	defer tcp.Close()

	s := statServer{}
	go fcgi.Serve(tcp, s)

	//Now we tested again above server
	r := &phpfpm{
		Urls: []string{"fcgi://" + tcp.Addr().String() + "/status"},
	}

	err = r.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = acc.GatherError(r.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"pool": "www",
		"url":  r.Urls[0],
	}

	fields := map[string]interface{}{
		"start_since":          int64(1991),
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

func TestPhpFpmGeneratesMetrics_From_Socket(t *testing.T) {
	// Create a socket in /tmp because we always have write permission and if the
	// removing of socket fail when system restart /tmp is clear so
	// we don't have junk files around
	var randomNumber int64
	binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
	tcp, err := net.Listen("unix", fmt.Sprintf("/tmp/test-fpm%d.sock", randomNumber))
	if err != nil {
		t.Fatal("Cannot initialize server on port ")
	}

	defer tcp.Close()
	s := statServer{}
	go fcgi.Serve(tcp, s)

	r := &phpfpm{
		Urls: []string{tcp.Addr().String()},
	}

	err = r.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(r.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"pool": "www",
		"url":  r.Urls[0],
	}

	fields := map[string]interface{}{
		"start_since":          int64(1991),
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

func TestPhpFpmGeneratesMetrics_From_Multiple_Sockets_With_Glob(t *testing.T) {
	// Create a socket in /tmp because we always have write permission and if the
	// removing of socket fail when system restart /tmp is clear so
	// we don't have junk files around
	var randomNumber int64
	binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
	socket1 := fmt.Sprintf("/tmp/test-fpm%d.sock", randomNumber)
	tcp1, err := net.Listen("unix", socket1)
	if err != nil {
		t.Fatal("Cannot initialize server on port ")
	}
	defer tcp1.Close()

	binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
	socket2 := fmt.Sprintf("/tmp/test-fpm%d.sock", randomNumber)
	tcp2, err := net.Listen("unix", socket2)
	if err != nil {
		t.Fatal("Cannot initialize server on port ")
	}
	defer tcp2.Close()

	s := statServer{}
	go fcgi.Serve(tcp1, s)
	go fcgi.Serve(tcp2, s)

	r := &phpfpm{
		Urls: []string{"/tmp/test-fpm[\\-0-9]*.sock"},
	}

	err = r.Init()
	require.NoError(t, err)

	var acc1, acc2 testutil.Accumulator

	err = acc1.GatherError(r.Gather)
	require.NoError(t, err)

	err = acc2.GatherError(r.Gather)
	require.NoError(t, err)

	tags1 := map[string]string{
		"pool": "www",
		"url":  socket1,
	}

	tags2 := map[string]string{
		"pool": "www",
		"url":  socket2,
	}

	fields := map[string]interface{}{
		"start_since":          int64(1991),
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

	acc1.AssertContainsTaggedFields(t, "phpfpm", fields, tags1)
	acc2.AssertContainsTaggedFields(t, "phpfpm", fields, tags2)
}

func TestPhpFpmGeneratesMetrics_From_Socket_Custom_Status_Path(t *testing.T) {
	// Create a socket in /tmp because we always have write permission. If the
	// removing of socket fail we won't have junk files around. Cuz when system
	// restart, it clears out /tmp
	var randomNumber int64
	binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
	tcp, err := net.Listen("unix", fmt.Sprintf("/tmp/test-fpm%d.sock", randomNumber))
	if err != nil {
		t.Fatal("Cannot initialize server on port ")
	}

	defer tcp.Close()
	s := statServer{}
	go fcgi.Serve(tcp, s)

	r := &phpfpm{
		Urls: []string{tcp.Addr().String() + ":custom-status-path"},
	}

	err = r.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(r.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"pool": "www",
		"url":  r.Urls[0],
	}

	fields := map[string]interface{}{
		"start_since":          int64(1991),
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
func TestPhpFpmDefaultGetFromLocalhost(t *testing.T) {
	r := &phpfpm{}

	err := r.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(r.Gather)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "127.0.0.1/status")
}

func TestPhpFpmGeneratesMetrics_Throw_Error_When_Fpm_Status_Is_Not_Responding(t *testing.T) {
	r := &phpfpm{
		Urls: []string{"http://aninvalidone"},
	}

	err := r.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(r.Gather)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unable to connect to phpfpm status page 'http://aninvalidone'`)
	assert.Contains(t, err.Error(), `lookup aninvalidone`)
}

func TestPhpFpmGeneratesMetrics_Throw_Error_When_Socket_Path_Is_Invalid(t *testing.T) {
	r := &phpfpm{
		Urls: []string{"/tmp/invalid.sock"},
	}

	err := r.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(r.Gather)
	require.Error(t, err)
	assert.Equal(t, `dial unix /tmp/invalid.sock: connect: no such file or directory`, err.Error())

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
