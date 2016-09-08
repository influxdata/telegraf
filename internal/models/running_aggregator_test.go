package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	a := &TestAggregator{}
	ra := RunningAggregator{
		Config: &AggregatorConfig{
			Name: "TestRunningAggregator",
			Filter: Filter{
				NamePass: []string{"*"},
			},
		},
		Aggregator: a,
	}
	assert.NoError(t, ra.Config.Filter.Compile())

	m := ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now(),
	)
	assert.False(t, ra.Apply(m))
	assert.Equal(t, int64(101), a.sum)
}

func TestApplyDropOriginal(t *testing.T) {
	ra := RunningAggregator{
		Config: &AggregatorConfig{
			Name: "TestRunningAggregator",
			Filter: Filter{
				NamePass: []string{"RI*"},
			},
			DropOriginal: true,
		},
		Aggregator: &TestAggregator{},
	}
	assert.NoError(t, ra.Config.Filter.Compile())

	m := ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now(),
	)
	assert.True(t, ra.Apply(m))

	// this metric name doesn't match the filter, so Apply will return false
	m2 := ra.MakeMetric(
		"foobar",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		time.Now(),
	)
	assert.False(t, ra.Apply(m2))
}

// make an untyped, counter, & gauge metric
func TestMakeMetricA(t *testing.T) {
	now := time.Now()
	ra := RunningAggregator{
		Config: &AggregatorConfig{
			Name: "TestRunningAggregator",
		},
	}
	assert.Equal(t, "aggregators.TestRunningAggregator", ra.Name())

	m := ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		now,
	)
	assert.Equal(
		t,
		m.String(),
		fmt.Sprintf("RITest value=101i %d", now.UnixNano()),
	)
	assert.Equal(
		t,
		m.Type(),
		telegraf.Untyped,
	)

	m = ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Counter,
		now,
	)
	assert.Equal(
		t,
		m.String(),
		fmt.Sprintf("RITest value=101i %d", now.UnixNano()),
	)
	assert.Equal(
		t,
		m.Type(),
		telegraf.Counter,
	)

	m = ra.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Gauge,
		now,
	)
	assert.Equal(
		t,
		m.String(),
		fmt.Sprintf("RITest value=101i %d", now.UnixNano()),
	)
	assert.Equal(
		t,
		m.Type(),
		telegraf.Gauge,
	)
}

type TestAggregator struct {
	sum int64
}

func (t *TestAggregator) Description() string                  { return "" }
func (t *TestAggregator) SampleConfig() string                 { return "" }
func (t *TestAggregator) Start(acc telegraf.Accumulator) error { return nil }
func (t *TestAggregator) Stop()                                {}

func (t *TestAggregator) Apply(in telegraf.Metric) {
	for _, v := range in.Fields() {
		if vi, ok := v.(int64); ok {
			t.sum += vi
		}
	}
}
