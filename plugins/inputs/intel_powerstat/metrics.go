//go:build linux && amd64

package intel_powerstat

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	ptel "github.com/intel/powertelemetry"

	"github.com/influxdata/telegraf"
)

// cpuMetricType is an enum type to identify core metrics.
type cpuMetricType int

// cpuMetricType enum defines supported core metrics.
const (
	// metric relying on cpuFreq
	cpuFrequency cpuMetricType = iota

	// metric relying on msr
	cpuTemperature

	// metrics relying on msr with storage
	cpuC0StateResidency
	cpuC1StateResidency
	cpuC3StateResidency
	cpuC6StateResidency
	cpuC7StateResidency
	cpuBusyCycles // alias of cpuC0StateResidency
	cpuBusyFrequency

	// metrics relying on perf
	cpuC0SubstateC01Percent
	cpuC0SubstateC02Percent
	cpuC0SubstateC0WaitPercent
)

// Helper method to return a string representation of a core metric.
func (m cpuMetricType) String() string {
	switch m {
	case cpuFrequency:
		return "cpu_frequency"
	case cpuTemperature:
		return "cpu_temperature"
	case cpuBusyFrequency:
		return "cpu_busy_frequency"
	case cpuC0StateResidency:
		return "cpu_c0_state_residency"
	case cpuC1StateResidency:
		return "cpu_c1_state_residency"
	case cpuC3StateResidency:
		return "cpu_c3_state_residency"
	case cpuC6StateResidency:
		return "cpu_c6_state_residency"
	case cpuC7StateResidency:
		return "cpu_c7_state_residency"
	case cpuBusyCycles:
		return "cpu_busy_cycles"
	case cpuC0SubstateC01Percent:
		return "cpu_c0_substate_c01"
	case cpuC0SubstateC02Percent:
		return "cpu_c0_substate_c02"
	case cpuC0SubstateC0WaitPercent:
		return "cpu_c0_substate_c0_wait"
	}
	return ""
}

// isValidCoreMetric returns true if a given metric is a supported core metric.
func isValidCoreMetric(metric string) bool {
	switch metric {
	case cpuFrequency.String():
	case cpuTemperature.String():
	case cpuBusyFrequency.String():
	case cpuC0StateResidency.String():
	case cpuC1StateResidency.String():
	case cpuC3StateResidency.String():
	case cpuC6StateResidency.String():
	case cpuC7StateResidency.String():
	case cpuBusyCycles.String():
	case cpuC0SubstateC01Percent.String():
	case cpuC0SubstateC02Percent.String():
	case cpuC0SubstateC0WaitPercent.String():
	default:
		return false
	}
	return true
}

// packageMetricType is an enum type to identify package metrics.
type packageMetricType int

// packageMetricType enum defines supported package metrics.
const (
	// metrics relying on rapl
	packageCurrentPowerConsumption packageMetricType = iota
	packageCurrentDramPowerConsumption
	packageThermalDesignPower

	// metrics relying on msr
	packageCPUBaseFrequency

	// hybrid metric relying on uncoreFreq as a primary mechanism and on msr as fallback mechanism.
	packageUncoreFrequency

	// metrics relying on msr
	packageTurboLimit
)

// Helper method to return a string representation of a package metric.
func (m packageMetricType) String() string {
	switch m {
	case packageCurrentPowerConsumption:
		return "current_power_consumption"
	case packageCurrentDramPowerConsumption:
		return "current_dram_power_consumption"
	case packageThermalDesignPower:
		return "thermal_design_power"
	case packageCPUBaseFrequency:
		return "cpu_base_frequency"
	case packageUncoreFrequency:
		return "uncore_frequency"
	case packageTurboLimit:
		return "max_turbo_frequency"
	}
	return ""
}

// isValidCoreMetric returns true if a given metric is a supported package metric.
func isValidPackageMetric(metric string) bool {
	switch metric {
	case packageCurrentPowerConsumption.String():
	case packageCurrentDramPowerConsumption.String():
	case packageThermalDesignPower.String():
	case packageCPUBaseFrequency.String():
	case packageUncoreFrequency.String():
	case packageTurboLimit.String():
	default:
		return false
	}
	return true
}

// Numeric is a type constraint definition.
type Numeric interface {
	float64 | uint64
}

// metricInfoProvider provides measurement name, fields, and tags needed by the accumulator to add a metric.
type metricInfoProvider interface {
	// measurement returns a string with the name of measurement.
	measurement() string

	// fields returns a map of string keys with metric name and metric values.
	fields() (map[string]interface{}, error)

	// tags returns a map of string key and string value to add additional metric-specific information.
	tags() map[string]string

	// name returns the name of a metric.
	name() string
}

// addMetric takes a metricInfoProvider interface and adds metric information to an accumulator.
func addMetric(acc telegraf.Accumulator, m metricInfoProvider, logOnceMap map[string]struct{}) {
	fields, err := m.fields()
	if err == nil {
		acc.AddGauge(
			m.measurement(),
			fields,
			m.tags(),
		)
		return
	}

	// Always add to the accumulator errors not related to module not initialized.
	var moduleErr *ptel.ModuleNotInitializedError
	if !errors.As(err, &moduleErr) {
		acc.AddError(err)
		return
	}

	// Add only once module not initialized error related to module and metric name.
	logErrorOnce(
		acc,
		logOnceMap,
		fmt.Sprintf("%s_%s", moduleErr.Name, m.name()),
		fmt.Errorf("failed to get %q: %w", m.name(), moduleErr),
	)
}

// metricCommon has metric information common to different types.
type metricCommon struct {
	metric interface{}
	units  string
}

func (m *metricCommon) name() string {
	switch m.metric.(type) {
	case cpuMetricType:
		return m.metric.(cpuMetricType).String()
	case packageMetricType:
		return m.metric.(packageMetricType).String()
	default:
		return ""
	}
}

func (m *metricCommon) measurement() string {
	switch m.metric.(type) {
	case cpuMetricType:
		return "powerstat_core"
	case packageMetricType:
		return "powerstat_package"
	default:
		return ""
	}
}

// cpuMetric is a generic type that has the information to identify a CPU-related metric,
// as well as function to retrieve its value at any time. Implements metricAdder interface.
type cpuMetric[T Numeric] struct {
	metricCommon

	cpuID     int
	coreID    int
	packageID int
	fetchFn   func(cpuID int) (T, error)
}

func (m *cpuMetric[T]) fields() (map[string]interface{}, error) {
	val, err := m.fetchFn(m.cpuID)
	if err != nil {
		return nil, fmt.Errorf("failed to get %q for CPU ID %v: %w", m.metric, m.cpuID, err)
	}

	return map[string]interface{}{
		fmt.Sprintf("%s_%s", m.metric, m.units): round(val),
	}, nil
}

func (m *cpuMetric[T]) tags() map[string]string {
	return map[string]string{
		"core_id":    strconv.Itoa(m.coreID),
		"cpu_id":     strconv.Itoa(m.cpuID),
		"package_id": strconv.Itoa(m.packageID),
	}
}

// packageMetric is a generic type that has the information to identify a package-related metric,
// as well as the function to retrieve its value at any time. Implements metricAdder interface.
type packageMetric[T Numeric] struct {
	metricCommon

	packageID int
	fetchFn   func(packageID int) (T, error)
}

//nolint:revive // Confusing-naming caused by a generic type that implements this interface method.
func (m *packageMetric[T]) fields() (map[string]interface{}, error) {
	val, err := m.fetchFn(m.packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get %q for package ID %v: %w", m.metric, m.packageID, err)
	}

	return map[string]interface{}{
		fmt.Sprintf("%s_%s", m.metric, m.units): round(val),
	}, nil
}

//nolint:revive // Confusing-naming caused by a generic type that implements this interface method.
func (m *packageMetric[T]) tags() map[string]string {
	return map[string]string{
		"package_id": strconv.Itoa(m.packageID),
	}
}

// round returns the result of rounding the argument, only if its a 64 bit floating-point type.
func round[T Numeric](val T) T {
	if v, ok := any(val).(float64); ok {
		val = T(math.Round(v*100) / 100)
	}
	return val
}
