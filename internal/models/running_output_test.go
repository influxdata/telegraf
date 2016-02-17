package internal_models

import (
	"fmt"
	"sort"
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

// Test that we can write metrics with simple default setup.
func TestRunningOutputDefault(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			IsActive: false,
		},
	}

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf)

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

// Test that the first metric gets overwritten if there is a buffer overflow.
func TestRunningOutputOverwrite(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			IsActive: false,
		},
	}

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf)
	ro.MetricBufferLimit = 4

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	require.Len(t, m.Metrics(), 0)

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 4)

	var expected, actual []string
	for i, exp := range first5[1:] {
		expected = append(expected, exp.String())
		actual = append(actual, m.Metrics()[i].String())
	}

	sort.Strings(expected)
	sort.Strings(actual)

	assert.Equal(t, expected, actual)
}

// Test that multiple buffer overflows are handled properly.
func TestRunningOutputMultiOverwrite(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			IsActive: false,
		},
	}

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf)
	ro.MetricBufferLimit = 3

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	require.Len(t, m.Metrics(), 0)

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 3)

	var expected, actual []string
	for i, exp := range next5[2:] {
		expected = append(expected, exp.String())
		actual = append(actual, m.Metrics()[i].String())
	}

	sort.Strings(expected)
	sort.Strings(actual)

	assert.Equal(t, expected, actual)
}

// Test that running output doesn't flush until it's full when
// FlushBufferWhenFull is set.
func TestRunningOutputFlushWhenFull(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			IsActive: false,
		},
	}

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf)
	ro.FlushBufferWhenFull = true
	ro.MetricBufferLimit = 5

	// Fill buffer to limit
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// no flush yet
	assert.Len(t, m.Metrics(), 0)

	// add one more metric
	ro.AddMetric(next5[0])
	// now it flushed
	assert.Len(t, m.Metrics(), 6)

	// add one more metric and write it manually
	ro.AddMetric(next5[1])
	err := ro.Write()
	assert.NoError(t, err)
	assert.Len(t, m.Metrics(), 7)
}

// Test that running output doesn't flush until it's full when
// FlushBufferWhenFull is set, twice.
func TestRunningOutputMultiFlushWhenFull(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			IsActive: false,
		},
	}

	m := &mockOutput{}
	ro := NewRunningOutput("test", m, conf)
	ro.FlushBufferWhenFull = true
	ro.MetricBufferLimit = 4

	// Fill buffer past limit twive
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// flushed twice
	assert.Len(t, m.Metrics(), 10)
}

func TestRunningOutputWriteFail(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			IsActive: false,
		},
	}

	m := &mockOutput{}
	m.failWrite = true
	ro := NewRunningOutput("test", m, conf)
	ro.FlushBufferWhenFull = true
	ro.MetricBufferLimit = 4

	// Fill buffer past limit twice
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
