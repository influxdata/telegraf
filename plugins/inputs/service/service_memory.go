package service

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type MemoryStats struct {
	ps PS

	ProcessNames []string `toml:"process_names"`
}

func (_ *MemoryStats) Description() string {
	return "Read memory usage about a particular service"
}

var sampleConfig = `
  ## names of services
  ProcessNames = ["process"]
`

func (_ *MemoryStats) SampleConfig() string {
	return sampleConfig
}

func (s *MemoryStats) Gather(acc telegraf.Accumulator) error {

	for _, processName := range s.ProcessNames {
		memInfo, err := s.ps.MemInfo(processName)
		if err != nil {
			return err
		}

		fields := map[string]interface{}{
			"rss":  memInfo.RSS,
			"vms":  memInfo.VMS,
			"swap": memInfo.Swap,
		}

		tags := map[string]string{
			"name": processName,
		}

		acc.AddGauge("service_mem", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("service_mem", func() telegraf.Input {
		return &MemoryStats{ps: &servicePs{}}
	})
}
