package librato_with_tags

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
)

// Librato structure for configuration and client
type LibratoWithTags struct {
	APIUser  string `toml:"api_user"`
	APIToken string `toml:"api_token"`
	Debug    bool
	Timeout  internal.Duration
	Prefix   string

	APIUrl string
	client *http.Client
}

// https://www.librato.com/docs/kb/faq/best_practices/naming_convention_metrics_sources.html#naming-limitations-for-sources-and-metrics
var reUnacceptedChar = regexp.MustCompile("[^.a-zA-Z0-9_-]")

var sampleConfig = `
  ## Librato API Docs
  ## http://dev.librato.com/v1/metrics-authentication
  ## Librato API user
  api_user = "telegraf@influxdb.com" # required.
  ## Librato API token
  api_token = "my-secret-token" # required.
  ## Debug
  # debug = false
  ## Connection timeout.
  # timeout = "5s"
  ## Metrics prefix, used for the metric/measurement name prefix
  prefix = "telegraf"
`

// Ligrato API endpoint
// See docs here: https://www.librato.com/docs/api/#create-a-measurement
const libratoAPI = "https://metrics-api.librato.com/v1/measurements"

// LMeasurement is the default struct for Librato APIs toplevel JSON format
type LMeasurements struct {
	Measurements []*Measurement `json:"measurements"`
}

// Measurement is one item in the list of measurements that can be sent in one request
type Measurement struct {
	Name        string            `json:"name"`
	Value       float64           `json:"value"`
	Tags        map[string]string `json:"tags"`
	MeasureTime int64             `json:"time"`
}

// NewLibratoWithTags is the main constructor for librato output plugin
func NewLibratoWithTags(apiURL string) *LibratoWithTags {
	return &LibratoWithTags{
		APIUrl: apiURL,
		Prefix: "telegraf",
	}
}

// Connect is the default output plugin connection function who make sure it
// can connect to the endpoint
func (l *LibratoWithTags) Connect() error {
	if l.APIUser == "" || l.APIToken == "" {
		return fmt.Errorf(
			"api_user and api_token are required fields for librato output")
	}
	l.client = &http.Client{
		Timeout: l.Timeout.Duration,
	}
	return nil
}

func (l *LibratoWithTags) Write(metrics []telegraf.Metric) error {

	if len(metrics) == 0 {
		return nil
	}

	tempMeasurements := []*Measurement{}

	for _, m := range metrics {
		if measurements, err := l.buildMeasurements(m); err == nil {
			for _, measurement := range measurements {
				tempMeasurements = append(tempMeasurements, measurement)
				log.Printf("D! Got a measurement: %v\n", measurement)

			}
		} else {
			log.Printf("I! unable to build Measurement for %s, skipping\n", m.Name())
			log.Printf("D! Couldn't build Measurement: %v\n", err)

		}
	}

	measurementCounter := len(tempMeasurements)
	// make sure we send a batch of maximum 500
	sizeBatch := 500
	for start := 0; start < measurementCounter; start += sizeBatch {
		lmeasurements := LMeasurements{}
		end := start + sizeBatch
		if end > measurementCounter {
			end = measurementCounter
			sizeBatch = end - start
		}
		lmeasurements.Measurements = make([]*Measurement, sizeBatch)
		copy(lmeasurements.Measurements, tempMeasurements[start:end])
		metricsBytes, err := json.Marshal(lmeasurements)
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

		ValidStatusCode := map[int]bool{
			200: true,
			202: true,
		}

		if !ValidStatusCode[resp.StatusCode] || l.Debug {
			htmlData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("D! Couldn't get response! (%v)\n", err)
			}
			if !ValidStatusCode[resp.StatusCode] {
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
func (l *LibratoWithTags) SampleConfig() string {
	return sampleConfig
}

// Description is function who return the Description of this output
func (l *LibratoWithTags) Description() string {
	return "Configuration for Librato API to send metrics to."
}

func (l *LibratoWithTags) buildMeasurements(m telegraf.Metric) ([]*Measurement, error) {

	measurements := []*Measurement{}

	for fieldName, value := range m.Fields() {

		if l.Prefix == "" {
			l.Prefix = "telegraf"
		}

		// prepare metric name:
		metricName := m.Name()
		if fieldName != "value" {
			metricName = fmt.Sprintf("%s.%s.%s", l.Prefix, m.Name(), fieldName)
		} else {
			metricName = fmt.Sprintf("%s.%s", l.Prefix, m.Name())
		}

		measurement := &Measurement{
			Name: reUnacceptedChar.ReplaceAllString(metricName, "-"),
			// Value: setting it below
			Tags: m.Tags(),
			// MeasureTime: m.Time().Unix(),
		}
		if !verifyValue(value) {
			continue
		}
		if err := measurement.setValue(value); err != nil {
			return measurements, fmt.Errorf(
				"unable to extract value from Fields, %s\n",
				err.Error())
		}
		measurements = append(measurements, measurement)
	}

	log.Printf("D! Built measurements: %v\n", measurements)
	return measurements, nil
}

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	}
	return true
}

func (g *Measurement) setValue(v interface{}) error {
	switch d := v.(type) {
	case int:
		g.Value = float64(int(d))
	case int32:
		g.Value = float64(int32(d))
	case int64:
		g.Value = float64(int64(d))
	case uint64:
		g.Value = float64(uint64(d))
	case float32:
		g.Value = float64(d)
	case float64:
		g.Value = float64(d)
	default:
		return fmt.Errorf("undeterminable type %+v", d)
	}
	return nil
}

//Close is used to close the connection to librato Output
func (l *LibratoWithTags) Close() error {
	return nil
}

func init() {
	outputs.Add("librato_with_tags", func() telegraf.Output {
		return NewLibratoWithTags(libratoAPI)
	})
}
