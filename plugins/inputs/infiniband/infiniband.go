package infiniband

import (
	"strconv"
	"github.com/willfurnell/rdmamap"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Infiniband struct {
}

var InfinibandConfig = `
  ## no config required
`

func (s *Infiniband) SampleConfig() string {
	return InfinibandConfig
}

func (s *Infiniband) Description() string {
	return "Gets counters from all Infiniband cards and ports installed, seperately"
}

func (s *Infiniband) Gather(acc telegraf.Accumulator) error {

	rdmaDevices := rdmamap.GetRdmaDeviceList()

	for _, dev := range rdmaDevices {
		devicePorts := rdmamap.GetPorts(dev)
		for _, port := range devicePorts {
			portInt, err := strconv.Atoi(port)
			stats, err2 := rdmamap.GetRdmaSysfsStats(dev, portInt)
			if err == nil && err2 == nil {

				tags := map[string]string{"card": dev, "port": port}
				fields := make(map[string]interface{})

				for _, entry := range stats {
					fields[entry.Name] = entry.Value
				}

				acc.AddFields("infiniband", fields, tags)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("infiniband", func() telegraf.Input { return &Infiniband{} })
}
