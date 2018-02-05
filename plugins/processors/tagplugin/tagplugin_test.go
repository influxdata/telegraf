package tagplugin

import (
	"testing"
	"time"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

// Test cases
var m0, _ = metric.New("m0",
	map[string]string{
		"host": "localhost",
		"cpu": "cpu0",
	},
	map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	},
	time.Now(),
)

var m1, _ = metric.New("m1",
	map[string]string{
		"host": "localhost",
		"cpu": "cpu1",
	},
	map[string]interface{}{
		"usage_idle": float64(50),
		"usage_busy": float64(50),
	},
	time.Now(),
)

var m2, _ = metric.New("m2",
	map[string]string{
		"host": "localhost",
		"cpu": "cpu2",
	},
	map[string]interface{}{
		"usage_idle": float64(1),
		"usage_busy": float64(99),
	},
	time.Now(),
)

var m3, _ = metric.New("m3",
	map[string]string{
		"host": "localhost",
		"cpu": "cpu3",
	},
	map[string]interface{}{
		"usage_idle": float64(70),
		"usage_busy": float64(30),
	},
	time.Now(),
)

func TestApply(t *testing.T) {
	tagplugin := newTagPlugin()
	tagplugin.ReferenceTagName = "cpu"
	tagplugin.NewTagName = "cpu_category"
	tagplugin.NewTagValueMap = map[string][]string{"system":{"cpu0","cpu1"}, "user":{"cpu2"}}
	tagplugin.NewTagDefaultValue = "vm"

	taggedMetrics := tagplugin.Apply(m0, m1, m2, m3)

	// Check that cpus get the tag they should have
	assert.Equal(t, "system", taggedMetrics[0].Tags()["cpu_category"])
	assert.Equal(t, "system", taggedMetrics[1].Tags()["cpu_category"])
	assert.Equal(t, "user", taggedMetrics[2].Tags()["cpu_category"])

	// Check that cpus not listed in the config gets the default tag
	assert.Equal(t, "vm", taggedMetrics[3].Tags()["cpu_category"])
}

func TestApplyNoNewTagDefaultValue(t *testing.T) {
	tagplugin := newTagPlugin()
	tagplugin.ReferenceTagName = "cpu"
	tagplugin.NewTagName = "cpu_category"
	tagplugin.NewTagValueMap = map[string][]string{"system":{"cpu0","cpu1"}, "user":{"cpu2"}}

	// Reset to no tags
	m0.RemoveTag("cpu_category")
	m1.RemoveTag("cpu_category")
	m2.RemoveTag("cpu_category")
	m3.RemoveTag("cpu_category")

	taggedMetrics := tagplugin.Apply(m0, m1, m2, m3)

	// Check that cpus get the tag they should have
	assert.Equal(t, "system", taggedMetrics[0].Tags()["cpu_category"])
	assert.Equal(t, "system", taggedMetrics[1].Tags()["cpu_category"])
	assert.Equal(t, "user", taggedMetrics[2].Tags()["cpu_category"])

	// Check that the cpu not listed in the config doesn't get a tag
	assert.False(t, taggedMetrics[3].HasTag("cpu_category"))
}

func TestApplyNoNewTagValueMap(t *testing.T) {
	tagplugin := newTagPlugin()
	tagplugin.ReferenceTagName = "cpu"
	tagplugin.NewTagName = "cpu_category"
	tagplugin.NewTagDefaultValue = "vm"

	// Reset to no tags
	m0.RemoveTag("cpu_category")
	m1.RemoveTag("cpu_category")
	m2.RemoveTag("cpu_category")
	m3.RemoveTag("cpu_category")

	taggedMetrics := tagplugin.Apply(m0, m1, m2, m3)

	// Check that all cpus get the default tag
	assert.Equal(t, "vm", taggedMetrics[0].Tags()["cpu_category"])
	assert.Equal(t, "vm", taggedMetrics[1].Tags()["cpu_category"])
	assert.Equal(t, "vm", taggedMetrics[2].Tags()["cpu_category"])
	assert.Equal(t, "vm", taggedMetrics[3].Tags()["cpu_category"])
}

func TestApplyNoNewTagValueMapNoNewTagDefaultValue(t *testing.T) {
	tagplugin := newTagPlugin()
	tagplugin.ReferenceTagName = "cpu"
	tagplugin.NewTagName = "cpu_category"

	// Reset to no tags
	m0.RemoveTag("cpu_category")
	m1.RemoveTag("cpu_category")
	m2.RemoveTag("cpu_category")
	m3.RemoveTag("cpu_category")

	taggedMetrics := tagplugin.Apply(m0, m1, m2, m3)

	// Check that cpus don't get tags when no config
	assert.False(t, taggedMetrics[0].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[1].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[2].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[3].HasTag("cpu_category"))
}

func TestApplyNoNewTagName(t *testing.T) {
	tagplugin := newTagPlugin()
	tagplugin.ReferenceTagName = "cpu"
	tagplugin.NewTagValueMap = map[string][]string{"system":{"cpu0","cpu1"}, "user":{"cpu2"}}
	tagplugin.NewTagDefaultValue = "vm"

	// Reset to no tags
	m0.RemoveTag("cpu_category")
	m1.RemoveTag("cpu_category")
	m2.RemoveTag("cpu_category")
	m3.RemoveTag("cpu_category")

	taggedMetrics := tagplugin.Apply(m0, m1, m2, m3)

	// Check that cpus don't get tags when no new tag name
	assert.False(t, taggedMetrics[0].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[1].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[2].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[3].HasTag("cpu_category"))
}

func TestApplyNoReferenceTagName(t *testing.T) {
	tagplugin := newTagPlugin()
	tagplugin.NewTagName = "cpu"
	tagplugin.NewTagValueMap = map[string][]string{"system":{"cpu0","cpu1"}, "user":{"cpu2"}}
	tagplugin.NewTagDefaultValue = "vm"

	// Reset to no tags
	m0.RemoveTag("cpu_category")
	m1.RemoveTag("cpu_category")
	m2.RemoveTag("cpu_category")
	m3.RemoveTag("cpu_category")

	taggedMetrics := tagplugin.Apply(m0, m1, m2, m3)

	// Check that cpus don't get tags when no reference tag name
	assert.False(t, taggedMetrics[0].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[1].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[2].HasTag("cpu_category"))
	assert.False(t, taggedMetrics[3].HasTag("cpu_category"))
}