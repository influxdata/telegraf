package statsd

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	defaultFieldName = "value"
	defaultTagKey    = "statsd_type"
)

type Statsd struct {
	// cache added metrics, reset every period
	// gauges, counters: hash of measurement & tags --> metric
	// sets: hash of measurement & tags --> metric value sets
	// timing: hash of measurement & tags --> RunningStats
	gauges   map[uint64]cachedGauge
	counters map[uint64]cachedCounter
	sets     map[uint64]cachedSet
	timings  map[uint64]cachedTimings

	// Percentiles specifies the percentiles that will be calculated for timing
	// and histogram stats.
	PercentileLimit int   `toml:"percentile_limit"`
	Percentiles     []int `toml:"percentiles"`

	DeleteTimings  bool `toml:"delete_timings"`
	DeleteCounters bool `toml:"delete_counters"`
	DeleteGauges   bool `toml:"delete_gauges"`
	DeleteSets     bool `toml:"delete_sets"`
}

type cachedGauge struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

type cachedCounter struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

type cachedSet struct {
	name   string
	fields map[string]map[string]bool
	tags   map[string]string
}

type cachedTimings struct {
	name   string
	fields map[string]*RunningStats
	tags   map[string]string
}

func NewStatsd() *Statsd {
	s := &Statsd{
		DeleteGauges:   true,
		DeleteCounters: true,
		DeleteSets:     true,
		DeleteTimings:  true,
	}
	s.Reset()
	return s
}

var sampleConfig = `
[[aggregators.statsd]]
  # General Aggregator Arguments:

  ## The period on which to flush & clear the aggregator.
  period = "5s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = true

  # Statsd Arguments:

  ## The following configuration options control when aggregator clears it's
  ## cache of previous values. If set to false, then telegraf will only clear
  ## it's cache when the daemon is restarted.
  ## Reset gauges every interval (default=true)
  delete_gauges = true
  ## Reset counters every interval (default=true)
  delete_counters = true
  ## Reset sets every interval (default=true)
  delete_sets = true
  ## Reset timings & histograms every interval (default=true)
  delete_timings = true

  ## Percentiles to calculate for timing & histogram stats
  percentiles = [90]

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000
`

func (s *Statsd) SampleConfig() string {
	return sampleConfig
}

func (s *Statsd) Description() string {
	return "Aggregate metrics from statsd."
}

func (s *Statsd) Add(in telegraf.Metric) {
	mtype, _ := in.Tags()[defaultTagKey]

	hashID := in.HashID()
	switch mtype {
	case "ms", "h":
		_, ok := s.timings[hashID]
		if !ok {
			s.timings[hashID] = cachedTimings{
				name:   in.Name(),
				tags:   in.Tags(),
				fields: make(map[string]*RunningStats),
			}
		}
		for field, value := range in.Fields() {
			_, ok := s.timings[hashID].fields[field]
			if !ok {
				s.timings[hashID].fields[field] = &RunningStats{
					PercLimit: s.PercentileLimit,
				}
			}
			s.timings[hashID].fields[field].AddValue(value.(float64))
		}
	case "c":
		_, ok := s.counters[hashID]
		if !ok {
			s.counters[hashID] = cachedCounter{
				name:   in.Name(),
				tags:   in.Tags(),
				fields: make(map[string]interface{}),
			}
		}
		for field, value := range in.Fields() {
			_, ok := s.counters[hashID].fields[field]
			if !ok {
				s.counters[hashID].fields[field] = int64(0)
			}
			s.counters[hashID].fields[field] = s.counters[hashID].fields[field].(int64) + value.(int64)
		}
	case "g":
		_, ok := s.gauges[hashID]
		if !ok {
			s.gauges[hashID] = cachedGauge{
				name:   in.Name(),
				tags:   in.Tags(),
				fields: make(map[string]interface{}),
			}
		}
		for field, value := range in.Fields() {
			_, ok = s.gauges[hashID].fields[field]
			if !ok {
				s.gauges[hashID].fields[field] = float64(0)
			}
			valueStr := value.(string)
			valueFloat, err := strconv.ParseFloat(valueStr, 10)
			if err != nil {
				log.Printf("!E Gauge value is not float %v", err)
				continue
			}
			if strings.HasPrefix(valueStr, "+") || strings.HasPrefix(valueStr, "-") {
				s.gauges[hashID].fields[field] = s.gauges[hashID].fields[field].(float64) + valueFloat
			} else {
				s.gauges[hashID].fields[field] = valueFloat
			}
		}
	case "s":
		_, ok := s.sets[hashID]
		if !ok {
			s.sets[hashID] = cachedSet{
				name:   in.Name(),
				tags:   in.Tags(),
				fields: make(map[string]map[string]bool),
			}
		}
		for field, value := range in.Fields() {
			_, ok = s.sets[hashID].fields[field]
			if !ok {
				s.sets[hashID].fields[field] = make(map[string]bool)
			}
			s.sets[hashID].fields[field][value.(string)] = true
		}
	}
	return
}

func (s *Statsd) Push(acc telegraf.Accumulator) {
	now := time.Now()

	for _, metric := range s.timings {
		fields := make(map[string]interface{})
		for fieldName, stats := range metric.fields {
			var prefix string
			if fieldName != defaultFieldName {
				prefix = fieldName + "_"
			}

			fields[prefix+"mean"] = stats.Mean()
			fields[prefix+"stddev"] = stats.Stddev()
			fields[prefix+"sum"] = stats.Sum()
			fields[prefix+"upper"] = stats.Upper()
			fields[prefix+"lower"] = stats.Lower()
			fields[prefix+"count"] = stats.Count()
			for _, percentile := range s.Percentiles {
				fields[fmt.Sprintf("%s%v_percentile", prefix, percentile)] = stats.Percentile(percentile)
			}
		}
		acc.AddFields(metric.name, fields, metric.tags, now)
	}

	for _, metric := range s.counters {
		acc.AddCounter(metric.name, metric.fields, metric.tags, now)
	}

	for _, metric := range s.gauges {
		acc.AddGauge(metric.name, metric.fields, metric.tags, now)
	}

	for _, metric := range s.sets {
		fields := make(map[string]interface{})
		for field, set := range metric.fields {
			fields[field] = int64(len(set))
		}
		acc.AddFields(metric.name, fields, metric.tags, now)
	}
}

func (s *Statsd) Reset() {
	if s.DeleteGauges {
		s.gauges = make(map[uint64]cachedGauge)
	}
	if s.DeleteCounters {
		s.counters = make(map[uint64]cachedCounter)
	}
	if s.DeleteSets {
		s.sets = make(map[uint64]cachedSet)
	}
	if s.DeleteTimings {
		s.timings = make(map[uint64]cachedTimings)
	}
}

func init() {
	aggregators.Add("statsd", func() telegraf.Aggregator {
		return NewStatsd()
	})
}
