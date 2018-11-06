package newrelic

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Newrelic struct {
	URL        string            `toml:"url"`
	Timeout    internal.Duration `toml:"timeout"`
	Method     string            `toml:"method"`
	Headers    map[string]string `toml:"headers"`
	client     *http.Client
	serializer serializers.Serializer
}

const (
	defaultClientTimeout = 30 * time.Second
	defaultContentType   = "application/json"
	defaultMethod        = http.MethodPost
)

var sampleConfig = `
  ## Newrelic Insights API URL
  url = "https://insights-collector.newrelic.com/v1/accounts/xxxxxxx/events"

  ## Additional HTTP headers
    [outputs.newrelic.headers]
    X-Insert-Key = "New Relic Insert Key"
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

func (nr *Newrelic) Connect() error {
	if nr.URL == "" {
		return fmt.Errorf("URL is a required field for newrelic output")
	}

	nr.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: nr.Timeout.Duration,
	}
	return nil
}

func (nr *Newrelic) Write(metrics []telegraf.Metric) error {
	reqBody, err := nr.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}

	if err := nr.write(reqBody); err != nil {
		return err
	}

	return nil
}

func (nr *Newrelic) write(reqBody []byte) error {
	req, err := http.NewRequest(nr.Method, nr.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", defaultContentType)
	for k, v := range nr.Headers {
		req.Header.Set(k, v)
	}

	resp, err := nr.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to [%s] received status code: %d", nr.URL, resp.StatusCode)
	}

	return nil
}

func (nr *Newrelic) SampleConfig() string {
	return sampleConfig
}

func (nr *Newrelic) Description() string {
	return "Configuration for NewRelic Insights API to send metrics to."
}

func (nr *Newrelic) Close() error {
	return nil
}

func (nr *Newrelic) SetSerializer(serializer serializers.Serializer) {
	nr.serializer = serializer
}

func init() {
	outputs.Add("newrelic", func() telegraf.Output {
		return &Newrelic{
			Timeout: internal.Duration{Duration: defaultClientTimeout},
			Method:  defaultMethod,
		}
	})
}
