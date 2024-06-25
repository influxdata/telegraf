package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestMemoryBuffer(t testing.TB, capacity int) Buffer {
	t.Helper()
	buf, err := NewBuffer("test", "", capacity, "memory", "")
	require.NoError(t, err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	return buf
}

func TestBuffer_AcceptCallsMetricAccept(t *testing.T) {
	var accept int
	mm := &MockMetric{
		Metric: Metric(),
		AcceptF: func() {
			accept++
		},
	}
	b := newTestMemoryBuffer(t, 5)
	b.Add(mm, mm, mm)
	batch := b.Batch(2)
	b.Accept(batch)
	require.Equal(t, 2, accept)
}

func BenchmarkAddMetrics(b *testing.B) {
	buf := newTestMemoryBuffer(b, 10000)
	m := Metric()
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}
