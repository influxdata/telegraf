package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Datadog struct {
	Apikey  string
	Timeout internal.Duration

	URL    string `toml:"url"`
	client *http.Client
}

var sampleConfig = `
  ## Datadog API key
  apikey = "my-secret-key" # required.

  # The base endpoint URL can optionally be specified but it defaults to:
  #url = "https://app.datadoghq.com/api/v1/series"

  ## Connection timeout.
  # timeout = "5s"
`

type TimeSeries struct {
	Series []*Metric `json:"series"`
}

type Metric struct {
	Metric string   `json:"metric"`
	Points [1]Point `json:"points"`
	Host   string   `json:"host"`
	Tags   []string `json:"tags,omitempty"`
}

type Point [2]float64

const datadog_api = "https://app.datadoghq.com/api/v1/series"

func (d *Datadog) Connect() error {
	if d.Apikey == "" {
		return fmt.Errorf("apikey is a required field for datadog output")
	}

	d.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: d.Timeout.Duration,
	}
	return nil
}

func (d *Datadog) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	ts := TimeSeries{}
	tempSeries := []*Metric{}
	metricCounter := 0

	for _, m := range metrics {
		if dogMs, err := buildMetrics(m); err == nil {
			metricTags := buildTags(m.TagList())
			host, _ := m.GetTag("host")

			for fieldName, dogM := range dogMs {
				// name of the datadog measurement
				var dname string
				if fieldName == "value" {
					// adding .value seems redundant here
					dname = m.Name()
				} else {
					dname = m.Name() + "." + fieldName
				}
				metric := &Metric{
					Metric: dname,
					Tags:   metricTags,
					Host:   host,
				}
				metric.Points[0] = dogM
				tempSeries = append(tempSeries, metric)
				metricCounter++
			}
		} else {
			log.Printf("I! unable to build Metric for %s due to error '%v', skipping\n", m.Name(), err)
		}
	}

	redactedApiKey := "****************"
	ts.Series = make([]*Metric, metricCounter)
	copy(ts.Series, tempSeries[0:])
	tsBytes, err := json.Marshal(ts)
	if err != nil {
		return fmt.Errorf("unable to marshal TimeSeries, %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", d.authenticatedUrl(), bytes.NewBuffer(tsBytes))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s\n", strings.Replace(err.Error(), d.Apikey, redactedApiKey, -1))
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s\n", strings.Replace(err.Error(), d.Apikey, redactedApiKey, -1))
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
	return fmt.Sprintf("%s?%s", d.URL, q.Encode())
}

func buildMetrics(m telegraf.Metric) (map[string]Point, error) {
	ms := make(map[string]Point)
	for _, field := range m.FieldList() {
		if !verifyValue(field.Value) {
			continue
		}
		var p Point
		if err := p.setValue(field.Value); err != nil {
			return ms, fmt.Errorf("unable to extract value from Fields %v error %v", field.Key, err.Error())
		}
		p[0] = float64(m.Time().Unix())
		ms[field.Key] = p
	}
	return ms, nil
}

func buildTags(tagList []*telegraf.Tag) []string {
	tags := make([]string, len(tagList))
	index := 0
	for _, tag := range tagList {
		tags[index] = fmt.Sprintf("%s:%s", tag.Key, tag.Value)
		index += 1
	}
	return tags
}

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	}
	return true
}

func (p *Point) setValue(v interface{}) error {
	switch d := v.(type) {
	case int64:
		p[1] = float64(d)
	case uint64:
		p[1] = float64(d)
	case float64:
		p[1] = float64(d)
	case bool:
		p[1] = float64(0)
		if d {
			p[1] = float64(1)
		}
	default:
		return fmt.Errorf("undeterminable field type: %T", v)
	}
	return nil
}

func (d *Datadog) Close() error {
	return nil
}

func init() {
	outputs.Add("datadog", func() telegraf.Output {
		return &Datadog{
			URL: datadog_api,
		}
	})
}
