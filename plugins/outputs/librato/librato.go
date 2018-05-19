package librato

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
)

// Librato structure for configuration and client
type Librato struct {
	APIUser   string `toml:"api_user"`
	APIToken  string `toml:"api_token"`
	Debug     bool
	SourceTag string // Deprecated, keeping for backward-compatibility
	Timeout   internal.Duration
	Template  string

	APIUrl string
	client *http.Client
}

// https://www.librato.com/docs/kb/faq/best_practices/naming_convention_metrics_sources.html#naming-limitations-for-sources-and-metrics
var reUnacceptedChar = regexp.MustCompile("[^.a-zA-Z0-9_-]")

var sampleConfig = `
  ## Librator API Docs
  ## http://dev.librato.com/v1/metrics-authentication
  ## Librato API user
  api_user = "telegraf@influxdb.com" # required.
  ## Librato API token
  api_token = "my-secret-token" # required.
  ## Debug
  # debug = false
  ## Connection timeout.
  # timeout = "5s"
  ## Output source Template (same as graphite buckets)
  ## see https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#graphite
  ## This template is used in librato's source (not metric's name)
  template = "host"

`

// LMetrics is the default struct for Librato's API fromat
type LMetrics struct {
	Gauges []*Gauge `json:"gauges"`
}

// Gauge is the gauge format for Librato's API fromat
type Gauge struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Source      string  `json:"source"`
	MeasureTime int64   `json:"measure_time"`
}

const libratoAPI = "https://metrics-api.librato.com/v1/metrics"

// NewLibrato is the main constructor for librato output plugins
func NewLibrato(apiURL string) *Librato {
	return &Librato{
		APIUrl:   apiURL,
		Template: "host",
	}
}

// Connect is the default output plugin connection function who make sure it
// can connect to the endpoint
func (l *Librato) Connect() error {
	if l.APIUser == "" || l.APIToken == "" {
		return fmt.Errorf(
			"api_user and api_token are required fields for librato output")
	}
	l.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: l.Timeout.Duration,
	}
	return nil
}

func (l *Librato) Write(metrics []telegraf.Metric) error {

	if len(metrics) == 0 {
		return nil
	}
	if l.Template == "" {
		l.Template = "host"
	}
	if l.SourceTag != "" {
		l.Template = l.SourceTag
	}

	tempGauges := []*Gauge{}

	for _, m := range metrics {
		if gauges, err := l.buildGauges(m); err == nil {
			for _, gauge := range gauges {
				tempGauges = append(tempGauges, gauge)
				log.Printf("D! Got a gauge: %v\n", gauge)
			}
		} else {
			log.Printf("I! unable to build Gauge for %s, skipping\n", m.Name())
			log.Printf("D! Couldn't build gauge: %v\n", err)

		}
	}

	metricCounter := len(tempGauges)
	// make sur we send a batch of maximum 300
	sizeBatch := 300
	for start := 0; start < metricCounter; start += sizeBatch {
		lmetrics := LMetrics{}
		end := start + sizeBatch
		if end > metricCounter {
			end = metricCounter
			sizeBatch = end - start
		}
		lmetrics.Gauges = make([]*Gauge, sizeBatch)
		copy(lmetrics.Gauges, tempGauges[start:end])
		metricsBytes, err := json.Marshal(lmetrics)
		if err != nil {
			return fmt.Errorf("unable to marshal Metrics, %s\n", err.Error())
		}

		log.Printf("D! Librato request: %v\n", string(metricsBytes))

		req, err := http.NewRequest(
			"POST",
			l.APIUrl,
			bytes.NewBuffer(metricsBytes))
		if err != nil {
			return fmt.Errorf(
				"unable to create http.Request, %s\n",
				err.Error())
		}
		req.Header.Add("Content-Type", "application/json")
		req.SetBasicAuth(l.APIUser, l.APIToken)

		resp, err := l.client.Do(req)
		if err != nil {
			log.Printf("D! Error POSTing metrics: %v\n", err.Error())
			return fmt.Errorf("error POSTing metrics, %s\n", err.Error())
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 || l.Debug {
			htmlData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("D! Couldn't get response! (%v)\n", err)
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf(
					"received bad status code, %d\n %s",
					resp.StatusCode,
					string(htmlData))
			}
			log.Printf("D! Librato response: %v\n", string(htmlData))
		}
	}

	return nil
}

// SampleConfig is function who return the default configuration for this
// output
func (l *Librato) SampleConfig() string {
	return sampleConfig
}

// Description is function who return the Description of this output
func (l *Librato) Description() string {
	return "Configuration for Librato API to send metrics to."
}

func (l *Librato) buildGauges(m telegraf.Metric) ([]*Gauge, error) {

	gauges := []*Gauge{}
	if m.Time().Unix() == 0 {
		return gauges, fmt.Errorf("time was zero %s", m.Name())
	}
	metricSource := graphite.InsertField(
		graphite.SerializeBucketName("", m.Tags(), l.Template, ""),
		"value")
	if metricSource == "" {
		return gauges,
			fmt.Errorf("undeterminable Source type from Field, %s\n",
				l.Template)
	}
	for fieldName, value := range m.Fields() {

		metricName := m.Name()
		if fieldName != "value" {
			metricName = fmt.Sprintf("%s.%s", m.Name(), fieldName)
		}

		gauge := &Gauge{
			Source:      reUnacceptedChar.ReplaceAllString(metricSource, "-"),
			Name:        reUnacceptedChar.ReplaceAllString(metricName, "-"),
			MeasureTime: m.Time().Unix(),
		}
		if !verifyValue(value) {
			continue
		}
		if err := gauge.setValue(value); err != nil {
			return gauges, fmt.Errorf(
				"unable to extract value from Fields, %s\n",
				err.Error())
		}
		gauges = append(gauges, gauge)
	}

	log.Printf("D! Built gauges: %v\n", gauges)
	return gauges, nil
}

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	}
	return true
}

func (g *Gauge) setValue(v interface{}) error {
	switch d := v.(type) {
	case int64:
		g.Value = float64(int64(d))
	case uint64:
		g.Value = float64(d)
	case float64:
		g.Value = float64(d)
	case bool:
		if d {
			g.Value = float64(1.0)
		} else {
			g.Value = float64(0.0)
		}
	default:
		return fmt.Errorf("undeterminable type %+v", d)
	}
	return nil
}

//Close is used to close the connection to librato Output
func (l *Librato) Close() error {
	return nil
}

func init() {
	outputs.Add("librato", func() telegraf.Output {
		return NewLibrato(libratoAPI)
	})
}
