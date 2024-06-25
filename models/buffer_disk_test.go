package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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

// WAL file tested here was written as:
// 1: Metric()
// 2: Metric()
// 3: Metric()
// 4: metric.WithTracking(Metric())
// 5: Metric()
//
// Expected to drop the 4th metric, as tracking metrics from
// previous instances  are dropped when the wal file is reopened.
func TestBuffer_TrackingDroppedFromOldWal(t *testing.T) {
	// copy the testdata so we do not destroy the testdata wal file
	path, err := os.MkdirTemp("", "*-buffer-test")
	require.NoError(t, err)
	f, err := os.Create(path + "/00000000000000000001")
	require.NoError(t, err)
	f1, err := os.Open("testdata/testwal/00000000000000000001")
	require.NoError(t, err)
	written, err := io.Copy(f, f1)
	require.NoError(t, err)
	fmt.Println(written)

	b := newTestDiskBufferWithPath(t, filepath.Base(path), filepath.Dir(path))
	batch := b.Batch(4)
	expected := []telegraf.Metric{
		Metric(), Metric(), Metric(), Metric(),
	}
	testutil.RequireMetricsEqual(t, expected, batch)
}

/*
// Function used to create the test data used in the test above
func Test_CreateTestData(t *testing.T) {
	metric.Init()
	walfile, _ := wal.Open("testdata/testwal", nil)
	data, err := metric.ToBytes(Metric())
	require.NoError(t, err)
	require.NoError(t, walfile.Write(1, data))
	data, err = metric.ToBytes(Metric())
	require.NoError(t, err)
	require.NoError(t, walfile.Write(2, data))
	data, err = metric.ToBytes(Metric())
	require.NoError(t, err)
	require.NoError(t, walfile.Write(3, data))
	m, _ := metric.WithTracking(Metric(), func(di telegraf.DeliveryInfo) {})
	data, err = metric.ToBytes(m)
	require.NoError(t, err)
	require.NoError(t, walfile.Write(4, data))
	data, err = metric.ToBytes(Metric())
	require.NoError(t, err)
	require.NoError(t, walfile.Write(5, data))
}
*/
