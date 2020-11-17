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
	"github.com/influxdata/telegraf/plugins/inputs/memcached"
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

func TestAssistant_GetInputPluginSchema(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	req := request{GET_PLUGIN_SCHEMA, "123", pluginInfo{"httpjson", "INPUT", nil}}
	res := ast.getSchema(req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema := res.Data.(schema)
	_, err := json.Marshal(res)
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

	d, _ := json.Marshal(s.Defaults)

	var config map[string]interface{}
	_ = json.Unmarshal([]byte(d), &config)

	fmt.Println(config)

	cancel()
}

func TestAssistant_GetOutputPluginSchema(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "slice_comment") // single output plugin http

	req := request{GET_PLUGIN_SCHEMA, "123", pluginInfo{"http", "OUTPUT", nil}}
	res := ast.getSchema(req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema := res.Data.(schema)
	_, err := json.Marshal(res)
	if err != nil {
		t.Log(err)
	}

	m := s.Types
	// d := s.Defaults

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
	}, m["ClientConfig"])
	assert.Equal(t, "string", m["Method"])
	assert.Equal(t, agent.ArrayFieldSchema{"string", 0}, m["Scopes"])
	assert.Equal(t, agent.MapFieldSchema{"string", "string"}, m["Headers"])
	assert.Equal(t, map[string]interface{}{
		"Duration": "int64",
	}, m["Timeout"])

	// TODO complete check for defaults
	// assert.Equal(t, "http://127.0.0.1:8080/telegraf", d["URL"])
	// assert.Equal(t, false, d["InsecureSkipVerify"])
	// assert.Equal(t, "", d["ClientID"])
	// assert.Equal(t, map[string]interface{}{
	// 	"Duration": 5000000000,
	// }, d["Timeout"])

	_, _ = json.Marshal(s.Defaults)

	// var config map[string]interface{}
	// _ = json.Unmarshal([]byte(d), &config)

	// fmt.Println(config)

	cancel()
}

func TestAssistant_GetSinglePlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	req := request{START_PLUGIN, "123", pluginInfo{"memcached", "INPUT", nil}}
	res := ast.getPlugin(req)
	assert.Equal(t, SUCCESS, res.Status)
	_, memcachedOk := res.Data.(*memcached.Memcached)
	assert.True(t, memcachedOk)
	_, err := json.Marshal(res)
	if err != nil {
		t.Log(err)
	}
	assert.NoError(t, err)
	cancel()
}

func TestAssistant_ValidateGetPluginsWithAllPlugins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ag, ast := initAgentAndAssistant(ctx, "single_plugin")

	for inputName := range inputs.Inputs {
		req := request{START_PLUGIN, "123", pluginInfo{inputName, "INPUT", nil}}
		ast.startPlugin(req)
	}

	for _, p := range ag.Config.Inputs {
		name := p.Config.Name
		req := request{GET_PLUGIN, "123", pluginInfo{name, "INPUT", nil}}
		res := ast.getPlugin(req)
		assert.Equal(t, SUCCESS, res.Status)
		_, err := json.Marshal(res)
		if err != nil {
			t.Log(name)
		}
		assert.NoError(t, err)
	}

	// for outputName := range outputs.Outputs {
	// 	ag.AddOutput(outputName)
	// }

	// for _, p := range ag.Config.Outputs {
	// 	name := p.Config.Name
	// 	req := request{GET_PLUGIN, "123", plugin{name, "OUTPUT", nil}}
	// 	res := ast.getPlugin(req)
	// 	assert.Equal(t, SUCCESS, res.Status)
	// 	_, err := json.Marshal(res)
	// 	if err != nil {
	// 		t.Log(name)
	// 	}
	// 	assert.NoError(t, err)
	// }

	cancel()
}

func TestAssistant_GetUnexistingPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	req := request{GET_PLUGIN, "123", pluginInfo{"VACCUM CLEANER", "INPUT", nil}}
	res := ast.getPlugin(req)
	assert.Equal(t, FAILURE, res.Status)
	cancel()
}

func TestAssistant_GetNotRunningPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	req := request{GET_PLUGIN, "123", pluginInfo{"cpu", "INPUT", nil}}
	res := ast.getPlugin(req)
	assert.Equal(t, FAILURE, res.Status)
	cancel()
}
func TestAssistant_UpdatePlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	servers := []string{"go", "jo", "bo"}
	unixSockets := []string{"ubuntu"}
	testReq := request{UPDATE_PLUGIN, "69", pluginInfo{"memcached", "INPUT", map[string]interface{}{
		"Servers":     servers,
		"UnixSockets": unixSockets,
	}}}

	response := ast.updatePlugin(testReq)
	assert.Equal(t, SUCCESS, response.Status)

	getReq := request{GET_PLUGIN, "000", pluginInfo{"memcached", "INPUT", nil}}
	plugin := ast.getPlugin(getReq)
	assert.Equal(t, SUCCESS, plugin.Status)
	data := plugin.Data

	memcached, memcachedOk := data.(*memcached.Memcached)
	assert.True(t, memcachedOk)

	assert.Equal(t, "go", memcached.Servers[0])
	assert.Equal(t, "jo", memcached.Servers[1])
	assert.Equal(t, "bo", memcached.Servers[2])
	assert.Equal(t, "ubuntu", memcached.UnixSockets[0])

	cancel()
}

func TestAssistant_UpdatePlugin_WithInvalidFieldName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	servers := []string{"go", "SLIM JIMS", "SANDWICH"}
	unixSockets := []string{"ubuntu"}
	invalidField := []string{"invalid value"}
	testReq := request{UPDATE_PLUGIN, "69", pluginInfo{"memcached", "INPUT", map[string]interface{}{
		"Servers":      servers,
		"UnixSockets":  unixSockets,
		"InvalidField": invalidField,
	}}}

	response := ast.updatePlugin(testReq)
	assert.Equal(t, FAILURE, response.Status)

	getReq := request{GET_PLUGIN, "000", pluginInfo{"memcached", "INPUT", nil}}
	plugin := ast.getPlugin(getReq)
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
	testReq := request{UPDATE_PLUGIN, "69", pluginInfo{"memcached", "INPUT", map[string]interface{}{
		"Servers":      servers,
		"UnixSockets":  testString,
		"InvalidField": invalidField,
	}}}

	response := ast.updatePlugin(testReq)
	assert.Equal(t, FAILURE, response.Status)

	getReq := request{GET_PLUGIN, "000", pluginInfo{"memcached", "INPUT", nil}}
	plugin := ast.getPlugin(getReq)
	data := plugin.Data

	memcached, memcachedOk := data.(*memcached.Memcached)

	assert.True(t, memcachedOk)
	assert.Equal(t, "localhost", memcached.Servers[0])
	assert.NotEqual(t, testString, memcached.UnixSockets)

	cancel()
}

// ? Unsure what Data will contain
// TODO Implement assertions on res.Data
func TestAssistant_GetAllPlugins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin")

	getReq := request{GET_ALL_PLUGINS, "000", pluginInfo{"memcached", "INPUT", nil}}
	res := ast.getAllPlugins(getReq)
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
	res := ast.getRunningPlugins(getReq)

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
	res := ast.stopPlugin(req)
	assert.Equal(t, SUCCESS, res.Status)

	getReq2 := request{GET_RUNNING_PLUGINS, "000", pluginInfo{"memcached", "INPUT", nil}}
	res2 := ast.getRunningPlugins(getReq2)

	t.Log(res2)

	time.Sleep(5)
	cancel()
}
