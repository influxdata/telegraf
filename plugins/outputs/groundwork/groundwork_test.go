package groundwork

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gwos/tcg/sdk/clients"
	"github.com/gwos/tcg/sdk/mapping"
	"github.com/gwos/tcg/sdk/transit"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/logger"
	"github.com/influxdata/telegraf/testutil"
)

const (
	defaultTestAgentID = "ec1676cc-583d-48ee-b035-7fb5ed0fcf88"
	defaultHost        = "telegraf"
	defaultAppType     = "TELEGRAF"
	customAppType      = "SYSLOG"
)

func TestWriteWithDebug(t *testing.T) {
	// Generate test metric with default name to test Write logic
	intMetric := testutil.TestMetric(42, "IntMetric")
	srvTok := "88fcf0de5bf7-530b-ee84-d385-cc6761ce"

	// Simulate Groundwork server that should receive custom metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}

		// Decode body to use in assertions below
		var obj transit.ResourcesWithServicesRequest
		if err = json.Unmarshal(body, &obj); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}

		// Check if server gets proper data
		if obj.Resources[0].Services[0].Name != "IntMetric" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "IntMetric", obj.Resources[0].Services[0].Name)
			return
		}
		if *obj.Resources[0].Services[0].Metrics[0].Value.IntegerValue != int64(42) {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %v, actual: %v", int64(42), *obj.Resources[0].Services[0].Metrics[0].Value.IntegerValue)
			return
		}

		// Send back details
		ans := "Content-type: application/json\n\n" + `{"message":"` + srvTok + `"}`
		if _, err = fmt.Fprintln(w, ans); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))

	i := Groundwork{
		Server:              server.URL,
		AgentID:             defaultTestAgentID,
		Username:            config.NewSecret([]byte(`tu ser`)),
		Password:            config.NewSecret([]byte(`pu ser`)),
		DefaultAppType:      defaultAppType,
		DefaultHost:         defaultHost,
		DefaultServiceState: string(transit.ServiceOk),
		ResourceTag:         "host",
		Log:                 testutil.Logger{},
	}

	buf := new(bytes.Buffer)
	require.NoError(t, logger.SetupLogging(&logger.Config{Debug: true}))
	logger.RedirectLogging(buf)

	require.NoError(t, i.Init())
	require.NoError(t, i.Write([]telegraf.Metric{intMetric}))

	require.NoError(t, logger.CloseLogging())
	require.Contains(t, buf.String(), defaultTestAgentID)
	require.Contains(t, buf.String(), srvTok)

	server.Close()
}

func TestWriteWithDefaults(t *testing.T) {
	// Generate test metric with default name to test Write logic
	intMetric := testutil.TestMetric(42, "IntMetric")

	// Simulate Groundwork server that should receive custom metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}

		// Decode body to use in assertions below
		var obj transit.ResourcesWithServicesRequest
		if err = json.Unmarshal(body, &obj); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}

		// Check if server gets proper data
		if obj.Context.AgentID != defaultTestAgentID {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", defaultTestAgentID, obj.Context.AgentID)
			return
		}
		if obj.Context.AppType != customAppType {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", customAppType, obj.Context.AppType)
			return
		}
		if obj.Resources[0].Name != defaultHost {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", defaultHost, obj.Resources[0].Name)
			return
		}
		if _, ok := obj.Resources[0].Properties["Alias"]; ok {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Unexpected host property %q", "Alias")
			return
		}
		if obj.Resources[0].Services[0].Status != transit.MonitorStatus("SERVICE_OK") {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", transit.MonitorStatus("SERVICE_OK"), obj.Resources[0].Services[0].Status)
			return
		}
		if obj.Resources[0].Services[0].Name != "IntMetric" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "IntMetric", obj.Resources[0].Services[0].Name)
			return
		}
		if *obj.Resources[0].Services[0].Metrics[0].Value.IntegerValue != int64(42) {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %v, actual: %v", int64(42), *obj.Resources[0].Services[0].Metrics[0].Value.IntegerValue)
			return
		}
		if len(obj.Groups) != 0 {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("'obj.Groups' should not be empty")
			return
		}

		if _, err = fmt.Fprintln(w, "OK"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))

	i := Groundwork{
		Log:            testutil.Logger{},
		Server:         server.URL,
		AgentID:        defaultTestAgentID,
		DefaultHost:    defaultHost,
		DefaultAppType: customAppType,
		client: clients.GWClient{
			AppName: "telegraf",
			AppType: customAppType,
			GWConnection: clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	err := i.Write([]telegraf.Metric{intMetric})
	require.NoError(t, err)

	server.Close()
}

func TestWriteWithFields(t *testing.T) {
	// Generate test metric with fields to test Write logic
	floatMetric := testutil.TestMetric(1.0, "FloatMetric")
	floatMetric.AddField("value_cr", 3.0)
	floatMetric.AddField("value_wn", 2.0)
	floatMetric.AddField("message", "Test Message")
	floatMetric.AddField("status", "SERVICE_WARNING")

	// Simulate Groundwork server that should receive custom metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}

		// Decode body to use in assertions below
		var obj transit.ResourcesWithServicesRequest
		if err = json.Unmarshal(body, &obj); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}

		// Check if server gets proper data
		if obj.Resources[0].Services[0].LastPluginOutput != "Test Message" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "Test Message", obj.Resources[0].Services[0].LastPluginOutput)
			return
		}
		if obj.Resources[0].Services[0].Status != transit.MonitorStatus("SERVICE_WARNING") {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", transit.MonitorStatus("SERVICE_WARNING"), obj.Resources[0].Services[0].Status)
			return
		}
		if dt := float64(1.0) - *obj.Resources[0].Services[0].Metrics[0].Value.DoubleValue; !testutil.WithinDefaultDelta(dt) {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Max difference between %v and %v allowed is %v, but difference was %v",
				float64(1.0), *obj.Resources[0].Services[0].Metrics[0].Value.DoubleValue, testutil.DefaultDelta, dt)
			return
		}
		if dt := float64(3.0) - *obj.Resources[0].Services[0].Metrics[0].Thresholds[0].Value.DoubleValue; !testutil.WithinDefaultDelta(dt) {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Max difference between %v and %v allowed is %v, but difference was %v",
				float64(3.0), *obj.Resources[0].Services[0].Metrics[0].Thresholds[0].Value.DoubleValue, testutil.DefaultDelta, dt)
			return
		}
		if dt := float64(2.0) - *obj.Resources[0].Services[0].Metrics[0].Thresholds[1].Value.DoubleValue; !testutil.WithinDefaultDelta(dt) {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Max difference between %v and %v allowed is %v, but difference was %v",
				float64(2.0), *obj.Resources[0].Services[0].Metrics[0].Thresholds[1].Value.DoubleValue, testutil.DefaultDelta, dt)
			return
		}

		if _, err = fmt.Fprintln(w, "OK"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))

	i := Groundwork{
		Log:            testutil.Logger{},
		Server:         server.URL,
		AgentID:        defaultTestAgentID,
		DefaultHost:    defaultHost,
		DefaultAppType: defaultAppType,
		GroupTag:       "group",
		ResourceTag:    "host",
		client: clients.GWClient{
			AppName: "telegraf",
			AppType: defaultAppType,
			GWConnection: clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	err := i.Write([]telegraf.Metric{floatMetric})
	require.NoError(t, err)

	server.Close()
}

func TestWriteWithTags(t *testing.T) {
	// Generate test metric with tags to test Write logic
	floatMetric := testutil.TestMetric(1.0, "FloatMetric")
	floatMetric.AddField("value_cr", 3.0)
	floatMetric.AddField("value_wn", 2.0)
	floatMetric.AddField("message", "Test Message")
	floatMetric.AddField("status", "SERVICE_WARNING")
	floatMetric.AddTag("value_cr", "9.0")
	floatMetric.AddTag("value_wn", "6.0")
	floatMetric.AddTag("message", "Test Tag")
	floatMetric.AddTag("status", "SERVICE_PENDING")
	floatMetric.AddTag("group-tag", "Group01")
	floatMetric.AddTag("resource-tag", "Host01")
	floatMetric.AddTag("alias-tag", "Host01 Alias")
	floatMetric.AddTag("service-tag", "Service01")
	floatMetric.AddTag("facility", "FACILITY")
	floatMetric.AddTag("severity", "SEVERITY")

	// Simulate Groundwork server that should receive custom metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}

		// Decode body to use in assertions below
		var obj transit.ResourcesWithServicesRequest
		if err = json.Unmarshal(body, &obj); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}

		// Check if server gets proper data
		if obj.Context.AgentID != defaultTestAgentID {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", defaultTestAgentID, obj.Context.AgentID)
			return
		}
		if obj.Context.AppType != defaultAppType {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", defaultAppType, obj.Context.AppType)
			return
		}
		if obj.Resources[0].Name != "Host01" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "Host01", obj.Resources[0].Name)
			return
		}
		if alias, ok := obj.Resources[0].Properties["Alias"]; !ok || *alias.StringValue != "Host01 Alias" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %v", "Host01 Alias", obj.Resources[0].Properties["Alias"])
			return
		}
		for _, tag := range []string{"alias-tag", "service-tag"} {
			if _, ok := obj.Resources[0].Services[0].Properties[tag]; ok {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("Unexpected service property %q", tag)
				return
			}
		}
		if obj.Resources[0].Services[0].Name != "Service01" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "Service01", obj.Resources[0].Services[0].Name)
			return
		}
		if *obj.Resources[0].Services[0].Properties["facility"].StringValue != "FACILITY" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "FACILITY", *obj.Resources[0].Services[0].Properties["facility"].StringValue)
			return
		}
		if *obj.Resources[0].Services[0].Properties["severity"].StringValue != "SEVERITY" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "SEVERITY", *obj.Resources[0].Services[0].Properties["severity"].StringValue)
			return
		}
		if obj.Groups[0].GroupName != "Group01" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "Group01", obj.Groups[0].GroupName)
			return
		}
		if obj.Groups[0].Resources[0].Name != "Host01" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "Host01", obj.Groups[0].Resources[0].Name)
			return
		}
		if obj.Resources[0].Services[0].LastPluginOutput != "Test Tag" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", "Test Tag", obj.Resources[0].Services[0].LastPluginOutput)
			return
		}
		if obj.Resources[0].Services[0].Status != transit.MonitorStatus("SERVICE_PENDING") {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Not equal, expected: %q, actual: %q", transit.MonitorStatus("SERVICE_PENDING"), obj.Resources[0].Services[0].Status)
			return
		}
		if dt := float64(1.0) - *obj.Resources[0].Services[0].Metrics[0].Value.DoubleValue; !testutil.WithinDefaultDelta(dt) {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Max difference between %v and %v allowed is %v, but difference was %v",
				float64(1.0), *obj.Resources[0].Services[0].Metrics[0].Value.DoubleValue, testutil.DefaultDelta, dt)
			return
		}
		if dt := float64(9.0) - *obj.Resources[0].Services[0].Metrics[0].Thresholds[0].Value.DoubleValue; !testutil.WithinDefaultDelta(dt) {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Max difference between %v and %v allowed is %v, but difference was %v",
				float64(9.0), *obj.Resources[0].Services[0].Metrics[0].Thresholds[0].Value.DoubleValue, testutil.DefaultDelta, dt)
			return
		}
		if dt := float64(6.0) - *obj.Resources[0].Services[0].Metrics[0].Thresholds[1].Value.DoubleValue; !testutil.WithinDefaultDelta(dt) {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("Max difference between %v and %v allowed is %v, but difference was %v",
				float64(6.0), *obj.Resources[0].Services[0].Metrics[0].Thresholds[1].Value.DoubleValue, testutil.DefaultDelta, dt)
			return
		}

		if _, err = fmt.Fprintln(w, "OK"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))

	i := Groundwork{
		Log:            testutil.Logger{},
		Server:         server.URL,
		AgentID:        defaultTestAgentID,
		DefaultHost:    defaultHost,
		DefaultAppType: defaultAppType,
		GroupTag:       "group-tag",
		ResourceTag:    "resource-tag",
		AliasTag:       "alias-tag",
		ServiceTag:     "service-tag",
		client: clients.GWClient{
			AppName: "telegraf",
			AppType: defaultAppType,
			GWConnection: clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	err := i.Write([]telegraf.Metric{floatMetric})
	require.NoError(t, err)

	server.Close()
}

func TestWriteWithMappings(t *testing.T) {
	floatMetric := testutil.TestMetric(1.0, "FloatMetric")
	floatMetric.AddTag("host", "host01.example.com")
	floatMetric.AddTag("cluster", "prod")
	floatMetric.AddTag("namespace", "web")
	floatMetric.AddTag("check", "http")
	floatMetric.AddTag("state", "warning")
	floatMetric.AddTag("status", "SERVICE_PENDING")
	floatMetric.AddTag("message", "Test Tag")

	server, requests := newCaptureServer(t)

	i := Groundwork{
		Log:            testutil.Logger{},
		Server:         server.URL,
		AgentID:        defaultTestAgentID,
		DefaultHost:    defaultHost,
		DefaultAppType: defaultAppType,
		GroupTag:       "group",
		ResourceTag:    "host",
		MapHostGroup:   mapping.Mappings{*mapping.NewMapping("cluster,namespace", "^(.+),(.+)$", "$1-$2")},
		MapHostAlias:   mapping.Mappings{*mapping.NewMapping("host", "^(.+)$", "alias-$1")},
		MapHostName: mapping.Mappings{
			*mapping.NewMapping("missing", "(.*)", "$1"),
			*mapping.NewMapping("host", "^([^.]+)", "$1"),
		},
		MapService: mapping.Mappings{*mapping.NewMapping("check", "(.*)", "svc-$1")},
		MapStatus:  mapping.Mappings{*mapping.NewMapping("state", "^warning$", "SERVICE_WARNING")},
		MapMessage: mapping.Mappings{*mapping.NewMapping("", "", "Mapped Message")},
		client: clients.GWClient{
			AppName: "telegraf",
			AppType: defaultAppType,
			GWConnection: clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	require.NoError(t, i.Write([]telegraf.Metric{floatMetric}))

	require.Len(t, *requests, 1)
	received := (*requests)[0]
	require.Len(t, received.Resources, 1)
	res := received.Resources[0]
	require.Equal(t, "host01", res.Name)
	require.Contains(t, res.Properties, "Alias")
	require.Equal(t, "alias-host01.example.com", *res.Properties["Alias"].StringValue)
	require.Len(t, res.Services, 1)
	require.Equal(t, "svc-http", res.Services[0].Name)
	require.Equal(t, "host01", res.Services[0].Owner)
	require.Equal(t, transit.ServiceWarning, res.Services[0].Status)
	require.Equal(t, "Mapped Message", res.Services[0].LastPluginOutput)
	require.Len(t, received.Groups, 1)
	require.Equal(t, "prod-web", received.Groups[0].GroupName)
	require.Equal(t, "host01", received.Groups[0].Resources[0].Name)
}

func TestWriteMappingRules(t *testing.T) {
	newMetric := func(name, host string, tags map[string]string, message string) telegraf.Metric {
		m := testutil.TestMetric(1.0, name)
		m.AddTag("host", host)
		m.AddTag("check", name)
		for k, v := range tags {
			m.AddTag(k, v)
		}
		if message != "" {
			m.AddField("message", message)
		}
		return m
	}
	// Mapped from the string field "message".
	matched := newMetric("matched", "host01.example.com", nil, "disk full")
	// The "message" tag takes precedence over the field.
	tagFirst := newMetric("tagFirst", "host01.example.com", map[string]string{"message": "intrusion"}, "disk full")
	// Optional mappings that do not apply keep their tag-based defaults.
	fallback := newMetric("fallback", "host02.example.com",
		map[string]string{"group": "Group02", "host_alias": "Alias02", "state": "bogus"}, "all good")
	// Dropped: hostname or service mapping does not match, or map_ignore matches.
	badHost := newMetric("badHost", "-bad-", nil, "")
	noService := newMetric("noService", "host01.example.com", nil, "")
	noService.RemoveTag("check")
	ignored := newMetric("ignored", "host01.example.com", nil, "DB client connected")

	server, requests := newCaptureServer(t)

	i := Groundwork{
		Log:                 testutil.Logger{},
		Server:              server.URL,
		AgentID:             defaultTestAgentID,
		DefaultHost:         defaultHost,
		DefaultAppType:      defaultAppType,
		DefaultServiceState: string(transit.ServiceOk),
		GroupTag:            "group",
		ResourceTag:         "host",
		AliasTag:            "host_alias",
		MapIgnore:           mapping.Mappings{*mapping.NewMapping("message", "client connected", "ignore")},
		MapHostName:         mapping.Mappings{*mapping.NewMapping("host", `^(\w+)\.`, "$1")},
		MapService:          mapping.Mappings{*mapping.NewMapping("check", "(.+)", "svc-$1")},
		MapHostGroup:        mapping.Mappings{*mapping.NewMapping("cluster", "(.+)", "$1")},
		MapHostAlias:        mapping.Mappings{*mapping.NewMapping("alias_src", "(.+)", "$1")},
		MapStatus: mapping.Mappings{
			*mapping.NewMapping("state", "(.*)", "$1"),
			*mapping.NewMapping("message", "intrusion", string(transit.ServiceUnscheduledCritical)),
			*mapping.NewMapping("message", "disk full", string(transit.ServiceWarning)),
		},
		MapMessage: mapping.Mappings{*mapping.NewMapping("summary", "(.+)", "$1")},
		client: clients.GWClient{
			AppName: "telegraf",
			AppType: defaultAppType,
			GWConnection: clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	require.NoError(t, i.Write([]telegraf.Metric{matched, tagFirst, fallback, badHost, noService, ignored}))
	require.Len(t, *requests, 1)
	received := (*requests)[0]

	services := make(map[string]transit.MonitoredService)
	hosts := make(map[string]transit.MonitoredResource)
	for _, res := range received.Resources {
		hosts[res.Name] = res
		for _, svc := range res.Services {
			services[svc.Name] = svc
		}
	}
	require.Len(t, services, 3)
	require.Equal(t, transit.ServiceWarning, services["svc-matched"].Status)
	require.Equal(t, "disk full", services["svc-matched"].LastPluginOutput)
	require.Equal(t, transit.ServiceUnscheduledCritical, services["svc-tagFirst"].Status)
	require.Equal(t, "host01", services["svc-tagFirst"].Owner)

	require.Equal(t, transit.ServiceOk, services["svc-fallback"].Status)
	require.Equal(t, "all good", services["svc-fallback"].LastPluginOutput)
	require.Equal(t, "host02", services["svc-fallback"].Owner)
	require.Equal(t, "Alias02", *hosts["host02"].Properties["Alias"].StringValue)
	require.Len(t, received.Groups, 1)
	require.Equal(t, "Group02", received.Groups[0].GroupName)

	// A batch where every metric is dropped must not be sent at all.
	require.NoError(t, i.Write([]telegraf.Metric{badHost, noService, ignored}))
	require.Len(t, *requests, 1)
}

func TestInitMappingCompileError(t *testing.T) {
	i := Groundwork{
		Log:                 testutil.Logger{},
		Server:              "http://localhost",
		AgentID:             defaultTestAgentID,
		Username:            config.NewSecret([]byte(`tu ser`)),
		Password:            config.NewSecret([]byte(`pu ser`)),
		DefaultAppType:      defaultAppType,
		DefaultHost:         defaultHost,
		DefaultServiceState: string(transit.ServiceOk),
		ResourceTag:         "host",
		MapService:          mapping.Mappings{{Tag: "service", Matcher: "(", Template: "$1"}},
	}
	require.ErrorContains(t, i.Init(), "map_service")
}

// newCaptureServer starts a fake GroundWork server that records every request it decodes.
func newCaptureServer(t *testing.T) (*httptest.Server, *[]transit.ResourcesWithServicesRequest) {
	var requests []transit.ResourcesWithServicesRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var obj transit.ResourcesWithServicesRequest
		if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
		requests = append(requests, obj)
		if _, err := fmt.Fprintln(w, "OK"); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	return server, &requests
}
