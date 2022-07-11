package t128_metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strconv"
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

	client  *http.Client
	limiter *requestLimiter
	metrics []RequestMetric
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

## The maximum number of requests to be in flight at once
# max_simultaneous_requests = 20

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

	if plugin.BaseURL[len(plugin.BaseURL)-1:] != "/" {
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
	plugin.metrics = configuredMetricsToRequestMetrics(plugin.ConfiguredMetrics)

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval"
func (plugin *T128Metrics) Gather(acc telegraf.Accumulator) error {
	timestamp := time.Now().Round(time.Second)

	var wg sync.WaitGroup
	wg.Add(len(plugin.metrics))

	for _, requestMetric := range plugin.metrics {
		go func(metric RequestMetric) {
			plugin.retrieveMetric(metric, acc, timestamp)
			wg.Done()
		}(requestMetric)
	}

	wg.Wait()

	return nil
}

func (plugin *T128Metrics) retrieveMetric(metric RequestMetric, acc telegraf.Accumulator, timestamp time.Time) {
	request, err := plugin.createRequest(plugin.BaseURL, metric)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to create a request for metric %s: %w", metric.ID, err))
		return
	}

	plugin.limiter.wait()
	response, err := plugin.client.Do(request)
	plugin.limiter.done()

	if err != nil {
		acc.AddError(fmt.Errorf("failed to retrieve metric %s: %w", metric.ID, err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, err := ioutil.ReadAll(response.Body)
		if err != nil {
			message = []byte("")
		}

		acc.AddError(fmt.Errorf("status code %d not OK for metric %s: %s", response.StatusCode, metric.ID, message))
		return
	}

	var responseMetrics []ResponseMetric
	if err := json.NewDecoder(response.Body).Decode(&responseMetrics); err != nil {
		acc.AddError(fmt.Errorf("failed to decode response for metric %s: %w", metric.ID, err))
		return
	}

	for _, responseMetric := range responseMetrics {
		for _, permutation := range responseMetric.Permutations {
			if permutation.Value == nil {
				continue
			}

			tags := make(map[string]string)
			for _, parameter := range permutation.Parameters {
				tags[parameter.Name] = parameter.Value
			}

			acc.AddFields(
				metric.OutMeasurement,
				map[string]interface{}{metric.OutField: tryNumericConversion(*permutation.Value)},
				tags,
				timestamp)
		}
	}
}

func (plugin *T128Metrics) createRequest(baseURL string, metric RequestMetric) (*http.Request, error) {
	content := struct {
		Parameters []RequestParameter `json:"parameters,omitempty"`
	}{
		metric.Parameters,
	}

	body, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to create request body for metric '%s': %w", metric.ID, err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", baseURL, metric.ID), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for metric '%s': %w", metric.ID, err)
	}

	request.Header.Add("Content-Type", "application/json")

	return request, nil
}

func configuredMetricsToRequestMetrics(configuredMetrics []ConfiguredMetric) []RequestMetric {
	requestMetrics := make([]RequestMetric, 0)

	for _, configMetric := range configuredMetrics {
		for fieldName, fieldPath := range configMetric.Fields {
			// Sort names for consistency in testing. It's not free, but only happens during startup.
			parameterNames := make([]string, 0, len(configMetric.Parameters))
			for parameterName := range configMetric.Parameters {
				parameterNames = append(parameterNames, parameterName)
			}
			sort.Strings(parameterNames)

			parameters := make([]RequestParameter, 0, len(configMetric.Parameters))
			for _, parameterName := range parameterNames {
				values := configMetric.Parameters[parameterName]
				parameters = append(parameters, RequestParameter{
					Name:    parameterName,
					Values:  values,
					Itemize: true,
				})
			}

			requestMetrics = append(requestMetrics, RequestMetric{
				ID:             fieldPath,
				Parameters:     parameters,
				OutMeasurement: configMetric.Name,
				OutField:       fieldName,
			})
		}
	}

	return requestMetrics
}

func tryNumericConversion(value string) interface{} {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	} else if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	} else {
		return value
	}
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
		}
	})
}
