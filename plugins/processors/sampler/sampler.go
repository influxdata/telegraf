package sampler

import (
	"encoding/binary"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Sampler struct {
	PercentOfMetrics int    `toml:"percent_of_metrics"`
	SampleField      string `toml:"sample_field"`
}

func (s *Sampler) SampleConfig() string {
	return `
[[processors.sampler]]

percent_of_metrics = 5

## field to be sampled over
sample_field = "trace_id"`
}

func (s *Sampler) Description() string {
	return "will pass through a random sampling of metrics"
}

func (s *Sampler) Apply(in ...telegraf.Metric) []telegraf.Metric {
	nMetrics := make([]telegraf.Metric, 0)
	for _, metric := range in {
		value := metric.Fields()[s.SampleField]
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
	processors.Add("sampler", func() telegraf.Processor {
		return &Sampler{}
	})
}
