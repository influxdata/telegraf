//go:generate ../../../tools/readme_config_includer/generator
package logzio

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultLogzioURL = "https://listener.logz.io:8071"
	logzioType       = "telegraf"
)

type Logzio struct {
	URL     string          `toml:"url"`
	Token   config.Secret   `toml:"token"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`

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

func (*Logzio) SampleConfig() string {
	return sampleConfig
}

// Connect to the Output
func (l *Logzio) Connect() error {
	l.Log.Debug("Connecting to logz.io output...")

	if l.Token.Empty() {
		return fmt.Errorf("token is required")
	}
	if equal, err := l.Token.EqualTo([]byte("your logz.io token")); err != nil {
		return err
	} else if equal {
		return fmt.Errorf("please replace 'token' with your actual token")
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
			return fmt.Errorf("unable to marshal metric: %w", err)
		}

		_, err = gz.Write(append(serialized, '\n'))
		if err != nil {
			return fmt.Errorf("unable to write gzip meric: %w", err)
		}
	}

	err := gz.Close()
	if err != nil {
		return fmt.Errorf("unable to close gzip: %w", err)
	}

	return l.send(buff.Bytes())
}

func (l *Logzio) send(metrics []byte) error {
	url, err := l.authURL()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(metrics))
	if err != nil {
		return fmt.Errorf("unable to create http.Request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d", resp.StatusCode)
	}

	return nil
}

func (l *Logzio) authURL() (string, error) {
	token, err := l.Token.Get()
	if err != nil {
		return "", fmt.Errorf("getting token failed: %w", err)
	}
	defer config.ReleaseSecret(token)

	return fmt.Sprintf("%s/?token=%s", l.URL, string(token)), nil
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
