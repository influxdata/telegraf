package logstash

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonParser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const sampleConfig = `
  ## The URL of the exposed Logstash API endpoint.
  url = "http://127.0.0.1:9600"

  ## Use Logstash 5 single pipeline API, set to true when monitoring
  ## Logstash 5.
  # single_pipeline = false

  ## Enable optional collection components.  Can contain
  ## "pipelines", "process", and "jvm".
  # collect = ["pipelines", "process", "jvm"]

  ## Timeout for HTTP requests.
  # timeout = "5s"

  ## Optional HTTP Basic Auth credentials.
  # username = "username"
  # password = "pa$$word"

  ## Optional TLS Config.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Use TLS but skip chain & host verification.
  # insecure_skip_verify = false

  ## Optional HTTP headers.
  # [inputs.logstash.headers]
  #   "X-Special-Header" = "Special-Value"
`

type Logstash struct {
	URL string `toml:"url"`

	SinglePipeline bool     `toml:"single_pipeline"`
	Collect        []string `toml:"collect"`

	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Headers  map[string]string `toml:"headers"`
	Timeout  internal.Duration `toml:"timeout"`
	tls.ClientConfig

	client *http.Client
}

// NewLogstash create an instance of the plugin with default settings
func NewLogstash() *Logstash {
	return &Logstash{
		URL:            "http://127.0.0.1:9600",
		SinglePipeline: false,
		Collect:        []string{"pipelines", "process", "jvm"},
		Headers:        make(map[string]string),
		Timeout:        internal.Duration{Duration: time.Second * 5},
	}
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

func (i *Logstash) Init() error {
	err := choice.CheckSlice(i.Collect, []string{"pipelines", "process", "jvm"})
	if err != nil {
		return fmt.Errorf(`cannot verify "collect" setting: %v`, err)
	}
	return nil
}

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
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if (logstash.Username != "") || (logstash.Password != "") {
		request.SetBasicAuth(logstash.Username, logstash.Password)
	}

	for header, value := range logstash.Headers {
		if strings.ToLower(header) == "host" {
			request.Host = value
		} else {
			request.Header.Add(header, value)
		}
	}

	response, err := logstash.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		// ignore the err here; LimitReader returns io.EOF and we're not interested in read errors.
		body, _ := ioutil.ReadAll(io.LimitReader(response.Body, 200))
		return fmt.Errorf("%s returned HTTP status %s: %q", url, response.Status, body)
	}

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
		"node_version": jvmStats.Version,
		"source":       jvmStats.Host,
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
		"node_version": processStats.Version,
		"source":       processStats.Host,
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
		"node_version": pipelineStats.Version,
		"source":       pipelineStats.Host,
	}

	flattener := jsonParser.JSONFlattener{}
	err = flattener.FlattenJSON("", pipelineStats.Pipeline.Events)
	if err != nil {
		return err
	}
	accumulator.AddFields("logstash_events", flattener.Fields, tags)

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

	err = logstash.gatherQueueStats(&pipelineStats.Pipeline.Queue, tags, accumulator)
	if err != nil {
		return err
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
			"node_version": pipelinesStats.Version,
			"pipeline":     pipelineName,
			"source":       pipelinesStats.Host,
		}

		flattener := jsonParser.JSONFlattener{}
		err := flattener.FlattenJSON("", pipeline.Events)
		if err != nil {
			return err
		}
		accumulator.AddFields("logstash_events", flattener.Fields, tags)

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

		err = logstash.gatherQueueStats(&pipeline.Queue, tags, accumulator)
		if err != nil {
			return err
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

	if choice.Contains("jvm", logstash.Collect) {
		jvmUrl, err := url.Parse(logstash.URL + jvmStats)
		if err != nil {
			return err
		}
		if err := logstash.gatherJVMStats(jvmUrl.String(), accumulator); err != nil {
			return err
		}
	}

	if choice.Contains("process", logstash.Collect) {
		processUrl, err := url.Parse(logstash.URL + processStats)
		if err != nil {
			return err
		}
		if err := logstash.gatherProcessStats(processUrl.String(), accumulator); err != nil {
			return err
		}
	}

	if choice.Contains("pipelines", logstash.Collect) {
		if logstash.SinglePipeline {
			pipelineUrl, err := url.Parse(logstash.URL + pipelineStats)
			if err != nil {
				return err
			}
			if err := logstash.gatherPipelineStats(pipelineUrl.String(), accumulator); err != nil {
				return err
			}
		} else {
			pipelinesUrl, err := url.Parse(logstash.URL + pipelinesStats)
			if err != nil {
				return err
			}
			if err := logstash.gatherPipelinesStats(pipelinesUrl.String(), accumulator); err != nil {
				return err
			}
		}
	}

	return nil
}

// init registers this plugin instance
func init() {
	inputs.Add("logstash", func() telegraf.Input {
		return NewLogstash()
	})
}
