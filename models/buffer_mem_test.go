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

func BenchmarkAddMetrics(b *testing.B) {
	buf := newTestMemoryBuffer(b, 10000)
	m := Metric()
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}
