package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
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
		Metric: metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
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

	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}
