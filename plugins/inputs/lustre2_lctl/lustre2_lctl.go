//go:build linux

//go:generate ../../../tools/readme_config_includer/generator
package lustre2_lctl

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const namespace = "lustre2"

//go:embed sample.conf
var sampleConfig string

// Lustre proc files can change between versions, so we want to future-proof
// by letting people choose what to look at.
type Lustre2Lctl struct {
	OstCollect    []string        `toml:"ost_collect"`
	MdtCollect    []string        `toml:"mdt_collect"`
	ClientCollect []string        `toml:"client_collect"`
	Log           telegraf.Logger `toml:"-"`
}

func (*Lustre2Lctl) SampleConfig() string {
	return sampleConfig
}

// Gather reads stats from all lustre targets
func (l *Lustre2Lctl) Gather(acc telegraf.Accumulator) error {
	gatherHealth(namespace, acc)
	gatherOST(l.OstCollect, namespace, acc)
	gatherMDT(l.MdtCollect, namespace, acc)
	gatherClient(l.ClientCollect, namespace, acc)
	return nil
}

func init() {
	inputs.Add("lustre2_lctl", func() telegraf.Input {
		return &Lustre2Lctl{}
	})
}
