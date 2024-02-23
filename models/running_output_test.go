package models

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

var first5 = []telegraf.Metric{
	testutil.TestMetric(101, "metric1"),
	testutil.TestMetric(101, "metric2"),
	testutil.TestMetric(101, "metric3"),
	testutil.TestMetric(101, "metric4"),
	testutil.TestMetric(101, "metric5"),
}

var next5 = []telegraf.Metric{
	testutil.TestMetric(101, "metric6"),
	testutil.TestMetric(101, "metric7"),
	testutil.TestMetric(101, "metric8"),
	testutil.TestMetric(101, "metric9"),
	testutil.TestMetric(101, "metric10"),
}

// Benchmark adding metrics.
func BenchmarkRunningOutputAddWrite(b *testing.B) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &perfOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for n := 0; n < b.N; n++ {
		ro.AddMetric(testutil.TestMetric(101, "metric1"))
		ro.Write() //nolint: errcheck // skip checking err for benchmark tests
	}
}

// Benchmark adding metrics.
func BenchmarkRunningOutputAddWriteEvery100(b *testing.B) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &perfOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for n := 0; n < b.N; n++ {
		ro.AddMetric(testutil.TestMetric(101, "metric1"))
		if n%100 == 0 {
			ro.Write() //nolint: errcheck // skip checking err for benchmark tests
		}
	}
}

// Benchmark adding metrics.
func BenchmarkRunningOutputAddFailWrites(b *testing.B) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &perfOutput{}
	m.failWrite = true
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for n := 0; n < b.N; n++ {
		ro.AddMetric(testutil.TestMetric(101, "metric1"))
	}
}

// Test that NameDrop filters ger properly applied.
func TestRunningOutput_DropFilter(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			NameDrop: []string{"metric1", "metric2"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 8)
}

// Test that NameDrop filters without a match do nothing.
func TestRunningOutput_PassFilter(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			NameDrop: []string{"metric1000", "foo*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 10)
}

// Test that tags are properly included
func TestRunningOutput_TagIncludeNoMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagInclude: []string{"nothing*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Empty(t, m.Metrics()[0].Tags())
}

// Test that tags are properly excluded
func TestRunningOutput_TagExcludeMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagExclude: []string{"tag*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Empty(t, m.Metrics()[0].Tags())
}

// Test that tags are properly Excluded
func TestRunningOutput_TagExcludeNoMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagExclude: []string{"nothing*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Len(t, m.Metrics()[0].Tags(), 1)
}

// Test that tags are properly included
func TestRunningOutput_TagIncludeMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagInclude: []string{"tag*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Len(t, m.Metrics()[0].Tags(), 1)
}

// Test that measurement name overriding correctly
func TestRunningOutput_NameOverride(t *testing.T) {
	conf := &OutputConfig{
		NameOverride: "new_metric_name",
	}

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Equal(t, "new_metric_name", m.Metrics()[0].Name())
}

// Test that measurement name prefix is added correctly
func TestRunningOutput_NamePrefix(t *testing.T) {
	conf := &OutputConfig{
		NamePrefix: "prefix_",
	}

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Equal(t, "prefix_metric1", m.Metrics()[0].Name())
}

// Test that measurement name suffix is added correctly
func TestRunningOutput_NameSuffix(t *testing.T) {
	conf := &OutputConfig{
		NameSuffix: "_suffix",
	}

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Equal(t, "metric1_suffix", m.Metrics()[0].Name())
}

// Test that we can write metrics with simple default setup.
func TestRunningOutputDefault(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 10)
}

func TestRunningOutputWriteFail(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput(m, conf, 4, 12)

	// Fill buffer to limit twice
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// manual write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	m.failWrite = false
	err = ro.Write()
	require.NoError(t, err)

	require.Len(t, m.Metrics(), 10)
}

// Verify that the order of points is preserved during write failure.
func TestRunningOutputWriteFailOrder(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput(m, conf, 100, 1000)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	m.failWrite = false
	// add 5 more metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	err = ro.Write()
	require.NoError(t, err)

	// Verify that 10 metrics were written
	require.Len(t, m.Metrics(), 10)
	// Verify that they are in order
	expected := append(first5, next5...)
	require.Equal(t, expected, m.Metrics())
}

// Verify that the order of points is preserved during many write failures.
func TestRunningOutputWriteFailOrder2(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput(m, conf, 5, 100)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// add 5 metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// add 5 metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	m.failWrite = false
	err = ro.Write()
	require.NoError(t, err)

	// Verify that 20 metrics were written
	require.Len(t, m.Metrics(), 20)
	// Verify that they are in order
	expected := append(first5, next5...)
	expected = append(expected, first5...)
	expected = append(expected, next5...)
	require.Equal(t, expected, m.Metrics())
}

// Verify that the order of points is preserved when there is a remainder
// of points for the batch.
func TestRunningOutputWriteFailOrder3(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput(m, conf, 5, 1000)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// add and attempt to write a single metric:
	ro.AddMetric(next5[0])
	err = ro.Write()
	require.Error(t, err)

	// unset fail and write metrics
	m.failWrite = false
	err = ro.Write()
	require.NoError(t, err)

	// Verify that 6 metrics were written
	require.Len(t, m.Metrics(), 6)
	// Verify that they are in order
	expected := []telegraf.Metric{first5[0], first5[1], first5[2], first5[3], first5[4], next5[0]}
	require.Equal(t, expected, m.Metrics())
}

func TestInternalMetrics(t *testing.T) {
	_ = NewRunningOutput(
		&mockOutput{},
		&OutputConfig{
			Filter: Filter{},
			Name:   "test_name",
			Alias:  "test_alias",
		},
		5,
		10)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"internal_write",
			map[string]string{
				"output": "test_name",
				"alias":  "test_alias",
			},
			map[string]interface{}{
				"buffer_limit":     10,
				"buffer_size":      0,
				"errors":           0,
				"metrics_added":    0,
				"metrics_dropped":  0,
				"metrics_filtered": 0,
				"metrics_written":  0,
				"write_time_ns":    0,
			},
			time.Unix(0, 0),
		),
	}

	var actual []telegraf.Metric
	for _, m := range selfstat.Metrics() {
		output, _ := m.GetTag("output")
		if m.Name() == "internal_write" && output == "test_name" {
			actual = append(actual, m)
		}
	}

	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
}

type mockOutput struct {
	sync.Mutex

	metrics []telegraf.Metric

	// if true, mock write failure
	failWrite bool
}

func (m *mockOutput) Connect() error {
	return nil
}

func (m *mockOutput) Close() error {
	return nil
}

func (m *mockOutput) Description() string {
	return ""
}

func (m *mockOutput) SampleConfig() string {
	return ""
}

func (m *mockOutput) Write(metrics []telegraf.Metric) error {
	m.Lock()
	defer m.Unlock()
	if m.failWrite {
		return errors.New("failed write")
	}

	if m.metrics == nil {
		m.metrics = []telegraf.Metric{}
	}

	m.metrics = append(m.metrics, metrics...)
	return nil
}

func (m *mockOutput) Metrics() []telegraf.Metric {
	m.Lock()
	defer m.Unlock()
	return m.metrics
}

type perfOutput struct {
	// if true, mock write failure
	failWrite bool
}

func (m *perfOutput) Connect() error {
	return nil
}

func (m *perfOutput) Close() error {
	return nil
}

func (m *perfOutput) Description() string {
	return ""
}

func (m *perfOutput) SampleConfig() string {
	return ""
}

func (m *perfOutput) Write(_ []telegraf.Metric) error {
	if m.failWrite {
		return errors.New("failed write")
	}
	return nil
}
