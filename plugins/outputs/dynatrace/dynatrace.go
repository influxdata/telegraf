//go:generate ../../../tools/readme_config_includer/generator
package dynatrace

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	dynatrace_metric "github.com/dynatrace-oss/dynatrace-metric-utils-go/metric"
	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/apiconstants"
	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/dimensions"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

// Dynatrace Configuration for the Dynatrace output plugin
type Dynatrace struct {
	URL                       string          `toml:"url"`
	APIToken                  config.Secret   `toml:"api_token"`
	Prefix                    string          `toml:"prefix"`
	Log                       telegraf.Logger `toml:"-"`
	Timeout                   config.Duration `toml:"timeout"`
	AddCounterMetrics         []string        `toml:"additional_counters"`
	AddCounterMetricsPatterns []string        `toml:"additional_counters_patterns"`

	DefaultDimensions map[string]string `toml:"default_dimensions"`

	normalizedDefaultDimensions dimensions.NormalizedDimensionList
	normalizedStaticDimensions  dimensions.NormalizedDimensionList

	tls.ClientConfig

	client *http.Client

	loggedMetrics map[string]bool // New empty set
}

func (*Dynatrace) SampleConfig() string {
	return sampleConfig
}

// Connect Connects the Dynatrace output plugin to the Telegraf stream
func (*Dynatrace) Connect() error {
	return nil
}

// Close Closes the Dynatrace output plugin
func (d *Dynatrace) Close() error {
	d.client = nil
	return nil
}

func (d *Dynatrace) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	lines := make([]string, 0, len(metrics))
	for _, tm := range metrics {
		dims := make([]dimensions.Dimension, 0, len(tm.TagList()))
		for _, tag := range tm.TagList() {
			// Ignore special tags for histogram and summary types.
			switch tm.Type() {
			case telegraf.Histogram:
				if tag.Key == "le" || tag.Key == "gt" {
					continue
				}
			case telegraf.Summary:
				if tag.Key == "quantile" {
					continue
				}
			}
			dims = append(dims, dimensions.NewDimension(tag.Key, tag.Value))
		}

		for _, field := range tm.FieldList() {
			metricName := tm.Name() + "." + field.Key

			typeOpt := d.getTypeOption(tm, field)

			if typeOpt == nil {
				// Unsupported type. Log only once per unsupported metric name
				if !d.loggedMetrics[metricName] {
					d.Log.Warnf("Unsupported type for %s", metricName)
					d.loggedMetrics[metricName] = true
				}
				continue
			}

			name := tm.Name() + "." + field.Key
			dm, err := dynatrace_metric.NewMetric(
				name,
				dynatrace_metric.WithPrefix(d.Prefix),
				dynatrace_metric.WithDimensions(
					dimensions.MergeLists(
						d.normalizedDefaultDimensions,
						dimensions.NewNormalizedDimensionList(dims...),
						d.normalizedStaticDimensions,
					),
				),
				dynatrace_metric.WithTimestamp(tm.Time()),
				typeOpt,
			)

			if err != nil {
				d.Log.Warn(fmt.Sprintf("failed to normalize metric: %s - %s", name, err.Error()))
				continue
			}

			line, err := dm.Serialize()

			if err != nil {
				d.Log.Warn(fmt.Sprintf("failed to serialize metric: %s - %s", name, err.Error()))
				continue
			}

			lines = append(lines, line)
		}
	}

	limit := apiconstants.GetPayloadLinesLimit()
	for i := 0; i < len(lines); i += limit {
		batch := lines[i:min(i+limit, len(lines))]

		output := strings.Join(batch, "\n")
		if output != "" {
			if err := d.send(output); err != nil {
				return fmt.Errorf("error processing data: %w", err)
			}
		}
	}

	return nil
}

func (d *Dynatrace) send(msg string) error {
	var err error
	req, err := http.NewRequest("POST", d.URL, bytes.NewBufferString(msg))
	if err != nil {
		d.Log.Errorf("Dynatrace error: %s", err.Error())
		return fmt.Errorf("error while creating HTTP request: %w", err)
	}
	req.Header.Add("Content-Type", "text/plain; charset=UTF-8")

	if !d.APIToken.Empty() {
		token, err := d.APIToken.Get()
		if err != nil {
			return fmt.Errorf("getting token failed: %w", err)
		}
		req.Header.Add("Authorization", "Api-Token "+token.String())
		token.Destroy()
	}
	// add user-agent header to identify metric source
	req.Header.Add("User-Agent", "telegraf")

	resp, err := d.client.Do(req)
	if err != nil {
		d.Log.Errorf("Dynatrace error: %s", err.Error())
		return fmt.Errorf("error while sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("request failed with response code: %d", resp.StatusCode)
	}

	// print metric line results as info log
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		d.Log.Errorf("Dynatrace error reading response")
	}
	bodyString := string(bodyBytes)
	d.Log.Debugf("Dynatrace returned: %s", bodyString)

	return nil
}

func (d *Dynatrace) Init() error {
	if len(d.URL) == 0 {
		d.Log.Infof("Dynatrace URL is empty, defaulting to OneAgent metrics interface")
		d.URL = apiconstants.GetDefaultOneAgentEndpoint()
	}
	if d.URL != apiconstants.GetDefaultOneAgentEndpoint() && d.APIToken.Empty() {
		d.Log.Errorf("Dynatrace api_token is a required field for Dynatrace output")
		return errors.New("api_token is a required field for Dynatrace output")
	}

	tlsCfg, err := d.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	d.client = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(d.Timeout),
	}

	dims := make([]dimensions.Dimension, 0, len(d.DefaultDimensions))
	for key, value := range d.DefaultDimensions {
		dims = append(dims, dimensions.NewDimension(key, value))
	}
	d.normalizedDefaultDimensions = dimensions.NewNormalizedDimensionList(dims...)
	d.normalizedStaticDimensions = dimensions.NewNormalizedDimensionList(dimensions.NewDimension("dt.metrics.source", "telegraf"))
	d.loggedMetrics = make(map[string]bool)

	return nil
}

func init() {
	outputs.Add("dynatrace", func() telegraf.Output {
		return &Dynatrace{
			Timeout: config.Duration(time.Second * 5),
		}
	})
}

func (d *Dynatrace) getTypeOption(metric telegraf.Metric, field *telegraf.Field) dynatrace_metric.MetricOption {
	metricName := metric.Name() + "." + field.Key
	if isCounterMetricsMatch(d.AddCounterMetrics, metricName) ||
		isCounterMetricsPatternsMatch(d.AddCounterMetricsPatterns, metricName) {
		switch v := field.Value.(type) {
		case float64:
			return dynatrace_metric.WithFloatCounterValueDelta(v)
		case uint64:
			return dynatrace_metric.WithIntCounterValueDelta(int64(v))
		case int64:
			return dynatrace_metric.WithIntCounterValueDelta(v)
		default:
			return nil
		}
	}
	switch v := field.Value.(type) {
	case float64:
		return dynatrace_metric.WithFloatGaugeValue(v)
	case uint64:
		return dynatrace_metric.WithIntGaugeValue(int64(v))
	case int64:
		return dynatrace_metric.WithIntGaugeValue(v)
	case bool:
		if v {
			return dynatrace_metric.WithIntGaugeValue(1)
		}
		return dynatrace_metric.WithIntGaugeValue(0)
	}

	return nil
}

func isCounterMetricsMatch(counterMetrics []string, metricName string) bool {
	for _, i := range counterMetrics {
		if i == metricName {
			return true
		}
	}
	return false
}

func isCounterMetricsPatternsMatch(counterPatterns []string, metricName string) bool {
	for _, pattern := range counterPatterns {
		regex, err := regexp.Compile(pattern)
		if err == nil && regex.MatchString(metricName) {
			return true
		}
	}
	return false
}
