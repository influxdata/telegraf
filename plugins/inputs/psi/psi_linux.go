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
func (psi *Psi) getPressureValues() (pressures map[string]procfs.PSIStats, err error) {
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
func (psi *Psi) uploadPressure(pressures map[string]procfs.PSIStats, acc telegraf.Accumulator) {
	// pressureTotal type=some
	for _, resource := range []string{"cpu", "memory", "io"} {
		acc.AddCounter("pressureTotal", map[string]interface{}{
			"total": pressures[resource].Some.Total,
		},
			map[string]string{
				"resource": resource,
				"type":     "some",
			},
		)
	}

	// pressureTotal type=full
	for _, resource := range []string{"memory", "io"} {
		acc.AddCounter("pressureTotal", map[string]interface{}{
			"total": pressures[resource].Full.Total,
		},
			map[string]string{
				"resource": resource,
				"type":     "full",
			},
		)
	}

	// pressure type=some
	for _, resource := range []string{"cpu", "memory", "io"} {
		acc.AddGauge("pressure", map[string]interface{}{
			"avg10":  pressures[resource].Some.Avg10,
			"avg60":  pressures[resource].Some.Avg60,
			"avg300": pressures[resource].Some.Avg300,
		},
			map[string]string{
				"resource": resource,
				"type":     "some",
			},
		)
	}

	// pressure type=full
	for _, resource := range []string{"memory", "io"} {
		acc.AddGauge("pressure", map[string]interface{}{
			"avg10":  pressures[resource].Full.Avg10,
			"avg60":  pressures[resource].Full.Avg60,
			"avg300": pressures[resource].Full.Avg300,
		},
			map[string]string{
				"resource": resource,
				"type":     "full",
			},
		)
	}
}
