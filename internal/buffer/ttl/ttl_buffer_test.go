package ttl

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func adjustTimer(d time.Duration) {
	now = func() time.Time {
		return time.Now().Add(d)
	}
}

func resetTimer() {
	now = time.Now
}

func makeBench5(b *testing.B, freq, batchSize int) {
	const k = 1000
	var wg sync.WaitGroup
	buf := NewTTLBuffer(time.Minute)
	m := testutil.TestMetric(1, "mymetric")

	for i := 0; i < b.N; i++ {
		buf.Add(m, m, m, m, m)
		if i%(freq*k) == 0 {
			wg.Add(1)
			go func() {
				buf.Batch(batchSize * k)
				wg.Done()
			}()
		}
	}
	// Flush
	buf.Batch(b.N)
	wg.Wait()

}
func makeBenchStrict(b *testing.B, freq, batchSize int) {
	const k = 1000
	var count uint64
	var wg sync.WaitGroup
	buf := NewTTLBuffer(time.Minute)
	m := testutil.TestMetric(1, "mymetric")

	for i := 0; i < b.N; i++ {
		buf.Add(m)
		if i%(freq*k) == 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				l := len(buf.Batch(batchSize * k))
				atomic.AddUint64(&count, uint64(l))
			}()
		}
	}
	// Flush
	wg.Add(1)
	go func() {
		l := len(buf.Batch(b.N))
		atomic.AddUint64(&count, uint64(l))
		wg.Done()
	}()

	wg.Wait()
	if count != uint64(b.N) {
		b.Errorf("not all metrics came out. %d of %d", count, b.N)
	}
}
func makeBench(b *testing.B, freq, batchSize int) {
	const k = 1000
	var wg sync.WaitGroup
	buf := NewTTLBuffer(time.Minute)
	m := testutil.TestMetric(1, "mymetric")

	for i := 0; i < b.N; i++ {
		buf.Add(m)
		if i%(freq*k) == 0 {
			wg.Add(1)
			go func() {
				buf.Batch(batchSize * k)
				wg.Done()
			}()
		}
	}
	wg.Wait()
	// Flush
	buf.Batch(b.N)
}

func BenchmarkBufferBatch5Add(b *testing.B) {
	makeBench5(b, 100, 101)
}
func BenchmarkBufferBigInfrequentBatchCatchup(b *testing.B) {
	makeBench(b, 100, 101)
}
func BenchmarkBufferOftenBatch(b *testing.B) {
	makeBench(b, 1, 1)
}
func BenchmarkBufferAlmostBatch(b *testing.B) {
	makeBench(b, 10, 9)
}
func BenchmarkBufferSlowBatch(b *testing.B) {
	makeBench(b, 10, 1)
}
func BenchmarkBufferBatchNoDrop(b *testing.B) {
	makeBenchStrict(b, 1, 4)
}
func BenchmarkBufferCatchup(b *testing.B) {
	buf := NewTTLBuffer(time.Minute)
	m := testutil.TestMetric(1, "mymetric")

	for i := 0; i < b.N; i++ {
		buf.Add(m)
	}
	buf.Batch(b.N)
}

func BenchmarkAddMetrics(b *testing.B) {
	buf := NewTTLBuffer(time.Minute)
	m := testutil.TestMetric(1, "mymetric")
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}

func TestNewTTLBufferBasicFuncs(t *testing.T) {
	b := NewTTLBuffer(time.Minute)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	assert.True(t, b.IsEmpty())
	assert.Zero(t, b.Len())
	assert.Zero(t, MetricsDropped.Get())
	assert.Zero(t, MetricsWritten.Get())

	m := testutil.TestMetric(1, "mymetric")
	b.Add(m)
	assert.False(t, b.IsEmpty())
	assert.Equal(t, 1, b.Len())
	assert.Equal(t, int64(0), MetricsDropped.Get())
	assert.Equal(t, int64(1), MetricsWritten.Get())

	b.Add(metricList...)
	assert.False(t, b.IsEmpty())
	assert.Equal(t, 6, b.Len())
	assert.Equal(t, int64(0), MetricsDropped.Get())
	assert.Equal(t, int64(6), MetricsWritten.Get())
}

func TestTTLDroppingMetrics(t *testing.T) {
	adjustTimer(-time.Minute * 2)

	b := NewTTLBuffer(time.Minute)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	b.Add(metricList...)
	assert.False(t, b.IsEmpty())
	assert.Equal(t, b.Len(), 5)
	assert.Zero(t, MetricsDropped.Get())
	assert.Equal(t, int64(5), MetricsWritten.Get())

	resetTimer()

	b.Add(metricList[0])
	assert.False(t, b.IsEmpty())
	assert.Equal(t, b.Len(), 1)
	assert.Equal(t, int64(5), MetricsDropped.Get())
	assert.Equal(t, int64(6), MetricsWritten.Get())
}

func TestTTLGettingBatches(t *testing.T) {
	b := NewTTLBuffer(time.Minute)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	b.Add(metricList...)
	batch := b.Batch(10)
	assert.Len(t, batch, 5)

	assert.True(t, b.IsEmpty())
	assert.Zero(t, b.Len())
	assert.Zero(t, MetricsDropped.Get())
	assert.Equal(t, int64(5), MetricsWritten.Get())

	b.Add(metricList...)
	batch = b.Batch(3)
	assert.Len(t, batch, 3)

	assert.False(t, b.IsEmpty())
	assert.Equal(t, b.Len(), 2)
	assert.Zero(t, MetricsDropped.Get())
	assert.Equal(t, int64(10), MetricsWritten.Get())
}
