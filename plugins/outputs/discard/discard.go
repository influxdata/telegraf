//go:generate ../../../tools/readme_config_includer/generator
package discard

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Discard struct{}

func (*Discard) SampleConfig() string {
	return sampleConfig
}

func (*Discard) Connect() error { return nil }
func (*Discard) Close() error   { return nil }
func (*Discard) Write([]telegraf.Metric) error {
	return nil
}

func init() {
	outputs.Add("discard", func() telegraf.Output { return &Discard{} })
}
