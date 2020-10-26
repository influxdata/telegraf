package stepped

import (
	"sort"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
	## Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

	## Unique Fields
	unique_fields = ["value"]

	## Step value offset
	step_offset = "1ns"

	## Maximum time to cache last value
	cache_interval = "720h"
`

type Stepped struct {
	Fields         []string          `toml:"unique_fields"`
	StepOffset     string            `toml:"step_offset"`
	RetainInterval internal.Duration `toml:"cache_interval"`
	FlushTime      time.Time
	Cache          map[uint64]telegraf.Metric
	Dur            time.Duration
}

func (d *Stepped) Init() error {
	dur, err := time.ParseDuration(d.StepOffset)

	if dur.Nanoseconds() > 0 {
		dur = dur * -1
	}
	if err != nil {
		return err
	}
	d.Dur = dur
	return nil
}

func (d *Stepped) SampleConfig() string {
	return sampleConfig
}

func (d *Stepped) Description() string {
	return "Insert a record for the previous unique field value and tag set just before the current one to display field as stepped."
}

// Remove single item from slice
func remove(slice []telegraf.Metric, i int) []telegraf.Metric {
	slice[len(slice)-1], slice[i] = slice[i], slice[len(slice)-1]
	return slice[:len(slice)-1]
}

// Remove expired items from cache
func (d *Stepped) cleanup() {
	// No need to cleanup cache too often. Lets save some CPU
	if time.Since(d.FlushTime) < d.RetainInterval.Duration {
		return
	}
	d.FlushTime = time.Now()
	keep := make(map[uint64]telegraf.Metric, 0)
	for id, metric := range d.Cache {
		if time.Since(metric.Time()) < d.RetainInterval.Duration {
			keep[id] = metric
		}
	}
	d.Cache = keep
}

// Save item to cache
func (d *Stepped) save(metric telegraf.Metric, id uint64) {
	d.Cache[id] = metric.Copy()
}

// Remove item from cache
func (d *Stepped) remove(id uint64) {
	d.Cache[id].Drop()
}

// Check if string in list of strings
func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

// Apply the main processing method
func (d *Stepped) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	steppedMetrics := []telegraf.Metric{}
	for _, metric := range metrics {
		id := metric.HashID()
		m, ok := d.Cache[id]

		// If not in cache then just save it
		if !ok {
			d.save(metric, id)
			continue
		}

		// If cache item has expired then remove it
		if time.Since(m.Time()) >= d.RetainInterval.Duration {
			d.remove(id)
			continue
		}

		// Otherwise lets refresh this cache value
		m.SetTime(metric.Time())

		// For each field compare value with the cached one
		changedFields := []string{}
		for _, f := range metric.FieldList() {
			found := contains(d.Fields, f.Key)
			if found {
				// Found check if value has changed
				if value, ok := m.GetField(f.Key); ok {
					if value != f.Value {
						// Record this has changed
						changedFields = append(changedFields, f.Key)
						break
					}
				} else {
					// This field isn't in the cached metric but it's the
					// same series and timestamp. Merge it into the cached
					// metric.
					m.AddField(f.Key, f.Value)
				}
			}
		}
		// If any field value has changed then refresh the cache
		if len(changedFields) > 0 {
			steppedMetric := m.Copy()

			for _, f := range m.FieldList() {
				found := contains(changedFields, f.Key)
				if !found {
					// Remove from stepped Metric
					steppedMetric.RemoveField(f.Key)
				} else {
					// Update Cache
					if value, ok := metric.GetField(f.Key); ok {
						m.AddField(f.Key, value)
					}
				}
			}
			steppedMetric.SetTime(steppedMetric.Time().Add(d.Dur))
			steppedMetric.Accept()
			steppedMetrics = append(steppedMetrics, steppedMetric)
		}

		d.save(m, id)
	}
	d.cleanup()
	return append(steppedMetrics, metrics...)
}

func init() {
	processors.Add("stepped", func() telegraf.Processor {
		return &Stepped{
			FlushTime: time.Now(),
			Cache:     make(map[uint64]telegraf.Metric),
		}
	})
}
