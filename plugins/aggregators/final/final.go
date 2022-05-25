//go:generate ../../../tools/readme_config_includer/generator
package final

import (
	_ "embed"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embedd the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

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

func (*Final) SampleConfig() string {
	return sampleConfig
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
