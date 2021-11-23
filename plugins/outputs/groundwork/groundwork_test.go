package groundwork

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	defaultTestAgentID = "ec1676cc-583d-48ee-b035-7fb5ed0fcf88"
	defaultHost        = "telegraf"
)

func TestWrite(t *testing.T) {
	// Generate test metric with default name to test Write logic
	floatMetric := testutil.TestMetric(1, "Float")
	stringMetric := testutil.TestMetric("Test", "String")

	// Simulate Groundwork server that should receive custom metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		// Decode body to use in assertations below
		var obj GroundworkObject
		err = json.Unmarshal(body, &obj)
		assert.NoError(t, err)

		// Check if server gets valid metrics object
		assert.Equal(t, obj.Context.AgentID, defaultTestAgentID)
		assert.Equal(t, obj.Resources[0].Name, defaultHost)
		assert.Equal(
			t,
			obj.Resources[0].Services[0].Name,
			"Float",
		)
		assert.Equal(
			t,
			obj.Resources[0].Services[0].Metrics[0].Value.DoubleValue,
			1.0,
		)
		assert.Equal(
			t,
			obj.Resources[0].Services[1].Metrics[0].Value.StringValue,
			"Test",
		)

		_, err = fmt.Fprintln(w, `OK`)
		assert.NoError(t, err)
	}))

	i := Groundwork{
		Server:      server.URL,
		AgentID:     defaultTestAgentID,
		DefaultHost: "telegraf",
	}

	err := i.Write([]telegraf.Metric{floatMetric, stringMetric})
	assert.NoError(t, err)

	defer server.Close()
}

type GroundworkObject struct {
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
}
