package logstash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const sampleConfig = `
  ## This plugin reads metrics exposed by Logstash Monitoring API.
  #
  # url = "http://localhost:9600"
  url = "http://localhost:9600"
`
const jvmStats = "/_node/stats/jvm"
const processStats = "/_node/stats/process"
const pipelineStats = "/_node/stats/pipelines"

type Logstash struct {
	URL    string
	client *http.Client
}

type JVMStats struct {
	ID  string      `json:"id"`
	JVM interface{} `json:"jvm"`
}

type ProcessStats struct {
	ID      string      `json:"id"`
	Process interface{} `json:"process"`
}

type PluginEvents struct {
	QueuePushDurationInMillis float64 `json:"queue_push_duration_in_millis"`
	DurationInMillis          float64 `json:"duration_in_millis"`
	In                        float64 `json:"in"`
	Out                       float64 `json:"out"`
}

type Plugin struct {
	ID     string       `json:"id"`
	Events PluginEvents `json:"events"`
	Name   string       `json:"name"`
}

type PipelinePlugins struct {
	Inputs  []Plugin `json:"inputs"`
	Filters []Plugin `json:"filters"`
	Outputs []Plugin `json:"outputs"`
}

type PipelineQueue struct {
	Events   float64     `json:"events"`
	Qtype    string      `json:"type"`
	Capacity interface{} `json:"capacity"`
	Data     interface{} `json:"data"`
}

type Pipeline struct {
	Events  interface{}     `json:"events"`
	Plugins PipelinePlugins `json:"plugins"`
	Reloads interface{}     `json:"reloads"`
	Queue   PipelineQueue   `json:"queue"`
}

type PipelineTopLevel struct {
	Pipeline_Monitoring interface{} `json:".monitoring-logstash"`
	Pipeline_Main       Pipeline    `json:"main"`
}

type PipelineStats struct {
	ID        string           `json:"id"`
	Pipelines PipelineTopLevel `json:"pipelines"`
}

//Description returns short info about plugin
func (l *Logstash) Description() string { return "Read metrics exposed by Logstash" }

//SampleConfig returns details how to configure plugin
func (l *Logstash) SampleConfig() string { return sampleConfig }

//createHttpClient create clients to access API
func (l *Logstash) createHTTPClient() (*http.Client, error) {

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(4 * time.Second),
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
		"node_id": JVMStats.ID,
	}

	stats := map[string]interface{}{
		"jvm": JVMStats.JVM,
	}

	for p, s := range stats {
		f := jsonparser.JSONFlattener{}
		// parse Json, ignoring strings and bools
		err := f.FlattenJSON("", s)
		if err != nil {
			return err
		}
		acc.AddFields("logstash_"+p, f.Fields, tags)
	}

	return nil
}

func (l *Logstash) gatherProcessStats(url string, acc telegraf.Accumulator) error {
	ProcessStats := &ProcessStats{}
	if err := l.gatherJSONData(url, ProcessStats); err != nil {
		return err
	}

	tags := map[string]string{
		"node_id": ProcessStats.ID,
	}

	stats := map[string]interface{}{
		"process": ProcessStats.Process,
	}

	for p, s := range stats {
		f := jsonparser.JSONFlattener{}
		// parse Json, ignoring strings and bools
		err := f.FlattenJSON("", s)
		if err != nil {
			return err
		}
		acc.AddFields("logstash_"+p, f.Fields, tags)
	}

	return nil
}

func (l *Logstash) gatherPipelineStats(url string, acc telegraf.Accumulator) error {
	PipelineStats := &PipelineStats{}
	if err := l.gatherJSONData(url, PipelineStats); err != nil {
		return err
	}

	tags := map[string]string{
		"node_id": PipelineStats.ID,
	}

	stats := map[string]interface{}{
		"events": PipelineStats.Pipelines.Pipeline_Main.Events,
	}

	// Events
	for p, s := range stats {
		f := jsonparser.JSONFlattener{}
		// parse Json, ignoring strings and bools
		err := f.FlattenJSON("", s)
		if err != nil {
			return err
		}
		acc.AddFields("logstash_"+p, f.Fields, tags)
	}
	// Input Plugins
	for _, plugin := range PipelineStats.Pipelines.Pipeline_Main.Plugins.Inputs {

		fields := map[string]interface{}{
			"queue_push_duration_in_millis": plugin.Events.QueuePushDurationInMillis,
			"duration_in_millis":            plugin.Events.DurationInMillis,
			"in":                            plugin.Events.In,
			"out":                           plugin.Events.Out,
		}
		tags := map[string]string{
			"plugin": plugin.Name,
			"type":   "input",
		}
		acc.AddFields("logstash_plugins", fields, tags)
	}

	// Filters Plugins
	for _, plugin := range PipelineStats.Pipelines.Pipeline_Main.Plugins.Filters {

		fields := map[string]interface{}{
			"duration_in_millis": plugin.Events.DurationInMillis,
			"in":                 plugin.Events.In,
			"out":                plugin.Events.Out,
		}
		tags := map[string]string{
			"plugin": plugin.Name,
			"type":   "filter",
		}
		acc.AddFields("logstash_plugins", fields, tags)
	}

	// Output Plugins
	for _, plugin := range PipelineStats.Pipelines.Pipeline_Main.Plugins.Outputs {

		fields := map[string]interface{}{
			"duration_in_millis": plugin.Events.DurationInMillis,
			"in":                 plugin.Events.In,
			"out":                plugin.Events.Out,
		}
		tags := map[string]string{
			"plugin": plugin.Name,
			"type":   "output",
		}
		acc.AddFields("logstash_plugins", fields, tags)
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

	jvm_url, err := url.Parse(l.URL + jvmStats)
	if err != nil {
		return err
	}
	if err := l.gatherJVMStats(jvm_url.String(), acc); err != nil {
		return err
	}

	process_url, err := url.Parse(l.URL + processStats)
	if err != nil {
		return err
	}
	if err := l.gatherProcessStats(process_url.String(), acc); err != nil {
		return err
	}

	pipeline_url, err := url.Parse(l.URL + pipelineStats)
	if err != nil {
		return err
	}
	if err := l.gatherPipelineStats(pipeline_url.String(), acc); err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("logstash", func() telegraf.Input { return &Logstash{} })
}
