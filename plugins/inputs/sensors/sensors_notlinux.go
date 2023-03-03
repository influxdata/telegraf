//go:build !linux

package sensors

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Sensors struct {
	Log telegraf.Logger `toml:"-"`
}

func (s *Sensors) Init() error {
	s.Log.Warn("current platform is not supported")
	return nil
}
func (s *Sensors) SampleConfig() string                { return sampleConfig }
func (s *Sensors) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("sensors", func() telegraf.Input {
		return &Sensors{}
	})
}
