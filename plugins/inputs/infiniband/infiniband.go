//go:generate ../../../tools/readme_config_includer/generator
package infiniband

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

// Stores the configuration values for the infiniband plugin - as there are no
// config values, this is intentionally empty
type Infiniband struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Infiniband) SampleConfig() string {
	return sampleConfig
}

// Initialise plugin
func init() {
	inputs.Add("infiniband", func() telegraf.Input { return &Infiniband{} })
}
