package librato

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Librato struct {
	ApiUser   string
	ApiToken  string
	SourceTag string
	Timeout   internal.Duration

	apiUrl string
	client *http.Client
}

var sampleConfig = `
  # Librator API Docs
  # http://dev.librato.com/v1/metrics-authentication

  # Librato API user
  api_user = "telegraf@influxdb.com" # required.

  # Librato API token
  api_token = "my-secret-token" # required.

  # Tag Field to populate source attribute (optional)
  # This is typically the _hostname_ from which the metric was obtained.
  source_tag = "hostname"

  # Connection timeout.
  # timeout = "5s"
`

type LMetrics struct {
	Gauges []*Gauge `json:"gauges"`
}

type Gauge struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Source      string  `json:"source"`
	MeasureTime int64   `json:"measure_time"`
}

const librato_api = "https://metrics-api.librato.com/v1/metrics"

func NewLibrato(apiUrl string) *Librato {
	return &Librato{
		apiUrl: apiUrl,
	}
}

func (l *Librato) Connect() error {
	if l.ApiUser == "" || l.ApiToken == "" {
		return fmt.Errorf("api_user and api_token are required fields for librato output")
	}
	l.client = &http.Client{
		Timeout: l.Timeout.Duration,
	}
	return nil
}

func (l *Librato) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	lmetrics := LMetrics{}
	tempGauges := []*Gauge{}
	metricCounter := 0

	for _, m := range metrics {
		if gauges, err := l.buildGauges(m); err == nil {
			for _, gauge := range gauges {
				tempGauges = append(tempGauges, gauge)
				metricCounter++
			}
		} else {
			log.Printf("unable to build Gauge for %s, skipping\n", m.Name())
		}
	}

	lmetrics.Gauges = make([]*Gauge, metricCounter)
	copy(lmetrics.Gauges, tempGauges[0:])
	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("unable to marshal Metrics, %s\n", err.Error())
	}
	req, err := http.NewRequest("POST", l.apiUrl, bytes.NewBuffer(metricsBytes))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s\n", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(l.ApiUser, l.ApiToken)

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s\n", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("received bad status code, %d\n", resp.StatusCode)
	}

	return nil
}

func (l *Librato) SampleConfig() string {
	return sampleConfig
}

func (l *Librato) Description() string {
	return "Configuration for Librato API to send metrics to."
}

func (l *Librato) buildGauges(m telegraf.Metric) ([]*Gauge, error) {
	gauges := []*Gauge{}
	for fieldName, value := range m.Fields() {
		gauge := &Gauge{
			Name:        m.Name() + "_" + fieldName,
			MeasureTime: m.Time().Unix(),
		}
		if err := gauge.setValue(value); err != nil {
			return gauges, fmt.Errorf("unable to extract value from Fields, %s\n",
				err.Error())
		}
		if l.SourceTag != "" {
			if source, ok := m.Tags()[l.SourceTag]; ok {
				gauge.Source = source
			} else {
				return gauges,
					fmt.Errorf("undeterminable Source type from Field, %s\n",
						l.SourceTag)
			}
		}
	}
	return gauges, nil
}

func (g *Gauge) setValue(v interface{}) error {
	switch d := v.(type) {
	case int:
		g.Value = float64(int(d))
	case int32:
		g.Value = float64(int32(d))
	case int64:
		g.Value = float64(int64(d))
	case float32:
		g.Value = float64(d)
	case float64:
		g.Value = float64(d)
	default:
		return fmt.Errorf("undeterminable type %+v", d)
	}
	return nil
}

func (l *Librato) Close() error {
	return nil
}

func init() {
	outputs.Add("librato", func() telegraf.Output {
		return NewLibrato(librato_api)
	})
}
