//go:build linux && amd64

package intel_powerstat

import (
	ptel "github.com/intel/powertelemetry"
)

// TopologyFetcher fetches topology information of the host.
type TopologyFetcher interface {
	// GetMsrCPUIDs returns a slice with available CPU IDs of the host for which msr will access to.
	GetMsrCPUIDs() []int

	// GetPerfCPUIDs returns a slice with available CPU IDs of the host for which perf will access to.
	GetPerfCPUIDs() []int

	// GetPackageIDs returns a slice with available package IDs of the host.
	GetPackageIDs() []int

	// GetCPUPackageID returns the package ID of the host corresponding to the given CPU ID.
	GetCPUPackageID(cpuID int) (int, error)

	// GetCPUCoreID returns the core ID of the host corresponding to the given CPU ID.
	GetCPUCoreID(cpuID int) (int, error)

	// GetPackageDieIDs returns the die IDs of the host corresponding to the given package ID.
	GetPackageDieIDs(packageID int) ([]int, error)
}

// CPUFreqFetcher fetches supported CPU-related metrics relying on core frequency.
type CPUFreqFetcher interface {
	// GetCPUFrequency returns the current frequency value of a given CPU ID, in MHz.
	GetCPUFrequency(cpuID int) (float64, error)
}

// CPUMsrFetcher fetches supported CPU-related metrics relying on msr registers.
type CPUMsrFetcher interface {
	// GetCPUTemperature returns the temperature value of a given CPU ID, in degrees Celsius.
	GetCPUTemperature(cpuID int) (uint64, error)

	// UpdatePerCPUMetrics reads multiple MSR offsets needed to get metric values that are time sensitive.
	// Below are the list of methods that need the update to be performed beforehand.
	UpdatePerCPUMetrics(cpuID int) error

	// GetCPUC0StateResidency returns the C0 state residency value of a given CPU ID, as a percentage.
	GetCPUC0StateResidency(cpuID int) (float64, error)

	// GetCPUC1StateResidency returns the C1 state residency value of a given CPU ID, as a percentage.
	GetCPUC1StateResidency(cpuID int) (float64, error)

	// GetCPUC3StateResidency returns the C3 state residency value of a given CPU ID, as a percentage.
	GetCPUC3StateResidency(cpuID int) (float64, error)

	// GetCPUC6StateResidency returns the C6 state residency value of a given CPU ID, as a percentage.
	GetCPUC6StateResidency(cpuID int) (float64, error)

	// GetCPUC7StateResidency returns the C7 state residency value of a given CPU ID, as a percentage.
	GetCPUC7StateResidency(cpuID int) (float64, error)

	// GetCPUBusyFrequencyMhz returns the busy frequency value of a given CPU ID, in MHz.
	GetCPUBusyFrequencyMhz(cpuID int) (float64, error)
}

// CPUPerfFetcher fetches supported CPU-related metrics relying on perf events.
type CPUPerfFetcher interface {
	// ReadPerfEvents reads values of perf events needed to get C0X state residency metrics.
	// Below getter methods that need this operation to be performed previously.
	ReadPerfEvents() error

	// DeactivatePerfEvents deactivates perf events. It closes file descriptors used to get perf event values.
	DeactivatePerfEvents() error

	// GetCPUC0SubstateC01Percent takes a CPU ID and returns a value indicating the percentage of time
	// the processor spent in its C0.1 substate out of the total time in the C0 state.
	// C0.1 is characterized by a light-weight slower wakeup time but more power-saving optimized state.
	GetCPUC0SubstateC01Percent(cpuID int) (float64, error)

	// GetCPUC0SubstateC02Percent takes a CPU ID and returns a value indicating the percentage of time
	// the processor spent in its C0.2 substate out of the total time in the C0 state.
	// C0.2 is characterized by a light-weight faster wakeup time but less power saving optimized state.
	GetCPUC0SubstateC02Percent(cpuID int) (float64, error)

	// GetCPUC0SubstateC0WaitPercent takes a CPU ID and returns a value indicating the percentage of time
	// the processor spent in its C0_Wait substate out of the total time in the C0 state.
	// CPU is in C0_Wait substate when the thread is in the C0.1 or C0.2 or running a PAUSE in C0 ACPI state.
	GetCPUC0SubstateC0WaitPercent(cpuID int) (float64, error)
}

// PackageRaplFetcher fetches supported package related metrics relying on rapl.
type PackageRaplFetcher interface {
	// GetCurrentPackagePowerConsumptionWatts returns the current package power consumption value of a given package ID, in watts.
	GetCurrentPackagePowerConsumptionWatts(packageID int) (float64, error)

	// GetCurrentDramPowerConsumptionWatts returns the current dram power consumption value of a given package ID, in watts.
	GetCurrentDramPowerConsumptionWatts(packageID int) (float64, error)

	// GetPackageThermalDesignPowerWatts returns the thermal power design value of a given package ID, in watts.
	GetPackageThermalDesignPowerWatts(packageID int) (float64, error)
}

// PackageUncoreFreqFetcher fetches supported package related metrics relying on uncore frequency.
type PackageUncoreFreqFetcher interface {
	// GetInitialUncoreFrequencyMin returns the minimum initial uncore frequency value of a given package ID, in MHz.
	GetInitialUncoreFrequencyMin(packageID, dieID int) (float64, error)

	// GetInitialUncoreFrequencyMax returns the maximum initial uncore frequency value of a given package ID, in MHz.
	GetInitialUncoreFrequencyMax(packageID, dieID int) (float64, error)

	// GetCustomizedUncoreFrequencyMin returns the minimum custom uncore frequency value of a given package ID, in MHz.
	GetCustomizedUncoreFrequencyMin(packageID, dieID int) (float64, error)

	// GetCustomizedUncoreFrequencyMax returns the maximum custom uncore frequency value of a given package ID, in MHz.
	GetCustomizedUncoreFrequencyMax(packageID, dieID int) (float64, error)

	// GetCurrentUncoreFrequency returns the current uncore frequency value of a given package ID, in MHz.
	GetCurrentUncoreFrequency(packageID, dieID int) (float64, error)
}

// PackageMsrFetcher fetches supported package related metrics relying on msr registers.
type PackageMsrFetcher interface {
	// GetCPUBaseFrequency returns the CPU base frequency value of a given package ID, in MHz.
	GetCPUBaseFrequency(packageID int) (uint64, error)

	// GetMaxTurboFreqList returns a list of max turbo frequencies and related active cores of a given package ID.
	GetMaxTurboFreqList(packageID int) ([]ptel.MaxTurboFreq, error)
}

// MetricFetcher fetches metrics supported by this plugin.
type MetricFetcher interface {
	TopologyFetcher

	CPUFreqFetcher
	CPUMsrFetcher
	CPUPerfFetcher

	PackageRaplFetcher
	PackageUncoreFreqFetcher
	PackageMsrFetcher
}
