package system

import (
	"fmt"

	"github.com/influxdb/telegraf/plugins"
)

type MemStats struct {
	ps PS
}

func (_ *MemStats) Description() string {
	return "Read metrics about memory usage"
}

func (_ *MemStats) SampleConfig() string { return "" }

func (s *MemStats) Gather(acc plugins.Accumulator) error {
	vm, err := s.ps.VMStat()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %s", err)
	}

	vmtags := map[string]string(nil)

	acc.Add("total", vm.Total, vmtags)
	acc.Add("available", vm.Available, vmtags)
	acc.Add("used", vm.Used, vmtags)
	acc.Add("used_prec", vm.UsedPercent, vmtags)
	acc.Add("free", vm.Free, vmtags)
	acc.Add("active", vm.Active, vmtags)
	acc.Add("inactive", vm.Inactive, vmtags)
	acc.Add("buffers", vm.Buffers, vmtags)
	acc.Add("cached", vm.Cached, vmtags)
	acc.Add("wired", vm.Wired, vmtags)
	acc.Add("shared", vm.Shared, vmtags)

	return nil
}

type SwapStats struct {
	ps PS
}

func (_ *SwapStats) Description() string {
	return "Read metrics about swap memory usage"
}

func (_ *SwapStats) SampleConfig() string { return "" }

func (s *SwapStats) Gather(acc plugins.Accumulator) error {
	swap, err := s.ps.SwapStat()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %s", err)
	}

	swaptags := map[string]string(nil)

	acc.Add("total", swap.Total, swaptags)
	acc.Add("used", swap.Used, swaptags)
	acc.Add("free", swap.Free, swaptags)
	acc.Add("used_perc", swap.UsedPercent, swaptags)
	acc.Add("in", swap.Sin, swaptags)
	acc.Add("out", swap.Sout, swaptags)

	return nil
}

func init() {
	plugins.Add("mem", func() plugins.Plugin {
		return &MemStats{ps: &systemPS{}}
	})

	plugins.Add("swap", func() plugins.Plugin {
		return &SwapStats{ps: &systemPS{}}
	})
}
