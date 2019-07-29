package jolokia2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestJolokia2_ClientAuthRequest(t *testing.T) {
	var username string
	var password string
	var requests []map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ = r.BasicAuth()

		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &requests)
		if err != nil {
			t.Error(err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plugin := setupPlugin(t, fmt.Sprintf(`
		[jolokia2_agent]
			urls = ["%s/jolokia"]
			username = "sally"
			password = "seashore"
		[[jolokia2_agent.metric]]
			name  = "hello"
			mbean = "hello:foo=bar"
	`, server.URL))

	var acc testutil.Accumulator
	plugin.Gather(&acc)

	if username != "sally" {
		t.Errorf("Expected to post with username %s, but was %s", "sally", username)
	}
	if password != "seashore" {
		t.Errorf("Expected to post with password %s, but was %s", "seashore", password)
	}
	if len(requests) == 0 {
		t.Fatal("Expected to post a request body, but was empty.")
	}

	request := requests[0]
	if expect := "hello:foo=bar"; request["mbean"] != expect {
		t.Errorf("Expected to query mbean %s, but was %s", expect, request["mbean"])
	}
}

func TestJolokia2_ClientProxyAuthRequest(t *testing.T) {
	var requests []map[string]interface{}

	var username string
	var password string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ = r.BasicAuth()

		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &requests)
		if err != nil {
			t.Error(err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plugin := setupPlugin(t, fmt.Sprintf(`
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
	plugin.Gather(&acc)

	if username != "sally" {
		t.Errorf("Expected to post with username %s, but was %s", "sally", username)
	}
	if password != "seashore" {
		t.Errorf("Expected to post with password %s, but was %s", "seashore", password)
	}
	if len(requests) == 0 {
		t.Fatal("Expected to post a request body, but was empty.")
	}

	request := requests[0]
	if expect := "hello:foo=bar"; request["mbean"] != expect {
		t.Errorf("Expected to query mbean %s, but was %s", expect, request["mbean"])
	}

	target, ok := request["target"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected a proxy target, but was empty.")
	}

	if expect := "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"; target["url"] != expect {
		t.Errorf("Expected proxy target url %s, but was %s", expect, target["url"])
	}

	if expect := "jack"; target["user"] != expect {
		t.Errorf("Expected proxy target username %s, but was %s", expect, target["user"])
	}

	if expect := "benimble"; target["password"] != expect {
		t.Errorf("Expected proxy target password %s, but was %s", expect, target["password"])
	}
}
