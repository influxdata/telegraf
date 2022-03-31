package logzio

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultLogzioURL = "https://listener.logz.io:8071"

	logzioDescription = "Send aggregate metrics to Logz.io"
	logzioType        = "telegraf"
)

type Logzio struct {
	Log     telegraf.Logger `toml:"-"`
	Timeout config.Duration `toml:"timeout"`
	Token   string          `toml:"token"`
	URL     string          `toml:"url"`

	tls.ClientConfig
	client *http.Client
}

type TimeSeries struct {
	Series []*Metric
}

type Metric struct {
	Metric     map[string]interface{} `json:"metrics"`
	Dimensions map[string]string      `json:"dimensions"`
	Time       time.Time              `json:"@timestamp"`
	Type       string                 `json:"type"`
}

// Connect to the Output
func (l *Logzio) Connect() error {
	l.Log.Debug("Connecting to logz.io output...")

	if l.Token == "" || l.Token == "your logz.io token" {
		return fmt.Errorf("token is required")
	}

	tlsCfg, err := l.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	l.client = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(l.Timeout),
	}

	return nil
}

// Close any connections to the Output
func (l *Logzio) Close() error {
	l.Log.Debug("Closing logz.io output")
	return nil
}

// Write takes in group of points to be written to the Output
func (l *Logzio) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	for _, metric := range metrics {
		m := l.parseMetric(metric)

		serialized, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("unable to marshal metric, %s", err.Error())
		}

		_, err = gz.Write(append(serialized, '\n'))
		if err != nil {
			return fmt.Errorf("unable to write gzip meric, %s", err.Error())
		}
	}

	err := gz.Close()
	if err != nil {
		return fmt.Errorf("unable to close gzip, %s", err.Error())
	}

	return l.send(buff.Bytes())
}

func (l *Logzio) send(metrics []byte) error {
	req, err := http.NewRequest("POST", l.authURL(), bytes.NewBuffer(metrics))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d", resp.StatusCode)
	}

	return nil
}

func (l *Logzio) authURL() string {
	return fmt.Sprintf("%s/?token=%s", l.URL, l.Token)
}

func (l *Logzio) parseMetric(metric telegraf.Metric) *Metric {
	return &Metric{
		Metric: map[string]interface{}{
			metric.Name(): metric.Fields(),
		},
		Dimensions: metric.Tags(),
		Time:       metric.Time(),
		Type:       logzioType,
	}
}

func init() {
	outputs.Add("logzio", func() telegraf.Output {
		return &Logzio{
			URL:     defaultLogzioURL,
			Timeout: config.Duration(time.Second * 5),
		}
	})
}
