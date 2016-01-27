package amon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf"
)

type Amon struct {
	ServerKey    string
	AmonInstance string
	Timeout      internal.Duration

	client *http.Client
}

var sampleConfig = `
  # Amon Server Key
  server_key = "my-server-key" # required.

  # Amon Instance URL
  amon_instance = "https://youramoninstance" # required

  # Connection timeout.
  # timeout = "5s"
`

type TimeSeries struct {
	Series []*Metric `json:"series"`
}

type Metric struct {
	Metric string   `json:"metric"`
	Points [1]Point `json:"points"`
}

type Point [2]float64

func (a *Amon) Connect() error {
	if a.ServerKey == "" || a.AmonInstance == "" {
		return fmt.Errorf("serverkey and amon_instance are required fields for amon output")
	}
	a.client = &http.Client{
		Timeout: a.Timeout.Duration,
	}
	return nil
}

func (a *Amon) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}
	ts := TimeSeries{}
	tempSeries := []*Metric{}
	metricCounter := 0

	for _, pt := range points {
		mname := strings.Replace(pt.Name(), "_", ".", -1)
		if amonPts, err := buildPoints(pt); err == nil {
			for fieldName, amonPt := range amonPts {
				metric := &Metric{
					Metric: mname + "_" + strings.Replace(fieldName, "_", ".", -1),
				}
				metric.Points[0] = amonPt
				tempSeries = append(tempSeries, metric)
				metricCounter++
			}
		} else {
			log.Printf("unable to build Metric for %s, skipping\n", pt.Name())
		}
	}

	ts.Series = make([]*Metric, metricCounter)
	copy(ts.Series, tempSeries[0:])
	tsBytes, err := json.Marshal(ts)
	if err != nil {
		return fmt.Errorf("unable to marshal TimeSeries, %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", a.authenticatedUrl(), bytes.NewBuffer(tsBytes))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s\n", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s\n", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d\n", resp.StatusCode)
	}

	return nil
}

func (a *Amon) SampleConfig() string {
	return sampleConfig
}

func (a *Amon) Description() string {
	return "Configuration for Amon Server to send metrics to."
}

func (a *Amon) authenticatedUrl() string {

	return fmt.Sprintf("%s/api/system/%s", a.AmonInstance, a.ServerKey)
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

func (a *Amon) Close() error {
	return nil
}

func init() {
	outputs.Add("amon", func() telegraf.Output {
		return &Amon{}
	})
}
