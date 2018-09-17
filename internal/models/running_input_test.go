package models

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeMetricNoFields(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	})

	m := ri.MakeMetric(
		"RITest",
		map[string]interface{}{},
		map[string]string{},
		telegraf.Untyped,
		now,
	)
	assert.Nil(t, m)
}

// nil fields should get dropped
func TestMakeMetricNilFields(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	})

	m := ri.MakeMetric(
		"RITest",
		map[string]interface{}{
			"value": int(101),
			"nil":   nil,
		},
		map[string]string{},
		telegraf.Untyped,
		now,
	)

	expected, err := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int(101),
		},
		now,
	)
	require.NoError(t, err)

	require.Equal(t, expected, m)
}

func TestMakeMetricWithPluginTags(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
	})

	ri.SetTrace(true)
	assert.Equal(t, true, ri.Trace())

	m := ri.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		nil,
		telegraf.Untyped,
		now,
	)

	expected, err := metric.New("RITest",
		map[string]string{
			"foo": "bar",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMakeMetricFilteredOut(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
		Filter: Filter{NamePass: []string{"foobar"}},
	})

	ri.SetTrace(true)
	assert.Equal(t, true, ri.Trace())
	assert.NoError(t, ri.Config.Filter.Compile())

	m := ri.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		nil,
		telegraf.Untyped,
		now,
	)
	assert.Nil(t, m)
}

func TestMakeMetricWithDaemonTags(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	})
	ri.SetDefaultTags(map[string]string{
		"foo": "bar",
	})

	ri.SetTrace(true)
	assert.Equal(t, true, ri.Trace())

	m := ri.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		now,
	)
	expected, err := metric.New("RITest",
		map[string]string{
			"foo": "bar",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMakeMetricNameOverride(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:         "TestRunningInput",
		NameOverride: "foobar",
	})

	m := ri.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		now,
	)
	expected, err := metric.New("foobar",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMakeMetricNamePrefix(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:              "TestRunningInput",
		MeasurementPrefix: "foobar_",
	})

	m := ri.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		now,
	)
	expected, err := metric.New("foobar_RITest",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMakeMetricNameSuffix(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:              "TestRunningInput",
		MeasurementSuffix: "_foobar",
	})

	m := ri.MakeMetric(
		"RITest",
		map[string]interface{}{"value": int(101)},
		map[string]string{},
		telegraf.Untyped,
		now,
	)
	expected, err := metric.New("RITest_foobar",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

type testInput struct{}

func (t *testInput) Description() string                   { return "" }
func (t *testInput) SampleConfig() string                  { return "" }
func (t *testInput) Gather(acc telegraf.Accumulator) error { return nil }
