package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/inputs/memcached"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	"github.com/stretchr/testify/assert"
)

func initAgentAndAssistant(ctx context.Context, configName string) (*agent.Agent, *Assistant) {
	c := config.NewConfig()
	_ = c.LoadConfig("../config/testdata/" + configName + ".toml")
	ag, _ := agent.NewAgent(c)
	ast, _ := NewAssistant(&AssistantConfig{Host: "localhost:8080", Path: "/echo", RetryInterval: 15}, ag)

	go func() {
		ag.Run(ctx)
	}()

	time.Sleep(5 * time.Second)

	return ag, ast
}

func buildRequest(rt requestType, params pluginInfo) (request, error) {
	paramsJSON, err := json.Marshal(params)
	var req request
	if err == nil {
		var pi pluginInfo
		err = json.Unmarshal(paramsJSON, &pi)
		return request{rt, "123", pi}, err
	}
	return req, err
}

func TestAssistant_GetInputPluginSchema(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	req, err := buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{"httpjson", "INPUT", nil})
	res := ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema := res.Data.(schema)
	_, err = json.Marshal(res)
	if err != nil {
		t.Log(err)
	}

	m := s.Types

	assert.NoError(t, err)
	assert.True(t, isSchema)
	assert.Equal(t, "string", m["Name"])
	assert.Equal(t, agent.ArrayFieldSchema{"string", 0}, m["Servers"])
	assert.Equal(t, "string", m["Method"])
	assert.Equal(t, agent.ArrayFieldSchema{"string", 0}, m["TagKeys"])
	assert.Equal(t, agent.MapFieldSchema{"string", "string"}, m["Parameters"])
	assert.Equal(t, agent.MapFieldSchema{"string", "string"}, m["Headers"])
	assert.Equal(t, map[string]interface{}{
		"Duration": "int64",
	}, m["ResponseTimeout"])

	cancel()
}

func TestAssistant_GetOutputPluginSchema(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "slice_comment") // single output plugin http

	req := request{GET_PLUGIN_SCHEMA, "123", pluginInfo{"http", "OUTPUT", nil}}
	res := ast.getSchema(&req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema := res.Data.(schema)
	_, err := json.Marshal(res)
	if err != nil {
		t.Log(err)
	}

	assert.NoError(t, err)
	assert.True(t, isSchema)
	assert.Equal(t, map[string]interface{}{
		"InsecureSkipVerify": "bool",
		"SSLCA":              "string",
		"SSLCert":            "string",
		"SSLKey":             "string",
		"TLSCA":              "string",
		"TLSCert":            "string",
		"TLSKey":             "string",
	}, s.Types["ClientConfig"])
	assert.Equal(t, "string", s.Types["Method"])
	assert.Equal(t, agent.ArrayFieldSchema{"string", 0}, s.Types["Scopes"])
	assert.Equal(t, agent.MapFieldSchema{"string", "string"}, s.Types["Headers"])
	assert.Equal(t, map[string]interface{}{
		"Duration": "int64",
	}, s.Types["Timeout"])

	assert.Equal(t, "http://127.0.0.1:8080/telegraf", s.Defaults["URL"])
	assert.Equal(t, "POST", s.Defaults["Method"])
	cancel()
}
func TestAssistant_GetPlugin(t *testing.T) {
	// Test getting an input plugin
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "telegraf-agent")

	req, err := buildRequest(GET_PLUGIN, pluginInfo{"memcached", "INPUT", nil})
	assert.NoError(t, err)
	res := ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)
	memcachedMap, dataIsMap := res.Data.(map[string]interface{})
	assert.True(t, dataIsMap)
	assert.Equal(t, []string{"localhost"}, memcachedMap["Servers"])

	// Test getting an output plugin
	req2, err2 := buildRequest(GET_PLUGIN, pluginInfo{"influxdb", "OUTPUT", nil})
	assert.NoError(t, err2)
	res2 := ast.handleRequests(&req2)
	assert.Equal(t, SUCCESS, res2.Status)
	_, dataIsMap = res2.Data.(map[string]interface{})
	assert.True(t, dataIsMap)

	cancel()
}

func TestAssistant_ValidatePluginToMap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	for inputName := range inputs.Inputs {
		req, _ := buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{inputName, "INPUT", nil})
		res := ast.handleRequests(&req)
		assert.Equal(t, SUCCESS, res.Status)

		schema, resIsSchema := res.Data.(schema)
		assert.True(t, resIsSchema)
		assert.NotNil(t, schema.Types)
		assert.NotNil(t, schema.Defaults)
	}

	for outputName := range outputs.Outputs {
		req, _ := buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{outputName, "OUTPUT", nil})
		res := ast.handleRequests(&req)
		assert.Equal(t, SUCCESS, res.Status)

		schema, resIsSchema := res.Data.(schema)
		assert.True(t, resIsSchema)
		assert.NotNil(t, schema.Types)
		assert.NotNil(t, schema.Defaults)
	}

	cancel()
}

func TestAssistant_GetUnexistingPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	req, err := buildRequest(GET_PLUGIN, pluginInfo{"VACCUM CLEANER", "INPUT", nil})
	assert.NoError(t, err)
	res := ast.handleRequests(&req)
	assert.Equal(t, FAILURE, res.Status)
	cancel()
}

func TestAssistant_GetNotRunningPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	req, err := buildRequest(GET_PLUGIN, pluginInfo{"cpu", "INPUT", nil})
	assert.NoError(t, err)
	res := ast.handleRequests(&req)
	assert.Equal(t, FAILURE, res.Status)
	cancel()
}
func TestAssistant_UpdatePlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	headers := map[string]string{
		"test": "result",
	}
	URLs := []string{"www.google.com"}

	req, err := buildRequest(UPDATE_PLUGIN, pluginInfo{"http", "INPUT", map[string]interface{}{
		"Headers": headers,
		"URLs":    URLs,
		"Timeout": map[string]interface{}{
			"Duration": 1,
		},
	}})

	assert.NoError(t, err)

	req2, err2 := buildRequest(START_PLUGIN, pluginInfo{"http", "INPUT", nil})
	assert.NoError(t, err2)
	ast.handleRequests(&req2)

	response := ast.handleRequests(&req)
	fmt.Println(response.Data)
	assert.Equal(t, SUCCESS, response.Status)

	req3, err3 := buildRequest(GET_PLUGIN, pluginInfo{"http", "INPUT", nil})
	assert.NoError(t, err3)
	plugin := ast.handleRequests(&req3)
	assert.Equal(t, SUCCESS, plugin.Status)
	data := plugin.Data

	h, isHTTP := data.(*http.HTTP)
	assert.True(t, isHTTP)

	assert.Equal(t, "www.google.com", h.URLs[0])
	assert.Equal(t, 1*time.Nanosecond, h.Timeout.Duration)
	assert.Equal(t, "result", h.Headers["test"])

	cancel()
}

func TestAssistant_UpdatePlugin_WithInvalidFieldName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	servers := []string{"go", "SLIM JIMS", "SANDWICH"}
	unixSockets := []string{"ubuntu"}
	invalidField := []string{"invalid value"}

	req, err := buildRequest(UPDATE_PLUGIN, pluginInfo{"memcached", "INPUT", map[string]interface{}{
		"Servers":      servers,
		"UnixSockets":  unixSockets,
		"InvalidField": invalidField,
	}})
	assert.NoError(t, err)

	response := ast.handleRequests(&req)
	assert.Equal(t, FAILURE, response.Status)

	req2, err2 := buildRequest(GET_PLUGIN, pluginInfo{"memcached", "INPUT", nil})
	assert.NoError(t, err2)
	plugin := ast.handleRequests(&req2)
	data := plugin.Data

	memcached, memcachedOk := data.(*memcached.Memcached)

	assert.True(t, memcachedOk)
	assert.Equal(t, "localhost", memcached.Servers[0])
	assert.Equal(t, 0, len(memcached.UnixSockets))

	cancel()
}

func TestAssistant_UpdatePlugins_WithInvalidFieldType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	servers := []string{"go", "bo", "jo"}
	testString := "harambe"
	invalidField := []string{"invalid value"}
	req, err := buildRequest(UPDATE_PLUGIN, pluginInfo{"memcached", "INPUT", map[string]interface{}{
		"Servers":      servers,
		"UnixSockets":  testString,
		"InvalidField": invalidField,
	}})
	assert.NoError(t, err)

	response := ast.handleRequests(&req)
	assert.Equal(t, FAILURE, response.Status)

	req2, err2 := buildRequest(GET_PLUGIN, pluginInfo{"memcached", "INPUT", nil})
	assert.NoError(t, err2)
	plugin := ast.handleRequests(&req2)
	data := plugin.Data

	memcached, memcachedOk := data.(*memcached.Memcached)

	assert.True(t, memcachedOk)
	assert.Equal(t, "localhost", memcached.Servers[0])
	assert.NotEqual(t, testString, memcached.UnixSockets)

	cancel()
}

func TestAssistant_GetAllPlugins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	getReq := request{GET_ALL_PLUGINS, "000", pluginInfo{"memcached", "INPUT", nil}}
	res := ast.handleRequests(&getReq)
	assert.Equal(t, SUCCESS, res.Status)

	pList, ok := res.Data.(pluginsList)
	assert.True(t, ok)
	assert.Equal(t, len(inputs.Inputs), len(pList.Inputs))

	cancel()
}

func TestAssistant_GetAllRunningPlugins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	getReq := request{GET_RUNNING_PLUGINS, "000", pluginInfo{"memcached", "INPUT", nil}}
	res := ast.handleRequests(&getReq)
	pList, ok := res.Data.(pluginsList)

	assert.True(t, ok)
	assert.Equal(t, 1, len(pList.Inputs))
	assert.Equal(t, 0, len(pList.Outputs))
	assert.Equal(t, "memcached", pList.Inputs[0])
	assert.Equal(t, SUCCESS, res.Status)

	cancel()
}

func TestAssistant_StopSinglePlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	req := request{STOP_PLUGIN, "123", pluginInfo{"memcached", "INPUT", nil}}
	res := ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)

	getReq := request{GET_RUNNING_PLUGINS, "000", pluginInfo{"memcached", "INPUT", nil}}
	res2 := ast.handleRequests(&getReq)

	t.Log(res2)

	time.Sleep(5)
	cancel()
}
