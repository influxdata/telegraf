//go:generate ../../../tools/readme_config_includer/generator
package dedup

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/processors"
	influxSerializer "github.com/influxdata/telegraf/plugins/serializers/influx"
)

//go:embed sample.conf
var sampleConfig string

type Dedup struct {
	DedupInterval config.Duration `toml:"dedup_interval"`
	FlushTime     time.Time
	Cache         map[uint64]telegraf.Metric
}

// Remove expired items from cache
func (d *Dedup) cleanup() {
	// No need to cleanup cache too often. Lets save some CPU
	if time.Since(d.FlushTime) < time.Duration(d.DedupInterval) {
		return
	}
	d.FlushTime = time.Now()
	keep := make(map[uint64]telegraf.Metric)
	for id, metric := range d.Cache {
		if time.Since(metric.Time()) < time.Duration(d.DedupInterval) {
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

func (*Dedup) SampleConfig() string {
	return sampleConfig
}

// main processing method
func (d *Dedup) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	idx := 0
	for _, metric := range metrics {
		id := metric.HashID()
		m, ok := d.Cache[id]

		// If not in cache then just save it
		if !ok {
			d.save(metric, id)
			metrics[idx] = metric
			idx++
			continue
		}

		// If cache item has expired then refresh it
		if time.Since(m.Time()) >= time.Duration(d.DedupInterval) {
			d.save(metric, id)
			metrics[idx] = metric
			idx++
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
			metrics[idx] = metric
			idx++
			continue
		}

		if sametime && added {
			metrics[idx] = metric
			idx++
			continue
		}

		// In any other case remove metric from the output
		metric.Drop()
	}
	metrics = metrics[:idx]
	d.cleanup()
	return metrics
}

func (d *Dedup) GetState() interface{} {
	s := &influxSerializer.Serializer{}
	v := make([]telegraf.Metric, 0, len(d.Cache))
	for _, value := range d.Cache {
		v = append(v, value)
	}
	state, _ := s.SerializeBatch(v)
	return state
}

func (d *Dedup) SetState(state interface{}) error {
	p := &influx.Parser{}
	if err := p.Init(); err != nil {
		return err
	}
	data, ok := state.([]byte)
	if !ok {
		return fmt.Errorf("state has wrong type %T", state)
	}
	metrics, err := p.Parse(data)
	if err == nil {
		d.Apply(metrics...)
	}
	return nil
}

func init() {
	processors.Add("dedup", func() telegraf.Processor {
		return &Dedup{
			DedupInterval: config.Duration(10 * time.Minute),
			FlushTime:     time.Now(),
			Cache:         make(map[uint64]telegraf.Metric),
		}
	})
}
