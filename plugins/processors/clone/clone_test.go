package clone

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func createTestMetric() telegraf.Metric {
	m := metric.New("m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Now(),
	)
	return m
}

func calculateProcessedTags(processor Clone, m telegraf.Metric) map[string]string {
	processed := processor.Apply(m)
	return processed[0].Tags()
}

func TestRetainsTags(t *testing.T) {
	processor := Clone{}

	tags := calculateProcessedTags(processor, createTestMetric())

	value, present := tags["metric_tag"]
	require.True(t, present, "Tag of metric was not present")
	require.Equal(t, "from_metric", value, "Value of Tag was changed")
}

func TestAddTags(t *testing.T) {
	processor := Clone{Tags: map[string]string{"added_tag": "from_config", "another_tag": ""}}

	tags := calculateProcessedTags(processor, createTestMetric())

	value, present := tags["added_tag"]
	require.True(t, present, "Additional Tag of metric was not present")
	require.Equal(t, "from_config", value, "Value of Tag was changed")
	require.Equal(t, 3, len(tags), "Should have one previous and two added tags.")
}

func TestOverwritesPresentTagValues(t *testing.T) {
	processor := Clone{Tags: map[string]string{"metric_tag": "from_config"}}

	tags := calculateProcessedTags(processor, createTestMetric())

	value, present := tags["metric_tag"]
	require.True(t, present, "Tag of metric was not present")
	require.Equal(t, 1, len(tags), "Should only have one tag.")
	require.Equal(t, "from_config", value, "Value of Tag was not changed")
}

func TestOverridesName(t *testing.T) {
	processor := Clone{NameOverride: "overridden"}

	processed := processor.Apply(createTestMetric())

	require.Equal(t, "overridden", processed[0].Name(), "Name was not overridden")
	require.Equal(t, "m1", processed[1].Name(), "Original metric was modified")
}

func TestNamePrefix(t *testing.T) {
	processor := Clone{NamePrefix: "Pre-"}

	processed := processor.Apply(createTestMetric())

	require.Equal(t, "Pre-m1", processed[0].Name(), "Prefix was not applied")
	require.Equal(t, "m1", processed[1].Name(), "Original metric was modified")
}

func TestNameSuffix(t *testing.T) {
	processor := Clone{NameSuffix: "-suff"}

	processed := processor.Apply(createTestMetric())

	require.Equal(t, "m1-suff", processed[0].Name(), "Suffix was not applied")
	require.Equal(t, "m1", processed[1].Name(), "Original metric was modified")
}
