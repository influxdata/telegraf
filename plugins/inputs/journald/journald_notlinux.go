//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package journald

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Journald struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Journald) SampleConfig() string { return sampleConfig }

func (j *Journald) Init() error {
	j.Log.Warn("Current platform is not supported")
	return nil
}

func (*Journald) Start(telegraf.Accumulator) error { return nil }

func (*Journald) Gather(telegraf.Accumulator) error { return nil }

func (*Journald) Stop() {}

func init() {
	inputs.Add("journald", func() telegraf.Input {
		return &Journald{}
	})
}
