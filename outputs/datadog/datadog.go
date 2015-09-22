package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/influxdb/influxdb/client"
	t "github.com/koksan83/telegraf"
	"github.com/koksan83/telegraf/outputs"
)

type Datadog struct {
	Apikey  string
	Timeout t.Duration

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
	Metric string   `json:"metric"`
	Points [1]Point `json:"points"`
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
		Timeout: d.Timeout.Duration,
	}
	return nil
}

func (d *Datadog) Write(bp client.BatchPoints) error {
	if len(bp.Points) == 0 {
		return nil
	}
	ts := TimeSeries{
		Series: make([]*Metric, len(bp.Points)),
	}
	for index, pt := range bp.Points {
		metric := &Metric{
			Metric: pt.Measurement,
			Tags:   buildTags(bp.Tags, pt.Tags),
		}
		if p, err := buildPoint(bp, pt); err == nil {
			metric.Points[0] = p
		}
		ts.Series[index] = metric
	}
	tsBytes, err := json.Marshal(ts)
	if err != nil {
		return fmt.Errorf("unable to marshal TimeSeries, %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", d.authenticatedUrl(), bytes.NewBuffer(tsBytes))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s\n", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s\n", err.Error())
	}

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

func buildTags(bpTags map[string]string, ptTags map[string]string) []string {
	tags := make([]string, (len(bpTags) + len(ptTags)))
	index := 0
	for k, v := range bpTags {
		tags[index] = fmt.Sprintf("%s:%s", k, v)
		index += 1
	}
	for k, v := range ptTags {
		tags[index] = fmt.Sprintf("%s:%s", k, v)
		index += 1
	}
	sort.Strings(tags)
	return tags
}

func buildPoint(bp client.BatchPoints, pt client.Point) (Point, error) {
	var p Point
	if err := p.setValue(pt.Fields["value"]); err != nil {
		return p, fmt.Errorf("unable to extract value from Fields, %s", err.Error())
	}
	if pt.Time.IsZero() {
		p[0] = float64(bp.Time.Unix())
	} else {
		p[0] = float64(pt.Time.Unix())
	}
	return p, nil
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
	outputs.Add("datadog", func() outputs.Output {
		return NewDatadog(datadog_api)
	})
}
