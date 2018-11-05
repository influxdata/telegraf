package models

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

type MockMetric struct {
	telegraf.Metric
	AcceptF func()
	RejectF func()
	DropF   func()
}

func (m *MockMetric) Accept() {
	m.AcceptF()
}

func (m *MockMetric) Reject() {
	m.RejectF()
}

func (m *MockMetric) Drop() {
	m.DropF()
}

func Metric() telegraf.Metric {
	m, err := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	if err != nil {
		panic(err)
	}
	return m
}

func BenchmarkAddMetrics(b *testing.B) {
	buf := NewBuffer("test", 10000)
	m := Metric()
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}

func setup(b *Buffer) *Buffer {
	b.MetricsAdded.Set(0)
	b.MetricsWritten.Set(0)
	b.MetricsDropped.Set(0)
	return b
}

func TestBuffer_LenEmpty(t *testing.T) {
	b := setup(NewBuffer("test", 5))

	require.Equal(t, 0, b.Len())
}

func TestBuffer_LenOne(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m)

	require.Equal(t, 1, b.Len())
}

func TestBuffer_LenFull(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m, m, m, m)

	require.Equal(t, 5, b.Len())
}

func TestBuffer_LenOverfill(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	setup(b)
	b.Add(m, m, m, m, m, m)

	require.Equal(t, 5, b.Len())
}

func TestBuffer_BatchLenZero(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	batch := b.Batch(0)

	require.Len(t, batch, 0)
}

func TestBuffer_BatchLenBufferEmpty(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	batch := b.Batch(2)

	require.Len(t, batch, 0)
}

func TestBuffer_BatchLenUnderfill(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m)
	batch := b.Batch(2)

	require.Len(t, batch, 1)
}

func TestBuffer_BatchLenFill(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m, m)
	batch := b.Batch(2)
	require.Len(t, batch, 2)
}

func TestBuffer_BatchLenExact(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m)
	batch := b.Batch(2)
	require.Len(t, batch, 2)
}

func TestBuffer_BatchLenLargerThanBuffer(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m, m, m, m)
	batch := b.Batch(6)
	require.Len(t, batch, 5)
}

func TestBuffer_BatchWrap(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Accept(batch)
	b.Add(m, m)
	batch = b.Batch(5)
	require.Len(t, batch, 5)
}

func TestBuffer_AddDropsOverwrittenMetrics(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))

	b.Add(m, m, m, m, m)
	b.Add(m, m, m, m, m)

	require.Equal(t, int64(5), b.MetricsDropped.Get())
	require.Equal(t, int64(0), b.MetricsWritten.Get())
}

func TestBuffer_AcceptRemovesBatch(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Accept(batch)
	require.Equal(t, 1, b.Len())
}

func TestBuffer_RejectLeavesBatch(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	require.Equal(t, 3, b.Len())
}

func TestBuffer_AcceptWritesOverwrittenBatch(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))

	b.Add(m, m, m, m, m)
	batch := b.Batch(5)
	b.Add(m, m, m, m, m)
	b.Accept(batch)

	require.Equal(t, int64(0), b.MetricsDropped.Get())
	require.Equal(t, int64(5), b.MetricsWritten.Get())
}

func TestBuffer_BatchRejectDropsOverwrittenBatch(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))

	b.Add(m, m, m, m, m)
	batch := b.Batch(5)
	b.Add(m, m, m, m, m)
	b.Reject(batch)

	require.Equal(t, int64(5), b.MetricsDropped.Get())
	require.Equal(t, int64(0), b.MetricsWritten.Get())
}

func TestBuffer_MetricsOverwriteBatchAccept(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Accept(batch)
	require.Equal(t, int64(0), b.MetricsDropped.Get())
	require.Equal(t, int64(3), b.MetricsWritten.Get())
}

func TestBuffer_MetricsOverwriteBatchReject(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Reject(batch)
	require.Equal(t, int64(3), b.MetricsDropped.Get())
	require.Equal(t, int64(0), b.MetricsWritten.Get())
}

func TestBuffer_MetricsBatchAcceptRemoved(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m, m, m)
	b.Accept(batch)
	require.Equal(t, int64(2), b.MetricsDropped.Get())
	require.Equal(t, int64(3), b.MetricsWritten.Get())
}

func TestBuffer_WrapWithBatch(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))

	b.Add(m, m, m)
	b.Batch(3)
	b.Add(m, m, m, m, m, m)

	require.Equal(t, int64(1), b.MetricsDropped.Get())
}

func TestBuffer_BatchNotRemoved(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m, m, m, m)
	b.Batch(2)
	require.Equal(t, 5, b.Len())
}

func TestBuffer_BatchRejectAcceptNoop(t *testing.T) {
	m := Metric()
	b := setup(NewBuffer("test", 5))
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	b.Accept(batch)
	require.Equal(t, 5, b.Len())
}

func TestBuffer_AcceptCallsMetricAccept(t *testing.T) {
	var accept int
	mm := &MockMetric{
		Metric: Metric(),
		AcceptF: func() {
			accept++
		},
	}
	b := setup(NewBuffer("test", 5))
	b.Add(mm, mm, mm)
	batch := b.Batch(2)
	b.Accept(batch)
	require.Equal(t, 2, accept)
}

func TestBuffer_AddCallsMetricRejectWhenNoBatch(t *testing.T) {
	var reject int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := setup(NewBuffer("test", 5))
	setup(b)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm)
	require.Equal(t, 2, reject)
}

func TestBuffer_AddCallsMetricRejectWhenNotInBatch(t *testing.T) {
	var reject int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := setup(NewBuffer("test", 5))
	setup(b)
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(2)
	b.Add(mm, mm, mm, mm)
	// metric[2] and metric[3] rejected
	require.Equal(t, 2, reject)
	b.Reject(batch)
	// metric[1] and metric[2] now rejected
	require.Equal(t, 4, reject)
}

func TestBuffer_RejectCallsMetricRejectWithOverwritten(t *testing.T) {
	var reject int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := setup(NewBuffer("test", 5))
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(5)
	b.Add(mm, mm)
	require.Equal(t, 0, reject)
	b.Reject(batch)
	require.Equal(t, 2, reject)
}

func TestBuffer_AddOverwriteAndReject(t *testing.T) {
	var reject int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
	}
	b := setup(NewBuffer("test", 5))
	b.Add(mm, mm, mm, mm, mm)
	batch := b.Batch(5)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	b.Add(mm, mm, mm, mm, mm)
	require.Equal(t, 15, reject)
	b.Reject(batch)
	require.Equal(t, 20, reject)
}

func TestBuffer_AddOverwriteAndRejectOffset(t *testing.T) {
	var reject int
	var accept int
	mm := &MockMetric{
		Metric: Metric(),
		RejectF: func() {
			reject++
		},
		AcceptF: func() {
			accept++
		},
	}
	b := setup(NewBuffer("test", 5))
	b.Add(mm, mm, mm)
	b.Add(mm, mm, mm, mm)
	require.Equal(t, 2, reject)
	batch := b.Batch(5)
	b.Add(mm, mm, mm, mm)
	require.Equal(t, 2, reject)
	b.Add(mm, mm, mm, mm)
	require.Equal(t, 5, reject)
	b.Add(mm, mm, mm, mm)
	require.Equal(t, 9, reject)
	b.Add(mm, mm, mm, mm)
	require.Equal(t, 13, reject)
	b.Accept(batch)
	require.Equal(t, 13, reject)
	require.Equal(t, 5, accept)
}
