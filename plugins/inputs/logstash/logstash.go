package logstash

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net/http"
	"net/url"
	"time"

	jsonParser "github.com/influxdata/telegraf/plugins/parsers/json"
)

// ##### Interface #####

const sampleConfig = `
  ## This plugin reads metrics exposed by Logstash Monitoring API.
  ## https://www.elastic.co/guide/en/logstash/current/monitoring.html

  ## The URL of the exposed Logstash API endpoint
  url = "http://127.0.0.1:9600"

  ## Enable Logstash 6+ multi-pipeline statistics support
  multi_pipeline = true

  ## Should the general process statistics be gathered
  collect_process_stats = true

  ## Should the JVM specific statistics be gathered
  collect_jvm_stats = true

  ## Should the event pipelines statistics be gathered
  collect_pipelines_stats = true

  ## Should the plugin statistics be gathered
  collect_plugins_stats = true

  ## Should the queue statistics be gathered
  collect_queue_stats = true

  ## HTTP method
  # method = "GET"

  ## Optional HTTP headers
  # headers = {"X-Special-Header" = "Special-Value"}

  ## Override HTTP "Host" header
  # host_header = "logstash.example.com"

  ## Timeout for HTTP requests
  timeout = "5s"

  ## Optional HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

type Logstash struct {
	URL string `toml:"url"`

	MultiPipeline         bool `toml:"multi_pipeline"`
	CollectProcessStats   bool `toml:"collect_process_stats"`
	CollectJVMStats       bool `toml:"collect_jvm_stats"`
	CollectPipelinesStats bool `toml:"collect_pipelines_stats"`
	CollectPluginsStats   bool `toml:"collect_plugins_stats"`
	CollectQueueStats     bool `toml:"collect_queue_stats"`

	Username   string            `toml:"username"`
	Password   string            `toml:"password"`
	Method     string            `toml:"method"`
	Headers    map[string]string `toml:"headers"`
	HostHeader string            `toml:"host_header"`
	Timeout    internal.Duration `toml:"timeout"`

	tls.ClientConfig
	client *http.Client
}

// NewLogstash create an instance of the plugin with default settings
func NewLogstash() *Logstash {
	return &Logstash{
		URL:                   "http://127.0.0.1:9600",
		MultiPipeline:         true,
		CollectProcessStats:   true,
		CollectJVMStats:       true,
		CollectPipelinesStats: true,
		CollectPluginsStats:   true,
		CollectQueueStats:     true,
		Method:                "GET",
		Headers:               make(map[string]string),
		HostHeader:            "",
		Timeout:               internal.Duration{Duration: time.Second * 5},
	}
}

// init initialise this plugin instance
func init() {
	inputs.Add("logstash", func() telegraf.Input {
		return NewLogstash()
	})
}

// Description returns short info about plugin
func (logstash *Logstash) Description() string {
	return "Read metrics exposed by Logstash"
}

// SampleConfig returns details how to configure plugin
func (logstash *Logstash) SampleConfig() string {
	return sampleConfig
}

type ProcessStats struct {
	ID      string      `json:"id"`
	Process interface{} `json:"process"`
	Name    string      `json:"name"`
	Host    string      `json:"host"`
	Version string      `json:"version"`
}

type JVMStats struct {
	ID      string      `json:"id"`
	JVM     interface{} `json:"jvm"`
	Name    string      `json:"name"`
	Host    string      `json:"host"`
	Version string      `json:"version"`
}

type PipelinesStats struct {
	ID        string              `json:"id"`
	Pipelines map[string]Pipeline `json:"pipelines"`
	Name      string              `json:"name"`
	Host      string              `json:"host"`
	Version   string              `json:"version"`
}

type PipelineStats struct {
	ID       string   `json:"id"`
	Pipeline Pipeline `json:"pipeline"`
	Name     string   `json:"name"`
	Host     string   `json:"host"`
	Version  string   `json:"version"`
}

type Pipeline struct {
	Events  interface{}     `json:"events"`
	Plugins PipelinePlugins `json:"plugins"`
	Reloads interface{}     `json:"reloads"`
	Queue   PipelineQueue   `json:"queue"`
}

type Plugin struct {
	ID     string      `json:"id"`
	Events interface{} `json:"events"`
	Name   string      `json:"name"`
}

type PipelinePlugins struct {
	Inputs  []Plugin `json:"inputs"`
	Filters []Plugin `json:"filters"`
	Outputs []Plugin `json:"outputs"`
}

type PipelineQueue struct {
	Events   float64     `json:"events"`
	Type     string      `json:"type"`
	Capacity interface{} `json:"capacity"`
	Data     interface{} `json:"data"`
}

const jvmStats = "/_node/stats/jvm"
const processStats = "/_node/stats/process"
const pipelinesStats = "/_node/stats/pipelines"
const pipelineStats = "/_node/stats/pipeline"

// createHttpClient create a clients to access API
func (logstash *Logstash) createHttpClient() (*http.Client, error) {
	tlsConfig, err := logstash.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: logstash.Timeout.Duration,
	}

	return client, nil
}

// gatherJsonData query the data source and parse the response JSON
func (logstash *Logstash) gatherJsonData(url string, value interface{}) error {

	var method string
	if logstash.Method != "" {
		method = logstash.Method
	} else {
		method = "GET"
	}

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	if (logstash.Username != "") || (logstash.Password != "") {
		request.SetBasicAuth(logstash.Username, logstash.Password)
	}
	for header, value := range logstash.Headers {
		request.Header.Add(header, value)
	}
	if logstash.HostHeader != "" {
		request.Host = logstash.HostHeader
	}

	response, err := logstash.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(value)
	if err != nil {
		return err
	}

	return nil
}

// gatherJVMStats gather the JVM metrics and add results to the accumulator
func (logstash *Logstash) gatherJVMStats(url string, accumulator telegraf.Accumulator) error {
	jvmStats := &JVMStats{}

	err := logstash.gatherJsonData(url, jvmStats)
	if err != nil {
		return err
	}

	tags := map[string]string{
		"node_id":      jvmStats.ID,
		"node_name":    jvmStats.Name,
		"node_host":    jvmStats.Host,
		"node_version": jvmStats.Version,
	}

	flattener := jsonParser.JSONFlattener{}
	err = flattener.FlattenJSON("", jvmStats.JVM)
	if err != nil {
		return err
	}
	accumulator.AddFields("logstash_jvm", flattener.Fields, tags)

	return nil
}

// gatherJVMStats gather the Process metrics and add results to the accumulator
func (logstash *Logstash) gatherProcessStats(url string, accumulator telegraf.Accumulator) error {
	processStats := &ProcessStats{}

	err := logstash.gatherJsonData(url, processStats)
	if err != nil {
		return err
	}

	tags := map[string]string{
		"node_id":      processStats.ID,
		"node_name":    processStats.Name,
		"node_host":    processStats.Host,
		"node_version": processStats.Version,
	}

	flattener := jsonParser.JSONFlattener{}
	err = flattener.FlattenJSON("", processStats.Process)
	if err != nil {
		return err
	}
	accumulator.AddFields("logstash_process", flattener.Fields, tags)

	return nil
}

// gatherPluginsStats go through a list of plugins and add their metrics to the accumulator
func (logstash *Logstash) gatherPluginsStats(
	plugins []Plugin,
	pluginType string,
	tags map[string]string,
	accumulator telegraf.Accumulator) error {

	for _, plugin := range plugins {
		pluginTags := map[string]string{
			"plugin_name": plugin.Name,
			"plugin_id":   plugin.ID,
			"plugin_type": pluginType,
		}
		for tag, value := range tags {
			pluginTags[tag] = value
		}
		flattener := jsonParser.JSONFlattener{}
		err := flattener.FlattenJSON("", plugin.Events)
		if err != nil {
			return err
		}
		accumulator.AddFields("logstash_plugins", flattener.Fields, pluginTags)
	}

	return nil
}

func (logstash *Logstash) gatherQueueStats(
	queue *PipelineQueue,
	tags map[string]string,
	accumulator telegraf.Accumulator) error {

	var err error
	queueTags := map[string]string{
		"queue_type": queue.Type,
	}
	for tag, value := range tags {
		queueTags[tag] = value
	}

	queueFields := map[string]interface{}{
		"events": queue.Events,
	}

	if queue.Type != "memory" {
		flattener := jsonParser.JSONFlattener{}
		err = flattener.FlattenJSON("", queue.Capacity)
		if err != nil {
			return err
		}
		err = flattener.FlattenJSON("", queue.Data)
		if err != nil {
			return err
		}
		for field, value := range flattener.Fields {
			queueFields[field] = value
		}
	}

	accumulator.AddFields("logstash_queue", queueFields, queueTags)

	return nil
}

// gatherJVMStats gather the Pipeline metrics and add results to the accumulator (for Logstash < 6)
func (logstash *Logstash) gatherPipelineStats(url string, accumulator telegraf.Accumulator) error {
	pipelineStats := &PipelineStats{}

	err := logstash.gatherJsonData(url, pipelineStats)
	if err != nil {
		return err
	}

	tags := map[string]string{
		"node_id":      pipelineStats.ID,
		"node_name":    pipelineStats.Name,
		"node_host":    pipelineStats.Host,
		"node_version": pipelineStats.Version,
	}

	flattener := jsonParser.JSONFlattener{}
	err = flattener.FlattenJSON("", pipelineStats.Pipeline.Events)
	if err != nil {
		return err
	}
	accumulator.AddFields("logstash_events", flattener.Fields, tags)

	if logstash.CollectPluginsStats {
		err = logstash.gatherPluginsStats(pipelineStats.Pipeline.Plugins.Inputs, "input", tags, accumulator)
		if err != nil {
			return err
		}
		err = logstash.gatherPluginsStats(pipelineStats.Pipeline.Plugins.Filters, "filter", tags, accumulator)
		if err != nil {
			return err
		}
		err = logstash.gatherPluginsStats(pipelineStats.Pipeline.Plugins.Outputs, "output", tags, accumulator)
		if err != nil {
			return err
		}
	}

	if logstash.CollectQueueStats {
		err = logstash.gatherQueueStats(&pipelineStats.Pipeline.Queue, tags, accumulator)
	}

	return nil
}

// gatherJVMStats gather the Pipelines metrics and add results to the accumulator (for Logstash >= 6)
func (logstash *Logstash) gatherPipelinesStats(url string, accumulator telegraf.Accumulator) error {
	pipelinesStats := &PipelinesStats{}

	err := logstash.gatherJsonData(url, pipelinesStats)
	if err != nil {
		return err
	}

	for pipelineName, pipeline := range pipelinesStats.Pipelines {
		tags := map[string]string{
			"node_id":      pipelinesStats.ID,
			"node_name":    pipelinesStats.Name,
			"node_host":    pipelinesStats.Host,
			"node_version": pipelinesStats.Version,
			"pipeline":     pipelineName,
		}

		flattener := jsonParser.JSONFlattener{}
		err := flattener.FlattenJSON("", pipeline.Events)
		if err != nil {
			return err
		}
		accumulator.AddFields("logstash_events", flattener.Fields, tags)

		if logstash.CollectPluginsStats {
			err = logstash.gatherPluginsStats(pipeline.Plugins.Inputs, "input", tags, accumulator)
			if err != nil {
				return err
			}
			err = logstash.gatherPluginsStats(pipeline.Plugins.Filters, "filter", tags, accumulator)
			if err != nil {
				return err
			}
			err = logstash.gatherPluginsStats(pipeline.Plugins.Outputs, "output", tags, accumulator)
			if err != nil {
				return err
			}
		}

		if logstash.CollectQueueStats {
			err = logstash.gatherQueueStats(&pipeline.Queue, tags, accumulator)
		}

	}

	return nil
}

// Gather ask this plugin to start gathering metrics
func (logstash *Logstash) Gather(accumulator telegraf.Accumulator) error {

	if logstash.client == nil {
		client, err := logstash.createHttpClient()

		if err != nil {
			return err
		}
		logstash.client = client
	}

	if logstash.CollectJVMStats {
		jvmUrl, err := url.Parse(logstash.URL + jvmStats)
		if err != nil {
			return err
		}
		if err := logstash.gatherJVMStats(jvmUrl.String(), accumulator); err != nil {
			return err
		}
	}

	if logstash.CollectProcessStats {
		processUrl, err := url.Parse(logstash.URL + processStats)
		if err != nil {
			return err
		}
		if err := logstash.gatherProcessStats(processUrl.String(), accumulator); err != nil {
			return err
		}
	}

	if logstash.CollectPipelinesStats {
		if logstash.MultiPipeline {
			pipelinesUrl, err := url.Parse(logstash.URL + pipelinesStats)
			if err != nil {
				return err
			}
			if err := logstash.gatherPipelinesStats(pipelinesUrl.String(), accumulator); err != nil {
				return err
			}
		} else {
			pipelineUrl, err := url.Parse(logstash.URL + pipelineStats)
			if err != nil {
				return err
			}
			if err := logstash.gatherPipelineStats(pipelineUrl.String(), accumulator); err != nil {
				return err
			}
		}
	}

	return nil
}
