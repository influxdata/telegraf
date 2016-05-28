package pagerduty

import (
	"fmt"
	pd "github.com/PagerDuty/go-pagerduty"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"go/token"
	"go/types"
	"log"
)

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
	event := pd.Event{
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
		_, err := pd.CreateEvent(event)
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
