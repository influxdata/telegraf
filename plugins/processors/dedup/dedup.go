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

func NewDedup() *Dedup {
	return &Dedup{
		DedupInterval: internal.Duration{Duration: 10 * time.Minute},
		FlushTime:     time.Now(),
		Cache:         make(map[uint64]telegraf.Metric),
	}
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
func (d *Dedup) save(metric telegraf.Metric) {
	id := metric.HashID()
	d.Cache[id] = metric.Copy()
	d.Cache[id].Accept()
}

// main processing method
func (d *Dedup) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for idx, metric := range metrics {
		id := metric.HashID()
		// check if metric is already in cache. Otherwise save it in cache
		if m, ok := d.Cache[id]; ok {
			// if cache is not expired then check values. Otherwise save it in cache
			if time.Since(m.Time()) < d.DedupInterval.Duration {
				for _, field := range metric.FieldList() {
					// if same value then drop it. Otherwise save it in cache
					if m.Fields()[field.Key] == field.Value {
						metrics = remove(metrics, idx)
						continue
					} else {
						d.save(metric)
					}
				}
			} else {
				d.save(metric)
			}
		} else {
			d.save(metric)
		}
	}
	d.cleanup()
	return metrics
}

func init() {
	processors.Add("dedup", func() telegraf.Processor {
		return NewDedup()
	})
}
