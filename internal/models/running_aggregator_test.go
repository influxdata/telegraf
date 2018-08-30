package models

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	assert.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}
	go ra.Run(&acc, make(chan struct{}))

	m := ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now().Add(time.Millisecond*150),
	)
	assert.False(t, ra.Add(m))

	for {
		time.Sleep(time.Millisecond)
		if atomic.LoadInt64(&a.sum) > 0 {
			break
		}
	}
	assert.Equal(t, int64(101), atomic.LoadInt64(&a.sum))
}

func TestAddMetricsOutsideCurrentPeriod(t *testing.T) {
	a := &TestAggregator{}
	ra := NewRunningAggregator(a, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"*"},
		},
		Period: time.Millisecond * 500,
	})
	assert.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}
	go ra.Run(&acc, make(chan struct{}))

	// metric before current period
	m := ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now().Add(-time.Hour),
	)
	assert.False(t, ra.Add(m))

	// metric after current period
	m = ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now().Add(time.Hour),
	)
	assert.False(t, ra.Add(m))

	// "now" metric
	m = ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now().Add(time.Millisecond*50),
	)
	assert.False(t, ra.Add(m))

	for {
		time.Sleep(time.Millisecond)
		if atomic.LoadInt64(&a.sum) > 0 {
			break
		}
	}
	assert.Equal(t, int64(101), atomic.LoadInt64(&a.sum))
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
	assert.NoError(t, ra.Config.Filter.Compile())
	acc := testutil.Accumulator{}
	shutdown := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ra.Run(&acc, shutdown)
	}()

	m := ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now().Add(time.Millisecond*100),
	)
	assert.False(t, ra.Add(m))

	for {
		time.Sleep(time.Millisecond)
		if acc.NMetrics() > 0 {
			break
		}
	}
	acc.AssertContainsFields(t, "TestMetric", map[string]interface{}{"sum": int64(101)})

	close(shutdown)
	wg.Wait()
}

func TestAddDropOriginal(t *testing.T) {
	ra := NewRunningAggregator(&TestAggregator{}, &AggregatorConfig{
		Name: "TestRunningAggregator",
		Filter: Filter{
			NamePass: []string{"RI*"},
		},
		DropOriginal: true,
	})
	assert.NoError(t, ra.Config.Filter.Compile())

	m := ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now(),
	)
	assert.True(t, ra.Add(m))

	// this metric name doesn't match the filter, so Add will return false
	m2 := ra.MakeMetric(
		"foobar",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now(),
	)
	assert.False(t, ra.Add(m2))
}

type TestAggregator struct {
	sum int64
}

func (t *TestAggregator) Description() string  { return "" }
func (t *TestAggregator) SampleConfig() string { return "" }
func (t *TestAggregator) Reset() {
	atomic.StoreInt64(&t.sum, 0)
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
			atomic.AddInt64(&t.sum, vi)
		}
	}
}
