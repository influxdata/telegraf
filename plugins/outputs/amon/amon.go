package amon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Amon struct {
	ServerKey    string          `toml:"server_key"`
	AmonInstance string          `toml:"amon_instance"`
	Timeout      config.Duration `toml:"timeout"`
	Log          telegraf.Logger `toml:"-"`

	client *http.Client
}

type TimeSeries struct {
	Series []*Metric `json:"series"`
}

type Metric struct {
	Metric string   `json:"metric"`
	Points [1]Point `json:"metrics"`
}

type Point [2]float64

func (a *Amon) Connect() error {
	if a.ServerKey == "" || a.AmonInstance == "" {
		return fmt.Errorf("serverkey and amon_instance are required fields for amon output")
	}
	a.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: time.Duration(a.Timeout),
	}
	return nil
}

func (a *Amon) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	ts := TimeSeries{}
	tempSeries := []*Metric{}
	metricCounter := 0

	for _, m := range metrics {
		mname := strings.Replace(m.Name(), "_", ".", -1)
		if amonPts, err := buildMetrics(m); err == nil {
			for fieldName, amonPt := range amonPts {
				metric := &Metric{
					Metric: mname + "_" + strings.Replace(fieldName, "_", ".", -1),
				}
				metric.Points[0] = amonPt
				tempSeries = append(tempSeries, metric)
				metricCounter++
			}
		} else {
			a.Log.Infof("Unable to build Metric for %s, skipping", m.Name())
		}
	}

	ts.Series = make([]*Metric, metricCounter)
	copy(ts.Series, tempSeries[0:])
	tsBytes, err := json.Marshal(ts)
	if err != nil {
		return fmt.Errorf("unable to marshal TimeSeries, %s", err.Error())
	}
	req, err := http.NewRequest("POST", a.authenticatedURL(), bytes.NewBuffer(tsBytes))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d", resp.StatusCode)
	}

	return nil
}

func (a *Amon) authenticatedURL() string {
	return fmt.Sprintf("%s/api/system/%s", a.AmonInstance, a.ServerKey)
}

func buildMetrics(m telegraf.Metric) (map[string]Point, error) {
	ms := make(map[string]Point)
	for k, v := range m.Fields() {
		var p Point
		if err := p.setValue(v); err != nil {
			return ms, fmt.Errorf("unable to extract value from Fields, %s", err.Error())
		}
		p[0] = float64(m.Time().Unix())
		ms[k] = p
	}
	return ms, nil
}

func (p *Point) setValue(v interface{}) error {
	switch d := v.(type) {
	case int:
		p[1] = float64(d)
	case int32:
		p[1] = float64(d)
	case int64:
		p[1] = float64(d)
	case float32:
		p[1] = float64(d)
	case float64:
		p[1] = d
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
