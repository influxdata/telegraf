// +build !linux

package synproxy

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Synproxy struct{}

func (k *Synproxy) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (k *Synproxy) Description() string {
	return ""
}

func (k *Synproxy) SampleConfig() string {
	return ""
}

func init() {
	inputs.Add("synproxy", func() telegraf.Input {
		log.Print("W! [inputs.synproxy] Current platform is not supported")
		return &Synproxy{}
	})
}
