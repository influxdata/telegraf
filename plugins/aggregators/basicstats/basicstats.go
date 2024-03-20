//go:generate ../../../tools/readme_config_includer/generator
package basicstats

import (
	_ "embed"
	"math"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

//go:embed sample.conf
var sampleConfig string

type BasicStats struct {
	Stats []string `toml:"stats"`
	Log   telegraf.Logger

	cache       map[uint64]aggregate
	statsConfig *configuredStats
}

type configuredStats struct {
	count           bool
	min             bool
	max             bool
	mean            bool
	variance        bool
	stdev           bool
	sum             bool
	diff            bool
	nonNegativeDiff bool
	rate            bool
	nonNegativeRate bool
	percentChange   bool
	interval        bool
	last            bool
}

func NewBasicStats() *BasicStats {
	return &BasicStats{
		cache: make(map[uint64]aggregate),
	}
}

type aggregate struct {
	fields map[string]basicstats
	name   string
	tags   map[string]string
}

type basicstats struct {
	count    float64
	min      float64
	max      float64
	sum      float64
	mean     float64
	diff     float64
	rate     float64
	interval time.Duration
	last     float64
	M2       float64   //intermediate value for variance/stdev
	PREVIOUS float64   //intermediate value for diff
	TIME     time.Time //intermediate value for rate
}

func (*BasicStats) SampleConfig() string {
	return sampleConfig
}

func (b *BasicStats) Add(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := b.cache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		a := aggregate{
			name:   in.Name(),
			tags:   in.Tags(),
			fields: make(map[string]basicstats),
		}
		for _, field := range in.FieldList() {
			if fv, ok := convert(field.Value); ok {
				a.fields[field.Key] = basicstats{
					count:    1,
					min:      fv,
					max:      fv,
					mean:     fv,
					sum:      fv,
					diff:     0.0,
					rate:     0.0,
					last:     fv,
					M2:       0.0,
					PREVIOUS: fv,
					TIME:     in.Time(),
				}
			}
		}
		b.cache[id] = a
	} else {
		for _, field := range in.FieldList() {
			if fv, ok := convert(field.Value); ok {
				if _, ok := b.cache[id].fields[field.Key]; !ok {
					// hit an uncached field of a cached metric
					b.cache[id].fields[field.Key] = basicstats{
						count:    1,
						min:      fv,
						max:      fv,
						mean:     fv,
						sum:      fv,
						diff:     0.0,
						rate:     0.0,
						interval: 0,
						last:     fv,
						M2:       0.0,
						PREVIOUS: fv,
						TIME:     in.Time(),
					}
					continue
				}

				tmp := b.cache[id].fields[field.Key]
				//https://en.m.wikipedia.org/wiki/Algorithms_for_calculating_variance
				//variable initialization
				x := fv
				mean := tmp.mean
				m2 := tmp.M2
				//counter compute
				n := tmp.count + 1
				tmp.count = n
				//mean compute
				delta := x - mean
				mean = mean + delta/n
				tmp.mean = mean
				//variance/stdev compute
				m2 = m2 + delta*(x-mean)
				tmp.M2 = m2
				//max/min compute
				if fv < tmp.min {
					tmp.min = fv
				} else if fv > tmp.max {
					tmp.max = fv
				}
				//sum compute
				tmp.sum += fv
				//diff compute
				tmp.diff = fv - tmp.PREVIOUS
				//interval compute
				tmp.interval = in.Time().Sub(tmp.TIME)
				//rate compute
				if !in.Time().Equal(tmp.TIME) {
					tmp.rate = tmp.diff / tmp.interval.Seconds()
				}
				//last compute
				tmp.last = fv
				//store final data
				b.cache[id].fields[field.Key] = tmp
			}
		}
	}
}

func (b *BasicStats) Push(acc telegraf.Accumulator) {
	for _, aggregate := range b.cache {
		fields := map[string]interface{}{}
		for k, v := range aggregate.fields {
			if b.statsConfig.count {
				fields[k+"_count"] = v.count
			}
			if b.statsConfig.min {
				fields[k+"_min"] = v.min
			}
			if b.statsConfig.max {
				fields[k+"_max"] = v.max
			}
			if b.statsConfig.mean {
				fields[k+"_mean"] = v.mean
			}
			if b.statsConfig.sum {
				fields[k+"_sum"] = v.sum
			}
			if b.statsConfig.last {
				fields[k+"_last"] = v.last
			}

			//v.count always >=1
			if v.count > 1 {
				variance := v.M2 / (v.count - 1)

				if b.statsConfig.variance {
					fields[k+"_s2"] = variance
				}
				if b.statsConfig.stdev {
					fields[k+"_stdev"] = math.Sqrt(variance)
				}
				if b.statsConfig.diff {
					fields[k+"_diff"] = v.diff
				}
				if b.statsConfig.nonNegativeDiff && v.diff >= 0 {
					fields[k+"_non_negative_diff"] = v.diff
				}
				if b.statsConfig.rate {
					fields[k+"_rate"] = v.rate
				}
				if b.statsConfig.percentChange {
					fields[k+"_percent_change"] = v.diff / v.PREVIOUS * 100
				}
				if b.statsConfig.nonNegativeRate && v.diff >= 0 {
					fields[k+"_non_negative_rate"] = v.rate
				}
				if b.statsConfig.interval {
					fields[k+"_interval"] = v.interval.Nanoseconds()
				}
			}
			//if count == 1 StdDev = infinite => so I won't send data
		}

		if len(fields) > 0 {
			acc.AddFields(aggregate.name, fields, aggregate.tags)
		}
	}
}

// member function for logging.
func (b *BasicStats) parseStats() *configuredStats {
	parsed := &configuredStats{}

	for _, name := range b.Stats {
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
		case "diff":
			parsed.diff = true
		case "non_negative_diff":
			parsed.nonNegativeDiff = true
		case "rate":
			parsed.rate = true
		case "non_negative_rate":
			parsed.nonNegativeRate = true
		case "percent_change":
			parsed.percentChange = true
		case "interval":
			parsed.interval = true
		case "last":
			parsed.last = true
		default:
			b.Log.Warnf("Unrecognized basic stat %q, ignoring", name)
		}
	}

	return parsed
}

func (b *BasicStats) getConfiguredStats() {
	if b.Stats == nil {
		b.statsConfig = &configuredStats{
			count:           true,
			min:             true,
			max:             true,
			mean:            true,
			variance:        true,
			stdev:           true,
			sum:             false,
			diff:            false,
			nonNegativeDiff: false,
			rate:            false,
			nonNegativeRate: false,
			percentChange:   false,
			interval:        false,
			last:            false,
		}
	} else {
		b.statsConfig = b.parseStats()
	}
}

func (b *BasicStats) Reset() {
	b.cache = make(map[uint64]aggregate)
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

func (b *BasicStats) Init() error {
	b.getConfiguredStats()

	return nil
}

func init() {
	aggregators.Add("basicstats", func() telegraf.Aggregator {
		return NewBasicStats()
	})
}
