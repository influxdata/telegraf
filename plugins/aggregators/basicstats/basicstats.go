package basicstats

import (
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type BasicStats struct {
	Stats []string `toml:"stats"`
	Log   telegraf.Logger

	cache       map[uint64]aggregate
	statsConfig *configuredStats
}

type configuredStats struct {
	count             bool
	min               bool
	max               bool
	mean              bool
	variance          bool
	stdev             bool
	sum               bool
	diff              bool
	non_negative_diff bool
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
	count float64
	min   float64
	max   float64
	sum   float64
	mean  float64
	diff  float64
	M2    float64 //intermediate value for variance/stdev
	LAST  float64 //intermediate value for diff
}

var sampleConfig = `
  ## The period on which to flush & clear the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Configures which basic stats to push as fields
  # stats = ["count", "min", "max", "mean", "stdev", "s2", "sum"]
`

func (*BasicStats) SampleConfig() string {
	return sampleConfig
}

func (*BasicStats) Description() string {
	return "Keep the aggregate basicstats of each metric passing through."
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
					count: 1,
					min:   fv,
					max:   fv,
					mean:  fv,
					sum:   fv,
					diff:  0.0,
					M2:    0.0,
					LAST:  fv,
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
						count: 1,
						min:   fv,
						max:   fv,
						mean:  fv,
						sum:   fv,
						diff:  0.0,
						M2:    0.0,
						LAST:  fv,
					}
					continue
				}

				tmp := b.cache[id].fields[field.Key]
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
				//sum compute
				tmp.sum += fv
				//diff compute
				tmp.diff = fv - tmp.LAST
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
				if b.statsConfig.non_negative_diff && v.diff >= 0 {
					fields[k+"_non_negative_diff"] = v.diff
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
			parsed.non_negative_diff = true

		default:
			b.Log.Warnf("Unrecognized basic stat %q, ignoring", name)
		}
	}

	return parsed
}

func (b *BasicStats) getConfiguredStats() {
	if b.Stats == nil {
		b.statsConfig = &configuredStats{
			count:             true,
			min:               true,
			max:               true,
			mean:              true,
			variance:          true,
			stdev:             true,
			sum:               false,
			non_negative_diff: false,
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
