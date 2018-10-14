// +build linux

package ipvs

import (
	"errors"
	"strconv"

	"github.com/amoghe/libipvs"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// IPVS holds the state for this input plugin
type IPVS struct {
	handle     libipvs.IPVSHandle
	handleOpen bool
}

// Description returns a description string
func (i *IPVS) Description() string {
	return "Read IPVS stats"
}

// SampleConfig returns a sample configuration for this input plugin
func (i *IPVS) SampleConfig() string {
	return `
  ## This plugin takes no configuration, it collects data on all the servers
  ## (virtual and real) discovered on the system it is running on.
  ## In the future we may add a way to monitor only a subset of the servers.
`
}

// Gather gathers the stats
func (i *IPVS) Gather(acc telegraf.Accumulator) error {

	// helper: given a Service, return tags that identify it
	serviceTags := func(s *libipvs.Service) map[string]string {
		if s.FWMark > 0 {
			return map[string]string{
				"fwmark": strconv.Itoa(int(s.FWMark)),
			}
		}
		return map[string]string{
			"protocol": s.Protocol.String(),
			"address":  s.Address.String(),
			"port":     strconv.Itoa(int(s.Port)),
		}
	}

	if i.handleOpen == false {
		h, err := libipvs.New()
		if err != nil {
			return errors.New("Unable to open IPVS handle")
		}
		i.handle = h
		i.handleOpen = true
	}

	services, err := i.handle.ListServices()
	if err != nil {
		i.handle.Close()
		i.handleOpen = false // trigger a reopen on next call to gather
		return errors.New("Failed to list IPVS services")
	}
	for _, s := range services {
		fields := map[string]interface{}{
			"connections": s.Stats.Connections,
			"pkts_in":     s.Stats.PacketsIn,
			"pkts_out":    s.Stats.PacketsOut,
			"bytes_in":    s.Stats.BytesIn,
			"bytes_out":   s.Stats.BytesOut,
			"pps_in":      s.Stats.PPSIn,
			"pps_out":     s.Stats.PPSOut,
			"cps":         s.Stats.CPS,
		}
		acc.AddGauge("ipvs.virtual_server", fields, serviceTags(s))
	}

	return nil
}

func init() {
	inputs.Add("ipvs", func() telegraf.Input { return &IPVS{} })
}
