package fastly

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Fastly struct {
	ApiKey     string
}

var sampleConfig = `
  ## Fastly API key
  api_key = "" # required
`

func (f *Fastly) SampleConfig() string {
	return sampleConfig
}

func (f *Fastly) Description() string {
	return "Gathers metrics from the Fastly API"
}


func (f *Fastly) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("fastly", func() telegraf.Input {
		return &Fastly{}
	})
}
