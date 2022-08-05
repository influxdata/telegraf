package groundwork

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gwos/tcg/sdk/clients"
	"github.com/gwos/tcg/sdk/transit"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

const (
	defaultTestAgentID = "ec1676cc-583d-48ee-b035-7fb5ed0fcf88"
	defaultHost        = "telegraf"
	defaultAppType     = "TELEGRAF"
	customAppType      = "SYSLOG"
)

func TestWriteWithDefaults(t *testing.T) {
	// Generate test metric with default name to test Write logic
	intMetric := testutil.TestMetric(42, "IntMetric")

	// Simulate Groundwork server that should receive custom metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		// Decode body to use in assertions below
		var obj transit.ResourcesWithServicesRequest
		err = json.Unmarshal(body, &obj)
		require.NoError(t, err)

		// Check if server gets proper data
		require.Equal(t, defaultTestAgentID, obj.Context.AgentID)
		require.Equal(t, customAppType, obj.Context.AppType)
		require.Equal(t, defaultHost, obj.Resources[0].Name)
		require.Equal(t, transit.MonitorStatus("SERVICE_OK"), obj.Resources[0].Services[0].Status)
		require.Equal(t, "IntMetric", obj.Resources[0].Services[0].Name)
		require.Equal(t, int64(42), *obj.Resources[0].Services[0].Metrics[0].Value.IntegerValue)
		require.Equal(t, 0, len(obj.Groups))

		_, err = fmt.Fprintln(w, "OK")
		require.NoError(t, err)
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
			GWConnection: &clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	err := i.Write([]telegraf.Metric{intMetric})
	require.NoError(t, err)

	defer server.Close()
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
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		// Decode body to use in assertions below
		var obj transit.ResourcesWithServicesRequest
		err = json.Unmarshal(body, &obj)
		require.NoError(t, err)

		// Check if server gets proper data
		require.Equal(t, "Test Message", obj.Resources[0].Services[0].LastPluginOutput)
		require.Equal(t, transit.MonitorStatus("SERVICE_WARNING"), obj.Resources[0].Services[0].Status)
		require.Equal(t, float64(1.0), *obj.Resources[0].Services[0].Metrics[0].Value.DoubleValue)
		require.Equal(t, float64(3.0), *obj.Resources[0].Services[0].Metrics[0].Thresholds[0].Value.DoubleValue)
		require.Equal(t, float64(2.0), *obj.Resources[0].Services[0].Metrics[0].Thresholds[1].Value.DoubleValue)

		_, err = fmt.Fprintln(w, "OK")
		require.NoError(t, err)
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
			GWConnection: &clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	err := i.Write([]telegraf.Metric{floatMetric})
	require.NoError(t, err)

	defer server.Close()
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
	floatMetric.AddTag("service", "Service01")
	floatMetric.AddTag("facility", "FACILITY")
	floatMetric.AddTag("severity", "SEVERITY")

	// Simulate Groundwork server that should receive custom metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		// Decode body to use in assertions below
		var obj transit.ResourcesWithServicesRequest
		err = json.Unmarshal(body, &obj)
		require.NoError(t, err)

		// Check if server gets proper data
		require.Equal(t, defaultTestAgentID, obj.Context.AgentID)
		require.Equal(t, defaultAppType, obj.Context.AppType)
		require.Equal(t, "Host01", obj.Resources[0].Name)
		require.Equal(t, "Service01", obj.Resources[0].Services[0].Name)
		require.Equal(t, "FACILITY", *obj.Resources[0].Services[0].Properties["facility"].StringValue)
		require.Equal(t, "SEVERITY", *obj.Resources[0].Services[0].Properties["severity"].StringValue)
		require.Equal(t, "Group01", obj.Groups[0].GroupName)
		require.Equal(t, "Host01", obj.Groups[0].Resources[0].Name)
		require.Equal(t, "Test Tag", obj.Resources[0].Services[0].LastPluginOutput)
		require.Equal(t, transit.MonitorStatus("SERVICE_PENDING"), obj.Resources[0].Services[0].Status)
		require.Equal(t, float64(1.0), *obj.Resources[0].Services[0].Metrics[0].Value.DoubleValue)
		require.Equal(t, float64(9.0), *obj.Resources[0].Services[0].Metrics[0].Thresholds[0].Value.DoubleValue)
		require.Equal(t, float64(6.0), *obj.Resources[0].Services[0].Metrics[0].Thresholds[1].Value.DoubleValue)

		_, err = fmt.Fprintln(w, "OK")
		require.NoError(t, err)
	}))

	i := Groundwork{
		Log:            testutil.Logger{},
		Server:         server.URL,
		AgentID:        defaultTestAgentID,
		DefaultHost:    defaultHost,
		DefaultAppType: defaultAppType,
		GroupTag:       "group-tag",
		ResourceTag:    "resource-tag",
		client: clients.GWClient{
			AppName: "telegraf",
			AppType: defaultAppType,
			GWConnection: &clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	err := i.Write([]telegraf.Metric{floatMetric})
	require.NoError(t, err)

	defer server.Close()
}
