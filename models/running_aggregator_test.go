package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestRunningAggregatorAdd(t *testing.T) {
	a := &mockAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}

	now := time.Now()
	ra.UpdateWindow(now, now.Add(ra.Config.Period))

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*150),
		telegraf.Untyped)
	require.False(t, ra.Add(m))
	ra.Push(&acc)

	require.Len(t, acc.Metrics, 1)
	require.Equal(t, int64(101), acc.Metrics[0].Fields["sum"])
}

func TestRunningAggregatorAddMetricsOutsideCurrentPeriod(t *testing.T) {
	a := &mockAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}
	now := time.Now()
	ra.UpdateWindow(now, now.Add(ra.Config.Period))

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now.Add(-time.Hour),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// metric after current period
	m = metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now.Add(time.Hour),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// "now" metric
	m = metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*50),
		telegraf.Untyped)
	require.False(t, ra.Add(m))

	ra.Push(&acc)
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, int64(101), acc.Metrics[0].Fields["sum"])
}

func TestRunningAggregatorAddMetricsOutsideCurrentPeriodWithGrace(t *testing.T) {
	a := &mockAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 1500,
		Grace:  time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}
	now := time.Now()
	ra.UpdateWindow(now, now.Add(ra.Config.Period))

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now.Add(-time.Hour),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// metric before current period (late)
	m = metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(100),
		},
		now.Add(-time.Millisecond*1000),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// metric before current period, but within grace period (late)
	m = metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(102),
		},
		now.Add(-time.Millisecond*200),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// "now" metric
	m = metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*50),
		telegraf.Untyped)
	require.False(t, ra.Add(m))

	ra.Push(&acc)
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, int64(203), acc.Metrics[0].Fields["sum"])
}

func TestRunningAggregatorAddAndPushOnePeriod(t *testing.T) {
	a := &mockAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}

	now := time.Now()
	ra.UpdateWindow(now, now.Add(ra.Config.Period))

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*100),
		telegraf.Untyped)
	require.False(t, ra.Add(m))

	ra.Push(&acc)

	acc.AssertContainsFields(t, "TestMetric", map[string]interface{}{"sum": int64(101)})
}

func TestRunningAggregatorAddDropOriginal(t *testing.T) {
	ra := NewRunningAggregator(&mockAggregator{}, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"RI*"},
		},
		DropOriginal: true,
	})
	require.NoError(t, ra.Config.Filter.Compile())

	now := time.Now()
	ra.UpdateWindow(now, now.Add(ra.Config.Period))

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	require.True(t, ra.Add(m))

	// this metric name doesn't match the filter, so Add will return false
	m2 := metric.New("foobar",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	require.False(t, ra.Add(m2))
}

func TestRunningAggregatorAddDoesNotModifyMetric(t *testing.T) {
	ra := NewRunningAggregator(&mockAggregator{}, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			FieldInclude: []string{"a"},
		},
		DropOriginal: true,
	})
	require.NoError(t, ra.Config.Filter.Compile())

	now := time.Now()

	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"a": int64(42),
			"b": int64(42),
		},
		now)
	expected := m.Copy()
	ra.Add(m)

	testutil.RequireMetricEqual(t, expected, m)
}

func TestRunningAggregatorPushAdvancesWindowWhenFiringEarly(t *testing.T) {
	ra := NewRunningAggregator(&mockAggregator{}, &AggregatorConfig{
		Name:   "TestRunningAggregator",
		Filter: Filter{NamePass: []string{"*"}},
		Period: time.Second,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}

	// Simulate a timer firing slightly before periodEnd: set periodEnd
	// 500ms in the future so time.Now() inside Push falls within the
	// current window. Before the drift safeguard was loosened, this
	// tripped the reset and left periodEnd unchanged, causing a second
	// Push for the same period.
	windowEnd := time.Now().Add(500 * time.Millisecond)
	ra.UpdateWindow(windowEnd.Add(-ra.Config.Period), windowEnd)

	ra.Push(&acc)

	require.True(t, ra.EndPeriod().Equal(windowEnd.Add(ra.Config.Period)),
		"expected periodEnd to advance by one period: want %v, got %v",
		windowEnd.Add(ra.Config.Period), ra.EndPeriod())
}

func TestRunningAggregatorPushResetsWindowOnLargeForwardJump(t *testing.T) {
	// Regression guard for PR #16375: the drift safeguard must still
	// reset the window when the wall clock is more than one period
	// beyond the expected window, such as after hibernation.
	ra := NewRunningAggregator(&mockAggregator{}, &AggregatorConfig{
		Name:   "TestRunningAggregator",
		Filter: Filter{NamePass: []string{"*"}},
		Period: time.Second,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}

	staleEnd := time.Now().Add(-2 * time.Minute)
	ra.UpdateWindow(staleEnd.Add(-ra.Config.Period), staleEnd)

	ra.Push(&acc)

	require.Less(t, time.Since(ra.EndPeriod()).Abs(), 2*ra.Config.Period,
		"expected safeguard to reset window near current time, got %v", ra.EndPeriod())
}

type mockAggregator struct {
	sum int64
}

func (*mockAggregator) SampleConfig() string {
	return ""
}

func (t *mockAggregator) Reset() {
	t.sum = 0
}

func (t *mockAggregator) Push(acc telegraf.Accumulator) {
	acc.AddFields("TestMetric",
		map[string]interface{}{"sum": t.sum},
		map[string]string{},
	)
}

func (t *mockAggregator) Add(in telegraf.Metric) {
	for _, v := range in.Fields() {
		if vi, ok := v.(int64); ok {
			t.sum += vi
		}
	}
}
