//go:generate ../../../tools/readme_config_includer/generator
package valuecounter

import (
	_ "embed"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

//go:embed sample.conf
var sampleConfig string

type ValueCounter struct {
	Fields []string `toml:"fields"`

	cache map[uint64]aggregate
}

type aggregate struct {
	name       string
	tags       map[string]string
	fieldCount map[string]int
}

func (*ValueCounter) SampleConfig() string {
	return sampleConfig
}

func (vc *ValueCounter) Add(in telegraf.Metric) {
	id := in.HashID()

	// Check if the cache already has an entry for this metric, if not create it
	if _, ok := vc.cache[id]; !ok {
		a := aggregate{
			name:       in.Name(),
			tags:       in.Tags(),
			fieldCount: make(map[string]int),
		}
		vc.cache[id] = a
	}

	// Check if this metric has fields which we need to count, if so increment
	// the count.
	for fk, fv := range in.Fields() {
		for _, cf := range vc.Fields {
			if fk == cf {
				fn := fmt.Sprintf("%v_%v", fk, fv)
				vc.cache[id].fieldCount[fn]++
			}
		}
	}
}

func (vc *ValueCounter) Push(acc telegraf.Accumulator) {
	for _, agg := range vc.cache {
		fields := make(map[string]interface{}, len(agg.fieldCount))
		for field, count := range agg.fieldCount {
			fields[field] = count
		}

		acc.AddFields(agg.name, fields, agg.tags)
	}
}

func (vc *ValueCounter) Reset() {
	vc.cache = make(map[uint64]aggregate)
}

func newValueCounter() telegraf.Aggregator {
	vc := &ValueCounter{}
	vc.Reset()
	return vc
}

func init() {
	aggregators.Add("valuecounter", func() telegraf.Aggregator {
		return newValueCounter()
	})
}
