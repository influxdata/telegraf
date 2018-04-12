package azuremetrics

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
	constants "github.com/influxdata/telegraf/plugins"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type AzureMetrics struct {
	Stats       []string `toml:"stats"`
	cache       map[uint64]aggregate
	statsConfig *configuredStats
	Period      string
}

type configuredStats struct {
	count              bool
	min                bool
	max                bool
	mean               bool
	totalSum           bool
	lastSample         bool
	beginningTimestamp bool
	endTimestamp       bool
}

func NewAzureMetrics() *AzureMetrics {
	mm := &AzureMetrics{}
	mm.Reset()
	return mm
}

type aggregate struct {
	fields map[string]azureMetrics
	name   string
	tags   map[string]string
}

type azureMetrics struct {
	count              float64
	min                float64
	max                float64
	mean               float64
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

func (m *AzureMetrics) SampleConfig() string {
	return sampleConfig
}

func (m *AzureMetrics) Description() string {
	return "Keep the aggregate metricAggregates of each metric passing through."
}

func (m *AzureMetrics) Add(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := m.cache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		a := aggregate{
			name:   in.Name(),
			tags:   in.Tags(),
			fields: make(map[string]azureMetrics),
		}
		for k, v := range in.Fields() {
			if fv, ok := convert(v); ok {
				a.fields[k] = azureMetrics{
					count:              1,
					min:                fv,
					max:                fv,
					mean:               fv,
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
					m.cache[id].fields[k] = azureMetrics{
						count:              1,
						min:                fv,
						max:                fv,
						mean:               fv,
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

func (m *AzureMetrics) Push(acc telegraf.Accumulator) {

	config := getConfiguredStats(m)

	for _, aggregate := range m.cache {
		fields := map[string]interface{}{}
		for k, v := range aggregate.fields {

			if config.count {
				fields[constants.SAMPLE_COUNT] = v.count
			}
			if config.min {
				fields[constants.MIN_SAMPLE] = v.min
			}
			if config.max {
				fields[constants.MAX_SAMPLE] = v.max
			}
			if config.mean {
				fields[constants.MEAN] = v.mean
			}
			if config.lastSample {
				fields[constants.LAST_SAMPLE] = v.lastSample
			}
			if config.beginningTimestamp {
				fields[constants.BEGIN_TIMESTAMP] = v.beginningTimestamp
			}
			if config.totalSum {
				fields[constants.TOTAL] = v.totalSum
			}
			if config.endTimestamp {
				fields[constants.END_TIMESTAMP] = v.endTimestamp
			}
			fields[constants.COUNTER_NAME] = k
		}

		if len(fields) > 0 {
			aggregate.tags[constants.PERIOD] = m.Period
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
	defaults.totalSum = true
	defaults.lastSample = true
	defaults.beginningTimestamp = true
	defaults.endTimestamp = true

	return defaults
}

func getConfiguredStats(m *AzureMetrics) *configuredStats {

	if m.statsConfig == nil {

		if m.Stats == nil {
			m.statsConfig = defaultStats()
		} else {
			m.statsConfig = parseStats(m.Stats)
		}
	}

	return m.statsConfig
}

func (m *AzureMetrics) Reset() {
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
	aggregators.Add("azuremetrics", func() telegraf.Aggregator {
		return NewAzureMetrics()
	})
}
