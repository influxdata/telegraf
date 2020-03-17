package dedup

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  ## Maximum time to suppress output
  dedup_interval = "600s"
`

type Dedup struct {
	DedupInterval internal.Duration `toml:"dedup_interval"`
	FlushTime     time.Time
	Cache         map[uint64]telegraf.Metric
}

func (d *Dedup) SampleConfig() string {
	return sampleConfig
}

func (d *Dedup) Description() string {
	return "Deduplicate repetitive metrics"
}

// Remove single item from slice
func remove(slice []telegraf.Metric, i int) []telegraf.Metric {
	slice[len(slice)-1], slice[i] = slice[i], slice[len(slice)-1]
	return slice[:len(slice)-1]
}

// Remove expired items from cache
func (d *Dedup) cleanup() {
	// No need to cleanup cache too often. Lets save some CPU
	if time.Since(d.FlushTime) < d.DedupInterval.Duration {
		return
	}
	d.FlushTime = time.Now()
	keep := make(map[uint64]telegraf.Metric, 0)
	for id, metric := range d.Cache {
		if time.Since(metric.Time()) < d.DedupInterval.Duration {
			keep[id] = metric
		}
	}
	d.Cache = keep
}

// Save item to cache
func (d *Dedup) save(metric telegraf.Metric, id uint64) {
	d.Cache[id] = metric.Copy()
	d.Cache[id].Accept()
}

// main processing method
func (d *Dedup) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for idx, metric := range metrics {
		id := metric.HashID()
		m, ok := d.Cache[id]

		// If not in cache then just save it
		if !ok {
			d.save(metric, id)
			continue
		}

		// If cache item has expired then refresh it
		if time.Since(m.Time()) >= d.DedupInterval.Duration {
			d.save(metric, id)
			continue
		}

		// For each filed compare value with the cached one
		changed := false
		for _, f := range metric.FieldList() {
			if value, ok := m.GetField(f.Key); ok && value != f.Value {
				changed = true
				continue
			}
		}
		// If any field value has changed then refresh the cache
		if changed {
			d.save(metric, id)
			continue
		}

		// In any other case remove metric from the output
		metrics = remove(metrics, idx)
	}
	d.cleanup()
	return metrics
}

func init() {
	processors.Add("dedup", func() telegraf.Processor {
		return &Dedup{
			DedupInterval: internal.Duration{Duration: 10 * time.Minute},
			FlushTime:     time.Now(),
			Cache:         make(map[uint64]telegraf.Metric),
		}
	})
}
