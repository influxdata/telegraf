//go:generate ../../../tools/readme_config_includer/generator
package swap

import (
	_ "embed"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type SwapStats struct {
	ps system.PS
}

func (*SwapStats) SampleConfig() string {
	return sampleConfig
}

func (ss *SwapStats) Gather(acc telegraf.Accumulator) error {
	swap, err := ss.ps.SwapStat()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %s", err)
	}

	fieldsG := map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,
	}
	fieldsC := map[string]interface{}{
		"in":  swap.Sin,
		"out": swap.Sout,
	}
	acc.AddGauge("swap", fieldsG, nil)
	acc.AddCounter("swap", fieldsC, nil)

	return nil
}

func init() {
	ps := system.NewSystemPS()
	inputs.Add("swap", func() telegraf.Input {
		return &SwapStats{ps: ps}
	})
}
