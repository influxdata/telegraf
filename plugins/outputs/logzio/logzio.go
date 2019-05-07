package logzio

import (
	. "bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf/internal"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultLogzioRequestTimeout = time.Second * 5
	defaultLogzioURL            = "https://listener.logz.io:8071"

	logzioDescription        = "Send aggregate metrics to Logz.io"
	logzioType               = "telegraf"
	logzioMaxRequestBodySize = 9 * 1024 * 1024 // 9MB
)

var sampleConfig = `
  ## Logz.io account token
  token = "your logz.io token" # required

  ## Use your listener URL for your Logz.io account region.
  # url = "https://listener.logz.io:8071"

  ## Timeout for HTTP requests
  # timeout = "5s"
`

type Logzio struct {
	Token   string            `toml:"token"`
	URL     string            `toml:"url"`
	Timeout internal.Duration `toml:"timeout"`

	client *http.Client
}

// Connect to the Output
func (l *Logzio) Connect() error {
	log.Printf("D! [logzio] Connecting to logz.io output...\n")
	if l.Token == "" || l.Token == "your logz.io token" {
		return fmt.Errorf("[logzio] token is required")
	}

	if l.URL == "" {
		l.URL = defaultLogzioURL
	}

	if l.Timeout.Duration == 0 {
		l.Timeout.Duration = defaultLogzioRequestTimeout
	}

	tlsConfig := &tls.Config{}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	l.client = &http.Client{
		Transport: transport,
		Timeout:   l.Timeout.Duration,
	}

	log.Printf("I! [logzio] Successfuly created Logz.io sender: %s\n", l.URL)
	return nil
}

// Close any connections to the Output
func (l *Logzio) Close() error {
	log.Printf("D! [logzio] Closing logz.io output\n")
	l.client.CloseIdleConnections()
	return nil
}

// Description returns a one-sentence description on the Output
func (l *Logzio) Description() string {
	return logzioDescription
}

// SampleConfig returns the default configuration of the Output
func (l *Logzio) SampleConfig() string {
	return sampleConfig
}

// Write takes in group of points to be written to the Output
func (l *Logzio) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	log.Printf("D! [logzio] Recived %d metrics\n", len(metrics))
	var body []byte
	for _, metric := range metrics {
		var name = metric.Name()
		m := make(map[string]interface{})

		m["@timestamp"] = metric.Time()
		m["measurement_name"] = name
		if len(metric.Tags()) != 0 {
			m["telegraf_tags"] = metric.Tags()
		}
		m["value_type"] = metric.Type()
		m["type"] = logzioType
		m[name] = metric.Fields()

		serialized, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("E! [logzio] Failed to marshal: %+v\n", m)
		}
		// Logz.io maximum request body size of 10MB. Send bulks that
		// exceed this size (with safety buffer) via separate write requests.
		if (len(body) + len(serialized) + 1) > logzioMaxRequestBodySize {
			err := l.sendBulk(body)
			if err != nil {
				return err
			}
			body = nil
		}
		body = append(body, serialized...)
		body = append(body, '\n')
	}

	return l.sendBulk(body)
}

func (l *Logzio) sendBulk(body []byte) error {
	if len(body) == 0 {
		return nil
	}

	var buf Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(body); err != nil {
		return err
	}
	if err := g.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", l.URL, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to write batch: [%v] %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func init() {
	outputs.Add("logzio", func() telegraf.Output {
		return &Logzio{}
	})
}
