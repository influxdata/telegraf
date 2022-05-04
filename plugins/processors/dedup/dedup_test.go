package dedup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
)

const metricName = "m1"

func createMetric(value int64, when time.Time) telegraf.Metric {
	m := metric.New(metricName,
		map[string]string{"tag": "tag_value"},
		map[string]interface{}{"value": value},
		when,
	)
	return m
}

func createDedup(initTime time.Time) Dedup {
	return Dedup{
		DedupInterval: config.Duration(10 * time.Minute),
		FlushTime:     initTime,
		Cache:         make(map[uint64]telegraf.Metric),
	}
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
	require.Equal(t, metricName, target[0].Name())
	// target metric has proper field
	tValue, present := target[0].GetField("value")
	require.True(t, present)
	sValue, present := source.GetField("value")
	require.True(t, present)
	require.Equal(t, tValue, sValue)
	// target metric has proper timestamp
	require.Equal(t, target[0].Time(), source.Time())
}

func assertMetricSuppressed(t *testing.T, target []telegraf.Metric) {
	// target is empty
	require.Equal(t, 0, len(target))
}

func TestProcRetainsMetric(t *testing.T) {
	deduplicate := createDedup(time.Now())
	source := createMetric(1, time.Now())
	target := deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestSuppressRepeatedValue(t *testing.T) {
	deduplicate := createDedup(time.Now())
	// Create metric in the past
	source := createMetric(1, time.Now().Add(-1*time.Second))
	_ = deduplicate.Apply(source)
	source = createMetric(1, time.Now())
	target := deduplicate.Apply(source)

	assertCacheHit(t, &deduplicate, source)
	assertMetricSuppressed(t, target)
}

func TestPassUpdatedValue(t *testing.T) {
	deduplicate := createDedup(time.Now())
	// Create metric in the past
	source := createMetric(1, time.Now().Add(-1*time.Second))
	target := deduplicate.Apply(source)
	assertMetricPassed(t, target, source)

	source = createMetric(2, time.Now())
	target = deduplicate.Apply(source)
	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestPassAfterCacheExpire(t *testing.T) {
	deduplicate := createDedup(time.Now())
	// Create metric in the past
	source := createMetric(1, time.Now().Add(-1*time.Hour))
	target := deduplicate.Apply(source)
	assertMetricPassed(t, target, source)

	source = createMetric(1, time.Now())
	target = deduplicate.Apply(source)
	assertCacheRefresh(t, &deduplicate, source)
	assertMetricPassed(t, target, source)
}

func TestCacheRetainsMetrics(t *testing.T) {
	deduplicate := createDedup(time.Now())
	// Create metric in the past 3sec
	source := createMetric(1, time.Now().Add(-3*time.Hour))
	deduplicate.Apply(source)
	// Create metric in the past 2sec
	source = createMetric(1, time.Now().Add(-2*time.Hour))
	deduplicate.Apply(source)
	source = createMetric(1, time.Now())
	deduplicate.Apply(source)

	assertCacheRefresh(t, &deduplicate, source)
}

func TestCacheShrink(t *testing.T) {
	// Time offset is more than 2 * DedupInterval
	deduplicate := createDedup(time.Now().Add(-2 * time.Hour))
	// Time offset is more than 1 * DedupInterval
	source := createMetric(1, time.Now().Add(-1*time.Hour))
	deduplicate.Apply(source)

	require.Equal(t, 0, len(deduplicate.Cache))
}

func TestSameTimestamp(t *testing.T) {
	now := time.Now()
	dedup := createDedup(now)
	var in telegraf.Metric
	var out []telegraf.Metric

	in = metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"foo": 1}, // field
		now,
	)
	out = dedup.Apply(in)
	require.Equal(t, []telegraf.Metric{in}, out) // pass

	in = metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"bar": 1}, // different field
		now,
	)
	out = dedup.Apply(in)
	require.Equal(t, []telegraf.Metric{in}, out) // pass

	in = metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"bar": 2}, // same field different value
		now,
	)
	out = dedup.Apply(in)
	require.Equal(t, []telegraf.Metric{in}, out) // pass

	in = metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"bar": 2}, // same field same value
		now,
	)
	out = dedup.Apply(in)
	require.Equal(t, []telegraf.Metric{}, out) // drop
}

func TestSuppressMultipleRepeatedValue(t *testing.T) {
	deduplicate := createDedup(time.Now())
	// Create metric in the past
	source := createMetric(1, time.Now().Add(-1*time.Second))
	_ = deduplicate.Apply(source)
	source = createMetric(1, time.Now())
	target := deduplicate.Apply(source, source, source, source)

	assertCacheHit(t, &deduplicate, source)
	assertMetricSuppressed(t, target)
}
