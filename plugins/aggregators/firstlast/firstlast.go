package firstlast

// FirstLast.go

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"time"
)

type FirstLast struct {
	SeriesTimeout internal.Duration `toml:"series_timeout"`
	Warmup        internal.Duration `toml:"warmup"`
	EnableFirst   bool              `toml:"first"`
	EnableLast    bool              `toml:"last"`

	// Time when Telegraf is started
	startTime time.Time
	// The latest metric for all series which are active
	metricCache map[uint64]telegraf.Metric
	// First metrics of the series which have newly appeared in this interval
	firstMetricCache map[uint64]telegraf.Metric
}

func NewFirstLast() *FirstLast {
	return &FirstLast{
		startTime:        time.Now(),
		SeriesTimeout:    internal.Duration{Duration: 30 * time.Second},
		Warmup:           internal.Duration{Duration: 10 * time.Second},
		EnableFirst:      true,
		EnableLast:       true,
		metricCache:      make(map[uint64]telegraf.Metric),
		firstMetricCache: make(map[uint64]telegraf.Metric),
	}
}
func (m *FirstLast) Reset() {

}

var sampleConfig = `
[[aggregators.firstlast]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The time that a series is not updated until considering it ended
  series_timeout = "30s"
  ## The amount of time to wait after Telegraf startup until evaluating new series
  warmup = "10s"
  ## Emit first entry of a series
  first = true
  ## Emit last entry of a series
  last = true
`

func (m *FirstLast) SampleConfig() string {
	return sampleConfig
}

func (m *FirstLast) Description() string {
	return "Get the first and/or last values of a series"
}

func (m *FirstLast) Add(in telegraf.Metric) {
	id := in.HashID()
	// Check if this is a new series
	if _, ok := m.metricCache[id]; !ok {
		m.firstMetricCache[id] = in
	}
	m.metricCache[id] = in
}

func (m *FirstLast) Push(acc telegraf.Accumulator) {
	acc.SetPrecision(time.Nanosecond)

	// Check if there are any new series
	for id, metric := range m.firstMetricCache {
		if metric.Time().After(m.startTime.Add(m.Warmup.Duration)) && m.EnableFirst {
			fields := map[string]interface{}{}
			for k, v := range metric.Fields() {
				fields[k+"_first"] = v
			}
			acc.AddFields(metric.Name(), fields, metric.Tags(), metric.Time())
		}
		// We clear all the firstMetricCache entries at the end of the interval
		delete(m.firstMetricCache, id)
	}
	for id, metric := range m.metricCache {
		if time.Since(metric.Time()) > m.SeriesTimeout.Duration {
			if m.EnableLast {
				fields := map[string]interface{}{}
				for k, v := range metric.Fields() {
					fields[k+"_last"] = v
				}
				acc.AddFields(metric.Name(), fields, metric.Tags(), metric.Time())
			}
			// Only clear timed out entries
			delete(m.metricCache, id)
		}
	}
}

func init() {
	aggregators.Add("firstlast", func() telegraf.Aggregator {
		return NewFirstLast()
	})
}
