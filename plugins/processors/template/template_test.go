package template

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestTagTemplateConcatenate(t *testing.T) {
	now := time.Now()

	// Create Template processor
	tmp := TemplateProcessor{Tag: "topic", Template: `{{.Tag "hostname"}}.{{ .Tag "level" }}`}
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

func TestMetricMissingTagsIsNotLost(t *testing.T) {
	now := time.Now()

	// Create Template processor
	tmp := TemplateProcessor{Tag: "topic", Template: `{{.Tag "hostname"}}.{{ .Tag "level" }}`}
	// manually init
	err := tmp.Init()

	if err != nil {
		panic(err)
	}

	// create metrics for testing
	m1 := testutil.MustMetric("Works", map[string]string{"hostname": "localhost", "level": "debug"}, nil, now)
	m2 := testutil.MustMetric("Fails", map[string]string{"hostname": "localhost"}, nil, now)

	// act
	actual := tmp.Apply(m1, m2)

	// assert
	// make sure no metrics are lost when a template process fails
	assert.Equal(t, 2, len(actual), "Number of metrics input should equal number of metrics output")
}

func TestTagAndFieldConcatenate(t *testing.T) {
	now := time.Now()

	// Create Template processor
	tmp := TemplateProcessor{Tag: "LocalTemp", Template: `{{.Tag "location"}} is {{ .Field "temperature" }}`}
	// manually init
	err := tmp.Init()

	if err != nil {
		panic(err)
	}

	// create metric for testing
	m1 := testutil.MustMetric("weather", map[string]string{"location": "us-midwest"}, map[string]interface{}{"temperature": "too warm"}, now)

	// act
	actual := tmp.Apply(m1)

	// assert
	expected := []telegraf.Metric{testutil.MustMetric("weather", map[string]string{"location": "us-midwest", "LocalTemp": "us-midwest is too warm"}, map[string]interface{}{"temperature": "too warm"}, now)}
	testutil.RequireMetricsEqual(t, expected, actual)
}
