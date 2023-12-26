package psi

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Psi - Plugins main structure
type Psi struct {
	Log telegraf.Logger `toml:"-"`
}

//go:embed sample.conf
var sampleConfig string

// SampleConfig returns sample configuration for this plugin
func (psi *Psi) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description
func (psi *Psi) Description() string {
	return "Gather Pressure Stall Information (PSI) from /proc/pressure/"
}

func init() {
	inputs.Add("psi", func() telegraf.Input {
		return &Psi{}
	})
}
