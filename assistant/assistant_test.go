package assistant

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	"github.com/stretchr/testify/assert"
)

func initAgentAndAssistant(ctx context.Context, configName string, t *testing.T) (*agent.Agent, *Assistant) {
	c := config.NewConfig()
	_ = c.LoadConfig("../config/testdata/" + configName + ".toml")
	ag, _ := agent.NewAgent(c)
	ast, _ := NewAssistant(&AssistantConfig{Host: "localhost:8080", Path: "/echo", RetryInterval: 15}, ag)

	go func() {
		ag.Run(ctx)
	}()

	time.Sleep(2 * time.Second)

	t.Cleanup(func() {
		os.Remove("./updated_config.conf")
	})

	return ag, ast
}

func buildRequest(rt requestType, params pluginInfo) (request, error) {
	paramsJSON, err := json.Marshal(params)
	var req request
	if err == nil {
		var pi pluginInfo
		err = json.Unmarshal(paramsJSON, &pi)
		return request{rt, "123", pi}, err // 123 is dummy request uuid
	}
	return req, err
}

func (ast *Assistant) getPluginID(name string) string {
	getAllReq, _ := buildRequest(GET_RUNNING_PLUGINS, pluginInfo{"", "", nil, ""})

	allRes := ast.handleRequests(&getAllReq)

	pluginsWithID, _ := allRes.Data.(pluginsWithIdList)

	for _, in := range pluginsWithID.Inputs {
		if in["name"] == name {
			return in["id"]
		}
	}

	for _, out := range pluginsWithID.Outputs {
		if out["name"] == name {
			return out["id"]
		}
	}

	return ""
}

func (ast *Assistant) getAllPluginsID(name string) []string {
	getAllReq, _ := buildRequest(GET_RUNNING_PLUGINS, pluginInfo{"", "", nil, ""})

	allRes := ast.handleRequests(&getAllReq)

	pluginsWithID, _ := allRes.Data.(pluginsWithIdList)

	res := []string{}

	for _, in := range pluginsWithID.Inputs {
		if in["name"] == name {
			res = append(res, in["id"])
		}
	}

	for _, out := range pluginsWithID.Outputs {
		if out["name"] == name {
			res = append(res, out["id"])
		}
	}

	return res
}

func TestAssistant_GetInputPluginSchema(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	req, err := buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{"httpjson", "INPUT", nil, ""})
	res := ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema := res.Data.(schema)
	_, err = json.Marshal(res)
	if err != nil {
		t.Log(err)
	}

	// test s.Types
	assert.NoError(t, err)
	assert.True(t, isSchema)
	assert.Equal(t, "string", s.Types["Name"])
	assert.Equal(t, agent.ArrayFieldSchema{"string", 0}, s.Types["Servers"])
	assert.Equal(t, "string", s.Types["Method"])
	assert.Equal(t, agent.ArrayFieldSchema{"string", 0}, s.Types["TagKeys"])
	assert.Equal(t, agent.MapFieldSchema{"string", "string"}, s.Types["Parameters"])
	assert.Equal(t, agent.MapFieldSchema{"string", "string"}, s.Types["Headers"])
	assert.Equal(t, map[string]interface{}{
		"Duration": "int64",
	}, s.Types["ResponseTimeout"])
	// test s.Defaults
	dur, _ := time.ParseDuration("5s")
	assert.Equal(t, map[string]interface{}{
		"Duration": dur,
	}, s.Defaults["ResponseTimeout"])

	// test aerospike (empty default values)
	req, err = buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{"aerospike", "INPUT", nil, ""})
	res = ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema = res.Data.(schema)
	_, err = json.Marshal(res)
	if err != nil {
		t.Log(err)
	}

	assert.NoError(t, err)
	assert.True(t, isSchema)
	assert.Equal(t, 8, len(s.Defaults)) // only bools and ints initialized

	cancel()
}

func TestAssistant_GetOutputPluginSchema(t *testing.T) {
	// currently running plugins are irrelevant to this test
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "slice_comment", t)

	req, err := buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{"http", "OUTPUT", nil, ""})
	res := ast.getSchema(&req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema := res.Data.(schema)
	_, err = json.Marshal(res)
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

	// test application_insights, empty map of length 2
	req, err = buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{"application_insights", "OUTPUT", nil, ""})
	res = ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema = res.Data.(schema)
	_, err = json.Marshal(res)
	if err != nil {
		t.Log(err)
	}

	assert.NoError(t, err)
	assert.True(t, isSchema)

	for _, field := range []string{"EndpointURL", "InstrumentationKey"} {
		_, ok := s.Defaults[field]
		assert.False(t, ok)
	}
	dur, _ := time.ParseDuration("5s")
	assert.Equal(t, map[string]interface{}{
		"Duration": dur,
	}, s.Defaults["Timeout"])
	assert.Equal(t, make(map[string]string, 2), s.Defaults["ContextTagSources"])

	// test cloudwatch (empty defaults)
	req, err = buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{"cloudwatch", "OUTPUT", nil, ""})
	res = ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)

	s, isSchema = res.Data.(schema)
	_, err = json.Marshal(res)
	if err != nil {
		t.Log(err)
	}

	assert.NoError(t, err)
	assert.True(t, isSchema)
	assert.Equal(t, 2, len(s.Defaults)) // only bools and ints intialized

	cancel()
}
func TestAssistant_GetInputPlugin(t *testing.T) {
	// Test getting an input plugin
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	memcachedID := ast.getPluginID("memcached")
	assert.NotEmpty(t, memcachedID)

	req, err := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, memcachedID})
	assert.NoError(t, err)
	res := ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)
	memcachedMap, dataIsMap := res.Data.(map[string]interface{})
	assert.True(t, dataIsMap)
	assert.Equal(t, []string{"localhost"}, memcachedMap["Servers"])

	cancel()
}

func TestAssistant_GetOutputPlugin(t *testing.T) {
	// Test getting an output plugin
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_output_plugin", t)

	influxdbID := ast.getPluginID("influxdb")
	assert.NotEmpty(t, influxdbID)
	req2, err2 := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, influxdbID})
	assert.NoError(t, err2)
	res2 := ast.handleRequests(&req2)
	assert.Equal(t, SUCCESS, res2.Status)
	_, dataIsMap := res2.Data.(map[string]interface{})
	assert.True(t, dataIsMap)

	cancel()
}

func TestAssistant_ValidatePluginToMap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	for inputName := range inputs.Inputs {
		req, _ := buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{inputName, "INPUT", nil, ""})
		res := ast.handleRequests(&req)
		assert.Equal(t, SUCCESS, res.Status)

		schema, resIsSchema := res.Data.(schema)
		assert.True(t, resIsSchema)
		assert.NotNil(t, schema.Types)
		assert.NotNil(t, schema.Defaults)
	}

	for outputName := range outputs.Outputs {
		req, _ := buildRequest(GET_PLUGIN_SCHEMA, pluginInfo{outputName, "OUTPUT", nil, ""})
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
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	req, err := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, "yeet"})
	assert.NoError(t, err)
	res := ast.handleRequests(&req)
	assert.Equal(t, FAILURE, res.Status)
	cancel()
}

func TestAssistant_GetNotRunningPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	req, err := buildRequest(GET_PLUGIN, pluginInfo{"cpu", "INPUT", nil, ""})
	assert.NoError(t, err)
	res := ast.handleRequests(&req)
	assert.Equal(t, FAILURE, res.Status)
	cancel()
}
func TestAssistant_UpdatePlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	headers := map[string]string{
		"test": "result",
	}
	URLs := []string{"www.google.com"}

	req2, err2 := buildRequest(START_PLUGIN, pluginInfo{"http", "INPUT", nil, ""})
	assert.NoError(t, err2)
	res2 := ast.handleRequests(&req2)
	assert.Equal(t, SUCCESS, res2.Status)

	httpID := ast.getPluginID("http")

	req, err := buildRequest(UPDATE_PLUGIN, pluginInfo{"", "INPUT", map[string]interface{}{
		"Headers": headers,
		"URLs":    URLs,
		"Timeout": map[string]interface{}{
			"Duration": 1,
		},
	}, httpID})
	assert.NoError(t, err)

	response := ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, response.Status)

	req3, err3 := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, httpID})
	assert.NoError(t, err3)
	plugin := ast.handleRequests(&req3)
	assert.Equal(t, SUCCESS, plugin.Status)
	data := plugin.Data

	dataMap := data.(map[string]interface{})

	assert.Equal(t, "www.google.com", (dataMap["URLs"]).([]string)[0])
	assert.Equal(t, 1*time.Nanosecond, (dataMap["Timeout"]).(map[string]interface{})["Duration"])
	assert.Equal(t, "result", (dataMap["Headers"]).(map[string]string)["test"])

	cancel()
}

func TestAssistant_UpdatePlugin_WithInvalidFieldName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	servers := []string{"localhost:11211", "10.0.0.1:11211", "10.0.0.2:11211"}
	unixSockets := []string{"/var/run/memcached_test.sock"}
	invalidField := []string{"invalid value"}

	memcachedID := ast.getPluginID("memcached")
	req, err := buildRequest(UPDATE_PLUGIN, pluginInfo{"", "INPUT", map[string]interface{}{
		"Servers":      servers,
		"UnixSockets":  unixSockets,
		"InvalidField": invalidField,
	}, memcachedID})
	assert.NoError(t, err)

	response := ast.handleRequests(&req)
	assert.Equal(t, FAILURE, response.Status)

	req2, err2 := buildRequest(GET_PLUGIN, pluginInfo{"", "INPUT", nil, memcachedID})
	assert.NoError(t, err2)
	plugin := ast.handleRequests(&req2)
	data := plugin.Data

	dataMap := data.(map[string]interface{})

	assert.Equal(t, "localhost", (dataMap["Servers"]).([]string)[0])
	assert.Equal(t, nil, dataMap["UnixSockets"])

	cancel()
}

func TestAssistant_UpdatePlugins_WithInvalidFieldType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	memcachedID := ast.getPluginID("memcached")
	servers := []string{"localhost:11211", "10.0.0.1:11211", "10.0.0.2:11211"}
	testString := "/var/run/memcached_test.sock"
	invalidField := []string{"invalid value"}
	req, err := buildRequest(UPDATE_PLUGIN, pluginInfo{"", "INPUT", map[string]interface{}{
		"Servers":      servers,
		"UnixSockets":  testString,
		"InvalidField": invalidField,
	}, memcachedID})
	assert.NoError(t, err)

	response := ast.handleRequests(&req)
	assert.Equal(t, FAILURE, response.Status)

	req2, err2 := buildRequest(GET_PLUGIN, pluginInfo{"", "INPUT", nil, memcachedID})
	assert.NoError(t, err2)
	plugin := ast.handleRequests(&req2)
	data := plugin.Data

	dataMap := data.(map[string]interface{})

	assert.Equal(t, "localhost", (dataMap["Servers"]).([]string)[0])
	assert.NotEqual(t, testString, dataMap["UnixSockets"])

	cancel()
}

func TestAssistant_GetAllPlugins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	getReq := request{GET_ALL_PLUGINS, "000", pluginInfo{"memcached", "INPUT", nil, ""}}
	res := ast.handleRequests(&getReq)
	assert.Equal(t, SUCCESS, res.Status)

	pList, ok := res.Data.(pluginsList)
	assert.True(t, ok)
	assert.Equal(t, len(inputs.Inputs), len(pList.Inputs))

	cancel()
}

func TestAssistant_GetAllRunningPlugins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	getReq := request{GET_RUNNING_PLUGINS, "000", pluginInfo{"", "", nil, ""}}
	res := ast.handleRequests(&getReq)
	pList, ok := res.Data.(pluginsWithIdList)

	assert.True(t, ok)
	assert.Equal(t, 1, len(pList.Inputs))
	assert.Equal(t, 0, len(pList.Outputs))
	assert.Equal(t, "memcached", pList.Inputs[0]["name"])
	assert.Equal(t, SUCCESS, res.Status)

	cancel()
}

func TestAssistant_StopSinglePlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "single_plugin", t)

	memcachedID := ast.getPluginID("memcached")
	req := request{STOP_PLUGIN, "123", pluginInfo{"", "INPUT", nil, memcachedID}}
	res := ast.handleRequests(&req)
	assert.Equal(t, SUCCESS, res.Status)

	getReq := request{GET_RUNNING_PLUGINS, "000", pluginInfo{"", "INPUT", nil, memcachedID}}
	res2 := ast.handleRequests(&getReq)

	t.Log(res2)

	time.Sleep(5)
	cancel()
}

func TestAssistant_MultiIDIntegrationTest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, ast := initAgentAndAssistant(ctx, "basic_config", t)

	// Step 1: User boots into dashboard and needs to see active plugin.
	getAllRunningReq, _ := buildRequest(GET_RUNNING_PLUGINS, pluginInfo{})
	allRunningRes := ast.handleRequests(&getAllRunningReq)
	assert.Equal(t, SUCCESS, allRunningRes.Status)
	pList, ok := allRunningRes.Data.(pluginsWithIdList)
	assert.True(t, ok)
	assert.Equal(t, 2, len(pList.Inputs))
	assert.Equal(t, 3, len(pList.Outputs))

	// Step 2: User wants to add a memcached input to their agent.
	startMemcachedReq, _ := buildRequest(START_PLUGIN, pluginInfo{"memcached", "INPUT", nil, ""})
	startMemcachedRes := ast.handleRequests(&startMemcachedReq)
	assert.Equal(t, SUCCESS, startMemcachedRes.Status)
	memcachedID, ok := startMemcachedRes.Data.(string)
	assert.True(t, ok)
	assert.Equal(t, ast.getPluginID("memcached"), memcachedID)

	// Step 3: User regrets her decision, and stops the memcached input.
	stopMemcachedReq, _ := buildRequest(STOP_PLUGIN, pluginInfo{"", "INPUT", nil, memcachedID})
	stopMemcachedRes := ast.handleRequests(&stopMemcachedReq)
	assert.Equal(t, SUCCESS, stopMemcachedRes.Status)
	assert.Equal(t, "", ast.getPluginID("memcached"))

	// Step 4: User wants to add a influxdb output to their agent.
	startIDBReq, _ := buildRequest(START_PLUGIN, pluginInfo{"influxdb", "OUTPUT", nil, ""})
	startIDBRes := ast.handleRequests(&startIDBReq)
	assert.Equal(t, SUCCESS, startIDBRes.Status)
	idbID, ok := startIDBRes.Data.(string)
	assert.True(t, ok)
	assert.Equal(t, 3, len(ast.getAllPluginsID("influxdb")))
	assert.Contains(t, ast.getAllPluginsID("influxdb"), idbID)

	// Step 5: User regrets their decision, and stops the influxdb output.
	stopIDBReq, _ := buildRequest(STOP_PLUGIN, pluginInfo{"", "OUTPUT", nil, idbID})
	stopIDBRes := ast.handleRequests(&stopIDBReq)
	assert.Equal(t, SUCCESS, stopIDBRes.Status)
	assert.Equal(t, 2, len(ast.getAllPluginsID("influxdb")))
	assert.NotContains(t, ast.getAllPluginsID("influxdb"), idbID)

	// Step 6: User wants to add a influxdb output to their agent (again).
	startIDBReq2, _ := buildRequest(START_PLUGIN, pluginInfo{"influxdb", "OUTPUT", nil, ""})
	startIDBRes2 := ast.handleRequests(&startIDBReq2)
	assert.Equal(t, SUCCESS, startIDBRes2.Status)
	idbID2, ok := startIDBRes2.Data.(string)
	assert.True(t, ok)
	assert.Equal(t, 3, len(ast.getAllPluginsID("influxdb")))
	assert.Contains(t, ast.getAllPluginsID("influxdb"), idbID2)
	assert.NotEqual(t, idbID, idbID2)

	// Step 7 Prep: Get the influxdb output with Database = "telegraf" for next test
	var idbIDWithSameSettings string

	for _, id := range ast.getAllPluginsID("influxdb") {
		req, _ := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, id})
		res := ast.handleRequests(&req)

		m, ok := res.Data.(map[string]interface{})

		if ok {
			if m["Database"] == "telegraf" {
				idbIDWithSameSettings = id
			}
		}
	}

	// Step 7: User wants to set the freshly added influxdb's settings to
	// urls = ["http://localhost:8086"] and database = "telegraf"

	desiredIDBSettings := map[string]interface{}{
		"URLs":     []string{"http://localhost:8086"},
		"Database": "telegraf",
	}

	updateIDBReq1, _ := buildRequest(UPDATE_PLUGIN, pluginInfo{"influxdb", "OUTPUT", desiredIDBSettings, idbID2})
	updateIDBRes1 := ast.handleRequests(&updateIDBReq1)
	assert.Equal(t, SUCCESS, updateIDBRes1.Status)

	getIDBReq1, _ := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, idbID2})
	getIDBRes1 := ast.handleRequests(&getIDBReq1)
	idbMap, ok := getIDBRes1.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, desiredIDBSettings["URLs"], idbMap["URLs"])
	assert.Equal(t, desiredIDBSettings["Database"], idbMap["Database"])

	// Step 8: User wants to change database to database = "udp-telegraf".
	// Ensure that the other influxdb plugin with the same settings doesn't get changed.

	desiredIDBSettings2 := map[string]interface{}{
		"Database": "udp-telegraf",
	}

	updateIDBReq2, _ := buildRequest(UPDATE_PLUGIN, pluginInfo{"influxdb", "OUTPUT", desiredIDBSettings2, idbID2})
	updateIDBRes2 := ast.handleRequests(&updateIDBReq2)
	assert.Equal(t, SUCCESS, updateIDBRes2.Status)

	getIDBReq2, _ := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, idbID2})
	getIDBRes2 := ast.handleRequests(&getIDBReq2)
	idbMap2, ok := getIDBRes2.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, desiredIDBSettings["URLs"], idbMap2["URLs"])
	assert.Equal(t, desiredIDBSettings2["Database"], idbMap2["Database"])

	getIDBReq3, _ := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, idbIDWithSameSettings})
	getIDBRes3 := ast.handleRequests(&getIDBReq3)
	idbMap3, ok := getIDBRes3.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, desiredIDBSettings["URLs"], idbMap3["URLs"])
	assert.Equal(t, desiredIDBSettings["Database"], idbMap3["Database"])

	// Step 9 Prep: Get the All Trues and All Falses CPU plugins.

	var allTrueCPUID string
	var allFalseCPUID string

	assert.NotEmpty(t, ast.getAllPluginsID("cpu"))

	for _, id := range ast.getAllPluginsID("cpu") {
		req, _ := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, id})
		res := ast.handleRequests(&req)

		m, ok := res.Data.(map[string]interface{})

		if ok {
			if m["PerCPU"] == true {
				allTrueCPUID = id
			} else {
				allFalseCPUID = id
			}
		}
	}

	assert.NotNil(t, allTrueCPUID)
	assert.NotNil(t, allFalseCPUID)

	// Step 9: User wants to update allFalse's PerCPU field to be true.

	desiredCPUSettings := map[string]interface{}{
		"PerCPU": true,
	}

	updateCPUReq1, _ := buildRequest(UPDATE_PLUGIN, pluginInfo{"cpu", "INPUT", desiredCPUSettings, allFalseCPUID})
	updateCPURes1 := ast.handleRequests(&updateCPUReq1)
	assert.Equal(t, SUCCESS, updateCPURes1.Status)

	getCPUReq1, _ := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, allFalseCPUID})
	getCPURes1 := ast.handleRequests(&getCPUReq1)
	cpuMap, ok := getCPURes1.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, cpuMap["PerCPU"])

	// Step 10: User wants to revert the changes she made to PerCPU in allFalse.
	// Ensure that allTrue doesn't get changed.

	desiredCPUSettings2 := map[string]interface{}{
		"PerCPU": false,
	}
	updateCPUReq2, _ := buildRequest(UPDATE_PLUGIN, pluginInfo{"cpu", "INPUT", desiredCPUSettings2, allFalseCPUID})
	updateCPURes2 := ast.handleRequests(&updateCPUReq2)
	assert.Equal(t, SUCCESS, updateCPURes2.Status)

	getCPUReq2, _ := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, allFalseCPUID})
	getCPURes2 := ast.handleRequests(&getCPUReq2)
	cpuMap2, ok := getCPURes2.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, false, cpuMap2["PerCPU"])

	getCPUReq3, _ := buildRequest(GET_PLUGIN, pluginInfo{"", "", nil, allTrueCPUID})
	getCPURes3 := ast.handleRequests(&getCPUReq3)
	cpuMap3, ok := getCPURes3.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, cpuMap3["PerCPU"])

	cancel()
}
