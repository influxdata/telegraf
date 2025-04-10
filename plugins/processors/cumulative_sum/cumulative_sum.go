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
	"github.com/influxdata/telegraf/metric"
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
	cache     map[uint64]telegraf.Metric
	nextReset time.Time
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

	c.cache = make(map[uint64]telegraf.Metric)

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
			a = metric.New(
				original.Name(),
				original.Tags(),
				map[string]interface{}{},
				time.Now(),
			)
		}
		for _, field := range original.FieldList() {
			if c.accept != nil {
				if !c.accept.Match(field.Key) {
					continue
				}
			}
			fv, err := internal.ToFloat64(field.Value)
			if err == nil {
				if v, found := a.GetField(field.Key); !found {
					a.AddField(field.Key, fv)
				} else {
					a.AddField(field.Key, v.(float64)+fv)
				}
				original.AddField(field.Key+"_sum", a.Fields()[field.Key])
				if !c.KeepOriginalField {
					original.RemoveField(field.Key)
				}
				a.SetTime(timeNow())
			}
		}
		c.cache[id] = a
	}
	return in
}

// Remove expired items from cache
func (c *CumulativeSum) cleanup() {
	// clean up not oftener than reset interval
	now := timeNow()
	if c.nextReset.After(now) {
		return
	}

	resetIntervalDuration := time.Duration(c.ResetInterval)

	// keep all fields that was updated not later than ResetInterval ago
	threshold := now.Add(-resetIntervalDuration)
	keep := make(map[uint64]telegraf.Metric)
	for id, a := range c.cache {
		if a.Time().After(threshold) {
			keep[id] = a
		}
	}
	c.cache = keep

	c.nextReset = now.Add(resetIntervalDuration)
}

func init() {
	processors.Add("cumulative_sum", func() telegraf.Processor {
		return &CumulativeSum{}
	})
}
