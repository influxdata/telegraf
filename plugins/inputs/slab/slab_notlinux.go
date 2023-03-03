//go:build !linux

package slab

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Slab struct {
	Log telegraf.Logger `toml:"-"`
}

func (s *Slab) Init() error {
	s.Log.Warn("current platform is not supported")
	return nil
}
func (s *Slab) SampleConfig() string                { return sampleConfig }
func (s *Slab) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("slab", func() telegraf.Input {
		return &Slab{}
	})
}
