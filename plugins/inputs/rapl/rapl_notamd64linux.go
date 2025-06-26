//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux || !amd64

package rapl

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type RAPL struct {
	Log telegraf.Logger `toml:"-"`
}

func (*RAPL) SampleConfig() string { return sampleConfig }

func (o *RAPL) Init() error {
	o.Log.Warn("Current platform is not supported")
	return nil
}

func (*RAPL) Gather(telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("rapl", func() telegraf.Input {
		return &RAPL{}
	})
}
