package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadSingleInputWithEnvVars(t *testing.T) {
	c := NewConfig()
	err := os.Setenv("MY_TEST_SERVER", "192.168.1.1")
	assert.NoError(t, err)
	err = os.Setenv("TEST_INTERVAL", "10s")
	assert.NoError(t, err)
	c.LoadConfig("./testdata/single_plugin_env_vars.toml")

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
	assert.NoError(t, filter.Compile())
	inputConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 10 * time.Second,
	}
	inputConfig.Tags = make(map[string]string)

	// Ignore Log and Parser
	c.Inputs[0].Input.(*MockupInputPlugin).Log = nil
	c.Inputs[0].Input.(*MockupInputPlugin).parser = nil
	assert.Equal(t, input, c.Inputs[0].Input, "Testdata did not produce a correct mockup struct.")
	assert.Equal(t, inputConfig, c.Inputs[0].Config, "Testdata did not produce correct input metadata.")
}

func TestConfig_LoadSingleInput(t *testing.T) {
	c := NewConfig()
	c.LoadConfig("./testdata/single_plugin.toml")

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
	assert.NoError(t, filter.Compile())
	inputConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 5 * time.Second,
	}
	inputConfig.Tags = make(map[string]string)

	// Ignore Log and Parser
	c.Inputs[0].Input.(*MockupInputPlugin).Log = nil
	c.Inputs[0].Input.(*MockupInputPlugin).parser = nil
	assert.Equal(t, input, c.Inputs[0].Input, "Testdata did not produce a correct memcached struct.")
	assert.Equal(t, inputConfig, c.Inputs[0].Config, "Testdata did not produce correct memcached metadata.")
}

func TestConfig_LoadDirectory(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("./testdata/single_plugin.toml")
	if err != nil {
		t.Error(err)
	}
	err = c.LoadDirectory("./testdata/subconfig")
	if err != nil {
		t.Error(err)
	}

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
	assert.NoError(t, filterMockup.Compile())
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
	assert.NoError(t, err)
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
	assert.NoError(t, filterMemcached.Compile())
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
	assert.Len(t, c.Inputs, len(expectedPlugins))
	assert.Len(t, c.Inputs, len(expectedConfigs))
	for i, plugin := range c.Inputs {
		input := plugin.Input.(*MockupInputPlugin)
		// Check the logger and ignore it for comparison
		require.NotNil(t, input.Log)
		input.Log = nil

		// Ignore the parser if not expected
		if expectedPlugins[i].parser == nil {
			input.parser = nil
		}

		assert.Equalf(t, expectedPlugins[i], plugin.Input, "Plugin %d: incorrect struct produced", i)
		assert.Equalf(t, expectedConfigs[i], plugin.Config, "Plugin %d: incorrect config produced", i)
	}

}

func TestConfig_LoadSpecialTypes(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("./testdata/special_types.toml")
	assert.NoError(t, err)
	require.Equal(t, 1, len(c.Inputs))

	input, ok := c.Inputs[0].Input.(*MockupInputPlugin)
	assert.Equal(t, true, ok)
	// Tests telegraf duration parsing.
	assert.Equal(t, Duration(time.Second), input.WriteTimeout)
	// Tests telegraf size parsing.
	assert.Equal(t, internal.Size{Size: 1024 * 1024}, input.MaxBodySize)
	// Tests toml multiline basic strings.
	assert.Equal(t, "/path/to/my/cert", strings.TrimRight(input.TLSCert, "\r\n"))
}

func TestConfig_FieldNotDefined(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("./testdata/invalid_field.toml")
	require.Error(t, err, "invalid field name")
	assert.Equal(t, "Error loading config file ./testdata/invalid_field.toml: plugin inputs.http_listener_v2: line 1: configuration specified the fields [\"not_a_field\"], but they weren't used", err.Error())
}

func TestConfig_WrongFieldType(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("./testdata/wrong_field_type.toml")
	require.Error(t, err, "invalid field type")
	assert.Equal(t, "Error loading config file ./testdata/wrong_field_type.toml: error parsing http_listener_v2, line 2: (config.MockupInputPlugin.Port) cannot unmarshal TOML string into int", err.Error())

	c = NewConfig()
	err = c.LoadConfig("./testdata/wrong_field_type2.toml")
	require.Error(t, err, "invalid field type2")
	assert.Equal(t, "Error loading config file ./testdata/wrong_field_type2.toml: error parsing http_listener_v2, line 2: (config.MockupInputPlugin.Methods) cannot unmarshal TOML string into []string", err.Error())
}

func TestConfig_InlineTables(t *testing.T) {
	// #4098
	c := NewConfig()
	err := c.LoadConfig("./testdata/inline_table.toml")
	assert.NoError(t, err)
	require.Equal(t, 2, len(c.Outputs))

	output, ok := c.Outputs[1].Output.(*MockupOuputPlugin)
	assert.Equal(t, true, ok)
	assert.Equal(t, map[string]string{"Authorization": "Token $TOKEN", "Content-Type": "application/json"}, output.Headers)
	assert.Equal(t, []string{"org_id"}, c.Outputs[0].Config.Filter.TagInclude)
}

func TestConfig_SliceComment(t *testing.T) {
	t.Skipf("Skipping until #3642 is resolved")

	c := NewConfig()
	err := c.LoadConfig("./testdata/slice_comment.toml")
	assert.NoError(t, err)
	require.Equal(t, 1, len(c.Outputs))

	output, ok := c.Outputs[0].Output.(*MockupOuputPlugin)
	assert.Equal(t, []string{"test"}, output.Scopes)
	assert.Equal(t, true, ok)
}

func TestConfig_BadOrdering(t *testing.T) {
	// #3444: when not using inline tables, care has to be taken so subsequent configuration
	// doesn't become part of the table. This is not a bug, but TOML syntax.
	c := NewConfig()
	err := c.LoadConfig("./testdata/non_slice_slice.toml")
	require.Error(t, err, "bad ordering")
	assert.Equal(t, "Error loading config file ./testdata/non_slice_slice.toml: error parsing http array, line 4: cannot unmarshal TOML array into string (need slice)", err.Error())
}

func TestConfig_AzureMonitorNamespacePrefix(t *testing.T) {
	// #8256 Cannot use empty string as the namespace prefix
	c := NewConfig()
	err := c.LoadConfig("./testdata/azure_monitor.toml")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(c.Outputs))

	expectedPrefix := []string{"Telegraf/", ""}
	for i, plugin := range c.Outputs {
		output, ok := plugin.Output.(*MockupOuputPlugin)
		assert.True(t, ok)
		assert.Equal(t, expectedPrefix[i], output.NamespacePrefix)
	}
}

func TestConfig_URLRetries3Fails(t *testing.T) {
	httpLoadConfigRetryInterval = 0 * time.Second
	responseCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		responseCounter++
	}))
	defer ts.Close()

	c := NewConfig()
	err := c.LoadConfig(ts.URL)
	require.Error(t, err)
	require.Equal(t, 4, responseCounter)
}

func TestConfig_URLRetries3FailsThenPasses(t *testing.T) {
	httpLoadConfigRetryInterval = 0 * time.Second
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

	c := NewConfig()
	err := c.LoadConfig(ts.URL)
	require.NoError(t, err)
	require.Equal(t, 4, responseCounter)
}

/*** Mockup INPUT plugin for testing to avoid cyclic dependencies ***/
type MockupInputPlugin struct {
	Servers      []string      `toml:"servers"`
	Methods      []string      `toml:"methods"`
	Timeout      Duration      `toml:"timeout"`
	ReadTimeout  Duration      `toml:"read_timeout"`
	WriteTimeout Duration      `toml:"write_timeout"`
	MaxBodySize  internal.Size `toml:"max_body_size"`
	Port         int           `toml:"port"`
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
	URL                 string            `toml:"url"`
	Headers             map[string]string `toml:"headers"`
	Scopes              []string          `toml:"scopes"`
	NamespacePrefix     string            `toml:"namespace_prefix"`
	Log                 telegraf.Logger   `toml:"-"`
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
	inputs.Add("exec", func() telegraf.Input { return &MockupInputPlugin{Timeout: Duration(time.Second * 5)} })
	inputs.Add("http_listener_v2", func() telegraf.Input { return &MockupInputPlugin{} })
	inputs.Add("memcached", func() telegraf.Input { return &MockupInputPlugin{} })
	inputs.Add("procstat", func() telegraf.Input { return &MockupInputPlugin{} })

	// Register the mockup output plugin for the required names
	outputs.Add("azure_monitor", func() telegraf.Output { return &MockupOuputPlugin{NamespacePrefix: "Telegraf/"} })
	outputs.Add("http", func() telegraf.Output { return &MockupOuputPlugin{} })
}
