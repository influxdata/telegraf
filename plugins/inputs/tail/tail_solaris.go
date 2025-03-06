// Skipping plugin on Solaris due to fsnotify support
//
//go:build solaris

package tail

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Tail struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Tail) SampleConfig() string {
	return sampleConfig
}

func (h *Tail) Init() error {
	h.Log.Warn("Current platform is not supported")
	return nil
}

func (*Tail) Gather(telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("tail", func() telegraf.Input {
		return &Tail{}
	})
}
