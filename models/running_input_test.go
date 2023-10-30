package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

func TestMakeMetricFilterAfterApplyingGlobalTags(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Filter: Filter{
			TagInclude: []string{"b"},
		},
	})
	require.NoError(t, ri.Config.Filter.Compile())
	ri.SetDefaultTags(map[string]string{"a": "x", "b": "y"})

	m := metric.New("cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42,
		},
		now)

	actual := ri.MakeMetric(m)

	expected := metric.New("cpu",
		map[string]string{
			"b": "y",
		},
		map[string]interface{}{
			"value": 42,
		},
		now)

	testutil.RequireMetricEqual(t, expected, actual)
}

func TestMakeMetricNoFields(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	})

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)
	require.Nil(t, actual)
}

// nil fields should get dropped
func TestMakeMetricNilFields(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	})

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
			"nil":   nil,
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)

	expected := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int(101),
		},
		now,
	)

	require.Equal(t, expected, actual)
}

func TestMakeMetricWithPluginTags(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
	})

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)

	expected := metric.New("RITest",
		map[string]string{
			"foo": "bar",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
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

	require.NoError(t, ri.Config.Filter.Compile())

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)
	require.Nil(t, actual)
}

func TestMakeMetricWithDaemonTags(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	})
	ri.SetDefaultTags(map[string]string{
		"foo": "bar",
	})

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)
	expected := metric.New("RITest",
		map[string]string{
			"foo": "bar",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
}

func TestMakeMetricNameOverride(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:         "TestRunningInput",
		NameOverride: "foobar",
	})

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)
	expected := metric.New("foobar",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
}

func TestMakeMetricNamePrefix(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:              "TestRunningInput",
		MeasurementPrefix: "foobar_",
	})

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)
	expected := metric.New("foobar_RITest",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
}

func TestMakeMetricNameSuffix(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:              "TestRunningInput",
		MeasurementSuffix: "_foobar",
	})

	m := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)
	expected := metric.New("RITest_foobar",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
}

func TestMetricErrorCounters(t *testing.T) {
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestMetricErrorCounters",
	})

	getGatherErrors := func() int64 {
		for _, r := range selfstat.Metrics() {
			tag, hasTag := r.GetTag("input")
			if r.Name() == "internal_gather" && hasTag && tag == "TestMetricErrorCounters" {
				errCount, ok := r.GetField("errors")
				if !ok {
					t.Fatal("Expected error field")
				}
				return errCount.(int64)
			}
		}
		return 0
	}

	before := getGatherErrors()

	ri.Log().Error("Oh no")

	after := getGatherErrors()

	require.Greater(t, after, before)
	require.GreaterOrEqual(t, int64(1), GlobalGatherErrors.Get())
}

func TestMakeMetricWithAlwaysKeepingPluginTagsDisabled(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
		Filter: Filter{
			TagInclude: []string{"b"},
		},
	})
	ri.SetDefaultTags(map[string]string{"logic": "rulez"})
	require.NoError(t, ri.Config.Filter.Compile())

	m := testutil.MustMetric("RITest",
		map[string]string{
			"b": "test",
		},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)

	expected := metric.New("RITest",
		map[string]string{
			"b": "test",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
}

func TestMakeMetricWithAlwaysKeepingLocalPluginTagsEnabled(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
		Filter: Filter{
			TagInclude: []string{"b"},
		},
		AlwaysIncludeLocalTags: true,
	})
	ri.SetDefaultTags(map[string]string{"logic": "rulez"})
	require.NoError(t, ri.Config.Filter.Compile())

	m := testutil.MustMetric("RITest",
		map[string]string{
			"b": "test",
		},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)

	expected := metric.New("RITest",
		map[string]string{
			"b":   "test",
			"foo": "bar",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
}

func TestMakeMetricWithAlwaysKeepingGlobalPluginTagsEnabled(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
		Filter: Filter{
			TagInclude: []string{"b"},
		},
		AlwaysIncludeGlobalTags: true,
	})
	ri.SetDefaultTags(map[string]string{"logic": "rulez"})
	require.NoError(t, ri.Config.Filter.Compile())

	m := testutil.MustMetric("RITest",
		map[string]string{
			"b": "test",
		},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)

	expected := metric.New("RITest",
		map[string]string{
			"b":     "test",
			"logic": "rulez",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
}

func TestMakeMetricWithAlwaysKeepingPluginTagsEnabled(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
		Filter: Filter{
			TagInclude: []string{"b"},
		},
		AlwaysIncludeLocalTags:  true,
		AlwaysIncludeGlobalTags: true,
	})
	ri.SetDefaultTags(map[string]string{"logic": "rulez"})
	require.NoError(t, ri.Config.Filter.Compile())

	m := testutil.MustMetric("RITest",
		map[string]string{
			"b": "test",
		},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	actual := ri.MakeMetric(m)

	expected := metric.New("RITest",
		map[string]string{
			"b":     "test",
			"foo":   "bar",
			"logic": "rulez",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.Equal(t, expected, actual)
}

type testInput struct{}

func (t *testInput) Description() string                 { return "" }
func (t *testInput) SampleConfig() string                { return "" }
func (t *testInput) Gather(_ telegraf.Accumulator) error { return nil }
