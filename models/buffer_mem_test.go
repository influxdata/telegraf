package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func TestMemoryBufferAcceptCallsMetricAccept(t *testing.T) {
	buf, err := NewBuffer("test", "123", "", 5, "memory", "", true)
	require.NoError(t, err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	defer buf.Close()

	var accept int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]any{"value": 42.0}, time.Unix(0, 0)),
		AcceptF: func() {
			accept++
		},
	}
	buf.Add(mm, mm, mm)
	tx := buf.BeginTransaction(2)
	tx.AcceptAll()
	buf.EndTransaction(tx)
	require.Equal(t, 2, accept)
}

func TestCheckBufferSettings(t *testing.T) {
	for _, strategy := range []string{"", "memory", "disk_write_through"} {
		require.NoError(t, CheckBufferSettings(strategy))
	}
	require.ErrorContains(t, CheckBufferSettings("discard"), `invalid buffer strategy "discard"`)
	require.ErrorContains(t, CheckBufferSettings("unknown"), `invalid buffer strategy "unknown"`)
}

func TestDiscardBufferDropsMetrics(t *testing.T) {
	buf, err := NewBuffer("test", "123", "", 5, "discard", "", true)
	require.NoError(t, err)
	buf.Stats().MetricsDropped.Set(0)
	defer buf.Close()

	var rejected int
	mm := &mockMetric{
		Metric: metric.New("cpu", map[string]string{}, map[string]any{"value": 42.0}, time.Unix(0, 0)),
		RejectF: func() {
			rejected++
		},
	}
	require.Equal(t, 1, buf.Add(mm))
	require.Equal(t, 0, buf.Len())
	require.Equal(t, 1, rejected)
	require.Equal(t, int64(1), buf.Stats().MetricsDropped.Get())
	require.Empty(t, buf.BeginTransaction(1).Batch)
}

func BenchmarkMemoryBufferAddMetrics(b *testing.B) {
	buf, err := NewBuffer("test", "123", "", 10000, "memory", "", true)
	require.NoError(b, err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]any{"value": 42.0}, time.Unix(0, 0))
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}

func TestCheckDropStrategy(t *testing.T) {
	for _, strategy := range []string{"", DropOldest, DropNewest} {
		require.NoError(t, CheckDropStrategy(strategy))
	}
	require.ErrorContains(t, CheckDropStrategy("unknown"), `invalid buffer drop strategy "unknown"`)

	_, err := NewBuffer("test", "123", "", 5, "memory", "", true, WithDropStrategy("unknown"))
	require.ErrorContains(t, err, `invalid buffer drop strategy "unknown"`)
}

func TestMemoryBufferDropStrategy(t *testing.T) {
	newMetric := func(v int) *mockMetric {
		return &mockMetric{
			Metric: metric.New("cpu", map[string]string{}, map[string]any{"value": v}, time.Unix(int64(v), 0)),
		}
	}
	values := func(batch []telegraf.Metric) []int {
		out := make([]int, 0, len(batch))
		for _, m := range batch {
			v, _ := m.GetField("value")
			out = append(out, int(v.(int64)))
		}
		return out
	}

	tests := []struct {
		name     string
		opts     []BufferOption
		expected []int
	}{
		{name: "default drops oldest", expected: []int{3, 4, 5}},
		{name: "drop_oldest drops oldest", opts: []BufferOption{WithDropStrategy(DropOldest)}, expected: []int{3, 4, 5}},
		{name: "drop_newest drops newest", opts: []BufferOption{WithDropStrategy(DropNewest)}, expected: []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := NewBuffer("test", "123", "", 3, "memory", "", true, tt.opts...)
			require.NoError(t, err)
			buf.Stats().MetricsAdded.Set(0)
			buf.Stats().MetricsDropped.Set(0)
			defer buf.Close()

			var rejected int
			for i := 1; i <= 5; i++ {
				m := newMetric(i)
				m.RejectF = func() { rejected++ }
				buf.Add(m)
			}
			require.Equal(t, 3, buf.Len())
			require.Equal(t, 2, rejected)
			require.Equal(t, int64(2), buf.Stats().MetricsDropped.Get())

			tx := buf.BeginTransaction(10)
			require.Equal(t, tt.expected, values(tx.Batch))
			tx.AcceptAll()
			buf.EndTransaction(tx)
		})
	}
}

func TestMemoryBufferDropNewestReturnsDroppedCount(t *testing.T) {
	buf, err := NewBuffer("test", "123", "", 2, "memory", "", true, WithDropStrategy(DropNewest))
	require.NoError(t, err)
	defer buf.Close()

	m := metric.New("cpu", map[string]string{}, map[string]any{"value": 1}, time.Unix(0, 0))
	require.Equal(t, 0, buf.Add(m, m))
	require.Equal(t, 3, buf.Add(m, m, m))
	require.Equal(t, 2, buf.Len())
}

func TestMemoryBufferDropNewestWithOpenTransaction(t *testing.T) {
	newMetric := func(v int) telegraf.Metric {
		return metric.New("cpu", map[string]string{}, map[string]any{"value": v}, time.Unix(int64(v), 0))
	}
	values := func(batch []telegraf.Metric) []int {
		out := make([]int, 0, len(batch))
		for _, m := range batch {
			v, _ := m.GetField("value")
			out = append(out, int(v.(int64)))
		}
		return out
	}

	buf, err := NewBuffer("test", "123", "", 3, "memory", "", true, WithDropStrategy(DropNewest))
	require.NoError(t, err)
	defer buf.Close()

	buf.Add(newMetric(1), newMetric(2), newMetric(3))
	tx := buf.BeginTransaction(3)

	// The batch is still in flight and counts against the capacity, so new
	// metrics must be dropped instead of displacing the batch on a failed write.
	require.Equal(t, 2, buf.Add(newMetric(4), newMetric(5)))
	tx.KeepAll()
	buf.EndTransaction(tx)

	tx = buf.BeginTransaction(10)
	require.Equal(t, []int{1, 2, 3}, values(tx.Batch))
	tx.AcceptAll()
	buf.EndTransaction(tx)
}
