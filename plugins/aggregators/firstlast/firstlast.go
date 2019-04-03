package firstlast

// FirstLast.go

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"time"
)

type FirstLast struct {
	LastTimeout internal.Duration `toml:"timeout"`
	WarmupTime  internal.Duration `toml:"warmup"`
	EnableFirst bool              `toml:"first"`
	EnableLast  bool              `toml:"last"`
	FirstSuffix string            `toml:"first_suffix"`
	LastSuffix  string            `toml:"last_suffix"`

	// caches for metric fields, names, and tags
	startTime        time.Time
	metricCache      map[uint64]telegraf.Metric
	firstMetricCache map[uint64]telegraf.Metric
}

func NewFirstLast() telegraf.Aggregator {
	return &FirstLast{
		startTime:        time.Now(),
		LastTimeout:      internal.Duration{Duration: 20 * time.Second},
		WarmupTime:       internal.Duration{Duration: 10 * time.Second},
		EnableFirst:      true,
		EnableLast:       true,
		FirstSuffix:      "_first",
		LastSuffix:       "_last",
		metricCache:      make(map[uint64]telegraf.Metric),
		firstMetricCache: make(map[uint64]telegraf.Metric),
	}
}
func (m *FirstLast) Reset() {

}

var sampleConfig = `
  ## period is the flush & clear interval of the aggregator.
  period = "30s"
  ## If true drop_original will drop the original metrics and
  ## only send aggregates.
  drop_original = false
  ## The amount of time until a series is considered ended
  timeout = "30s"
  ## The amount of time before we start issuing _first 
  warmup = "10s"
  ## Emit first entry of a series
  first = true
  ## Suffix for the measurement names of first entries
  first_suffix = "_first"
  ## Emit last entry of a series
  last = true
  ## Suffix for the measurement names of last entries
  last_suffix = "_last"

 `

func (m *FirstLast) SampleConfig() string {
	return sampleConfig
}

func (m *FirstLast) Description() string {
	return "Get the first and/or last values of a series"
}

func (m *FirstLast) Add(in telegraf.Metric) {
	id := in.HashID()
	if _, ok := m.metricCache[id]; !ok {
		m.firstMetricCache[id] = in
	}
	m.metricCache[id] = in
}

func (m *FirstLast) Push(acc telegraf.Accumulator) {
	acc.SetPrecision(time.Nanosecond)
	for id, metric := range m.firstMetricCache {
		if time.Since(metric.Time()) > m.WarmupTime.Duration && m.EnableFirst {
			acc.AddFields(metric.Name()+m.FirstSuffix, metric.Fields(), metric.Tags(), metric.Time())
		}
		delete(m.firstMetricCache, id)
	}
	for id, metric := range m.metricCache {
		if time.Since(metric.Time()) > m.LastTimeout.Duration {
			if m.EnableLast {
				acc.AddFields(metric.Name()+m.LastSuffix, metric.Fields(), metric.Tags(), metric.Time())
			}
			delete(m.metricCache, id)
		}
	}
}

func init() {
	aggregators.Add("firstlast", func() telegraf.Aggregator {
		return NewFirstLast()
	})
}
