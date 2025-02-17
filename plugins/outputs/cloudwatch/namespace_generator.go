package cloudwatch

import (
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/influxdata/telegraf"
)

type NamespaceGenerator struct {
	metric   telegraf.Metric
	template *template.Template
}

func NewNamespaceGenerator(namespace string) (*NamespaceGenerator, error) {
	nt, err := template.New("namespace").Funcs(sprig.TxtFuncMap()).Parse(namespace)
	if err != nil {
		return nil, err
	}
	return &NamespaceGenerator{template: nt}, nil
}

func (t *NamespaceGenerator) Generate(m telegraf.Metric) (string, error) {
	var b strings.Builder
	err := t.template.Execute(&b, t)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
