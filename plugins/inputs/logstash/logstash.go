package logstash

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonParser "github.com/influxdata/telegraf/plugins/parsers/json"
)

type Logstash struct {
	URL string `toml:"url"`

	SinglePipeline bool     `toml:"single_pipeline"`
	Collect        []string `toml:"collect"`

	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Headers  map[string]string `toml:"headers"`
	Timeout  config.Duration   `toml:"timeout"`
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
		Timeout:        config.Duration(time.Second * 5),
	}
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
	ID           string                 `json:"id"`
	Events       interface{}            `json:"events"`
	Name         string                 `json:"name"`
	BulkRequests map[string]interface{} `json:"bulk_requests"`
	Documents    map[string]interface{} `json:"documents"`
}

type PipelinePlugins struct {
	Inputs  []Plugin `json:"inputs"`
	Filters []Plugin `json:"filters"`
	Outputs []Plugin `json:"outputs"`
}

type PipelineQueue struct {
	Events              float64     `json:"events"`
	EventsCount         *float64    `json:"events_count"`
	Type                string      `json:"type"`
	Capacity            interface{} `json:"capacity"`
	Data                interface{} `json:"data"`
	QueueSizeInBytes    *float64    `json:"queue_size_in_bytes"`
	MaxQueueSizeInBytes *float64    `json:"max_queue_size_in_bytes"`
}

const jvmStats = "/_node/stats/jvm"
const processStats = "/_node/stats/process"
const pipelinesStats = "/_node/stats/pipelines"
const pipelineStats = "/_node/stats/pipeline"

func (logstash *Logstash) Init() error {
	err := choice.CheckSlice(logstash.Collect, []string{"pipelines", "process", "jvm"})
	if err != nil {
		return fmt.Errorf(`cannot verify "collect" setting: %v`, err)
	}
	return nil
}

// createHTTPClient create a clients to access API
func (logstash *Logstash) createHTTPClient() (*http.Client, error) {
	tlsConfig, err := logstash.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(logstash.Timeout),
	}

	return client, nil
}

// gatherJSONData query the data source and parse the response JSON
func (logstash *Logstash) gatherJSONData(address string, value interface{}) error {
	request, err := http.NewRequest("GET", address, nil)
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
		body, _ := io.ReadAll(io.LimitReader(response.Body, 200))
		return fmt.Errorf("%s returned HTTP status %s: %q", address, response.Status, body)
	}

	err = json.NewDecoder(response.Body).Decode(value)
	if err != nil {
		return err
	}

	return nil
}

// gatherJVMStats gather the JVM metrics and add results to the accumulator
func (logstash *Logstash) gatherJVMStats(address string, accumulator telegraf.Accumulator) error {
	jvmStats := &JVMStats{}

	err := logstash.gatherJSONData(address, jvmStats)
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
func (logstash *Logstash) gatherProcessStats(address string, accumulator telegraf.Accumulator) error {
	processStats := &ProcessStats{}

	err := logstash.gatherJSONData(address, processStats)
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
	accumulator telegraf.Accumulator,
) error {
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
		/*
			The elasticsearch output produces additional stats around
			bulk requests and document writes (that are elasticsearch specific).
			Collect those here
		*/
		if pluginType == "output" && plugin.Name == "elasticsearch" {
			/*
				The "bulk_requests" section has details about batch writes
				into Elasticsearch

				  "bulk_requests" : {
					"successes" : 2870,
					"responses" : {
					  "200" : 2870
					},
					"failures": 262,
					"with_errors": 9089
				  },
			*/
			flattener := jsonParser.JSONFlattener{}
			err := flattener.FlattenJSON("", plugin.BulkRequests)
			if err != nil {
				return err
			}
			for k, v := range flattener.Fields {
				if strings.HasPrefix(k, "bulk_requests") {
					continue
				}
				newKey := fmt.Sprintf("bulk_requests_%s", k)
				flattener.Fields[newKey] = v
				delete(flattener.Fields, k)
			}
			accumulator.AddFields("logstash_plugins", flattener.Fields, pluginTags)

			/*
				The "documents" section has counts of individual documents
				written/retried/etc.
				  "documents" : {
					"successes" : 2665549,
					"retryable_failures": 13733
				  }
			*/
			flattener = jsonParser.JSONFlattener{}
			err = flattener.FlattenJSON("", plugin.Documents)
			if err != nil {
				return err
			}
			for k, v := range flattener.Fields {
				if strings.HasPrefix(k, "documents") {
					continue
				}
				newKey := fmt.Sprintf("documents_%s", k)
				flattener.Fields[newKey] = v
				delete(flattener.Fields, k)
			}
			accumulator.AddFields("logstash_plugins", flattener.Fields, pluginTags)
		}
	}

	return nil
}

func (logstash *Logstash) gatherQueueStats(
	queue *PipelineQueue,
	tags map[string]string,
	accumulator telegraf.Accumulator,
) error {
	queueTags := map[string]string{
		"queue_type": queue.Type,
	}
	for tag, value := range tags {
		queueTags[tag] = value
	}

	events := queue.Events
	if queue.EventsCount != nil {
		events = *queue.EventsCount
	}

	queueFields := map[string]interface{}{
		"events": events,
	}

	if queue.Type != "memory" {
		flattener := jsonParser.JSONFlattener{}
		err := flattener.FlattenJSON("", queue.Capacity)
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

		if queue.MaxQueueSizeInBytes != nil {
			queueFields["max_queue_size_in_bytes"] = *queue.MaxQueueSizeInBytes
		}

		if queue.QueueSizeInBytes != nil {
			queueFields["queue_size_in_bytes"] = *queue.QueueSizeInBytes
		}
	}

	accumulator.AddFields("logstash_queue", queueFields, queueTags)

	return nil
}

// gatherJVMStats gather the Pipeline metrics and add results to the accumulator (for Logstash < 6)
func (logstash *Logstash) gatherPipelineStats(address string, accumulator telegraf.Accumulator) error {
	pipelineStats := &PipelineStats{}

	err := logstash.gatherJSONData(address, pipelineStats)
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
func (logstash *Logstash) gatherPipelinesStats(address string, accumulator telegraf.Accumulator) error {
	pipelinesStats := &PipelinesStats{}

	err := logstash.gatherJSONData(address, pipelinesStats)
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
		client, err := logstash.createHTTPClient()

		if err != nil {
			return err
		}
		logstash.client = client
	}

	if choice.Contains("jvm", logstash.Collect) {
		jvmURL, err := url.Parse(logstash.URL + jvmStats)
		if err != nil {
			return err
		}
		if err := logstash.gatherJVMStats(jvmURL.String(), accumulator); err != nil {
			return err
		}
	}

	if choice.Contains("process", logstash.Collect) {
		processURL, err := url.Parse(logstash.URL + processStats)
		if err != nil {
			return err
		}
		if err := logstash.gatherProcessStats(processURL.String(), accumulator); err != nil {
			return err
		}
	}

	if choice.Contains("pipelines", logstash.Collect) {
		if logstash.SinglePipeline {
			pipelineURL, err := url.Parse(logstash.URL + pipelineStats)
			if err != nil {
				return err
			}
			if err := logstash.gatherPipelineStats(pipelineURL.String(), accumulator); err != nil {
				return err
			}
		} else {
			pipelinesURL, err := url.Parse(logstash.URL + pipelinesStats)
			if err != nil {
				return err
			}
			if err := logstash.gatherPipelinesStats(pipelinesURL.String(), accumulator); err != nil {
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
