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
	return "Filter metrics with repeating field values"
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

		// For each field compare value with the cached one
		changed := false
		added := false
		sametime := metric.Time() == m.Time()
		for _, f := range metric.FieldList() {
			if value, ok := m.GetField(f.Key); ok {
				if value != f.Value {
					changed = true
					break
				}
			} else if sametime {
				// This field isn't in the cached metric but it's the
				// same series and timestamp. Merge it into the cached
				// metric.

				// Metrics have a ValueType that applies to all values
				// in the metric. If an input needs to produce values
				// with different ValueTypes but the same timestamp,
				// they have to produce multiple metrics. (See the
				// system input for an example.) In this case, dedup
				// ignores the ValueTypes of the metrics and merges
				// the fields into one metric for the dup check.

				m.AddField(f.Key, f.Value)
				added = true
			}
		}
		// If any field value has changed then refresh the cache
		if changed {
			d.save(metric, id)
			continue
		}

		if sametime && added {
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
