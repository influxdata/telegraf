package clone

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestRetainsTags(t *testing.T) {
	input := metric.New(
		"m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Unix(0, 0),
	)

	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
	}

	plugin := &Clone{}
	actual := plugin.Apply(input)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestAddTags(t *testing.T) {
	input := metric.New(
		"m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Unix(0, 0),
	)

	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{
				"metric_tag":  "from_metric",
				"added_tag":   "from_config",
				"another_tag": "",
			},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
	}

	plugin := &Clone{
		Tags: map[string]string{
			"added_tag":   "from_config",
			"another_tag": "",
		},
	}
	actual := plugin.Apply(input)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestOverwritesPresentTagValues(t *testing.T) {
	input := metric.New(
		"m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Unix(0, 0),
	)

	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_config"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
	}

	plugin := &Clone{
		Tags: map[string]string{"metric_tag": "from_config"},
	}
	actual := plugin.Apply(input)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestOverridesName(t *testing.T) {
	input := metric.New(
		"m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Unix(0, 0),
	)

	expected := []telegraf.Metric{
		metric.New(
			"overridden",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
	}

	plugin := &Clone{NameOverride: "overridden"}
	actual := plugin.Apply(input)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestNamePrefix(t *testing.T) {
	input := metric.New(
		"m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Unix(0, 0),
	)

	expected := []telegraf.Metric{
		metric.New(
			"Pre-m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
	}

	plugin := &Clone{NamePrefix: "Pre-"}
	actual := plugin.Apply(input)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestNameSuffix(t *testing.T) {
	input := metric.New(
		"m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Unix(0, 0),
	)

	expected := []telegraf.Metric{
		metric.New(
			"m1-suff",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
	}

	plugin := &Clone{NameSuffix: "-suff"}
	actual := plugin.Apply(input)
	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{"metric_tag": "from_metric"},
			map[string]interface{}{"value": int64(1)},
			time.Now(),
		),
		metric.New(
			"m2",
			map[string]string{"metric_tag": "foo_metric"},
			map[string]interface{}{"value": int64(2)},
			time.Now(),
		),
	}

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}
	input := make([]telegraf.Metric, 0, len(inputRaw))
	expected := make([]telegraf.Metric, 0, 2*len(input))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
		expected = append(expected, m)
	}
	expected = append(expected, input...)

	// Process expected metrics and compare with resulting metrics
	plugin := &Clone{}
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
