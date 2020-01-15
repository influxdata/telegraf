package template

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func newMetric(name string, tags map[string]string, fields map[string]interface{}) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m, _ := metric.New(name, tags, fields, time.Now())
	return m
}

func TestTagTemplateConcatenate(t *testing.T) {
	// Create Template processor
	tmp := TemplateProcessor{Tag: "topic", Template: "{{.Tag \"hostname\"}}.{{ .Tag \"level\" }}"}
	// manually init
	err := tmp.Init()

	if err != nil {
		panic(err)
	}

	// test metric
	m1 := newMetric("Tags", map[string]string{"hostname": "localhost", "level": "debug"}, nil)

	// act
	result := tmp.Apply(m1)

	// assert
	resultTaglist := result[0].TagList()
	
	assert.True(t, contains(resultTaglist, "topic", "localhost.debug"))
}

func contains(s []*telegraf.Tag, name string, value string) bool {
	for _, a := range s {
		if a.Key == name && a.Value == value {
			return true
		}
	}
	return false
}
