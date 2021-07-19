package stepped

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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

func createStepped(initTime time.Time, fields []string) Stepped {
	result := Stepped{
		Fields:         fields,
		FlushTime:      initTime,
		StepOffset:     "1ns",
		RetainInterval: internal.Duration{Duration: 30 * 24 * time.Hour},
		Cache:          make(map[uint64]telegraf.Metric),
	}

	err := result.Init()

	if err != nil {
		panic(err)
	}

	return result
}

// Check if field list contains a key
func assertFieldListContainsKey(t *testing.T, fieldList []*telegraf.Field, searchterm string) {
	found := false
	for _, f := range fieldList {
		if searchterm == f.Key {
			found = true
			break
		}
	}
	require.True(t, found)
}

func assertCacheRefresh(t *testing.T, proc *Stepped, item telegraf.Metric) {
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

func assertCacheHit(t *testing.T, proc *Stepped, item telegraf.Metric) {
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

func assertMetricStepped(t *testing.T, proc *Stepped, items []telegraf.Metric) {
	// At Least 2 metrics
	require.Greater(t, len(items), 1)

	firstMetric := items[0]
	secondMetric := items[1]

	// Check Metric Name & Tags
	require.Equal(t, firstMetric.Name(), secondMetric.Name(), "Metric names should match")
	require.Equal(t, firstMetric.TagList(), secondMetric.TagList(), "Metric tags should match")

	// Check Fields Don't Match
	require.NotEqual(t, firstMetric.FieldList(), secondMetric.FieldList(), "Metric fields should not match")

	// Check for Field Unique
	for _, f := range proc.Fields {
		assertFieldListContainsKey(t, firstMetric.FieldList(), f)
		assertFieldListContainsKey(t, secondMetric.FieldList(), f)
	}

	// Check Timestamp off by offset
	firstMetricTime := secondMetric.Time().Add(proc.Dur)
	require.True(t, firstMetricTime.Equal(firstMetric.Time()))
}

func assertNotMetricStepped(t *testing.T, proc *Stepped, items []telegraf.Metric) {
	// At Least 2 metrics
	require.LessOrEqual(t, len(items), 1)
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

func TestStepped_ProcRetainsMetric(t *testing.T) {
	stepped := createStepped(time.Now(), []string{"value"})
	source := createMetric("m1", 1, time.Now())
	target := stepped.Apply(source)

	assertCacheRefresh(t, &stepped, source)
	assertMetricPassed(t, target, source)
}

func TestStepped_RepeatedValueUpdatesCache(t *testing.T) {
	stepped := createStepped(time.Now(), []string{"value"})
	// Create metric in the past
	source := createMetric("m1", 1, time.Now().Add(-1*time.Second))
	target := stepped.Apply(source)
	source = createMetric("m1", 1, time.Now())
	target = stepped.Apply(source)

	assertCacheRefresh(t, &stepped, source)
	assertMetricPassed(t, target, source)
}

func TestStepped_SingleStepped(t *testing.T) {
	stepped := createStepped(time.Now(), []string{"value"})
	ts1 := createMetric("m1", 1, time.Now())
	target := stepped.Apply(ts1)

	assertCacheRefresh(t, &stepped, ts1)

	ts2 := createMetric("m1", 2, time.Now())
	target = stepped.Apply(ts2)

	assertCacheRefresh(t, &stepped, ts2)
	assertMetricStepped(t, &stepped, target)
}

func TestStepped_PassAfterCacheExpire(t *testing.T) {
	stepped := createStepped(time.Now(), []string{"value"})
	// Create metric in the past
	source := createMetric("m1", 1, time.Now())
	target := stepped.Apply(source)

	source = createMetric("m1", 1, time.Now().Add(-90*24*time.Hour))
	target = stepped.Apply(source)
	assertMetricPassed(t, target, source)

	source = createMetric("m1", 2, time.Now())
	target = stepped.Apply(source)

	assertNotMetricStepped(t, &stepped, target)
}

func TestStepped_CacheRetainsMetrics(t *testing.T) {
	stepped := createStepped(time.Now(), []string{"value"})
	// Create metric in the past 3sec
	source := createMetric("m1", 1, time.Now().Add(-3*time.Hour))
	stepped.Apply(source)
	// Create metric in the past 2sec
	source = createMetric("m1", 1, time.Now().Add(-2*time.Hour))
	stepped.Apply(source)
	source = createMetric("m1", 1, time.Now())
	stepped.Apply(source)

	assertCacheRefresh(t, &stepped, source)
}

func TestStepped_CacheShrink(t *testing.T) {
	// Time offset is more than 2 * RetainInterval
	stepped := createStepped(time.Now().Add(-60*24*time.Hour), []string{"value"})
	// Time offset is more than 1 * RetainInterval
	source := createMetric("m1", 1, time.Now().Add(-30*24*time.Hour))
	stepped.Apply(source)

	require.Equal(t, 0, len(stepped.Cache))
}

func TestStepped_SameTimestamp(t *testing.T) {
	now := time.Now()
	stepped := createStepped(now, []string{"foo"})
	var in telegraf.Metric
	var out []telegraf.Metric

	in, _ = metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"foo": 1}, // field
		now,
	)
	out = stepped.Apply(in)
	require.Equal(t, []telegraf.Metric{in}, out) // pass

	in, _ = metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"bar": 1}, // different field
		now,
	)
	out = stepped.Apply(in)
	require.Equal(t, []telegraf.Metric{in}, out) // pass

	in, _ = metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"bar": 2}, // same field different value
		now,
	)
	out = stepped.Apply(in)
	require.Equal(t, []telegraf.Metric{in}, out) // pass

	in, _ = metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"bar": 2}, // same field same value
		now,
	)
	out = stepped.Apply(in)
	require.Equal(t, []telegraf.Metric{in}, out) // pass
}

func TestStepped_LotsOfTagsAndFields(t *testing.T) {
	now := time.Now()
	stepped := createStepped(time.Now(), []string{"value"})
	tags := make(map[string]string)
	tags["host"] = "user-Virtual-Machine"
	tags["topic"] = "Site/Area/Line/Status/StateCurrent"
	fields := make(map[string]interface{})
	fields["value"] = "Idle"
	fields["lower"] = 10
	source, _ := metric.New("mqtt_consumer", tags, fields, now.Add(-1*time.Minute))
	stepped.Apply(source)
	fields["value"] = "Starting"
	fields["lower"] = 10

	source, _ = metric.New("mqtt_consumer", tags, fields, now)
	target := stepped.Apply(source)

	assertMetricStepped(t, &stepped, target)

}

func TestStepped_MergedMetric(t *testing.T) {
	t1 := time.Date(2020, 10, 20, 8, 0, 0, 0, time.UTC)
	t2 := time.Date(2020, 10, 20, 9, 0, 0, 0, time.UTC)
	t3 := time.Date(2020, 10, 20, 10, 0, 0, 0, time.UTC)

	stepped := createStepped(t1, []string{"value1", "value2"})

	in1, _ := metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"value1": 1},
		t1,
	)
	in2, _ := metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"value2": 2},
		t2,
	)
	in3, _ := metric.New("metric",
		map[string]string{"tag": "value"},
		map[string]interface{}{"value1": 2},
		t3,
	)
	target := stepped.Apply(in1)
	target = stepped.Apply(in2)

	// Assert Cache Contains both values
	id := in1.HashID()
	name := in1.Name()
	require.Equal(t, id, in2.HashID())
	require.Equal(t, name, in2.Name())

	require.Equal(t, 1, len(stepped.Cache))
	// cache has metric with proper id
	cache, present := stepped.Cache[id]
	require.True(t, present)
	// cache has metric with proper name
	require.Equal(t, name, cache.Name())
	// cached metric has proper field
	cValue, present := cache.GetField("value1")
	require.True(t, present)
	iValue, _ := in1.GetField("value1")
	require.Equal(t, cValue, iValue)
	cValue, present = cache.GetField("value2")
	require.True(t, present)
	iValue, _ = in2.GetField("value2")
	require.Equal(t, cValue, iValue)
	// cached metric has proper timestamp
	log.Printf("%s = %s\n", in2.Time(), cache.Time())
	require.Equal(t, cache.Time(), in2.Time())

	target = stepped.Apply(in3)

	// Asset Stepped
	require.Equal(t, 2, len(target))
	cache, present = stepped.Cache[id]
	require.True(t, present)
	// cache has metric with proper name
	require.Equal(t, name, cache.Name())

	require.Equal(t, cache.Time(), in3.Time())
	// cached metric has proper field
	cValue, present = cache.GetField("value1")
	require.True(t, present)
	iValue, _ = in3.GetField("value1")
	require.Equal(t, cValue, iValue)
	cValue, present = cache.GetField("value2")
	require.True(t, present)
	iValue, _ = in2.GetField("value2")
	require.Equal(t, cValue, iValue)

	// Target has old field
	m1 := target[0]
	cValue, present = m1.GetField("value1")
	require.True(t, present)
	iValue, _ = in1.GetField("value1")
	require.Equal(t, cValue, iValue)
}
