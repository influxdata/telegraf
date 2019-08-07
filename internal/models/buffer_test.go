package models

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
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
	return MetricTime(0)
}

func MetricTime(sec int64) telegraf.Metric {
	m, err := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(sec, 0),
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

func TestBuffer_BatchLatest(t *testing.T) {
	b := setup(NewBuffer("test", 4))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)

	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(3),
			MetricTime(2),
		}, batch)
}

func TestBuffer_BatchLatestWrap(t *testing.T) {
	b := setup(NewBuffer("test", 4))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	batch := b.Batch(2)

	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(5),
			MetricTime(4),
		}, batch)
}

func TestBuffer_MultipleBatch(t *testing.T) {
	b := setup(NewBuffer("test", 10))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	batch := b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(6),
			MetricTime(5),
			MetricTime(4),
			MetricTime(3),
			MetricTime(2),
		}, batch)
	b.Accept(batch)
	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(1),
		}, batch)
	b.Accept(batch)
}

func TestBuffer_RejectWithRoom(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Reject(batch)

	require.Equal(t, int64(0), b.MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(5),
			MetricTime(4),
			MetricTime(3),
			MetricTime(2),
			MetricTime(1),
		}, batch)
}

func TestBuffer_RejectNothingNewFull(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	batch := b.Batch(2)
	b.Reject(batch)

	require.Equal(t, int64(0), b.MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(5),
			MetricTime(4),
			MetricTime(3),
			MetricTime(2),
			MetricTime(1),
		}, batch)
}

func TestBuffer_RejectNoRoom(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	b.Add(MetricTime(1))

	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)

	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Add(MetricTime(8))

	b.Reject(batch)

	require.Equal(t, int64(3), b.MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(8),
			MetricTime(7),
			MetricTime(6),
			MetricTime(5),
			MetricTime(4),
		}, batch)
}

func TestBuffer_RejectRoomExact(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	batch := b.Batch(2)
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))

	b.Reject(batch)

	require.Equal(t, int64(0), b.MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(5),
			MetricTime(4),
			MetricTime(3),
			MetricTime(2),
			MetricTime(1),
		}, batch)
}

func TestBuffer_RejectRoomOverwriteOld(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(1)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))

	b.Reject(batch)

	require.Equal(t, int64(1), b.MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(6),
			MetricTime(5),
			MetricTime(4),
			MetricTime(3),
			MetricTime(2),
		}, batch)
}

func TestBuffer_RejectPartialRoom(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	b.Add(MetricTime(1))

	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)

	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Reject(batch)

	require.Equal(t, int64(2), b.MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(7),
			MetricTime(6),
			MetricTime(5),
			MetricTime(4),
			MetricTime(3),
		}, batch)
}

func TestBuffer_RejectNewMetricsWrapped(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(2)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))

	// buffer: 1, 4, 5; batch: 2, 3
	require.Equal(t, int64(0), b.MetricsDropped.Get())

	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Add(MetricTime(8))
	b.Add(MetricTime(9))
	b.Add(MetricTime(10))

	// buffer: 8, 9, 10, 6, 7; batch: 2, 3
	require.Equal(t, int64(3), b.MetricsDropped.Get())

	b.Add(MetricTime(11))
	b.Add(MetricTime(12))
	b.Add(MetricTime(13))
	b.Add(MetricTime(14))
	b.Add(MetricTime(15))
	// buffer: 13, 14, 15, 11, 12; batch: 2, 3
	require.Equal(t, int64(8), b.MetricsDropped.Get())
	b.Reject(batch)

	require.Equal(t, int64(10), b.MetricsDropped.Get())

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(15),
			MetricTime(14),
			MetricTime(13),
			MetricTime(12),
			MetricTime(11),
		}, batch)
}

func TestBuffer_RejectWrapped(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))

	b.Add(MetricTime(6))
	b.Add(MetricTime(7))
	b.Add(MetricTime(8))
	batch := b.Batch(3)

	b.Add(MetricTime(9))
	b.Add(MetricTime(10))
	b.Add(MetricTime(11))
	b.Add(MetricTime(12))

	b.Reject(batch)

	batch = b.Batch(5)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(12),
			MetricTime(11),
			MetricTime(10),
			MetricTime(9),
			MetricTime(8),
		}, batch)
}

func TestBuffer_RejectAdjustFirst(t *testing.T) {
	b := setup(NewBuffer("test", 10))
	b.Add(MetricTime(1))
	b.Add(MetricTime(2))
	b.Add(MetricTime(3))
	batch := b.Batch(3)
	b.Add(MetricTime(4))
	b.Add(MetricTime(5))
	b.Add(MetricTime(6))
	b.Reject(batch)

	b.Add(MetricTime(7))
	b.Add(MetricTime(8))
	b.Add(MetricTime(9))
	batch = b.Batch(3)
	b.Add(MetricTime(10))
	b.Add(MetricTime(11))
	b.Add(MetricTime(12))
	b.Reject(batch)

	b.Add(MetricTime(13))
	b.Add(MetricTime(14))
	b.Add(MetricTime(15))
	batch = b.Batch(3)
	b.Add(MetricTime(16))
	b.Add(MetricTime(17))
	b.Add(MetricTime(18))
	b.Reject(batch)

	b.Add(MetricTime(19))

	batch = b.Batch(10)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(19),
			MetricTime(18),
			MetricTime(17),
			MetricTime(16),
			MetricTime(15),
			MetricTime(14),
			MetricTime(13),
			MetricTime(12),
			MetricTime(11),
			MetricTime(10),
		}, batch)
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
	require.Equal(t, int64(0), b.MetricsDropped.Get(), "dropped")
	require.Equal(t, int64(3), b.MetricsWritten.Get(), "written")
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
	require.Equal(t, 2, reject)
	b.Reject(batch)
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

func TestBuffer_RejectEmptyBatch(t *testing.T) {
	b := setup(NewBuffer("test", 5))
	batch := b.Batch(2)
	b.Add(MetricTime(1))
	b.Reject(batch)
	b.Add(MetricTime(2))
	batch = b.Batch(2)
	for _, m := range batch {
		require.NotNil(t, m)
	}
}
