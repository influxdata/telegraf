//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package lustre2_lctl

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Lustre2Lctl struct {
	Log telegraf.Logger `toml:"-"`
}

func (l *Lustre2Lctl) Init() error {
	l.Log.Warn("current platform is not supported")
	return nil
}
func (*Lustre2Lctl) SampleConfig() string                { return sampleConfig }
func (*Lustre2Lctl) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("lustre2_lctl", func() telegraf.Input {
		return &Lustre2Lctl{}
	})
}
