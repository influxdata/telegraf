package models

import "github.com/influxdata/telegraf"

type discardBuffer struct {
	BufferStats
}

func newDiscardBuffer(stats BufferStats) *discardBuffer {
	return &discardBuffer{BufferStats: stats}
}

func (*discardBuffer) Len() int {
	return 0
}

func (b *discardBuffer) Add(metrics ...telegraf.Metric) int {
	for _, m := range metrics {
		b.metricDropped(m)
	}
	return len(metrics)
}

func (*discardBuffer) BeginTransaction(int) *Transaction {
	return &Transaction{}
}

func (*discardBuffer) EndTransaction(*Transaction) {
}

func (b *discardBuffer) Stats() BufferStats {
	return b.BufferStats
}

func (*discardBuffer) Close() error {
	return nil
}
