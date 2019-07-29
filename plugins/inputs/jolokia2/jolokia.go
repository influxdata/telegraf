package jolokia2

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	inputs.Add("jolokia2_agent", func() telegraf.Input {
		return &JolokiaAgent{
			Metrics:               []MetricConfig{},
			DefaultFieldSeparator: ".",
		}
	})
	inputs.Add("jolokia2_proxy", func() telegraf.Input {
		return &JolokiaProxy{
			Metrics:               []MetricConfig{},
			DefaultFieldSeparator: ".",
		}
	})
}
