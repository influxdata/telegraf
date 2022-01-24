package groundwork

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gwos/tcg/sdk/clients"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

const (
	defaultTestAgentID = "ec1676cc-583d-48ee-b035-7fb5ed0fcf88"
	defaultHost        = "telegraf"
)

func TestWrite(t *testing.T) {
	// Generate test metric with default name to test Write logic
	floatMetric := testutil.TestMetric(1.0, "Float")
	floatMetric.AddTag("host", "Host01")
	floatMetric.AddTag("group", "Group01")

	// Simulate Groundwork server that should receive custom metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		// Decode body to use in assertions below
		var obj groundworkObject
		err = json.Unmarshal(body, &obj)
		require.NoError(t, err)

		// Check if server gets valid metrics object
		require.Equal(t, defaultTestAgentID, obj.Context.AgentID)
		require.Equal(t, "Host01", obj.Resources[0].Name)
		require.Equal(t, "Float", obj.Resources[0].Services[0].Name)
		require.Equal(t, 1.0, obj.Resources[0].Services[0].Metrics[0].Value.DoubleValue)
		require.Equal(t, "Group01", obj.Groups[0].GroupName)
		require.Equal(t, "Host01", obj.Groups[0].Resources[0].Name)

		_, err = fmt.Fprintln(w, `OK`)
		require.NoError(t, err)
	}))

	i := Groundwork{
		Server:      server.URL,
		AgentID:     defaultTestAgentID,
		GroupTag:    "group",
		ResourceTag: "host",
		DefaultHost: "telegraf",
		client: clients.GWClient{
			AppName: "telegraf",
			AppType: "TELEGRAF",
			GWConnection: &clients.GWConnection{
				HostName: server.URL,
			},
		},
	}

	err := i.Write([]telegraf.Metric{floatMetric})
	require.NoError(t, err)

	defer server.Close()
}

type groundworkObject struct {
	Context struct {
		AgentID string `json:"agentId"`
	} `json:"context"`
	Resources []struct {
		Name     string `json:"name"`
		Services []struct {
			Name    string `json:"name"`
			Metrics []struct {
				Value struct {
					StringValue string  `json:"stringValue"`
					DoubleValue float64 `json:"doubleValue"`
				} `json:"value"`
			}
		} `json:"services"`
	} `json:"resources"`
	Groups []struct {
		Type      string `json:"type"`
		GroupName string `json:"groupName"`
		Resources []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"resources"`
	} `json:"groups"`
}
