package template

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestTagTemplateConcatenate(t *testing.T) {
	now := time.Now()

	// Create Template processor
	tmp := TemplateProcessor{Tag: "topic", Template: "{{.Tag \"hostname\"}}.{{ .Tag \"level\" }}"}
	// manually init
	err := tmp.Init()

	if err != nil {
		panic(err)
	}

	// create metric for testing
	input := []telegraf.Metric{testutil.MustMetric("Tags", map[string]string{"hostname": "localhost", "level": "debug"}, nil, now)}

	// act
	actual := tmp.Apply(input[0])

	// assert
	expected := []telegraf.Metric{testutil.MustMetric("Tags", map[string]string{"hostname": "localhost", "level": "debug", "topic": "localhost.debug"}, nil, now)}
	testutil.RequireMetricsEqual(t, expected, actual)
}
