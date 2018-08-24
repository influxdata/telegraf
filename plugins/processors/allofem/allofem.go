package allofem

import (
	"encoding/binary"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/sampler"
	"github.com/influxdata/telegraf/plugins/processors/stats"
	"github.com/influxdata/telegraf/plugins/processors/threshold"
)

type AllOfEm struct {
	WindowSize       int     `toml:"window_size"`
	StatsField       string  `toml:"stats_field"`
	OutlierDistance  float64 `toml:"outlier_distance"`
	PercentOfMetrics int     `toml:"percent_of_metrics"`
	Stats            stats.Stats
	Threshold        threshold.Threshold
	Sampler          sampler.Sampler
}

func (s *AllOfEm) SampleConfig() string {
	return `
[[processors.sampler]]

percent_of_metrics = 5

## field to be sampled over
sample_field = "trace_id"`
}

func (s *AllOfEm) Description() string {
	return "will pass through a random sampling of metrics"
}

func (s *AllOfEm) Apply(in ...telegraf.Metric) []telegraf.Metric {
	nMetrics := make([]telegraf.Metric, 0)
	for _, metric := range in {
		value := metric.Fields()[s.StatsField]
		if value == "" {
			return nil
		}

		if metric.Fields()["stddev_away"] != nil {
			nMetrics = append(nMetrics, metric)
		}

		hash := binary.BigEndian.Uint64([]byte(value.(string)))
		hash = hash % 100
		if hash >= 0 && hash <= uint64(s.PercentOfMetrics) {
			nMetrics = append(nMetrics, metric)
		}
	}
	return nMetrics
}

func init() {
	processors.Add("allofem", func() telegraf.Processor {
		return &AllOfEm{}
	})
}
