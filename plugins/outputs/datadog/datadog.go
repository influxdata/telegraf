package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Datadog struct {
	Apikey  string
	Timeout internal.Duration

	apiUrl string
	client *http.Client
}

var sampleConfig = `
  # Datadog API key
  apikey = "my-secret-key" # required.

  # Connection timeout.
  # timeout = "5s"
`

type TimeSeries struct {
	Series []*Metric `json:"series"`
}

type Metric struct {
	Metric     string   `json:"metric"`
	Points     [1]Point `json:"points"`
	Host       string   `json:"host"`
	Tags       []string `json:"tags,omitempty"`
	Interval   float64  `json:"interval,omitempty"`
	Type       string   `json:"type"`
	DeviceName string   `json:"device_name"`
}

type Point [2]float64

const datadog_api = "https://app.datadoghq.com/api/v1/series"

func NewDatadog(apiUrl string) *Datadog {
	return &Datadog{
		apiUrl: apiUrl,
	}
}

func (d *Datadog) Connect() error {
	if d.Apikey == "" {
		return fmt.Errorf("apikey is a required field for datadog output")
	}
	d.client = &http.Client{
		Timeout: d.Timeout.Duration,
	}
	return nil
}

func (d *Datadog) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	ts := TimeSeries{}

	for _, m := range metrics {
		if dogMs, err := buildMetrics(m); err == nil {
			ts.Series = append(ts.Series, dogMs...)
		} else {
			log.Printf("unable to build Metric for %s, skipping\n", m.Name())
		}
	}

	tsBytes, err := json.Marshal(ts)
	if err != nil {
		return fmt.Errorf("unable to marshal TimeSeries, %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", d.authenticatedUrl(), bytes.NewBuffer(tsBytes))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s\n", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	// If we don't pretend to be dogstatsd, it ignores our carefully
	// calculated type.
	req.Header.Set("DD-Dogstatsd-Version", "5.6.3")
	req.Header.Set("User-Agent", "python-requests/2.6.0 CPython/2.7.10")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s\n", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d\n", resp.StatusCode)
	}

	return nil
}

func (d *Datadog) SampleConfig() string {
	return sampleConfig
}

func (d *Datadog) Description() string {
	return "Configuration for DataDog API to send metrics to."
}

func (d *Datadog) authenticatedUrl() string {
	q := url.Values{
		"api_key": []string{d.Apikey},
	}
	return fmt.Sprintf("%s?%s", d.apiUrl, q.Encode())
}

// Convert a telegraf metric to datadog metrics;
// we need a separate metric for each field.
// (also has magic for statsd field names)
func buildMetrics(m telegraf.Metric) ([]*Metric, error) {
	var datadogMetrics []*Metric
	tags := m.Tags()
	metricType := tags["metric_type"]
	baseDatadogMetric := Metric{
		Metric:   m.Name(),
		Tags:     buildTags(tags),
		Host:     tags["host"],
		Interval: float64(m.Interval()) / float64(time.Second),
	}

	for field, value := range m.Fields() {
		metric := baseDatadogMetric
		if field != "value" {
			metric.Metric += "_" + field
		}
		metric.Metric = strings.Replace(metric.Metric, "_", ".", -1)
		p := &metric.Points[0]
		p[0] = float64(m.Time().Unix())
		if err := p.setValue(value); err != nil {
			return nil, fmt.Errorf("unable to extract value from Fields, %s", err.Error())
		}

		if metricType == "counter" ||
			((metricType == "histogram" || metricType == "timing") && field == "count") {
			metric.Type = "rate"
			p[1] /= baseDatadogMetric.Interval
		} else {
			metric.Type = "gauge"
		}

		datadogMetrics = append(datadogMetrics, &metric)
	}

	return datadogMetrics, nil
}

func buildTags(mTags map[string]string) []string {
	tags := make([]string, len(mTags))
	index := 0
	for k, v := range mTags {
		tags[index] = fmt.Sprintf("%s:%s", k, v)
		index += 1
	}
	sort.Strings(tags)
	return tags
}

func (p *Point) setValue(v interface{}) error {
	switch d := v.(type) {
	case int:
		p[1] = float64(int(d))
	case int32:
		p[1] = float64(int32(d))
	case int64:
		p[1] = float64(int64(d))
	case float32:
		p[1] = float64(d)
	case float64:
		p[1] = float64(d)
	default:
		return fmt.Errorf("undeterminable type")
	}
	return nil
}

func (d *Datadog) Close() error {
	return nil
}

func init() {
	outputs.Add("datadog", func() telegraf.Output {
		return NewDatadog(datadog_api)
	})
}
