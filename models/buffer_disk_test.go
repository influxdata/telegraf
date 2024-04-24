package models

import (
	"os"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func newTestDiskBuffer(name string, alias string, capacity int) (*DiskBuffer, string) {
	bm := NewBufferMetrics(name, alias, capacity)
	path, _ := os.MkdirTemp("", "*-disk_buffer-"+name)
	return NewDiskBuffer(name, capacity, path, bm), path
}

func setupDisk(b *DiskBuffer) *DiskBuffer {
	b.MetricsAdded.Set(0)
	b.MetricsWritten.Set(0)
	b.MetricsDropped.Set(0)
	return b
}

func TestDiskBuffer(t *testing.T) {
	db, path := newTestDiskBuffer("test", "", 5000)
	defer os.RemoveAll(path)
	db = setupDisk(db)

	db.Add(MetricTime(1))
	db.Add(MetricTime(2))
	db.Add(MetricTime(3))
	batch := db.Batch(2)
	testutil.RequireMetricsEqual(t,
		[]telegraf.Metric{
			MetricTime(1),
			MetricTime(2),
		}, batch)
}
