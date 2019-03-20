package basicstats

import (
	"regexp"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs/statsd"
)

type BasicStats struct {
	Stats  []string            `toml:"stats"`
	Fields map[string][]string `toml:"fields"`
	Log    telegraf.Logger

	cache   map[uint64]aggregate
	configs map[string]configuredStats
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
	percentiles       []int
}

func NewBasicStats() *BasicStats {
	return &BasicStats{
		cache: make(map[uint64]aggregate),
	}
}

type aggregate struct {
	fields map[string]statsd.RunningStats
	name   string
	tags   map[string]string
}

var sampleConfig = `
  ## The period on which to flush & clear the aggregator.
  period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Configures which basic stats to push as fields. This option
  ## is deprecated and only kept for backward compatibility. If any
  ## fields is configured, this option will be ignored.
  # stats = ["count", "min", "max", "mean", "stdev", "s2", "sum"]

  ## Configures which basic stats to push as fields. "*" is the default configuration for all fields.
  ## Use strings like "p95" to add 95th percentile. Supported percentile range is [0, 100].
  # [aggregators.basicstats.fields]
  #   "*" = ["count", "min", "max", "mean", "stdev", "s2", "sum"]
  #   "some_field" = ["count", "p90", "p95"]
  ## If "*" is not provided, unmatched fields will be ignored.
  # [aggregators.basicstats.fields]
  #   "only_field" = ["count", "sum"]
`

func (*BasicStats) SampleConfig() string {
	return sampleConfig
}

func (*BasicStats) Description() string {
	return "Keep the aggregate statsd.RunningStats of each metric passing through."
}

func (b *BasicStats) Add(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := b.cache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		a := aggregate{
			name:   in.Name(),
			tags:   in.Tags(),
			fields: make(map[string]statsd.RunningStats),
		}
		for _, field := range in.FieldList() {
			if fv, ok := convert(field.Value); ok {
				rs := statsd.RunningStats{}
				rs.AddValue(fv)
				a.fields[field.Key] = rs
			}
		}
		b.cache[id] = a
	} else {
		for _, field := range in.FieldList() {
			if fv, ok := convert(field.Value); ok {
				if _, ok := b.cache[id].fields[field.Key]; !ok {
					// hit an uncached field of a cached metric
					b.cache[id].fields[field.Key] = statsd.RunningStats{}
				}

				rs := b.cache[id].fields[field.Key]
				rs.AddValue(fv)
				b.cache[id].fields[field.Key] = rs
			}
		}
	}
}

func (b *BasicStats) Push(acc telegraf.Accumulator) {
	for _, aggregate := range b.cache {
		fields := map[string]interface{}{}
		for k, v := range aggregate.fields {
			config := b.getConfiguredStatsForField(k)

			if config.count {
				fields[k+"_count"] = v.Count()
			}
			if config.min {
				fields[k+"_min"] = v.Lower()
			}
			if config.max {
				fields[k+"_max"] = v.Upper()
			}
			if config.mean {
				fields[k+"_mean"] = v.Mean()
			}
			if config.sum {
				fields[k+"_sum"] = v.Sum()
			}

			for _, p := range config.percentiles {
				fields[k+"_p"+strconv.Itoa(p)] = v.Percentile(float64(p))
			}

			// backward compatibility
			if v.Count() > 1 {
				if config.variance {
					fields[k+"_s2"] = v.Variance()
				}
				if config.stdev {
					fields[k+"_stdev"] = v.Stddev()
				}
				if config.diff {
					fields[k+"_diff"] = v.Diff()
				}
				if config.non_negative_diff && v.Diff() >= 0 {
					fields[k+"_non_negative_diff"] = v.Diff()
				}
			}
		}

		if len(fields) > 0 {
			acc.AddFields(aggregate.name, fields, aggregate.tags)
		}
	}
}

func (b *BasicStats) parseStats(stats []string) configuredStats {

	PRECENTILE_PATTERN := regexp.MustCompile(`^p([0-9]|[1-9][0-9]|100)$`)

	parsed := configuredStats{}

	for _, stat := range stats {

		// parse percentile stats, e.g. "p90" "p95"
		match := PRECENTILE_PATTERN.FindStringSubmatch(stat)
		if len(match) >= 2 {
			if p, err := strconv.Atoi(match[1]); err == nil {
				parsed.percentiles = append(parsed.percentiles, p)
				continue
			}
		}

		switch stat {

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
			b.Log.Warnf("Unrecognized basic stat %q, ignoring", stat)
		}
	}

	return parsed
}

func (b *BasicStats) getConfiguredStatsForField(field string) configuredStats {
	DEFAULT_FIELD := "*"
	DEFAULT_STATS := []string{"count", "min", "max", "mean", "s2", "stdev"}

	if b.configs == nil {

		if b.Fields == nil {
			b.Fields = make(map[string][]string)

			if b.Stats == nil {
				// neither b.Fileds nor b.Stats provided, use DEFAULT_STATS
				b.Fields[DEFAULT_FIELD] = DEFAULT_STATS
			} else {
				// make b.Stats default for all fields
				b.Fields[DEFAULT_FIELD] = b.Stats
			}
		}
		// b.Fields provided, b.Stats ignored

		b.configs = make(map[string]configuredStats)

		for k, stats := range b.Fields {
			b.configs[k] = b.parseStats(stats)
		}
	}

	if _, ok := b.configs[field]; !ok {
		// field-specfic stats not found, fallback to DEFAULT_FIELD
		field = DEFAULT_FIELD
	}
	// it's OK if DEFAULT_FIELD is not specified, the return below won't
	// result in any error or aggregated field, which is what we desired

	return b.configs[field]
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

func init() {
	aggregators.Add("basicstats", func() telegraf.Aggregator {
		return NewBasicStats()
	})
}
