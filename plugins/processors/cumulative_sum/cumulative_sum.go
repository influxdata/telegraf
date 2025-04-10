//go:generate ../../../tools/readme_config_includer/generator
package cumulative_sum

import (
	_ "embed"
	"fmt"
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
	Fields            []string        `toml:"fields"`
	KeepOriginalField bool            `toml:"keep_original_field"`
	ResetInterval     config.Duration `toml:"reset_interval"`
	Log               telegraf.Logger `toml:"-"`

	accept    filter.Filter
	cache     map[uint64]aggregate
	nextReset time.Time
}

type aggregate struct {
	name       string
	tags       map[string]string
	fields     map[string]float64
	expireTime time.Time
}

var timeNow = time.Now

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

	c.cache = make(map[uint64]aggregate)

	if c.ResetInterval > 0 {
		c.nextReset = timeNow().Add(time.Duration(c.ResetInterval))
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
				expireTime: timeNow().Add(time.Duration(c.ResetInterval)),
			}
		}
		for _, field := range original.FieldList() {
			if c.accept != nil {
				if !c.accept.Match(field.Key) {
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
				if !c.KeepOriginalField {
					original.RemoveField(field.Key)
				}
				a.expireTime = timeNow().Add(time.Duration(c.ResetInterval))
			}
		}
		c.cache[id] = a
	}
	return in
}

// Remove expired items from cache
func (c *CumulativeSum) cleanup() {
	now := timeNow()
	if c.nextReset.After(now) {
		return
	}
	c.nextReset = now.Add(time.Duration(c.ResetInterval))
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
		return &CumulativeSum{}
	})
}
