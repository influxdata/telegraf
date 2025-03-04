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
	RDMA bool            `toml:"gather_rdma"`
	Log  telegraf.Logger `toml:"-"`
}

func (*Infiniband) SampleConfig() string {
	return sampleConfig
}

// Initialise plugin
func init() {
	inputs.Add("infiniband", func() telegraf.Input { return &Infiniband{} })
}
