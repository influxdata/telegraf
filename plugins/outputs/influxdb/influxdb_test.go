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
	require.NoError(t, i.Close())
}

func TestHTTPInflux(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/write":
			// test that database is set properly
			if r.FormValue("db") != "test" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
			}
			// test that user agent is set properly
			if r.UserAgent() != "telegraf" {
				w.WriteHeader(http.StatusTeapot)
				w.Header().Set("Content-Type", "application/json")
			}
			w.WriteHeader(http.StatusNoContent)
			w.Header().Set("Content-Type", "application/json")
		case "/query":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"results":[{}]}`)
		}
	}))
	defer ts.Close()

	i := newInflux()
	i.URLs = []string{ts.URL}
	i.Database = "test"
	i.UserAgent = "telegraf"

	err := i.Connect()
	require.NoError(t, err)
	err = i.Write(testutil.MockMetrics())
	require.NoError(t, err)
	require.NoError(t, i.Close())
}

func TestUDPConnectError(t *testing.T) {
	i := InfluxDB{
		URLs: []string{"udp://foobar:8089"},
	}

	err := i.Connect()
	require.Error(t, err)

	i = InfluxDB{
		URLs: []string{"udp://localhost:9999999"},
	}

	err = i.Connect()
	require.Error(t, err)
}

func TestHTTPConnectError_InvalidURL(t *testing.T) {
	i := InfluxDB{
		URLs: []string{"http://foobar:8089"},
	}

	err := i.Connect()
	require.Error(t, err)

	i = InfluxDB{
		URLs: []string{"http://localhost:9999999"},
	}

	err = i.Connect()
	require.Error(t, err)
}

func TestHTTPConnectError_DatabaseCreateFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/query":
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"results":[{}],"error":"test error"}`)
		}
	}))
	defer ts.Close()

	i := InfluxDB{
		URLs:     []string{ts.URL},
		Database: "test",
	}

	// database creation errors do not return an error from Connect
	// they are only logged.
	err := i.Connect()
	require.NoError(t, err)
	require.NoError(t, i.Close())
}

func TestHTTPError_DatabaseNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/write":
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"results":[{}],"error":"database not found"}`)
		case "/query":
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"results":[{}],"error":"database not found"}`)
		}
	}))
	defer ts.Close()

	i := InfluxDB{
		URLs:     []string{ts.URL},
		Database: "test",
	}

	err := i.Connect()
	require.NoError(t, err)
	err = i.Write(testutil.MockMetrics())
	require.Error(t, err)
	require.NoError(t, i.Close())
}
