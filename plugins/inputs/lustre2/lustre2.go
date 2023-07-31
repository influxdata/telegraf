//go:build linux

//go:generate ../../../tools/readme_config_includer/generator
package lustre2

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
type Lustre2 struct {
	OST bool            `toml:"ost"`
	MDT bool            `toml:"mdt"`
	MGS bool            `toml:"mgs"`
	Log telegraf.Logger `toml:"-"`
}

func (*Lustre2) SampleConfig() string {
	return sampleConfig
}

// Gather reads stats from all lustre targets
func (l *Lustre2) Gather(acc telegraf.Accumulator) error {

	/* OST metrics. */
	if l.OST {
		mesurementPrefix := namespace + "_ost"
		if fields, tags, err := RetrvOSTRecoveryStatus(); err != nil {
			return err
		} else {
			for k, field := range fields {
				acc.AddGauge(mesurementPrefix, field, tags[k])
			}

		}

		if fields, tags, err := RetrvOSTJobStats(); err != nil {
			return err
		} else {
			for k, field := range fields {
				acc.AddGauge(mesurementPrefix, field, tags[k])
			}
		}

		if fields, tags, err := RetrvHealthCheck(); err != nil {
			return err
		} else {
			acc.AddGauge(mesurementPrefix, fields, tags)
		}
	}

	/* MDT metrics. */
	if l.MDT {
		mesurementPrefix := namespace + "_mdt"
		if fields, tags, err := RetrvMDTJobStats(); err != nil {
			return err
		} else {
			for k, field := range fields {
				acc.AddGauge(mesurementPrefix, field, tags[k])
			}
		}

		if fields, tags, err := RetrvMDTRecoveryStatus(); err != nil {
			return err
		} else {
			for k, field := range fields {
				acc.AddGauge(mesurementPrefix, field, tags[k])
			}
		}

		if fields, tags, err := RetrvHealthCheck(); err != nil {
			return err
		} else {
			acc.AddGauge(mesurementPrefix, fields, tags)
		}
	}

	/* MGS metrics. */
	// if l.MGS {

	// }

	return nil
}

func init() {
	inputs.Add("lustre2", func() telegraf.Input {
		return &Lustre2{}
	})
}
