//go:generate ../../../tools/readme_config_includer/generator
package trig

import (
	_ "embed"
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type Trig struct {
	x         float64
	Amplitude float64
}

func (*Trig) SampleConfig() string {
	return sampleConfig
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
