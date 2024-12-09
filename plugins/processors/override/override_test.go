package override

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func createTestMetric() telegraf.Metric {
	m := metric.New("m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Now(),
	)
	return m
}

func calculateProcessedTags(processor Override, m telegraf.Metric) map[string]string {
	processed := processor.Apply(m)
	return processed[0].Tags()
}

func TestRetainsTags(t *testing.T) {
	processor := Override{}

	tags := calculateProcessedTags(processor, createTestMetric())

	value, present := tags["metric_tag"]
	require.True(t, present, "Tag of metric was not present")
	require.Equal(t, "from_metric", value, "Value of Tag was changed")
}

func TestAddTags(t *testing.T) {
	processor := Override{Tags: map[string]string{"added_tag": "from_config", "another_tag": ""}}

	tags := calculateProcessedTags(processor, createTestMetric())

	value, present := tags["added_tag"]
	require.True(t, present, "Additional Tag of metric was not present")
	require.Equal(t, "from_config", value, "Value of Tag was changed")
	require.Len(t, tags, 3, "Should have one previous and two added tags.")
}

func TestOverwritesPresentTagValues(t *testing.T) {
	processor := Override{Tags: map[string]string{"metric_tag": "from_config"}}

	tags := calculateProcessedTags(processor, createTestMetric())

	value, present := tags["metric_tag"]
	require.True(t, present, "Tag of metric was not present")
	require.Len(t, tags, 1, "Should only have one tag.")
	require.Equal(t, "from_config", value, "Value of Tag was not changed")
}

func TestOverridesName(t *testing.T) {
	processor := Override{NameOverride: "overridden"}

	processed := processor.Apply(createTestMetric())

	require.Equal(t, "overridden", processed[0].Name(), "Name was not overridden")
}

func TestNamePrefix(t *testing.T) {
	processor := Override{NamePrefix: "Pre-"}

	processed := processor.Apply(createTestMetric())

	require.Equal(t, "Pre-m1", processed[0].Name(), "Prefix was not applied")
}

func TestNameSuffix(t *testing.T) {
	processor := Override{NameSuffix: "-suff"}

	processed := processor.Apply(createTestMetric())

	require.Equal(t, "m1-suff", processed[0].Name(), "Suffix was not applied")
}
func TestTracking(t *testing.T) {
	// Setup raw input and expected output
	inputRaw := []telegraf.Metric{
		metric.New(
			"zero_uint64",
			map[string]string{},
			map[string]interface{}{"value": uint64(3)},
			time.Unix(0, 0),
		),
		metric.New(
			"zero_int64",
			map[string]string{},
			map[string]interface{}{"value": int64(4)},
			time.Unix(0, 0),
		),
		metric.New(
			"zero_float",
			map[string]string{},
			map[string]interface{}{"value": float64(5.5)},
			time.Unix(0, 0),
		),
	}

	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": uint64(3)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": int64(4)},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{},
			map[string]interface{}{"value": float64(5.5)},
			time.Unix(0, 0),
		),
	}

	// Create fake notification for testing
	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	// Convert raw input to tracking metric
	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	// Prepare and start the plugin
	plugin := &Override{
		NameOverride: "test",
	}

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
