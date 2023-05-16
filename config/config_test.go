package config

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/persister"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/parsers/all" // Blank import to have all parsers for testing
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/serializers"
	_ "github.com/influxdata/telegraf/plugins/serializers/all" // Blank import to have all serializers for testing
	"github.com/influxdata/telegraf/testutil"
)

func TestReadBinaryFile(t *testing.T) {
	// Create a temporary binary file using the Telegraf tool custom_builder to pass as a config
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		err := os.Chdir(wd)
		require.NoError(t, err)
	})

	err = os.Chdir("../")
	require.NoError(t, err)
	tmpdir := t.TempDir()
	binaryFile := filepath.Join(tmpdir, "custom_builder")
	cmd := exec.Command("go", "build", "-o", binaryFile, "./tools/custom_builder")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()

	require.NoError(t, err, fmt.Sprintf("stdout: %s, stderr: %s", outb.String(), errb.String()))
	c := NewConfig()
	err = c.LoadConfig(binaryFile)
	require.Error(t, err)
	require.ErrorContains(t, err, "provided config is not a TOML file")
}

func TestConfig_LoadSingleInputWithEnvVars(t *testing.T) {
	c := NewConfig()
	t.Setenv("MY_TEST_SERVER", "192.168.1.1")
	t.Setenv("TEST_INTERVAL", "10s")
	require.NoError(t, c.LoadConfig("./testdata/single_plugin_env_vars.toml"))

	input := inputs.Inputs["memcached"]().(*MockupInputPlugin)
	input.Servers = []string{"192.168.1.1"}

	filter := models.Filter{
		NameDrop:  []string{"metricname2"},
		NamePass:  []string{"metricname1", "ip_192.168.1.1_name"},
		FieldDrop: []string{"other", "stuff"},
		FieldPass: []string{"some", "strings"},
		TagDropFilters: []models.TagFilter{
			{
				Name:   "badtag",
				Values: []string{"othertag"},
			},
		},
		TagPassFilters: []models.TagFilter{
			{
				Name:   "goodtag",
				Values: []string{"mytag"},
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

	// Ignore Log, Parser and ID
	c.Inputs[0].Input.(*MockupInputPlugin).Log = nil
	c.Inputs[0].Input.(*MockupInputPlugin).parser = nil
	c.Inputs[0].Config.ID = ""
	require.Equal(t, input, c.Inputs[0].Input, "Testdata did not produce a correct mockup struct.")
	require.Equal(t, inputConfig, c.Inputs[0].Config, "Testdata did not produce correct input metadata.")
}

func TestConfig_LoadSingleInput(t *testing.T) {
	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/single_plugin.toml"))

	input := inputs.Inputs["memcached"]().(*MockupInputPlugin)
	input.Servers = []string{"localhost"}

	filter := models.Filter{
		NameDrop:  []string{"metricname2"},
		NamePass:  []string{"metricname1"},
		FieldDrop: []string{"other", "stuff"},
		FieldPass: []string{"some", "strings"},
		TagDropFilters: []models.TagFilter{
			{
				Name:   "badtag",
				Values: []string{"othertag"},
			},
		},
		TagPassFilters: []models.TagFilter{
			{
				Name:   "goodtag",
				Values: []string{"mytag"},
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

	// Ignore Log, Parser and ID
	c.Inputs[0].Input.(*MockupInputPlugin).Log = nil
	c.Inputs[0].Input.(*MockupInputPlugin).parser = nil
	c.Inputs[0].Config.ID = ""
	require.Equal(t, input, c.Inputs[0].Input, "Testdata did not produce a correct memcached struct.")
	require.Equal(t, inputConfig, c.Inputs[0].Config, "Testdata did not produce correct memcached metadata.")
}

func TestConfig_LoadDirectory(t *testing.T) {
	c := NewConfig()

	files, err := WalkDirectory("./testdata/subconfig")
	files = append([]string{"./testdata/single_plugin.toml"}, files...)
	require.NoError(t, err)
	require.NoError(t, c.LoadAll(files...))

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
		TagDropFilters: []models.TagFilter{
			{
				Name:   "badtag",
				Values: []string{"othertag"},
			},
		},
		TagPassFilters: []models.TagFilter{
			{
				Name:   "goodtag",
				Values: []string{"mytag"},
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
	parser := &json.Parser{
		MetricName: "exec",
		Strict:     true,
	}
	require.NoError(t, parser.Init())

	expectedPlugins[1].SetParser(parser)
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
		TagDropFilters: []models.TagFilter{
			{
				Name:   "badtag",
				Values: []string{"othertag"},
			},
		},
		TagPassFilters: []models.TagFilter{
			{
				Name:   "goodtag",
				Values: []string{"mytag"},
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

		// Check the parsers if any
		if expectedPlugins[i].parser != nil {
			runningParser, ok := input.parser.(*models.RunningParser)
			require.True(t, ok)

			// We only use the JSON parser here
			parser, ok := runningParser.Parser.(*json.Parser)
			require.True(t, ok)

			// Prepare parser for comparison
			require.NoError(t, parser.Init())
			parser.Log = nil

			// Compare the parser
			require.Equalf(t, expectedPlugins[i].parser, parser, "Plugin %d: incorrect parser produced", i)
		}

		// Ignore the parsers for further comparisons
		input.parser = nil
		expectedPlugins[i].parser = nil

		// Ignore the ID
		plugin.Config.ID = ""

		require.Equalf(t, expectedPlugins[i], plugin.Input, "Plugin %d: incorrect struct produced", i)
		require.Equalf(t, expectedConfigs[i], plugin.Config, "Plugin %d: incorrect config produced", i)
	}
}

func TestConfig_WrongCertPath(t *testing.T) {
	c := NewConfig()
	require.Error(t, c.LoadConfig("./testdata/wrong_cert_path.toml"))
}

func TestConfig_DefaultParser(t *testing.T) {
	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/default_parser.toml"))
}

func TestConfig_DefaultExecParser(t *testing.T) {
	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/default_parser_exec.toml"))
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
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "in input plugin without parser",
			filename: "./testdata/invalid_field.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in input plugin with parser",
			filename: "./testdata/invalid_field_with_parser.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in input plugin with parser func",
			filename: "./testdata/invalid_field_with_parserfunc.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in parser of input plugin",
			filename: "./testdata/invalid_field_in_parser_table.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in parser of input plugin with parser-func",
			filename: "./testdata/invalid_field_in_parserfunc_table.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in processor plugin without parser",
			filename: "./testdata/invalid_field_processor.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in processor plugin with parser",
			filename: "./testdata/invalid_field_processor_with_parser.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in processor plugin with parser func",
			filename: "./testdata/invalid_field_processor_with_parserfunc.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in parser of processor plugin",
			filename: "./testdata/invalid_field_processor_in_parser_table.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
		{
			name:     "in parser of processor plugin with parser-func",
			filename: "./testdata/invalid_field_processor_in_parserfunc_table.toml",
			expected: `line 1: configuration specified the fields ["not_a_field"], but they weren't used`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			err := c.LoadConfig(tt.filename)
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestConfig_WrongFieldType(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("./testdata/wrong_field_type.toml")
	require.Error(t, err, "invalid field type")
	require.Equal(
		t,
		"error loading config file ./testdata/wrong_field_type.toml: error parsing http_listener_v2, line 2: "+
			"(config.MockupInputPlugin.Port) cannot unmarshal TOML string into int",
		err.Error(),
	)

	c = NewConfig()
	err = c.LoadConfig("./testdata/wrong_field_type2.toml")
	require.Error(t, err, "invalid field type2")
	require.Equal(
		t,
		"error loading config file ./testdata/wrong_field_type2.toml: error parsing http_listener_v2, line 2: "+
			"(config.MockupInputPlugin.Methods) cannot unmarshal TOML string into []string",
		err.Error(),
	)
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
	require.Equal(
		t,
		"error loading config file ./testdata/non_slice_slice.toml: error parsing http array, line 4: cannot unmarshal TOML array into string (need slice)",
		err.Error(),
	)
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

	expected := fmt.Sprintf("error loading config file %s: retry 3 of 3 failed to retrieve remote config: 404 Not Found", ts.URL)

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
	t.Setenv("TELEGRAF_CONFIG_PATH", ts.URL)
	configPath, err := getDefaultConfigPath()
	require.NoError(t, err)
	require.Equal(t, []string{ts.URL}, configPath)
	err = c.LoadConfig("")
	require.NoError(t, err)
}

func TestConfig_URLLikeFileName(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("http:##www.example.com.conf")
	require.Error(t, err)

	if runtime.GOOS == "windows" {
		// The error file not found error message is different on Windows
		require.Equal(
			t,
			"error loading config file http:##www.example.com.conf: open http:##www.example.com.conf: The system cannot find the file specified.",
			err.Error(),
		)
	} else {
		require.Equal(t, "error loading config file http:##www.example.com.conf: open http:##www.example.com.conf: no such file or directory", err.Error())
	}
}

func TestConfig_Filtering(t *testing.T) {
	c := NewConfig()
	require.NoError(t, c.LoadAll("./testdata/filter_metricpass.toml"))
	require.Len(t, c.Processors, 1)

	in := []telegraf.Metric{
		metric.New(
			"machine",
			map[string]string{"state": "on"},
			map[string]interface{}{"value": 42.0},
			time.Date(2023, time.April, 23, 01, 15, 30, 0, time.UTC),
		),
		metric.New(
			"machine",
			map[string]string{"state": "off"},
			map[string]interface{}{"value": 23.0},
			time.Date(2023, time.April, 23, 23, 59, 01, 0, time.UTC),
		),
		metric.New(
			"temperature",
			map[string]string{},
			map[string]interface{}{"value": 23.5},
			time.Date(2023, time.April, 24, 02, 15, 30, 0, time.UTC),
		),
	}
	expected := []telegraf.Metric{
		metric.New(
			"machine",
			map[string]string{
				"state":     "on",
				"processed": "yes",
			},
			map[string]interface{}{"value": 42.0},
			time.Date(2023, time.April, 23, 01, 15, 30, 0, time.UTC),
		),
		metric.New(
			"machine",
			map[string]string{"state": "off"},
			map[string]interface{}{"value": 23.0},
			time.Date(2023, time.April, 23, 23, 59, 01, 0, time.UTC),
		),
		metric.New(
			"temperature",
			map[string]string{
				"processed": "yes",
			},
			map[string]interface{}{"value": 23.5},
			time.Date(2023, time.April, 24, 02, 15, 30, 0, time.UTC),
		),
	}

	plugin := c.Processors[0]
	var acc testutil.Accumulator
	for _, m := range in {
		require.NoError(t, plugin.Add(m, &acc))
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func TestConfig_SerializerInterfaceNewFormat(t *testing.T) {
	formats := []string{
		"carbon2",
		"csv",
		"graphite",
		"influx",
		"json",
		"msgpack",
		"nowmetric",
		"prometheus",
		"prometheusremotewrite",
		"splunkmetric",
		"wavefront",
	}

	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/serializers_new.toml"))
	require.Len(t, c.Outputs, len(formats))

	cfg := serializers.Config{}
	override := map[string]struct {
		param map[string]interface{}
		mask  []string
	}{}

	expected := make([]telegraf.Serializer, 0, len(formats))
	for _, format := range formats {
		formatCfg := &cfg
		formatCfg.DataFormat = format

		logger := models.NewLogger("serializers", format, "test")

		var serializer telegraf.Serializer
		if creator, found := serializers.Serializers[format]; found {
			serializer = creator()
		} else {
			var err error
			serializer, err = serializers.NewSerializer(formatCfg)
			require.NoErrorf(t, err, "No serializer for format %q", format)
		}

		if settings, found := override[format]; found {
			s := reflect.Indirect(reflect.ValueOf(serializer))
			for key, value := range settings.param {
				v := reflect.ValueOf(value)
				s.FieldByName(key).Set(v)
			}
		}
		models.SetLoggerOnPlugin(serializer, logger)
		if s, ok := serializer.(telegraf.Initializer); ok {
			require.NoError(t, s.Init())
		}
		expected = append(expected, serializer)
	}
	require.Len(t, expected, len(formats))

	actual := make([]interface{}, 0)
	for _, plugin := range c.Outputs {
		output, ok := plugin.Output.(*MockupOutputPluginSerializerNew)
		require.True(t, ok)
		// Get the parser set with 'SetParser()'
		if p, ok := output.Serializer.(*models.RunningSerializer); ok {
			actual = append(actual, p.Serializer)
		} else {
			actual = append(actual, output.Serializer)
		}
	}
	require.Len(t, actual, len(formats))

	for i, format := range formats {
		// Determine the underlying type of the serializer
		stype := reflect.Indirect(reflect.ValueOf(expected[i])).Interface()
		// Ignore all unexported fields and fields not relevant for functionality
		options := []cmp.Option{
			cmpopts.IgnoreUnexported(stype),
			cmpopts.IgnoreTypes(sync.Mutex{}, regexp.Regexp{}),
			cmpopts.IgnoreInterfaces(struct{ telegraf.Logger }{}),
		}
		if settings, found := override[format]; found {
			options = append(options, cmpopts.IgnoreFields(stype, settings.mask...))
		}

		// Do a manual comparision as require.EqualValues will also work on unexported fields
		// that cannot be cleared or ignored.
		diff := cmp.Diff(expected[i], actual[i], options...)
		require.Emptyf(t, diff, "Difference in SetSerializer() for %q", format)
	}
}

func TestConfig_SerializerInterfaceOldFormat(t *testing.T) {
	formats := []string{
		"carbon2",
		"csv",
		"graphite",
		"influx",
		"json",
		"msgpack",
		"nowmetric",
		"prometheus",
		"prometheusremotewrite",
		"splunkmetric",
		"wavefront",
	}

	c := NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/serializers_old.toml"))
	require.Len(t, c.Outputs, len(formats))

	cfg := serializers.Config{}
	override := map[string]struct {
		param map[string]interface{}
		mask  []string
	}{}

	expected := make([]telegraf.Serializer, 0, len(formats))
	for _, format := range formats {
		formatCfg := &cfg
		formatCfg.DataFormat = format

		logger := models.NewLogger("serializers", format, "test")

		var serializer serializers.Serializer
		if creator, found := serializers.Serializers[format]; found {
			serializer = creator()
		} else {
			var err error
			serializer, err = serializers.NewSerializer(formatCfg)
			require.NoErrorf(t, err, "No serializer for format %q", format)
		}

		if settings, found := override[format]; found {
			s := reflect.Indirect(reflect.ValueOf(serializer))
			for key, value := range settings.param {
				v := reflect.ValueOf(value)
				s.FieldByName(key).Set(v)
			}
		}
		models.SetLoggerOnPlugin(serializer, logger)
		if s, ok := serializer.(telegraf.Initializer); ok {
			require.NoError(t, s.Init())
		}
		expected = append(expected, serializer)
	}
	require.Len(t, expected, len(formats))

	actual := make([]interface{}, 0)
	for _, plugin := range c.Outputs {
		output, ok := plugin.Output.(*MockupOutputPluginSerializerOld)
		require.True(t, ok)
		// Get the parser set with 'SetParser()'
		if p, ok := output.Serializer.(*models.RunningSerializer); ok {
			actual = append(actual, p.Serializer)
		} else {
			actual = append(actual, output.Serializer)
		}
	}
	require.Len(t, actual, len(formats))

	for i, format := range formats {
		// Determine the underlying type of the serializer
		stype := reflect.Indirect(reflect.ValueOf(expected[i])).Interface()
		// Ignore all unexported fields and fields not relevant for functionality
		options := []cmp.Option{
			cmpopts.IgnoreUnexported(stype),
			cmpopts.IgnoreTypes(sync.Mutex{}, regexp.Regexp{}),
			cmpopts.IgnoreInterfaces(struct{ telegraf.Logger }{}),
		}
		if settings, found := override[format]; found {
			options = append(options, cmpopts.IgnoreFields(stype, settings.mask...))
		}

		// Do a manual comparison as require.EqualValues will also work on unexported fields
		// that cannot be cleared or ignored.
		diff := cmp.Diff(expected[i], actual[i], options...)
		require.Emptyf(t, diff, "Difference in SetSerializer() for %q", format)
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
		param map[string]interface{}
		mask  []string
	}{
		"csv": {
			param: map[string]interface{}{
				"HeaderRowCount": cfg.CSVHeaderRowCount,
			},
			mask: []string{"TimeFunc", "ResetMode"},
		},
		"xpath_protobuf": {
			param: map[string]interface{}{
				"ProtobufMessageDef":  "testdata/addressbook.proto",
				"ProtobufMessageType": "addressbook.AddressBook",
			},
		},
	}

	expected := make([]telegraf.Parser, 0, len(formats))
	for _, format := range formats {
		formatCfg := &cfg
		formatCfg.DataFormat = format

		logger := models.NewLogger("parsers", format, cfg.MetricName)

		creator, found := parsers.Parsers[format]
		require.Truef(t, found, "No parser for format %q", format)

		parser := creator(formatCfg.MetricName)
		if settings, found := override[format]; found {
			s := reflect.Indirect(reflect.ValueOf(parser))
			for key, value := range settings.param {
				v := reflect.ValueOf(value)
				s.FieldByName(key).Set(v)
			}
		}
		models.SetLoggerOnPlugin(parser, logger)
		if p, ok := parser.(telegraf.Initializer); ok {
			require.NoError(t, p.Init())
		}
		expected = append(expected, parser)
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
		param map[string]interface{}
		mask  []string
	}{
		"csv": {
			param: map[string]interface{}{
				"HeaderRowCount": cfg.CSVHeaderRowCount,
			},
			mask: []string{"TimeFunc", "ResetMode"},
		},
		"xpath_protobuf": {
			param: map[string]interface{}{
				"ProtobufMessageDef":  "testdata/addressbook.proto",
				"ProtobufMessageType": "addressbook.AddressBook",
			},
		},
	}

	expected := make([]telegraf.Parser, 0, len(formats))
	for _, format := range formats {
		formatCfg := &cfg
		formatCfg.DataFormat = format

		logger := models.NewLogger("parsers", format, cfg.MetricName)

		creator, found := parsers.Parsers[format]
		require.Truef(t, found, "No parser for format %q", format)

		parser := creator(formatCfg.MetricName)
		if settings, found := override[format]; found {
			s := reflect.Indirect(reflect.ValueOf(parser))
			for key, value := range settings.param {
				v := reflect.ValueOf(value)
				s.FieldByName(key).Set(v)
			}
		}
		models.SetLoggerOnPlugin(parser, logger)
		if p, ok := parser.(telegraf.Initializer); ok {
			require.NoError(t, p.Init())
		}
		expected = append(expected, parser)
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

		// Do a manual comparison as require.EqualValues will also work on unexported fields
		// that cannot be cleared or ignored.
		diff := cmp.Diff(expected[i], actual[i], options...)
		require.Emptyf(t, diff, "Difference in SetParser() for %q", format)
		diff = cmp.Diff(expected[i], generated[i], options...)
		require.Emptyf(t, diff, "Difference in SetParserFunc() for %q", format)
	}
}

func TestConfig_MultipleProcessorsOrder(t *testing.T) {
	tests := []struct {
		name          string
		filename      []string
		expectedOrder []string
	}{
		{
			name:     "Test the order of multiple unique processosr",
			filename: []string{"multiple_processors.toml"},
			expectedOrder: []string{
				"processor",
				"parser_test",
				"processor_parser",
				"processor_parserfunc",
			},
		},
		{
			name:     "Test using a single 'order' configuration",
			filename: []string{"multiple_processors_simple_order.toml"},
			expectedOrder: []string{
				"parser_test",
				"processor_parser",
				"processor_parserfunc",
				"processor",
			},
		},
		{
			name:     "Test using multiple 'order' configurations",
			filename: []string{"multiple_processors_messy_order.toml"},
			expectedOrder: []string{
				"parser_test",
				"processor_parserfunc",
				"processor",
				"processor_parser",
				"processor_parser",
				"processor_parserfunc",
			},
		},
		{
			name: "Test loading multiple configuration files",
			filename: []string{
				"multiple_processors.toml",
				"multiple_processors_simple_order.toml",
			},
			expectedOrder: []string{
				"processor",
				"parser_test",
				"processor_parser",
				"processor_parserfunc",
				"parser_test",
				"processor_parser",
				"processor_parserfunc",
				"processor",
			},
		},
		{
			name: "Test loading multiple configuration files both with order",
			filename: []string{
				"multiple_processors_simple_order.toml",
				"multiple_processors_messy_order.toml",
			},
			expectedOrder: []string{
				"parser_test",
				"processor_parser",
				"processor_parserfunc",
				"parser_test",
				"processor_parserfunc",
				"processor",
				"processor",
				"processor_parser",
				"processor_parser",
				"processor_parserfunc",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewConfig()
			filenames := make([]string, 0, len(test.filename))
			for _, fn := range test.filename {
				filenames = append(filenames, filepath.Join("./testdata/processor_order", fn))
			}
			require.NoError(t, c.LoadAll(filenames...))

			require.Equal(t, len(test.expectedOrder), len(c.Processors))

			var order []string
			for _, p := range c.Processors {
				order = append(order, p.Config.Name)
			}

			require.Equal(t, test.expectedOrder, order)
		})
	}
}

func TestConfig_ProcessorsWithParsers(t *testing.T) {
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
	require.NoError(t, c.LoadAll("./testdata/processors_with_parsers.toml"))
	require.Len(t, c.Processors, len(formats))

	override := map[string]struct {
		param map[string]interface{}
		mask  []string
	}{
		"csv": {
			param: map[string]interface{}{
				"HeaderRowCount": 42,
			},
			mask: []string{"TimeFunc", "ResetMode"},
		},
		"xpath_protobuf": {
			param: map[string]interface{}{
				"ProtobufMessageDef":  "testdata/addressbook.proto",
				"ProtobufMessageType": "addressbook.AddressBook",
			},
		},
	}

	expected := make([]telegraf.Parser, 0, len(formats))
	for _, format := range formats {
		logger := models.NewLogger("parsers", format, "processors_with_parsers")

		creator, found := parsers.Parsers[format]
		require.Truef(t, found, "No parser for format %q", format)

		parser := creator("parser_test")
		if settings, found := override[format]; found {
			s := reflect.Indirect(reflect.ValueOf(parser))
			for key, value := range settings.param {
				v := reflect.ValueOf(value)
				s.FieldByName(key).Set(v)
			}
		}
		models.SetLoggerOnPlugin(parser, logger)
		if p, ok := parser.(telegraf.Initializer); ok {
			require.NoError(t, p.Init())
		}
		expected = append(expected, parser)
	}
	require.Len(t, expected, len(formats))

	actual := make([]interface{}, 0)
	generated := make([]interface{}, 0)
	for _, plugin := range c.Processors {
		var processorIF telegraf.Processor
		if p, ok := plugin.Processor.(unwrappable); ok {
			processorIF = p.Unwrap()
		} else {
			processorIF = plugin.Processor.(telegraf.Processor)
		}
		require.NotNil(t, processorIF)

		processor, ok := processorIF.(*MockupProcessorPluginParser)
		require.True(t, ok)

		// Get the parser set with 'SetParser()'
		if p, ok := processor.Parser.(*models.RunningParser); ok {
			actual = append(actual, p.Parser)
		} else {
			actual = append(actual, processor.Parser)
		}
		// Get the parser set with 'SetParserFunc()'
		if processor.ParserFunc != nil {
			g, err := processor.ParserFunc()
			require.NoError(t, err)
			if rp, ok := g.(*models.RunningParser); ok {
				generated = append(generated, rp.Parser)
			} else {
				generated = append(generated, g)
			}
		} else {
			generated = append(generated, nil)
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

func TestConfigPluginIDsDifferent(t *testing.T) {
	c := NewConfig()
	c.Agent.Statefile = "/dev/null"
	require.NoError(t, c.LoadConfig("./testdata/state_persistence_input_all_different.toml"))
	require.NotEmpty(t, c.Inputs)

	// Compare generated IDs
	for i, pi := range c.Inputs {
		refid := pi.Config.ID
		require.NotEmpty(t, refid)

		// Cross-comparison
		for j, pj := range c.Inputs {
			testid := pj.Config.ID
			if i == j {
				require.Equal(t, refid, testid)
				continue
			}
			require.NotEqualf(t, refid, testid, "equal for %d, %d", i, j)
		}
	}
}

func TestConfigPluginIDsSame(t *testing.T) {
	c := NewConfig()
	c.Agent.Statefile = "/dev/null"
	require.NoError(t, c.LoadConfig("./testdata/state_persistence_input_all_same.toml"))
	require.NotEmpty(t, c.Inputs)

	// Compare generated IDs
	for i, pi := range c.Inputs {
		refid := pi.Config.ID
		require.NotEmpty(t, refid)

		// Cross-comparison
		for j, pj := range c.Inputs {
			testid := pj.Config.ID
			require.Equal(t, refid, testid, "not equal for %d, %d", i, j)
		}
	}
}

func TestPersisterInputStoreLoad(t *testing.T) {
	// Reserve a temporary state file
	file, err := os.CreateTemp("", "telegraf_state-*.json")
	require.NoError(t, err)
	filename := file.Name()
	require.NoError(t, file.Close())
	defer os.Remove(filename)

	// Load the plugins
	cstore := NewConfig()
	require.NoError(t, cstore.LoadConfig("testdata/state_persistence_input_store_load.toml"))

	// Initialize the persister for storing the state
	persisterStore := persister.Persister{
		Filename: filename,
	}
	require.NoError(t, persisterStore.Init())

	expected := make(map[string]interface{})
	for i, plugin := range cstore.Inputs {
		require.NoError(t, plugin.Init())

		// Register
		p := plugin.Input.(*MockupStatePlugin)
		require.NoError(t, persisterStore.Register(plugin.ID(), p))

		// Change the state
		p.state.Name += "_" + strings.Repeat("a", i+1)
		p.state.Version++
		p.state.Offset += uint64(i + 1)
		p.state.Bits = append(p.state.Bits, len(p.state.Bits))
		p.state.Modified, _ = time.Parse(time.RFC3339, "2022-11-03T16:49:00+02:00")

		// Store the state for later comparison
		expected[plugin.ID()] = p.GetState()
	}

	// Write state
	require.NoError(t, persisterStore.Store())

	// Load the plugins
	cload := NewConfig()
	require.NoError(t, cload.LoadConfig("testdata/state_persistence_input_store_load.toml"))
	require.Len(t, cload.Inputs, len(expected))

	// Initialize the persister for loading the state
	persisterLoad := persister.Persister{
		Filename: filename,
	}
	require.NoError(t, persisterLoad.Init())

	for _, plugin := range cload.Inputs {
		require.NoError(t, plugin.Init())

		// Register
		p := plugin.Input.(*MockupStatePlugin)
		require.NoError(t, persisterLoad.Register(plugin.ID(), p))

		// Check that the states are not yet restored
		require.NotNil(t, expected[plugin.ID()])
		require.NotEqual(t, expected[plugin.ID()], p.GetState())
	}

	// Restore states
	require.NoError(t, persisterLoad.Load())

	// Check we got what we saved.
	for _, plugin := range cload.Inputs {
		p := plugin.Input.(*MockupStatePlugin)
		require.Equal(t, expected[plugin.ID()], p.GetState())
	}
}

func TestPersisterProcessorRegistration(t *testing.T) {
	// Load the plugins
	c := NewConfig()
	require.NoError(t, c.LoadConfig("testdata/state_persistence_processors.toml"))
	require.NotEmpty(t, c.Processors)
	require.NotEmpty(t, c.AggProcessors)

	// Initialize the persister for test
	dut := persister.Persister{
		Filename: "/tmp/doesn_t_matter.json",
	}
	require.NoError(t, dut.Init())

	// Register the processors
	for _, plugin := range c.Processors {
		unwrapped := plugin.Processor.(unwrappable).Unwrap()

		p := unwrapped.(*MockupProcessorPlugin)
		require.NoError(t, dut.Register(plugin.ID(), p))
	}

	// Register the after-aggregator processors
	for _, plugin := range c.AggProcessors {
		unwrapped := plugin.Processor.(unwrappable).Unwrap()

		p := unwrapped.(*MockupProcessorPlugin)
		require.NoError(t, dut.Register(plugin.ID(), p))
	}
}

/*** Mockup INPUT plugin for (old) parser testing to avoid cyclic dependencies ***/
type MockupInputPluginParserOld struct {
	Parser     parsers.Parser
	ParserFunc parsers.ParserFunc
}

func (m *MockupInputPluginParserOld) SampleConfig() string {
	return "Mockup old parser test plugin"
}
func (m *MockupInputPluginParserOld) Gather(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupInputPluginParserOld) SetParser(parser parsers.Parser) {
	m.Parser = parser
}
func (m *MockupInputPluginParserOld) SetParserFunc(f parsers.ParserFunc) {
	m.ParserFunc = f
}

/*** Mockup INPUT plugin for (new) parser testing to avoid cyclic dependencies ***/
type MockupInputPluginParserNew struct {
	Parser     telegraf.Parser
	ParserFunc telegraf.ParserFunc
}

func (m *MockupInputPluginParserNew) SampleConfig() string {
	return "Mockup old parser test plugin"
}
func (m *MockupInputPluginParserNew) Gather(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupInputPluginParserNew) SetParser(parser telegraf.Parser) {
	m.Parser = parser
}
func (m *MockupInputPluginParserNew) SetParserFunc(f telegraf.ParserFunc) {
	m.ParserFunc = f
}

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
	Files        []string
	PidFile      string
	Log          telegraf.Logger `toml:"-"`
	tls.ServerConfig

	parser telegraf.Parser
}

func (m *MockupInputPlugin) SampleConfig() string {
	return "Mockup test input plugin"
}
func (m *MockupInputPlugin) Gather(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupInputPlugin) SetParser(parser telegraf.Parser) {
	m.parser = parser
}

/*** Mockup INPUT plugin with ParserFunc interface ***/
type MockupInputPluginParserFunc struct {
	parserFunc telegraf.ParserFunc
}

func (m *MockupInputPluginParserFunc) SampleConfig() string {
	return "Mockup test input plugin"
}
func (m *MockupInputPluginParserFunc) Gather(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupInputPluginParserFunc) SetParserFunc(pf telegraf.ParserFunc) {
	m.parserFunc = pf
}

/*** Mockup INPUT plugin without ParserFunc interface ***/
type MockupInputPluginParserOnly struct {
	parser telegraf.Parser
}

func (m *MockupInputPluginParserOnly) SampleConfig() string {
	return "Mockup test input plugin"
}
func (m *MockupInputPluginParserOnly) Gather(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupInputPluginParserOnly) SetParser(p telegraf.Parser) {
	m.parser = p
}

/*** Mockup PROCESSOR plugin for testing to avoid cyclic dependencies ***/
type MockupProcessorPluginParser struct {
	Parser     telegraf.Parser
	ParserFunc telegraf.ParserFunc
}

func (m *MockupProcessorPluginParser) Start(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupProcessorPluginParser) Stop() {
}
func (m *MockupProcessorPluginParser) SampleConfig() string {
	return "Mockup test processor plugin with parser"
}
func (m *MockupProcessorPluginParser) Apply(_ ...telegraf.Metric) []telegraf.Metric {
	return nil
}
func (m *MockupProcessorPluginParser) Add(_ telegraf.Metric, _ telegraf.Accumulator) error {
	return nil
}
func (m *MockupProcessorPluginParser) SetParser(parser telegraf.Parser) {
	m.Parser = parser
}
func (m *MockupProcessorPluginParser) SetParserFunc(f telegraf.ParserFunc) {
	m.ParserFunc = f
}

/*** Mockup PROCESSOR plugin without parser ***/
type MockupProcessorPlugin struct {
	Option string `toml:"option"`
	state  []uint64
}

func (m *MockupProcessorPlugin) Start(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupProcessorPlugin) Stop() {
}
func (m *MockupProcessorPlugin) SampleConfig() string {
	return "Mockup test processor plugin with parser"
}
func (m *MockupProcessorPlugin) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0, len(in))
	for _, m := range in {
		m.AddTag("processed", "yes")
		out = append(out, m)
	}
	return out
}
func (m *MockupProcessorPlugin) GetState() interface{} {
	return m.state
}
func (m *MockupProcessorPlugin) SetState(state interface{}) error {
	s, ok := state.([]uint64)
	if !ok {
		return fmt.Errorf("invalid state type %T", state)
	}
	m.state = s

	return nil
}

/*** Mockup PROCESSOR plugin with parser ***/
type MockupProcessorPluginParserOnly struct {
	Parser telegraf.Parser
}

func (m *MockupProcessorPluginParserOnly) Start(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupProcessorPluginParserOnly) Stop() {
}
func (m *MockupProcessorPluginParserOnly) SampleConfig() string {
	return "Mockup test processor plugin with parser"
}
func (m *MockupProcessorPluginParserOnly) Apply(_ ...telegraf.Metric) []telegraf.Metric {
	return nil
}
func (m *MockupProcessorPluginParserOnly) Add(_ telegraf.Metric, _ telegraf.Accumulator) error {
	return nil
}
func (m *MockupProcessorPluginParserOnly) SetParser(parser telegraf.Parser) {
	m.Parser = parser
}

/*** Mockup PROCESSOR plugin with parser-function ***/
type MockupProcessorPluginParserFunc struct {
	Parser telegraf.ParserFunc
}

func (m *MockupProcessorPluginParserFunc) Start(_ telegraf.Accumulator) error {
	return nil
}
func (m *MockupProcessorPluginParserFunc) Stop() {
}
func (m *MockupProcessorPluginParserFunc) SampleConfig() string {
	return "Mockup test processor plugin with parser"
}
func (m *MockupProcessorPluginParserFunc) Apply(_ ...telegraf.Metric) []telegraf.Metric {
	return nil
}
func (m *MockupProcessorPluginParserFunc) Add(_ telegraf.Metric, _ telegraf.Accumulator) error {
	return nil
}
func (m *MockupProcessorPluginParserFunc) SetParserFunc(pf telegraf.ParserFunc) {
	m.Parser = pf
}

/*** Mockup OUTPUT plugin for testing to avoid cyclic dependencies ***/
type MockupOuputPlugin struct {
	URL             string            `toml:"url"`
	Headers         map[string]string `toml:"headers"`
	Scopes          []string          `toml:"scopes"`
	NamespacePrefix string            `toml:"namespace_prefix"`
	Log             telegraf.Logger   `toml:"-"`
	tls.ClientConfig
}

func (m *MockupOuputPlugin) Connect() error {
	return nil
}
func (m *MockupOuputPlugin) Close() error {
	return nil
}
func (m *MockupOuputPlugin) SampleConfig() string {
	return "Mockup test output plugin"
}
func (m *MockupOuputPlugin) Write(_ []telegraf.Metric) error {
	return nil
}

/*** Mockup OUTPUT plugin for serializer testing to avoid cyclic dependencies ***/
type MockupOutputPluginSerializerOld struct {
	Serializer serializers.Serializer
}

func (m *MockupOutputPluginSerializerOld) SetSerializer(s serializers.Serializer) {
	m.Serializer = s
}
func (*MockupOutputPluginSerializerOld) Connect() error {
	return nil
}
func (*MockupOutputPluginSerializerOld) Close() error {
	return nil
}
func (*MockupOutputPluginSerializerOld) SampleConfig() string {
	return "Mockup test output plugin"
}
func (*MockupOutputPluginSerializerOld) Write(_ []telegraf.Metric) error {
	return nil
}

type MockupOutputPluginSerializerNew struct {
	Serializer telegraf.Serializer
}

func (m *MockupOutputPluginSerializerNew) SetSerializer(s telegraf.Serializer) {
	m.Serializer = s
}
func (*MockupOutputPluginSerializerNew) Connect() error {
	return nil
}
func (*MockupOutputPluginSerializerNew) Close() error {
	return nil
}
func (*MockupOutputPluginSerializerNew) SampleConfig() string {
	return "Mockup test output plugin"
}
func (*MockupOutputPluginSerializerNew) Write(_ []telegraf.Metric) error {
	return nil
}

/*** Mockup INPUT plugin with state for testing to avoid cyclic dependencies ***/
type MockupState struct {
	Name     string
	Version  uint64
	Offset   uint64
	Bits     []int
	Modified time.Time
}

type MockupStatePluginSettings struct {
	Name     string  `toml:"name"`
	Factor   float64 `toml:"factor"`
	Enabled  bool    `toml:"enabled"`
	BitField []int   `toml:"bits"`
}

type MockupStatePlugin struct {
	Servers  []string                    `toml:"servers"`
	Method   string                      `toml:"method"`
	Settings map[string]string           `toml:"params"`
	Port     int                         `toml:"port"`
	Setups   []MockupStatePluginSettings `toml:"setup"`
	state    MockupState
}

func (m *MockupStatePlugin) Init() error {
	t0, _ := time.Parse(time.RFC3339, "2021-04-24T23:42:00+02:00")
	m.state = MockupState{
		Name:     "mockup",
		Bits:     []int{},
		Modified: t0,
	}

	return nil
}

func (m *MockupStatePlugin) GetState() interface{} {
	return m.state
}

func (m *MockupStatePlugin) SetState(state interface{}) error {
	s, ok := state.(MockupState)
	if !ok {
		return fmt.Errorf("invalid state type %T", state)
	}
	m.state = s

	return nil
}

func (m *MockupStatePlugin) SampleConfig() string {
	return "Mockup test plugin"
}

func (m *MockupStatePlugin) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Register the mockup plugin on loading
func init() {
	// Register the mockup input plugin for the required names
	inputs.Add("parser_test_new", func() telegraf.Input {
		return &MockupInputPluginParserNew{}
	})
	inputs.Add("parser_test_old", func() telegraf.Input {
		return &MockupInputPluginParserOld{}
	})
	inputs.Add("parser", func() telegraf.Input {
		return &MockupInputPluginParserOnly{}
	})
	inputs.Add("parser_func", func() telegraf.Input {
		return &MockupInputPluginParserFunc{}
	})
	inputs.Add("exec", func() telegraf.Input {
		return &MockupInputPlugin{Timeout: Duration(time.Second * 5)}
	})
	inputs.Add("file", func() telegraf.Input {
		return &MockupInputPlugin{}
	})
	inputs.Add("http_listener_v2", func() telegraf.Input {
		return &MockupInputPlugin{}
	})
	inputs.Add("memcached", func() telegraf.Input {
		return &MockupInputPlugin{}
	})
	inputs.Add("procstat", func() telegraf.Input {
		return &MockupInputPlugin{}
	})
	inputs.Add("statetest", func() telegraf.Input {
		return &MockupStatePlugin{}
	})

	// Register the mockup processor plugin for the required names
	processors.Add("parser_test", func() telegraf.Processor {
		return &MockupProcessorPluginParser{}
	})
	processors.Add("processor", func() telegraf.Processor {
		return &MockupProcessorPlugin{}
	})
	processors.Add("processor_parser", func() telegraf.Processor {
		return &MockupProcessorPluginParserOnly{}
	})
	processors.Add("processor_parserfunc", func() telegraf.Processor {
		return &MockupProcessorPluginParserFunc{}
	})
	processors.Add("statetest", func() telegraf.Processor {
		return &MockupProcessorPlugin{}
	})

	// Register the mockup output plugin for the required names
	outputs.Add("azure_monitor", func() telegraf.Output {
		return &MockupOuputPlugin{NamespacePrefix: "Telegraf/"}
	})
	outputs.Add("http", func() telegraf.Output {
		return &MockupOuputPlugin{}
	})
	outputs.Add("serializer_test_new", func() telegraf.Output {
		return &MockupOutputPluginSerializerNew{}
	})
	outputs.Add("serializer_test_old", func() telegraf.Output {
		return &MockupOutputPluginSerializerOld{}
	})
}
