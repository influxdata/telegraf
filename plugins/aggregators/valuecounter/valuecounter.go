package valuecounter

import (
	"fmt"
	"github.com/influxdata/telegraf/internal/choice"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type aggregate struct {
	name       string
	tags       map[string]string
	fieldCount map[string]int
}

type predicate func(int, int) bool

type Predicate struct {
	Type          string `toml:"type"`
	Value         int    `toml:"value"`
	predicateFunc predicate
}

// ValueCounter an aggregation plugin
type ValueCounter struct {
	cache      map[uint64]aggregate
	Fields     []string
	Predicates []*Predicate `toml:"predicate"`
}

// NewValueCounter create a new aggregation plugin which counts the occurrences
// of fields and emits the count.
func NewValueCounter() telegraf.Aggregator {
	vc := &ValueCounter{}
	vc.Reset()
	return vc
}

var sampleConfig = `
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The fields for which the values will be counted
  fields = []
  ## Only emit fields whose aggregated value matches the predicate
  [[aggregators.valuecounter.predicate]]
	type = "greater_than"
	value = 0
 
`

// SampleConfig generates a sample config for the ValueCounter plugin
func (vc *ValueCounter) SampleConfig() string {
	return sampleConfig
}

// Description returns the description of the ValueCounter plugin
func (vc *ValueCounter) Description() string {
	return "Count the occurrence of values in fields."
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
			var matched = 0
			for _, predicate := range vc.Predicates {
				if predicate.predicateFunc(count, predicate.Value) {
					matched++
				}
			}

			if matched == len(vc.Predicates) {
				fields[field] = count
			}
		}

		acc.AddFields(agg.name, fields, agg.tags)
	}
}

// Reset the cache, executed after each push
func (vc *ValueCounter) Reset() {
	vc.cache = make(map[uint64]aggregate)
}

func knownPredicateFunctions() (map[string]predicate, []string) {
	predicates := map[string]predicate{
		"greater_than": func(count int, compareAgainst int) bool { return count > compareAgainst },
		"less_than":    func(count int, compareAgainst int) bool { return count < compareAgainst },
		"equal_to":     func(count int, compareAgainst int) bool { return count == compareAgainst },
		"not_equal_to": func(count int, compareAgainst int) bool { return count != compareAgainst },
	}

	keys := make([]string, 0, len(predicates))
	for k := range predicates {
		keys = append(keys, k)
	}

	return predicates, keys
}

func (vc *ValueCounter) configurePredicates() error {

	predicateFunctions, knownPredicates := knownPredicateFunctions()
	var requestedPredicateTypes []string
	for _, t := range vc.Predicates {
		requestedPredicateTypes = append(requestedPredicateTypes, t.Type)
	}
	err := choice.CheckSlice(requestedPredicateTypes, knownPredicates)
	if err != nil {
		return fmt.Errorf(`cannot verify "predicate" settings: %v`, err)
	}

	for _, predicate := range vc.Predicates {
		if predicateFunction, ok := predicateFunctions[predicate.Type]; ok {
			predicate.predicateFunc = predicateFunction
		}
	}
	return nil
}

func (vc *ValueCounter) Init() error {
	return vc.configurePredicates()
}

func init() {
	aggregators.Add("valuecounter", func() telegraf.Aggregator {
		return NewValueCounter()
	})
}
