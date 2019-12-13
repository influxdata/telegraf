// +build !linux

package synproxy

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Synproxy struct {
	Log telegraf.Logger `toml:"-"`
}

func (k *Synproxy) Init() error {
	log.Warn("Current platform is not supported")
}

func (k *Synproxy) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (k *Synproxy) Description() string {
	return "Get synproxy counter statistics from procfs"
}

func (k *Synproxy) SampleConfig() string {
	return ""
}

func init() {
	inputs.Add("synproxy", func() telegraf.Input {
		return &Synproxy{}
	})
}
