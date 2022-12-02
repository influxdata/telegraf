package mqtt

import (
	"strings"

	"github.com/influxdata/telegraf"
)

type TopicGenerator struct {
	Hostname    string
	metric      telegraf.Metric
	topicPrefix string
}

func (t *TopicGenerator) Tag(key string) string {
	tagString, _ := t.metric.GetTag(key)
	return tagString
}

func (t *TopicGenerator) TopicPrefix() string {
	return t.topicPrefix
}

func (t *TopicGenerator) PluginName() string {
	return t.metric.Name()
}

func (t *TopicGenerator) Generate(m *MQTT) string {
	var b strings.Builder
	t.topicPrefix = m.TopicPrefix
	err := m.template.Execute(&b, t)
	if err != nil {
		panic(err)
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
