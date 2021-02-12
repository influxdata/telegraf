package models

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		DropOriginal: true,
		Period:       time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := &testutil.Accumulator{}

	now := time.Now()
	ra.updateWindow(now, now.Add(ra.Config.Period))

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*150),
		telegraf.Untyped)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)
	ra.Push(acc)

	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, int64(101), acc.Metrics[0].Fields["sum"])
}

func TestAddWithoutDrop(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		DropOriginal: false,
		Period:       time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := &testutil.Accumulator{}

	now := time.Now()
	ra.updateWindow(now, now.Add(ra.Config.Period))

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*150),
		telegraf.Untyped)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)
	ra.Push(acc)

	require.Equal(t, 2, len(acc.Metrics))
	metric := acc.Metrics[0]
	if _, ok := metric.Fields["sum"]; !ok {
		metric = acc.Metrics[1]
	}
	require.Equal(t, int64(101), metric.Fields["sum"])
}

func TestAddMetricsOutsideCurrentPeriod(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		DropOriginal: true,
		Period:       time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := &testutil.Accumulator{}
	now := time.Now()
	ra.updateWindow(now, now.Add(ra.Config.Period))

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now.Add(-time.Hour),
		telegraf.Untyped,
	)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)

	// metric after current period
	m = testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now.Add(time.Hour),
		telegraf.Untyped,
	)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)

	// "now" metric
	m = testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*50),
		telegraf.Untyped)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)

	ra.Push(acc)
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, int64(101), acc.Metrics[0].Fields["sum"])
}

func TestAddMetricsOutsideCurrentPeriodWithGrace(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period:       time.Millisecond * 1500,
		Grace:        time.Millisecond * 500,
		DropOriginal: true,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := &testutil.Accumulator{}
	now := time.Now()
	ra.updateWindow(now, now.Add(ra.Config.Period))

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now.Add(-time.Hour),
		telegraf.Untyped,
	)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)

	// metric before current period (late)
	m = testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(100),
		},
		now.Add(-time.Millisecond*1000),
		telegraf.Untyped,
	)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)

	// metric before current period, but within grace period (late)
	m = testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(102),
		},
		now.Add(-time.Millisecond*200),
		telegraf.Untyped,
	)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)

	// "now" metric
	m = testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*50),
		telegraf.Untyped)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)

	ra.Push(acc)
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, int64(203), acc.Metrics[0].Fields["sum"])
}

func TestAddAndPushOnePeriod(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := &testutil.Accumulator{}

	now := time.Now()
	ra.updateWindow(now, now.Add(ra.Config.Period))

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*100),
		telegraf.Untyped)
	require.NoError(t, ra.Add(m, acc))
	require.Len(t, m.FieldList(), 1)

	ra.Push(acc)

	acc.AssertContainsFields(t, "TestMetric", map[string]interface{}{"sum": int64(101)})
}

func TestAddDropOriginal(t *testing.T) {
	ra := NewRunningAggregator(&TestAggregator{}, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"RI*"},
		},
		DropOriginal: true,
	})
	require.NoError(t, ra.Config.Filter.Compile())

	now := time.Now()
	ra.updateWindow(now, now.Add(ra.Config.Period))

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	require.NoError(t, ra.Add(m, &testutil.Accumulator{}))
	require.Len(t, m.FieldList(), 1)

	// this metric name doesn't match the filter, so Add will return false
	m2 := testutil.MustMetric("foobar",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	require.NoError(t, ra.Add(m2, &testutil.Accumulator{}))
	require.Len(t, m.FieldList(), 1)
}

func TestAddDoesNotModifyMetric(t *testing.T) {
	ra := NewRunningAggregator(&TestAggregator{}, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			FieldPass: []string{"a"},
		},
		DropOriginal: true,
	})
	require.NoError(t, ra.Config.Filter.Compile())

	now := time.Now()

	m := testutil.MustMetric(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"a": int64(42),
			"b": int64(42),
		},
		now)
	expected := m.Copy()
	ra.Add(m, &testutil.Accumulator{})

	testutil.RequireMetricEqual(t, expected, m)
}

type TestAggregator struct {
	sum int64
}

func (t *TestAggregator) Description() string  { return "" }
func (t *TestAggregator) SampleConfig() string { return "" }
func (t *TestAggregator) Reset() {
	t.sum = 0
}

func (t *TestAggregator) Push(acc telegraf.Accumulator) {
	acc.AddFields("TestMetric",
		map[string]interface{}{"sum": t.sum},
		map[string]string{},
	)
}

func (t *TestAggregator) Add(in telegraf.Metric) {
	for _, v := range in.Fields() {
		if vi, ok := v.(int64); ok {
			t.sum += vi
		}
	}
}
