package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/parsers/all" // Blank import to have all parsers for testing
)

func TestConfig_LoadSingleInputWithEnvVars(t *testing.T) {
	c := NewConfig()
	require.NoError(t, os.Setenv("MY_TEST_SERVER", "192.168.1.1"))
	require.NoError(t, os.Setenv("TEST_INTERVAL", "10s"))
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
	require.NoError(t, filter.Compile())
	inputConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 10 * time.Second,
	}
	inputConfig.Tags = make(map[string]string)

	// Ignore Log and Parser
	c.Inputs[0].Input.(*MockupInputPlugin).Log = nil
	c.Inputs[0].Input.(*MockupInputPlugin).parser = nil
	require.Equal(t, input, c.Inputs[0].Input, "Testdata did not produce a correct mockup struct.")
	require.Equal(t, inputConfig, c.Inputs[0].Config, "Testdata did not produce correct input metadata.")
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
	require.NoError(t, filter.Compile())
	inputConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 5 * time.Second,
	}
	inputConfig.Tags = make(map[string]string)

	// Ignore Log and Parser
	c.Inputs[0].Input.(*MockupInputPlugin).Log = nil
	c.Inputs[0].Input.(*MockupInputPlugin).parser = nil
	require.Equal(t, input, c.Inputs[0].Input, "Testdata did not produce a correct memcached struct.")
	require.Equal(t, inputConfig, c.Inputs[0].Config, "Testdata did not produce correct memcached metadata.")
}

func TestConfig_LoadDirectory(t *testing.T) {
	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/single_plugin.toml"))
	require.NoError(t, c.LoadDirectory("./testdata/subconfig"))

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
	parserConfig := &parsers.Config{
		MetricName: "exec",
		DataFormat: "json",
		JSONStrict: true,
	}
	p, err := parsers.NewParser(parserConfig)
	require.NoError(t, err)

	// Inject logger to have proper struct for comparison
	models.SetLoggerOnPlugin(p, models.NewLogger("parsers", parserConfig.DataFormat, parserConfig.MetricName))

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
	require.Len(t, c.Inputs, len(expectedPlugins))
	require.Len(t, c.Inputs, len(expectedConfigs))
	for i, plugin := range c.Inputs {
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

func TestConfig_WrongCertPath(t *testing.T) {
	c := NewConfig()
	require.Error(t, c.LoadConfig("./testdata/wrong_cert_path.toml"))
}

func TestConfig_LoadSpecialTypes(t *testing.T) {
	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/special_types.toml"))
	require.Len(t, c.Inputs, 1)

	input, ok := c.Inputs[0].Input.(*MockupInputPlugin)
	require.True(t, ok)
	// Tests telegraf duration parsing.
	require.Equal(t, Duration(time.Second), input.WriteTimeout)
	// Tests telegraf size parsing.
	require.Equal(t, Size(1024*1024), input.MaxBodySize)
	// Tests toml multiline basic strings on single line.
	require.Equal(t, "./testdata/special_types.pem", input.TLSCert)
	// Tests toml multiline basic strings on single line.
	require.Equal(t, "./testdata/special_types.key", input.TLSKey)
	// Tests toml multiline basic strings on multiple lines.
	require.Equal(t, "/path/", strings.TrimRight(input.Paths[0], "\r\n"))
}

func TestConfig_FieldNotDefined(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("./testdata/invalid_field.toml")
	require.Error(t, err, "invalid field name")
	require.Equal(t, "Error loading config file ./testdata/invalid_field.toml: plugin inputs.http_listener_v2: line 1: configuration specified the fields [\"not_a_field\"], but they weren't used", err.Error())
}

func TestConfig_WrongFieldType(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("./testdata/wrong_field_type.toml")
	require.Error(t, err, "invalid field type")
	require.Equal(t, "Error loading config file ./testdata/wrong_field_type.toml: error parsing http_listener_v2, line 2: (config.MockupInputPlugin.Port) cannot unmarshal TOML string into int", err.Error())

	c = NewConfig()
	err = c.LoadConfig("./testdata/wrong_field_type2.toml")
	require.Error(t, err, "invalid field type2")
	require.Equal(t, "Error loading config file ./testdata/wrong_field_type2.toml: error parsing http_listener_v2, line 2: (config.MockupInputPlugin.Methods) cannot unmarshal TOML string into []string", err.Error())
}

func TestConfig_InlineTables(t *testing.T) {
	// #4098
	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/inline_table.toml"))
	require.Len(t, c.Outputs, 2)

	output, ok := c.Outputs[1].Output.(*MockupOuputPlugin)
	require.True(t, ok)
	require.Equal(t, map[string]string{"Authorization": "Token $TOKEN", "Content-Type": "application/json"}, output.Headers)
	require.Equal(t, []string{"org_id"}, c.Outputs[0].Config.Filter.TagInclude)
}

func TestConfig_SliceComment(t *testing.T) {
	t.Skipf("Skipping until #3642 is resolved")

	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/slice_comment.toml"))
	require.Len(t, c.Outputs, 1)

	output, ok := c.Outputs[0].Output.(*MockupOuputPlugin)
	require.True(t, ok)
	require.Equal(t, []string{"test"}, output.Scopes)
}

func TestConfig_BadOrdering(t *testing.T) {
	// #3444: when not using inline tables, care has to be taken so subsequent configuration
	// doesn't become part of the table. This is not a bug, but TOML syntax.
	c := NewConfig()
	err := c.LoadConfig("./testdata/non_slice_slice.toml")
	require.Error(t, err, "bad ordering")
	require.Equal(t, "Error loading config file ./testdata/non_slice_slice.toml: error parsing http array, line 4: cannot unmarshal TOML array into string (need slice)", err.Error())
}

func TestConfig_AzureMonitorNamespacePrefix(t *testing.T) {
	// #8256 Cannot use empty string as the namespace prefix
	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/azure_monitor.toml"))
	require.Len(t, c.Outputs, 2)

	expectedPrefix := []string{"Telegraf/", ""}
	for i, plugin := range c.Outputs {
		output, ok := plugin.Output.(*MockupOuputPlugin)
		require.True(t, ok)
		require.Equal(t, expectedPrefix[i], output.NamespacePrefix)
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

	expected := fmt.Sprintf("Error loading config file %s: Retry 3 of 3 failed to retrieve remote config: 404 Not Found", ts.URL)

	c := NewConfig()
	err := c.LoadConfig(ts.URL)
	require.Error(t, err)
	require.Equal(t, expected, err.Error())
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
	require.NoError(t, c.LoadConfig(ts.URL))
	require.Equal(t, 4, responseCounter)
}

func TestConfig_getDefaultConfigPathFromEnvURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewConfig()
	err := os.Setenv("TELEGRAF_CONFIG_PATH", ts.URL)
	require.NoError(t, err)
	configPath, err := getDefaultConfigPath()
	require.NoError(t, err)
	require.Equal(t, ts.URL, configPath)
	err = c.LoadConfig("")
	require.NoError(t, err)
}

func TestConfig_URLLikeFileName(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("http:##www.example.com.conf")
	require.Error(t, err)

	if runtime.GOOS == "windows" {
		// The error file not found error message is different on windows
		require.Equal(t, "Error loading config file http:##www.example.com.conf: open http:##www.example.com.conf: The system cannot find the file specified.", err.Error())
	} else {
		require.Equal(t, "Error loading config file http:##www.example.com.conf: open http:##www.example.com.conf: no such file or directory", err.Error())
	}
}

func TestConfig_ParserInterfaceNewFormat(t *testing.T) {
	formats := []string{
		"collectd",
		"csv",
		"dropwizard",
		"form_urlencoded",
		"graphite",
		"grok",
		"influx",
		"json",
		"json_v2",
		"logfmt",
		"nagios",
		"prometheus",
		"prometheusremotewrite",
		"value",
		"wavefront",
		"xml", "xpath_json", "xpath_msgpack", "xpath_protobuf",
	}

	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/parsers_new.toml"))
	require.Len(t, c.Inputs, len(formats))

	cfg := parsers.Config{
		CSVHeaderRowCount:     42,
		DropwizardTagPathsMap: make(map[string]string),
		GrokPatterns:          []string{"%{COMBINED_LOG_FORMAT}"},
		JSONStrict:            true,
		MetricName:            "parser_test_new",
	}

	override := map[string]struct {
		cfg   *parsers.Config
		param map[string]interface{}
		mask  []string
	}{
		"csv": {
			param: map[string]interface{}{
				"HeaderRowCount": cfg.CSVHeaderRowCount,
			},
			mask: []string{"TimeFunc"},
		},
		"logfmt": {
			mask: []string{"Now"},
		},
		"xpath_protobuf": {
			cfg: &parsers.Config{
				MetricName:        "parser_test_new",
				XPathProtobufFile: "testdata/addressbook.proto",
				XPathProtobufType: "addressbook.AddressBook",
			},
			param: map[string]interface{}{
				"ProtobufMessageDef":  "testdata/addressbook.proto",
				"ProtobufMessageType": "addressbook.AddressBook",
			},
		},
	}

	expected := make([]telegraf.Parser, 0, len(formats))
	for _, format := range formats {
		formatCfg := &cfg
		settings, hasOverride := override[format]
		if hasOverride && settings.cfg != nil {
			formatCfg = settings.cfg
		}
		formatCfg.DataFormat = format

		logger := models.NewLogger("parsers", format, cfg.MetricName)

		// Try with the new format
		if creator, found := parsers.Parsers[format]; found {
			t.Logf("using new format parser for %q...", format)
			parserNew := creator(formatCfg.MetricName)
			if settings, found := override[format]; found {
				s := reflect.Indirect(reflect.ValueOf(parserNew))
				for key, value := range settings.param {
					v := reflect.ValueOf(value)
					s.FieldByName(key).Set(v)
				}
			}
			models.SetLoggerOnPlugin(parserNew, logger)
			if p, ok := parserNew.(telegraf.Initializer); ok {
				require.NoError(t, p.Init())
			}
			expected = append(expected, parserNew)
			continue
		}

		// Try with the old format
		parserOld, err := parsers.NewParser(formatCfg)
		if err == nil {
			t.Logf("using old format parser for %q...", format)
			models.SetLoggerOnPlugin(parserOld, logger)
			if p, ok := parserOld.(telegraf.Initializer); ok {
				require.NoError(t, p.Init())
			}
			expected = append(expected, parserOld)
			continue
		}
		require.Containsf(t, err.Error(), "invalid data format:", "setup %q failed: %v", format, err)
		require.Failf(t, "%q neither found in old nor new format", format)
	}
	require.Len(t, expected, len(formats))

	actual := make([]interface{}, 0)
	generated := make([]interface{}, 0)
	for _, plugin := range c.Inputs {
		input, ok := plugin.Input.(*MockupInputPluginParserNew)
		require.True(t, ok)
		// Get the parser set with 'SetParser()'
		if p, ok := input.Parser.(*models.RunningParser); ok {
			actual = append(actual, p.Parser)
		} else {
			actual = append(actual, input.Parser)
		}
		// Get the parser set with 'SetParserFunc()'
		g, err := input.ParserFunc()
		require.NoError(t, err)
		if rp, ok := g.(*models.RunningParser); ok {
			generated = append(generated, rp.Parser)
		} else {
			generated = append(generated, g)
		}
	}
	require.Len(t, actual, len(formats))

	for i, format := range formats {
		// Determine the underlying type of the parser
		stype := reflect.Indirect(reflect.ValueOf(expected[i])).Interface()
		// Ignore all unexported fields and fields not relevant for functionality
		options := []cmp.Option{
			cmpopts.IgnoreUnexported(stype),
			cmpopts.IgnoreTypes(sync.Mutex{}),
			cmpopts.IgnoreInterfaces(struct{ telegraf.Logger }{}),
		}
		if settings, found := override[format]; found {
			options = append(options, cmpopts.IgnoreFields(stype, settings.mask...))
		}

		// Do a manual comparision as require.EqualValues will also work on unexported fields
		// that cannot be cleared or ignored.
		diff := cmp.Diff(expected[i], actual[i], options...)
		require.Emptyf(t, diff, "Difference in SetParser() for %q", format)
		diff = cmp.Diff(expected[i], generated[i], options...)
		require.Emptyf(t, diff, "Difference in SetParserFunc() for %q", format)
	}
}

func TestConfig_ParserInterfaceOldFormat(t *testing.T) {
	formats := []string{
		"collectd",
		"csv",
		"dropwizard",
		"form_urlencoded",
		"graphite",
		"grok",
		"influx",
		"json",
		"json_v2",
		"logfmt",
		"nagios",
		"prometheus",
		"prometheusremotewrite",
		"value",
		"wavefront",
		"xml", "xpath_json", "xpath_msgpack", "xpath_protobuf",
	}

	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/parsers_old.toml"))
	require.Len(t, c.Inputs, len(formats))

	cfg := parsers.Config{
		CSVHeaderRowCount:     42,
		DropwizardTagPathsMap: make(map[string]string),
		GrokPatterns:          []string{"%{COMBINED_LOG_FORMAT}"},
		JSONStrict:            true,
		MetricName:            "parser_test_old",
	}

	override := map[string]struct {
		cfg   *parsers.Config
		param map[string]interface{}
		mask  []string
	}{
		"csv": {
			param: map[string]interface{}{
				"HeaderRowCount": cfg.CSVHeaderRowCount,
			},
			mask: []string{"TimeFunc"},
		},
		"logfmt": {
			mask: []string{"Now"},
		},
		"xpath_protobuf": {
			cfg: &parsers.Config{
				MetricName:        "parser_test_new",
				XPathProtobufFile: "testdata/addressbook.proto",
				XPathProtobufType: "addressbook.AddressBook",
			},
			param: map[string]interface{}{
				"ProtobufMessageDef":  "testdata/addressbook.proto",
				"ProtobufMessageType": "addressbook.AddressBook",
			},
		},
	}

	expected := make([]telegraf.Parser, 0, len(formats))
	for _, format := range formats {
		formatCfg := &cfg
		settings, hasOverride := override[format]
		if hasOverride && settings.cfg != nil {
			formatCfg = settings.cfg
		}
		formatCfg.DataFormat = format

		logger := models.NewLogger("parsers", format, cfg.MetricName)

		// Try with the new format
		if creator, found := parsers.Parsers[format]; found {
			t.Logf("using new format parser for %q...", format)
			parserNew := creator(formatCfg.MetricName)
			if settings, found := override[format]; found {
				s := reflect.Indirect(reflect.ValueOf(parserNew))
				for key, value := range settings.param {
					v := reflect.ValueOf(value)
					s.FieldByName(key).Set(v)
				}
			}
			models.SetLoggerOnPlugin(parserNew, logger)
			if p, ok := parserNew.(telegraf.Initializer); ok {
				require.NoError(t, p.Init())
			}
			expected = append(expected, parserNew)
			continue
		}

		// Try with the old format
		parserOld, err := parsers.NewParser(formatCfg)
		if err == nil {
			t.Logf("using old format parser for %q...", format)
			models.SetLoggerOnPlugin(parserOld, logger)
			if p, ok := parserOld.(telegraf.Initializer); ok {
				require.NoError(t, p.Init())
			}
			expected = append(expected, parserOld)
			continue
		}
		require.Containsf(t, err.Error(), "invalid data format:", "setup %q failed: %v", format, err)
		require.Failf(t, "%q neither found in old nor new format", format)
	}
	require.Len(t, expected, len(formats))

	actual := make([]interface{}, 0)
	generated := make([]interface{}, 0)
	for _, plugin := range c.Inputs {
		input, ok := plugin.Input.(*MockupInputPluginParserOld)
		require.True(t, ok)
		// Get the parser set with 'SetParser()'
		if p, ok := input.Parser.(*models.RunningParser); ok {
			actual = append(actual, p.Parser)
		} else {
			actual = append(actual, input.Parser)
		}
		// Get the parser set with 'SetParserFunc()'
		g, err := input.ParserFunc()
		require.NoError(t, err)
		if rp, ok := g.(*models.RunningParser); ok {
			generated = append(generated, rp.Parser)
		} else {
			generated = append(generated, g)
		}
	}
	require.Len(t, actual, len(formats))

	for i, format := range formats {
		// Determine the underlying type of the parser
		stype := reflect.Indirect(reflect.ValueOf(expected[i])).Interface()
		// Ignore all unexported fields and fields not relevant for functionality
		options := []cmp.Option{
			cmpopts.IgnoreUnexported(stype),
			cmpopts.IgnoreTypes(sync.Mutex{}),
			cmpopts.IgnoreInterfaces(struct{ telegraf.Logger }{}),
		}
		if settings, found := override[format]; found {
			options = append(options, cmpopts.IgnoreFields(stype, settings.mask...))
		}

		// Do a manual comparision as require.EqualValues will also work on unexported fields
		// that cannot be cleared or ignored.
		diff := cmp.Diff(expected[i], actual[i], options...)
		require.Emptyf(t, diff, "Difference in SetParser() for %q", format)
		diff = cmp.Diff(expected[i], generated[i], options...)
		require.Emptyf(t, diff, "Difference in SetParserFunc() for %q", format)
	}
}

/*** Mockup INPUT plugin for (old) parser testing to avoid cyclic dependencies ***/
type MockupInputPluginParserOld struct {
	Parser     parsers.Parser
	ParserFunc parsers.ParserFunc
}

func (m *MockupInputPluginParserOld) SampleConfig() string                  { return "Mockup old parser test plugin" }
func (m *MockupInputPluginParserOld) Gather(acc telegraf.Accumulator) error { return nil }
func (m *MockupInputPluginParserOld) SetParser(parser parsers.Parser)       { m.Parser = parser }
func (m *MockupInputPluginParserOld) SetParserFunc(f parsers.ParserFunc)    { m.ParserFunc = f }

/*** Mockup INPUT plugin for (new) parser testing to avoid cyclic dependencies ***/
type MockupInputPluginParserNew struct {
	Parser     telegraf.Parser
	ParserFunc telegraf.ParserFunc
}

func (m *MockupInputPluginParserNew) SampleConfig() string                  { return "Mockup old parser test plugin" }
func (m *MockupInputPluginParserNew) Gather(acc telegraf.Accumulator) error { return nil }
func (m *MockupInputPluginParserNew) SetParser(parser telegraf.Parser)      { m.Parser = parser }
func (m *MockupInputPluginParserNew) SetParserFunc(f telegraf.ParserFunc)   { m.ParserFunc = f }

/*** Mockup INPUT plugin for testing to avoid cyclic dependencies ***/
type MockupInputPlugin struct {
	Servers      []string `toml:"servers"`
	Methods      []string `toml:"methods"`
	Timeout      Duration `toml:"timeout"`
	ReadTimeout  Duration `toml:"read_timeout"`
	WriteTimeout Duration `toml:"write_timeout"`
	MaxBodySize  Size     `toml:"max_body_size"`
	Paths        []string `toml:"paths"`
	Port         int      `toml:"port"`
	Command      string
	PidFile      string
	Log          telegraf.Logger `toml:"-"`
	tls.ServerConfig

	parser telegraf.Parser
}

func (m *MockupInputPlugin) SampleConfig() string                  { return "Mockup test input plugin" }
func (m *MockupInputPlugin) Gather(acc telegraf.Accumulator) error { return nil }
func (m *MockupInputPlugin) SetParser(parser telegraf.Parser)      { m.parser = parser }

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
func (m *MockupOuputPlugin) SampleConfig() string                  { return "Mockup test output plugin" }
func (m *MockupOuputPlugin) Write(metrics []telegraf.Metric) error { return nil }

// Register the mockup plugin on loading
func init() {
	// Register the mockup input plugin for the required names
	inputs.Add("parser_test_new", func() telegraf.Input { return &MockupInputPluginParserNew{} })
	inputs.Add("parser_test_old", func() telegraf.Input { return &MockupInputPluginParserOld{} })
	inputs.Add("exec", func() telegraf.Input { return &MockupInputPlugin{Timeout: Duration(time.Second * 5)} })
	inputs.Add("http_listener_v2", func() telegraf.Input { return &MockupInputPlugin{} })
	inputs.Add("memcached", func() telegraf.Input { return &MockupInputPlugin{} })
	inputs.Add("procstat", func() telegraf.Input { return &MockupInputPlugin{} })

	// Register the mockup output plugin for the required names
	outputs.Add("azure_monitor", func() telegraf.Output { return &MockupOuputPlugin{NamespacePrefix: "Telegraf/"} })
	outputs.Add("http", func() telegraf.Output { return &MockupOuputPlugin{} })
}
