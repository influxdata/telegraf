package models

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/wal"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func newTestDiskBuffer(t testing.TB) Buffer {
	path, err := os.MkdirTemp("", "*-buffer-test")
	require.NoError(t, err)
	return newTestDiskBufferWithPath(t, "test", path)
}

func newTestDiskBufferWithPath(t testing.TB, name string, path string) Buffer {
	t.Helper()
	buf, err := NewBuffer(name, "", 0, "disk", path)
	require.NoError(t, err)
	buf.Stats().MetricsAdded.Set(0)
	buf.Stats().MetricsWritten.Set(0)
	buf.Stats().MetricsDropped.Set(0)
	return buf
}

func TestBuffer_RetainsTrackingInformation(t *testing.T) {
	var delivered int
	mm, _ := metric.WithTracking(Metric(), func(_ telegraf.DeliveryInfo) {
		delivered++
	})
	b := newTestDiskBuffer(t)
	b.Add(mm)
	batch := b.Batch(1)
	b.Accept(batch)
	require.Equal(t, 1, delivered)
}

func TestBuffer_TrackingDroppedFromOldWal(t *testing.T) {
	path, err := os.MkdirTemp("", "*-buffer-test")
	require.NoError(t, err)
	walfile, err := wal.Open(path, nil)
	require.NoError(t, err)

	tm, _ := metric.WithTracking(Metric(), func(_ telegraf.DeliveryInfo) {})

	metrics := []telegraf.Metric{
		// Basic metric with 1 field, 0 timestamp
		Metric(),
		// Basic metric with 1 field, different timestamp
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 20.0,
			},
			time.Now(),
		),
		// Metric with a field
		metric.New(
			"cpu",
			map[string]string{
				"x": "y",
			},
			map[string]interface{}{
				"value": 18.0,
			},
			time.Now(),
		),
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

	for i, m := range metrics {
		data, err := metric.ToBytes(m)
		require.NoError(t, err)
		require.NoError(t, walfile.Write(uint64(i+1), data))
	}

	b := newTestDiskBufferWithPath(t, filepath.Base(path), filepath.Dir(path))
	batch := b.Batch(4)
	// expected skips the tracking metric
	expected := []telegraf.Metric{
		metrics[0], metrics[1], metrics[2], metrics[4],
	}
	testutil.RequireMetricsEqual(t, expected, batch)
}
