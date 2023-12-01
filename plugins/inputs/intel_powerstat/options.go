//go:build linux && amd64

package intel_powerstat

import (
	"time"

	ptel "github.com/intel/powertelemetry"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
)

// optConfig represents plugin configuration fields needed to generate options.
type optConfig struct {
	cpuMetrics     []string
	packageMetrics []string
	includedCPUs   []int
	excludedCPUs   []int
	perfEventFile  string
	msrReadTimeout time.Duration
	log            telegraf.Logger
}

// optionGenerator takes a struct with the plugin configuration, and generates options
// needed to gather metrics.
type optionGenerator interface {
	generate(cfg optConfig) []ptel.Option
}

// optGenerator implements optionGenerator interface.
type optGenerator struct{}

// generate takes plugin configuration options and generates options needed
// to gather requested metrics.
func (g *optGenerator) generate(cfg optConfig) []ptel.Option {
	opts := make([]ptel.Option, 0)

	metrics := make([]string, 0, len(cfg.cpuMetrics)+len(cfg.packageMetrics))
	metrics = append(metrics, cfg.cpuMetrics...)
	metrics = append(metrics, cfg.packageMetrics...)

	if len(cfg.includedCPUs) != 0 {
		opts = append(opts, ptel.WithIncludedCPUs(cfg.includedCPUs))
	}

	if len(cfg.excludedCPUs) != 0 {
		opts = append(opts, ptel.WithExcludedCPUs(cfg.excludedCPUs))
	}

	if needsMsr(metrics) {
		if cfg.msrReadTimeout == 0 {
			opts = append(opts, ptel.WithMsr())
		} else {
			opts = append(opts, ptel.WithMsrTimeout(cfg.msrReadTimeout))
		}
	}

	if needsRapl(metrics) {
		opts = append(opts, ptel.WithRapl())
	}

	if needsCoreFreq(metrics) {
		opts = append(opts, ptel.WithCoreFrequency())
	}

	if needsUncoreFreq(metrics) {
		opts = append(opts, ptel.WithUncoreFrequency())
	}

	if needsPerf(metrics) {
		opts = append(opts, ptel.WithPerf(cfg.perfEventFile))
	}

	if cfg.log != nil {
		opts = append(opts, ptel.WithLogger(cfg.log))
	}

	return opts
}

// needsMsr takes a slice of strings, representing supported metrics, and
// returns true if any relies on msr registers.
func needsMsr(metrics []string) bool {
	slice := []string{
		packageCPUBaseFrequency.String(),
		cpuTemperature.String(),
		cpuC0StateResidency.String(),
		cpuC1StateResidency.String(),
		cpuC3StateResidency.String(),
		cpuC6StateResidency.String(),
		cpuC7StateResidency.String(),
		cpuBusyCycles.String(),
		cpuBusyFrequency.String(),
		packageTurboLimit.String(),

		// Fallback mechanism retrieves this value from MSR registers.
		packageUncoreFrequency.String(),
	}

	for i := range metrics {
		if choice.Contains(metrics[i], slice) {
			return true
		}
	}
	return false
}

// needsTimeRelatedMsr takes a slice of strings, representing supported metrics, and
// returns true if any relies on time-related reads of msr registers.
func needsTimeRelatedMsr(metrics []string) bool {
	slice := []string{
		cpuC0StateResidency.String(),
		cpuC1StateResidency.String(),
		cpuC3StateResidency.String(),
		cpuC6StateResidency.String(),
		cpuC7StateResidency.String(),
		cpuBusyCycles.String(),
		cpuBusyFrequency.String(),
	}

	for i := range metrics {
		if choice.Contains(metrics[i], slice) {
			return true
		}
	}
	return false
}

// needsRapl takes a slice of strings, representing supported metrics, and
// returns true if any relies on intel-rapl control zone.
func needsRapl(metrics []string) bool {
	slice := []string{
		packageCurrentPowerConsumption.String(),
		packageCurrentDramPowerConsumption.String(),
		packageThermalDesignPower.String(),
	}

	for i := range metrics {
		if choice.Contains(metrics[i], slice) {
			return true
		}
	}
	return false
}

// needsCoreFreq takes a slice of strings, representing supported metrics, and
// returns true if any relies on sysfs "/sys/devices/system/cpu/" with global and
// individual CPU attributes.
func needsCoreFreq(metrics []string) bool {
	return choice.Contains(cpuFrequency.String(), metrics)
}

// needsUncoreFreq takes a slice of strings, representing supported metrics, and returns
// true if any relies on sysfs interface "/sys/devices/system/cpu/intel_uncore_frequency/"
// provided by intel_uncore_frequency kernel module.
func needsUncoreFreq(metrics []string) bool {
	return choice.Contains(packageUncoreFrequency.String(), metrics)
}

// needsPerf takes a slice of strings, representing supported metrics, and
// returns true if any relies on perf_events interface.
func needsPerf(metrics []string) bool {
	slice := []string{
		cpuC0SubstateC01Percent.String(),
		cpuC0SubstateC02Percent.String(),
		cpuC0SubstateC0WaitPercent.String(),
	}

	for i := range metrics {
		if choice.Contains(metrics[i], slice) {
			return true
		}
	}
	return false
}
