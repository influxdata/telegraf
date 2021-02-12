package configapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	"github.com/influxdata/telegraf/plugins/inputs/kafka_consumer"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	_ "github.com/influxdata/telegraf/plugins/processors/all"
	// _ "github.com/influxdata/telegraf/plugins/inputs/cpu"
	// _ "github.com/influxdata/telegraf/plugins/outputs/file"
	// _ "github.com/influxdata/telegraf/plugins/processors/rename"
)

// TestListPluginTypes tests that the config api can scrape all existing plugins
// for type information to build a schema.
func TestListPluginTypes(t *testing.T) {
	cfg := config.NewConfig() // initalizes API
	a, err := agent.NewAgent(cfg)
	require.NoError(t, err)

	newAPI(context.Background(), a)

	pluginConfigs := API.ListPluginTypes()
	require.Greater(t, len(pluginConfigs), 10)
	// b, _ := json.Marshal(pluginConfigs)
	// fmt.Println(string(b))

	// find the gnmi plugin
	var gnmi PluginConfig
	for _, conf := range pluginConfigs {
		if conf.Name == "gnmi" {
			gnmi = conf
			break
		}
	}

	// find the cloudwatch plugin
	var cloudwatch PluginConfig
	for _, conf := range pluginConfigs {
		if conf.Name == "cloudwatch" {
			cloudwatch = conf
			break
		}
	}

	// validate a slice of objects
	require.EqualValues(t, "array", gnmi.Config["Subscriptions"].Type)
	require.EqualValues(t, "object", gnmi.Config["Subscriptions"].SubType)
	require.NotNil(t, gnmi.Config["Subscriptions"].SubFields)
	require.EqualValues(t, "string", gnmi.Config["Subscriptions"].SubFields["Name"].Type)

	// validate a slice of pointer objects
	require.EqualValues(t, "array", cloudwatch.Config["Metrics"].Type)
	require.EqualValues(t, "object", cloudwatch.Config["Metrics"].SubType)
	require.NotNil(t, cloudwatch.Config["Metrics"].SubFields)
	require.EqualValues(t, "array", cloudwatch.Config["Metrics"].SubFields["StatisticExclude"].Type)
	require.EqualValues(t, "array", cloudwatch.Config["Metrics"].SubFields["MetricNames"].Type)

	// validate a map of strings
	require.EqualValues(t, "map", gnmi.Config["Aliases"].Type)
	require.EqualValues(t, "string", gnmi.Config["Aliases"].SubType)

	// check a default value
	require.EqualValues(t, "proto", gnmi.Config["Encoding"].Default)
	require.EqualValues(t, 10*1e9, gnmi.Config["Redial"].Default)

	// check anonymous composed fields
	require.EqualValues(t, "bool", gnmi.Config["InsecureSkipVerify"].Type)

	// TODO: check named composed fields

}

func TestInputPluginLifecycle(t *testing.T) {
	cfg := config.NewConfig() // initalizes API
	a, err := agent.NewAgent(cfg)
	require.NoError(t, err)
	_, outputCancel := newAPI(context.Background(), a)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Log("Running")
	go a.RunWithAPI(ctx, outputCancel)
	// TODO: defer a.Shutdown()

	// create
	t.Log("Create plugin")
	newPluginID, err := API.CreatePlugin(PluginConfigCreate{
		Name: "inputs.cpu",
		Config: map[string]interface{}{
			"percpu":           true,
			"totalcpu":         true,
			"collect_cpu_time": true,
			"report_active":    true,
		},
	})
	require.NoError(t, err)
	require.NotZero(t, len(newPluginID))

	// get plugin status
	t.Log("wait for start state")
	waitForStatus(t, newPluginID, "running", 1*time.Second)

	// list running
	t.Log("List running plugins")
	runningPlugins := API.ListRunningPlugins()
	require.Len(t, runningPlugins, 1)

	status := API.GetPluginStatus(newPluginID)
	require.Equal(t, "running", status.String())
	// delete
	t.Log("delete plugin")
	err = API.DeletePlugin(newPluginID)
	require.NoError(t, err)

	t.Log("wait for dead state")
	waitForStatus(t, newPluginID, "dead", 300*time.Millisecond)

	// get plugin status until dead
	status = API.GetPluginStatus(newPluginID)
	require.Equal(t, "dead", status.String())

	// list running should have none
	t.Log("list plugins")
	runningPlugins = API.ListRunningPlugins()
	require.Len(t, runningPlugins, 0)
}

func TestAllPluginLifecycle(t *testing.T) {
	cfg := config.NewConfig()
	a, err := agent.NewAgent(cfg)
	require.NoError(t, err)

	appCtx := context.Background()

	_, outputCancel := newAPI(appCtx, a)

	runCtx, cancel := context.WithCancel(appCtx)
	defer cancel()

	t.Log("Running")
	go a.RunWithAPI(runCtx, outputCancel)

	// create
	t.Log("Create plugin")
	pluginIDs := []models.PluginID{}
	newPluginID, err := API.CreatePlugin(PluginConfigCreate{
		Name:   "inputs.cpu",
		Config: map[string]interface{}{},
	})
	pluginIDs = append(pluginIDs, newPluginID)
	require.NoError(t, err)
	require.NotZero(t, len(newPluginID))

	newPluginID, err = API.CreatePlugin(PluginConfigCreate{
		Name: "processors.rename",
		Config: map[string]interface{}{
			"replace": map[string]interface{}{
				"tag":  "hostname",
				"dest": "a_host",
			},
		},
	})
	pluginIDs = append(pluginIDs, newPluginID)
	require.NoError(t, err)
	require.NotZero(t, len(newPluginID))

	newPluginID, err = API.CreatePlugin(PluginConfigCreate{
		Name: "outputs.file",
		Config: map[string]interface{}{
			"files": []string{"stdout"},
		},
	})
	pluginIDs = append(pluginIDs, newPluginID)
	require.NoError(t, err)
	require.NotZero(t, len(newPluginID))

	for _, id := range pluginIDs {
		waitForStatus(t, id, "running", 1*time.Second)
	}

	// list running
	t.Log("List running plugins")
	runningPlugins := API.ListRunningPlugins()
	require.Len(t, runningPlugins, 3)

	time.Sleep(5 * time.Second)

	// delete
	t.Log("delete plugins")
	for _, id := range pluginIDs {
		err = API.DeletePlugin(id)
		require.NoError(t, err)
	}

	t.Log("wait for dead state")
	for _, id := range pluginIDs {
		waitForStatus(t, id, "dead", 300*time.Millisecond)
	}

	// list running should have none
	runningPlugins = API.ListRunningPlugins()
	require.Len(t, runningPlugins, 0)

}

func waitForStatus(t *testing.T, newPluginID models.PluginID, waitStatus string, timeout time.Duration) {
	timeoutAt := time.Now().Add(timeout)
	for timeoutAt.After(time.Now()) {
		status := API.GetPluginStatus(newPluginID)
		if status.String() == waitStatus {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.FailNow(t, "timed out waiting for status "+waitStatus)
}

func TestSetFieldConfig(t *testing.T) {
	creator := inputs.Inputs["kafka_consumer"]
	cfg := map[string]interface{}{
		"name":               "alias",
		"alias":              "bar",
		"interval":           "30s",
		"collection_jitter":  "5s",
		"precision":          "1ms",
		"name_override":      "my",
		"measurement_prefix": "prefix_",
		"measurement_suffix": "_suffix",
		"tags": map[string]interface{}{
			"tag1": "value",
		},
		"filter": map[string]interface{}{
			"namedrop":  []string{"namedrop"},
			"namepass":  []string{"namepass"},
			"fielddrop": []string{"fielddrop"},
			"fieldpass": []string{"fieldpass"},
			"tagdrop": []map[string]interface{}{{
				"name":   "tagfilter",
				"filter": []string{"filter"},
			}},
			"tagpass": []map[string]interface{}{{
				"name":   "tagpass",
				"filter": []string{"tagpassfilter"},
			}},
			"tagexclude": []string{"tagexclude"},
			"taginclude": []string{"taginclude"},
		},
		"brokers":              []string{"localhost:9092"},
		"topics":               []string{"foo"},
		"topic_tag":            "foo",
		"client_id":            "tg123",
		"tls_ca":               "/etc/telegraf/ca.pem",
		"tls_cert":             "/etc/telegraf/cert.pem",
		"tls_key":              "/etc/telegraf/key.pem",
		"insecure_skip_verify": true,
		"sasl_mechanism":       "SCRAM-SHA-256",
		"sasl_version":         1,
		"compression_codec":    1,
		"sasl_username":        "Some-Username",
		"data_format":          "influx",
	}
	i := creator()
	err := setFieldConfig(cfg, i)
	require.NoError(t, err)
	expect := &kafka_consumer.KafkaConsumer{
		Brokers:  []string{"localhost:9092"},
		Topics:   []string{"foo"},
		TopicTag: "foo",
		ReadConfig: kafka.ReadConfig{
			Config: kafka.Config{
				ClientID:         "tg123",
				CompressionCodec: 1,
				SASLAuth: kafka.SASLAuth{
					SASLUsername:  "Some-Username",
					SASLMechanism: "SCRAM-SHA-256",
					SASLVersion:   intptr(1),
				},
				ClientConfig: tls.ClientConfig{
					TLSCA:              "/etc/telegraf/ca.pem",
					TLSCert:            "/etc/telegraf/cert.pem",
					TLSKey:             "/etc/telegraf/key.pem",
					InsecureSkipVerify: true,
				},
			},
		},
	}

	require.Equal(t, expect, i)

	icfg := &models.InputConfig{}
	err = setFieldConfig(cfg, icfg)
	require.NoError(t, err)
	expected := &models.InputConfig{
		Name:              "alias",
		Alias:             "bar",
		Interval:          30 * time.Second,
		CollectionJitter:  5 * time.Second,
		Precision:         1 * time.Millisecond,
		NameOverride:      "my",
		MeasurementPrefix: "prefix_",
		MeasurementSuffix: "_suffix",
		Tags: map[string]string{
			"tag1": "value",
		},
		Filter: models.Filter{
			NameDrop:  []string{"namedrop"},
			NamePass:  []string{"namepass"},
			FieldDrop: []string{"fielddrop"},
			FieldPass: []string{"fieldpass"},
			TagDrop: []models.TagFilter{{
				Name:   "tagfilter",
				Filter: []string{"filter"},
			}},
			TagPass: []models.TagFilter{{
				Name:   "tagpass",
				Filter: []string{"tagpassfilter"},
			}},
			TagExclude: []string{"tagexclude"},
			TagInclude: []string{"taginclude"},
		},
	}

	require.Equal(t, expected, icfg)
}

// TODO: test different plugin types

func TestExampleWorstPlugin(t *testing.T) {
	input := map[string]interface{}{
		"elapsed":     "3s",
		"elapsed2":    "4s",
		"readtimeout": "5s",
		"size1":       "8MiB",
		"size2":       "9MiB",
		"pointerstruct": map[string]interface{}{
			"field": "f",
		},
		"b":   true,
		"i":   1,
		"i8":  2,
		"i32": 3,
		"u8":  4,
		"f":   5.0,
		"pf":  6.0,
		"ps":  "I am a string pointer",
		// type Header map[string][]string
		"header": map[string]interface{}{
			"Content-Type": []interface{}{
				"json/application", "text/html",
			},
		},
		"fields": map[string]interface{}{
			"field1": "field1",
			"field2": 1,
			"field3": float64(5),
		},
		"reservedkeys": map[string]bool{
			"key": true,
		},
		"stringtonumber": map[string][]map[string]float64{
			"s": {
				{
					"n": 1.0,
				},
			},
		},
		"clean": []map[string]interface{}{
			{
				"field": "fieldtest",
			},
		},
		"templates": []map[string]interface{}{
			{
				"tag": "tagtest",
			},
		},
		"value": "string",
		"devicetags": map[string][]map[string]string{
			"s": {
				{
					"n": "1.0",
				},
			},
		},
		"percentiles": []interface{}{
			1,
		},
		"floatpercentiles": []interface{}{
			1.0,
		},
		"mapofstructs": map[string]interface{}{
			"src": map[string]interface{}{
				"dest": "d",
			},
		},
		"command": []interface{}{
			"string",
			1,
			2.0,
		},
		"tagslice": []interface{}{
			[]interface{}{
				"s",
			},
		},
		"address": []interface{}{
			1,
		},
	}
	readTimeout := internal.Duration{Duration: 5 * time.Second}
	b := true
	var i int = 1
	f := float64(6)
	s := "I am a string pointer"
	header := http.Header{
		"Content-Type": []string{"json/application", "text/html"},
	}
	expected := ExampleWorstPlugin{
		Elapsed:       config.Duration(3 * time.Second),
		Elapsed2:      internal.Duration{Duration: 4 * time.Second},
		ReadTimeout:   &readTimeout,
		Size1:         config.Size(8 * 1024 * 1024),
		Size2:         internal.Size{Size: 9 * 1024 * 1024},
		PointerStruct: &baseopts{Field: "f"},
		B:             &b,
		I:             &i,
		I8:            2,
		I32:           3,
		U8:            4,
		F:             5,
		PF:            &f,
		PS:            &s,
		Header:        header,
		DefaultFieldsSets: map[string]interface{}{
			"field1": "field1",
			"field2": 1,
			"field3": float64(5),
		},
		ReservedKeys: map[string]bool{
			"key": true,
		},
		StringToNumber: map[string][]map[string]float64{
			"s": {
				{
					"n": 1.0,
				},
			},
		},
		Clean: []baseopts{
			{Field: "fieldtest"},
		},
		Templates: []*baseopts{
			{Tag: "tagtest"},
		},
		Value: "string",
		DeviceTags: map[string][]map[string]string{
			"s": {
				{
					"n": "1.0",
				},
			},
		},
		Percentiles: []internal.Number{
			{Value: 1},
		},
		FloatPercentiles: []float64{1.0},
		MapOfStructs: map[string]baseopts{
			"src": {
				Dest: "d",
			},
		},
		Command: []interface{}{
			"string",
			1,
			2.0,
		},
		TagSlice: [][]string{
			{"s"},
		},
		Address: []uint16{
			1,
		},
	}

	actual := ExampleWorstPlugin{}
	err := setFieldConfig(input, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

type ExampleWorstPlugin struct {
	Elapsed           config.Duration
	Elapsed2          internal.Duration
	ReadTimeout       *internal.Duration
	Size1             config.Size
	Size2             internal.Size
	PointerStruct     *baseopts
	B                 *bool
	I                 *int
	I8                int8
	I32               int32
	U8                uint8
	F                 float64
	PF                *float64
	PS                *string
	Header            http.Header
	DefaultFieldsSets map[string]interface{} `toml:"fields"`
	ReservedKeys      map[string]bool
	StringToNumber    map[string][]map[string]float64
	Clean             []baseopts
	Templates         []*baseopts
	Value             interface{} `json:"value"`
	DeviceTags        map[string][]map[string]string
	Percentiles       []internal.Number
	FloatPercentiles  []float64
	MapOfStructs      map[string]baseopts
	Command           []interface{}
	TagSlice          [][]string
	Address           []uint16 `toml:"address"`
}

type baseopts struct {
	Field string
	Tag   string
	Dest  string
}

func intptr(i int) *int {
	return &i
}
