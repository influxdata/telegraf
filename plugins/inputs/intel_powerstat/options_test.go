//go:build linux && amd64

package intel_powerstat

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	t.Run("NoCPUsSpecified", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			cpuMetrics: []string{
				cpuFrequency.String(),            // needs coreFreq
				cpuC0SubstateC01Percent.String(), // needs perf
			},
			packageMetrics: []string{
				packageCurrentPowerConsumption.String(), // needs rapl
				packageUncoreFrequency.String(),         // needs uncoreFreq and msr
			},
		})

		require.Len(t, opts, 5)
	})

	t.Run("ExcludedCPUs", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			excludedCPUs: []int{0, 1, 2, 3},
			cpuMetrics: []string{
				// needs msr
				cpuTemperature.String(),
			},
			packageMetrics: []string{
				// needs rapl
				packageCurrentPowerConsumption.String(),
			},
		})

		require.Len(t, opts, 3)
	})

	t.Run("IncludedCPUs", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			includedCPUs: []int{0, 1, 2, 3},
			cpuMetrics: []string{
				cpuFrequency.String(),               // needs coreFreq
				cpuC0SubstateC0WaitPercent.String(), // needs perf
			},
			packageMetrics: []string{
				packageTurboLimit.String(),                  // needs msr
				packageCurrentDramPowerConsumption.String(), // needs rapl
				packageUncoreFrequency.String(),             // needs uncoreFreq
			},
		})

		require.Len(t, opts, 6)
	})

	t.Run("WithMsrTimeout", func(t *testing.T) {
		g := &optGenerator{}
		opts := g.generate(optConfig{
			cpuMetrics: []string{
				cpuTemperature.String(),
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
			cpuMetrics: []string{
				cpuC7StateResidency.String(),
			},
			msrReadTimeout: 0, //timeout disabled
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
			cpuMetrics: []string{
				cpuC3StateResidency.String(),
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

func TestNeedsMsr(t *testing.T) {
	metrics := []string{
		cpuFrequency.String(),              // needs cpuFreq
		cpuC0SubstateC01Percent.String(),   // needs perf
		packageThermalDesignPower.String(), // needs rapl
		"",
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsMsr(metrics))
	})

	t.Run("True", func(t *testing.T) {
		t.Run("CPUBaseFreq", func(t *testing.T) {
			metrics[len(metrics)-1] = packageCPUBaseFrequency.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("CPUTemperature", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuTemperature.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("CPUC0StateResidency", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC0StateResidency.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("CPUC1StateResidency", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC1StateResidency.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("CPUC3StateResidency", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC3StateResidency.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("CPUC6StateResidency", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC6StateResidency.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("CPUC7StateResidency", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC7StateResidency.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("CPUBusyCycles", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuBusyCycles.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("CPUBusyFrequency", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuBusyFrequency.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("PackageTurboLimit", func(t *testing.T) {
			metrics[len(metrics)-1] = packageTurboLimit.String()
			require.True(t, needsMsr(metrics))
		})

		t.Run("PackageUncoreFrequency", func(t *testing.T) {
			metrics[len(metrics)-1] = packageUncoreFrequency.String()
			require.True(t, needsMsr(metrics))
		})
	})
}

func TestNeedsRapl(t *testing.T) {
	metrics := []string{
		cpuFrequency.String(),            // needs cpuFreq
		packageCPUBaseFrequency.String(), // needs msr
		cpuC0SubstateC01Percent.String(), // needs perf
		packageUncoreFrequency.String(),  // needs uncoreFreq
		"",
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsRapl(metrics))
	})

	t.Run("True", func(t *testing.T) {
		t.Run("PackageCurrentPowerConsumption", func(t *testing.T) {
			metrics[len(metrics)-1] = packageCurrentPowerConsumption.String()
			require.True(t, needsRapl(metrics))
		})

		t.Run("PackageCurrentDramPowerConsumption", func(t *testing.T) {
			metrics[len(metrics)-1] = packageCurrentDramPowerConsumption.String()
			require.True(t, needsRapl(metrics))
		})

		t.Run("PackageThermalDesignPower", func(t *testing.T) {
			metrics[len(metrics)-1] = packageThermalDesignPower.String()
			require.True(t, needsRapl(metrics))
		})
	})
}

func TestNeedsCoreFreq(t *testing.T) {
	metrics := []string{
		cpuTemperature.String(),            // needs msr
		cpuC0SubstateC01Percent.String(),   // needs perf
		packageThermalDesignPower.String(), // needs rapl
		packageUncoreFrequency.String(),    // needs uncoreFreq
		"",
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsCoreFreq(metrics))
	})

	t.Run("True", func(t *testing.T) {
		metrics[len(metrics)-1] = cpuFrequency.String()
		require.True(t, needsCoreFreq(metrics))
	})
}

func TestNeedsUncoreFreq(t *testing.T) {
	metrics := []string{
		cpuFrequency.String(),              // needs cpuFreq
		packageCPUBaseFrequency.String(),   // needs msr
		cpuC0SubstateC01Percent.String(),   // needs perf
		packageThermalDesignPower.String(), // needs rapl
		"",
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsUncoreFreq(metrics))
	})

	t.Run("True", func(t *testing.T) {
		metrics[len(metrics)-1] = packageUncoreFrequency.String()
		require.True(t, needsUncoreFreq(metrics))
	})
}

func TestNeedsPerf(t *testing.T) {
	metrics := []string{
		cpuFrequency.String(),              // needs cpuFreq
		packageCPUBaseFrequency.String(),   // needs msr
		packageThermalDesignPower.String(), // needs rapl
		packageUncoreFrequency.String(),    // needs uncoreFreq
		"",
	}

	t.Run("False", func(t *testing.T) {
		require.False(t, needsPerf(metrics))
	})

	t.Run("True", func(t *testing.T) {
		t.Run("CPUC0SubstateC01Percent", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC0SubstateC01Percent.String()
			require.True(t, needsPerf(metrics))
		})

		t.Run("CPUC0SubstateC02Percent", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC0SubstateC02Percent.String()
			require.True(t, needsPerf(metrics))
		})

		t.Run("CPUC0SubstateC0WaitPercent", func(t *testing.T) {
			metrics[len(metrics)-1] = cpuC0SubstateC0WaitPercent.String()
			require.True(t, needsPerf(metrics))
		})
	})
}
