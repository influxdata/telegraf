// +build !linux

package infiniband

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (e *Infiniband) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("infiniband", func() telegraf.Input {
		log.Print("W! [inputs.infiniband] Current platform is not supported")
		return &Infiniband{}
	})
}
