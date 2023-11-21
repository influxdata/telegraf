//go:generate ../../../tools/readme_config_includer/generator
package filter

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Noop struct{}

func (*Noop) SampleConfig() string {
	return sampleConfig
}

func (p *Noop) Apply(in ...telegraf.Metric) []telegraf.Metric {
	return in
}

func init() {
	processors.Add("noop", func() telegraf.Processor {
		return &Noop{}
	})
}
