package dedup

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  ## Maximum time to suppress output
  dedup_interval = "600s"
  ## Maximum time to keep cached metric without update
  evict_interval = "1h"
`

type Dedup struct {
	DedupInterval internal.Duration `toml:"dedup_interval"`
	EvictInterval internal.Duration `toml:"evict_interval"`
	Cache         map[uint64]telegraf.Metric
}

func NewDedup() *Dedup {
	return &Dedup{
		DedupInterval: internal.Duration{Duration: 10 * time.Minute},
		EvictInterval: internal.Duration{Duration: 1 * time.Hour},
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
	keep := make(map[uint64]telegraf.Metric, 0)
	for id, metric := range d.Cache {
		if time.Since(metric.Time()) < d.EvictInterval.Duration {
			keep[id] = metric
		}
	}
	d.Cache = keep
}

func (d *Dedup) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for idx, metric := range metrics {
		id := metric.HashID()
		if m, ok := d.Cache[id]; ok {
			// compare all fields values
			sameValue := true
			for k, v := range metric.Fields() {
				if m.Fields()[k] != v {
					log.Printf("D! [processors.dedup]: metric value has changed: %s", metric.Name())
					sameValue = false
				}
			}
			// value has not changed and cache is not expired
			if sameValue && time.Since(m.Time()) < d.DedupInterval.Duration {
				// Deduplicate this metric
				log.Printf("D! [processors.dedup]: suppress metric: %s", metric.Name())
				metrics = remove(metrics, idx)
				continue
			}
		}
		log.Printf("D! [processors.dedup]: update cached metric: %s", metric.Name())
		d.Cache[id] = metric.Copy()
	}
	d.cleanup()
	return metrics
}

func init() {
	processors.Add("dedup", func() telegraf.Processor {
		return NewDedup()
	})
}
