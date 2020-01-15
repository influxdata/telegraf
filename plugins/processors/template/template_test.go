package template

import (
	"io/ioutil"
	"os"
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
	tmp := TemplateProcessor{Tag: "topic", Template: "{{ index .Tags \"hostname\" }}.{{ index .Tags \"level\" }}"}

	tmp.Init()

	m1 := newMetric("Tags", map[string]string{"hostname": "localhost", "level": "debug"}, nil)
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	result := tmp.Apply(m1)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	t.Logf("Captured: %s", out) // prints: Captured: Hello, playground
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
