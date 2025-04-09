//go:generate ../../../tools/readme_config_includer/generator
package cumulative_sum

import (
	_ "embed"
	"fmt"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type CumulativeSum struct {
	Fields            []string        `toml:"fields"`
	DropOriginalField bool            `toml:"drop_original_field"`
	CleanUpInterval   config.Duration `toml:"clean_up_interval"`
	Log               telegraf.Logger `toml:"-"`

	fieldFilter filter.Filter
	cache       map[uint64]aggregate
	nextCleanUp time.Time
}

type aggregate struct {
	name       string
	tags       map[string]string
	fields     map[string]float64
	expireTime time.Time
}

var timeNow = time.Now

func NewCumulativeSum() *CumulativeSum {
	return &CumulativeSum{
		DropOriginalField: true,
		CleanUpInterval:   config.Duration(10 * time.Minute),
		cache:             make(map[uint64]aggregate),
	}
}

func (*CumulativeSum) SampleConfig() string {
	return sampleConfig
}

func (c *CumulativeSum) Init() error {
	c.nextCleanUp = timeNow().Add(time.Duration(c.CleanUpInterval))
	if c.Fields != nil {
		fieldFilter, err := filter.Compile(c.Fields)
		if err != nil {
			return fmt.Errorf("failed to create new field filter: %w", err)
		}
		c.fieldFilter = fieldFilter
	}
	return nil
}

func (c *CumulativeSum) Apply(in ...telegraf.Metric) []telegraf.Metric {
	c.cleanup()
	for _, original := range in {
		id := original.HashID()
		a, ok := c.cache[id]
		if !ok {
			a = aggregate{
				name:       original.Name(),
				tags:       original.Tags(),
				fields:     make(map[string]float64),
				expireTime: timeNow().Add(time.Duration(c.CleanUpInterval)),
			}
		}
		for _, field := range original.FieldList() {
			if c.fieldFilter != nil {
				if !c.fieldFilter.Match(field.Key) {
					continue
				}
			}
			fv, err := internal.ToFloat64(field.Value)
			if err == nil {
				if _, ok := a.fields[field.Key]; !ok {
					// hit an uncached field of a cached metric
					a.fields[field.Key] = fv
				} else {
					a.fields[field.Key] = a.fields[field.Key] + fv
				}
				original.AddField(field.Key+"_sum", a.fields[field.Key])
				if c.DropOriginalField {
					original.RemoveField(field.Key)
				}
				a.expireTime = timeNow().Add(time.Duration(c.CleanUpInterval))
			}
		}
		c.cache[id] = a
	}
	return in
}

// Remove expired items from cache
func (c *CumulativeSum) cleanup() {
	now := timeNow()
	if c.nextCleanUp.After(now) {
		return
	}
	c.nextCleanUp = now.Add(time.Duration(c.CleanUpInterval))
	keep := make(map[uint64]aggregate)
	for id, a := range c.cache {
		if a.expireTime.After(now) {
			keep[id] = a
		}
	}
	c.cache = keep
}

func init() {
	processors.Add("cumulative_sum", func() telegraf.Processor {
		return NewCumulativeSum()
	})
}
