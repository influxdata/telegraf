package mqtt

import (
	"strings"

	"github.com/influxdata/telegraf"
)

type TemplateTopic struct {
	Hostname, TopicPrefix string
	metric                telegraf.Metric
}

func (t *TemplateTopic) Tag(key string) string {
	tagString, _ := t.metric.GetTag(key)
	return tagString
}

func (t *TemplateTopic) PluginName() string {
	return t.metric.Name()
}

func (t *TemplateTopic) Parse(m *MQTT) string {
	var b strings.Builder
	err := m.template.Execute(&b, t)
	if err != nil {
		panic("err")
	}
	var ts []string
	for _, p := range strings.Split(b.String(), "/") {
		if p != "" {
			ts = append(ts, p)
		}
	}
	topic := strings.Join(ts, "/")
	// This is to keep backward compatibility with previous behaviour where the plugin name was always present
	if topic == "" {
		return t.PluginName()
	}
	return topic
}
