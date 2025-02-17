//go:generate ../../../tools/readme_config_includer/generator
package datadog

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

//go:embed sample.conf
var sampleConfig string

type Datadog struct {
	Apikey       string          `toml:"apikey"`
	Timeout      config.Duration `toml:"timeout"`
	URL          string          `toml:"url"`
	Compression  string          `toml:"compression"`
	RateInterval config.Duration `toml:"rate_interval"`
	Log          telegraf.Logger `toml:"-"`

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

func (*Datadog) SampleConfig() string {
	return sampleConfig
}

func (d *Datadog) Connect() error {
	if d.Apikey == "" {
		return errors.New("apikey is a required field for datadog output")
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

func (d *Datadog) convertToDatadogMetric(metrics []telegraf.Metric) []*Metric {
	tempSeries := make([]*Metric, 0, len(metrics))
	for _, m := range metrics {
		if dogMs, err := buildMetrics(m); err == nil {
			metricTags := buildTags(m.TagList())
			host, _ := m.GetTag("host")
			// Retrieve the metric_type tag created by inputs.statsd
			statsDMetricType, _ := m.GetTag("metric_type")

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
				var interval int64
				interval = 1
				switch m.Type() {
				case telegraf.Counter, telegraf.Untyped:
					if d.RateInterval > 0 && isRateable(statsDMetricType, fieldName) {
						// interval is expected to be in seconds
						rateIntervalSeconds := time.Duration(d.RateInterval).Seconds()
						interval = int64(rateIntervalSeconds)
						dogM[1] = dogM[1] / rateIntervalSeconds
						tname = "rate"
					} else if m.Type() == telegraf.Counter {
						tname = "count"
					} else {
						tname = ""
					}
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
					Interval: interval,
				}
				metric.Points[0] = dogM
				tempSeries = append(tempSeries, metric)
			}
		} else {
			d.Log.Infof("Unable to build Metric for %s due to error '%v', skipping", m.Name(), err)
		}
	}
	return tempSeries
}

func (d *Datadog) Write(metrics []telegraf.Metric) error {
	ts := TimeSeries{}
	tempSeries := d.convertToDatadogMetric(metrics)

	if len(tempSeries) == 0 {
		return nil
	}

	redactedAPIKey := "****************"
	ts.Series = make([]*Metric, len(tempSeries))
	copy(ts.Series, tempSeries[0:])
	tsBytes, err := json.Marshal(ts)
	if err != nil {
		return fmt.Errorf("unable to marshal TimeSeries: %w", err)
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
		return fmt.Errorf("unable to create http.Request, %s", strings.ReplaceAll(err.Error(), d.Apikey, redactedAPIKey))
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s", strings.ReplaceAll(err.Error(), d.Apikey, redactedAPIKey))
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		//nolint:errcheck // err can be ignored since it is just for logging
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received bad status code, %d: %s", resp.StatusCode, string(body))
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
			return ms, fmt.Errorf("unable to extract value from Field %v: %w", field.Key, err)
		}
		p[0] = float64(m.Time().Unix())
		ms[field.Key] = p
	}
	return ms, nil
}

func buildTags(tagList []*telegraf.Tag) []string {
	tags := make([]string, 0, len(tagList))
	for _, tag := range tagList {
		tags = append(tags, fmt.Sprintf("%s:%s", tag.Key, tag.Value))
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

func isRateable(statsDMetricType, fieldName string) bool {
	switch statsDMetricType {
	case
		"counter":
		return true
	case
		"timing",
		"histogram":
		return fieldName == "count"
	default:
		return false
	}
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

func (*Datadog) Close() error {
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
