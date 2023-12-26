package psi

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/prometheus/procfs"
)

// Gather Psi metrics
func (psi *Psi) Gather(acc telegraf.Accumulator) error {
	cpuPressure, memoryPressure, ioPressure, err := psi.getPressureValues()
	if err != nil {
		return err
	}

	psi.uploadPressure(cpuPressure, memoryPressure, ioPressure, acc)
	return nil
}

// getPressureValues - Get the pressure values from /proc/pressure/*
func (psi *Psi) getPressureValues() (cpuPressure procfs.PSIStats, memoryPressure procfs.PSIStats, ioPressure procfs.PSIStats, err error) {
	var fs procfs.FS
	fs, err = procfs.NewDefaultFS()
	if err != nil {
		err = fmt.Errorf("procfs not available: %w", err)
		return
	}

	cpuPressure, err = fs.PSIStatsForResource("cpu")
	if err != nil {
		err = fmt.Errorf("no CPU pressure found: %w", err)
		return
	}

	memoryPressure, err = fs.PSIStatsForResource("memory")
	if err != nil {
		err = fmt.Errorf("no memory pressure found: %w", err)
		return
	}

	ioPressure, err = fs.PSIStatsForResource("io")
	if err != nil {
		err = fmt.Errorf("no io pressure found: %w", err)
		return
	}

	return
}

// uploadPressure Uploads all pressure value to corrosponding fields
func (psi *Psi) uploadPressure(cpuPressure procfs.PSIStats, memoryPressure procfs.PSIStats, ioPressure procfs.PSIStats, acc telegraf.Accumulator) {

	// pressureTotal some

	acc.AddCounter("pressureTotal", map[string]interface{}{
		"total": cpuPressure.Some.Total,
	},
		map[string]string{
			"resource": "cpu",
			"type":     "some",
		},
	)

	acc.AddCounter("pressureTotal", map[string]interface{}{
		"total": memoryPressure.Some.Total,
	},
		map[string]string{
			"resource": "memory",
			"type":     "some",
		},
	)

	acc.AddCounter("pressureTotal", map[string]interface{}{
		"total": ioPressure.Some.Total,
	},
		map[string]string{
			"resource": "io",
			"type":     "some",
		},
	)

	// pressureTotal full

	acc.AddCounter("pressureTotal", map[string]interface{}{
		"total": memoryPressure.Full.Total,
	},
		map[string]string{
			"resource": "memory",
			"type":     "full",
		},
	)

	acc.AddCounter("pressureTotal", map[string]interface{}{
		"total": ioPressure.Full.Total,
	},
		map[string]string{
			"resource": "io",
			"type":     "full",
		},
	)

	// pressure some

	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  cpuPressure.Some.Avg10,
		"avg60":  cpuPressure.Some.Avg60,
		"avg300": cpuPressure.Some.Avg300,
	},
		map[string]string{
			"resource": "cpu",
			"type":     "some",
		},
	)

	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  memoryPressure.Some.Avg10,
		"avg60":  memoryPressure.Some.Avg60,
		"avg300": memoryPressure.Some.Avg300,
	},
		map[string]string{
			"resource": "memory",
			"type":     "some",
		},
	)

	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  ioPressure.Some.Avg10,
		"avg60":  ioPressure.Some.Avg60,
		"avg300": ioPressure.Some.Avg300,
	},
		map[string]string{
			"resource": "io",
			"type":     "some",
		},
	)

	// pressure full
	// NOTE: cpu.full is omitted because it is always zero

	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  memoryPressure.Full.Avg10,
		"avg60":  memoryPressure.Full.Avg60,
		"avg300": memoryPressure.Full.Avg300,
	},
		map[string]string{
			"resource": "memory",
			"type":     "full",
		},
	)

	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  ioPressure.Full.Avg10,
		"avg60":  ioPressure.Full.Avg60,
		"avg300": ioPressure.Full.Avg300,
	},
		map[string]string{
			"resource": "io",
			"type":     "full",
		},
	)
}
