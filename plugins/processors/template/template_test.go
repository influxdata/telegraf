package template

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestName(t *testing.T) {
	plugin := TemplateProcessor{
		Tag:      "measurement",
		Template: "{{ .Name }}",
	}

	err := plugin.Init()
	require.NoError(t, err)

	input := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	}

	actual := plugin.Apply(input...)
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"measurement": "cpu",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestNameTemplate(t *testing.T) {
	plugin := TemplateProcessor{
		Tag:      `{{ .Tag "foo" }}`,
		Template: `{{ .Name }}`,
	}

	err := plugin.Init()
	require.NoError(t, err)

	input := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{"foo": "measurement"},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	}

	actual := plugin.Apply(input...)
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"foo":         "measurement",
				"measurement": "cpu",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, actual)
}

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
	expected := []telegraf.Metric{
		testutil.MustMetric("Tags", map[string]string{"hostname": "localhost", "level": "debug", "topic": "localhost.debug"}, nil, now),
	}
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
	require.Len(t, actual, 2, "Number of metrics input should equal number of metrics output")
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
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"weather",
			map[string]string{"location": "us-midwest", "LocalTemp": "us-midwest is too warm"},
			map[string]interface{}{"temperature": "too warm"},
			now,
		),
	}
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestFieldList(t *testing.T) {
	// Prepare
	plugin := TemplateProcessor{Tag: "fields", Template: "{{.FieldList}}"}
	require.NoError(t, plugin.Init())

	// Run
	m := testutil.TestMetric(1.23)
	actual := plugin.Apply(m)

	// Verify
	expected := m.Copy()
	expected.AddTag("fields", "map[value:1.23]")
	testutil.RequireMetricsEqual(t, []telegraf.Metric{expected}, actual)
}

func TestTagList(t *testing.T) {
	// Prepare
	plugin := TemplateProcessor{Tag: "tags", Template: "{{.TagList}}"}
	require.NoError(t, plugin.Init())

	// Run
	m := testutil.TestMetric(1.23)
	actual := plugin.Apply(m)

	// Verify
	expected := m.Copy()
	expected.AddTag("tags", "map[tag1:value1]")
	testutil.RequireMetricsEqual(t, []telegraf.Metric{expected}, actual)
}

func TestFields(t *testing.T) {
	// Prepare
	plugin := TemplateProcessor{
		Tag:      "fields",
		Template: "{{.Fields}}",
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Run
	m := testutil.TestMetric(1.23)
	actual := plugin.Apply(m)

	// Verify
	expected := m.Copy()
	expected.AddTag("fields", "map[value:1.23]")
	testutil.RequireMetricsEqual(t, []telegraf.Metric{expected}, actual)
}

func TestTags(t *testing.T) {
	// Prepare
	plugin := TemplateProcessor{
		Tag:      "tags",
		Template: "{{.Tags}}",
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Run
	m := testutil.TestMetric(1.23)
	actual := plugin.Apply(m)

	// Verify
	expected := m.Copy()
	expected.AddTag("tags", "map[tag1:value1]")
	testutil.RequireMetricsEqual(t, []telegraf.Metric{expected}, actual)
}

func TestString(t *testing.T) {
	// Prepare
	plugin := TemplateProcessor{
		Tag:      "tags",
		Template: "{{.}}",
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Run
	m := testutil.TestMetric(1.23)
	actual := plugin.Apply(m)

	// Verify
	expected := m.Copy()
	expected.AddTag("tags", "test1 map[tag1:value1] map[value:1.23] 1257894000000000000")
	testutil.RequireMetricsEqual(t, []telegraf.Metric{expected}, actual)
}

func TestDot(t *testing.T) {
	// Prepare
	plugin := TemplateProcessor{Tag: "metric", Template: "{{.}}"}
	require.NoError(t, plugin.Init())

	// Run
	m := testutil.TestMetric(1.23)
	actual := plugin.Apply(m)

	// Verify
	expected := m.Copy()
	expected.AddTag("metric", "test1 map[tag1:value1] map[value:1.23] 1257894000000000000")
	testutil.RequireMetricsEqual(t, []telegraf.Metric{expected}, actual)
}

func TestTracking(t *testing.T) {
	// Create a tracking metric and tap the delivery information
	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, 1)
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}
	m := testutil.TestMetric(1.23)
	input, _ := metric.WithTracking(m, notify)

	// Create an expectation
	e := m.Copy()
	e.AddTag("metric", "test1 map[tag1:value1] map[value:1.23] 1257894000000000000")
	expected := []telegraf.Metric{e}

	// Configure the plugin
	plugin := TemplateProcessor{Tag: "metric", Template: "{{.}}"}
	require.NoError(t, plugin.Init())

	// Process expected metrics and compare with resulting metrics
	actual := plugin.Apply(input)
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(delivered) == 1
	}, time.Second, 100*time.Millisecond, "%d delivered but 1 expected", len(delivered))
}

func TestSprig(t *testing.T) {
	plugin := TemplateProcessor{
		Tag:      `{{ .Tag "foo" | lower }}`,
		Template: `{{ .Name | upper }}`,
	}

	err := plugin.Init()
	require.NoError(t, err)

	input := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{"foo": "MEASUREMENT"},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	}

	actual := plugin.Apply(input...)
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"foo":         "MEASUREMENT",
				"measurement": "CPU",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, actual)
}
