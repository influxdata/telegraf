package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Datadog struct {
	Apikey      string          `toml:"apikey"`
	Timeout     config.Duration `toml:"timeout"`
	URL         string          `toml:"url"`
	Compression string          `toml:"compression"`
	Log         telegraf.Logger `toml:"-"`

	client *http.Client
	proxy.HTTPProxy
}

type TimeSeries struct {
	Series []*Metric `json:"series"`
}

type Metric struct {
	Metric   string   `json:"metric"`
	Points   [1]Point `json:"points"`
	Host     string   `json:"host"`
	Type     string   `json:"type,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Interval int64    `json:"interval"`
}

type Point [2]float64

const datadogAPI = "https://app.datadoghq.com/api/v1/series"

func (d *Datadog) Connect() error {
	if d.Apikey == "" {
		return fmt.Errorf("apikey is a required field for datadog output")
	}

	proxyFunc, err := d.Proxy()
	if err != nil {
		return err
	}

	d.client = &http.Client{
		Transport: &http.Transport{
			Proxy: proxyFunc,
		},
		Timeout: time.Duration(d.Timeout),
	}
	return nil
}

func (d *Datadog) Write(metrics []telegraf.Metric) error {
	ts := TimeSeries{}
	tempSeries := []*Metric{}
	metricCounter := 0

	for _, m := range metrics {
		if dogMs, err := buildMetrics(m); err == nil {
			metricTags := buildTags(m.TagList())
			host, _ := m.GetTag("host")

			if len(dogMs) == 0 {
				continue
			}

			for fieldName, dogM := range dogMs {
				// name of the datadog measurement
				var dname string
				if fieldName == "value" {
					// adding .value seems redundant here
					dname = m.Name()
				} else {
					dname = m.Name() + "." + fieldName
				}
				var tname string
				switch m.Type() {
				case telegraf.Counter:
					tname = "count"
				case telegraf.Gauge:
					tname = "gauge"
				default:
					tname = ""
				}
				metric := &Metric{
					Metric:   dname,
					Tags:     metricTags,
					Host:     host,
					Type:     tname,
					Interval: 1,
				}
				metric.Points[0] = dogM
				tempSeries = append(tempSeries, metric)
				metricCounter++
			}
		} else {
			d.Log.Infof("Unable to build Metric for %s due to error '%v', skipping", m.Name(), err)
		}
	}

	if len(tempSeries) == 0 {
		return nil
	}

	redactedAPIKey := "****************"
	ts.Series = make([]*Metric, metricCounter)
	copy(ts.Series, tempSeries[0:])
	tsBytes, err := json.Marshal(ts)
	if err != nil {
		return fmt.Errorf("unable to marshal TimeSeries, %s", err.Error())
	}

	var req *http.Request
	c := strings.ToLower(d.Compression)
	switch c {
	case "zlib":
		encoder, err := internal.NewContentEncoder(c)
		if err != nil {
			return err
		}
		buf, err := encoder.Encode(tsBytes)
		if err != nil {
			return err
		}
		req, err = http.NewRequest("POST", d.authenticatedURL(), bytes.NewBuffer(buf))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Encoding", "deflate")
	case "none":
		fallthrough
	default:
		req, err = http.NewRequest("POST", d.authenticatedURL(), bytes.NewBuffer(tsBytes))
	}

	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s", strings.Replace(err.Error(), d.Apikey, redactedAPIKey, -1))
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s", strings.Replace(err.Error(), d.Apikey, redactedAPIKey, -1))
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d", resp.StatusCode)
	}

	return nil
}

func (d *Datadog) authenticatedURL() string {
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
		index++
	}
	return tags
}

func verifyValue(v interface{}) bool {
	switch v := v.(type) {
	case string:
		return false
	case float64:
		// The payload will be encoded as JSON, which does not allow NaN or Inf.
		return !math.IsNaN(v) && !math.IsInf(v, 0)
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
		p[1] = d
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
			URL:         datadogAPI,
			Compression: "none",
		}
	})
}
