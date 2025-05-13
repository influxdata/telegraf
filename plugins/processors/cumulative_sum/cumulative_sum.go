//go:generate ../../../tools/readme_config_includer/generator
package cumulative_sum

import (
	_ "embed"
	"fmt"
	"maps"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type CumulativeSum struct {
	Fields         []string        `toml:"fields"`
	ExpiryInterval config.Duration `toml:"expiry_interval"`
	Log            telegraf.Logger `toml:"-"`

	accept filter.Filter
	cache  map[uint64]*entry
}

type entry struct {
	sums map[string]float64
	seen time.Time
}

func (*CumulativeSum) SampleConfig() string {
	return sampleConfig
}

func (c *CumulativeSum) Init() error {
	if len(c.Fields) == 0 {
		c.Fields = []string{"*"}
	}
	f, err := filter.Compile(c.Fields)
	if err != nil {
		return fmt.Errorf("failed to create new field filter: %w", err)
	}
	c.accept = f

	c.cache = make(map[uint64]*entry)

	return nil
}

func (c *CumulativeSum) Apply(in ...telegraf.Metric) []telegraf.Metric {
	now := time.Now()

	out := make([]telegraf.Metric, 0, len(in))
	for _, original := range in {
		id := original.HashID()
		// Create a new entry for unseen metrics
		stored, ok := c.cache[id]
		if !ok {
			stored = &entry{sums: make(map[string]float64)}
		}
		// Create a metric with the summed fields
		m := original.Copy()
		for _, field := range m.FieldList() {
			// Ignore all non-sum fields and keep them
			if c.accept != nil && !c.accept.Match(field.Key) {
				continue
			}

			// Ignore all fields not convertible to float
			fv, err := internal.ToFloat64(field.Value)
			if err != nil {
				c.Log.Tracef("Skipping field %q with value %v (%T) as it is not convertible to float: %v", field.Key, field.Value, field.Value, err)
				continue
			}

			// Compute the sum and create the new field
			sum := stored.sums[field.Key] + fv
			m.AddField(field.Key+"_sum", sum)
			stored.sums[field.Key] = sum
		}
		stored.seen = now
		c.cache[id] = stored

		out = append(out, m)
		original.Accept()
	}

	// Cleanup cache entries that are too old
	if c.ExpiryInterval > 0 {
		threshold := now.Add(-time.Duration(c.ExpiryInterval))
		maps.DeleteFunc(c.cache, func(_ uint64, e *entry) bool {
			return e.seen.Before(threshold)
		})
	}

	return out
}

func init() {
	processors.Add("cumulative_sum", func() telegraf.Processor {
		return &CumulativeSum{}
	})
}
