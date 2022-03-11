//go:build !windows
// +build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

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

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

type statServer struct{}

// We create a fake server to return test data
func (s statServer) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprint(len(outputSample)))
	// Ignore the returned error as the tests will fail anyway
	//nolint:errcheck,revive
	fmt.Fprint(w, outputSample)
}

func TestPhpFpmGeneratesMetrics_From_Http(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "ok", r.URL.Query().Get("test"))
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", fmt.Sprint(len(outputSample)))
		_, err := fmt.Fprint(w, outputSample)
		require.NoError(t, err)
	}))
	defer ts.Close()

	url := ts.URL + "?test=ok"
	r := &phpfpm{
		Urls: []string{url},
	}

	require.NoError(t, r.Init())

	var acc testutil.Accumulator

	require.NoError(t, acc.GatherError(r.Gather))

	tags := map[string]string{
		"pool": "www",
		"url":  url,
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
	require.NoError(t, err, "Cannot initialize test server")
	defer tcp.Close()

	s := statServer{}
	//nolint:errcheck,revive
	go fcgi.Serve(tcp, s)

	//Now we tested again above server
	r := &phpfpm{
		Urls: []string{"fcgi://" + tcp.Addr().String() + "/status"},
	}

	require.NoError(t, r.Init())

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(r.Gather))

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
	require.NoError(t, binary.Read(rand.Reader, binary.LittleEndian, &randomNumber))
	tcp, err := net.Listen("unix", fmt.Sprintf("/tmp/test-fpm%d.sock", randomNumber))
	require.NoError(t, err, "Cannot initialize server on port ")

	defer tcp.Close()
	s := statServer{}
	//nolint:errcheck,revive
	go fcgi.Serve(tcp, s)

	r := &phpfpm{
		Urls: []string{tcp.Addr().String()},
	}

	require.NoError(t, r.Init())

	var acc testutil.Accumulator

	require.NoError(t, acc.GatherError(r.Gather))

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
	require.NoError(t, binary.Read(rand.Reader, binary.LittleEndian, &randomNumber))
	socket1 := fmt.Sprintf("/tmp/test-fpm%d.sock", randomNumber)
	tcp1, err := net.Listen("unix", socket1)
	require.NoError(t, err, "Cannot initialize server on port ")
	defer tcp1.Close()

	require.NoError(t, binary.Read(rand.Reader, binary.LittleEndian, &randomNumber))
	socket2 := fmt.Sprintf("/tmp/test-fpm%d.sock", randomNumber)
	tcp2, err := net.Listen("unix", socket2)
	require.NoError(t, err, "Cannot initialize server on port ")
	defer tcp2.Close()

	s := statServer{}
	//nolint:errcheck,revive
	go fcgi.Serve(tcp1, s)
	//nolint:errcheck,revive
	go fcgi.Serve(tcp2, s)

	r := &phpfpm{
		Urls: []string{"/tmp/test-fpm[\\-0-9]*.sock"},
	}

	require.NoError(t, r.Init())

	var acc1, acc2 testutil.Accumulator

	require.NoError(t, acc1.GatherError(r.Gather))

	require.NoError(t, acc2.GatherError(r.Gather))

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
	require.NoError(t, binary.Read(rand.Reader, binary.LittleEndian, &randomNumber))
	tcp, err := net.Listen("unix", fmt.Sprintf("/tmp/test-fpm%d.sock", randomNumber))
	require.NoError(t, err, "Cannot initialize server on port ")

	defer tcp.Close()
	s := statServer{}
	//nolint:errcheck,revive
	go fcgi.Serve(tcp, s)

	r := &phpfpm{
		Urls: []string{tcp.Addr().String() + ":custom-status-path"},
	}

	require.NoError(t, r.Init())

	var acc testutil.Accumulator

	require.NoError(t, acc.GatherError(r.Gather))

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
	r := &phpfpm{Urls: []string{"http://bad.localhost:62001/status"}}

	require.NoError(t, r.Init())

	var acc testutil.Accumulator

	err := acc.GatherError(r.Gather)
	require.Error(t, err)
	require.Contains(t, err.Error(), "/status")
}

func TestPhpFpmGeneratesMetrics_Throw_Error_When_Fpm_Status_Is_Not_Responding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long test in short mode")
	}

	r := &phpfpm{
		Urls: []string{"http://aninvalidone"},
	}

	require.NoError(t, r.Init())

	var acc testutil.Accumulator

	err := acc.GatherError(r.Gather)
	require.Error(t, err)
	require.Contains(t, err.Error(), `unable to connect to phpfpm status page 'http://aninvalidone'`)
	require.Contains(t, err.Error(), `lookup aninvalidone`)
}

func TestPhpFpmGeneratesMetrics_Throw_Error_When_Socket_Path_Is_Invalid(t *testing.T) {
	r := &phpfpm{
		Urls: []string{"/tmp/invalid.sock"},
	}

	require.NoError(t, r.Init())

	var acc testutil.Accumulator

	err := acc.GatherError(r.Gather)
	require.Error(t, err)
	require.Equal(t, `socket doesn't exist "/tmp/invalid.sock"`, err.Error())
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
