package fastly

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
	"time"
)

const statPrefix = "fastly."

type Fastly struct {
	ApiKey              string            `toml:"api_key"`
	ServiceUpdatePeriod internal.Duration `toml:"service_update_period"`

	client          *fastly.Client
	services        []*fastly.Service
	rtClient        *fastly.RTSClient
	rtUpdateTracker *realtimeUpdateTracker
}

var sampleConfig = `
  ## Fastly API key.
  api_key = "" # required
  ## How often to refresh our local Fastly service list.
  service_update_period = "1m"
`

func (f *Fastly) SampleConfig() string {
	return sampleConfig
}

func (f *Fastly) Description() string {
	return "Gathers metrics from the Fastly API"
}

func (f *Fastly) Gather(acc telegraf.Accumulator) error {
	log.Printf("D! [inputs.fastly] Gathering Fastly")
	if err := f.ensureFastlyClients(); err != nil {
		return err
	}
	if err := f.ensureFastlyServiceList(); err != nil {
		return err
	}
	if err := f.collectRealtimeStats(acc); err != nil {
		return err
	}
	return nil
}

func init() {
	fastlyInput := Fastly{
		ServiceUpdatePeriod: internal.Duration{
			Duration: time.Duration(1 * time.Minute),
		},
		rtUpdateTracker: newRealtimeUpdateTracker(),
	}
	inputs.Add("fastly", func() telegraf.Input { return &fastlyInput })
}
