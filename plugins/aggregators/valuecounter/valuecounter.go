package valuecounter

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type aggregate struct {
	name       string
	tags       map[string]string
	fieldCount map[string]int
}

// ValueCounter an aggregation plugin
type ValueCounter struct {
	cache  map[uint64]aggregate
	Fields []string
}

// NewValueCounter create a new aggregation plugin which counts the occurrences
// of fields and emits the count.
func NewValueCounter() telegraf.Aggregator {
	vc := &ValueCounter{}
	vc.Reset()
	return vc
}

// Add is run on every metric which passes the plugin
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

// Push emits the counters
func (vc *ValueCounter) Push(acc telegraf.Accumulator) {
	for _, agg := range vc.cache {
		fields := map[string]interface{}{}

		for field, count := range agg.fieldCount {
			fields[field] = count
		}

		acc.AddFields(agg.name, fields, agg.tags)
	}
}

// Reset the cache, executed after each push
func (vc *ValueCounter) Reset() {
	vc.cache = make(map[uint64]aggregate)
}

func init() {
	aggregators.Add("valuecounter", func() telegraf.Aggregator {
		return NewValueCounter()
	})
}
