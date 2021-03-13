package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/exec"
	"github.com/influxdata/telegraf/plugins/inputs/http_listener_v2"
	"github.com/influxdata/telegraf/plugins/inputs/memcached"
	"github.com/influxdata/telegraf/plugins/inputs/procstat"
	"github.com/influxdata/telegraf/plugins/outputs/azure_monitor"
	httpOut "github.com/influxdata/telegraf/plugins/outputs/http"
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
	assert.NoError(t, filter.Compile())
	mConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 10 * time.Second,
	}
	mConfig.Tags = make(map[string]string)

	assert.Equal(t, memcached, c.Inputs[0].Input,
		"Testdata did not produce a correct memcached struct.")
	assert.Equal(t, mConfig, c.Inputs[0].Config,
		"Testdata did not produce correct memcached metadata.")
}

func TestConfig_LoadSingleInput(t *testing.T) {
	c := NewConfig()
	c.LoadConfig("./testdata/single_plugin.toml")

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
	assert.NoError(t, filter.Compile())
	mConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 5 * time.Second,
	}
	mConfig.Tags = make(map[string]string)

	assert.Equal(t, memcached, c.Inputs[0].Input,
		"Testdata did not produce a correct memcached struct.")
	assert.Equal(t, mConfig, c.Inputs[0].Config,
		"Testdata did not produce correct memcached metadata.")
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
	assert.NoError(t, filter.Compile())
	mConfig := &models.InputConfig{
		Name:     "memcached",
		Filter:   filter,
		Interval: 5 * time.Second,
	}
	mConfig.Tags = make(map[string]string)

	assert.Equal(t, memcached, c.Inputs[0].Input,
		"Testdata did not produce a correct memcached struct.")
	assert.Equal(t, mConfig, c.Inputs[0].Config,
		"Testdata did not produce correct memcached metadata.")

	ex := inputs.Inputs["exec"]().(*exec.Exec)
	p, err := parsers.NewParser(&parsers.Config{
		MetricName: "exec",
		DataFormat: "json",
		JSONStrict: true,
	})
	assert.NoError(t, err)
	ex.SetParser(p)
	ex.Command = "/usr/bin/myothercollector --foo=bar"
	eConfig := &models.InputConfig{
		Name:              "exec",
		MeasurementSuffix: "_myothercollector",
	}
	eConfig.Tags = make(map[string]string)

	exec := c.Inputs[1].Input.(*exec.Exec)
	require.NotNil(t, exec.Log)
	exec.Log = nil

	assert.Equal(t, ex, c.Inputs[1].Input,
		"Merged Testdata did not produce a correct exec struct.")
	assert.Equal(t, eConfig, c.Inputs[1].Config,
		"Merged Testdata did not produce correct exec metadata.")

	memcached.Servers = []string{"192.168.1.1"}
	assert.Equal(t, memcached, c.Inputs[2].Input,
		"Testdata did not produce a correct memcached struct.")
	assert.Equal(t, mConfig, c.Inputs[2].Config,
		"Testdata did not produce correct memcached metadata.")

	pstat := inputs.Inputs["procstat"]().(*procstat.Procstat)
	pstat.PidFile = "/var/run/grafana-server.pid"

	pConfig := &models.InputConfig{Name: "procstat"}
	pConfig.Tags = make(map[string]string)

	assert.Equal(t, pstat, c.Inputs[3].Input,
		"Merged Testdata did not produce a correct procstat struct.")
	assert.Equal(t, pConfig, c.Inputs[3].Config,
		"Merged Testdata did not produce correct procstat metadata.")
}

func TestConfig_LoadSpecialTypes(t *testing.T) {
	c := NewConfig()
	err := c.LoadConfig("./testdata/special_types.toml")
	assert.NoError(t, err)
	require.Equal(t, 1, len(c.Inputs))

	inputHTTPListener, ok := c.Inputs[0].Input.(*http_listener_v2.HTTPListenerV2)
	assert.Equal(t, true, ok)
	// Tests telegraf duration parsing.
	assert.Equal(t, internal.Duration{Duration: time.Second}, inputHTTPListener.WriteTimeout)
	// Tests telegraf size parsing.
	assert.Equal(t, internal.Size{Size: 1024 * 1024}, inputHTTPListener.MaxBodySize)
	// Tests toml multiline basic strings.
	assert.Equal(t, "/path/to/my/cert", strings.TrimRight(inputHTTPListener.TLSCert, "\r\n"))
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
	assert.Equal(t, "Error loading config file ./testdata/wrong_field_type.toml: error parsing http_listener_v2, line 2: (http_listener_v2.HTTPListenerV2.Port) cannot unmarshal TOML string into int", err.Error())

	c = NewConfig()
	err = c.LoadConfig("./testdata/wrong_field_type2.toml")
	require.Error(t, err, "invalid field type2")
	assert.Equal(t, "Error loading config file ./testdata/wrong_field_type2.toml: error parsing http_listener_v2, line 2: (http_listener_v2.HTTPListenerV2.Methods) cannot unmarshal TOML string into []string", err.Error())
}

func TestConfig_InlineTables(t *testing.T) {
	// #4098
	c := NewConfig()
	err := c.LoadConfig("./testdata/inline_table.toml")
	assert.NoError(t, err)
	require.Equal(t, 2, len(c.Outputs))

	outputHTTP, ok := c.Outputs[1].Output.(*httpOut.HTTP)
	assert.Equal(t, true, ok)
	assert.Equal(t, map[string]string{"Authorization": "Token $TOKEN", "Content-Type": "application/json"}, outputHTTP.Headers)
	assert.Equal(t, []string{"org_id"}, c.Outputs[0].Config.Filter.TagInclude)
}

func TestConfig_SliceComment(t *testing.T) {
	t.Skipf("Skipping until #3642 is resolved")

	c := NewConfig()
	err := c.LoadConfig("./testdata/slice_comment.toml")
	assert.NoError(t, err)
	require.Equal(t, 1, len(c.Outputs))

	outputHTTP, ok := c.Outputs[0].Output.(*httpOut.HTTP)
	assert.Equal(t, []string{"test"}, outputHTTP.Scopes)
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
	defaultPrefixConfig := `[[outputs.azure_monitor]]`
	err := c.LoadConfigData([]byte(defaultPrefixConfig))
	assert.NoError(t, err)
	azureMonitor, ok := c.Outputs[0].Output.(*azure_monitor.AzureMonitor)
	assert.Equal(t, "Telegraf/", azureMonitor.NamespacePrefix)
	assert.Equal(t, true, ok)

	c = NewConfig()
	customPrefixConfig := `[[outputs.azure_monitor]]
	namespace_prefix = ""`
	err = c.LoadConfigData([]byte(customPrefixConfig))
	assert.NoError(t, err)
	azureMonitor, ok = c.Outputs[0].Output.(*azure_monitor.AzureMonitor)
	assert.Equal(t, "", azureMonitor.NamespacePrefix)
	assert.Equal(t, true, ok)
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
