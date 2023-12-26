package psi

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/prometheus/procfs"
)

// Psi - Plugins main structure

type Psi struct{}

// Description returns the plugin description
func (psi *Psi) Description() string {
	return `
    A plugin to gather resource pressure metrics from the Linux kernel.
    Pressure Stall Information (PSI) is available at
    "/proc/pressure/" -- cpu, memory and io.
    
    Examples:
    /proc/pressure/cpu
    some avg10=1.53 avg60=1.87 avg300=1.73 total=1088168194
    
    /proc/pressure/memory
    some avg10=0.00 avg60=0.00 avg300=0.00 total=3463792
    full avg10=0.00 avg60=0.00 avg300=0.00 total=1429641
    
    /proc/pressure/io
    some avg10=0.00 avg60=0.00 avg300=0.00 total=68568296
    full avg10=0.00 avg60=0.00 avg300=0.00 total=54982338
    `
}

// SampleConfig returns sample configuration for this plugin
func (psi *Psi) SampleConfig() string {
	return `
    [[inputs.execd]]
    command = ["/usr/local/bin/psi"]
    signal = "none"
    `
}

// Gather Psi metrics
func (psi *Psi) Gather(acc telegraf.Accumulator) error {

	cpuPressure, memoryPressure, ioPressure, err := psi.getPressureValues()
	if err == nil {
		psi.uploadPressure(cpuPressure, memoryPressure, ioPressure, acc)
	}

	return nil
}

// run initially when the package is imported
func init() {
	inputs.Add("psi", func() telegraf.Input { return &Psi{} })
}

// getPressureValues - Get the pressure values from /proc/pressure/*
func (psi *Psi) getPressureValues() (cpuPressure procfs.PSIStats, memoryPressure procfs.PSIStats, ioPressure procfs.PSIStats, err error) {
	procfs, err := procfs.NewFS("/proc")
	if err != nil {
		log.Fatalf("proc not available: %s", err)
	}

	cpuPressure, err = procfs.PSIStatsForResource("cpu")
	if err != nil {
		log.Fatalf("No CPU pressure found: %s", err)
	}

	memoryPressure, err = procfs.PSIStatsForResource("memory")
	if err != nil {
		log.Fatalf("No memory pressure found: %s", err)
	}

	ioPressure, err = procfs.PSIStatsForResource("io")
	if err != nil {
		log.Fatalf("No io pressure found: %s", err)
	}

	return cpuPressure, memoryPressure, ioPressure, nil

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
