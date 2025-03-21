package filter

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var testmetrics = []telegraf.Metric{
	metric.New(
		"packing",
		map[string]string{
			"source":   "machine A",
			"location": "main building",
			"status":   "OK",
		},
		map[string]interface{}{
			"operating_hours": 37,
			"temperature":     23.1,
		},
		time.Unix(0, 0),
	),
	metric.New(
		"foundry",
		map[string]string{
			"source":   "machine B",
			"location": "factory X",
			"status":   "OK",
		},
		map[string]interface{}{
			"operating_hours": 1337,
			"temperature":     19.9,
			"pieces":          96878,
		},
		time.Unix(0, 0),
	),
	metric.New(
		"welding",
		map[string]string{
			"source":   "machine C",
			"location": "factory X",
			"status":   "failure",
		},
		map[string]interface{}{
			"operating_hours": 1009,
			"temperature":     67.3,
			"message":         "temperature alert",
		},
		time.Unix(0, 0),
	),
	metric.New(
		"welding",
		map[string]string{
			"source":   "machine D",
			"location": "factory Y",
			"status":   "OK",
		},
		map[string]interface{}{
			"operating_hours": 825,
			"temperature":     31.2,
		},
		time.Unix(0, 0),
	),
}

func TestNoRules(t *testing.T) {
	logger := &testutil.CaptureLogger{}
	plugin := &Filter{
		DefaultAction: "drop",
		Log:           logger,
	}
	require.NoError(t, plugin.Init())

	warnings := logger.Warnings()
	require.Len(t, warnings, 1)
	require.Contains(t, warnings[0], "dropping all metrics")
}

func TestInvalidDefaultAction(t *testing.T) {
	plugin := &Filter{
		Rules:         []rule{{Name: []string{"foo"}}},
		DefaultAction: "foo",
	}
	require.ErrorContains(t, plugin.Init(), "invalid default action")
}

func TestNoMetric(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{{Name: []string{"*"}}},
	}
	require.NoError(t, plugin.Init())

	var input []telegraf.Metric
	require.Empty(t, plugin.Apply(input...))
}

func TestDropAll(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{{Name: []string{"*"}}},
	}
	require.NoError(t, plugin.Init())
	require.Empty(t, plugin.Apply(testmetrics...))
}

func TestDropDefault(t *testing.T) {
	plugin := &Filter{
		Rules:         []rule{{Name: []string{"foo"}, Action: "pass"}},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())
	require.Empty(t, plugin.Apply(testmetrics...))
}

func TestPassAll(t *testing.T) {
	plugin := &Filter{
		Rules:         []rule{{Name: []string{"*"}, Action: "pass"}},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	expected := testmetrics
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestPassDefault(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{{Name: []string{"foo"}, Action: "drop"}},
	}
	require.NoError(t, plugin.Init())

	expected := testmetrics
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestNamePass(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Name:   []string{"welding"},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine C",
				"location": "factory X",
				"status":   "failure",
			},
			map[string]interface{}{
				"operating_hours": 1009,
				"temperature":     67.3,
				"message":         "temperature alert",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine D",
				"location": "factory Y",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 825,
				"temperature":     31.2,
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestNameDrop(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Name:   []string{"welding"},
				Action: "drop",
			},
		},
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"packing",
			map[string]string{
				"source":   "machine A",
				"location": "main building",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 37,
				"temperature":     23.1,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"foundry",
			map[string]string{
				"source":   "machine B",
				"location": "factory X",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 1337,
				"temperature":     19.9,
				"pieces":          96878,
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestNameGlob(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Name:   []string{"*ing"},
				Action: "drop",
			},
		},
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"foundry",
			map[string]string{
				"source":   "machine B",
				"location": "factory X",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 1337,
				"temperature":     19.9,
				"pieces":          96878,
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTagPass(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Tags:   map[string][]string{"status": {"OK"}},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"packing",
			map[string]string{
				"source":   "machine A",
				"location": "main building",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 37,
				"temperature":     23.1,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"foundry",
			map[string]string{
				"source":   "machine B",
				"location": "factory X",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 1337,
				"temperature":     19.9,
				"pieces":          96878,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine D",
				"location": "factory Y",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 825,
				"temperature":     31.2,
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTagDrop(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Tags:   map[string][]string{"status": {"OK"}},
				Action: "drop",
			},
		},
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine C",
				"location": "factory X",
				"status":   "failure",
			},
			map[string]interface{}{
				"operating_hours": 1009,
				"temperature":     67.3,
				"message":         "temperature alert",
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTagMultiple(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Tags: map[string][]string{
					"location": {"factory X", "factory Y"},
					"status":   {"OK"},
				},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"foundry",
			map[string]string{
				"source":   "machine B",
				"location": "factory X",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 1337,
				"temperature":     19.9,
				"pieces":          96878,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine D",
				"location": "factory Y",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 825,
				"temperature":     31.2,
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTagGlob(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Tags:   map[string][]string{"location": {"factory *"}},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"foundry",
			map[string]string{
				"source":   "machine B",
				"location": "factory X",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 1337,
				"temperature":     19.9,
				"pieces":          96878,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine C",
				"location": "factory X",
				"status":   "failure",
			},
			map[string]interface{}{
				"operating_hours": 1009,
				"temperature":     67.3,
				"message":         "temperature alert",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine D",
				"location": "factory Y",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 825,
				"temperature":     31.2,
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTagDoesNotExist(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Tags: map[string][]string{
					"operator": {"peter"},
					"status":   {"OK"},
				},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	require.Empty(t, plugin.Apply(testmetrics...))
}

func TestFieldPass(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Fields: []string{"message", "pieces"},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"foundry",
			map[string]string{
				"source":   "machine B",
				"location": "factory X",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 1337,
				"temperature":     19.9,
				"pieces":          96878,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine C",
				"location": "factory X",
				"status":   "failure",
			},
			map[string]interface{}{
				"operating_hours": 1009,
				"temperature":     67.3,
				"message":         "temperature alert",
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestFieldDrop(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Fields: []string{"message", "pieces"},
				Action: "drop",
			},
		},
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"packing",
			map[string]string{
				"source":   "machine A",
				"location": "main building",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 37,
				"temperature":     23.1,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine D",
				"location": "factory Y",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 825,
				"temperature":     31.2,
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestFieldGlob(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Fields: []string{"{message,piece*}"},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"foundry",
			map[string]string{
				"source":   "machine B",
				"location": "factory X",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 1337,
				"temperature":     19.9,
				"pieces":          96878,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine C",
				"location": "factory X",
				"status":   "failure",
			},
			map[string]interface{}{
				"operating_hours": 1009,
				"temperature":     67.3,
				"message":         "temperature alert",
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestRuleOrder(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Name:   []string{"welding"},
				Action: "drop",
			},
			{
				Name:   []string{"welding"},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())
	require.Empty(t, plugin.Apply(testmetrics...))
}

func TestRuleMultiple(t *testing.T) {
	plugin := &Filter{
		Rules: []rule{
			{
				Name:   []string{"welding"},
				Action: "drop",
			},
			{
				Name:   []string{"foundry"},
				Action: "drop",
			},
		},
		DefaultAction: "pass",
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"packing",
			map[string]string{
				"source":   "machine A",
				"location": "main building",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 37,
				"temperature":     23.1,
			},
			time.Unix(0, 0),
		),
	}
	actual := plugin.Apply(testmetrics...)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTracking(t *testing.T) {
	inputRaw := testmetrics

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m.Copy(), notify)
		input = append(input, tm)
	}

	expected := []telegraf.Metric{
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine C",
				"location": "factory X",
				"status":   "failure",
			},
			map[string]interface{}{
				"operating_hours": 1009,
				"temperature":     67.3,
				"message":         "temperature alert",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"welding",
			map[string]string{
				"source":   "machine D",
				"location": "factory Y",
				"status":   "OK",
			},
			map[string]interface{}{
				"operating_hours": 825,
				"temperature":     31.2,
			},
			time.Unix(0, 0),
		),
	}

	plugin := &Filter{
		Rules: []rule{
			{
				Name:   []string{"welding"},
				Action: "pass",
			},
		},
		DefaultAction: "drop",
	}
	require.NoError(t, plugin.Init())

	// Process expected metrics and compare with resulting metrics
	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
