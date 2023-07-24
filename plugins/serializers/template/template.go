package template

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Serializer struct {
	Template string          `toml:"template"`
	Log      telegraf.Logger `toml:"-"`

	outTemplate *template.Template
}

func (s *Serializer) Init() error {
	// Setting defaults
	var err error

	s.outTemplate, err = template.New("template").Parse(s.Template)
	if err != nil {
		return fmt.Errorf("creating template failed: %w", err)
	}
	return nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	m, ok := metric.(telegraf.TemplateMetric)
	if !ok {
		s.Log.Errorf("metric of type %T is not a template metric", metric)
		return nil, nil
	}
	var b strings.Builder
	if err := s.outTemplate.Execute(&b, &m); err != nil {
		s.Log.Errorf("failed to execute template: %v", err)
		return nil, nil
	}

	return []byte(b.String()), nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	if len(metrics) < 1 {
		return nil, nil
	}
	newMetrics := make([]telegraf.TemplateMetric, 0, len(metrics))

	for _, metric := range metrics {
		m, ok := metric.(telegraf.TemplateMetric)
		if !ok {
			s.Log.Errorf("metric of type %T is not a template metric", metric)
			return nil, nil
		}
		newMetrics = append(newMetrics, m)
	}

	var b strings.Builder
	if err := s.outTemplate.Execute(&b, &newMetrics); err != nil {
		s.Log.Errorf("failed to execute template: %v", err)
		return nil, nil
	}

	return []byte(b.String()), nil
}

func init() {
	serializers.Add("template",
		func() serializers.Serializer {
			return &Serializer{}
		},
	)
}

// InitFromConfig is a compatibility function to construct the parser the old way
func (s *Serializer) InitFromConfig(cfg *serializers.Config) error {
	s.Template = cfg.Template

	return nil
}
