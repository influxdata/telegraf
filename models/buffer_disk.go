package models

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/aarthikrao/wal"
	"github.com/influxdata/telegraf"
	"go.uber.org/zap"
)

type DiskBuffer struct {
	BufferMetrics

	walFile *wal.WriteAheadLog
}

func NewDiskBuffer(name string, capacity int, path string, metrics BufferMetrics) *DiskBuffer {
	log, err := zap.NewProduction()
	if err != nil {
		return nil // todo error handling
	}
	walFile, err := wal.NewWriteAheadLog(&wal.WALOptions{
		LogDir:            path + "/" + name,
		MaxLogSize:        int64(capacity),
		MaxSegments:       2,
		Log:               log,
		MaxWaitBeforeSync: 1 * time.Second,
		SyncMaxBytes:      1000,
	})
	if err != nil {
		return nil // todo error handling
	}
	return &DiskBuffer{
		BufferMetrics: metrics,
		walFile:       walFile,
	}
}

func (b *DiskBuffer) Len() int {
	return -1 // todo
}

func (b *DiskBuffer) Add(metrics ...telegraf.Metric) int {
	written := 0
	for _, m := range metrics {
		data := b.metricToBytes(m)
		m.Accept() // accept here, since the metric object is no longer retained from here
		_, err := b.walFile.Write(data)
		if err == nil {
			written++
			b.metricAdded()
		}
	}
	return written
}

func (b *DiskBuffer) Batch(batchSize int) []telegraf.Metric {
	metrics := make([]telegraf.Metric, batchSize)
	index := 0
	_ = b.walFile.Replay(0, func(bytes []byte) error {
		if index == batchSize {
			return errors.New("stop parsing WAL file error")
		}
		metrics[index] = b.bytesToMetric(bytes)
		index++
		return nil
	})
	return metrics
}

func (b *DiskBuffer) Accept(batch []telegraf.Metric) {
	for _, m := range batch {
		b.metricWritten(m)
	}
}

func (b *DiskBuffer) Reject(batch []telegraf.Metric) {
	for _, m := range batch {
		b.metricDropped(m)
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
