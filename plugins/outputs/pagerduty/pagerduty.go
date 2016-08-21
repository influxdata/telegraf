package pagerduty

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Tag struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

type PD struct {
	ServiceKey  string `toml:"service_key"`
	Desc        string `toml:"description"`
	Metric      string `toml:"metric"`
	Field       string `toml:"field"`
	Expression  string `toml:"expression"`
	TagFilter   []Tag  `toml:"tags"`
	incidentKey string `toml:"-"`
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
  ## Tag filter, when present only metrics with the specified tag value
  ## will be considered for further processing
  [[outputs.pagerduty.tag_filter]]
    role = "web-server"
  ## Expression is used to evaluate the alert
  expression = "> 50.0"
`

func (p *PD) Connect() error {
	return nil
}

func (p *PD) Close() error {
	return nil
}

func (p *PD) SampleConfig() string {
	return sampleConfig
}

func (p *PD) Description() string {
	return "Send PagerDuty alert based on metric values"
}

func (p *PD) isMatch(metric telegraf.Metric) bool {
	if p.Metric != metric.Name() {
		return false
	}
	for _, tag := range p.TagFilter {
		v, ok := metric.Tags()[tag.Name]
		if !ok || (v != tag.Value) {
			return false
		}
	}
	return true
}

func init() {
	outputs.Add("pagerduty", func() telegraf.Output {
		return &PD{}
	})
}

func (p *PD) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		if p.isMatch(metric) {
			p.processForEvent(metric)
		}
	}
	return nil
}

func (p *PD) processForEvent(metric telegraf.Metric) error {
	value, ok := metric.Fields()[p.Field]
	if !ok {
		return fmt.Errorf("Filed '%s' absent", p.Field)
	}
	expr := fmt.Sprintf("%v %s", value, p.Expression)
	trigger, err := evalBoolExpr(expr)
	if err != nil {
		return err
	}

	event := Event{
		ServiceKey: p.ServiceKey,
		Client:     "telegraf",
	}
	m := make(map[string]interface{})
	m["tags"] = metric.Tags()
	m["fields"] = metric.Fields()
	m["name"] = metric.Name()
	m["timestamp"] = metric.UnixNano() / 1000000000
	event.Details = m
	event.Description = p.Desc
	// either retrigger of create a new event
	if trigger {
		event.Type = "trigger"
		if p.incidentKey != "" {
			// already triggered incident, retriggering it
			// PagerDuty dedups alerts when we reuse the incidentkey
			event.IncidentKey = p.incidentKey
		}
		resp, err := createEvent(event)
		if err != nil {
			return err
		}
		p.incidentKey = resp.IncidentKey
		return nil
	}
	if p.incidentKey != "" {
		event.IncidentKey = p.incidentKey
		event.Type = "resolve"
		_, err := createEvent(event)
		if err != nil {
			return err
		}
		// incident is resolved hence reset incident key
		p.incidentKey = ""
	}
	return nil
}
