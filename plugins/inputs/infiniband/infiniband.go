package infiniband

import (
	"strconv"
	"fmt"
	"github.com/Mellanox/rdmamap"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Stores the configuration values for the infiniband plugin - as there are no
// config values, this is intentionally empty
type Infiniband struct {

}

// Sample configuration for plugin
var InfinibandConfig = `
  ## no config required
`

func (s *Infiniband) SampleConfig() string {
	return InfinibandConfig
}

func (s *Infiniband) Description() string {
	return "Gets counters from all InfiniBand cards and ports installed"
}

// Gather statistics from our infiniband cards
func (s *Infiniband) Gather(acc telegraf.Accumulator) error {

	rdmaDevices := rdmamap.GetRdmaDeviceList()

	if len(rdmaDevices) == 0 {
		return fmt.Errorf("No InfiniBand devices found on this system! Check /sys/class/infiniband/ exists")
	}

	for _, dev := range rdmaDevices {
		devicePorts := rdmamap.GetPorts(dev)
		for _, port := range devicePorts {
			portInt, err := strconv.Atoi(port)
			if err != nil {
				return err
			}

			stats, err := rdmamap.GetRdmaSysfsStats(dev, portInt)
			if err != nil {
				return err
			}
			
			addStats(dev, port, stats, acc)
		}
	}

	return nil
}

// Add the statistics to the accumulator
func addStats(dev string, port string, stats[] rdmamap.RdmaStatEntry, acc telegraf.Accumulator) {
	
	// Allow users to filter by card and port
	tags := map[string]string{"device": dev, "port": port}
	fields := make(map[string]interface{})

	for _, entry := range stats {
		fields[entry.Name] = entry.Value
	}

	acc.AddFields("infiniband", fields, tags)
}

// Initialise plugin
func init() {
	inputs.Add("infiniband", func() telegraf.Input { return &Infiniband{} })
}
