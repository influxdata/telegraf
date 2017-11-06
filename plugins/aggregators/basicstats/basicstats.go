package basicstats

import (
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type BasicStats struct {
	cache map[uint64]aggregate
}

func NewBasicStats() telegraf.Aggregator {
	mm := &BasicStats{}
	mm.Reset()
	return mm
}

type aggregate struct {
	fields map[string]basicstats
	name   string
	tags   map[string]string
}

type basicstats struct {
	count float64
	min   float64
	max   float64
	mean  float64
	M2    float64 //intermedia value for variance/stdev
}

var sampleConfig = `
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
`

func (m *BasicStats) SampleConfig() string {
	return sampleConfig
}

func (m *BasicStats) Description() string {
	return "Keep the aggregate basicstats of each metric passing through."
}

func (m *BasicStats) Add(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := m.cache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		a := aggregate{
			name:   in.Name(),
			tags:   in.Tags(),
			fields: make(map[string]basicstats),
		}
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				a.fields[k] = basicstats{
					count: 1,
					min:   fv,
					max:   fv,
					mean:  fv,
					M2:    0.0,
				}
			}
		}
		m.cache[id] = a
	} else {
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				if _, ok := m.cache[id].fields[k]; !ok {
					// hit an uncached field of a cached metric
					m.cache[id].fields[k] = basicstats{
						count: 1,
						min:   fv,
						max:   fv,
						mean:  fv,
						M2:    0.0,
					}
					continue
				}

				tmp := m.cache[id].fields[k]
				//https://en.m.wikipedia.org/wiki/Algorithms_for_calculating_variance
				//variable initialization
				x := fv
				mean := tmp.mean
				M2 := tmp.M2
				//counter compute
				n := tmp.count + 1
				tmp.count = n
				//mean compute
				delta := x - mean
				mean = mean + delta/n
				tmp.mean = mean
				//variance/stdev compute
				M2 = M2 + delta*(x-mean)
				tmp.M2 = M2
				//max/min compute
				if fv < tmp.min {
					tmp.min = fv
				} else if fv > tmp.max {
					tmp.max = fv
				}
				//store final data
				m.cache[id].fields[k] = tmp
			}
		}
	}
}

func (m *BasicStats) Push(acc telegraf.Accumulator) {
	for _, aggregate := range m.cache {
		fields := map[string]interface{}{}
		for k, v := range aggregate.fields {
			fields[k+"_count"] = v.count
			fields[k+"_min"] = v.min
			fields[k+"_max"] = v.max
			fields[k+"_mean"] = v.mean
			//v.count always >=1
			if v.count > 1 {
				variance := v.M2 / (v.count - 1)
				fields[k+"_s2"] = variance
				fields[k+"_stdev"] = math.Sqrt(variance)
			}
			//if count == 1 StdDev = infinite => so I won't send data
		}
		acc.AddFields(aggregate.name, fields, aggregate.tags)
	}
}

func (m *BasicStats) Reset() {
	m.cache = make(map[uint64]aggregate)
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func init() {
	aggregators.Add("basicstats", func() telegraf.Aggregator {
		return NewBasicStats()
	})
}
