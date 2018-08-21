package average

import (
	"encoding/binary"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Average struct {
	GroupSize      int    `toml:"group_size"`
	AverageField   string `toml:"average_field"`
	Sum            int
	TotalProcessed int
}

func (s *Sampler) SampleConfig() string {
	return `
[[processors.sampler]]

count = 5

## fields added all up
field_sum

## field to be sampled over
average_field = "trace_id"`
}

func (a *Average) Description() string {
	return "will pass through an average sampling of metrics"
}

func (a *Average) Apply(in ...telegraf.Metric) []telegraf.Metric {
	nMetrics := make([]telegraf.Metric, 0)
	for _, metric := range in {
		value := metric.Fields()[a.AverageField]
		if value == "" {
			return nil
		}

		hash := binary.BigEndian.Uint64([]byte(value.(string)))
		hash = hash % 100
		if hash >= 0 && hash <= uint64(s.PercentOfMetrics) {
			nMetrics = append(nMetrics, in...)
		}
	}
	return nMetrics
}

func init() {
	processors.Add("average", func() telegraf.Processor {
		return &Average{}
	})
}
