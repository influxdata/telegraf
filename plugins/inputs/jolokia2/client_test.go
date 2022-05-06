package jolokia2_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestJolokia2_ClientAuthRequest(t *testing.T) {
	var username string
	var password string
	var requests []map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ = r.BasicAuth()

		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &requests))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plugin := SetupPlugin(t, fmt.Sprintf(`
		[jolokia2_agent]
			urls = ["%s/jolokia"]
			username = "sally"
			password = "seashore"
		[[jolokia2_agent.metric]]
			name  = "hello"
			mbean = "hello:foo=bar"
	`, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.EqualValuesf(t, "sally", username, "Expected to post with username %s, but was %s", "sally", username)
	require.EqualValuesf(t, "seashore", password, "Expected to post with password %s, but was %s", "seashore", password)
	require.NotZero(t, len(requests), "Expected to post a request body, but was empty.")

	request := requests[0]["mbean"]
	require.EqualValuesf(t, "hello:foo=bar", request, "Expected to query mbean %s, but was %s", "hello:foo=bar", request)
}

func TestJolokia2_ClientProxyAuthRequest(t *testing.T) {
	var requests []map[string]interface{}

	var username string
	var password string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ = r.BasicAuth()

		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &requests))
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, "[]")
		require.NoError(t, err)
	}))
	defer server.Close()

	plugin := SetupPlugin(t, fmt.Sprintf(`
		[jolokia2_proxy]
			url = "%s/jolokia"
			username = "sally"
			password = "seashore"

		[[jolokia2_proxy.target]]
			url = "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
			username = "jack"
			password = "benimble"

		[[jolokia2_proxy.metric]]
			name  = "hello"
			mbean = "hello:foo=bar"
	`, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.EqualValuesf(t, "sally", username, "Expected to post with username %s, but was %s", "sally", username)
	require.EqualValuesf(t, "seashore", password, "Expected to post with password %s, but was %s", "seashore", password)
	require.NotZero(t, len(requests), "Expected to post a request body, but was empty.")

	request := requests[0]
	expected := "hello:foo=bar"
	require.EqualValuesf(t, expected, request["mbean"], "Expected to query mbean %s, but was %s", expected, request["mbean"])

	target, ok := request["target"].(map[string]interface{})
	require.True(t, ok, "Expected a proxy target, but was empty.")

	expected = "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
	require.Equalf(t, expected, target["url"], "Expected proxy target url %s, but was %s", expected, target["url"])
	expected = "jack"
	require.Equalf(t, expected, target["user"], "Expected proxy target username %s, but was %s", expected, target["user"])
	expected = "benimble"
	require.Equalf(t, expected, target["password"], "Expected proxy target username %s, but was %s", expected, target["password"])
}
