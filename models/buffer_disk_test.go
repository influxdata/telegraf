package models

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/wal"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// TestDiskBufferTruncate is a regression test for
// https://github.com/influxdata/telegraf/issues/16696
func TestDiskBufferTruncate(t *testing.T) {
	// Create a disk buffer
	buf, err := NewBuffer("test", "id123", "", 0, "disk_write_through", t.TempDir())
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

	// Ensure the buffer was fully truncated on disk and the mask is consistent with that
	require.Zero(t, diskBuf.entries())
	require.Empty(t, diskBuf.mask)

	// We shouldn't get any metric when requesting a new batch
	tx = buf.BeginTransaction(4)
	require.Empty(t, tx.Batch)
}

// TestDiskBufferEmptyReuse is a regression test for making sure all metrics are
// output after being added to an fully drained (i.e. empty) buffer. Related to
// https://github.com/influxdata/telegraf/issues/16981
func TestDiskBufferEmptyReuse(t *testing.T) {
	// Create a disk buffer
	buf, err := NewBuffer("test", "id123", "", 0, "disk_write_through", t.TempDir())
	require.NoError(t, err)
	defer buf.Close()
	diskBuf, ok := buf.(*DiskBuffer)
	require.True(t, ok, "buffer is not a disk buffer")

	// Add some metrics to the buffer
	expected := make([]telegraf.Metric, 0, 5)
	for i := range 5 {
		m := metric.New("test", map[string]string{}, map[string]interface{}{"value": i}, time.Now())
		buf.Add(m)
		expected = append(expected, m)
	}

	// Read the complete set of metrics such that the buffer is empty again
	tx := buf.BeginTransaction(5)
	testutil.RequireMetricsEqual(t, expected, tx.Batch)
	tx.AcceptAll()
	buf.EndTransaction(tx)

	// Ensure all storage elements of the buffer are consistent with it being empty
	require.Zero(t, diskBuf.length())
	require.Zero(t, diskBuf.entries())
	require.Empty(t, diskBuf.mask)

	// Try to read the buffer again. This should return an empty transaction...
	tx = buf.BeginTransaction(5)
	require.Empty(t, tx.Batch)
	buf.EndTransaction(tx)

	// Now add another set of metrics and make sure we can read it
	m := metric.New("test", map[string]string{}, map[string]interface{}{"value": 42}, time.Now())
	buf.Add(m)

	// Read the complete set of metrics such that the buffer is empty again
	tx = buf.BeginTransaction(5)
	testutil.RequireMetricsEqual(t, []telegraf.Metric{m}, tx.Batch)
	tx.AcceptAll()
	buf.EndTransaction(tx)
}

// TestDiskBufferEmptyClose is a regression test for making sure that we do not
// encounter any metrics in a reopened buffer if it was closed in an empty state.
// This should be the normal case if all metrics are successfully written when
// stopping Telegraf. On next startup the buffer should be empty. Related to
// https://github.com/influxdata/telegraf/issues/16981
func TestDiskBufferEmptyClose(t *testing.T) {
	tmpdir := t.TempDir()

	// Create a disk buffer
	buf, err := NewBuffer("test", "id123", "", 0, "disk_write_through", tmpdir)
	require.NoError(t, err)
	defer buf.Close()
	diskBuf, ok := buf.(*DiskBuffer)
	require.True(t, ok, "buffer is not a disk buffer")

	// Add some metrics to the buffer
	expected := make([]telegraf.Metric, 0, 5)
	for i := range 5 {
		m := metric.New("test", map[string]string{}, map[string]interface{}{"value": i}, time.Now())
		buf.Add(m)
		expected = append(expected, m)
	}

	// Read the complete set of metrics such that the buffer is empty again
	tx := buf.BeginTransaction(5)
	testutil.RequireMetricsEqual(t, expected, tx.Batch)
	tx.AcceptAll()
	buf.EndTransaction(tx)

	// Make sure the buffer was fully emptied
	require.Zero(t, diskBuf.length())
	require.Zero(t, diskBuf.entries())

	// Close the buffer to simulate stopping Telegraf in a normal shutdown
	require.NoError(t, diskBuf.Close())

	// Reopen the buffer with the parameters above to see the same buffer
	reopened, err := NewBuffer("test", "id123", "", 0, "disk_write_through", tmpdir)
	require.NoError(t, err)
	defer reopened.Close()
	_, ok = reopened.(*DiskBuffer)
	require.True(t, ok, "reopened buffer is not a disk buffer")

	// Try to read the buffer again. This should return an empty transaction...
	tx = reopened.BeginTransaction(5)
	require.Empty(t, tx.Batch)
	reopened.EndTransaction(tx)

	// However, adding a new metric to the buffer should work
	// Now add another set of metrics and make sure we can read it
	m := metric.New("test", map[string]string{}, map[string]interface{}{"value": 42}, time.Now())
	reopened.Add(m)

	// Read the complete set of metrics such that the buffer is empty again
	tx = reopened.BeginTransaction(5)
	testutil.RequireMetricsEqual(t, []telegraf.Metric{m}, tx.Batch)
	tx.AcceptAll()
	reopened.EndTransaction(tx)
}

func TestDiskBufferRetainsTrackingInformation(t *testing.T) {
	m := metric.New("cpu", map[string]string{}, map[string]interface{}{"value": 42.0}, time.Unix(0, 0))

	var delivered int
	mm, _ := metric.WithTracking(m, func(telegraf.DeliveryInfo) { delivered++ })

	buf, err := NewBuffer("test", "123", "", 0, "disk_write_through", t.TempDir())
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
	walfile, err := wal.Open(filepath.Join(path, "123"), &wal.Options{
		AllowEmpty: true,
	})
	require.NoError(t, err)
	defer walfile.Close()
	for i, m := range metrics {
		data, err := metric.ToBytes(m)
		require.NoError(t, err)
		require.NoError(t, walfile.Write(uint64(i+1), data))
	}
	walfile.Close()

	// Create a buffer
	buf, err := NewBuffer("123", "123", "", 0, "disk_write_through", path)
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

// TestDiskBufferTrackingOnOutputOutage is a regression test for making sure
// that we send all metrics if an output goes down and comes up again. In this
// special test we use tracking metrics as e.g. used for Kafka or MQTT.
// Related to https://github.com/influxdata/telegraf/issues/16981
func TestDiskBufferTrackingOnOutputOutage(t *testing.T) {
	// Make sure we can serialize the metrics by manually registering the binary
	// serializer. In real-world this is done during setting up Telegraf.
	registerGob()

	// Create some tracking metrics with a callback that records the accepted
	// metrics (or at least the tracking ID).
	const count = 10

	var mu sync.Mutex

	created := make([]telegraf.TrackingID, 0, count)
	delivered := make([]telegraf.TrackingID, 0, count)
	inputs := make([]telegraf.Metric, 0, count)
	expected := make([]telegraf.Metric, 0, count)
	for i := range count {
		m := metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{"value": i},
			time.Unix(0, 0),
		)
		tm, tid := metric.WithTracking(m, func(di telegraf.DeliveryInfo) {
			mu.Lock()
			defer mu.Unlock()
			t.Logf("delivered metric %v successfully: %v", di.ID(), di.Delivered())
			delivered = append(delivered, di.ID())
		})
		t.Logf("tracking metric %v", tid)
		inputs = append(inputs, tm)
		expected = append(expected, tm)
		created = append(created, tid)
	}

	// Create a disk buffer
	buf, err := NewBuffer("test", "id123", "", 0, "disk_write_through", t.TempDir())
	require.NoError(t, err)
	defer buf.Close()
	diskBuf, ok := buf.(*DiskBuffer)
	require.True(t, ok, "buffer is not a disk buffer")

	// Make sure the new buffer is fully empty
	require.Zero(t, diskBuf.length())
	require.Zero(t, diskBuf.entries())

	// Add a first metric and make sure we get it on transaction. Accept the
	// metric simulating that the buffer is up.
	t.Log("checking first accepted metric")
	require.Zero(t, buf.Add(inputs[0]))
	tx := buf.BeginTransaction(count)
	testutil.RequireMetricsEqual(t, expected[:1], tx.Batch)
	tx.AcceptAll()
	buf.EndTransaction(tx)

	// Add the remaining metrics except the last one
	middle := inputs[1 : count-1]
	middleExpected := expected[1 : count-1]
	require.Zero(t, buf.Add(middle...))

	// Get the metrics into a batch and keep them to simulate the output was
	// not able to deliver the metrics.
	t.Log("checking rejected batch")
	tx = buf.BeginTransaction(count)
	testutil.RequireMetricsEqual(t, middleExpected, tx.Batch)
	tx.KeepAll()
	buf.EndTransaction(tx)

	// Make sure we see the same, kept metrics again on next read
	t.Log("checking rejected batch a second time")
	tx = buf.BeginTransaction(count)
	testutil.RequireMetricsEqual(t, middleExpected, tx.Batch)
	tx.KeepAll()
	buf.EndTransaction(tx)

	// Now read the metrics again but this time we accept the metric to simulate
	// the output is back up again.
	t.Log("checking rejected batch but now accept")
	tx = buf.BeginTransaction(count)
	testutil.RequireMetricsEqual(t, middleExpected, tx.Batch)
	tx.AcceptAll()
	buf.EndTransaction(tx)

	// Add the last metric to the buffer, read it into a batch and accept it
	t.Log("checking last accepted metric")
	require.Zero(t, buf.Add(inputs[count-1:]...))
	tx = buf.BeginTransaction(count)
	testutil.RequireMetricsEqual(t, expected[count-1:], tx.Batch)
	tx.AcceptAll()
	buf.EndTransaction(tx)

	// Check that we got a delivery signal for all of the metrics
	mu.Lock()
	defer mu.Unlock()
	require.ElementsMatch(t, created, delivered, "tracking information mismatch")
}
