package mqtt

import (
	"strings"
	"text/template"

	"github.com/influxdata/telegraf"
)

type TopicNameGenerator struct {
	Hostname    string
	TopicPrefix string
	metric      telegraf.Metric
	template    *template.Template
}

func NewTopicNameGenerator(hostname string, topicPrefix string, temp *template.Template) *TopicNameGenerator {
	return &TopicNameGenerator{Hostname: hostname, TopicPrefix: topicPrefix, template: temp}
}

func (t *TopicNameGenerator) Tag(key string) string {
	tagString, _ := t.metric.GetTag(key)
	return tagString
}

func (t *TopicNameGenerator) PluginName() string {
	return t.metric.Name()
}

func (t *TopicNameGenerator) Generate(m telegraf.Metric) string {
	t.metric = m
	var b strings.Builder
	err := t.template.Execute(&b, t)
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
