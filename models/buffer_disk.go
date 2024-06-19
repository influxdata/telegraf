package models

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/tidwall/wal"
)

type DiskBuffer struct {
	BufferStats
	sync.Mutex

	walFile     *wal.Log
	walFilePath string

	batchFirst uint64 // Index of the first metric in the batch
	batchSize  uint64 // Number of metrics currently in the batch

	// Ending point of metrics read from disk on telegraf launch.
	// Used to know whether to discard tracking metrics.
	originalEnd uint64
}

func NewDiskBuffer(name string, path string, stats BufferStats) (*DiskBuffer, error) {
	filePath := path + "/" + name
	walFile, err := wal.Open(filePath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open wal file: %w", err)
	}
	return &DiskBuffer{
		BufferStats: stats,
		walFile:     walFile,
		walFilePath: filePath,
	}, nil
}

func (b *DiskBuffer) Len() int {
	b.Lock()
	defer b.Unlock()
	return b.length()
}

func (b *DiskBuffer) length() int {
	// Special case for when the read index is zero, it must be empty (otherwise it would be >= 1)
	if b.readIndex() == 0 {
		return 0
	}
	return int(b.writeIndex() - b.readIndex())
}

// readIndex is the first index to start reading metrics from, or the head of the buffer
func (b *DiskBuffer) readIndex() uint64 {
	index, err := b.walFile.FirstIndex()
	if err != nil {
		panic(err) // can only occur with a corrupt wal file
	}
	return index
}

// writeIndex is the first index to start writing metrics to, or the tail of the buffer
func (b *DiskBuffer) writeIndex() uint64 {
	index, err := b.walFile.LastIndex()
	if err != nil {
		panic(err) // can only occur with a corrupt wal file
	}
	return index + 1
}

func (b *DiskBuffer) Add(metrics ...telegraf.Metric) int {
	b.Lock()
	defer b.Unlock()

	dropped := 0
	for _, m := range metrics {
		if !b.addSingle(m) {
			dropped++
		}
	}
	b.BufferSize.Set(int64(b.length()))
	return dropped
	// todo implement batched writes
}

func (b *DiskBuffer) addSingle(m telegraf.Metric) bool {
	data, err := metric.ToBytes(m)
	if err != nil {
		panic(err)
	}
	err = b.walFile.Write(b.writeIndex(), data)
	if err == nil {
		b.metricAdded()
		return true
	}
	return false
}

//nolint:unused // to be implemented in the future
func (b *DiskBuffer) addBatch(metrics []telegraf.Metric) int {
	written := 0
	batch := new(wal.Batch)
	for _, m := range metrics {
		data, err := metric.ToBytes(m)
		if err != nil {
			panic(err)
		}
		batch.Write(b.writeIndex(), data)
		b.metricAdded()
		written++
	}
	err := b.walFile.WriteBatch(batch)
	if err != nil {
		return 0 // todo error handle, test if a partial write occur
	}
	return written
}

func (b *DiskBuffer) Batch(batchSize int) []telegraf.Metric {
	b.Lock()
	defer b.Unlock()

	if b.length() == 0 {
		// no metrics in the wal file, so return an empty array
		return make([]telegraf.Metric, 0)
	}
	b.batchFirst = b.readIndex()
	var metrics []telegraf.Metric

	b.batchSize = 0
	readIndex := b.batchFirst
	endIndex := b.writeIndex()
	for batchSize > 0 && readIndex < endIndex {
		data, err := b.walFile.Read(readIndex)
		if err != nil {
			panic(err)
		}
		readIndex++

		m, err := metric.FromBytes(data)
		if errors.Is(err, metric.ErrSkipTracking) {
			// could not look up tracking information for metric, skip
			continue
		}
		if err != nil {
			// non-recoverable error in deserialization, abort
			panic(err)
		}
		if _, ok := m.(telegraf.TrackingMetric); ok && readIndex < b.originalEnd {
			// tracking metric left over from previous instance, skip
			continue
		}

		metrics = append(metrics, m)
		b.batchSize++
		batchSize--
	}
	return metrics
}

func (b *DiskBuffer) Accept(batch []telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	if b.batchSize == 0 || len(batch) == 0 {
		// nothing to accept
		return
	}
	for _, m := range batch {
		b.metricWritten(m)
	}
	if b.length() == len(batch) {
		b.resetWalFile()
	} else {
		err := b.walFile.TruncateFront(b.batchFirst + uint64(len(batch)))
		if err != nil {
			panic(err)
		}
	}

	// check if the original end index is still valid, clear if not
	if b.originalEnd < b.readIndex() {
		b.originalEnd = 0
	}

	b.resetBatch()
	b.BufferSize.Set(int64(b.length()))
}

func (b *DiskBuffer) Reject(_ []telegraf.Metric) {
	// very little to do here as the disk buffer retains metrics in
	// the wal file until a call to accept
	b.Lock()
	defer b.Unlock()
	b.resetBatch()
}

func (b *DiskBuffer) Stats() BufferStats {
	return b.BufferStats
}

func (b *DiskBuffer) resetBatch() {
	b.batchFirst = 0
	b.batchSize = 0
}

// todo This is very messy and not ideal, but serves as the only way I can find currently
// todo to actually clear the walfile completely if needed, since Truncate() calls require
// todo at least one entry remains in them otherwise they return an error.
func (b *DiskBuffer) resetWalFile() {
	b.walFile.Close()
	os.Remove(b.walFilePath)
	walFile, err := wal.Open(b.walFilePath, nil)
	if err != nil {
		panic(err)
	}
	b.walFile = walFile
}
