package buffer

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var metricList = []telegraf.Metric{
	testutil.TestMetric(2, "mymetric1"),
	testutil.TestMetric(1, "mymetric2"),
	testutil.TestMetric(11, "mymetric3"),
	testutil.TestMetric(15, "mymetric4"),
	testutil.TestMetric(8, "mymetric5"),
}

func makeBench5(b *testing.B, freq, batchSize int) {
	const k = 1000
	var wg sync.WaitGroup
	buf := NewBuffer(10000)
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
	buf := NewBuffer(10000)
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
	buf := NewBuffer(10000)
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
	buf := NewBuffer(10000)
	m := testutil.TestMetric(1, "mymetric")

	for i := 0; i < b.N; i++ {
		buf.Add(m)
	}
	buf.Batch(b.N)
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
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	require.Zero(t, b.Len())
	require.Zero(t, MetricsDropped.Get())
	require.Zero(t, MetricsWritten.Get())

	m := testutil.TestMetric(1, "mymetric")
	b.Add(m)
	require.Equal(t, b.Len(), 1)
	require.Equal(t, int64(0), MetricsDropped.Get())
	require.Equal(t, int64(1), MetricsWritten.Get())

	b.Add(metricList...)
	require.Equal(t, b.Len(), 6)
	require.Equal(t, int64(0), MetricsDropped.Get())
	require.Equal(t, int64(6), MetricsWritten.Get())
}

func TestDroppingMetrics(t *testing.T) {
	b := NewBuffer(10)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	// Add up to the size of the buffer
	b.Add(metricList...)
	b.Add(metricList...)
	require.Equal(t, b.Len(), 10)
	require.Equal(t, int64(0), MetricsDropped.Get())
	require.Equal(t, int64(10), MetricsWritten.Get())

	// Add 5 more and verify they were dropped
	b.Add(metricList...)
	require.Equal(t, b.Len(), 10)
	require.Equal(t, int64(5), MetricsDropped.Get())
	require.Equal(t, int64(15), MetricsWritten.Get())
}

func TestGettingBatches(t *testing.T) {
	b := NewBuffer(20)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	// Verify that the buffer returned is smaller than requested when there are
	// not as many items as requested.
	b.Add(metricList...)
	batch := b.Batch(10)
	require.Len(t, batch, 5)

	// Verify that the buffer is now empty
	require.Zero(t, b.Len())
	require.Zero(t, MetricsDropped.Get())
	require.Equal(t, int64(5), MetricsWritten.Get())

	// Verify that the buffer returned is not more than the size requested
	b.Add(metricList...)
	batch = b.Batch(3)
	require.Len(t, batch, 3)

	// Verify that buffer is not empty
	require.Equal(t, b.Len(), 2)
	require.Equal(t, int64(0), MetricsDropped.Get())
	require.Equal(t, int64(10), MetricsWritten.Get())
}
