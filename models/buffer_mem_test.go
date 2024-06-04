package models

import (
	"testing"

	models "github.com/influxdata/telegraf/models/mock"
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
	mm := &models.MockMetric{
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

func TestBuffer_RejectCallsMetricRejectWithOverwritten(t *testing.T) {
	var reject int
	mm := &models.MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := newTestMemoryBuffer(t, 5)
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(5)
	b.Add(mm, mm)
	require.Equal(t, 0, reject)
	b.Reject(batch)
	require.Equal(t, 2, reject)
}

func BenchmarkMemoryAddMetrics(b *testing.B) {
	buf := newTestMemoryBuffer(b, 10000)
	m := Metric()
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}
