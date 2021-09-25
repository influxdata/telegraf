package nginx_upstream_check

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const sampleStatusResponse = `
{
  "servers": {
    "total": 2,
    "generation": 1,
    "server": [
      {
        "index": 0,
        "upstream": "upstream-1",
        "name": "127.0.0.1:8081",
        "status": "up",
        "rise": 1000,
        "fall": 0,
        "type": "http",
        "port": 0
      },
      {
        "index": 1,
        "upstream": "upstream-2",
        "name": "127.0.0.1:8082",
        "status": "down",
        "rise": 0,
        "fall": 2000,
        "type": "tcp",
        "port": 8080
      }
    ]
  }
}
`

func TestNginxUpstreamCheckData(test *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		var response string

		require.Equal(test, request.URL.Path, "/status", "Cannot handle request")

		response = sampleStatusResponse
		responseWriter.Header()["Content-Type"] = []string{"application/json"}

		_, err := fmt.Fprintln(responseWriter, response)
		require.NoError(test, err)
	}))
	defer testServer.Close()

	check := NewNginxUpstreamCheck()
	check.URL = fmt.Sprintf("%s/status", testServer.URL)

	var accumulator testutil.Accumulator

	checkError := check.Gather(&accumulator)
	require.NoError(test, checkError)

	accumulator.AssertContainsTaggedFields(
		test,
		"nginx_upstream_check",
		map[string]interface{}{
			"status":      string("up"),
			"status_code": uint8(1),
			"rise":        uint64(1000),
			"fall":        uint64(0),
		},
		map[string]string{
			"upstream": string("upstream-1"),
			"type":     string("http"),
			"name":     string("127.0.0.1:8081"),
			"port":     string("0"),
			"url":      fmt.Sprintf("%s/status", testServer.URL),
		})

	accumulator.AssertContainsTaggedFields(
		test,
		"nginx_upstream_check",
		map[string]interface{}{
			"status":      string("down"),
			"status_code": uint8(2),
			"rise":        uint64(0),
			"fall":        uint64(2000),
		},
		map[string]string{
			"upstream": string("upstream-2"),
			"type":     string("tcp"),
			"name":     string("127.0.0.1:8082"),
			"port":     string("8080"),
			"url":      fmt.Sprintf("%s/status", testServer.URL),
		})
}

func TestNginxUpstreamCheckRequest(test *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		var response string

		require.Equal(test, request.URL.Path, "/status", "Cannot handle request")

		response = sampleStatusResponse
		responseWriter.Header()["Content-Type"] = []string{"application/json"}

		_, err := fmt.Fprintln(responseWriter, response)
		require.NoError(test, err)

		require.Equal(test, request.Method, "POST")
		require.Equal(test, request.Header.Get("X-Test"), "test-value")
		require.Equal(test, request.Header.Get("Authorization"), "Basic dXNlcjpwYXNzd29yZA==")
		require.Equal(test, request.Host, "status.local")
	}))
	defer testServer.Close()

	check := NewNginxUpstreamCheck()
	check.URL = fmt.Sprintf("%s/status", testServer.URL)
	check.Headers["X-test"] = "test-value"
	check.HostHeader = "status.local"
	check.Username = "user"
	check.Password = "password"
	check.Method = "POST"

	var accumulator testutil.Accumulator

	checkError := check.Gather(&accumulator)
	require.NoError(test, checkError)
}
