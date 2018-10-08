package seq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"net/http"
	"os"
)

type Seq struct {
	SeqInstance string
	SeqApiKey   string
	Timeout     internal.Duration

	client *http.Client
}

var sampleConfig = `
  ## Seq Instance URL
  seq_instance = "https://localhost:5341" # required

  ## Seq API Key
  seq_api_key = "MYAPIKEY"

  ## Connection timeout.
  # timeout = "5s"
`

func (a *Seq) Connect() error {
	if a.SeqInstance == "" {
		return fmt.Errorf("seq_instance is a required field for seq output")
	}
	a.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: a.Timeout.Duration,
	}
	return nil
}

func serialize(m telegraf.Metric) ([]byte, error) {
	se := make(map[string]interface{})
	se["Name"] = m.Name()
	se["@mt"] = "Telegraf Measurement {Name} on {Host}"
	se["@t"] = m.Time().Format("2006-01-02T15:04:05Z07:00")

	if host, ok := m.Tags()["host"]; ok {
		se["Host"] = host
	} else {
		host, err := os.Hostname()
		if err != nil {
			return []byte{}, err
		}
		se["Host"] = host
	}

	se["Tags"] = m.Tags()
	se["Fields"] = m.Fields()

	serialized, err := json.Marshal(se)
	if err != nil {
		return []byte{}, err
	}

	return serialized, nil
}

func (a *Seq) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	var buffer bytes.Buffer

	for _, metric := range metrics {

		line, err := serialize(metric)
		if err != nil {
			return err
		}

		buffer.Write(line)
		buffer.WriteString("\n")
	}

	req, err := http.NewRequest("POST", a.authenticatedUrl(), &buffer)

	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s\n", err.Error())
	}

	req.Header.Add("Content-Type", "application/vnd.serilog.clef")

	if a.SeqApiKey != "" {
		req.Header.Add("X-Seq-ApiKey", a.SeqApiKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s\n", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d\n", resp.StatusCode)
	}

	return nil
}

func (a *Seq) SampleConfig() string {
	return sampleConfig
}

func (a *Seq) Description() string {
	return "Configuration for Seq output target."
}

func (a *Seq) authenticatedUrl() string {
	return fmt.Sprintf("%s/api/events/raw", a.SeqInstance)
}

func (a *Seq) Close() error {
	return nil
}

func init() {
	outputs.Add("seq", func() telegraf.Output {
		return &Seq{}
	})
}
