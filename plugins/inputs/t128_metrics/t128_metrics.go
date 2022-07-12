package t128_metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// DefaultRequestTimeout is the request timeout if none is configured
	DefaultRequestTimeout = config.Duration(time.Second * 5)
	// DefaultMaxSimultaneousRequests is the maximum simultaneous requests if none is configured
	DefaultMaxSimultaneousRequests = 20
)

// T128Metrics is an input for metrics of a 128T router instance
type T128Metrics struct {
	BaseURL                 string             `toml:"base_url"`
	UnixSocket              string             `toml:"unix_socket"`
	ConfiguredMetrics       []ConfiguredMetric `toml:"metric"`
	Timeout                 config.Duration    `toml:"timeout"`
	MaxSimultaneousRequests int                `toml:"max_simultaneous_requests"`
	UseIntegerConversion    bool               `toml:"use_integer_conversion"`
	UseBulkRetrieval        bool               `toml:"use_bulk_retrieval"`

	client    *http.Client
	limiter   *requestLimiter
	retriever Retriever
}

// ConfiguredMetric represents a single configured metric element
type ConfiguredMetric struct {
	Name       string              `toml:"name"`
	Fields     map[string]string   `toml:"fields"`
	Parameters map[string][]string `toml:"parameters"`
}

var sampleConfig = `
## Read metrics from a 128T instance
[[inputs.t128_metrics]]
## Required. The base url for metrics collection
# base_url = "http://localhost:31517/api/v1/router/Fabric128/"

## A socket to use for retrieving metrics - unused by default
# unix_socket = "/var/run/128technology/web-server.sock"

## Whether or not to use the bulk retrieval API. If used, the base_url should
## point directly to the bulk metrics endpoint.
# use_bulk_retrieval = false

## The maximum number of requests to be in flight at once
# max_simultaneous_requests = 20

## Whether to attempt conversion of values to integer before conversion to float
# use_integer_conversion = false

## Amount of time allowed to complete a single HTTP request
# timeout = "5s"

## The metrics to collect
# [[inputs.t128_metrics.metric]]
# name = "cpu"
#
# [inputs.t128_metrics.metric.fields]
## Refer to the 128T REST swagger documentation for the list of available metrics
#     key_name = "stats/<path_to_metric>"
#     utilization = "stats/cpu/utilization"
#
## [inputs.t128_metrics.metric.parameters]
#     parameter_name = ["value1", "value2"]
#     core = ["1", "2"]
`

// SampleConfig returns the default configuration of the Input
func (*T128Metrics) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the Input
func (*T128Metrics) Description() string {
	return "Read metrics from a 128T router instance"
}

// Init sets up the input to be ready for action
func (plugin *T128Metrics) Init() error {
	if plugin.BaseURL == "" {
		return fmt.Errorf("base_url is a required configuration field")
	}

	if !plugin.UseBulkRetrieval && plugin.BaseURL[len(plugin.BaseURL)-1:] != "/" {
		plugin.BaseURL += "/"
	}

	if plugin.MaxSimultaneousRequests <= 0 {
		return fmt.Errorf("max_simultaneous_requests must be greater than 0")
	}

	transport := http.DefaultTransport

	if plugin.UnixSocket != "" {
		transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", plugin.UnixSocket)
			},
		}
	}

	plugin.client = &http.Client{Transport: transport, Timeout: time.Duration(plugin.Timeout)}
	plugin.limiter = newRequestLimiter(plugin.MaxSimultaneousRequests)

	if plugin.UseBulkRetrieval {
		var err error
		plugin.retriever, err = NewBulkRetriever(plugin.UseIntegerConversion, plugin.ConfiguredMetrics)
		if err != nil {
			return fmt.Errorf("failed to create retriever: %w", err)
		}
	} else {
		plugin.retriever = NewIndividualRetriever(plugin.UseIntegerConversion, plugin.ConfiguredMetrics)
	}

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval"
func (plugin *T128Metrics) Gather(acc telegraf.Accumulator) error {
	timestamp := time.Now().Round(time.Second)

	var wg sync.WaitGroup
	wg.Add(plugin.retriever.RequestCount())

	for index := 0; index < plugin.retriever.RequestCount(); index++ {
		go func(idx int) {
			plugin.retrieveMetrics(idx, acc, timestamp)
			wg.Done()
		}(index)
	}

	wg.Wait()

	return nil
}

func (plugin *T128Metrics) retrieveMetrics(index int, acc telegraf.Accumulator, timestamp time.Time) {
	request, err := plugin.retriever.CreateRequest(index, plugin.BaseURL)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to create a request for %s: %w", plugin.retriever.Describe(index), err))
		return
	}

	plugin.limiter.wait()
	response, err := plugin.client.Do(request)
	plugin.limiter.done()

	if err != nil {
		acc.AddError(fmt.Errorf("failed to retrieve %s: %w", plugin.retriever.Describe(index), err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, err := ioutil.ReadAll(response.Body)
		if err != nil {
			message = []byte("")
		}

		acc.AddError(fmt.Errorf("status code %d not OK for %s: %s", response.StatusCode, plugin.retriever.Describe(index), message))
		return
	}

	var responseMetrics []ResponseMetric
	if err := json.NewDecoder(response.Body).Decode(&responseMetrics); err != nil {
		acc.AddError(fmt.Errorf("failed to decode response for %s: %w", plugin.retriever.Describe(index), err))
		return
	}

	plugin.retriever.PopulateResponse(index, acc, responseMetrics, timestamp)
}

type requestLimiter struct {
	synchronizer chan struct{}
}

func newRequestLimiter(limit int) *requestLimiter {
	limiter := &requestLimiter{make(chan struct{}, limit)}

	for i := 0; i < limit; i++ {
		limiter.done()
	}

	return limiter
}

func (l *requestLimiter) wait() {
	<-l.synchronizer
}

func (l *requestLimiter) done() {
	l.synchronizer <- struct{}{}
}

func init() {
	inputs.Add("t128_metrics", func() telegraf.Input {
		return &T128Metrics{
			Timeout:                 config.Duration(DefaultRequestTimeout),
			MaxSimultaneousRequests: DefaultMaxSimultaneousRequests,
			UseIntegerConversion:    false,
			UseBulkRetrieval:        false,
		}
	})
}
