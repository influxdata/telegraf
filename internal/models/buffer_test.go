package models

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

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

func MetricUnix(seconds int64) telegraf.Metric {
	m, err := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(seconds, 0),
	)
	if err != nil {
		panic(err)
	}
	return m
}

func BenchmarkAddMetrics(b *testing.B) {
	buf := NewBuffer(10000)
	m := Metric()
	for n := 0; n < b.N; n++ {
		buf.Add(m)
	}
}

func TestBuffer_LenEmpty(t *testing.T) {
	b := NewBuffer(5)
	require.Equal(t, 0, b.Len())
}

func TestBuffer_LenOne(t *testing.T) {
	m := Metric()

	b := NewBuffer(5)
	b.Add(m)
	require.Equal(t, 1, b.Len())
}

func TestBuffer_LenFull(t *testing.T) {
	m := Metric()

	b := NewBuffer(5)
	b.Add(m, m, m, m, m)
	require.Equal(t, 5, b.Len())
}

func TestBuffer_LenOverfill(t *testing.T) {
	m := Metric()

	b := NewBuffer(5)
	b.Add(m, m, m, m, m, m)
	require.Equal(t, 5, b.Len())
}

func TestBuffer_BatchEmpty(t *testing.T) {
	b := NewBuffer(5)
	batch := b.Batch(2)
	require.Len(t, batch, 0)
}

func TestBuffer_BatchUnderfill(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	b.Add(m)
	batch := b.Batch(2)
	require.Len(t, batch, 1)
}

func TestBuffer_BatchExact(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	b.Add(m, m)
	batch := b.Batch(2)
	require.Len(t, batch, 2)
}

func TestBuffer_BatchPartial(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	require.Len(t, batch, 2)
}

func TestBuffer_BatchOverSize(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(6)
	require.Len(t, batch, 5)
}

func TestBuffer_BatchReject(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	b.Add(m, m, m)
	batch := b.Batch(2)
	b.Add(m, m)
	b.Reject(batch)
	require.Equal(t, 5, b.Len())
	batch = b.Batch(3)
	require.Equal(t, 3, len(batch))
}

func TestBuffer_BatchWrapped(t *testing.T) {
	b := NewBuffer(5)
	b.Add(
		MetricUnix(1),
		MetricUnix(2),
		MetricUnix(3),
		MetricUnix(4),
		MetricUnix(5))
	batch := b.Batch(2)
	require.Len(t, batch, 2)
	require.Equal(t, int64(1), batch[0].Time().Unix())
	require.Equal(t, int64(2), batch[1].Time().Unix())
	b.Ack()
	b.Add(
		MetricUnix(6),
		MetricUnix(7))
	batch = b.Batch(5)
	require.Len(t, batch, 5)
	require.Equal(t, int64(3), batch[0].Time().Unix())
	require.Equal(t, int64(4), batch[1].Time().Unix())
	require.Equal(t, int64(5), batch[2].Time().Unix())
	require.Equal(t, int64(6), batch[3].Time().Unix())
	require.Equal(t, int64(7), batch[4].Time().Unix())
}

func TestBuffer_BatchOverwriteBatch(t *testing.T) {
	b := NewBuffer(5)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	b.Add(
		MetricUnix(1),
		MetricUnix(2),
		MetricUnix(3),
		MetricUnix(4),
		MetricUnix(5))
	batch := b.Batch(2)
	require.Len(t, batch, 2)
	require.Equal(t, int64(1), batch[0].Time().Unix())
	require.Equal(t, int64(2), batch[1].Time().Unix())
	b.Reject(batch)
	b.Add(
		MetricUnix(6), // overwrite 1
		MetricUnix(7), // overwrite 2
		MetricUnix(8), // overwrite 3
		MetricUnix(9)) // overwrite 4
	batch = b.Batch(2)
	require.Len(t, batch, 2)
	require.Equal(t, int64(5), batch[0].Time().Unix())
	require.Equal(t, int64(6), batch[1].Time().Unix())

	require.Equal(t, int64(4), MetricsDropped.Get())
	require.Equal(t, int64(9), MetricsWritten.Get())
}

func TestBuffer_MetricsOverwrite(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	b.Add(m, m, m, m, m)
	b.Add(m, m, m)
	require.Equal(t, int64(3), MetricsDropped.Get())
	require.Equal(t, int64(8), MetricsWritten.Get())
}

func TestBuffer_MetricsOverwriteBatchAck(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	b.Add(m, m, m, m, m)
	b.Batch(3)
	b.Add(m, m, m)
	b.Ack()
	require.Equal(t, int64(0), MetricsDropped.Get())
	require.Equal(t, int64(8), MetricsWritten.Get())
}

func TestBuffer_MetricsOverwriteBatchReject(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	b.Add(m, m, m, m, m)
	batch := b.Batch(3)
	b.Add(m, m, m)
	b.Reject(batch)
	require.Equal(t, int64(3), MetricsDropped.Get())
	require.Equal(t, int64(8), MetricsWritten.Get())
}

func TestBuffer_MetricsBatchAckRemoved(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	MetricsDropped.Set(0)
	MetricsWritten.Set(0)

	b.Add(m, m, m, m, m)
	b.Batch(3)
	b.Add(m, m, m, m, m)
	b.Ack()
	require.Equal(t, int64(0), MetricsDropped.Get())
	require.Equal(t, int64(10), MetricsWritten.Get())
}

func TestBuffer_BatchKept(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	b.Add(m, m, m, m, m)
	b.Batch(2)
	require.Equal(t, 5, b.Len())
}

func TestBuffer_BatchRemovedOnAck(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	b.Add(m, m, m, m, m)
	b.Batch(2)
	b.Ack()
	require.Equal(t, 3, b.Len())
}

func TestBuffer_BatchRejectAckNoop(t *testing.T) {
	m := Metric()
	b := NewBuffer(5)
	b.Add(m, m, m, m, m)
	batch := b.Batch(2)
	b.Reject(batch)
	b.Ack()
	require.Equal(t, 5, b.Len())
}
