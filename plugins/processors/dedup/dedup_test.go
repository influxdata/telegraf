package dedup

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func createMetric(name string, value int64) telegraf.Metric {
	m, _ := metric.New(name,
		map[string]string{"tag": "tag_value"},
		map[string]interface{}{"value": value},
		time.Now(),
	)
	return m
}

func createProcessor() Dedup {
	return Dedup{
		DedupInterval: internal.Duration{Duration: 2 * time.Second},
		FlushTime:     time.Now(),
		Cache:         make(map[uint64]telegraf.Metric),
	}
}

func assertCacheRefresh(t *testing.T, proc *Dedup, item telegraf.Metric) {
	id := item.HashID()
	name := item.Name()
	// cache is not empty
	assert.NotEqual(t, 0, len(proc.Cache))
	// cache has metric with proper id
	cached, present := proc.Cache[id]
	assert.True(t, present)
	// cache has metric with proper name
	assert.Equal(t, name, cached.Name())
	// cached metric has proper field
	value, present := cached.Fields()["value"]
	assert.True(t, present)
	assert.Equal(t, value, item.Fields()["value"])
	// cached metric has proper timestamp
	assert.Equal(t, cached.Time(), item.Time())
}

func assertCacheHit(t *testing.T, proc *Dedup, item telegraf.Metric) {
	id := item.HashID()
	name := item.Name()
	// cache is not empty
	assert.NotEqual(t, 0, len(proc.Cache))
	// cache has metric with proper id
	cached, present := proc.Cache[id]
	assert.True(t, present)
	// cache has metric with proper name
	assert.Equal(t, name, cached.Name())
	// cached metric has proper field
	value, present := cached.Fields()["value"]
	assert.True(t, present)
	assert.Equal(t, value, item.Fields()["value"])
	// cached metric did NOT change timestamp
	assert.NotEqual(t, cached.Time(), item.Time())
}

func assertMetricPassed(t *testing.T, target []telegraf.Metric, source telegraf.Metric) {
	// target is not empty
	assert.NotEqual(t, 0, len(target))
	// target has metric with proper name
	assert.Equal(t, "m1", target[0].Name())
	// target metric has proper field
	value, present := target[0].Fields()["value"]
	assert.True(t, present)
	assert.Equal(t, value, source.Fields()["value"])
	// target metric has proper timestamp
	assert.Equal(t, target[0].Time(), source.Time())
}

func assertMetricSuppressed(t *testing.T, target []telegraf.Metric, source telegraf.Metric) {
	// target is empty
	assert.Equal(t, 0, len(target))
}

func TestProcRetainsMetric(t *testing.T) {
	deduplicate := createProcessor()
	source := createMetric("m1", 1)
	target := deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestSuppressRepeatedValue(t *testing.T) {
	deduplicate := createProcessor()
	source := createMetric("m1", 1)
	target := deduplicate.Apply(source)
	// wait less than deduplication interval
	time.Sleep(1 * time.Second)
	source = createMetric("m1", 1)
	target = deduplicate.Apply(source)

	assertCacheHit(t, &deduplicate, source)
	assertMetricSuppressed(t, target, source)
}

func TestPassUpdatedValue(t *testing.T) {
	deduplicate := createProcessor()
	source := createMetric("m1", 1)
	target := deduplicate.Apply(source)
	// wait less than deduplication interval
	time.Sleep(1 * time.Second)
	source = createMetric("m1", 2)
	target = deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestPassAfterCacheExpire(t *testing.T) {
	deduplicate := createProcessor()
	source := createMetric("m1", 1)
	target := deduplicate.Apply(source)
	// wait more than deduplication interval
	time.Sleep(3 * time.Second)
	source = createMetric("m1", 1)
	target = deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestCacheRetainsMetrics(t *testing.T) {
	deduplicate := createProcessor()
	source := createMetric("m1", 1)
	deduplicate.Apply(source)
	// wait less than deduplication interval
	time.Sleep(1 * time.Second)
	source = createMetric("m1", 1)
	deduplicate.Apply(source)
	// wait same as deduplication interval
	time.Sleep(2 * time.Second)
	source = createMetric("m1", 1)
	deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
}

func TestCacheShrink(t *testing.T) {
	deduplicate := createProcessor()
	src1 := createMetric("m1", 1)
	deduplicate.Apply(src1)
	// wait same as deduplication interval
	time.Sleep(2 * time.Second)
	deduplicate.cleanup()

	assert.Equal(t, 0, len(deduplicate.Cache))
}
