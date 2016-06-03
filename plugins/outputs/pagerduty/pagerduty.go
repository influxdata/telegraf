package pagerduty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"go/token"
	"go/types"
	"log"
	"net/http"
)

const EventEndPoint = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

type Event struct {
	Type        string        `json:"event_type"`
	ServiceKey  string        `json:"service_key"`
	Description string        `json:"description,omitempty"`
	Client      string        `json:"client,omitempty"`
	ClientURL   string        `json:"client_url,omitempty"`
	Details     interface{}   `json:"details,omitempty"`
	Contexts    []interface{} `json:"contexts,omitempty"`
}

type PD struct {
	ServiceKey string            `toml:"service_key"`
	Desc       string            `toml:"description"`
	Metric     string            `toml:"metric"`
	Field      string            `toml:"field"`
	Expression string            `toml:"expression"`
	Tags       map[string]string `toml:"tags"`
}

var sampleConfig = `
## PagerDuty service key
service_key = <SERVICE KEY>
## Metric name that will be checked
metric = "cpu"
## Description of the check
description = "Check CPU"
## Name of the metric field which will be used to check
field = "time_iowait"
## Expression is used to evaluate the alert
expression = "> 50.0"
`

func createEvent(e Event) (*http.Response, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", EventEndPoint, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("HTTP Status Code: %d", resp.StatusCode)
	}
	return resp, nil
}

func (p *PD) Connect() error {
	return nil
}

func (p *PD) Close() error {
	return nil
}

func (p *PD) Match(metric telegraf.Metric) bool {
	if p.Metric != metric.Name() {
		log.Printf("Metric name is not matched. Expected: '%s' Found: '%s'", p.Metric, metric.Name())
		return false
	}
	for k, v := range p.Tags {
		t, ok := metric.Tags()[k]
		if !ok {
			log.Printf("Tag value absent. Tag name: '%s'", k)
			return false
		}
		if t != v {
			log.Printf("Tag '%s' value not matched. Expected: '%s' Found: '%s'", k, v, t)
			return false
		}
	}
	field, ok := metric.Fields()[p.Field]
	if !ok {
		log.Printf("Filed '%s' absent", p.Field)
		return false
	}
	expr := fmt.Sprintf("%v %s", field, p.Expression)
	fs := token.NewFileSet()
	tv, err := types.Eval(fs, nil, token.NoPos, expr)
	if err != nil {
		log.Printf("Error in parsing expression. Message:%s", err)
		return false
	}
	return tv.Value.String() == "true"
}

func (p *PD) SampleConfig() string {
	return sampleConfig
}

func (p *PD) Description() string {
	return "Output metrics as PagerDuty event"
}

func (p *PD) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	event := Event{
		Type:        "trigger",
		ServiceKey:  p.ServiceKey,
		Description: p.Desc,
		Client:      "telegraf",
	}
	for _, metric := range metrics {
		if !p.Match(metric) {
			log.Println("Metric is not matched by threshold, skipping")
			continue
		}
		m := make(map[string]interface{})
		m["tags"] = metric.Tags()
		m["fields"] = metric.Fields()
		m["name"] = metric.Name()
		m["timestamp"] = metric.UnixNano() / 1000000000
		event.Details = m
		_, err := createEvent(event)
		if err != nil {
			return err
		}
		log.Println("Created PagerDuty event for metric: ", metric.Name())
	}
	return nil
}

func init() {
	outputs.Add("pagerduty", func() telegraf.Output { return &PD{} })
}
