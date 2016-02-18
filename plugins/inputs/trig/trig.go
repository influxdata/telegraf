package trig

import (
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Trig struct {
	x         float64
	Amplitude float64
}

var TrigConfig = `
  ## Set the amplitude
  amplitude = 10.0
`

func (s *Trig) SampleConfig() string {
	return TrigConfig
}

func (s *Trig) Description() string {
	return "Inserts sine and cosine waves for demonstration purposes"
}

func (s *Trig) Gather(acc telegraf.Accumulator) error {
	sinner := math.Sin((s.x*math.Pi)/5.0) * s.Amplitude
	cosinner := math.Cos((s.x*math.Pi)/5.0) * s.Amplitude

	fields := make(map[string]interface{})
	fields["sine"] = sinner
	fields["cosine"] = cosinner

	tags := make(map[string]string)

	s.x += 1.0
	acc.AddFields("trig", fields, tags)

	return nil
}

func init() {
	inputs.Add("trig", func() telegraf.Input { return &Trig{x: 0.0} })
}
