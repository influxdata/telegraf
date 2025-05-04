package mqtt

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"

	"github.com/influxdata/telegraf"
)

type TopicNameGenerator struct {
	TopicPrefix string
	metric      telegraf.TemplateMetric
	template    *template.Template
}

func NewTopicNameGenerator(topicPrefix, topic string) (*TopicNameGenerator, error) {
	topic = hostnameRe.ReplaceAllString(topic, `$1.Tag "host"$2`)
	topic = pluginNameRe.ReplaceAllString(topic, `$1.Name$2`)

	tt, err := template.New("topic_name").Funcs(sprig.TxtFuncMap()).Parse(topic)
	if err != nil {
		return nil, err
	}
	for _, p := range strings.Split(topic, "/") {
		if strings.ContainsAny(p, "#+") {
			return nil, fmt.Errorf("found forbidden character %s in the topic name %s", p, topic)
		}
	}
	return &TopicNameGenerator{TopicPrefix: topicPrefix, template: tt}, nil
}

func (t *TopicNameGenerator) Name() string {
	return t.metric.Name()
}

func (t *TopicNameGenerator) Tag(key string) string {
	return t.metric.Tag(key)
}

func (t *TopicNameGenerator) Field(key string) interface{} {
	return t.metric.Field(key)
}

func (t *TopicNameGenerator) Time() time.Time {
	return t.metric.Time()
}

func (t *TopicNameGenerator) Tags() map[string]string {
	return t.metric.Tags()
}

func (t *TopicNameGenerator) Fields() map[string]interface{} {
	return t.metric.Fields()
}

func (t *TopicNameGenerator) String() string {
	return t.metric.String()
}

func (m *MQTT) generateTopic(metric telegraf.Metric) (string, error) {
	m.generator.metric = metric.(telegraf.TemplateMetric)

	// Cannot directly pass TemplateMetric since TopicNameGenerator still contains TopicPrefix (until v1.35.0)
	var b strings.Builder
	err := m.generator.template.Execute(&b, m.generator)
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
		return metric.Name(), nil
	}
	if strings.HasPrefix(b.String(), "/") {
		topic = "/" + topic
	}
	return topic, nil
}
