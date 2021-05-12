package config_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	_ "github.com/influxdata/telegraf/plugins/aggregators/minmax"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/exec"
	_ "github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/inputs/http_listener_v2"
	"github.com/influxdata/telegraf/plugins/inputs/memcached"
	"github.com/influxdata/telegraf/plugins/inputs/procstat"
	"github.com/influxdata/telegraf/plugins/outputs/azure_monitor"
	_ "github.com/influxdata/telegraf/plugins/outputs/file"
	httpOut "github.com/influxdata/telegraf/plugins/outputs/http"
	"github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/processors/rename"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadSingleInputWithEnvVars(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := os.Setenv("MY_TEST_SERVER", "192.168.1.1")
	require.NoError(t, err)
	err = os.Setenv("TEST_INTERVAL", "10s")
	require.NoError(t, err)
	err = c.LoadConfig(context.Background(), context.Background(), "./testdata/single_plugin_env_vars.toml")
	require.NoError(t, err)

	memcached := inputs.Inputs["memcached"]().(*memcached.Memcached)
	memcached.Servers = []string{"192.168.1.1"}

	filter := models.Filter{
		NameDrop:  []string{"metricname2"},
		NamePass:  []string{"metricname1", "ip_192.168.1.1_name"},
		FieldDrop: []string{"other", "stuff"},
		FieldPass: []string{"some", "strings"},
		TagDrop: []models.TagFilter{
			{
				Name:   "badtag",
				Filter: []string{"othertag"},
			},
		},
		TagPass: []models.TagFilter{
			{
				Name:   "goodtag",
				Filter: []string{"mytag"},
			},
		},
	}
	require.NoError(t, filter.Compile())
	mConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 10 * time.Second,
	}
	mConfig.Tags = make(map[string]string)

	require.Len(t, c.Inputs(), 1)

	require.Equal(t, memcached, c.Inputs()[0].Input,
		"Testdata did not produce a correct memcached struct.")
	require.Equal(t, mConfig, c.Inputs()[0].Config,
		"Testdata did not produce correct memcached metadata.")
}

func TestConfig_LoadSingleInput(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/single_plugin.toml")
	require.NoError(t, err)

	memcached := inputs.Inputs["memcached"]().(*memcached.Memcached)
	memcached.Servers = []string{"localhost"}

	filter := models.Filter{
		NameDrop:  []string{"metricname2"},
		NamePass:  []string{"metricname1"},
		FieldDrop: []string{"other", "stuff"},
		FieldPass: []string{"some", "strings"},
		TagDrop: []models.TagFilter{
			{
				Name:   "badtag",
				Filter: []string{"othertag"},
			},
		},
		TagPass: []models.TagFilter{
			{
				Name:   "goodtag",
				Filter: []string{"mytag"},
			},
		},
	}
	require.NoError(t, filter.Compile())
	mConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 5 * time.Second,
	}
	mConfig.Tags = make(map[string]string)

	require.Len(t, c.Inputs(), 1)
	require.Equal(t, memcached, c.Inputs()[0].Input,
		"Testdata did not produce a correct memcached struct.")
	require.Equal(t, mConfig, c.Inputs()[0].Config,
		"Testdata did not produce correct memcached metadata.")
}

func TestConfig_LoadDirectory(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/single_plugin.toml")
	if err != nil {
		t.Error(err)
	}
	err = c.LoadDirectory(context.Background(), context.Background(), "./testdata/subconfig")
	if err != nil {
		t.Error(err)
	}

	memcached := inputs.Inputs["memcached"]().(*memcached.Memcached)
	memcached.Servers = []string{"localhost"}

	filter := models.Filter{
		NameDrop:  []string{"metricname2"},
		NamePass:  []string{"metricname1"},
		FieldDrop: []string{"other", "stuff"},
		FieldPass: []string{"some", "strings"},
		TagDrop: []models.TagFilter{
			{
				Name:   "badtag",
				Filter: []string{"othertag"},
			},
		},
		TagPass: []models.TagFilter{
			{
				Name:   "goodtag",
				Filter: []string{"mytag"},
			},
		},
	}
	require.NoError(t, filter.Compile())
	mConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 5 * time.Second,
	}
	mConfig.Tags = make(map[string]string)

	require.Len(t, c.Inputs(), 4)
	require.Equal(t, memcached, c.Inputs()[0].Input,
		"Testdata did not produce a correct memcached struct.")
	require.Equal(t, mConfig, c.Inputs()[0].Config,
		"Testdata did not produce correct memcached metadata.")

	ex := inputs.Inputs["exec"]().(*exec.Exec)
	p, err := parsers.NewParser(&parsers.Config{
		MetricName: "exec",
		DataFormat: "json",
		JSONStrict: true,
	})
	require.NoError(t, err)
	ex.SetParser(p)
	ex.Command = "/usr/bin/myothercollector --foo=bar"
	eConfig := &models.InputConfig{
		Name:              "exec",
		MeasurementSuffix: "_myothercollector",
	}
	eConfig.Tags = make(map[string]string)

	exec := c.Inputs()[1].Input.(*exec.Exec)
	require.NotNil(t, exec.Log)
	exec.Log = nil

	require.Len(t, c.Inputs(), 4)
	require.Equal(t, ex, c.Inputs()[1].Input,
		"Merged Testdata did not produce a correct exec struct.")
	require.Equal(t, eConfig, c.Inputs()[1].Config,
		"Merged Testdata did not produce correct exec metadata.")

	memcached.Servers = []string{"192.168.1.1"}
	require.Equal(t, memcached, c.Inputs()[2].Input,
		"Testdata did not produce a correct memcached struct.")
	require.Equal(t, mConfig, c.Inputs()[2].Config,
		"Testdata did not produce correct memcached metadata.")

	pstat := inputs.Inputs["procstat"]().(*procstat.Procstat)
	pstat.PidFile = "/var/run/grafana-server.pid"

	pConfig := &models.InputConfig{Name: "procstat"}
	pConfig.Tags = make(map[string]string)

	require.Equal(t, pstat, c.Inputs()[3].Input,
		"Merged Testdata did not produce a correct procstat struct.")
	require.Equal(t, pConfig, c.Inputs()[3].Config,
		"Merged Testdata did not produce correct procstat metadata.")
}

func TestConfig_LoadSpecialTypes(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/special_types.toml")
	require.NoError(t, err)
	require.Equal(t, 1, len(c.Inputs()))

	inputHTTPListener, ok := c.Inputs()[0].Input.(*http_listener_v2.HTTPListenerV2)
	require.Equal(t, true, ok)
	// Tests telegraf duration parsing.
	require.Equal(t, internal.Duration{Duration: time.Second}, inputHTTPListener.WriteTimeout)
	// Tests telegraf size parsing.
	require.Equal(t, internal.Size{Size: 1024 * 1024}, inputHTTPListener.MaxBodySize)
	// Tests toml multiline basic strings.
	require.Equal(t, "/path/to/my/cert", strings.TrimRight(inputHTTPListener.TLSCert, "\r\n"))
}

func TestConfig_FieldNotDefined(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/invalid_field.toml")
	require.Error(t, err, "invalid field name")
	require.Equal(t, "Error loading config file ./testdata/invalid_field.toml: plugin inputs.http_listener_v2: line 1: configuration specified the fields [\"not_a_field\"], but they weren't used", err.Error())
}

func TestConfig_WrongFieldType(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/wrong_field_type.toml")
	require.Error(t, err, "invalid field type")
	require.Equal(t, "Error loading config file ./testdata/wrong_field_type.toml: error parsing http_listener_v2, line 2: (http_listener_v2.HTTPListenerV2.Port) cannot unmarshal TOML string into int", err.Error())

	c = config.NewConfig()
	c.SetAgent(agentController)
	err = c.LoadConfig(context.Background(), context.Background(), "./testdata/wrong_field_type2.toml")
	require.Error(t, err, "invalid field type2")
	require.Equal(t, "Error loading config file ./testdata/wrong_field_type2.toml: error parsing http_listener_v2, line 2: (http_listener_v2.HTTPListenerV2.Methods) cannot unmarshal TOML string into []string", err.Error())
}

func TestConfig_InlineTables(t *testing.T) {
	// #4098
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/inline_table.toml")
	require.NoError(t, err)
	require.Equal(t, 2, len(c.Outputs()))

	outputHTTP, ok := c.Outputs()[1].Output.(*httpOut.HTTP)
	require.Equal(t, true, ok)
	require.Equal(t, map[string]string{"Authorization": "Token $TOKEN", "Content-Type": "application/json"}, outputHTTP.Headers)
	require.Equal(t, []string{"org_id"}, c.Outputs()[0].Config.Filter.TagInclude)
}

func TestConfig_SliceComment(t *testing.T) {
	t.Skipf("Skipping until #3642 is resolved")

	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/slice_comment.toml")
	require.NoError(t, err)
	require.Equal(t, 1, len(c.Outputs()))

	outputHTTP, ok := c.Outputs()[0].Output.(*httpOut.HTTP)
	require.Equal(t, []string{"test"}, outputHTTP.Scopes)
	require.Equal(t, true, ok)
}

func TestConfig_BadOrdering(t *testing.T) {
	// #3444: when not using inline tables, care has to be taken so subsequent configuration
	// doesn't become part of the table. This is not a bug, but TOML syntax.
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/non_slice_slice.toml")
	require.Error(t, err, "bad ordering")
	require.Equal(t, "Error loading config file ./testdata/non_slice_slice.toml: error parsing http array, line 4: cannot unmarshal TOML array into string (need slice)", err.Error())
}

func TestConfig_AzureMonitorNamespacePrefix(t *testing.T) {
	// #8256 Cannot use empty string as the namespace prefix
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	defaultPrefixConfig := `[[outputs.azure_monitor]]`
	err := c.LoadConfigData(context.Background(), context.Background(), []byte(defaultPrefixConfig))
	require.NoError(t, err)
	require.Len(t, c.Outputs(), 1)
	azureMonitor, ok := c.Outputs()[0].Output.(*azure_monitor.AzureMonitor)
	require.Equal(t, "Telegraf/", azureMonitor.NamespacePrefix)
	require.Equal(t, true, ok)

	agentController.reset()
	c = config.NewConfig()
	c.SetAgent(agentController)
	customPrefixConfig := `[[outputs.azure_monitor]]
	namespace_prefix = ""`
	err = c.LoadConfigData(context.Background(), context.Background(), []byte(customPrefixConfig))
	require.NoError(t, err)
	azureMonitor, ok = c.Outputs()[0].Output.(*azure_monitor.AzureMonitor)
	require.Equal(t, "", azureMonitor.NamespacePrefix)
	require.Equal(t, true, ok)
}

func TestConfig_OrderingProcessorsWithAggregators(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/processor_and_aggregator_order.toml")
	require.NoError(t, err)
	require.Equal(t, 1, len(c.Inputs()))
	require.Equal(t, 4, len(c.Processors()))
	require.Equal(t, 1, len(c.Outputs()))

	actual := map[string]int64{}
	expected := map[string]int64{
		"aggregators.minmax::one":   1,
		"processors.rename::two":    2,
		"aggregators.minmax::three": 3,
		"processors.rename::four":   4,
	}
	for _, p := range c.Processors() {
		actual[p.LogName()] = p.Order()
	}
	require.EqualValues(t, expected, actual)
}

func TestConfig_DefaultOrderingProcessorsWithAggregators(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/processor_and_aggregator_unordered.toml")
	require.NoError(t, err)
	require.Equal(t, 1, len(c.Inputs()))
	require.Equal(t, 4, len(c.Processors()))
	require.Equal(t, 1, len(c.Outputs()))

	actual := map[string]int64{}
	// negative orders are defaults based on file position order + -10,000,000
	expected := map[string]int64{
		"aggregators.minmax::one":   -9999984,
		"processors.rename::two":    -9999945,
		"aggregators.minmax::three": -9999907,
		"processors.rename::four":   4,
	}
	for _, p := range c.Processors() {
		actual[p.LogName()] = p.Order()
	}
	require.EqualValues(t, expected, actual)
}

type testAgentController struct {
	inputs     []*models.RunningInput
	processors []models.ProcessorRunner
	outputs    []*models.RunningOutput
	// configs    []*config.RunningConfigPlugin
}

func (a *testAgentController) reset() {
	a.inputs = nil
	a.processors = nil
	a.outputs = nil
	// a.configs = nil
}

func (a *testAgentController) RunningInputs() []*models.RunningInput {
	return a.inputs
}
func (a *testAgentController) RunningProcessors() []models.ProcessorRunner {
	return a.processors
}
func (a *testAgentController) RunningOutputs() []*models.RunningOutput {
	return a.outputs
}
func (a *testAgentController) AddInput(input *models.RunningInput) {
	a.inputs = append(a.inputs, input)
}
func (a *testAgentController) AddProcessor(processor models.ProcessorRunner) {
	a.processors = append(a.processors, processor)
}
func (a *testAgentController) AddOutput(output *models.RunningOutput) {
	a.outputs = append(a.outputs, output)
}
func (a *testAgentController) RunInput(input *models.RunningInput, startTime time.Time)        {}
func (a *testAgentController) RunProcessor(p models.ProcessorRunner)                           {}
func (a *testAgentController) RunOutput(ctx context.Context, output *models.RunningOutput)     {}
func (a *testAgentController) RunConfigPlugin(ctx context.Context, plugin config.ConfigPlugin) {}
func (a *testAgentController) StopInput(i *models.RunningInput)                                {}
func (a *testAgentController) StopProcessor(p models.ProcessorRunner)                          {}
func (a *testAgentController) StopOutput(p *models.RunningOutput)                              {}
