package models

import (
	"sync"
	"sync/atomic"
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
		map[string]any{
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
		map[string]any{
			"value": int64(101),
		},
		now.Add(-time.Hour),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// metric after current period
	m = metric.New("RITest",
		map[string]string{},
		map[string]any{
			"value": int64(101),
		},
		now.Add(time.Hour),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// "now" metric
	m = metric.New("RITest",
		map[string]string{},
		map[string]any{
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
		map[string]any{
			"value": int64(101),
		},
		now.Add(-time.Hour),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// metric before current period (late)
	m = metric.New("RITest",
		map[string]string{},
		map[string]any{
			"value": int64(100),
		},
		now.Add(-time.Millisecond*1000),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// metric before current period, but within grace period (late)
	m = metric.New("RITest",
		map[string]string{},
		map[string]any{
			"value": int64(102),
		},
		now.Add(-time.Millisecond*200),
		telegraf.Untyped,
	)
	require.False(t, ra.Add(m))

	// "now" metric
	m = metric.New("RITest",
		map[string]string{},
		map[string]any{
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
		map[string]any{
			"value": int64(101),
		},
		time.Now().Add(time.Millisecond*100),
		telegraf.Untyped)
	require.False(t, ra.Add(m))

	ra.Push(&acc)

	acc.AssertContainsFields(t, "TestMetric", map[string]any{"sum": int64(101)})
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
		map[string]any{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	require.True(t, ra.Add(m))

	// this metric name doesn't match the filter, so Add will return false
	m2 := metric.New("foobar",
		map[string]string{},
		map[string]any{
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
		map[string]any{
			"a": int64(42),
			"b": int64(42),
		},
		now)
	expected := m.Copy()
	ra.Add(m)

	testutil.RequireMetricEqual(t, expected, m)
}

func TestRunningAggregatorPushResetsWindowOnLargeForwardJump(t *testing.T) {
	// Regression guard for PR #16375: the drift safeguard must reset the
	// window when the wall clock is beyond the expected window, such as
	// after hibernation.
	ra := NewRunningAggregator(&mockAggregator{}, &AggregatorConfig{
		Name:   "TestRunningAggregator",
		Filter: Filter{NamePass: []string{"*"}},
		Period: 10 * time.Second,
	})
	require.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}

	staleEnd := time.Now().Add(-2 * time.Minute)
	ra.UpdateWindow(staleEnd.Add(-ra.Config.Period), staleEnd)

	ra.Push(&acc)

	// After a reset the new window end lands in (now, now+period].
	require.WithinDuration(t, time.Now(), ra.EndPeriod(), ra.Config.Period)
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
		map[string]any{"sum": t.sum},
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

// slowPushAggregator produces a configurable number of metrics on Push and
// blocks for a configurable duration while doing so, simulating a
// high-cardinality aggregation (e.g. many distinct tag combinations) that
// takes a while to serialize.
type slowPushAggregator struct {
	metricsToPush int
	pushDuration  time.Duration
}

func (*slowPushAggregator) SampleConfig() string { return "" }
func (*slowPushAggregator) Reset()               {}
func (*slowPushAggregator) Add(telegraf.Metric)  {}

func (a *slowPushAggregator) Push(acc telegraf.Accumulator) {
	perMetric := time.Duration(0)
	if a.metricsToPush > 0 {
		perMetric = a.pushDuration / time.Duration(a.metricsToPush)
	}
	for i := 0; i < a.metricsToPush; i++ {
		time.Sleep(perMetric)
		acc.AddFields("slow", map[string]any{"value": int64(i)}, map[string]string{})
	}
}

// blockingAccumulator embeds testutil.Accumulator but blocks every call that
// would normally deliver a metric downstream, until release() is called.
// This stands in for a bounded channel to a slow/busy next stage.
type blockingAccumulator struct {
	testutil.Accumulator
	blockUntil chan struct{}
}

func newBlockingAccumulator() *blockingAccumulator {
	return &blockingAccumulator{blockUntil: make(chan struct{})}
}

func (a *blockingAccumulator) release() {
	close(a.blockUntil)
}

func (a *blockingAccumulator) AddMetric(m telegraf.Metric) {
	<-a.blockUntil
	a.Accumulator.AddMetric(m)
}

func (a *blockingAccumulator) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {
	<-a.blockUntil
	a.Accumulator.AddFields(measurement, fields, tags, t...)
}

// TestRunningAggregatorAddNotBlockedByPush verifies that Add() can proceed
// while Push() is still delivering the previous period's metrics downstream,
// even if that delivery is blocked for a while. Before the fix, Push() held
// the aggregator's lock for the entire delivery, so a slow or momentarily
// full downstream stage (e.g. an output straining under a large,
// high-cardinality aggregation) would stall every concurrent Add() call --
// and, transitively, any input plugin (such as influxdb_listener) blocked
// delivering a metric through Add(). That, in turn, could make upstream
// clients experience request timeouts and silently lose metrics that never
// made it into Telegraf, without Telegraf itself ever logging a drop.
func TestRunningAggregatorAddNotBlockedByPush(t *testing.T) {
	agg := &slowPushAggregator{metricsToPush: 50, pushDuration: 200 * time.Millisecond}
	ra := NewRunningAggregator(agg, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period:       time.Minute,
		DropOriginal: false,
	})
	require.NoError(t, ra.Config.Filter.Compile())

	now := time.Now()
	ra.UpdateWindow(now, now.Add(ra.Config.Period))

	acc := newBlockingAccumulator()

	var pushWG sync.WaitGroup
	pushWG.Add(1)
	go func() {
		defer pushWG.Done()
		ra.Push(acc)
	}()

	// Give Push() a head start so it is actively running (and, before the
	// fix, holding the lock) when we call Add() below.
	time.Sleep(20 * time.Millisecond)

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{"value": int64(1)},
		now.Add(time.Second),
		telegraf.Untyped)

	addDone := make(chan bool, 1)
	go func() {
		addDone <- ra.Add(m)
	}()

	select {
	case dropOriginal := <-addDone:
		require.False(t, dropOriginal)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Add() did not return while Push() was delivering metrics; " +
			"it is stuck behind the aggregator's lock")
	}

	// Let Push() finish delivering so the goroutine and test can exit cleanly.
	acc.release()
	pushWG.Wait()

	require.Len(t, acc.Metrics, 50)
}

// TestRunningAggregatorPushDeliversAllMetricsWithoutHoldingLock is a
// regression test ensuring the lock-decoupling refactor of Push() still
// delivers every metric produced by the aggregator, even when downstream
// delivery is slow.
func TestRunningAggregatorPushDeliversAllMetricsWithoutHoldingLock(t *testing.T) {
	const metricsToPush = 25
	agg := &slowPushAggregator{metricsToPush: metricsToPush, pushDuration: 50 * time.Millisecond}
	ra := NewRunningAggregator(agg, &AggregatorConfig{
		Name:   "TestRunningAggregator",
		Filter: Filter{NamePass: []string{"*"}},
		Period: time.Minute,
	})
	require.NoError(t, ra.Config.Filter.Compile())

	now := time.Now()
	ra.UpdateWindow(now, now.Add(ra.Config.Period))

	acc := newBlockingAccumulator()
	var delivered atomic.Int32
	done := make(chan struct{})
	go func() {
		ra.Push(acc)
		close(done)
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		acc.release()
	}()

	<-done
	delivered.Store(int32(len(acc.Metrics)))
	require.EqualValues(t, metricsToPush, delivered.Load())
}
