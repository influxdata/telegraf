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

	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/internal"
	"github.com/influxdb/telegraf/outputs"
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
		Timeout: d.Timeout.Duration,
	}
	return nil
}

func (d *Datadog) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}
	ts := TimeSeries{}
	var tempSeries = make([]*Metric, len(points))
	var acceptablePoints = 0

	for _, pt := range points {
		mname := strings.Replace(pt.Name(), "_", ".", -1)
		if amonPts, err := buildPoints(pt); err == nil {
			for fieldName, amonPt := range amonPts {
				metric := &Metric{
					Metric: mname + strings.Replace(fieldName, "_", ".", -1),
				}
				metric.Points[0] = amonPt
				tempSeries[acceptablePoints] = metric
				acceptablePoints += 1
			}
		} else {
			log.Printf("unable to build Metric for %s, skipping\n", pt.Name())
		}
	}

	ts.Series = make([]*Metric, acceptablePoints)
	copy(ts.Series, tempSeries[0:])
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

func buildPoints(pt *client.Point) (map[string]Point, error) {
	pts := make(map[string]Point)
	for k, v := range pt.Fields() {
		var p Point
		if err := p.setValue(v); err != nil {
			return pts, fmt.Errorf("unable to extract value from Fields, %s", err.Error())
		}
		p[0] = float64(pt.Time().Unix())
		pts[k] = p
	}
	return pts, nil
}

func buildTags(ptTags map[string]string) []string {
	tags := make([]string, len(ptTags))
	index := 0
	for k, v := range ptTags {
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
	outputs.Add("datadog", func() outputs.Output {
		return NewDatadog(datadog_api)
	})
}
