package batch

import (
	_ "embed"
	"strconv"
	"sync/atomic"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Batch struct {
	BatchTag     string `toml:"batch_tag"`
	NumBatches   uint64 `toml:"batches"`
	SkipExisting bool   `toml:"skip_existing"`

	// the number of metrics that have been processed so far
	count atomic.Uint64
}

func (*Batch) SampleConfig() string {
	return sampleConfig
}

func (b *Batch) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0, len(in))
	for _, m := range in {
		if b.SkipExisting && m.HasTag(b.BatchTag) {
			out = append(out, m)
			continue
		}

		oldCount := b.count.Add(1) - 1
		batchID := oldCount % b.NumBatches
		m.AddTag(b.BatchTag, strconv.FormatUint(batchID, 10))
		out = append(out, m)
	}

	return out
}

func init() {
	processors.Add("batch", func() telegraf.Processor {
		return &Batch{}
	})
}
