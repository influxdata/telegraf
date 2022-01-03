package final

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

var sampleConfig = `
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## The time that a series is not updated until considering it final.
  series_timeout = "5m"
`

type Final struct {
	SeriesTimeout config.Duration `toml:"series_timeout"`

	// The last metric for all series which are active
	metricCache map[uint64]telegraf.Metric
}

func NewFinal() *Final {
	return &Final{
		SeriesTimeout: config.Duration(5 * time.Minute),
		metricCache:   make(map[uint64]telegraf.Metric),
	}
}

func (m *Final) SampleConfig() string {
	return sampleConfig
}

func (m *Final) Description() string {
	return "Report the final metric of a series"
}

func (m *Final) Add(in telegraf.Metric) {
	id := in.HashID()
	m.metricCache[id] = in
}

func (m *Final) Push(acc telegraf.Accumulator) {
	// Preserve timestamp of original metric
	acc.SetPrecision(time.Nanosecond)

	for id, metric := range m.metricCache {
		if time.Since(metric.Time()) > time.Duration(m.SeriesTimeout) {
			fields := map[string]interface{}{}
			for _, field := range metric.FieldList() {
				fields[field.Key+"_final"] = field.Value
			}
			acc.AddFields(metric.Name(), fields, metric.Tags(), metric.Time())
			delete(m.metricCache, id)
		}
	}
}

func (m *Final) Reset() {
}

func init() {
	aggregators.Add("final", func() telegraf.Aggregator {
		return NewFinal()
	})
}
