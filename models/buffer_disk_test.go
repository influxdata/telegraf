package models

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/wal"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestDiskBufferRetainsTrackingInformation(t *testing.T) {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))

	var delivered int
	mm, _ := metric.WithTracking(m, func(telegraf.DeliveryInfo) { delivered++ })

	buf, err := NewBuffer("test", "123", "", 0, "disk", t.TempDir())
	require.NoError(t, err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	defer buf.Close()

	buf.Add(mm)
	tx := buf.BeginTransaction(1)
	tx.AcceptAll()
	buf.EndTransaction(tx)
	require.Equal(t, 1, delivered)
}

func TestDiskBufferTrackingDroppedFromOldWal(t *testing.T) {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))

	tm, _ := metric.WithTracking(m, func(telegraf.DeliveryInfo) {})
	metrics := []telegraf.Metric{
		// Basic metric with 1 field, 0 timestamp
		metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0)),
		// Basic metric with 1 field, different timestamp
		metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 20.0}, time.Now()),
		// Metric with a field
		metric.New("cpu", map[string]string{"x": "y"}, map[string]interface{}{"value": 18.0}, time.Now()),
		// Tracking metric
		tm,
		// Metric with lots of tag types
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value_f64":        20.0,
				"value_uint64":     uint64(10),
				"value_int16":      int16(5),
				"value_string":     "foo",
				"value_boolean":    true,
				"value_byte_array": []byte{1, 2, 3, 4, 5},
			},
			time.Now(),
		),
	}

	// call manually so that we can properly use metric.ToBytes() without having initialized a buffer
	registerGob()

	// Prefill the WAL file
	path := t.TempDir()
	walfile, err := wal.Open(filepath.Join(path, "123"), nil)
	require.NoError(t, err)
	defer walfile.Close()
	for i, m := range metrics {
		data, err := metric.ToBytes(m)
		require.NoError(t, err)
		require.NoError(t, walfile.Write(uint64(i+1), data))
	}
	walfile.Close()

	// Create a buffer
	buf, err := NewBuffer("123", "123", "", 0, "disk", path)
	require.NoError(t, err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	defer buf.Close()

	tx := buf.BeginTransaction(4)

	// Check that the tracking metric is skipped
	expected := []telegraf.Metric{
		metrics[0], metrics[1], metrics[2], metrics[4],
	}
	testutil.RequireMetricsEqual(t, expected, tx.Batch)
}

// TestDiskBufferTruncate is a regression test for
// https://github.com/influxdata/telegraf/issues/16696
func TestDiskBufferTruncate(t *testing.T) {
	// Create a disk buffer
	buf, err := NewBuffer("test", "id123", "", 0, "disk", t.TempDir())
	require.NoError(t, err)
	defer buf.Close()
	diskBuf, ok := buf.(*DiskBuffer)
	require.True(t, ok, "buffer is not a disk buffer")

	// Add some metrics to the buffer
	expected := make([]telegraf.Metric, 0, 10)
	for i := range 10 {
		m := metric.New("test", map[string]string{}, map[string]interface{}{"value": i}, time.Now())
		buf.Add(m)
		expected = append(expected, m)
	}

	// Get a batch, test the metrics and acknowledge all metrics
	tx := buf.BeginTransaction(4)
	testutil.RequireMetricsEqual(t, expected[:4], tx.Batch)
	tx.AcceptAll()
	buf.EndTransaction(tx)

	// The buffer must have been truncated on disk and the mask should be empty
	require.Equal(t, 6, diskBuf.entries())
	require.Empty(t, diskBuf.mask)

	// Get a second batch, test the metrics and acknowledge all metrics except
	// for the first one.
	tx = buf.BeginTransaction(4)
	testutil.RequireMetricsEqual(t, expected[4:8], tx.Batch)
	tx.Accept = []int{1, 2, 3}
	buf.EndTransaction(tx)

	// The buffer cannot be truncated on disk as the first metric must be kept.
	// However, the mask now must contain the accepted indices...
	require.Equal(t, 6, diskBuf.entries())
	require.Equal(t, []int{1, 2, 3}, diskBuf.mask)

	// Get a third batch with all the remaining metrics, test them and
	// acknowledge all
	tx = buf.BeginTransaction(4)
	remaining := append([]telegraf.Metric{expected[4]}, expected[8:]...)
	testutil.RequireMetricsEqual(t, remaining, tx.Batch)
	tx.AcceptAll()
	buf.EndTransaction(tx)

	// The buffer should be truncated completely, however due to the WAL
	// implementation the file cannot be completely empty. So we expect one
	// entry left on disk but this one being masked...
	require.Equal(t, 1, diskBuf.entries())
	require.Equal(t, []int{0}, diskBuf.mask)

	// We shouldn't get any metric when requesting a new batch
	tx = buf.BeginTransaction(4)
	require.Empty(t, tx.Batch)
}
