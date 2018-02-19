package tags

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func createTestMetric() telegraf.Metric {
	metric, _ := metric.New("m1",
		map[string]string{"metric_tag": "from_metric"},
		map[string]interface{}{"value": int64(1)},
		time.Now(),
	)
	return metric
}

func calculateProcessedTags(adder TagAdder, metric telegraf.Metric) map[string]string {
	processed := adder.Apply(metric)
	return processed[0].Tags()
}

func TestRetainsTags(t *testing.T) {
	adder := TagAdder{}

	tags := calculateProcessedTags(adder, createTestMetric())

	value, present := tags["metric_tag"]
	assert.True(t, present, "Tag of metric was not present")
	assert.Equal(t, "from_metric", value, "Value of Tag was changed")
}

func TestAddTags(t *testing.T) {
	adder := TagAdder{Tags: map[string]string{"added_tag": "from_config", "another_tag": ""}}

	tags := calculateProcessedTags(adder, createTestMetric())

	value, present := tags["added_tag"]
	assert.True(t, present, "Additional Tag of metric was not present")
	assert.Equal(t, "from_config", value, "Value of Tag was changed")
	assert.Equal(t, 3, len(tags), "Should have one previous and two added tags.")
}

func TestOverwritesPresentTagValues(t *testing.T) {
	adder := TagAdder{Tags: map[string]string{"metric_tag": "from_config"}}

	tags := calculateProcessedTags(adder, createTestMetric())

	value, present := tags["metric_tag"]
	assert.True(t, present, "Tag of metric was not present")
	assert.Equal(t, 1, len(tags), "Should only have one tag.")
	assert.Equal(t, "from_config", value, "Value of Tag was not changed")
}

func TestOverridesName(t *testing.T) {
	adder := TagAdder{NameOverride: "overridden"}

	processed := adder.Apply(createTestMetric())

	assert.Equal(t, "overridden", processed[0].Name(), "Name was not overridden")
}

func TestNamePrefix(t *testing.T) {
	adder := TagAdder{NamePrefix: "Pre-"}

	processed := adder.Apply(createTestMetric())

	assert.Equal(t, "Pre-m1", processed[0].Name(), "Prefix was not applied")
}

func TestNameSuffix(t *testing.T) {
	adder := TagAdder{NameSuffix: "-suff"}

	processed := adder.Apply(createTestMetric())

	assert.Equal(t, "m1-suff", processed[0].Name(), "Suffix was not applied")
}
