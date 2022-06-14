//go:generate ../../../tools/readme_config_includer/generator
package temp

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type Temperature struct {
	ps system.PS
}

func (*Temperature) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("temp", func() telegraf.Input {
		return &Temperature{ps: system.NewSystemPS()}
	})
}
