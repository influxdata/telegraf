//go:build linux && amd64

package intel_powerstat

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestGenerate(t *testing.T) {
	t.Run("NoCPUsSpecified", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			cpuMetrics: []cpuMetricType{
				cpuFrequency,            // needs coreFreq
				cpuC0SubstateC01Percent, // needs perf
			},
			packageMetrics: []packageMetricType{
				packageCurrentPowerConsumption, // needs rapl
				packageUncoreFrequency,         // needs uncoreFreq and msr
			},
		})

		require.Len(t, opts, 5)
	})

	t.Run("ExcludedCPUs", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			excludedCPUs: []int{0, 1, 2, 3},
			cpuMetrics: []cpuMetricType{
				// needs msr
				cpuTemperature,
			},
			packageMetrics: []packageMetricType{
				// needs rapl
				packageCurrentPowerConsumption,
			},
		})

		require.Len(t, opts, 3)
	})

	t.Run("IncludedCPUs", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			includedCPUs: []int{0, 1, 2, 3},
			cpuMetrics: []cpuMetricType{
				cpuFrequency,               // needs coreFreq
				cpuC0SubstateC0WaitPercent, // needs perf
			},
			packageMetrics: []packageMetricType{
				packageTurboLimit,                  // needs msr
				packageCurrentDramPowerConsumption, // needs rapl
				packageUncoreFrequency,             // needs uncoreFreq
			},
		})

		require.Len(t, opts, 6)
	})

	t.Run("WithMsrTimeout", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			cpuMetrics: []cpuMetricType{
				cpuTemperature,
			},
			msrReadTimeout: time.Second,
		})

		require.Len(t, opts, 1)

		withMsrTimeoutUsed := false
		for _, opt := range opts {
			if strings.Contains(runtime.FuncForPC(reflect.ValueOf(opt).Pointer()).Name(), ".WithMsrTimeout.") {
				withMsrTimeoutUsed = true
				continue
			}
		}
		require.True(t, withMsrTimeoutUsed, "WithMsrTimeout wasn't included in the generated options")
	})

	t.Run("WithMsr", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			cpuMetrics: []cpuMetricType{
				cpuC7StateResidency,
			},
			msrReadTimeout: 0, // timeout disabled
		})

		require.Len(t, opts, 1)

		withMsrUsed := false
		for _, opt := range opts {
			if strings.Contains(runtime.FuncForPC(reflect.ValueOf(opt).Pointer()).Name(), ".WithMsr.") {
				withMsrUsed = true
				continue
			}
		}
		require.True(t, withMsrUsed, "WithMsr wasn't included in the generated options")
	})

	t.Run("WithLogger", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			cpuMetrics: []cpuMetricType{
				cpuC3StateResidency,
			},
			log: &testutil.Logger{},
		})

		require.Len(t, opts, 2)

		withLoggerUsed := false
		for _, opt := range opts {
			if strings.Contains(runtime.FuncForPC(reflect.ValueOf(opt).Pointer()).Name(), ".WithLogger.") {
				withLoggerUsed = true
				continue
			}
		}
		require.True(t, withLoggerUsed, "WithLogger wasn't included in the generated options")
	})
}

func TestNeedsMsrPackage(t *testing.T) {
	packageMetrics := []packageMetricType{
		packageThermalDesignPower,          // needs rapl
		packageCurrentDramPowerConsumption, // needs rapl
		packageMetricType(420),
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsMsrPackage(packageMetrics))
	})

	t.Run("True", func(t *testing.T) {
		t.Run("CPUBaseFreq", func(t *testing.T) {
			packageMetrics[len(packageMetrics)-1] = packageCPUBaseFrequency
			require.True(t, needsMsrPackage(packageMetrics))
		})

		t.Run("PackageTurboLimit", func(t *testing.T) {
			packageMetrics[len(packageMetrics)-1] = packageTurboLimit
			require.True(t, needsMsrPackage(packageMetrics))
		})

		t.Run("PackageUncoreFrequency", func(t *testing.T) {
			packageMetrics[len(packageMetrics)-1] = packageUncoreFrequency
			require.True(t, needsMsrPackage(packageMetrics))
		})
	})
}

func TestNeedsMsrCPU(t *testing.T) {
	cpuMetrics := []cpuMetricType{
		cpuFrequency,            // needs cpuFreq
		cpuC0SubstateC01Percent, // needs perf
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsMsrCPU(cpuMetrics))
	})

	t.Run("True", func(t *testing.T) {
		t.Run("CPUTemperature", func(t *testing.T) {
			cpuMetrics[len(cpuMetrics)-1] = cpuTemperature
			require.True(t, needsMsrCPU(cpuMetrics))
		})

		t.Run("CPUC0StateResidency", func(t *testing.T) {
			cpuMetrics[len(cpuMetrics)-1] = cpuC0StateResidency
			require.True(t, needsMsrCPU(cpuMetrics))
		})

		t.Run("CPUC1StateResidency", func(t *testing.T) {
			cpuMetrics[len(cpuMetrics)-1] = cpuC1StateResidency
			require.True(t, needsMsrCPU(cpuMetrics))
		})

		t.Run("CPUC3StateResidency", func(t *testing.T) {
			cpuMetrics[len(cpuMetrics)-1] = cpuC3StateResidency
			require.True(t, needsMsrCPU(cpuMetrics))
		})

		t.Run("CPUC6StateResidency", func(t *testing.T) {
			cpuMetrics[len(cpuMetrics)-1] = cpuC6StateResidency
			require.True(t, needsMsrCPU(cpuMetrics))
		})

		t.Run("CPUC7StateResidency", func(t *testing.T) {
			cpuMetrics[len(cpuMetrics)-1] = cpuC7StateResidency
			require.True(t, needsMsrCPU(cpuMetrics))
		})

		t.Run("CPUBusyCycles", func(t *testing.T) {
			cpuMetrics[len(cpuMetrics)-1] = cpuBusyCycles
			require.True(t, needsMsrCPU(cpuMetrics))
		})

		t.Run("CPUBusyFrequency", func(t *testing.T) {
			cpuMetrics[len(cpuMetrics)-1] = cpuBusyFrequency
			require.True(t, needsMsrCPU(cpuMetrics))
		})
	})
}

func TestNeedsRapl(t *testing.T) {
	metrics := []packageMetricType{
		packageCPUBaseFrequency, // needs msr
		packageUncoreFrequency,  // needs uncoreFreq
		packageMetricType(420),
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsRapl(metrics))
	})

	t.Run("True", func(t *testing.T) {
		t.Run("PackageCurrentPowerConsumption", func(t *testing.T) {
			metrics[len(metrics)-1] = packageCurrentPowerConsumption
			require.True(t, needsRapl(metrics))
		})

		t.Run("PackageCurrentDramPowerConsumption", func(t *testing.T) {
			metrics[len(metrics)-1] = packageCurrentDramPowerConsumption
			require.True(t, needsRapl(metrics))
		})

		t.Run("PackageThermalDesignPower", func(t *testing.T) {
			metrics[len(metrics)-1] = packageThermalDesignPower
			require.True(t, needsRapl(metrics))
		})
	})
}

func TestNeedsCoreFreq(t *testing.T) {
	metrics := []cpuMetricType{
		cpuTemperature,          // needs msr
		cpuC1StateResidency,     // needs msr
		cpuC0SubstateC01Percent, // needs perf
		cpuMetricType(420),
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsCoreFreq(metrics))
	})

	t.Run("True", func(t *testing.T) {
		metrics[len(metrics)-1] = cpuFrequency
		require.True(t, needsCoreFreq(metrics))
	})
}

func TestNeedsUncoreFreq(t *testing.T) {
	metrics := []packageMetricType{
		packageCPUBaseFrequency,   // needs msr
		packageThermalDesignPower, // needs rapl
		packageMetricType(420),
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsUncoreFreq(metrics))
	})

	t.Run("True", func(t *testing.T) {
		metrics[len(metrics)-1] = packageUncoreFrequency
		require.True(t, needsUncoreFreq(metrics))
	})
}

func TestNeedsPerf(t *testing.T) {
	metrics := []cpuMetricType{
		cpuFrequency,        // needs cpuFreq
		cpuC1StateResidency, // needs msr
		cpuMetricType(420),
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsPerf(metrics))
	})

	t.Run("True", func(t *testing.T) {
		t.Run("CPUC0SubstateC01Percent", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC0SubstateC01Percent
			require.True(t, needsPerf(metrics))
		})

		t.Run("CPUC0SubstateC02Percent", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC0SubstateC02Percent
			require.True(t, needsPerf(metrics))
		})

		t.Run("CPUC0SubstateC0WaitPercent", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC0SubstateC0WaitPercent
			require.True(t, needsPerf(metrics))
		})
	})
}
