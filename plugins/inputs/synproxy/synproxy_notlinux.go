// +build !linux

package synproxy

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (k *Synproxy) Init() error {
	log.Warn("Current platform is not supported")
}

func (k *Synproxy) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("synproxy", func() telegraf.Input {
		return &Synproxy{}
	})
}
