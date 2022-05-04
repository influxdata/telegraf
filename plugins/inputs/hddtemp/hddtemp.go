package hddtemp

import (
	"net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	gohddtemp "github.com/influxdata/telegraf/plugins/inputs/hddtemp/go-hddtemp"
)

const defaultAddress = "127.0.0.1:7634"

type HDDTemp struct {
	Address string
	Devices []string
	fetcher Fetcher
}

type Fetcher interface {
	Fetch(address string) ([]gohddtemp.Disk, error)
}

var hddtempSampleConfig = `
  ## By default, telegraf gathers temps data from all disks detected by the
  ## hddtemp.
  ##
  ## Only collect temps from the selected disks.
  ##
  ## A * as the device name will return the temperature values of all disks.
  ##
  # address = "127.0.0.1:7634"
  # devices = ["sda", "*"]
`

func (h *HDDTemp) Gather(acc telegraf.Accumulator) error {
	if h.fetcher == nil {
		h.fetcher = gohddtemp.New()
	}
	source, _, err := net.SplitHostPort(h.Address)
	if err != nil {
		source = h.Address
	}

	disks, err := h.fetcher.Fetch(h.Address)
	if err != nil {
		return err
	}

	for _, disk := range disks {
		for _, chosenDevice := range h.Devices {
			if chosenDevice == "*" || chosenDevice == disk.DeviceName {
				tags := map[string]string{
					"device": disk.DeviceName,
					"model":  disk.Model,
					"unit":   disk.Unit,
					"status": disk.Status,
					"source": source,
				}

				fields := map[string]interface{}{
					"temperature": disk.Temperature,
				}

				acc.AddFields("hddtemp", fields, tags)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("hddtemp", func() telegraf.Input {
		return &HDDTemp{
			Address: defaultAddress,
			Devices: []string{"*"},
		}
	})
}
