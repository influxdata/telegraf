package models

import (
	"fmt"
	"sync"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	for n := 0; n < b.N; n++ {
		ro.AddMetric(testutil.TestMetric(101, "metric1"))
		ro.Write()
	}
}

// Benchmark adding metrics.
func BenchmarkRunningOutputAddWriteEvery100(b *testing.B) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &perfOutput{}
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	for n := 0; n < b.N; n++ {
		ro.AddMetric(testutil.TestMetric(101, "metric1"))
		if n%100 == 0 {
			ro.Write()
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
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

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
	assert.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	assert.Len(t, m.Metrics(), 0)

	err := ro.Write()
	assert.NoError(t, err)
	assert.Len(t, m.Metrics(), 8)
}

// Test that NameDrop filters without a match do nothing.
func TestRunningOutput_PassFilter(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			NameDrop: []string{"metric1000", "foo*"},
		},
	}
	assert.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	assert.Len(t, m.Metrics(), 0)

	err := ro.Write()
	assert.NoError(t, err)
	assert.Len(t, m.Metrics(), 10)
}

// Test that tags are properly included
func TestRunningOutput_TagIncludeNoMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagInclude: []string{"nothing*"},
		},
	}
	assert.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	assert.Len(t, m.Metrics(), 0)

	err := ro.Write()
	assert.NoError(t, err)
	assert.Len(t, m.Metrics(), 1)
	assert.Empty(t, m.Metrics()[0].Tags())
}

// Test that tags are properly excluded
func TestRunningOutput_TagExcludeMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagExclude: []string{"tag*"},
		},
	}
	assert.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	assert.Len(t, m.Metrics(), 0)

	err := ro.Write()
	assert.NoError(t, err)
	assert.Len(t, m.Metrics(), 1)
	assert.Len(t, m.Metrics()[0].Tags(), 0)
}

// Test that tags are properly Excluded
func TestRunningOutput_TagExcludeNoMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagExclude: []string{"nothing*"},
		},
	}
	assert.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	assert.Len(t, m.Metrics(), 0)

	err := ro.Write()
	assert.NoError(t, err)
	assert.Len(t, m.Metrics(), 1)
	assert.Len(t, m.Metrics()[0].Tags(), 1)
}

// Test that tags are properly included
func TestRunningOutput_TagIncludeMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagInclude: []string{"tag*"},
		},
	}
	assert.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	assert.Len(t, m.Metrics(), 0)

	err := ro.Write()
	assert.NoError(t, err)
	assert.Len(t, m.Metrics(), 1)
	assert.Len(t, m.Metrics()[0].Tags(), 1)
}

// Test that we can write metrics with simple default setup.
func TestRunningOutputDefault(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	assert.Len(t, m.Metrics(), 0)

	err := ro.Write()
	assert.NoError(t, err)
	assert.Len(t, m.Metrics(), 10)
}

func TestRunningOutputWriteFail(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput("test", m, conf, 4, 12)

	// Fill buffer to limit twice
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	// manual write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	m.failWrite = false
	err = ro.Write()
	require.NoError(t, err)

	assert.Len(t, m.Metrics(), 10)
}

// Verify that the order of points is preserved during a write failure.
func TestRunningOutputWriteFailOrder(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput("test", m, conf, 100, 1000)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	m.failWrite = false
	// add 5 more metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	err = ro.Write()
	require.NoError(t, err)

	// Verify that 10 metrics were written
	assert.Len(t, m.Metrics(), 10)
	// Verify that they are in order
	expected := append(first5, next5...)
	assert.Equal(t, expected, m.Metrics())
}

// Verify that the order of points is preserved during many write failures.
func TestRunningOutputWriteFailOrder2(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput("test", m, conf, 5, 100)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	// add 5 metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	// add 5 metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	m.failWrite = false
	err = ro.Write()
	require.NoError(t, err)

	// Verify that 10 metrics were written
	assert.Len(t, m.Metrics(), 20)
	// Verify that they are in order
	expected := append(first5, next5...)
	expected = append(expected, first5...)
	expected = append(expected, next5...)
	assert.Equal(t, expected, m.Metrics())
}

// Verify that the order of points is preserved when there is a remainder
// of points for the batch.
//
// ie, with a batch size of 5:
//
//     1 2 3 4 5 6 <-- order, failed points
//     6 1 2 3 4 5 <-- order, after 1st write failure (1 2 3 4 5 was batch)
//     1 2 3 4 5 6 <-- order, after 2nd write failure, (6 was batch)
//
func TestRunningOutputWriteFailOrder3(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput("test", m, conf, 5, 1000)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	assert.Len(t, m.Metrics(), 0)

	// add and attempt to write a single metric:
	ro.AddMetric(next5[0])
	err = ro.Write()
	require.Error(t, err)

	// unset fail and write metrics
	m.failWrite = false
	err = ro.Write()
	require.NoError(t, err)

	// Verify that 6 metrics were written
	assert.Len(t, m.Metrics(), 6)
	// Verify that they are in order
	expected := append(first5, next5[0])
	assert.Equal(t, expected, m.Metrics())
}

type mockOutput struct {
	sync.Mutex

	metrics []telegraf.Metric

	// if true, mock a write failure
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
		return fmt.Errorf("Failed Write!")
	}

	if m.metrics == nil {
		m.metrics = []telegraf.Metric{}
	}

	for _, metric := range metrics {
		m.metrics = append(m.metrics, metric)
	}
	return nil
}

func (m *mockOutput) Metrics() []telegraf.Metric {
	m.Lock()
	defer m.Unlock()
	return m.metrics
}

type perfOutput struct {
	// if true, mock a write failure
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

func (m *perfOutput) Write(metrics []telegraf.Metric) error {
	if m.failWrite {
		return fmt.Errorf("Failed Write!")
	}
	return nil
}
