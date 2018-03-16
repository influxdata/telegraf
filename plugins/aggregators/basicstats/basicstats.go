package basicstats

import (
	"log"
	//"math"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type BasicStats struct {
	Stats       []string `toml:"stats"`
	cache       map[uint64]aggregate
	statsConfig *configuredStats
	Period      string
}

type configuredStats struct {
	count bool
	min   bool
	max   bool
	mean  bool
	//	variance           bool
	//	stdev              bool
	totalSum           bool
	lastSample         bool
	beginningTimestamp bool
	endTimestamp       bool
}

func NewBasicStats() *BasicStats {
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
	//	M2    float64 //intermedia value for variance/stdev
	totalSum           float64
	lastSample         float64
	beginningTimestamp string
	endTimestamp       string
}

const layout = "02/01/2006 03:04:05 PM"

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
					//M2:    0.0,
					totalSum:           fv,
					lastSample:         fv,
					beginningTimestamp: time.Now().Format(layout),
					endTimestamp:       time.Now().Format(layout),
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
						//	M2:    0.0,
						totalSum:           fv,
						lastSample:         fv,
						beginningTimestamp: time.Now().Format(layout),
						endTimestamp:       time.Now().Format(layout),
					}
					continue
				}

				tmp := m.cache[id].fields[k]
				//https://en.m.wikipedia.org/wiki/Algorithms_for_calculating_variance
				//variable initialization
				x := fv
				mean := tmp.mean
				//counter compute
				n := tmp.count + 1
				tmp.count = n
				//mean compute
				delta := x - mean
				mean = mean + delta/n
				tmp.mean = mean
				//variance/stdev compute
				//	M2 = M2 + delta*(x-mean)
				//	tmp.M2 = M2
				//max/min compute
				if fv < tmp.min {
					tmp.min = fv
				} else if fv > tmp.max {
					tmp.max = fv
				}
				tmp.totalSum = tmp.totalSum + fv
				tmp.lastSample = fv
				tmp.endTimestamp = time.Now().Format(layout)
				//store final data
				m.cache[id].fields[k] = tmp
			}
		}
	}
}

func (m *BasicStats) Push(acc telegraf.Accumulator) {

	config := getConfiguredStats(m)

	for _, aggregate := range m.cache {
		fields := map[string]interface{}{}
		for k, v := range aggregate.fields {

			if config.count {
				fields["Count"] = v.count
			}
			if config.min {
				fields["Minimum"] = v.min
			}
			if config.max {
				fields["Maximum"] = v.max
			}
			if config.mean {
				fields["Average"] = v.mean
			}

			//v.count always >=1
			//	if v.count > 1 {
			//		variance := v.M2 / (v.count - 1)

			//		if config.variance {
			//			fields[k+"_s2"] = variance
			//		}
			//		if config.stdev {
			//			fields[k+"_stdev"] = math.Sqrt(variance)
			//		}
			//	}
			if config.lastSample {
				fields["Last"] = v.lastSample
			}
			if config.beginningTimestamp {
				fields["TIMESTAMP"] = v.beginningTimestamp
			}
			if config.totalSum {
				fields["Total"] = v.totalSum
			}
			if config.endTimestamp {
				fields["Timestamp"] = v.endTimestamp
			}
			fields["CounterName"] = k
			//if count == 1 StdDev = infinite => so I won't send data
		}

		if len(fields) > 0 {
			aggregate.tags["Period"] = m.Period
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
			//	case "s2":
			//		parsed.variance = true
			//	case "stdev":
			//		parsed.stdev = true
		case "totalSum":
			parsed.totalSum = true
		case "lastSample":
			parsed.lastSample = true
		case "beginningTimestamp":
			parsed.beginningTimestamp = true
		case "endTimestamp":
			parsed.endTimestamp = true

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
	//	defaults.variance = true
	//	defaults.stdev = true
	defaults.totalSum = true
	defaults.lastSample = true
	defaults.beginningTimestamp = true
	defaults.endTimestamp = true

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
