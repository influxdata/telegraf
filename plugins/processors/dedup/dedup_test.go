package dedup

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
)

func createMetric(name string, value int64, when time.Time) telegraf.Metric {
	m, _ := metric.New(name,
		map[string]string{"tag": "tag_value"},
		map[string]interface{}{"value": value},
		when,
	)
	return m
}

func assertCacheRefresh(t *testing.T, proc *Dedup, item telegraf.Metric) {
	id := item.HashID()
	name := item.Name()
	// cache is not empty
	require.NotEqual(t, 0, len(proc.Cache))
	// cache has metric with proper id
	cache, present := proc.Cache[id]
	require.True(t, present)
	// cache has metric with proper name
	require.Equal(t, name, cache.Name())
	// cached metric has proper field
	cValue, present := cache.GetField("value")
	require.True(t, present)
	iValue, _ := item.GetField("value")
	require.Equal(t, cValue, iValue)
	// cached metric has proper timestamp
	require.Equal(t, cache.Time(), item.Time())
}

func assertCacheHit(t *testing.T, proc *Dedup, item telegraf.Metric) {
	id := item.HashID()
	name := item.Name()
	// cache is not empty
	require.NotEqual(t, 0, len(proc.Cache))
	// cache has metric with proper id
	cache, present := proc.Cache[id]
	require.True(t, present)
	// cache has metric with proper name
	require.Equal(t, name, cache.Name())
	// cached metric has proper field
	cValue, present := cache.GetField("value")
	require.True(t, present)
	iValue, _ := item.GetField("value")
	require.Equal(t, cValue, iValue)
	// cached metric did NOT change timestamp
	require.NotEqual(t, cache.Time(), item.Time())
}

func assertMetricPassed(t *testing.T, target []telegraf.Metric, source telegraf.Metric) {
	// target is not empty
	require.NotEqual(t, 0, len(target))
	// target has metric with proper name
	require.Equal(t, "m1", target[0].Name())
	// target metric has proper field
	tValue, present := target[0].GetField("value")
	require.True(t, present)
	sValue, present := source.GetField("value")
	require.Equal(t, tValue, sValue)
	// target metric has proper timestamp
	require.Equal(t, target[0].Time(), source.Time())
}

func assertMetricSuppressed(t *testing.T, target []telegraf.Metric, source telegraf.Metric) {
	// target is empty
	require.Equal(t, 0, len(target))
}

func TestProcRetainsMetric(t *testing.T) {
	deduplicate := *NewDedup()
	source := createMetric("m1", 1, time.Now())
	target := deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestSuppressRepeatedValue(t *testing.T) {
	deduplicate := *NewDedup()
	// Create metric in the past
	source := createMetric("m1", 1, time.Now().Add(-1*time.Second))
	target := deduplicate.Apply(source)
	source = createMetric("m1", 1, time.Now())
	target = deduplicate.Apply(source)

	assertCacheHit(t, &deduplicate, source)
	assertMetricSuppressed(t, target, source)
}

func TestPassUpdatedValue(t *testing.T) {
	deduplicate := *NewDedup()
	// Create metric in the past
	source := createMetric("m1", 1, time.Now().Add(-1*time.Second))
	target := deduplicate.Apply(source)
	source = createMetric("m1", 2, time.Now())
	target = deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestPassAfterCacheExpire(t *testing.T) {
	deduplicate := *NewDedup()
	// Create metric in the past
	source := createMetric("m1", 1, time.Now().Add(-1*time.Hour))
	target := deduplicate.Apply(source)
	source = createMetric("m1", 1, time.Now())
	target = deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestCacheRetainsMetrics(t *testing.T) {
	deduplicate := *NewDedup()
	// Create metric in the past 3sec
	source := createMetric("m1", 1, time.Now().Add(-3*time.Hour))
	deduplicate.Apply(source)
	// Create metric in the past 2sec
	source = createMetric("m1", 1, time.Now().Add(-2*time.Hour))
	deduplicate.Apply(source)
	source = createMetric("m1", 1, time.Now())
	deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
}

func TestCacheShrink(t *testing.T) {
	deduplicate := &Dedup{
		DedupInterval: internal.Duration{Duration: 10 * time.Minute},
		FlushTime:     time.Now().Add(-2 * time.Hour),
		Cache:         make(map[uint64]telegraf.Metric),
	}
	source := createMetric("m1", 1, time.Now().Add(-1*time.Hour))
	deduplicate.Apply(source)

	require.Equal(t, 0, len(deduplicate.Cache))
}
