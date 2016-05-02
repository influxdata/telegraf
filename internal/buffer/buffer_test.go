package buffer

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

var metricList = []telegraf.Metric{
	testutil.TestMetric(2, "mymetric1"),
	testutil.TestMetric(1, "mymetric2"),
	testutil.TestMetric(11, "mymetric3"),
	testutil.TestMetric(15, "mymetric4"),
	testutil.TestMetric(8, "mymetric5"),
}

func BenchmarkAddMetrics(b *testing.B) {
	buf := NewBuffer(10000)
	m := testutil.TestMetric(1, "mymetric")
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}

func TestNewBufferBasicFuncs(t *testing.T) {
	b := NewBuffer(10)

	assert.True(t, b.IsEmpty())
	assert.Zero(t, b.Len())
	assert.Zero(t, b.Drops())
	assert.Zero(t, b.Total())

	m := testutil.TestMetric(1, "mymetric")
	b.Add(m)
	assert.False(t, b.IsEmpty())
	assert.Equal(t, b.Len(), 1)
	assert.Equal(t, b.Drops(), 0)
	assert.Equal(t, b.Total(), 1)

	b.Add(metricList...)
	assert.False(t, b.IsEmpty())
	assert.Equal(t, b.Len(), 6)
	assert.Equal(t, b.Drops(), 0)
	assert.Equal(t, b.Total(), 6)
}

func TestDroppingMetrics(t *testing.T) {
	b := NewBuffer(10)

	// Add up to the size of the buffer
	b.Add(metricList...)
	b.Add(metricList...)
	assert.False(t, b.IsEmpty())
	assert.Equal(t, b.Len(), 10)
	assert.Equal(t, b.Drops(), 0)
	assert.Equal(t, b.Total(), 10)

	// Add 5 more and verify they were dropped
	b.Add(metricList...)
	assert.False(t, b.IsEmpty())
	assert.Equal(t, b.Len(), 10)
	assert.Equal(t, b.Drops(), 5)
	assert.Equal(t, b.Total(), 15)
}

func TestGettingBatches(t *testing.T) {
	b := NewBuffer(20)

	// Verify that the buffer returned is smaller than requested when there are
	// not as many items as requested.
	b.Add(metricList...)
	batch := b.Batch(10)
	assert.Len(t, batch, 5)

	// Verify that the buffer is now empty
	assert.True(t, b.IsEmpty())
	assert.Zero(t, b.Len())
	assert.Zero(t, b.Drops())
	assert.Equal(t, b.Total(), 5)

	// Verify that the buffer returned is not more than the size requested
	b.Add(metricList...)
	batch = b.Batch(3)
	assert.Len(t, batch, 3)

	// Verify that buffer is not empty
	assert.False(t, b.IsEmpty())
	assert.Equal(t, b.Len(), 2)
	assert.Equal(t, b.Drops(), 0)
	assert.Equal(t, b.Total(), 10)
}
