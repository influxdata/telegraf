package assistant

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/inputs/memcached"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	"github.com/stretchr/testify/assert"
)

func initAgentAndAssistant(ctx context.Context) (*agent.Agent, *Assistant) {
	c := config.NewConfig()
	_ = c.LoadConfig("../config/testdata/single_plugin.toml")
	ag, _ := agent.NewAgent(c)
	ast, _ := NewAssistant(&AssistantConfig{Host: "localhost:8080", Path: "/echo", RetryInterval: 15}, ag)

	go func() {
		ag.Run(ctx)
	}()

	time.Sleep(2 * time.Second)

	return ag, ast
}

func TestAssistant_GetPluginAsJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ag, ast := initAgentAndAssistant(ctx)

	for inputName := range inputs.Inputs {
		ag.AddInput(inputName)
	}

	for outputName := range outputs.Outputs {
		ag.AddOutput(outputName)
	}

	for _, p := range ag.Config.Inputs {
		name := p.Config.Name
		req := request{GET_PLUGIN, "123", plugin{name, "INPUT", nil}}
		res := ast.getPlugin(req)
		assert.Equal(t, SUCCESS, res.Status)
		_, err := json.Marshal(res)
		if err != nil {
			t.Log(name)
		}
		assert.NoError(t, err)
	}

	for _, p := range ag.Config.Outputs {
		name := p.Config.Name
		req := request{GET_PLUGIN, "123", plugin{name, "OUTPUT", nil}}
		res := ast.getPlugin(req)
		assert.Equal(t, SUCCESS, res.Status)
		_, err := json.Marshal(res)
		if err != nil {
			t.Log(name)
		}
		assert.NoError(t, err)
	}

	cancel()
}

func TestAssistant_GetUnexistingPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx)

	req := request{GET_PLUGIN, "123", plugin{"VACCUM CLEANER", "INPUT", nil}}
	res := ast.getPlugin(req)
	assert.Equal(t, FAILURE, res.Status)
	cancel()
}

func TestAssistant_UpdatePlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx)

	servers := []string{"go", "jo", "bo"}
	unixSockets := []string{"ubuntu"}
	testReq := request{UPDATE_PLUGIN, "69", plugin{"memcached", "INPUT", map[string]interface{}{
		"Servers":     servers,
		"UnixSockets": unixSockets,
	}}}

	response := ast.updatePlugin(testReq)
	t.Log(response)
	assert.Equal(t, SUCCESS, response.Status)

	getReq := request{GET_PLUGIN, "000", plugin{"memcached", "INPUT", map[string]interface{}{}}}
	plugin := ast.getPlugin(getReq)
	data := plugin.Data
	v, ok := data.(*memcached.Memcached)

	assert.True(t, ok)
	assert.Equal(t, "go", v.Servers[0])
	assert.Equal(t, "jo", v.Servers[1])
	assert.Equal(t, "bo", v.Servers[2])
	assert.Equal(t, "ubuntu", v.UnixSockets[0])

	cancel()
}

func TestAssistant_UpdatePlugin_WithInvalidFieldName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx)

	servers := []string{"go", "SLIM JIMS", "SANDWICH"}
	unixSockets := []string{"ubuntu"}
	invalidField := []string{"invalid value"}
	testReq := request{UPDATE_PLUGIN, "69", plugin{"memcached", "INPUT", map[string]interface{}{
		"Servers":      servers,
		"UnixSockets":  unixSockets,
		"InvalidField": invalidField,
	}}}

	response := ast.updatePlugin(testReq)
	assert.Equal(t, FAILURE, response.Status)

	getReq := request{GET_PLUGIN, "000", plugin{"memcached", "INPUT", map[string]interface{}{}}}
	plugin := ast.getPlugin(getReq)
	data := plugin.Data
	v, ok := data.(*memcached.Memcached)

	assert.True(t, ok)
	assert.Equal(t, "localhost", v.Servers[0])
	assert.Equal(t, 0, len(v.UnixSockets))

	cancel()
}

func TestAssistant_UpdatePlugins_WithInvalidFieldType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx)

	servers := []string{"go", "bo", "jo"}
	testString := "harambe"
	invalidField := []string{"invalid value"}
	testReq := request{UPDATE_PLUGIN, "69", plugin{"memcached", "INPUT", map[string]interface{}{
		"Servers":      servers,
		"UnixSockets":  testString,
		"InvalidField": invalidField,
	}}}

	response := ast.updatePlugin(testReq)
	assert.Equal(t, FAILURE, response.Status)

	getReq := request{GET_PLUGIN, "000", plugin{"memcached", "INPUT", map[string]interface{}{}}}
	plugin := ast.getPlugin(getReq)
	data := plugin.Data
	v, ok := data.(*memcached.Memcached)

	assert.True(t, ok)
	t.Log("servers")
	t.Log(v.Servers)
	assert.Equal(t, "localhost", v.Servers[0])
	assert.NotEqual(t, testString, v.UnixSockets)

	cancel()
}

// ? Unsure what Data will contain
// TODO Implement assertions on res.Data
func TestAssistant_GetAllPlugins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx)

	res, err := ast.getAllPlugins()
	assert.NoError(t, err)
	assert.Equal(t, SUCCESS, res.Status)

	cancel()
}

func TestAssistant_GetRunningPlugins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx)

	res, err := ast.getRunningPlugins()

	v, ok := res.Data.(map[string][]string)

	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, "memcached", v["inputs"][0])
	assert.Equal(t, SUCCESS, res.Status)

	cancel()
}
