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
  ## Datadog API key
  apikey = "my-secret-key" # required.

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
			for fieldName, dogM := range dogMs {
				// name of the datadog measurement
				var dname string
				if fieldName == "value" {
					// adding .value seems redundant here
					dname = m.Name()
				} else {
					dname = m.Name() + "." + fieldName
				}
				var host string
				host, _ = m.Tags()["host"]
				metric := &Metric{
					Metric: dname,
					Tags:   buildTags(m.Tags()),
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
	return fmt.Sprintf("%s?%s", d.apiUrl, q.Encode())
}

func buildMetrics(m telegraf.Metric) (map[string]Point, error) {
	ms := make(map[string]Point)
	for k, v := range m.Fields() {
		if !verifyValue(v) {
			continue
		}
		var p Point
		if err := p.setValue(v); err != nil {
			return ms, fmt.Errorf("unable to extract value from Fields %v error %v", k, err.Error())
		}
		p[0] = float64(m.Time().Unix())
		ms[k] = p
	}
	return ms, nil
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

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	}
	return true
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
	case bool:
		p[1] = float64(0)
		if d {
			p[1] = float64(1)
		}
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
