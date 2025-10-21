package models

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"sort"
	"sync"

	"github.com/tidwall/wal"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type DiskBuffer struct {
	BufferStats
	sync.Mutex

	file *wal.Log
	path string

	batchFirst uint64 // Index of the first metric in the batch
	batchSize  uint64 // Number of metrics currently in the batch

	// Ending point of metrics read from disk on telegraf launch.
	// Used to know whether to discard tracking metrics.
	originalEnd uint64

	// The mask contains offsets of metric already removed during a previous
	// transaction. Metrics at those offsets should not be contained in new
	// batches.
	mask []int
}

func NewDiskBuffer(id, path string, stats BufferStats) (*DiskBuffer, error) {
	filePath := filepath.Join(path, id)
	walFile, err := wal.Open(filePath, &wal.Options{
		AllowEmpty: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open wal file: %w", err)
	}

	buf := &DiskBuffer{
		BufferStats: stats,
		file:        walFile,
		path:        filePath,
	}
	if buf.Len() > 0 {
		buf.originalEnd = buf.writeIndex()
	}
	return buf, nil
}

func (b *DiskBuffer) Len() int {
	b.Lock()
	defer b.Unlock()
	return b.length()
}

func (b *DiskBuffer) length() int {
	return b.entries() - len(b.mask)
}

func (b *DiskBuffer) entries() int {
	if b.readIndex() == 0 {
		return 0
	}
	return int(b.writeIndex() - b.readIndex())
}

// readIndex is the first index to start reading metrics from, or the head of the buffer
func (b *DiskBuffer) readIndex() uint64 {
	index, err := b.file.FirstIndex()
	if err != nil {
		panic(err) // can only occur with a corrupt or closed wal file
	}
	return index
}

// writeIndex is the first index to start writing metrics to, or the tail of the buffer
func (b *DiskBuffer) writeIndex() uint64 {
	index, err := b.file.LastIndex()
	if err != nil {
		panic(err) // can only occur with a corrupt or closed wal file
	}
	return index + 1
}

func (b *DiskBuffer) Add(metrics ...telegraf.Metric) int {
	b.Lock()
	defer b.Unlock()

	dropped := 0
	for _, m := range metrics {
		if !b.addSingleMetric(m) {
			dropped++
		}
	}
	b.BufferSize.Set(int64(b.length()))
	return dropped
}

func (b *DiskBuffer) addSingleMetric(m telegraf.Metric) bool {
	data, err := metric.ToBytes(m)
	if err != nil {
		panic(err)
	}
	if err := b.file.Write(b.writeIndex(), data); err != nil {
		return false
	}
	b.metricAdded()
	return true
}

func (b *DiskBuffer) BeginTransaction(batchSize int) *Transaction {
	b.Lock()
	defer b.Unlock()

	if b.length() == 0 {
		return &Transaction{}
	}
	b.batchFirst = b.readIndex()
	b.batchSize = 0

	metrics := make([]telegraf.Metric, 0, batchSize)
	offsets := make([]int, 0, batchSize)
	readIndex := b.batchFirst
	endIndex := b.writeIndex()
	for offset := 0; batchSize > 0 && readIndex < endIndex; offset++ {
		data, err := b.file.Read(readIndex)
		if err != nil {
			panic(err)
		}
		readIndex++

		if slices.Contains(b.mask, offset) {
			// Metric is masked by a previous write and is scheduled for removal
			continue
		}

		// Validate that a tracking metric is from this instance of telegraf and skip ones from older instances.
		// A tracking metric can be skipped here because metric.Accept() is only called once data is successfully
		// written to an output, so any tracking metrics from older instances can be dropped and reacquired to
		// have an accurate tracking information.
		// There are two primary cases here:
		// - ErrSkipTracking:  means that the tracking information was unable to be found for a tracking ID.
		// - Outside of range: means that the metric was guaranteed to be left over from the previous instance
		//                     as it was here when we opened the wal file in this instance.
		m, err := metric.FromBytes(data)
		if err != nil {
			if errors.Is(err, metric.ErrSkipTracking) {
				// Could not look up tracking information for metric so skip
				// the metric and mask it so it is truncated later on.
				b.mask = append(b.mask, offset)
				continue
			}
			// non-recoverable error in deserialization, abort
			log.Printf("E! raw metric data: %v", data)
			panic(err)
		}
		if _, ok := m.(telegraf.TrackingMetric); ok && readIndex < b.originalEnd {
			// This tracking metric is a left-over from a previous instance e.g.
			// after restarting Telegraf. Skip the metric and mask it so it is
			// trucated later on
			b.mask = append(b.mask, offset)
			continue
		}

		metrics = append(metrics, m)
		offsets = append(offsets, offset)
		b.batchSize++
		batchSize--
	}
	return &Transaction{Batch: metrics, valid: true, state: offsets}
}

func (b *DiskBuffer) EndTransaction(tx *Transaction) {
	if len(tx.Batch) == 0 {
		return
	}

	// Ignore invalid transactions and make sure they can only be finished once
	if !tx.valid {
		return
	}
	tx.valid = false

	// Get the metric offsets from the transaction
	offsets := tx.state.([]int)

	b.Lock()
	defer b.Unlock()

	// Mark metrics which should be removed in the internal mask
	remove := make([]int, 0, len(tx.Accept)+len(tx.Reject))
	for _, idx := range tx.Accept {
		b.metricWritten(tx.Batch[idx])
		remove = append(remove, offsets[idx])
	}
	for _, idx := range tx.Reject {
		b.metricRejected(tx.Batch[idx])
		remove = append(remove, offsets[idx])
	}
	b.mask = append(b.mask, remove...)
	sort.Ints(b.mask)

	// Remove the metrics that are marked for removal from the front of the
	// WAL file. All other metrics must be kept.
	if len(b.mask) == 0 || b.mask[0] != 0 {
		// Mask is empty or the first index is not the front of the file, so
		// exit early as there is nothing to remove
		return
	}

	// Determine up to which index we can remove the entries from the WAL file
	var correction int
	for i, offset := range b.mask {
		if offset != i {
			break
		}
		correction = offset
	}
	// The 'correction' denotes the offset to subtract from the remaining mask
	// (if any) and the 'removalIdx' denotes the index to use when truncating
	// the file and mask. Keep them separate to be able to handle the special
	// "the file cannot be empty" property of the WAL file.
	removeIdx := correction + 1

	// Remove the metrics in front from the WAL file
	if err := b.file.TruncateFront(b.batchFirst + uint64(removeIdx)); err != nil {
		log.Printf("E! batch length: %d, first: %d, size: %d", len(tx.Batch), b.batchFirst, b.batchSize)
		panic(err)
	}

	// Truncate the mask and update the relative offsets
	b.mask = b.mask[removeIdx:]
	for i := range b.mask {
		b.mask[i] -= correction
	}

	// check if the original end index is still valid, clear if not
	if b.originalEnd < b.readIndex() {
		b.originalEnd = 0
	}

	b.resetBatch()
	b.BufferSize.Set(int64(b.length()))
}

func (b *DiskBuffer) Stats() BufferStats {
	return b.BufferStats
}

func (b *DiskBuffer) Close() error {
	if err := b.file.Close(); err != nil {
		return fmt.Errorf("closing buffer failed: %w", err)
	}
	return nil
}

func (b *DiskBuffer) resetBatch() {
	b.batchFirst = 0
	b.batchSize = 0
}
