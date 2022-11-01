package jolokia2_proxy_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	common "github.com/influxdata/telegraf/plugins/common/jolokia2"
	"github.com/influxdata/telegraf/plugins/inputs/jolokia2_proxy"
	"github.com/influxdata/telegraf/testutil"
)

func TestJolokia2_ProxyTargets(t *testing.T) {
	config := `
	[jolokia2_proxy]
		url = "%s"

	[[jolokia2_proxy.target]]
		url = "service:jmx:rmi:///jndi/rmi://target1:9010/jmxrmi"

	[[jolokia2_proxy.target]]
		url = "service:jmx:rmi:///jndi/rmi://target2:9010/jmxrmi"

	[[jolokia2_proxy.metric]]
		name  = "hello"
		mbean = "hello:foo=bar"`

	response := `[{
		"request": {
			"type": "read",
			"mbean": "hello:foo=bar",
			"target": {
				"url": "service:jmx:rmi:///jndi/rmi://target1:9010/jmxrmi"
			}
		},
		"value": 123,
		"status": 200
	}, {
		"request": {
			"type": "read",
			"mbean": "hello:foo=bar",
			"target": {
				"url": "service:jmx:rmi:///jndi/rmi://target2:9010/jmxrmi"
			}
		},
		"value": 456,
		"status": 200
	}]`

	server := setupServer(response)
	defer server.Close()
	plugin := SetupPlugin(t, fmt.Sprintf(config, server.URL))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t, "hello", map[string]interface{}{
		"value": 123.0,
	}, map[string]string{
		"jolokia_proxy_url": server.URL,
		"jolokia_agent_url": "service:jmx:rmi:///jndi/rmi://target1:9010/jmxrmi",
	})
	acc.AssertContainsTaggedFields(t, "hello", map[string]interface{}{
		"value": 456.0,
	}, map[string]string{
		"jolokia_proxy_url": server.URL,
		"jolokia_agent_url": "service:jmx:rmi:///jndi/rmi://target2:9010/jmxrmi",
	})
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

func TestFillFields(t *testing.T) {
	complexPoint := map[string]interface{}{"Value": []interface{}{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
	scalarPoint := []interface{}{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	results := map[string]interface{}{}
	common.NewPointBuilder(common.Metric{Name: "test", Mbean: "complex"}, []string{"this", "that"}, "/").FillFields("", complexPoint, results)
	require.Equal(t, map[string]interface{}{}, results)

	results = map[string]interface{}{}
	common.NewPointBuilder(common.Metric{Name: "test", Mbean: "scalar"}, []string{"this", "that"}, "/").FillFields("", scalarPoint, results)
	require.Equal(t, map[string]interface{}{}, results)
}

func setupServer(resp string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, resp)
	}))
}

func SetupPlugin(t *testing.T, conf string) telegraf.Input {
	table, err := toml.Parse([]byte(conf))
	if err != nil {
		t.Fatalf("Unable to parse config! %v", err)
	}

	for name := range table.Fields {
		object := table.Fields[name]
		if name == "jolokia2_proxy" {
			plugin := jolokia2_proxy.JolokiaProxy{
				Metrics:               []common.MetricConfig{},
				DefaultFieldSeparator: ".",
			}

			if err := toml.UnmarshalTable(object.(*ast.Table), &plugin); err != nil {
				t.Fatalf("Unable to parse jolokia_proxy plugin config! %v", err)
			}

			return &plugin
		}
	}

	return nil
}
