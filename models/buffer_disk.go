package models

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/tidwall/wal"
)

type DiskBuffer struct {
	BufferMetrics

	walFile *wal.Log
}

func NewDiskBuffer(name string, capacity int, path string, metrics BufferMetrics) *DiskBuffer {
	// todo capacity
	walFile, err := wal.Open(path+"/"+name, nil)
	if err != nil {
		return nil // todo error handling
	}
	return &DiskBuffer{
		BufferMetrics: metrics,
		walFile:       walFile,
	}
}

func (b *DiskBuffer) Len() int {
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
	return index
}

func (b *DiskBuffer) Add(metrics ...telegraf.Metric) int {
	// one metric to write, can write directly
	if len(metrics) == 1 {
		if b.addSingle(metrics[0]) {
			return 1
		}
		return 0
	}

	// multiple metrics to write, batch them
	return b.addBatch(metrics)
}

func (b *DiskBuffer) addSingle(metric telegraf.Metric) bool {
	err := b.walFile.Write(b.writeIndex(), b.metricToBytes(metric))
	metric.Accept()
	if err == nil {
		b.metricAdded()
		return true
	}
	return false
}

func (b *DiskBuffer) addBatch(metrics []telegraf.Metric) int {
	written := 0
	batch := new(wal.Batch)
	for _, m := range metrics {
		data := b.metricToBytes(m)
		m.Accept() // accept here, since the metric object is no longer retained from here
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
	if b.Len() == 0 {
		// no metrics in the wal file, so return an empty array
		return make([]telegraf.Metric, 0)
	}
	metrics := make([]telegraf.Metric, batchSize)
	index := 0
	for i := b.readIndex(); i < b.readIndex()+uint64(batchSize); i++ {
		data, err := b.walFile.Read(i)
		if err != nil {
			// todo error handle
		}
		metrics[index] = b.bytesToMetric(data)
		index++
	}
	return metrics
}

func (b *DiskBuffer) Accept(batch []telegraf.Metric) {
	for _, m := range batch {
		b.metricWritten(m)
	}
	err := b.walFile.TruncateFront(b.readIndex() + uint64(len(batch)))
	if err != nil {
		panic(err) // can only occur with a corrupt wal file
	}
}

func (b *DiskBuffer) Reject(batch []telegraf.Metric) {
	for _, m := range batch {
		b.metricDropped(m)
	}
	err := b.walFile.TruncateFront(b.readIndex() + uint64(len(batch)))
	if err != nil {
		panic(err) // can only occur with a corrupt wal file
	}
}

func (b *DiskBuffer) metricToBytes(metric telegraf.Metric) []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(metric); err != nil {
		// todo error handle
		fmt.Println("Error encoding:", err)
		panic(1)
	}
	return buf.Bytes()
}

func (b *DiskBuffer) bytesToMetric(data []byte) telegraf.Metric {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	var m telegraf.Metric
	if err := decoder.Decode(&m); err != nil {
		// todo error handle
		fmt.Println("Error decoding:", err)
		panic(1)
	}
	return m
}
