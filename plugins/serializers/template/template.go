package template

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Serializer struct {
	Template      string          `toml:"template"`
	BatchTemplate string          `toml:"batch_template"`
	Log           telegraf.Logger `toml:"-"`

	tmplMetric *template.Template
	tmplBatch  *template.Template
}

func (s *Serializer) Init() error {
	// Setting defaults
	var err error

	s.tmplMetric, err = template.New("template").Parse(s.Template)
	if err != nil {
		return fmt.Errorf("creating template failed: %w", err)
	}
	if s.BatchTemplate == "" {
		s.BatchTemplate = fmt.Sprintf("{{range .}}%s{{end}}", s.Template)
	}
	s.tmplBatch, err = template.New("batch template").Parse(s.BatchTemplate)
	if err != nil {
		return fmt.Errorf("creating batch template failed: %w", err)
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
	if s.Template == "" && s.BatchTemplate != "" {
		// If the user has only defined a batch template, fall back to batches of 1
		metrics := []telegraf.TemplateMetric{m}

		if err := s.tmplBatch.Execute(&b, &metrics); err != nil {
			s.Log.Errorf("failed to execute batch template: %v", err)
			return nil, nil
		}
	} else {
		if err := s.tmplMetric.Execute(&b, &m); err != nil {
			s.Log.Errorf("failed to execute template: %v", err)
			return nil, nil
		}
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
	if err := s.tmplBatch.Execute(&b, &newMetrics); err != nil {
		s.Log.Errorf("failed to execute batch template: %v", err)
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
