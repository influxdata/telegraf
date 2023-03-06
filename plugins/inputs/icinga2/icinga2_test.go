package icinga2

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestIcinga2Default(t *testing.T) {
	// This test should succeed with the default initialization.
	icinga2 := &Icinga2{
		Server:          "https://localhost:5665",
		Objects:         []string{"services"},
		ResponseTimeout: config.Duration(time.Second * 5),
	}
	require.NoError(t, icinga2.Init())

	require.Equal(t, config.Duration(5*time.Second), icinga2.ResponseTimeout)
	require.Equal(t, "https://localhost:5665", icinga2.Server)
	require.Equal(t, []string{"services"}, icinga2.Objects)
}

func TestIcinga2DeprecatedHostConfig(t *testing.T) {
	icinga2 := &Icinga2{
		ObjectType: "hosts", //deprecated
		Objects:    []string{},
	}
	require.NoError(t, icinga2.Init())

	require.Equal(t, []string{"hosts"}, icinga2.Objects)
}

func TestIcinga2DeprecatedServicesConfig(t *testing.T) {
	icinga2 := &Icinga2{
		ObjectType: "services", //deprecated
		Objects:    []string{},
	}
	require.NoError(t, icinga2.Init())

	require.Equal(t, []string{"services"}, icinga2.Objects)
}

const icinga2ServiceResponse = `{
	"results": [
		{
			"attrs": {
				"check_command": "check-bgp-juniper-netconf",
				"display_name": "eq-par.dc2.fr",
				"host_name": "someserverfqdn.net",
				"name": "ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
				"state": 0
			},
			"joins": {},
			"meta": {},
			"name": "eq-par.dc2.fr!ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
			"type": "Service"
		}
	]
}`

func TestGatherServicesStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/objects/services" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(icinga2ServiceResponse))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			t.Logf("Req: %s %s\n", r.Host, r.URL.Path)
		}
	}))
	defer ts.Close()

	var icinga2 = &Icinga2{
		Server:  ts.URL,
		Objects: []string{"services"},
	}
	require.NoError(t, icinga2.Init())
	var acc testutil.Accumulator
	err := icinga2.Gather(&acc)
	require.NoError(t, err)

	requestURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"name":       "ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
		"state_code": int64(0),
	}

	expectedTags := map[string]string{
		"display_name":  "eq-par.dc2.fr",
		"check_command": "check-bgp-juniper-netconf",
		"state":         "ok",
		"source":        "someserverfqdn.net",
		"server":        requestURL.Hostname(),
		"port":          requestURL.Port(),
		"scheme":        "http",
	}

	acc.AssertContainsTaggedFields(t, "icinga2_services", expectedFields, expectedTags)
}

const icinga2HostResponse = `{
	"results": [
		{
			"attrs": {
				"address": "192.168.1.1",
				"check_command": "ping",
				"display_name": "apache",
				"name": "webserver",
				"state": 2.0
			},
			"joins": {},
			"meta": {},
			"name": "webserver",
			"type": "Host"
		}
	]
}
`

func TestGatherHostsStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/objects/hosts" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(icinga2HostResponse))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			t.Logf("Req: %s %s\n", r.Host, r.URL.Path)
		}
	}))
	defer ts.Close()

	var icinga2 = &Icinga2{
		Server:  ts.URL,
		Objects: []string{"hosts"},
	}
	require.NoError(t, icinga2.Init())

	requestURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = icinga2.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"name":       "webserver",
		"state_code": int64(2),
	}

	expectedTags := map[string]string{
		"display_name":  "apache",
		"check_command": "ping",
		"state":         "critical",
		"source":        "webserver",
		"server":        requestURL.Hostname(),
		"port":          requestURL.Port(),
		"scheme":        "http",
	}

	acc.AssertContainsTaggedFields(t, "icinga2_hosts", expectedFields, expectedTags)
}

const icinga2StatusCIB = `{
  "results": [
    {
      "name": "CIB",
      "perfdata": [],
      "status": {
        "active_host_checks": 3.6,
        "avg_latency": 2.187678621145969e-06,
        "max_latency": 0.001603841781616211
      }
    }
  ]
}`

func TestGatherStatusCIB(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/status/CIB" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(icinga2StatusCIB))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			t.Logf("Req: %s %s\n", r.Host, r.URL.Path)
		}
	}))
	defer ts.Close()

	var icinga2 = &Icinga2{
		Server: ts.URL,
		Status: []string{"CIB"},
	}
	require.NoError(t, icinga2.Init())

	var acc testutil.Accumulator
	err := icinga2.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"active_host_checks": float64(3.6),
		"avg_latency":        float64(2.187678621145969e-06),
		"max_latency":        float64(0.001603841781616211),
	}

	expectedTags := map[string]string{
		"component": "CIB",
	}

	acc.AssertContainsTaggedFields(t, "icinga2_status", expectedFields, expectedTags)
}

const icinga2StatusPgsql = `{
  "results": [
    {
      "name": "IdoPgsqlConnection",
      "perfdata": [
        {
          "counter": false,
          "crit": null,
          "label": "idopgsqlconnection_ido-pgsql_queries_rate",
          "max": null,
          "min": null,
          "type": "PerfdataValue",
          "unit": "",
          "value": 649.8666666666667,
          "warn": null
        },
        {
          "counter": false,
          "crit": null,
          "label": "idopgsqlconnection_ido-pgsql_query_queue_item_rate",
          "max": null,
          "min": null,
          "type": "PerfdataValue",
          "unit": "",
          "value": 1295.1166666666666,
          "warn": null
        }
      ]
    }
  ]
}
`

func TestGatherStatusPgsql(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/status/IdoPgsqlConnection" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(icinga2StatusPgsql))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
			t.Logf("Req: %s %s\n", r.Host, r.URL.Path)
		}
	}))
	defer ts.Close()

	var icinga2 = &Icinga2{
		Server: ts.URL,
		Status: []string{"IdoPgsqlConnection"},
	}
	require.NoError(t, icinga2.Init())

	var acc testutil.Accumulator
	err := icinga2.Gather(&acc)
	require.NoError(t, err)

	expectedFields := map[string]interface{}{
		"pgsql_queries_rate":          float64(649.8666666666667),
		"pgsql_query_queue_item_rate": float64(1295.1166666666666),
	}

	expectedTags := map[string]string{
		"component": "IdoPgsqlConnection",
	}

	acc.AssertContainsTaggedFields(t, "icinga2_status", expectedFields, expectedTags)
}
