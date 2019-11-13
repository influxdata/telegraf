// +build !linux

package ethtool

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (e *Ethtool) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		log.Print("W! [inputs.ethtool] Current platform is not supported")
		return &Ethtool{}
	})
}
