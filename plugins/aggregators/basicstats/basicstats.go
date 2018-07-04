package basicstats

import (
	"log"
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type BasicStats struct {
	Stats []string `toml:"stats"`
	CoStatsConfig []costat `toml:"costat"`

	statsConfig *configuredStats
	cache       map[uint64]aggregate

	e *CoStats
}

type configuredStats struct {
	count    bool
	min      bool
	max      bool
	mean     bool
	variance bool
	stdev    bool
	sum      bool
}

func NewBasicStats() *BasicStats {
	mm := &BasicStats{}
	mm.Reset()
	return mm
}

type aggregate struct {
	name    string
	tags    map[string]string
	fields	map[string]basicstats
}

type basicstats struct {
	count float64
	min   float64
	max   float64
	sum   float64
	mean  float64
	M2    float64 //intermedia value for variance/stdev
	delta float64 //last value delta from the mean
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
					sum:   fv,
					M2:    0.0,
					delta: 0.0,
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
						sum:   fv,
						M2:    0.0,
						delta: 0.0,
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
				tmp.delta = delta
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
				//sum compute
				tmp.sum += fv
				//store final data
				m.cache[id].fields[k] = tmp

				// coStats products update
				(*(m.e)).Add(in, id, k)
			}
		}
	}
}


func (m *BasicStats) Push(acc telegraf.Accumulator) {

//	log.Printf("Push stats ");

	config := getConfiguredStats(m)

	for id, aggregate := range m.cache {
		fields := map[string]interface{}{}
		for k, v := range aggregate.fields {

			if config.count {
				fields[k+"_count"] = v.count
			}
			if config.min {
				fields[k+"_min"] = v.min
			}
			if config.max {
				fields[k+"_max"] = v.max
			}
			if config.mean {
				fields[k+"_mean"] = v.mean
			}
			if config.sum {
				fields[k+"_sum"] = v.sum
			}

			//v.count always >=1
			if v.count > 1 {
				variance := v.M2 / (v.count - 1)

				if config.variance {
					fields[k+"_s2"] = variance
				}
				if config.stdev {
					fields[k+"_stdev"] = math.Sqrt(variance)
				}

//				log.Printf("Push costats coCache %+v", fields);
				// co stats
				(*(m.e)).Push(id, k, fields)

			}
		}
		//if count == 1 StdDev = infinite => so I won't send data

		if len(fields) > 0 {
			acc.AddFields(aggregate.name, fields, aggregate.tags)
		}
	}

}

func parseStats(names []string) *configuredStats {

	parsed := &configuredStats{}

	for _, name := range names {

		switch name {

		case "count":
			parsed.count = true
		case "min":
			parsed.min = true
		case "max":
			parsed.max = true
		case "mean":
			parsed.mean = true
		case "s2":
			parsed.variance = true
		case "stdev":
			parsed.stdev = true
		case "sum":
			parsed.sum = true

		default:
			log.Printf("W! Unrecognized basic stat '%s', ignoring", name)
		}
	}

	return parsed
}

func defaultStats() *configuredStats {

	defaults := &configuredStats{}

	defaults.count = true
	defaults.min = true
	defaults.max = true
	defaults.mean = true
	defaults.variance = true
	defaults.stdev = true
	defaults.sum = false

	return defaults
}

func getConfiguredStats(m *BasicStats) *configuredStats {

	if m.statsConfig == nil {

		if m.Stats == nil {
			m.statsConfig = defaultStats()
		} else {
			m.statsConfig = parseStats(m.Stats)
		}
	}

	return m.statsConfig
}

func (m *BasicStats) Reset() {
	m.cache = make(map[uint64]aggregate)
	m.e = NewCoStats(m)
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
