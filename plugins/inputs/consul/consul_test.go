package consul

import (
	"net/http"
	"net/url"
	"testing"

	"encoding/json"
	"net/http/httptest"

	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var sampleChecks = []*api.HealthCheck{
	&api.HealthCheck{
		Node:        "localhost",
		CheckID:     "foo.health123",
		Name:        "foo.health",
		Status:      "passing",
		Notes:       "lorem ipsum",
		Output:      "OK",
		ServiceID:   "foo.123",
		ServiceName: "foo",
	},
}

func TestGatherHealtCheck(t *testing.T) {
	expectedFields := map[string]interface{}{
		"check_name": "foo.health",
		"status":     "passing",
		"passing":    1,
		"critical":   0,
		"warning":    0,
		"service_id": "foo.123",
	}

	expectedTags := map[string]string{
		"node":         "localhost",
		"service_name": "foo",
		"check_id":     "foo.health123",
	}

	var acc testutil.Accumulator

	c := &Consul{}
	c.GatherHealthCheck(&acc, sampleChecks)

	acc.AssertContainsTaggedFields(t, "consul_health_checks", expectedFields, expectedTags)
}

func setupTestMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/v1/status/peers", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]string{"10.1.10.12:8300", "10.1.10.11:8300", "10.1.10.10:8300"})
	}))

	mux.Handle("/v1/status/leader", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`"10.1.10.11:8300"`)
	}))

	mux.Handle("/v1/catalog/nodes", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[{"ID":"40e4a748-2192-161a-0510-9bf59fe950b5","Node":"baz","Address":"10.1.10.11",
		"TaggedAddresses":{"lan":"10.1.10.11","wan":"10.1.10.11"},"Meta":{"instance_type":"t2.medium"}},
		{"ID":"8f246b77-f3e1-ff88-5b48-8ec93abf3e05","Node":"foobar","Address":"10.1.10.12",
		"TaggedAddresses":{"lan":"10.1.10.11","wan":"10.1.10.12"},"Meta":{"instance_type":"t2.large"}}]`)
	}))

	mux.Handle("/v1/catalog/services", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"consul": [], "redis": [], "postgresql": ["primary","secondary"]}`)
	}))

	mux.Handle("/v1/health/service/consul", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[{"Node":{"ID":"40e4a748-2192-161a-0510-9bf59fe950b5","Node":"foobar","Address":"10.1.10.12",
		"TaggedAddresses":{"lan":"10.1.10.12","wan":"10.1.10.12"},"Meta":{"instance_type":"t2.medium"}},
		"Service":{"ID":"consul-1","Service":"consul","Tags":null,"Address":"10.1.10.12","Port":8000},
		"Checks":[{"Node":"foobar","CheckID":"service:consul","Name":"Service 'consul' check","Status":"passing",
		"Notes":"","Output":"","ServiceID":"consul","ServiceName":"consul"},{"Node":"foobar","CheckID":"serfHealth",
		"Name":"Serf Health Status","Status":"passing","Notes":"","Output":"","ServiceID":"","ServiceName":""}]}]`)
	}))

	mux.Handle("/v1/health/service/redis", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[{"Node":{"ID":"40e4a748-2192-161a-0510-9bf59fe950b5","Node":"foobar","Address":"10.1.10.12",
		"TaggedAddresses":{"lan":"10.1.10.12","wan":"10.1.10.12"},"Meta":{"instance_type":"t2.medium"}},
		"Service":{"ID":"redis-2","Service":"redis","Tags":null,"Address":"10.1.10.12","Port":8000},
		"Checks":[{"Node":"foobar","CheckID":"service:redis","Name":"Service 'redis' check","Status":"passing",
		"Notes":"","Output":"","ServiceID":"redis","ServiceName":"redis"},{"Node":"foobar","CheckID":"serfHealth",
		"Name":"Serf Health Status","Status":"passing","Notes":"","Output":"","ServiceID":"","ServiceName":""}]}]`)
	}))

	mux.Handle("/v1/health/service/postgresql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[{"Node":{"ID":"40e4a748-2192-161a-0510-9bf59fe950b5","Node":"foobar","Address":"10.1.10.12",
		"TaggedAddresses":{"lan":"10.1.10.12","wan":"10.1.10.12"},"Meta":{"instance_type":"t2.medium"}},
		"Service":{"ID":"postgresql-3","Service":"postgresql","Tags":null,"Address":"10.1.10.12","Port":8000},
		"Checks":[{"Node":"foobar","CheckID":"service:postgresql","Name":"Service 'postgresql' check","Status":"critical",
		"Notes":"","Output":"","ServiceID":"postgresql","ServiceName":"postgresql"},{"Node":"foobar","CheckID":"serfHealth",
		"Name":"Serf Health Status","Status":"critical","Notes":"","Output":"","ServiceID":"","ServiceName":""}]}]`)
	}))

	return mux
}

func TestGatherServerStats(t *testing.T) {
	var acc testutil.Accumulator

	c := &Consul{}

	ts := httptest.NewServer(setupTestMux())

	defer ts.Close()
	parts, _ := url.Parse(ts.URL)
	c.Address = parts.Host
	c.Scheme = parts.Scheme
	c.client, _ = c.createAPIClient()

	err := c.GatherServerStats(&acc)
	require.NoError(t, err)

	expecteFields := map[string]interface{}{
		"peers":    3.0,
		"leader":   0.0,
		"nodes":    2.0,
		"services": 3.0,
	}

	acc.AssertContainsFields(t, "consul_server_stats", expecteFields)
	acc.AssertDoesNotContainMeasurement(t, "consul_service_health")
}

func TestGatherServiceStats(t *testing.T) {
	var acc testutil.Accumulator

	c := &Consul{}

	ts := httptest.NewServer(setupTestMux())

	defer ts.Close()
	parts, _ := url.Parse(ts.URL)
	c.Address = parts.Host
	c.Scheme = parts.Scheme
	c.client, _ = c.createAPIClient()
	c.CollectServiceHealth = true

	err := c.GatherServerStats(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(
		t,
		"consul_service_health",
		map[string]interface{}{"healthy": 1.0},
		map[string]string{"service_name": "consul", "node": "foobar"},
	)

	acc.AssertContainsTaggedFields(
		t,
		"consul_service_health",
		map[string]interface{}{"healthy": 1.0},
		map[string]string{"service_name": "redis", "node": "foobar"},
	)

	acc.AssertContainsTaggedFields(
		t,
		"consul_service_health",
		map[string]interface{}{"healthy": 0.0},
		map[string]string{"service_name": "postgresql", "node": "foobar"},
	)
}
