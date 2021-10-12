package gw8

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
	defaultTestAgentID     = "ec1676cc-583d-48ee-b035-7fb5ed0fcf88"
	defaultTestAppType     = "TELEGRAF"
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
		assert.Equal(t, obj["context"].(map[string]interface{})["appType"], defaultTestAppType)
		assert.Equal(t, obj["context"].(map[string]interface{})["agentId"], defaultTestAgentID)
		assert.Equal(t, obj["resources"].([]interface{})[0].(map[string]interface{})["name"], "default_telegraf")
		assert.Equal(
			t,
			obj["resources"].([]interface{})[0].(map[string]interface{})["services"].([]interface{})[0].(map[string]interface{})["name"],
			"Float",
		)
		assert.Equal(
			t,
			obj["resources"].([]interface{})[0].(map[string]interface{})["services"].([]interface{})[0].(map[string]interface{})["metrics"].([]interface{})[0].(map[string]interface{})["value"].(map[string]interface{})["doubleValue"].(float64),
			1.0,
		)
		assert.Equal(
			t,
			obj["resources"].([]interface{})[0].(map[string]interface{})["services"].([]interface{})[1].(map[string]interface{})["metrics"].([]interface{})[0].(map[string]interface{})["value"].(map[string]interface{})["stringValue"].(string),
			"Test",
		)

		_, err = fmt.Fprintln(w, `OK`)
		assert.NoError(t, err)
	}))

	i := GW8{
		Server:  server.URL,
		AppType: defaultTestAppType,
		AgentID: defaultTestAgentID,
	}

	err := i.Write([]telegraf.Metric{floatMetric, stringMetric})
	assert.NoError(t, err)

	defer server.Close()
}

type GroundworkObject map[string]interface{}
