package mqtt

import (
	"strings"
	"text/template"

	"github.com/influxdata/telegraf"
)

type TopicNameGenerator struct {
	Hostname    string
	TopicPrefix string
	PluginName  string
	metric      telegraf.Metric
	template    *template.Template
}

func NewTopicNameGenerator(topicPrefix string, topic string) (*TopicNameGenerator, error) {
	tt, err := template.New("topic_name").Parse(topic)
	if err != nil {
		return nil, err
	}
	return &TopicNameGenerator{TopicPrefix: topicPrefix, template: tt}, nil
}

func (t *TopicNameGenerator) Tag(key string) string {
	tagString, _ := t.metric.GetTag(key)
	return tagString
}

func (t *TopicNameGenerator) Generate(hostname string, m telegraf.Metric) (string, error) {
	t.Hostname = hostname
	t.metric = m
	t.PluginName = m.Name()
	var b strings.Builder
	err := t.template.Execute(&b, t)
	if err != nil {
		return "", err
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
		return m.Name(), nil
	}
	return topic, nil
}
