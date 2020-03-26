package apm_server

import (
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func newTestServer() *APMServer {
	server := &APMServer{
		Log:            testutil.Logger{},
		ServiceAddress: "localhost:0",
	}
	_ = internal.SetVersion("0.0.1")
	return server
}

func TestNotMappedPath(t *testing.T) {

	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	// get server information
	resp, err := http.Get(createURL(server, "http", "/not-mapped", ""))
	require.NoError(t, err)
	require.Equal(t, "application/json", resp.Header["Content-Type"][0])
	require.EqualValues(t, 404, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "{\"error\":\"404 page not found\"}", string(body))
}

func TestServerInformation(t *testing.T) {
	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	server.buildSHA = "bc4d9a286a65b4283c2462404add86a26be61dca"
	server.buildDate = time.Date(
		2009, 11, 17, 20, 34, 58, 0, time.UTC)

	// get server information
	resp, err := http.Get(createURL(server, "http", "/", ""))
	require.NoError(t, err)
	require.Equal(t, "application/json", resp.Header["Content-Type"][0])
	require.EqualValues(t, 200, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "{\"build_date\":\"2009-11-17T20:34:58Z\","+
		"\"build_sha\":\"bc4d9a286a65b4283c2462404add86a26be61dca\",\"version\":\"0.0.1\"}", string(body))
}

func TestAgentConfiguration(t *testing.T) {
	server := newTestServer()
	acc := &testutil.Accumulator{}
	require.NoError(t, server.Init())
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	// get agent configuration
	resp, err := http.Get(createURL(server, "http", "/config/v1/agents", "service.name=TEST"))
	require.NoError(t, err)
	require.EqualValues(t, 403, resp.StatusCode)
	require.Equal(t, 0, len(acc.Metrics))
	defer resp.Body.Close()
}

func createURL(server *APMServer, scheme string, path string, rawquery string) string {
	u := url.URL{
		Scheme:   scheme,
		Host:     "localhost:" + strconv.Itoa(server.port),
		Path:     path,
		RawQuery: rawquery,
	}
	return u.String()
}
