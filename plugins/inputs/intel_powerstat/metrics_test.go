//go:build linux && amd64

package intel_powerstat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreMetric_String(t *testing.T) {
	testCases := []struct {
		name       string
		metricName string
	}{
		{
			name:       "CPUFrequency",
			metricName: "cpu_frequency",
		},
		{
			name:       "CPUTemperature",
			metricName: "cpu_temperature",
		},
		{
			name:       "CPUC0StateResidency",
			metricName: "cpu_c0_state_residency",
		},
		{
			name:       "CPUC1StateResidency",
			metricName: "cpu_c1_state_residency",
		},
		{
			name:       "CPUC3StateResidency",
			metricName: "cpu_c3_state_residency",
		},
		{
			name:       "CPUC6StateResidency",
			metricName: "cpu_c6_state_residency",
		},
		{
			name:       "CPUC7StateResidency",
			metricName: "cpu_c7_state_residency",
		},
		{
			name:       "CPUBusyFrequency",
			metricName: "cpu_busy_frequency",
		},
		{
			name:       "CPUC0SubstateC01Percent",
			metricName: "cpu_c0_substate_c01",
		},
		{
			name:       "CPUC0SubstateC02Percent",
			metricName: "cpu_c0_substate_c02",
		},
		{
			name:       "CPUC0SubstateC0WaitPercent",
			metricName: "cpu_c0_substate_c0_wait",
		},
		{
			name:       "Invalid",
			metricName: "",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metric := cpuMetricType(i)
			require.Equal(t, tc.metricName, metric.String())
		})
	}
}

func TestPackageMetric_String(t *testing.T) {
	testCases := []struct {
		name       string
		metricName string
	}{
		{
			name:       "PackageCurrentPowerConsumption",
			metricName: "current_power_consumption",
		},
		{
			name:       "PackageCurrentDramPowerConsumption",
			metricName: "current_dram_power_consumption",
		},
		{
			name:       "PackageThermalDesignPower",
			metricName: "thermal_design_power",
		},
		{
			name:       "PackageCPUBaseFrequency",
			metricName: "cpu_base_frequency",
		},
		{
			name:       "PackageUncoreFrequency",
			metricName: "uncore_frequency",
		},
		{
			name:       "PackageTurboLimit",
			metricName: "max_turbo_frequency",
		},
		{
			name:       "Invalid",
			metricName: "",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metric := packageMetricType(i)
			require.Equal(t, tc.metricName, metric.String())
		})
	}
}

func TestCPUMetricTypeFromString(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for m := cpuMetricType(0); m < cpuC0SubstateC0WaitPercent+1; m++ {
			val, err := cpuMetricTypeFromString(m.String())
			require.NoError(t, err)
			require.Equal(t, m, val)
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		val, err := cpuMetricTypeFromString("invalid")
		require.Error(t, err)
		require.Equal(t, cpuMetricType(-1), val)
	})
}

func TestPackageMetricTypeFromString(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for m := packageMetricType(0); m < packageTurboLimit+1; m++ {
			val, err := packageMetricTypeFromString(m.String())
			require.NoError(t, err)
			require.Equal(t, m, val)
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		val, err := packageMetricTypeFromString("invalid")
		require.Error(t, err)
		require.Equal(t, packageMetricType(-1), val)
	})
}
