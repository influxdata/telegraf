package librato

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
)

// Librato structure for configuration and client
type Librato struct {
	APIUser   string          `toml:"api_user"`
	APIToken  string          `toml:"api_token"`
	Debug     bool            `toml:"debug"`
	SourceTag string          `toml:"source_tag" deprecated:"1.0.0;use 'template' instead"`
	Timeout   config.Duration `toml:"timeout"`
	Template  string          `toml:"template"`
	Log       telegraf.Logger `toml:"-"`

	APIUrl string
	client *http.Client
}

// https://www.librato.com/docs/kb/faq/best_practices/naming_convention_metrics_sources.html#naming-limitations-for-sources-and-metrics
var reUnacceptedChar = regexp.MustCompile("[^.a-zA-Z0-9_-]")

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
		Timeout: time.Duration(l.Timeout),
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
				l.Log.Debugf("Got a gauge: %v", gauge)
			}
		} else {
			l.Log.Infof("Unable to build Gauge for %s, skipping", m.Name())
			l.Log.Debugf("Couldn't build gauge: %v", err)
		}
	}

	metricCounter := len(tempGauges)
	// make sur we send a batch of maximum 300
	sizeBatch := 300
	for start := 0; start < metricCounter; start += sizeBatch {
		err := l.writeBatch(start, sizeBatch, metricCounter, tempGauges)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Librato) writeBatch(start int, sizeBatch int, metricCounter int, tempGauges []*Gauge) error {
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
		return fmt.Errorf("unable to marshal Metrics, %s", err.Error())
	}

	l.Log.Debugf("Librato request: %v", string(metricsBytes))

	req, err := http.NewRequest(
		"POST",
		l.APIUrl,
		bytes.NewBuffer(metricsBytes))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(l.APIUser, l.APIToken)

	resp, err := l.client.Do(req)
	if err != nil {
		l.Log.Debugf("Error POSTing metrics: %v", err.Error())
		return fmt.Errorf("error POSTing metrics, %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 || l.Debug {
		htmlData, err := io.ReadAll(resp.Body)
		if err != nil {
			l.Log.Debugf("Couldn't get response! (%v)", err)
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf(
				"received bad status code, %d\n %s",
				resp.StatusCode,
				string(htmlData))
		}
		l.Log.Debugf("Librato response: %v", string(htmlData))
	}
	return nil
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
			fmt.Errorf("undeterminable Source type from Field, %s", l.Template)
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
			return gauges, fmt.Errorf("unable to extract value from Fields, %s", err.Error())
		}
		gauges = append(gauges, gauge)
	}

	l.Log.Debugf("Built gauges: %v", gauges)
	return gauges, nil
}

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	default:
		return true
	}
}

func (g *Gauge) setValue(v interface{}) error {
	switch d := v.(type) {
	case int64:
		g.Value = float64(d)
	case uint64:
		g.Value = float64(d)
	case float64:
		g.Value = d
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
