//go:generate ../../../tools/readme_config_includer/generator
package infiniband

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Infiniband struct {
	Log  telegraf.Logger `toml:"-"`
	RDMA bool            `toml:"rdma" default:"false"`
}

func (*Infiniband) SampleConfig() string {
	return sampleConfig
}

// Initialise plugin
func init() {
	inputs.Add("infiniband", func() telegraf.Input { return &Infiniband{} })
}
