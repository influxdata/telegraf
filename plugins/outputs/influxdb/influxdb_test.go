package influxdb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

func TestUDPInflux(t *testing.T) {
	i := InfluxDB{
		URLs: []string{"udp://localhost:8089"},
	}

	err := i.Connect()
	require.NoError(t, err)
	err = i.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestHTTPInflux(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"results":[{}]}`)
	}))
	defer ts.Close()

	i := InfluxDB{
		URLs: []string{ts.URL},
	}

	err := i.Connect()
	require.NoError(t, err)
	err = i.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

// TestHTTPInfluxFollow301RedirectFalse verifies that the default behavior is for
// InsecureFollowRedirect is false and that it will generate an error.
func TestHTTPInfluxFollow301RedirectFalse(t *testing.T) {

	// The influxDB HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"results":[{}]}`)
	}))
	defer ts.Close()

	// A intermediate HTTP server which sends a redirect
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "test/html")
		w.Header().Set("Location", ts.URL)
		w.WriteHeader(http.StatusMovedPermanently)
		fmt.Fprintln(w, `<html><head><title>301 Moved Permanently</title></head>`)
		fmt.Fprintln(w, `<body bgcolor="white"><center><h1>301 Moved Permanently</h1></center><hr><center>nginx/1.10.0 (Ubuntu)</center></body></html>`)
	}))
	defer redirectServer.Close()

	i := InfluxDB{
		URLs: []string{redirectServer.URL},
	}

	err := i.Connect()
	require.NoError(t, err)

	// This should be empty because we have one URL, with a redirect. Since we don't
	//  follow it, the connection is not added.
	if len(i.conns) != 0 {
		t.Errorf("Did not get an empty list of connections: %s\n", i.conns)
	}

}

// TestHTTPInfluxFollow301RedirectTrue verifies that if InsecureFollowRedirect
// is set to true, then Ping() will be called and the CheckRedirect function will
// be enabled and will set URL to the last redirect in the chain. Up to 10
// redirects max will be followed ( This is net/http behavior )
func TestHTTPInfluxFollow301RedirectTrue(t *testing.T) {

	// The influxDB HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"results":[{}]}`)
	}))
	defer ts.Close()

	// A intermediate HTTP server which sends a redirect
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "test/html")
		w.Header().Set("Location", ts.URL)
		w.WriteHeader(http.StatusMovedPermanently)
		fmt.Fprintln(w, `<html><head><title>301 Moved Permanently</title></head>`)
		fmt.Fprintln(w, `<body bgcolor="white"><center><h1>301 Moved Permanently</h1></center><hr><center>nginx/1.10.0 (Ubuntu)</center></body></html>`)
	}))
	defer redirectServer.Close()

	i := InfluxDB{
		URLs: []string{redirectServer.URL},
		InsecureFollowRedirect: true,
	}

	err := i.Connect()

	require.NoError(t, err)

}
