package psi

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/prometheus/procfs"
)

// Gather Psi metrics
func (psi *Psi) Gather(acc telegraf.Accumulator) error {
	pressures, err := psi.getPressureValues()
	if err != nil {
		return err
	}
	psi.uploadPressure(pressures, acc)
	return nil
}

// getPressureValues - Get the pressure values from /proc/pressure/*
func (*Psi) getPressureValues() (pressures map[string]procfs.PSIStats, err error) {
	var fs procfs.FS
	fs, err = procfs.NewDefaultFS()
	if err != nil {
		return nil, fmt.Errorf("procfs not available: %w", err)
	}

	pressures = make(map[string]procfs.PSIStats)
	for _, resource := range []string{"cpu", "memory", "io"} {
		pressures[resource], err = fs.PSIStatsForResource(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s pressure: %w", resource, err)
		}
	}
	return pressures, nil
}

// uploadPressure Uploads all pressure value to corrosponding fields
// NOTE: resource=cpu,type=full is omitted because it is always zero
func (*Psi) uploadPressure(pressures map[string]procfs.PSIStats, acc telegraf.Accumulator) {
	for _, typ := range []string{"some", "full"} {
		for _, resource := range []string{"cpu", "memory", "io"} {
			if resource == "cpu" && typ == "full" {
				continue
			}

			tags := map[string]string{
				"resource": resource,
				"type":     typ,
			}

			var stat *procfs.PSILine
			switch typ {
			case "some":
				stat = pressures[resource].Some
			case "full":
				stat = pressures[resource].Full
			}

			// pressureTotal
			acc.AddCounter("pressureTotal", map[string]interface{}{
				"total": stat.Total,
			}, tags)

			// pressure
			acc.AddGauge("pressure", map[string]interface{}{
				"avg10":  stat.Avg10,
				"avg60":  stat.Avg60,
				"avg300": stat.Avg300,
			}, tags)
		}
	}
}
