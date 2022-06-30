//go:generate ../../../tools/readme_config_includer/generator
package discard

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type Discard struct{}

func (*Discard) SampleConfig() string {
	return sampleConfig
}

func (d *Discard) Connect() error { return nil }
func (d *Discard) Close() error   { return nil }
func (d *Discard) Write(_ []telegraf.Metric) error {
	return nil
}

func init() {
	outputs.Add("discard", func() telegraf.Output { return &Discard{} })
}
