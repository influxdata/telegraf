package config_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	_ "github.com/influxdata/telegraf/plugins/aggregators/minmax"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/outputs"
	_ "github.com/influxdata/telegraf/plugins/outputs/file"
	"github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/processors/rename"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadSingleInputWithEnvVars(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	require.NoError(t, os.Setenv("MY_TEST_SERVER", "192.168.1.1"))
	require.NoError(t, os.Setenv("TEST_INTERVAL", "10s"))
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/single_plugin_env_vars.toml")
	require.NoError(t, err)

	input := inputs.Inputs["memcached"]().(*MockupInputPlugin)
	input.Servers = []string{"192.168.1.1"}

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
	inputConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 10 * time.Second,
	}
	inputConfig.Tags = make(map[string]string)

	require.Len(t, c.Inputs(), 1)

	// Ignore Log and Parser
	c.Inputs()[0].Input.(*MockupInputPlugin).Log = nil
	c.Inputs()[0].Input.(*MockupInputPlugin).parser = nil
	require.Equal(t, input, c.Inputs()[0].Input, "Testdata did not produce a correct mockup struct.")
	require.Equal(t, inputConfig, c.Inputs()[0].Config, "Testdata did not produce correct input metadata.")
}

func TestConfig_LoadSingleInput(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/single_plugin.toml")
	require.NoError(t, err)

	input := inputs.Inputs["memcached"]().(*MockupInputPlugin)
	input.Servers = []string{"localhost"}

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
	inputConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 5 * time.Second,
	}
	inputConfig.Tags = make(map[string]string)

	require.Len(t, c.Inputs(), 1)

	// Ignore Log and Parser
	c.Inputs()[0].Input.(*MockupInputPlugin).Log = nil
	c.Inputs()[0].Input.(*MockupInputPlugin).parser = nil
	require.Equal(t, input, c.Inputs()[0].Input, "Testdata did not produce a correct memcached struct.")
	require.Equal(t, inputConfig, c.Inputs()[0].Config, "Testdata did not produce correct memcached metadata.")
}

func TestConfig_LoadDirectory(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	require.NoError(t, c.LoadConfig(context.Background(), context.Background(), "./testdata/single_plugin.toml"))
	require.NoError(t, c.LoadDirectory(context.Background(), context.Background(), "./testdata/subconfig"))

	// Create the expected data
	expectedPlugins := make([]*MockupInputPlugin, 4)
	expectedConfigs := make([]*models.InputConfig, 4)

	expectedPlugins[0] = inputs.Inputs["memcached"]().(*MockupInputPlugin)
	expectedPlugins[0].Servers = []string{"localhost"}

	filterMockup := models.Filter{
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
	require.NoError(t, filterMockup.Compile())
	expectedConfigs[0] = &models.InputConfig{
		Name:     "memcached",
		Filter:   filterMockup,
		Interval: 5 * time.Second,
	}
	expectedConfigs[0].Tags = make(map[string]string)

	expectedPlugins[1] = inputs.Inputs["exec"]().(*MockupInputPlugin)
	p, err := parsers.NewParser(&parsers.Config{
		MetricName: "exec",
		DataFormat: "json",
		JSONStrict: true,
	})
	require.NoError(t, err)
	expectedPlugins[1].SetParser(p)
	expectedPlugins[1].Command = "/usr/bin/myothercollector --foo=bar"
	expectedConfigs[1] = &models.InputConfig{
		Name:              "exec",
		MeasurementSuffix: "_myothercollector",
	}
	expectedConfigs[1].Tags = make(map[string]string)

	expectedPlugins[2] = inputs.Inputs["memcached"]().(*MockupInputPlugin)
	expectedPlugins[2].Servers = []string{"192.168.1.1"}

	filterMemcached := models.Filter{
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
	require.NoError(t, filterMemcached.Compile())
	expectedConfigs[2] = &models.InputConfig{
		Name:     "memcached",
		Filter:   filterMemcached,
		Interval: 5 * time.Second,
	}
	expectedConfigs[2].Tags = make(map[string]string)

	expectedPlugins[3] = inputs.Inputs["procstat"]().(*MockupInputPlugin)
	expectedPlugins[3].PidFile = "/var/run/grafana-server.pid"
	expectedConfigs[3] = &models.InputConfig{Name: "procstat"}
	expectedConfigs[3].Tags = make(map[string]string)

	// Check the generated plugins
	require.Len(t, c.Inputs(), len(expectedPlugins))
	require.Len(t, c.Inputs(), len(expectedConfigs))
	for i, plugin := range c.Inputs() {
		input := plugin.Input.(*MockupInputPlugin)
		// Check the logger and ignore it for comparison
		require.NotNil(t, input.Log)
		input.Log = nil

		// Ignore the parser if not expected
		if expectedPlugins[i].parser == nil {
			input.parser = nil
		}

		require.Equalf(t, expectedPlugins[i], plugin.Input, "Plugin %d: incorrect struct produced", i)
		require.Equalf(t, expectedConfigs[i], plugin.Config, "Plugin %d: incorrect config produced", i)
	}
}

func TestConfig_LoadSpecialTypes(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "./testdata/special_types.toml")
	require.NoError(t, err)
	require.Equal(t, 1, len(c.Inputs()))

	input, ok := c.Inputs()[0].Input.(*MockupInputPlugin)
	require.True(t, ok)
	// Tests telegraf duration parsing.
	require.Equal(t, config.Duration(time.Second), input.WriteTimeout) // Tests telegraf size parsing.
	require.Equal(t, config.Size(1024*1024), input.MaxBodySize)
	// Tests toml multiline basic strings.
	require.Equal(t, "/path/to/my/cert", strings.TrimRight(input.TLSCert, "\r\n"))
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
	require.Equal(t, "Error loading config file ./testdata/wrong_field_type.toml: error parsing http_listener_v2, line 2: (config.MockupInputPlugin.Port) cannot unmarshal TOML string into int", err.Error())

	c = config.NewConfig()
	c.SetAgent(agentController)
	err = c.LoadConfig(context.Background(), context.Background(), "./testdata/wrong_field_type2.toml")
	require.Error(t, err, "invalid field type2")
	require.Equal(t, "Error loading config file ./testdata/wrong_field_type2.toml: error parsing http_listener_v2, line 2: (config.MockupInputPlugin.Methods) cannot unmarshal TOML string into []string", err.Error())
}

func TestConfig_InlineTables(t *testing.T) {
	// #4098
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	require.NoError(t, c.LoadConfig(context.Background(), context.Background(), "./testdata/inline_table.toml"))
	require.Len(t, c.Outputs, 2)

	output, ok := c.Outputs()[1].Output.(*MockupOuputPlugin)
	require.True(t, ok)
	require.Equal(t, map[string]string{"Authorization": "Token $TOKEN", "Content-Type": "application/json"}, output.Headers)
	require.Equal(t, []string{"org_id"}, c.Outputs()[0].Config.Filter.TagInclude)
}

func TestConfig_SliceComment(t *testing.T) {
	t.Skipf("Skipping until #3642 is resolved")

	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	require.NoError(t, c.LoadConfig(context.Background(), context.Background(), "./testdata/slice_comment.toml"))
	require.Len(t, c.Outputs, 1)

	output, ok := c.Outputs()[0].Output.(*MockupOuputPlugin)
	require.True(t, ok)
	require.Equal(t, []string{"test"}, output.Scopes)
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
	require.NoError(t, c.LoadConfig(context.Background(), context.Background(), "./testdata/azure_monitor.toml"))
	require.Len(t, c.Outputs, 2)

	expectedPrefix := []string{"Telegraf/", ""}
	for i, plugin := range c.Outputs() {
		output, ok := plugin.Output.(*MockupOuputPlugin)
		require.True(t, ok)
		require.Equal(t, expectedPrefix[i], output.NamespacePrefix)
	}
}

func TestConfig_URLRetries3Fails(t *testing.T) {
	config.HttpLoadConfigRetryInterval = 0 * time.Second
	responseCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		responseCounter++
	}))
	defer ts.Close()

	expected := fmt.Sprintf("Error loading config file %s: Retry 3 of 3 failed to retrieve remote config: 404 Not Found", ts.URL)

	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), ts.URL)
	require.Error(t, err)
	require.Equal(t, expected, err.Error())
	require.Equal(t, 4, responseCounter)
}

func TestConfig_URLRetries3FailsThenPasses(t *testing.T) {
	config.HttpLoadConfigRetryInterval = 0 * time.Second
	responseCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if responseCounter <= 2 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		responseCounter++
	}))
	defer ts.Close()

	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	require.NoError(t, c.LoadConfig(context.Background(), context.Background(), ts.URL))
	require.Equal(t, 4, responseCounter)
}

func TestConfig_getDefaultConfigPathFromEnvURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := os.Setenv("TELEGRAF_CONFIG_PATH", ts.URL)
	require.NoError(t, err)
	configPath, err := config.GetDefaultConfigPath()
	require.NoError(t, err)
	require.Equal(t, ts.URL, configPath)
	err = c.LoadConfig(context.Background(), context.Background(), "")
	require.NoError(t, err)
}

func TestConfig_URLLikeFileName(t *testing.T) {
	c := config.NewConfig()
	agentController := &testAgentController{}
	c.SetAgent(agentController)
	defer agentController.reset()
	err := c.LoadConfig(context.Background(), context.Background(), "http:##www.example.com.conf")
	require.Error(t, err)

	if runtime.GOOS == "windows" {
		// The error file not found error message is different on windows
		require.Equal(t, "Error loading config file http:##www.example.com.conf: open http:##www.example.com.conf: The system cannot find the file specified.", err.Error())
	} else {
		require.Equal(t, "Error loading config file http:##www.example.com.conf: open http:##www.example.com.conf: no such file or directory", err.Error())
	}
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
	// negative orders are defaults based on file position order  -10,000,000
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

/*** Mockup INPUT plugin for testing to avoid cyclic dependencies ***/
type MockupInputPlugin struct {
	Servers      []string        `toml:"servers"`
	Methods      []string        `toml:"methods"`
	Timeout      config.Duration `toml:"timeout"`
	ReadTimeout  config.Duration `toml:"read_timeout"`
	WriteTimeout config.Duration `toml:"write_timeout"`
	MaxBodySize  config.Size     `toml:"max_body_size"`
	Port         int             `toml:"port"`
	Command      string
	PidFile      string
	Log          telegraf.Logger `toml:"-"`
	tls.ServerConfig

	parser parsers.Parser
}

func (m *MockupInputPlugin) SampleConfig() string                  { return "Mockup test intput plugin" }
func (m *MockupInputPlugin) Description() string                   { return "Mockup test intput plugin" }
func (m *MockupInputPlugin) Gather(acc telegraf.Accumulator) error { return nil }
func (m *MockupInputPlugin) SetParser(parser parsers.Parser)       { m.parser = parser }

/*** Mockup OUTPUT plugin for testing to avoid cyclic dependencies ***/
type MockupOuputPlugin struct {
	URL             string            `toml:"url"`
	Headers         map[string]string `toml:"headers"`
	Scopes          []string          `toml:"scopes"`
	NamespacePrefix string            `toml:"namespace_prefix"`
	Log             telegraf.Logger   `toml:"-"`
	tls.ClientConfig
}

func (m *MockupOuputPlugin) Connect() error                        { return nil }
func (m *MockupOuputPlugin) Close() error                          { return nil }
func (m *MockupOuputPlugin) Description() string                   { return "Mockup test output plugin" }
func (m *MockupOuputPlugin) SampleConfig() string                  { return "Mockup test output plugin" }
func (m *MockupOuputPlugin) Write(metrics []telegraf.Metric) error { return nil }

// Register the mockup plugin on loading
func init() {
	// Register the mockup input plugin for the required names
	inputs.Add("exec", func() telegraf.Input { return &MockupInputPlugin{Timeout: config.Duration(time.Second * 5)} })
	inputs.Add("http_listener_v2", func() telegraf.Input { return &MockupInputPlugin{} })
	inputs.Add("memcached", func() telegraf.Input { return &MockupInputPlugin{} })
	inputs.Add("procstat", func() telegraf.Input { return &MockupInputPlugin{} })

	// Register the mockup output plugin for the required names
	outputs.Add("azure_monitor", func() telegraf.Output { return &MockupOuputPlugin{NamespacePrefix: "Telegraf/"} })
	outputs.Add("http", func() telegraf.Output { return &MockupOuputPlugin{} })
}
