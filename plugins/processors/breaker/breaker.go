package breaker

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Breaker struct {
	Enabled      bool
	Name         string
	Field        string
	ValueEnable  interface{}
	ValueDisable interface{}
}

var sampleConfig = `
[[processors.breaker]]
  # By default is false, meaning it will block
  enabled = false

  # Select which metric should be used to define the state of the breaker
  name = "flag_metric"
  field = "value"

  # Which values of the selected metric will be used to enable or disable the breaker
  # These values will be transformed to string to be compared with metrics values (also converted to strings)
  value_enable = "foo"
  value_disable = "bar"
`

func (b *Breaker) SampleConfig() string {
	return sampleConfig
}

func (b *Breaker) Description() string {
	return "Print all metrics that pass through this filter."
}

func (b *Breaker) Apply(in ...telegraf.Metric) []telegraf.Metric {
	acceptedMetrics := []telegraf.Metric{}
	for _, metric := range in {
		// Capture metric used as the breaker setter
		if metric.Name() == b.Name {
			value, exists := metric.GetField(b.Field)
			if exists {
				// Compare values as string, to avoid returning false because int(1) != int32(1)
				if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", b.ValueEnable) {
					b.Enabled = true
				} else if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", b.ValueDisable) {
					b.Enabled = false
				}
			}
		}

		// Ignore metrics if breaker is active
		if !b.Enabled {
			acceptedMetrics = append(acceptedMetrics, metric)
		}
	}
	return acceptedMetrics
}

func init() {
	processors.Add("breaker", func() telegraf.Processor {
		return &Breaker{
			Enabled: true,
		}
	})
}
