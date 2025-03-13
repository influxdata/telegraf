//go:generate ../../../tools/readme_config_includer/generator
package cumulative_sum

import (
	_ "embed"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type CumulativeSum struct {
	Log               telegraf.Logger
	Fields            []string        `toml:"fields"`
	DropOriginalField bool            `toml:"drop_original_field"`
	CleanUpInterval   config.Duration `toml:"clean_up_interval"`

	fieldMap    map[string]bool
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

func (c *CumulativeSum) Apply(in ...telegraf.Metric) []telegraf.Metric {
	c.cleanup()
	for _, original := range in {
		id := original.HashID()
		if _, ok := c.cache[id]; !ok {
			a := aggregate{
				name:       original.Name(),
				tags:       original.Tags(),
				fields:     make(map[string]float64),
				expireTime: timeNow().Add(time.Duration(c.CleanUpInterval)),
			}
			for _, field := range original.FieldList() {
				if c.fieldMap != nil {
					if _, ok := c.fieldMap[field.Key]; !ok {
						continue
					}
				}
				if fv, ok := convert(field.Value); ok {
					a.fields[field.Key] = fv
					original.AddField(field.Key+"_sum", fv)
					if c.DropOriginalField {
						original.RemoveField(field.Key)
					}
				}
			}
			c.cache[id] = a
		} else {
			for _, field := range original.FieldList() {
				if c.fieldMap != nil {
					if _, ok := c.fieldMap[field.Key]; !ok {
						continue
					}
				}
				if fv, ok := convert(field.Value); ok {
					a := c.cache[id]
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
					c.cache[id] = a
				}
			}
		}
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

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func (c *CumulativeSum) Init() error {
	c.nextCleanUp = timeNow().Add(time.Duration(c.CleanUpInterval))
	if c.Fields != nil {
		c.fieldMap = make(map[string]bool, len(c.Fields))
		for _, field := range c.Fields {
			c.fieldMap[field] = true
		}
	}
	return nil
}

func init() {
	processors.Add("cumulative_sum", func() telegraf.Processor {
		return NewCumulativeSum()
	})
}
