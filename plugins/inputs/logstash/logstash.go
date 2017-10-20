package logstash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const sampleConfig = `
  ## This plugin reads metrics exposed by Logstash Monitoring API.
  #
  # logstashURL = "http://localhost:9600"
  logstashURL = "http://localhost:9600"
`
const jvmStats = "/_node/stats/jvm"
const processStats = "/_node/stats/process"
const pipelineStats = "/_node/stats/pipeline"

// Logstash Plugin Struct
type Logstash struct {
	LogstashURL string
	client      *http.Client
}

// CommonData provides struct shared between all API calls
type CommonData struct {
	Host        string `json:"host"`
	Version     string `json:"version"`
	HTTPAddress string `json:"http_address"`
	ID          string `json:"id"`
	Name        string `json:"name"`
}

// JVMStats data structure
type JVMStats struct {
	CommonData CommonData
	JVM        interface{} `json:"jvm"`
}

//ProcessStats data structure
type ProcessStats struct {
	CommonData CommonData
	Process    interface{} `json:"process"`
}

//PluginEvents data structure
type PluginEvents struct {
	DurationInMillis float64 `json:"duration_in_millis"`
	In               float64 `json:"in"`
	Out              float64 `json:"out"`
}

//Plugin data structure
type Plugin struct {
	ID     string       `json:"id"`
	Events PluginEvents `json:"events"`
	Name   string       `json:"name"`
}

//PipelinePlugins data structure
type PipelinePlugins struct {
	Inputs  []Plugin `json:"inputs"`
	Filters []Plugin `json:"filters"`
	Outputs []Plugin `json:"outputs"`
}

//PipelineQueue data structure
type PipelineQueue struct {
	Events   float64     `json:"events"`
	Qtype    string      `json:"type"`
	Capacity interface{} `json:"capacity"`
	Data     interface{} `json:"data"`
}

//Pipeline data structure
type Pipeline struct {
	Events  interface{}     `json:"events"`
	Plugins PipelinePlugins `json:"plugins"`
	Reloads interface{}     `json:"reloads"`
	Queue   PipelineQueue   `json:"queue"`
}

//PipelineStats data structure
type PipelineStats struct {
	CommonData CommonData
	Pipeline   Pipeline `json:"pipeline"`
}

//Description returns short info about plugin
func (l *Logstash) Description() string { return "Read metricts exposed by Logstash" }

//SampleConfig returns details how to configure plugin
func (l *Logstash) SampleConfig() string { return sampleConfig }

//createHttpClient create clients to access API
func (l *Logstash) createHTTPClient() (*http.Client, error) {

	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(3 * time.Second),
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(4 * time.Second),
	}

	return client, nil
}

func (l *Logstash) gatherJSONData(url string, v interface{}) error {

	r, err := l.client.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("Logstash: API responded with status-code %d, expected %d",
			r.StatusCode, http.StatusOK)
	}

	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func (l *Logstash) gatherJVMStats(url string, acc telegraf.Accumulator) error {
	JVMStats := &JVMStats{}
	if err := l.gatherJSONData(url, JVMStats); err != nil {
		return err
	}

	tags := map[string]string{
		"node_id": JVMStats.CommonData.ID,
	}

	stats := map[string]interface{}{
		"jvm": JVMStats.JVM,
	}

	now := time.Now()
	for p, s := range stats {
		f := jsonparser.JSONFlattener{}
		// parse Json, ignoring strings and bools
		err := f.FlattenJSON("", s)
		if err != nil {
			return err
		}
		acc.AddFields("logstash_"+p, f.Fields, tags, now)
	}

	return nil
}

func (l *Logstash) gatherProcessStats(url string, acc telegraf.Accumulator) error {
	ProcessStats := &ProcessStats{}
	if err := l.gatherJSONData(url, ProcessStats); err != nil {
		return err
	}

	tags := map[string]string{
		"node_id": ProcessStats.CommonData.ID,
	}

	stats := map[string]interface{}{
		"process": ProcessStats.Process,
	}

	now := time.Now()
	for p, s := range stats {
		f := jsonparser.JSONFlattener{}
		// parse Json, ignoring strings and bools
		err := f.FlattenJSON("", s)
		if err != nil {
			return err
		}
		acc.AddFields("logstash_"+p, f.Fields, tags, now)
	}

	return nil
}

func (l *Logstash) gatherPipelineStats(url string, acc telegraf.Accumulator) error {
	PipelineStats := &PipelineStats{}
	if err := l.gatherJSONData(url, PipelineStats); err != nil {
		return err
	}

	tags := map[string]string{
		"node_id": PipelineStats.CommonData.ID,
	}

	stats := map[string]interface{}{
		"events": PipelineStats.Pipeline.Events,
	}

	now := time.Now()
	// Events
	for p, s := range stats {
		f := jsonparser.JSONFlattener{}
		// parse Json, ignoring strings and bools
		err := f.FlattenJSON("", s)
		if err != nil {
			return err
		}
		acc.AddFields("logstash_"+p, f.Fields, tags, now)
	}

	// Input Plugins
	for _, plugin := range PipelineStats.Pipeline.Plugins.Inputs {
		//plugin := &plugin
		fields := map[string]interface{}{
			"name":               plugin.Name,
			"duration_in_millis": plugin.Events.DurationInMillis,
			"in":                 plugin.Events.In,
			"out":                plugin.Events.Out,
		}
		acc.AddFields("logstash_plugin_input_"+plugin.Name, fields, tags, now)
	}

	// Filters Plugins
	for _, plugin := range PipelineStats.Pipeline.Plugins.Filters {
		//plugin := &plugin
		fields := map[string]interface{}{
			"name":               plugin.Name,
			"duration_in_millis": plugin.Events.DurationInMillis,
			"in":                 plugin.Events.In,
			"out":                plugin.Events.Out,
		}
		acc.AddFields("logstash_plugin_filter_"+plugin.Name, fields, tags, now)
	}

	// Output Plugins
	for _, plugin := range PipelineStats.Pipeline.Plugins.Outputs {
		//plugin := &plugin
		fields := map[string]interface{}{
			"name":               plugin.Name,
			"duration_in_millis": plugin.Events.DurationInMillis,
			"in":                 plugin.Events.In,
			"out":                plugin.Events.Out,
		}
		acc.AddFields("logstash_plugin_output_"+plugin.Name, fields, tags, now)
	}

	return nil
}

//Gather is main function to gather all metrics provided by this plugin
func (l *Logstash) Gather(acc telegraf.Accumulator) error {

	if l.client == nil {
		client, err := l.createHTTPClient()

		if err != nil {
			return err
		}
		l.client = client
	}

	if err := l.gatherJVMStats(l.LogstashURL+jvmStats, acc); err != nil {
		err = fmt.Errorf("Can't gather JVM Stats from: " + l.LogstashURL + jvmStats)
		return err
	}

	if err := l.gatherProcessStats(l.LogstashURL+processStats, acc); err != nil {
		err = fmt.Errorf("Can't gather Process Stats from: " + l.LogstashURL + processStats)
		return err
	}

	if err := l.gatherPipelineStats(l.LogstashURL+pipelineStats, acc); err != nil {
		err = fmt.Errorf("Can't gather Pipeline Stats from: " + l.LogstashURL + pipelineStats)
		return err
	}

	return nil
}

func init() {
	inputs.Add("logstash", func() telegraf.Input { return &Logstash{} })
}
