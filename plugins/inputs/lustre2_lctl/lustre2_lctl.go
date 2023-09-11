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
	OST    OST             `toml:"ost"`
	MDT    MDT             `toml:"mdt"`
	Client bool            `toml:"client"`
	Log    telegraf.Logger `toml:"-"`
}

type OST struct {
	Obdfilter Obdfilter `toml:"obdfilter"`
	// Zfs       bool      `toml:"osd-zfs"`
	// Ldiskfs   bool      `toml:"osd-ldiskfs"`
}

type MDT struct {
	RecoveryStatus bool  `toml:"recovery_status"`
	Jobstats       Stats `toml:"job_stats"`
	Stats          Stats `toml:"stats"`
}

type Obdfilter struct {
	RecoveryStatus bool  `toml:"recovery_status"`
	Jobstats       Stats `toml:"job_stats"`
	Stats          Stats `toml:"stats"`
	Capacity       bool  `toml:"capacity"`
}

type Stats struct {
	RW bool `toml:"rw"`
	OP bool `toml:"operation"`
}

func (*Lustre2Lctl) SampleConfig() string {
	return sampleConfig
}

// Gather reads stats from all lustre targets
func (l *Lustre2Lctl) Gather(acc telegraf.Accumulator) error {
	gatherHealth(namespace, acc)
	gatherOST(l.OST, namespace, acc, l.Log)
	gatherMDT(l.MDT, namespace, acc)
	gatherClient(l.Client, namespace, acc)
	return nil
}

func init() {
	inputs.Add("lustre2_lctl", func() telegraf.Input {
		return &Lustre2Lctl{}
	})
}
