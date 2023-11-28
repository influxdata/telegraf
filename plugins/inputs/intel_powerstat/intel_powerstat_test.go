//go:build linux && amd64

package intel_powerstat

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	ptel "github.com/intel/powertelemetry"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

type parseMetricTestCase struct {
	name    string
	metrics []string
	parsed  []string
	err     error
}

func TestParsePackageMetrics(t *testing.T) {
	testCases := []parseMetricTestCase{
		{
			name:    "NilSlice",
			metrics: nil,
			parsed: []string{
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},
		},
		{
			name:    "EmptySlice",
			metrics: []string{},
			parsed:  []string{},
		},
		{
			name: "UnsupportedMetric",
			metrics: []string{
				packageCurrentPowerConsumption.String(),
				"current_energy_consumption", // unsupported metric
				packageCurrentDramPowerConsumption.String(),
			},
			parsed: nil,
			err:    errors.New("invalid package metric specified: \"current_energy_consumption\""),
		},
		{
			name: "InvalidPackageMetric",
			metrics: []string{
				packageCurrentPowerConsumption.String(),
				cpuTemperature.String(), // not a package metric
				packageCurrentDramPowerConsumption.String(),
			},
			parsed: nil,
			err:    fmt.Errorf("invalid package metric specified: %q", cpuTemperature.String()),
		},
		{
			name: "HasDuplicates",
			metrics: []string{
				packageCurrentPowerConsumption.String(),
				packageThermalDesignPower.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(), // duplicate
			},
			parsed: nil,
			err:    errors.New("package metrics contains duplicates"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedOut, err := parsePackageMetrics(tc.metrics)

			require.Equal(t, tc.parsed, parsedOut)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseCPUMetrics(t *testing.T) {
	testCases := []parseMetricTestCase{
		{
			name:    "NilSlice",
			metrics: nil,
			parsed:  nil,
		},
		{
			name:    "EmptySlice",
			metrics: []string{},
			parsed:  []string{},
		},
		{
			name: "UnsupportedMetric",
			metrics: []string{
				cpuC0StateResidency.String(),
				"cpu_c9_state_residency", // unsupported metric
				cpuC1StateResidency.String(),
				cpuTemperature.String(),
			},
			parsed: nil,
			err:    errors.New("invalid core metric specified: \"cpu_c9_state_residency\""),
		},
		{
			name: "InvalidCoreMetric",
			metrics: []string{
				cpuC0StateResidency.String(),
				packageTurboLimit.String(), // not a core metric
				cpuC1StateResidency.String(),
				cpuTemperature.String(),
			},
			parsed: nil,
			err:    fmt.Errorf("invalid core metric specified: %q", packageTurboLimit.String()),
		},
		{
			name: "HasDuplicates",
			metrics: []string{
				cpuC0StateResidency.String(),
				cpuC1StateResidency.String(),
				cpuTemperature.String(),
				cpuC0StateResidency.String(), // duplicate
			},
			parsed: nil,
			err:    errors.New("core metrics contains duplicates"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedOut, err := parseCPUMetrics(tc.metrics)

			require.Equal(t, tc.parsed, parsedOut)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseCPUTimeRelatedMsrMetrics(t *testing.T) {
	testCases := []parseMetricTestCase{
		{
			name:    "EmptySlice",
			metrics: []string{},
			parsed:  []string{},
		},
		{
			name: "NotFound",
			metrics: []string{
				// Metric not relying on MSR.
				cpuFrequency.String(),

				// Metric relying on single MSR read.
				cpuTemperature.String(),

				// Metrics relying on perf events.
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
				cpuC0SubstateC0WaitPercent.String(),
			},
			parsed: []string{},
		},
		{
			name: "Found",
			metrics: []string{
				// Metric not relying on MSR.
				cpuFrequency.String(),

				// Metric relying on single MSR read.
				cpuTemperature.String(),

				// Metrics relying on time-related MSR offset reads.
				cpuC0StateResidency.String(),
				cpuC1StateResidency.String(),
				cpuC3StateResidency.String(),
				cpuC6StateResidency.String(),
				cpuC7StateResidency.String(),
				cpuBusyCycles.String(),
				cpuBusyFrequency.String(),

				// Metrics relying on perf events.
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
			},
			parsed: []string{
				cpuC0StateResidency.String(),
				cpuC1StateResidency.String(),
				cpuC3StateResidency.String(),
				cpuC6StateResidency.String(),
				cpuC7StateResidency.String(),
				cpuBusyCycles.String(),
				cpuBusyFrequency.String(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedOut := parseCPUTimeRelatedMsrMetrics(tc.metrics)
			require.Equal(t, tc.parsed, parsedOut)
		})
	}
}

func TestParseCPUPerfMetrics(t *testing.T) {
	testCases := []parseMetricTestCase{
		{
			name:    "EmptySlice",
			metrics: []string{},
			parsed:  []string{},
		},
		{
			name: "NotFound",
			metrics: []string{
				// Metric not relying on MSR.
				cpuFrequency.String(),

				// Metric relying on single MSR read.
				cpuTemperature.String(),

				// Metrics relying on time-related MSR offset reads.
				cpuC3StateResidency.String(),
				cpuC6StateResidency.String(),
				cpuBusyFrequency.String(),
			},
			parsed: []string{},
		},
		{
			name: "Found",
			metrics: []string{
				// Metric not relying on MSR.
				cpuFrequency.String(),

				// Metric relying on single MSR read.
				cpuTemperature.String(),

				// Metrics relying on perf events.
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
				cpuC0SubstateC0WaitPercent.String(),

				// Metrics relying on time-related MSR offset reads.
				cpuC3StateResidency.String(),
				cpuC6StateResidency.String(),
				cpuBusyFrequency.String(),
			},
			parsed: []string{
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
				cpuC0SubstateC0WaitPercent.String(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedOut := parseCPUPerfMetrics(tc.metrics)
			require.Equal(t, tc.parsed, parsedOut)
		})
	}
}

func TestParsePackageRaplMetrics(t *testing.T) {
	testCases := []parseMetricTestCase{
		{
			name:    "EmptySlice",
			metrics: []string{},
			parsed:  []string{},
		},
		{
			name: "NotFound",
			metrics: []string{
				// Metrics not relying on rapl.
				packageTurboLimit.String(),
				packageCPUBaseFrequency.String(),
				packageUncoreFrequency.String(),
			},
			parsed: []string{},
		},
		{
			name: "Found",
			metrics: []string{
				// Metrics not relying on rapl.
				packageTurboLimit.String(),
				packageCPUBaseFrequency.String(),
				packageUncoreFrequency.String(),

				// Metrics relying on rapl.
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},
			parsed: []string{
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedOut := parsePackageRaplMetrics(tc.metrics)
			require.Equal(t, tc.parsed, parsedOut)
		})
	}
}

func TestParsePackageMsrMetrics(t *testing.T) {
	testCases := []parseMetricTestCase{
		{
			name:    "EmptySlice",
			metrics: []string{},
			parsed:  []string{},
		},
		{
			name: "NotFound",
			metrics: []string{
				// Metrics not relying on msr.
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},
			parsed: []string{},
		},
		{
			name: "Found",
			metrics: []string{
				// Metrics not relying on msr.
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),

				// Metrics relying uniquely on msr.
				packageTurboLimit.String(),
				packageCPUBaseFrequency.String(),
			},
			parsed: []string{
				packageTurboLimit.String(),
				packageCPUBaseFrequency.String(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedOut := parsePackageMsrMetrics(tc.metrics)
			require.Equal(t, tc.parsed, parsedOut)
		})
	}
}

func TestParseCoreRange(t *testing.T) {
	testCases := []struct {
		name      string
		coreRange string
		cores     []int
		err       error
	}{
		{
			name:      "InvalidFormat",
			coreRange: "1,3",
			cores:     nil,
			err:       errors.New("invalid core range format"),
		},
		{
			name:      "LowerBoundNonNumeric",
			coreRange: "a-10",
			cores:     nil,
			err:       errors.New("failed to parse low bounds' core range"),
		},
		{
			name:      "MissingLowerBound",
			coreRange: "-10",
			cores:     nil,
			err:       errors.New("failed to parse low bounds' core range"),
		},
		{
			name:      "HigherBoundNonNumeric",
			coreRange: "0-a",
			cores:     nil,
			err:       errors.New("failed to parse high bounds' core range"),
		},
		{
			name:      "MissingHigherBound",
			coreRange: "0-",
			cores:     nil,
			err:       errors.New("failed to parse high bounds' core range"),
		},
		{
			name:      "InvalidBounds",
			coreRange: "10-1",
			cores:     nil,
			err:       errors.New("high bound of core range cannot be less than low bound"),
		},
		{
			name:      "SingleCore",
			coreRange: "1-1",
			cores:     []int{1},
		},
		{
			name:      "CoreRange",
			coreRange: "5-10",
			cores:     []int{5, 6, 7, 8, 9, 10},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coresOut, err := parseCoreRange(tc.coreRange)

			require.Equal(t, tc.cores, coresOut)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseGroupCores(t *testing.T) {
	testCases := []struct {
		name      string
		coreGroup string
		cores     []int
		err       error
	}{
		{
			name:      "FailedToParseCoreRange",
			coreGroup: "1-a,7,9,11",
			cores:     nil,
			err:       errors.New("failed to parse core range \"1-a\""),
		},
		{
			name:      "FailedToParseSingleCore",
			coreGroup: "1-5,7,b,11",
			cores:     nil,
			err:       errors.New("failed to parse single core"),
		},
		{
			name:      "Ok",
			coreGroup: "1-5,7,9,11",
			cores:     []int{1, 2, 3, 4, 5, 7, 9, 11},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coresOut, err := parseGroupCores(tc.coreGroup)

			require.Equal(t, tc.cores, coresOut)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHasDuplicate(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		t.Run("False", func(t *testing.T) {
			nums := []int{0, 1, 2, 3}
			require.False(t, hasDuplicate(nums))
		})

		t.Run("True", func(t *testing.T) {
			nums := []uint32{0, 1, 2, 3, 4, 5, 6, 1}
			require.True(t, hasDuplicate(nums))
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Run("False", func(t *testing.T) {
			strs := []string{"1", "2", "3", "4"}
			require.False(t, hasDuplicate(strs))
		})

		t.Run("True", func(t *testing.T) {
			strs := []string{"1", "2", "3", "1"}
			require.True(t, hasDuplicate(strs))
		})
	})
}

func TestParseCores(t *testing.T) {
	testCases := []struct {
		name       string
		coreGroups []string
		cores      []int
		err        error
	}{
		{
			name:       "InvalidCoreGroup",
			coreGroups: []string{"1-4,11", "10-b"},
			cores:      nil,
			err:        errors.New("failed to parse core group"),
		},
		{
			name:       "FoundDuplicates",
			coreGroups: []string{"1-4,11", "10-12"},
			cores:      nil,
			err:        errors.New("core values cannot be duplicated"),
		},
		{
			name:       "CoresIsNil",
			coreGroups: nil,
			cores:      []int{},
		},
		{
			name:       "CoresIsEmpty",
			coreGroups: []string{},
			cores:      []int{},
		},
		{
			name:       "Ok",
			coreGroups: []string{"1-4,6", "8", "10-12", "15,20"},
			cores:      []int{1, 2, 3, 4, 6, 8, 10, 11, 12, 15, 20},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coresOut, err := parseCores(tc.coreGroups)

			require.Equal(t, tc.cores, coresOut)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	t.Run("BothCPUOptionsProvided", func(t *testing.T) {
		p := &PowerStat{
			IncludedCPUs: []string{"0-10,20-22"},
			ExcludedCPUs: []string{"0"},
		}

		require.ErrorContains(t, p.parseConfig(), "configuration error. Provide either one 'included_cpus'/'excluded_cpus' configuration option, or none")
	})

	t.Run("FailedToParseIncludedCPUs", func(t *testing.T) {
		p := &PowerStat{
			// has duplicates
			IncludedCPUs: []string{"1-4,11", "10-12"},
		}

		require.ErrorContains(t, p.parseConfig(), "failed to parse included CPUs")
	})

	t.Run("FailedToParseExcludedCPUs", func(t *testing.T) {
		p := &PowerStat{
			// has non-numeric CPU ID
			ExcludedCPUs: []string{"1-4,b"},
		}

		require.ErrorContains(t, p.parseConfig(), "failed to parse excluded CPUs")
	})

	t.Run("FailedToParsePackageMetrics", func(t *testing.T) {
		p := &PowerStat{
			// has invalid metric
			PackageMetrics: []string{
				"invalid",
			},
		}

		require.ErrorContains(t, p.parseConfig(), "failed to parse package metrics")
	})

	t.Run("FailedToParseCPUMetrics", func(t *testing.T) {
		p := &PowerStat{
			// has duplicates
			CPUMetrics: []string{
				cpuFrequency.String(),
				cpuTemperature.String(),
				cpuFrequency.String(), // duplicate
			},
		}

		require.ErrorContains(t, p.parseConfig(), "failed to parse core metrics")
	})

	t.Run("EventDefinitionsNotProvidedForPerf", func(t *testing.T) {
		p := &PowerStat{
			// has duplicates
			CPUMetrics: []string{
				cpuC0SubstateC01Percent.String(),
			},
		}

		require.ErrorContains(t, p.parseConfig(), "failed to parse event definitions path: file path is empty")
	})

	t.Run("EventDefinitionsDoesNotExist", func(t *testing.T) {
		p := &PowerStat{
			// has duplicates
			CPUMetrics: []string{
				cpuC0SubstateC02Percent.String(),
			},
			EventDefinitions: "./testdata/doesNotExist.json",
		}

		require.ErrorContains(t, p.parseConfig(), "failed to parse event definitions path: file \"./testdata/doesNotExist.json\" does not exist")
	})

	t.Run("NoMetricsProvided", func(t *testing.T) {
		p := &PowerStat{
			// Disable default package metrics.
			PackageMetrics: []string{},
		}

		require.ErrorContains(t, p.parseConfig(), "no metrics were found in the configuration file")
	})

	t.Run("DisablePackageMetrics", func(t *testing.T) {
		p := &PowerStat{
			CPUMetrics: []string{
				cpuBusyFrequency.String(),
			},
			// Disable default package metrics.
			PackageMetrics: []string{},
		}

		require.NoError(t, p.parseConfig())
		require.Empty(t, p.PackageMetrics)
		require.Len(t, p.CPUMetrics, 1)
	})

	t.Run("DefaultPackageMetrics", func(t *testing.T) {
		p := &PowerStat{
			PackageMetrics: nil, // default package metrics
		}

		require.NoError(t, p.parseConfig())
		require.Equal(t,
			[]string{
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},
			p.PackageMetrics)
	})

	t.Run("IncludedCPUs", func(t *testing.T) {
		p := &PowerStat{
			IncludedCPUs: []string{"0-5"},
		}

		require.NoError(t, p.parseConfig())
		require.Equal(t, []int{0, 1, 2, 3, 4, 5}, p.parsedIncludedCores)
		require.Nil(t, p.parsedExcludedCores)
	})

	t.Run("ExcludedCPUs", func(t *testing.T) {
		p := &PowerStat{
			ExcludedCPUs: []string{"2-6", "8", "10"},
		}

		require.NoError(t, p.parseConfig())
		require.Equal(t, []int{2, 3, 4, 5, 6, 8, 10}, p.parsedExcludedCores)
		require.Nil(t, p.parsedIncludedCores)
	})

	t.Run("MetricsWithIncludedCPUs", func(t *testing.T) {
		p := &PowerStat{
			IncludedCPUs: []string{"0-3,6"},
			CPUMetrics: []string{
				cpuC0StateResidency.String(),
				cpuC1StateResidency.String(),
				cpuC3StateResidency.String(),
			},
			PackageMetrics: []string{
				packageUncoreFrequency.String(),
				packageTurboLimit.String(),
			},
		}

		require.NoError(t, p.parseConfig())
		require.Equal(t, []int{0, 1, 2, 3, 6}, p.parsedIncludedCores)

		// Check flags
		require.True(t, p.needsMsrCPU)
		require.True(t, p.needsTimeRelatedMsr)
		require.False(t, p.needsCoreFreq)
		require.False(t, p.needsPerf)
	})
}

type mockOptGenerator struct {
	mock.Mock
}

func (m *mockOptGenerator) Generate(cfg OptConfig) []ptel.Option {
	args := m.Called(cfg)
	return args.Get(0).([]ptel.Option)
}

func TestSampleConfig(t *testing.T) {
	p := &PowerStat{}
	require.NotZero(t, p.SampleConfig())
}

func TestStart(t *testing.T) {
	t.Run("FailedToDisableUnsupportedMetrics", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata/cpu_model_missing")

		acc := &testutil.Accumulator{}

		p := &PowerStat{}

		require.ErrorContains(t, p.Start(acc), "error occurred while parsing CPU model")
	})

	t.Run("FailedToParseConfigWithDuplicates", func(t *testing.T) {
		acc := &testutil.Accumulator{}
		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			// has duplicates
			IncludedCPUs: []string{"1-4,11", "10-12"},
			Log:          logger,
		}

		require.ErrorContains(t, p.Start(acc), "failed to parse included CPUs")
		require.Empty(t, logger.Warnings())
	})

	t.Run("FailedToParseConfigWithNegativeTimeout", func(t *testing.T) {
		acc := &testutil.Accumulator{}
		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			// negative value
			MsrReadTimeout: -2,
			Log:            logger,
		}

		require.ErrorContains(t, p.Start(acc), "msr_read_timeout should be positive number or equal to 0 (to disable timeouts)")
		require.Empty(t, logger.Warnings())
	})

	t.Run("FailedToCreateMetricFetcher", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata")

		acc := &testutil.Accumulator{}
		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			// has CPU ID out of bounds
			IncludedCPUs: []string{"0-9"},
			Log:          logger,

			option: &optGenerator{},
		}

		require.ErrorContains(t, p.Start(acc), "failed to initialize metric fetcher interface")
		require.Empty(t, logger.Warnings())
	})

	t.Run("WithWarning", func(t *testing.T) {
		acc := &testutil.Accumulator{}
		logger := &testutil.CaptureLogger{}

		mOptGenerator := &mockOptGenerator{}
		mOptGenerator.On("Generate", mock.AnythingOfType("OptConfig")).Return(
			[]ptel.Option{
				ptel.WithRapl("/dummy/path"),
			},
		)

		p := &PowerStat{
			IncludedCPUs: []string{"0"},
			PackageMetrics: []string{
				packageCurrentPowerConsumption.String(), // needs rapl
			},
			Log: logger,

			option: mOptGenerator,
		}

		require.NoError(t, p.Start(acc))
		require.Len(t, logger.Warnings(), 1)
		require.Contains(t, logger.Warnings()[0], "Plugin initialized with errors")
	})
}

func TestGather(t *testing.T) {
	t.Run("WithoutMetrics", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		p := &PowerStat{
			PackageMetrics: []string{},
		}

		require.NoError(t, p.Gather(acc))

		require.Empty(t, acc.Errors)
		require.Empty(t, acc.GetTelegrafMetrics())
	})

	t.Run("WithDefaultPackageMetrics", func(t *testing.T) {
		packageID := 0

		packagePower := 10.0
		dramPower := 5.0
		thermalDesignPower := 20.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available package IDs.
		mFetcher.On("GetPackageIDs").Return([]int{packageID}).Once()

		// mock getting current package power consumption metric.
		mFetcher.On("GetCurrentPackagePowerConsumptionWatts", packageID).Return(packagePower, nil).Once()

		// mock getting current dram power consumption metric.
		mFetcher.On("GetCurrentDramPowerConsumptionWatts", packageID).Return(dramPower, nil).Once()

		// mock getting package thermal design power metric.
		mFetcher.On("GetPackageThermalDesignPowerWatts", packageID).Return(thermalDesignPower, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		require.NoError(t, p.Gather(acc))

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 3)
		require.True(t, acc.HasField("powerstat_package", "current_power_consumption_watts"))
		require.True(t, acc.HasField("powerstat_package", "current_dram_power_consumption_watts"))
		require.True(t, acc.HasField("powerstat_package", "thermal_design_power_watts"))
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithPackageMetrics", func(t *testing.T) {
		packageIDs := []int{0, 1, 2, 3}

		baseFreq := uint64(200)
		packagePower := 30.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available package IDs.
		mFetcher.On("GetPackageIDs").Return(packageIDs).Once()

		// mock getting CPU base frequency metric.
		mFetcher.On("GetCPUBaseFrequency", mock.AnythingOfType("int")).Return(baseFreq, nil).Times(len(packageIDs))

		// mock getting current package power consumption metric.
		mFetcher.On("GetCurrentPackagePowerConsumptionWatts", mock.AnythingOfType("int")).Return(packagePower, nil).Times(len(packageIDs))

		p := &PowerStat{
			PackageMetrics: []string{
				packageCurrentPowerConsumption.String(),
				packageCPUBaseFrequency.String(),
			},
			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		require.NoError(t, p.Gather(acc))

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 8)
		require.True(t, acc.HasField("powerstat_package", "cpu_base_frequency_mhz"))
		require.True(t, acc.HasField("powerstat_package", "current_power_consumption_watts"))
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithCPUMetrics", func(t *testing.T) {
		cpuIDs := []int{0, 1, 2, 3}

		cpuFreq := 123.5
		cpuTemp := uint64(20)
		cpuBusyFreq := 456.7

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available CPU IDs with access to msr registers and coreFreq.
		mFetcher.On("GetMsrCPUIDs").Return(cpuIDs).Once()

		// mock getting core ID for CPU IDs.
		mFetcher.On("GetCPUCoreID", mock.AnythingOfType("int")).Return(1, nil).Times(len(cpuIDs))

		// mock getting package ID for CPU IDs.
		mFetcher.On("GetCPUPackageID", mock.AnythingOfType("int")).Return(1, nil).Times(len(cpuIDs))

		// mock updating msr time-related metrics for CPU IDs.
		mFetcher.On("UpdatePerCPUMetrics", mock.AnythingOfType("int")).Return(nil).Times(len(cpuIDs))

		// mock getting CPU frequency for CPU IDs.
		mFetcher.On("GetCPUFrequency", mock.AnythingOfType("int")).Return(cpuFreq, nil).Times(len(cpuIDs))

		// mock getting CPU temperature metric for CPU IDs.
		mFetcher.On("GetCPUTemperature", mock.AnythingOfType("int")).Return(cpuTemp, nil).Times(len(cpuIDs))

		// mock getting CPU busy frequency metric for CPU IDs.
		mFetcher.On("GetCPUBusyFrequencyMhz", mock.AnythingOfType("int")).Return(cpuBusyFreq, nil).Times(len(cpuIDs))

		p := &PowerStat{
			// Disables package metrics
			PackageMetrics: []string{},
			CPUMetrics: []string{
				cpuFrequency.String(),
				cpuTemperature.String(),
				cpuBusyFrequency.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		require.NoError(t, p.Gather(acc))

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 12)
		require.True(t, acc.HasField("powerstat_core", "cpu_frequency_mhz"))
		require.True(t, acc.HasField("powerstat_core", "cpu_temperature_celsius"))
		require.True(t, acc.HasField("powerstat_core", "cpu_busy_frequency_mhz"))
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithPerfMetrics", func(t *testing.T) {
		cpuIDs := []int{0, 1, 2}

		c01 := 0.1
		c02 := 1.2
		c0Wait := 2.3

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock reading perf events.
		mFetcher.On("ReadPerfEvents").Return(nil).Once()

		// mock getting available CPU IDs with access to msr registers and coreFreq.
		mFetcher.On("GetPerfCPUIDs").Return(cpuIDs).Once()

		// mock getting core ID for CPU IDs.
		mFetcher.On("GetCPUCoreID", mock.AnythingOfType("int")).Return(1, nil).Times(len(cpuIDs))

		// mock getting package ID for CPU IDs.
		mFetcher.On("GetCPUPackageID", mock.AnythingOfType("int")).Return(1, nil).Times(len(cpuIDs))

		// mock getting CPU C01 metric.
		mFetcher.On("GetCPUC0SubstateC01Percent", mock.AnythingOfType("int")).Return(c01, nil).Times(len(cpuIDs))

		// mock getting CPU C02 metric.
		mFetcher.On("GetCPUC0SubstateC02Percent", mock.AnythingOfType("int")).Return(c02, nil).Times(len(cpuIDs))

		// mock getting CPU C0Wait metric.
		mFetcher.On("GetCPUC0SubstateC0WaitPercent", mock.AnythingOfType("int")).Return(c0Wait, nil).Times(len(cpuIDs))

		p := &PowerStat{
			// Disables package metrics
			PackageMetrics: []string{},
			CPUMetrics: []string{
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
				cpuC0SubstateC0WaitPercent.String(),
			},
			EventDefinitions: "./testdata/sapphirerapids_core.json",

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		require.NoError(t, p.Gather(acc))

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 9)
		require.True(t, acc.HasField("powerstat_core", "cpu_c0_substate_c01_percent"))
		require.True(t, acc.HasField("powerstat_core", "cpu_c0_substate_c02_percent"))
		require.True(t, acc.HasField("powerstat_core", "cpu_c0_substate_c0_wait_percent"))
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithPerfAndMsrCPUMetrics", func(t *testing.T) {
		cpuIDsMsr := []int{0, 1, 2, 3}
		cpuIDsPerf := []int{0, 1}

		c1 := 0.5
		c6 := 1.5
		c01 := 0.1
		c02 := 1.2

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available CPU IDs with access to msr registers.
		mFetcher.On("GetMsrCPUIDs").Return(cpuIDsMsr).Once()

		// mock getting core ID for CPU IDs.
		mFetcher.On("GetCPUCoreID", mock.AnythingOfType("int")).Return(1, nil).Times(len(cpuIDsMsr) + len(cpuIDsPerf))

		// mock getting package ID for CPU IDs.
		mFetcher.On("GetCPUPackageID", mock.AnythingOfType("int")).Return(1, nil).Times(len(cpuIDsMsr) + len(cpuIDsPerf))

		// mock updating msr time-related metrics for CPU IDs.
		mFetcher.On("UpdatePerCPUMetrics", mock.AnythingOfType("int")).Return(nil).Times(len(cpuIDsMsr))

		// mock getting CPU C1 state residency metric for CPU IDs.
		mFetcher.On("GetCPUC1StateResidency", mock.AnythingOfType("int")).Return(c1, nil).Times(len(cpuIDsMsr))

		// mock getting CPU C6 state residency metric for CPU IDs.
		mFetcher.On("GetCPUC6StateResidency", mock.AnythingOfType("int")).Return(c6, nil).Times(len(cpuIDsMsr))

		// mock reading perf events.
		mFetcher.On("ReadPerfEvents").Return(nil).Once()

		// mock getting available CPU IDs with access to msr registers and coreFreq.
		mFetcher.On("GetPerfCPUIDs").Return(cpuIDsPerf).Once()

		// mock getting CPU C01 metric.
		mFetcher.On("GetCPUC0SubstateC01Percent", mock.AnythingOfType("int")).Return(c01, nil).Times(len(cpuIDsPerf))

		// mock getting CPU C02 metric.
		mFetcher.On("GetCPUC0SubstateC02Percent", mock.AnythingOfType("int")).Return(c02, nil).Times(len(cpuIDsPerf))

		p := &PowerStat{
			// Disables package metrics
			PackageMetrics: []string{},
			CPUMetrics: []string{
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
				cpuC1StateResidency.String(),
				cpuC6StateResidency.String(),
			},
			EventDefinitions: "./testdata/sapphirerapids_core.json",

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		require.NoError(t, p.Gather(acc))

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 12)
		require.True(t, acc.HasField("powerstat_core", "cpu_c0_substate_c01_percent"))
		require.True(t, acc.HasField("powerstat_core", "cpu_c0_substate_c02_percent"))
		require.True(t, acc.HasField("powerstat_core", "cpu_c1_state_residency_percent"))
		require.True(t, acc.HasField("powerstat_core", "cpu_c6_state_residency_percent"))
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithCPUAndPackageMetrics", func(t *testing.T) {
		cpuIDs := []int{10, 12}
		packageIDs := []int{0, 1, 2, 3}
		dieIDs := []int{0, 1}

		c7 := 0.12

		initMin := 200.0
		initMax := 1200.0
		currMin := 300.0
		currMax := 1300.0
		curr := 800.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available CPU IDs with access to msr registers and coreFreq.
		mFetcher.On("GetMsrCPUIDs").Return(cpuIDs).Once()

		// mock getting core ID for CPU IDs.
		mFetcher.On("GetCPUCoreID", mock.AnythingOfType("int")).Return(1, nil).Times(len(cpuIDs))

		// mock getting package ID for CPU IDs.
		mFetcher.On("GetCPUPackageID", mock.AnythingOfType("int")).Return(1, nil).Times(len(cpuIDs))

		// mock updating msr time-related metrics for CPU IDs.
		mFetcher.On("UpdatePerCPUMetrics", mock.AnythingOfType("int")).Return(nil).Times(len(cpuIDs))

		// mock getting C7 state residency metric for CPU IDs.
		mFetcher.On("GetCPUC7StateResidency", mock.AnythingOfType("int")).Return(c7, nil).Times(len(cpuIDs))

		// mock getting available package IDs.
		mFetcher.On("GetPackageIDs").Return(packageIDs).Once()

		// mock getting die IDs for package ID.
		mFetcher.On("GetPackageDieIDs", mock.AnythingOfType("int")).Return(dieIDs, nil).Times(len(packageIDs))

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(initMin, nil).
			Times(len(packageIDs) * len(dieIDs))

		// mock getting initial maximum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMax", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(initMax, nil).
			Times(len(packageIDs) * len(dieIDs))

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(currMin, nil).
			Times(len(packageIDs) * len(dieIDs))

		// mock getting custom maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(currMax, nil).
			Times(len(packageIDs) * len(dieIDs))

		// mock getting current uncore frequency value.
		mFetcher.On("GetCurrentUncoreFrequency", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(curr, nil).Times(len(packageIDs) * len(dieIDs))

		p := &PowerStat{
			PackageMetrics: []string{
				packageUncoreFrequency.String(),
			},
			CPUMetrics: []string{
				cpuC7StateResidency.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		require.NoError(t, p.Gather(acc))

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 18)
		require.True(t, acc.HasField("powerstat_core", "cpu_c7_state_residency_percent"))
		require.True(t, acc.HasField("powerstat_package", "uncore_frequency_limit_mhz_min"))
		require.True(t, acc.HasField("powerstat_package", "uncore_frequency_limit_mhz_max"))
		require.True(t, acc.HasField("powerstat_package", "uncore_frequency_mhz_cur"))
		require.True(t, acc.HasTag("powerstat_package", "type"))
		mFetcher.AssertExpectations(t)
	})
}

func TestStop(t *testing.T) {
	t.Run("NoErrorWithoutPerf", func(t *testing.T) {
		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			Log: logger,

			needsPerf: false,
		}

		p.Stop()

		require.Empty(t, logger.Errors())
	})

	t.Run("FailedToDeactivatePerfEvents", func(t *testing.T) {
		logger := &testutil.CaptureLogger{}

		mFetcher := &fetcherMock{}

		// mock deactivating perf events.
		mFetcher.On("DeactivatePerfEvents").Return(errors.New("mock error")).Once()

		p := &PowerStat{
			Log: logger,

			fetcher: mFetcher,

			needsPerf: true,
		}

		p.Stop()

		require.Len(t, logger.Errors(), 1)
		require.Contains(t, logger.Errors()[0], "Failed to deactivate perf events")
	})

	t.Run("NoErrorWithPerf", func(t *testing.T) {
		logger := &testutil.CaptureLogger{}

		mFetcher := &fetcherMock{}

		// mock deactivating perf events.
		mFetcher.On("DeactivatePerfEvents").Return(nil).Once()

		p := &PowerStat{
			Log: logger,

			fetcher: mFetcher,

			needsPerf: true,
		}

		p.Stop()

		require.Empty(t, logger.Errors())
	})
}

func TestDisableUnsupportedMetrics(t *testing.T) {
	t.Run("ModelMissing", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata/cpu_model_missing")

		p := &PowerStat{}

		err := p.disableUnsupportedMetrics()

		require.Error(t, err, "error occurred while parsing CPU model")
	})

	t.Run("MsrFlagNotFound", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata/msr_flag_not_found")

		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			CPUMetrics: []string{
				// Metrics relying on msr flag
				cpuC0StateResidency.String(),
				cpuC1StateResidency.String(),
				cpuC6StateResidency.String(),
				cpuBusyFrequency.String(),
				cpuBusyCycles.String(),
				cpuTemperature.String(),
			},
			PackageMetrics: []string{
				// Metrics relying on msr flag
				packageCPUBaseFrequency.String(),
				packageTurboLimit.String(),

				// Metrics not relying on msr flag
				packageCurrentPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},

			Log: logger,
		}

		err := p.disableUnsupportedMetrics()

		require.NoError(t, err)
		require.Empty(t, p.CPUMetrics)
		require.Len(t, p.PackageMetrics, 2)
		require.Contains(t, p.PackageMetrics, packageCurrentPowerConsumption.String())
		require.Contains(t, p.PackageMetrics, packageThermalDesignPower.String())
		require.Len(t, logger.Warnings(), 8)
	})

	t.Run("AperfMperfFlagNotFound", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata/aperfmperf_flag_not_found")

		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			CPUMetrics: []string{
				// Metrics relying on aperfmperf flag
				cpuC0StateResidency.String(),
				cpuC1StateResidency.String(),
				cpuBusyFrequency.String(),
				cpuBusyCycles.String(),

				// Metrics not relying on aperfmperf flag
				cpuTemperature.String(),
			},

			Log: logger,
		}

		err := p.disableUnsupportedMetrics()

		require.NoError(t, err)
		require.Len(t, p.CPUMetrics, 1)
		require.Contains(t, p.CPUMetrics, cpuTemperature.String())
		require.Len(t, logger.Warnings(), 4)
	})

	t.Run("DtsFlagNotFound", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata/dts_flag_not_found")

		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			CPUMetrics: []string{
				// Metrics relying on dts flag
				cpuTemperature.String(),

				// Metrics not relying on dts flag
				cpuBusyFrequency.String(),
			},
			PackageMetrics: []string{},

			Log: logger,
		}

		err := p.disableUnsupportedMetrics()

		require.NoError(t, err)
		require.Len(t, p.CPUMetrics, 1)
		require.Contains(t, p.CPUMetrics, cpuBusyFrequency.String())
		require.Len(t, logger.Warnings(), 1)
	})

	t.Run("ModelNotSupported", func(t *testing.T) {
		t.Setenv("HOST_PROC", "./testdata/model_not_supported")

		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			CPUMetrics: []string{
				// Metrics not supported by CPU
				cpuTemperature.String(),
				cpuC1StateResidency.String(),
				cpuC3StateResidency.String(),
				cpuC6StateResidency.String(),
				cpuC7StateResidency.String(),
			},
			PackageMetrics: []string{
				// Metrics not supported by CPU
				packageCPUBaseFrequency.String(),

				packageUncoreFrequency.String(),
			},

			Log: logger,
		}

		err := p.disableUnsupportedMetrics()

		require.NoError(t, err)
		require.Empty(t, p.CPUMetrics)
		require.Contains(t, p.PackageMetrics, packageUncoreFrequency.String())
		require.Len(t, logger.Warnings(), 6)
	})
}

func TestDisableCPUMetric(t *testing.T) {
	t.Run("NoMetricsRemoved", func(t *testing.T) {
		expStartLen := 1
		expEndLen := 1
		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			CPUMetrics: []string{"cpu_c1_state_residency"},
			Log:        logger,
		}

		require.Len(t, p.CPUMetrics, expStartLen)
		p.disableCPUMetric(cpuC3StateResidency)
		require.Len(t, p.CPUMetrics, expEndLen)

		require.Empty(t, logger.Warnings())
	})
	t.Run("TwoMetricsRemoved", func(t *testing.T) {
		expStartLen := 3
		expEndLen := 1
		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			CPUMetrics: []string{"cpu_c1_state_residency", "cpu_c3_state_residency", "cpu_c6_state_residency"},
			Log:        logger,
		}

		require.Len(t, p.CPUMetrics, expStartLen)
		p.disableCPUMetric(cpuC3StateResidency)
		p.disableCPUMetric(cpuC1StateResidency)
		require.Len(t, p.CPUMetrics, expEndLen)

		require.Len(t, logger.Warnings(), 2)
		require.Contains(t, logger.Warnings()[0], "\"cpu_c3_state_residency\" is not supported by CPU, metric will not be gathered")
		require.Contains(t, logger.Warnings()[1], "\"cpu_c1_state_residency\" is not supported by CPU, metric will not be gathered")
	})
}

func TestDisablePackageMetric(t *testing.T) {
	t.Run("NoMetricsRemoved", func(t *testing.T) {
		expStartLen := 1
		expEndLen := 1
		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			PackageMetrics: []string{"current_power_consumption"},
			Log:            logger,
		}

		require.Len(t, p.PackageMetrics, expStartLen)
		p.disablePackageMetric(packageCPUBaseFrequency)
		require.Len(t, p.PackageMetrics, expEndLen)

		require.Empty(t, logger.Warnings())
	})
	t.Run("TwoMetricsRemoved", func(t *testing.T) {
		expStartLen := 3
		expEndLen := 1
		logger := &testutil.CaptureLogger{}

		p := &PowerStat{
			PackageMetrics: []string{"current_power_consumption", "max_turbo_frequency", "cpu_base_frequency"},
			Log:            logger,
		}

		require.Len(t, p.PackageMetrics, expStartLen)
		p.disablePackageMetric(packageCPUBaseFrequency)
		p.disablePackageMetric(packageTurboLimit)
		require.Len(t, p.PackageMetrics, expEndLen)

		require.Len(t, logger.Warnings(), 2)
		require.Contains(t, logger.Warnings()[0], "\"cpu_base_frequency\" is not supported by CPU, metric will not be gathered")
		require.Contains(t, logger.Warnings()[1], "\"max_turbo_frequency\" is not supported by CPU, metric will not be gathered")
	})
}

type fetcherMock struct {
	mock.Mock
}

func (m *fetcherMock) GetMsrCPUIDs() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *fetcherMock) GetPerfCPUIDs() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *fetcherMock) GetPackageIDs() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *fetcherMock) GetCPUPackageID(cpuID int) (int, error) {
	args := m.Called(cpuID)
	return args.Int(0), args.Error(1)
}

func (m *fetcherMock) GetCPUCoreID(cpuID int) (int, error) {
	args := m.Called(cpuID)
	return args.Int(0), args.Error(1)
}

func (m *fetcherMock) GetPackageDieIDs(packageID int) ([]int, error) {
	args := m.Called(packageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int), args.Error(1)
}

func (m *fetcherMock) GetCPUFrequency(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) UpdatePerCPUMetrics(cpuID int) error {
	args := m.Called(cpuID)
	return args.Error(0)
}

func (m *fetcherMock) GetCPUTemperature(cpuID int) (uint64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *fetcherMock) GetCPUC0StateResidency(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCPUC1StateResidency(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCPUC3StateResidency(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCPUC6StateResidency(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCPUC7StateResidency(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCPUBusyFrequencyMhz(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) ReadPerfEvents() error {
	args := m.Called()
	return args.Error(0)
}

func (m *fetcherMock) DeactivatePerfEvents() error {
	args := m.Called()
	return args.Error(0)
}

func (m *fetcherMock) GetCPUC0SubstateC01Percent(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCPUC0SubstateC02Percent(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCPUC0SubstateC0WaitPercent(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCPUBaseFrequency(packageID int) (uint64, error) {
	args := m.Called(packageID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *fetcherMock) GetInitialUncoreFrequencyMin(packageID, dieID int) (float64, error) {
	args := m.Called(packageID, dieID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCustomizedUncoreFrequencyMin(packageID, dieID int) (float64, error) {
	args := m.Called(packageID, dieID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetInitialUncoreFrequencyMax(packageID, dieID int) (float64, error) {
	args := m.Called(packageID, dieID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCustomizedUncoreFrequencyMax(packageID, dieID int) (float64, error) {
	args := m.Called(packageID, dieID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCurrentUncoreFrequency(packageID, dieID int) (float64, error) {
	args := m.Called(packageID, dieID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCurrentPackagePowerConsumptionWatts(packageID int) (float64, error) {
	args := m.Called(packageID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetCurrentDramPowerConsumptionWatts(packageID int) (float64, error) {
	args := m.Called(packageID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetPackageThermalDesignPowerWatts(packageID int) (float64, error) {
	args := m.Called(packageID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *fetcherMock) GetMaxTurboFreqList(packageID int) ([]ptel.MaxTurboFreq, error) {
	args := m.Called(packageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ptel.MaxTurboFreq), args.Error(1)
}

func TestAddCPUMetrics(t *testing.T) {
	// Disable package metrics when parseConfig method is called.
	packageMetrics := []string{}

	t.Run("NoAvailableCPUs", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available CPU IDs with access to msr registers.
		mFetcher.On("GetMsrCPUIDs").Return(nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		p.addCPUMetrics(acc)

		require.Empty(t, acc.Errors)
		require.Empty(t, acc.GetTelegrafMetrics())
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithDataCPUIDErrors", func(t *testing.T) {
		t.Run("SingleCPUID", func(t *testing.T) {
			cpuID := 0

			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock getting available CPU IDs with access to msr registers and coreFreq.
			mFetcher.On("GetMsrCPUIDs").Return([]int{cpuID}).Once()

			// mock getting core ID for CPU ID.
			mFetcher.On("GetCPUCoreID", cpuID).Return(0, errors.New("mock error")).Once()

			p := &PowerStat{
				fetcher: mFetcher,
			}

			p.addCPUMetrics(acc)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get coreFreq and/or msr metrics for CPU ID %v", cpuID))
			require.Empty(t, acc.GetTelegrafMetrics())
			mFetcher.AssertExpectations(t)
		})

		t.Run("MultipleCPUIDs", func(t *testing.T) {
			cpuID := 1
			coreID := 2
			packageID := 3
			cpuFreq := 500.0

			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock getting available CPU IDs with access to msr registers and coreFreq.
			mFetcher.On("GetMsrCPUIDs").Return([]int{0, cpuID}).Once()

			// mock getting core ID for CPU ID 0.
			mFetcher.On("GetCPUCoreID", 0).Return(0, errors.New("mock error")).Once()

			// mock getting core ID for CPU ID 1.
			mFetcher.On("GetCPUCoreID", cpuID).Return(coreID, nil).Once()

			// mock getting package ID for CPU ID 1.
			mFetcher.On("GetCPUPackageID", cpuID).Return(packageID, nil).Once()

			// mock getting CPU frequency for CPU ID 1.
			mFetcher.On("GetCPUFrequency", cpuID).Return(cpuFreq, nil).Once()

			p := &PowerStat{
				CPUMetrics: []string{
					// Metric which relies on coreFreq.
					cpuFrequency.String(),

					// Metrics which do not rely on coreFreq nor msr.
					cpuC0SubstateC01Percent.String(),
					cpuC0SubstateC02Percent.String(),
				},
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,
			}

			require.NoError(t, p.parseConfig())

			p.addCPUMetrics(acc)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), "failed to get coreFreq and/or msr metrics for CPU ID 0")
			require.Len(t, acc.GetTelegrafMetrics(), 1)
			acc.AssertContainsTaggedFields(
				t,
				// measurement
				"powerstat_core",
				// fields
				map[string]interface{}{
					"cpu_frequency_mhz": cpuFreq,
				},
				// tags
				map[string]string{
					"cpu_id":     strconv.Itoa(cpuID),
					"core_id":    strconv.Itoa(coreID),
					"package_id": strconv.Itoa(packageID),
				},
			)
			mFetcher.AssertExpectations(t)
		})
	})

	t.Run("WithCoreFreqMetrics", func(t *testing.T) {
		cpuFreq := 500.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available CPU IDs with access to coreFreq.
		mFetcher.On("GetMsrCPUIDs").Return([]int{0, 1}).Once()

		// mock getting corresponding core ID to CPU IDs 0 and 1.
		mFetcher.On("GetCPUCoreID", mock.AnythingOfType("int")).Return(0, nil).Twice()

		// mock getting corresponding package ID to CPU IDs 0 and 1.
		mFetcher.On("GetCPUPackageID", mock.AnythingOfType("int")).Return(1, nil).Twice()

		// mock getting CPU frequency for CPU ID 0.
		mFetcher.On("GetCPUFrequency", 0).Return(cpuFreq, nil).Once()

		// mock getting CPU frequency for CPU ID 1.
		mFetcher.On("GetCPUFrequency", 1).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			CPUMetrics: []string{
				// Metric which relies on coreFreq.
				cpuFrequency.String(),

				// Metrics which do not rely on coreFreq nor msr
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
			},
			PackageMetrics:   packageMetrics,
			EventDefinitions: "./testdata/sapphirerapids_core.json",

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addCPUMetrics(acc)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID 1", cpuFrequency))
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_frequency_mhz": cpuFreq,
			},
			// tags
			map[string]string{
				"cpu_id":     "0",
				"core_id":    "0",
				"package_id": "1",
			},
		)
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithMsrMetrics", func(t *testing.T) {
		t.Run("SingleRead", func(t *testing.T) {
			cpuTemp := uint64(18)

			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock getting available CPU IDs with access to msr registers.
			mFetcher.On("GetMsrCPUIDs").Return([]int{0, 1}).Once()

			// mock getting corresponding core ID to CPU IDs 0 and 1.
			mFetcher.On("GetCPUCoreID", mock.AnythingOfType("int")).Return(0, nil).Twice()

			// mock getting corresponding package ID to CPU IDs 0 and 1.
			mFetcher.On("GetCPUPackageID", mock.AnythingOfType("int")).Return(1, nil).Twice()

			// mock getting CPU temperature metric for CPU ID 0.
			mFetcher.On("GetCPUTemperature", 0).Return(cpuTemp, nil).Once()

			// mock getting CPU temperature metric for CPU ID 1.
			mFetcher.On("GetCPUTemperature", 1).Return(uint64(0), errors.New("mock error")).Once()

			p := &PowerStat{
				CPUMetrics: []string{
					// Metrics which rely on single-read msr registers.
					cpuTemperature.String(),

					// Metrics which do not rely on coreFreq nor msr
					cpuC0SubstateC01Percent.String(),
					cpuC0SubstateC02Percent.String(),
				},
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,
			}

			require.NoError(t, p.parseConfig())

			p.addCPUMetrics(acc)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID 1", cpuTemperature))
			require.Len(t, acc.GetTelegrafMetrics(), 1)
			acc.AssertContainsTaggedFields(
				t,
				// measurement
				"powerstat_core",
				// fields
				map[string]interface{}{
					"cpu_temperature_celsius": cpuTemp,
				},
				// tags
				map[string]string{
					"cpu_id":     "0",
					"core_id":    "0",
					"package_id": "1",
				},
			)
			mFetcher.AssertExpectations(t)
		})

		t.Run("TimeRelated", func(t *testing.T) {
			cpuBusyFreq := 750.0

			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock getting available CPU IDs with access to msr registers.
			mFetcher.On("GetMsrCPUIDs").Return([]int{0, 1}).Once()

			// mock getting corresponding core ID to CPU IDs 0 and 1.
			mFetcher.On("GetCPUCoreID", mock.AnythingOfType("int")).Return(0, nil).Twice()

			// mock getting corresponding package ID to CPU IDs 0 and 1.
			mFetcher.On("GetCPUPackageID", mock.AnythingOfType("int")).Return(1, nil).Twice()

			// mock updating msr time-related metrics for CPU ID 0.
			mFetcher.On("UpdatePerCPUMetrics", 0).Return(errors.New("mock error")).Once()

			// mock updating msr time-related metrics for CPU ID 1.
			mFetcher.On("UpdatePerCPUMetrics", 1).Return(nil).Once()

			// mock getting CPU busy frequency metric for CPU ID 1.
			mFetcher.On("GetCPUBusyFrequencyMhz", 1).Return(cpuBusyFreq, nil).Once()

			p := &PowerStat{
				CPUMetrics: []string{
					// Metrics which rely on time-related msr reads.
					cpuBusyFrequency.String(),

					// Metrics which do not rely on coreFreq nor msr
					cpuC0SubstateC01Percent.String(),
					cpuC0SubstateC02Percent.String(),
					cpuC0SubstateC0WaitPercent.String(),
				},
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,
			}

			require.NoError(t, p.parseConfig())

			p.addCPUMetrics(acc)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), "failed to update MSR time-related metrics for CPU ID 0")
			require.Len(t, acc.GetTelegrafMetrics(), 1)
			acc.AssertContainsTaggedFields(
				t,
				// measurement
				"powerstat_core",
				// fields
				map[string]interface{}{
					"cpu_busy_frequency_mhz": cpuBusyFreq,
				},
				// tags
				map[string]string{
					"cpu_id":     "1",
					"core_id":    "0",
					"package_id": "1",
				},
			)
			mFetcher.AssertExpectations(t)
		})
	})
}

func TestAddPerCPUMsrMetrics(t *testing.T) {
	// Disable package metrics when parseConfig method is called.
	packageMetrics := []string{}

	t.Run("WithoutMsrMetrics", func(t *testing.T) {
		cpuID := 0
		coreID := 1
		packageID := 0

		acc := &testutil.Accumulator{}

		p := &PowerStat{
			CPUMetrics: []string{
				// metrics which do not rely on msr
				cpuFrequency.String(),
				cpuC0SubstateC01Percent.String(),
			},
		}

		p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.Errors)
		require.Empty(t, acc.GetTelegrafMetrics())
	})

	t.Run("WithSingleMsrReadMetrics", func(t *testing.T) {
		cpuID := 0
		coreID := 1
		packageID := 0
		cpuMetrics := []string{
			// metric that relies on a single msr read.
			cpuTemperature.String(),

			// metrics that do not rely on msr.
			cpuFrequency.String(),
			cpuC0SubstateC01Percent.String(),
		}

		t.Run("WithError", func(t *testing.T) {
			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock getting CPU temperature metric.
			mFetcher.On("GetCPUTemperature", cpuID).Return(uint64(0), errors.New("mock error")).Once()

			p := &PowerStat{
				CPUMetrics:       cpuMetrics,
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,

				logOnce: map[string]struct{}{},
			}

			require.NoError(t, p.parseConfig())

			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuTemperature, cpuID))
			require.Empty(t, p.logOnce)
			require.Empty(t, acc.GetTelegrafMetrics())
			mFetcher.AssertExpectations(t)
		})

		t.Run("WithModuleNotInitializedError", func(t *testing.T) {
			acc := &testutil.Accumulator{}

			mErr := &ptel.ModuleNotInitializedError{Name: "msr"}
			mFetcher := &fetcherMock{}

			// mock getting CPU temperature metric.
			mFetcher.On("GetCPUTemperature", cpuID).Return(uint64(0), mErr).Twice()

			p := &PowerStat{
				CPUMetrics:       cpuMetrics,
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,

				logOnce: map[string]struct{}{},
			}

			require.NoError(t, p.parseConfig())

			// First call adds the error to the accumulator and logOnce map.
			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

			// Second call detects previous error in logOnce map and skips adding it to the accumulator.
			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q: %v", cpuTemperature, mErr))
			require.Empty(t, acc.GetTelegrafMetrics())

			require.Len(t, p.logOnce, 1)
			require.Contains(t, p.logOnce, "msr_cpu_temperature")

			mFetcher.AssertExpectations(t)
		})

		t.Run("WithoutErrors", func(t *testing.T) {
			cpuTemp := uint64(20)

			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock getting CPU temperature metric.
			mFetcher.On("GetCPUTemperature", cpuID).Return(cpuTemp, nil).Once()

			p := &PowerStat{
				CPUMetrics:       cpuMetrics,
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,
			}

			require.NoError(t, p.parseConfig())

			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

			require.Empty(t, acc.Errors)
			require.Len(t, acc.GetTelegrafMetrics(), 1)
			acc.AssertContainsTaggedFields(
				t,
				// measurement
				"powerstat_core",
				// fields
				map[string]interface{}{
					"cpu_temperature_celsius": cpuTemp,
				},
				// tags
				map[string]string{
					"cpu_id":     strconv.Itoa(cpuID),
					"core_id":    strconv.Itoa(coreID),
					"package_id": strconv.Itoa(packageID),
				},
			)
			mFetcher.AssertExpectations(t)
		})
	})

	t.Run("WithTimeRelatedMsrMetrics", func(t *testing.T) {
		cpuID := 0
		coreID := 1
		packageID := 0

		c1State := 5.15
		c6State := 8.10

		cpuMetrics := []string{
			// metrics that rely on a time-related msr.
			cpuC1StateResidency.String(),
			cpuC6StateResidency.String(),

			// metrics which do not rely on msr.
			cpuFrequency.String(),
			cpuC0SubstateC01Percent.String(),
		}

		t.Run("FailedToUpdate", func(t *testing.T) {
			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock updating msr time-related metrics.
			mFetcher.On("UpdatePerCPUMetrics", cpuID).Return(errors.New("mock error")).Once()

			p := &PowerStat{
				CPUMetrics:       cpuMetrics,
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,
			}

			require.NoError(t, p.parseConfig())

			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to update MSR time-related metrics for CPU ID %v", cpuID))
			require.Empty(t, acc.GetTelegrafMetrics())
			mFetcher.AssertExpectations(t)
		})

		t.Run("FailedToUpdateModuleNotInitializedError", func(t *testing.T) {
			acc := &testutil.Accumulator{}

			mErr := &ptel.ModuleNotInitializedError{Name: "msr"}
			mFetcher := &fetcherMock{}

			// mock updating msr time-related metrics.
			mFetcher.On("UpdatePerCPUMetrics", cpuID).Return(mErr).Twice()

			p := &PowerStat{
				CPUMetrics:       cpuMetrics,
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,

				logOnce: map[string]struct{}{},
			}

			require.NoError(t, p.parseConfig())

			// First call adds the error to the accumulator and key to logOnce map.
			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

			// Second call detects previous error in logOnce map and skips adding it to the accumulator.
			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to update MSR time-related metrics: %v", mErr))
			require.Empty(t, acc.GetTelegrafMetrics())

			require.Len(t, p.logOnce, 1)
			require.Contains(t, p.logOnce, "msr_time_related")

			mFetcher.AssertExpectations(t)
		})

		t.Run("WithoutErrors", func(t *testing.T) {
			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock updating msr time-related metrics.
			mFetcher.On("UpdatePerCPUMetrics", cpuID).Return(nil).Once()

			// mock getting C1 state residency.
			mFetcher.On("GetCPUC1StateResidency", cpuID).Return(c1State, nil).Once()

			// mock getting C6 state residency.
			mFetcher.On("GetCPUC6StateResidency", cpuID).Return(c6State, nil).Once()

			p := &PowerStat{
				CPUMetrics:       cpuMetrics,
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,
			}

			require.NoError(t, p.parseConfig())

			p.addPerCPUMsrMetrics(acc, cpuID, coreID, packageID)

			require.Empty(t, acc.Errors)
			require.Len(t, acc.GetTelegrafMetrics(), 2)
			acc.AssertContainsTaggedFields(
				t,
				// measurement
				"powerstat_core",
				// fields
				map[string]interface{}{
					"cpu_c1_state_residency_percent": c1State,
				},
				// flags
				map[string]string{
					"cpu_id":     strconv.Itoa(cpuID),
					"core_id":    strconv.Itoa(coreID),
					"package_id": strconv.Itoa(packageID),
				},
			)
			acc.AssertContainsTaggedFields(
				t,
				// measurement
				"powerstat_core",
				// fields
				map[string]interface{}{
					"cpu_c6_state_residency_percent": c6State,
				},
				// flags
				map[string]string{
					"cpu_id":     strconv.Itoa(cpuID),
					"core_id":    strconv.Itoa(coreID),
					"package_id": strconv.Itoa(packageID),
				},
			)
			mFetcher.AssertExpectations(t)
		})
	})
}

func TestAddCPUTimeRelatedMsrMetrics(t *testing.T) {
	cpuID := 0
	coreID := 1
	packageID := 0

	c0State := 3.0
	c1State := 2.0
	c6State := 1.0

	acc := &testutil.Accumulator{}

	mFetcher := &fetcherMock{}

	// mock getting CPU C0 state residency value.
	mFetcher.On("GetCPUC0StateResidency", cpuID).Return(c0State, nil).Once()

	// mock getting CPU C1 state residency value.
	mFetcher.On("GetCPUC1StateResidency", cpuID).Return(c1State, nil).Once()

	// mock getting CPU C6 state residency value.
	mFetcher.On("GetCPUC6StateResidency", cpuID).Return(c6State, nil).Once()

	p := &PowerStat{
		CPUMetrics: []string{
			// Metrics which are not time-related MSR.
			cpuFrequency.String(),
			cpuTemperature.String(),
			cpuC0SubstateC01Percent.String(),

			// Time-related MSR metrics.
			cpuC0StateResidency.String(),
			cpuC1StateResidency.String(),
			cpuC6StateResidency.String(),
		},
		PackageMetrics:   []string{},
		EventDefinitions: "./testdata/sapphirerapids_core.json",

		fetcher: mFetcher,
	}

	require.NoError(t, p.parseConfig())
	require.Empty(t, acc.GetTelegrafMetrics())

	p.addCPUTimeRelatedMsrMetrics(acc, cpuID, coreID, packageID)

	require.Empty(t, acc.Errors)
	require.Len(t, acc.GetTelegrafMetrics(), 3)
	acc.AssertContainsTaggedFields(
		t,
		// measurement
		"powerstat_core",
		// fields
		map[string]interface{}{
			"cpu_c0_state_residency_percent": c0State,
		},
		// tags
		map[string]string{
			"cpu_id":     strconv.Itoa(cpuID),
			"core_id":    strconv.Itoa(coreID),
			"package_id": strconv.Itoa(packageID),
		},
	)
	acc.AssertContainsTaggedFields(
		t,
		// measurement
		"powerstat_core",
		// fields
		map[string]interface{}{
			"cpu_c1_state_residency_percent": c1State,
		},
		// tags
		map[string]string{
			"cpu_id":     strconv.Itoa(cpuID),
			"core_id":    strconv.Itoa(coreID),
			"package_id": strconv.Itoa(packageID),
		},
	)
	acc.AssertContainsTaggedFields(
		t,
		// measurement
		"powerstat_core",
		// fields
		map[string]interface{}{
			"cpu_c6_state_residency_percent": c6State,
		},
		// tags
		map[string]string{
			"cpu_id":     strconv.Itoa(cpuID),
			"core_id":    strconv.Itoa(coreID),
			"package_id": strconv.Itoa(packageID),
		},
	)
	mFetcher.AssertExpectations(t)
}

func TestAddCPUPerfMetrics(t *testing.T) {
	// Disable package metrics when parseConfig method is called.
	packageMetrics := []string{}

	t.Run("FailedToReadPerfEvents", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock reading perf events.
		mFetcher.On("ReadPerfEvents").Return(errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		p.addCPUPerfMetrics(acc)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), "failed to read perf events")
		mFetcher.AssertExpectations(t)
	})

	t.Run("FailedToReadPerfEventsModuleNotInitializedError", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		mErr := &ptel.ModuleNotInitializedError{Name: "perf"}
		mFetcher := &fetcherMock{}

		// mock reading perf events.
		mFetcher.On("ReadPerfEvents").Return(mErr).Twice()

		p := &PowerStat{
			fetcher: mFetcher,

			logOnce: map[string]struct{}{},
		}

		// First call adds the error to the accumulator and key to logOnce map.
		p.addCPUPerfMetrics(acc)

		// Second call detects previous error in logOnce map and skips adding it to the accumulator.
		p.addCPUPerfMetrics(acc)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to read perf events: %v", mErr))
		require.Empty(t, acc.GetTelegrafMetrics())

		require.Len(t, p.logOnce, 1)
		require.Contains(t, p.logOnce, "perf_read")

		mFetcher.AssertExpectations(t)
	})

	t.Run("NoAvailableCPUs", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock reading perf events.
		mFetcher.On("ReadPerfEvents").Return(nil).Once()

		// mock getting available CPU IDs for perf events.
		mFetcher.On("GetPerfCPUIDs").Return(nil).Once()

		p := &PowerStat{
			CPUMetrics: []string{
				// Metrics which do not rely on perf.
				cpuFrequency.String(),
				cpuTemperature.String(),
				cpuBusyCycles.String(),

				// Metrics which rely on perf.
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
				cpuC0SubstateC0WaitPercent.String(),
			},
			PackageMetrics:   []string{},
			EventDefinitions: "./testdata/sapphirerapids_core.json",

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addCPUPerfMetrics(acc)

		require.Empty(t, acc.Errors)
		require.Empty(t, acc.GetTelegrafMetrics())
		mFetcher.AssertExpectations(t)
	})

	t.Run("FailedToGetDataCPUID", func(t *testing.T) {
		t.Run("SingleCPUID", func(t *testing.T) {
			cpuID := 0

			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock reading perf events.
			mFetcher.On("ReadPerfEvents").Return(nil).Once()

			// mock getting available CPU IDs for perf events.
			mFetcher.On("GetPerfCPUIDs").Return([]int{cpuID}).Once()

			// mock getting corresponding core ID to CPU ID 0.
			mFetcher.On("GetCPUCoreID", cpuID).Return(1, nil).Once()

			// mock getting corresponding package ID to CPU ID 0.
			mFetcher.On("GetCPUPackageID", cpuID).Return(0, errors.New("mock error")).Once()

			p := &PowerStat{
				CPUMetrics: []string{
					// Metrics which do not rely on perf.
					cpuFrequency.String(),
					cpuTemperature.String(),
					cpuBusyCycles.String(),

					// Metrics which rely on perf.
					cpuC0SubstateC01Percent.String(),
					cpuC0SubstateC02Percent.String(),
					cpuC0SubstateC0WaitPercent.String(),
				},
				PackageMetrics:   []string{},
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,
			}

			require.NoError(t, p.parseConfig())

			p.addCPUPerfMetrics(acc)

			require.Empty(t, acc.GetTelegrafMetrics())
			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get perf metrics for CPU ID %v", cpuID))
			mFetcher.AssertExpectations(t)
		})

		t.Run("MultipleCPUIDs", func(t *testing.T) {
			cpuID := 0
			coreID := 2
			packageID := 3

			c01Percent := 0.2

			acc := &testutil.Accumulator{}

			mFetcher := &fetcherMock{}

			// mock reading perf events.
			mFetcher.On("ReadPerfEvents").Return(nil).Once()

			// mock getting available CPU IDs for perf events.
			mFetcher.On("GetPerfCPUIDs").Return([]int{cpuID, 1}).Once()

			// mock getting corresponding core ID to CPU ID 0.
			mFetcher.On("GetCPUCoreID", cpuID).Return(coreID, nil).Once()

			// mock getting corresponding package ID to CPU ID 0.
			mFetcher.On("GetCPUPackageID", cpuID).Return(packageID, nil).Once()

			// mock getting CPU C01 metric.
			mFetcher.On("GetCPUC0SubstateC01Percent", cpuID).Return(c01Percent, nil).Once()

			// mock getting corresponding core ID to CPU ID 1.
			mFetcher.On("GetCPUCoreID", 1).Return(5, nil).Once()

			// mock getting corresponding package ID to CPU ID 1.
			mFetcher.On("GetCPUPackageID", 1).Return(0, errors.New("mock error")).Once()

			p := &PowerStat{
				CPUMetrics: []string{
					// Metrics which do not rely on perf.
					cpuFrequency.String(),
					cpuTemperature.String(),
					cpuC6StateResidency.String(),

					// Metrics which rely on perf.
					cpuC0SubstateC01Percent.String(),
				},
				PackageMetrics:   packageMetrics,
				EventDefinitions: "./testdata/sapphirerapids_core.json",

				fetcher: mFetcher,
			}

			require.NoError(t, p.parseConfig())

			p.addCPUPerfMetrics(acc)

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.FirstError(), "failed to get perf metrics for CPU ID 1")
			require.Len(t, acc.GetTelegrafMetrics(), 1)
			acc.AssertContainsTaggedFields(
				t,
				// measurement
				"powerstat_core",
				// fields
				map[string]interface{}{
					"cpu_c0_substate_c01_percent": c01Percent,
				},
				// tags
				map[string]string{
					"cpu_id":     strconv.Itoa(cpuID),
					"core_id":    strconv.Itoa(coreID),
					"package_id": strconv.Itoa(packageID),
				},
			)
			mFetcher.AssertExpectations(t)
		})
	})

	t.Run("WithError", func(t *testing.T) {
		cpuID := 0
		coreID := 1
		packageID := 0

		c01Percent := 0.5
		c0Wait := 2.5

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock reading perf events.
		mFetcher.On("ReadPerfEvents").Return(nil).Once()

		// mock getting available CPU IDs for perf events.
		mFetcher.On("GetPerfCPUIDs").Return([]int{cpuID}).Once()

		// mock getting corresponding core ID to CPU ID 0.
		mFetcher.On("GetCPUCoreID", cpuID).Return(coreID, nil).Once()

		// mock getting corresponding package ID to CPU ID 0.
		mFetcher.On("GetCPUPackageID", cpuID).Return(packageID, nil).Once()

		// mock getting CPU C01 metric.
		mFetcher.On("GetCPUC0SubstateC01Percent", cpuID).Return(c01Percent, nil).Once()

		// mock getting CPU C02 metric.
		mFetcher.On("GetCPUC0SubstateC02Percent", cpuID).Return(0.0, errors.New("mock error")).Once()

		// mock getting CPU C0Wait metric.
		mFetcher.On("GetCPUC0SubstateC0WaitPercent", cpuID).Return(c0Wait, nil).Once()

		p := &PowerStat{
			CPUMetrics: []string{
				// Metrics which do not rely on perf.
				cpuFrequency.String(),
				cpuTemperature.String(),
				cpuC6StateResidency.String(),

				// Metrics which rely on perf.
				cpuC0SubstateC01Percent.String(),
				cpuC0SubstateC02Percent.String(),
				cpuC0SubstateC0WaitPercent.String(),
			},
			PackageMetrics:   packageMetrics,
			EventDefinitions: "./testdata/sapphirerapids_core.json",

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addCPUPerfMetrics(acc)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC0SubstateC02Percent, cpuID))
		require.Len(t, acc.GetTelegrafMetrics(), 2)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c0_substate_c01_percent": c01Percent,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c0_substate_c0_wait_percent": c0Wait,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
	})
}

func TestAddPerCPUPerfMetrics(t *testing.T) {
	cpuID := 0
	coreID := 1
	packageID := 0

	c01Percent := 1.09
	c02Percent := 2.12

	acc := &testutil.Accumulator{}

	mFetcher := &fetcherMock{}

	// mock getting CPU C01 metric.
	mFetcher.On("GetCPUC0SubstateC01Percent", cpuID).Return(c01Percent, nil).Once()

	// mock getting CPU C02 metric.
	mFetcher.On("GetCPUC0SubstateC02Percent", cpuID).Return(c02Percent, nil).Once()

	p := &PowerStat{
		CPUMetrics: []string{
			// Metrics which do not rely on perf.
			cpuFrequency.String(),
			cpuTemperature.String(),
			cpuC6StateResidency.String(),

			// Metrics which rely on perf.
			cpuC0SubstateC01Percent.String(),
			cpuC0SubstateC02Percent.String(),
		},
		PackageMetrics:   []string{},
		EventDefinitions: "./testdata/sapphirerapids_core.json",

		fetcher: mFetcher,
	}

	require.NoError(t, p.parseConfig())
	require.Empty(t, acc.GetTelegrafMetrics())

	p.addPerCPUPerfMetrics(acc, cpuID, coreID, packageID)

	require.Len(t, acc.GetTelegrafMetrics(), 2)
	acc.AssertContainsTaggedFields(
		t,
		// measurement
		"powerstat_core",
		// fields
		map[string]interface{}{
			"cpu_c0_substate_c01_percent": c01Percent,
		},
		// tags
		map[string]string{
			"cpu_id":     strconv.Itoa(cpuID),
			"core_id":    strconv.Itoa(coreID),
			"package_id": strconv.Itoa(packageID),
		},
	)
	acc.AssertContainsTaggedFields(
		t,
		// measurement
		"powerstat_core",
		// fields
		map[string]interface{}{
			"cpu_c0_substate_c02_percent": c02Percent,
		},
		// tags
		map[string]string{
			"cpu_id":     strconv.Itoa(cpuID),
			"core_id":    strconv.Itoa(coreID),
			"package_id": strconv.Itoa(packageID),
		},
	)
	mFetcher.AssertExpectations(t)
}

func TestGetDataCPUID(t *testing.T) {
	t.Run("FailedToGetCoreID", func(t *testing.T) {
		cpuID := 1

		mFetcher := &fetcherMock{}

		// mock getting core ID corresponding to the CPU ID.
		mFetcher.On("GetCPUCoreID", cpuID).Return(0, errors.New("mock error")).Once()

		coreID, packageID, err := getDataCPUID(mFetcher, cpuID)

		require.Equal(t, 0, coreID)
		require.Equal(t, 0, packageID)
		require.ErrorContains(t, err, fmt.Sprintf("failed to get core ID from CPU ID %v", cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("FailedToGetPackageID", func(t *testing.T) {
		cpuID := 1

		mFetcher := &fetcherMock{}

		// mock getting core ID corresponding to the CPU ID.
		mFetcher.On("GetCPUCoreID", cpuID).Return(1, nil).Once()

		// mock getting package ID corresponding to the CPU ID.
		mFetcher.On("GetCPUPackageID", cpuID).Return(0, errors.New("mock error")).Once()

		coreID, packageID, err := getDataCPUID(mFetcher, cpuID)

		require.Equal(t, 0, coreID)
		require.Equal(t, 0, packageID)
		require.ErrorContains(t, err, fmt.Sprintf("failed to get package ID from CPU ID %v", cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		cpuID := 1

		mFetcher := &fetcherMock{}

		// mock getting core ID corresponding to the CPU ID.
		mFetcher.On("GetCPUCoreID", cpuID).Return(1, nil).Once()

		// mock getting package ID corresponding to the CPU ID.
		mFetcher.On("GetCPUPackageID", cpuID).Return(2, nil).Once()

		coreID, packageID, err := getDataCPUID(mFetcher, cpuID)

		require.Equal(t, 1, coreID)
		require.Equal(t, 2, packageID)
		require.NoError(t, err)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddPackageMetrics(t *testing.T) {
	t.Run("NoPackageIDs", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available package IDs.
		mFetcher.On("GetPackageIDs").Return(nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		p.addPackageMetrics(acc)

		require.Empty(t, acc.Errors)
		require.Empty(t, acc.GetTelegrafMetrics())
	})

	t.Run("WithRaplMetrics", func(t *testing.T) {
		tdp := 80.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available package IDs.
		mFetcher.On("GetPackageIDs").Return([]int{0, 1}).Once()

		// mock getting package thermal design power metric for CPU ID 0.
		mFetcher.On("GetPackageThermalDesignPowerWatts", 0).Return(tdp, nil).Once()

		// mock getting package thermal design power metric for CPU ID 1.
		mFetcher.On("GetPackageThermalDesignPowerWatts", 1).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which rely on rapl
				packageThermalDesignPower.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addPackageMetrics(acc)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID 1", packageThermalDesignPower))
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"thermal_design_power_watts": tdp,
			},
			// tags
			map[string]string{
				"package_id": "0",
			},
		)
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithMsrMetrics", func(t *testing.T) {
		baseFreq := uint64(400)

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available package IDs.
		mFetcher.On("GetPackageIDs").Return([]int{0, 1}).Once()

		// mock getting CPU base frequency metric, for package ID 0.
		mFetcher.On("GetCPUBaseFrequency", 0).Return(uint64(0), errors.New("mock error")).Once()

		// mock getting CPU base frequency metric, for package ID 1.
		mFetcher.On("GetCPUBaseFrequency", 1).Return(baseFreq, nil).Once()

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which rely on msr
				packageCPUBaseFrequency.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addPackageMetrics(acc)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID 0", packageCPUBaseFrequency))
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"cpu_base_frequency_mhz": baseFreq,
			},
			// tags
			map[string]string{
				"package_id": "1",
			},
		)
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithUncoreFreqMetric", func(t *testing.T) {
		dieID := 0

		initMin := 500.0
		initMax := 2500.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting available package IDs.
		mFetcher.On("GetPackageIDs").Return([]int{0, 1}).Once()

		// mock getting die IDs for package ID.
		mFetcher.On("GetPackageDieIDs", mock.AnythingOfType("int")).Return([]int{dieID}, nil).Twice()

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", mock.AnythingOfType("int"), dieID).Return(initMin, nil).Twice()

		// mock getting initial maximum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMax", mock.AnythingOfType("int"), dieID).Return(initMax, nil).Twice()

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", mock.AnythingOfType("int"), dieID).Return(600.0, nil).Twice()

		// mock getting custom maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", mock.AnythingOfType("int"), dieID).Return(0.0, errors.New("mock error")).Twice()

		p := &PowerStat{
			PackageMetrics: []string{
				packageUncoreFrequency.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addPackageMetrics(acc)

		require.Len(t, acc.Errors, 2)
		require.ErrorContains(t, acc.Errors[0], fmt.Sprintf("failed to get current uncore frequency values for package ID 0 and die ID %v", dieID))
		require.ErrorContains(t, acc.Errors[1], fmt.Sprintf("failed to get current uncore frequency values for package ID 1 and die ID %v", dieID))
		require.Len(t, acc.GetTelegrafMetrics(), 2)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": initMin,
				"uncore_frequency_limit_mhz_max": initMax,
			},
			// tags
			map[string]string{
				"package_id": "0",
				"type":       "initial",
				"die":        strconv.Itoa(dieID),
			},
		)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": initMin,
				"uncore_frequency_limit_mhz_max": initMax,
			},
			// tags
			map[string]string{
				"package_id": "1",
				"type":       "initial",
				"die":        strconv.Itoa(dieID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddPerPackageRaplMetrics(t *testing.T) {
	t.Run("WithoutRaplMetrics", func(t *testing.T) {
		packageID := 0

		acc := &testutil.Accumulator{}

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which do not rely on rapl
				packageCPUBaseFrequency.String(),
				packageUncoreFrequency.String(),
				packageTurboLimit.String(),
			},
		}

		p.addPerPackageRaplMetrics(acc, packageID)

		require.Empty(t, acc.Errors)
		require.Empty(t, acc.GetTelegrafMetrics())
	})

	t.Run("WithModuleNotInitializedError", func(t *testing.T) {
		packageID := 0

		acc := &testutil.Accumulator{}

		raplNotInitErr := &ptel.ModuleNotInitializedError{Name: "rapl"}
		mError := fmt.Errorf("mock error: %w", raplNotInitErr)
		mFetcher := &fetcherMock{}

		// mock getting current dram power consumption metric.
		mFetcher.On("GetCurrentDramPowerConsumptionWatts", packageID).Return(0.0, mError).Twice()

		// mock getting package thermal design power metric.
		mFetcher.On("GetPackageThermalDesignPowerWatts", packageID).Return(0.0, mError).Twice()

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which rely on rapl
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),

				// metrics which do not rely on rapl
				packageCPUBaseFrequency.String(),
				packageUncoreFrequency.String(),
				packageTurboLimit.String(),
			},

			fetcher: mFetcher,

			logOnce: map[string]struct{}{},
		}

		require.NoError(t, p.parseConfig())

		// First call adds the error to the accumulator and logOnce map.
		p.addPerPackageRaplMetrics(acc, packageID)

		// Second call detects previous error in logOnce map and skips adding it to the accumulator.
		p.addPerPackageRaplMetrics(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 2)
		require.ErrorContains(t, acc.Errors[0], fmt.Sprintf("failed to get %q: %v", packageCurrentDramPowerConsumption, raplNotInitErr))
		require.ErrorContains(t, acc.Errors[1], fmt.Sprintf("failed to get %q: %v", packageThermalDesignPower, raplNotInitErr))

		require.Len(t, p.logOnce, 2)
		require.Contains(t, p.logOnce, "rapl_current_dram_power_consumption")
		require.Contains(t, p.logOnce, "rapl_thermal_design_power")
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithErrors", func(t *testing.T) {
		packageID := 0
		currPower := 30.0
		tdp := 80.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting current package power consumption metric.
		mFetcher.On("GetCurrentPackagePowerConsumptionWatts", packageID).Return(currPower, nil).Once()

		// mock getting current dram power consumption metric.
		mFetcher.On("GetCurrentDramPowerConsumptionWatts", packageID).Return(0.0, errors.New("mock error")).Once()

		// mock getting package thermal design power metric.
		mFetcher.On("GetPackageThermalDesignPowerWatts", packageID).Return(tdp, nil).Once()

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which rely on rapl
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),

				// metrics which do not rely on rapl
				packageCPUBaseFrequency.String(),
				packageUncoreFrequency.String(),
				packageTurboLimit.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addPerPackageRaplMetrics(acc, packageID)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID %v", packageCurrentDramPowerConsumption, packageID))
		require.Len(t, acc.GetTelegrafMetrics(), 2)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"current_power_consumption_watts": currPower,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"thermal_design_power_watts": tdp,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithoutErrors", func(t *testing.T) {
		packageID := 0
		currPower := 10.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting current dram power consumption metric.
		mFetcher.On("GetCurrentDramPowerConsumptionWatts", packageID).Return(currPower, nil).Once()

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which rely on rapl
				packageCurrentDramPowerConsumption.String(),

				// metrics which do not rely on rapl
				packageCPUBaseFrequency.String(),
				packageUncoreFrequency.String(),
				packageTurboLimit.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addPerPackageRaplMetrics(acc, packageID)

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"current_dram_power_consumption_watts": currPower,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddPerPackageMsrMetrics(t *testing.T) {
	t.Run("WithoutMsrMetrics", func(t *testing.T) {
		packageID := 0

		acc := &testutil.Accumulator{}

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which do not rely on msr
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},
		}

		p.addPerPackageMsrMetrics(acc, packageID)

		require.Empty(t, acc.Errors)
		require.Empty(t, acc.GetTelegrafMetrics())
	})

	t.Run("WithModuleNotInitializedError", func(t *testing.T) {
		packageID := 0

		acc := &testutil.Accumulator{}

		msrNotInitErr := &ptel.ModuleNotInitializedError{Name: "msr"}
		mError := fmt.Errorf("mock error: %w", msrNotInitErr)
		mFetcher := &fetcherMock{}

		// mock getting CPU base frequency metric.
		mFetcher.On("GetCPUBaseFrequency", packageID).Return(uint64(400), mError).Twice()

		// mock getting max turbo frequency list.
		mFetcher.On("GetMaxTurboFreqList", packageID).Return(nil, mError).Twice()

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which rely on msr
				packageCPUBaseFrequency.String(),
				packageTurboLimit.String(),

				// metrics which do not rely on msr
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},

			fetcher: mFetcher,

			logOnce: map[string]struct{}{},
		}

		require.NoError(t, p.parseConfig())

		// First call adds the error to the accumulator and logOnce map.
		p.addPerPackageMsrMetrics(acc, packageID)

		// Second call detects previous error in logOnce map and skips adding it to the accumulator.
		p.addPerPackageMsrMetrics(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 2)
		require.ErrorContains(t, acc.Errors[0], fmt.Sprintf("failed to get %q: %v", packageCPUBaseFrequency, msrNotInitErr))
		require.ErrorContains(t, acc.Errors[1], fmt.Sprintf("failed to get %q: %v", packageTurboLimit, msrNotInitErr))

		require.Len(t, p.logOnce, 2)
		require.Contains(t, p.logOnce, "msr_cpu_base_frequency")
		require.Contains(t, p.logOnce, "msr_max_turbo_frequency")
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithErrors", func(t *testing.T) {
		packageID := 0
		baseFreq := uint64(400)

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting CPU base frequency metric.
		mFetcher.On("GetCPUBaseFrequency", packageID).Return(baseFreq, nil).Once()

		// mock getting max turbo frequency list.
		mFetcher.On("GetMaxTurboFreqList", packageID).Return(nil, errors.New("mock error")).Once()

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which rely on msr
				packageCPUBaseFrequency.String(),
				packageTurboLimit.String(),

				// metrics which do not rely on msr
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addPerPackageMsrMetrics(acc, packageID)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID %v", packageTurboLimit, packageID))
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"cpu_base_frequency_mhz": baseFreq,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithoutErrors", func(t *testing.T) {
		packageID := 0
		baseFreq := uint64(400)
		maxTurboFreqList := []ptel.MaxTurboFreq{
			{
				Value:       1000,
				ActiveCores: 10,
			},
		}

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting CPU base frequency metric.
		mFetcher.On("GetCPUBaseFrequency", packageID).Return(baseFreq, nil).Once()

		// mock getting max turbo frequency list.
		mFetcher.On("GetMaxTurboFreqList", packageID).Return(maxTurboFreqList, nil).Once()

		p := &PowerStat{
			PackageMetrics: []string{
				// metrics which rely on msr
				packageCPUBaseFrequency.String(),
				packageTurboLimit.String(),

				// metrics which do not rely on msr
				packageCurrentPowerConsumption.String(),
				packageCurrentDramPowerConsumption.String(),
				packageThermalDesignPower.String(),
			},

			fetcher: mFetcher,
		}

		require.NoError(t, p.parseConfig())

		p.addPerPackageMsrMetrics(acc, packageID)

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 2)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"cpu_base_frequency_mhz": baseFreq,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"cpu_base_frequency_mhz": baseFreq,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"max_turbo_frequency_mhz": maxTurboFreqList[0].Value,
			},
			// tags
			map[string]string{
				"package_id":   strconv.Itoa(packageID),
				"active_cores": strconv.Itoa(int(maxTurboFreqList[0].ActiveCores)),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUFrequency(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU frequency metric.
		mFetcher.On("GetCPUFrequency", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUFrequency(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuFrequency, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		cpuID := 0
		coreID := 1
		packageID := 0
		cpuFreq := 800.001
		cpuFreqExp := 800.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting CPU frequency metric.
		mFetcher.On("GetCPUFrequency", cpuID).Return(cpuFreq, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUFrequency(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_frequency_mhz"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_frequency_mhz": cpuFreqExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUTemperature(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU temperature metric.
		mFetcher.On("GetCPUTemperature", cpuID).Return(uint64(0), errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUTemperature(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuTemperature, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		cpuTemp := uint64(25)

		mFetcher := &fetcherMock{}

		// mock getting cpu temperature metric.
		mFetcher.On("GetCPUTemperature", cpuID).Return(cpuTemp, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUTemperature(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasUIntField("powerstat_core", "cpu_temperature_celsius"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_temperature_celsius": cpuTemp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUC0StateResidency(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU C0 state residency metric.
		mFetcher.On("GetCPUC0StateResidency", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC0StateResidency(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC0StateResidency, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		cpuID := 0
		coreID := 1
		packageID := 0
		c0State := 10.1199
		c0StateExp := 10.12

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting CPU C0 state residency metric.
		mFetcher.On("GetCPUC0StateResidency", cpuID).Return(c0State, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC0StateResidency(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_c0_state_residency_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c0_state_residency_percent": c0StateExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUC1StateResidency(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU C1 state residency metric.
		mFetcher.On("GetCPUC1StateResidency", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC1StateResidency(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC1StateResidency, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		c1State := 10.1144
		c1StateExp := 10.11

		mFetcher := &fetcherMock{}

		// mock getting CPU C1 state residency metric.
		mFetcher.On("GetCPUC1StateResidency", cpuID).Return(c1State, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC1StateResidency(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_c1_state_residency_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c1_state_residency_percent": c1StateExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUC3StateResidency(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU C3 state residency metric.
		mFetcher.On("GetCPUC3StateResidency", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC3StateResidency(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC3StateResidency, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		c3State := 20.1178
		c3StateExp := 20.12

		mFetcher := &fetcherMock{}

		// mock getting CPU C3 state residency metric.
		mFetcher.On("GetCPUC3StateResidency", cpuID).Return(c3State, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC3StateResidency(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_c3_state_residency_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c3_state_residency_percent": c3StateExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUC6StateResidency(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU C6 state residency metric.
		mFetcher.On("GetCPUC6StateResidency", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC6StateResidency(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC6StateResidency, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		c6State := 9.115
		c6StateExp := 9.12

		mFetcher := &fetcherMock{}

		// mock getting CPU C6 state residency metric.
		mFetcher.On("GetCPUC6StateResidency", cpuID).Return(c6State, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC6StateResidency(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_c6_state_residency_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c6_state_residency_percent": c6StateExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUC7StateResidency(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU C7 state residency metric.
		mFetcher.On("GetCPUC7StateResidency", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC7StateResidency(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC7StateResidency, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		c7State := 9.1149
		c7StateExp := 9.11

		mFetcher := &fetcherMock{}

		// mock getting CPU C7 state residency metric.
		mFetcher.On("GetCPUC7StateResidency", cpuID).Return(c7State, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC7StateResidency(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_c7_state_residency_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c7_state_residency_percent": c7StateExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUBusyFrequency(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU busy frequency metric.
		mFetcher.On("GetCPUBusyFrequencyMhz", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUBusyFrequency(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuBusyFrequency, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		cpuBusyFreq := 800.119
		cpuBusyFreqExp := 800.12

		mFetcher := &fetcherMock{}

		// mock getting CPU busy frequency metric.
		mFetcher.On("GetCPUBusyFrequencyMhz", cpuID).Return(cpuBusyFreq, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUBusyFrequency(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_busy_frequency_mhz"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_busy_frequency_mhz": cpuBusyFreqExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUBusyCycles(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU busy cycles metric.
		mFetcher.On("GetCPUC0StateResidency", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUBusyCycles(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuBusyCycles, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		cpuBusyCycles := 10.1149
		cpuBusyCyclesExp := 10.11

		mFetcher := &fetcherMock{}

		// mock getting CPU C0 state residency metric.
		mFetcher.On("GetCPUC0StateResidency", cpuID).Return(cpuBusyCycles, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUBusyCycles(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_busy_cycles_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_busy_cycles_percent": cpuBusyCyclesExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUC0SubstateC01Percent(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU C01 metric.
		mFetcher.On("GetCPUC0SubstateC01Percent", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC0SubstateC01Percent(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC0SubstateC01Percent, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		c01Percent := 5.9229
		c01PercentExp := 5.92

		mFetcher := &fetcherMock{}

		// mock getting CPU C01 metric.
		mFetcher.On("GetCPUC0SubstateC01Percent", cpuID).Return(c01Percent, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC0SubstateC01Percent(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_c0_substate_c01_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c0_substate_c01_percent": c01PercentExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUC0SubstateC02Percent(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU C02 metric.
		mFetcher.On("GetCPUC0SubstateC02Percent", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC0SubstateC02Percent(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC0SubstateC02Percent, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		c02Percent := 0.001
		c02PercentExp := 0.0

		mFetcher := &fetcherMock{}

		// mock getting CPU C02 metric.
		mFetcher.On("GetCPUC0SubstateC02Percent", cpuID).Return(c02Percent, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC0SubstateC02Percent(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_c0_substate_c02_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c0_substate_c02_percent": c02PercentExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUC0SubstateC0WaitPercent(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU C0Wait metric.
		mFetcher.On("GetCPUC0SubstateC0WaitPercent", cpuID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC0SubstateC0WaitPercent(acc, cpuID, coreID, packageID)

		require.Empty(t, acc.GetTelegrafMetrics(), 0)
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for CPU ID %v", cpuC0SubstateC0WaitPercent, cpuID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		cpuID := 0
		coreID := 1
		packageID := 0
		c0WaitPercent := 0.995
		c0WaitPercentExp := 1.0

		mFetcher := &fetcherMock{}

		// mock getting CPU C0Wait metric.
		mFetcher.On("GetCPUC0SubstateC0WaitPercent", cpuID).Return(c0WaitPercent, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUC0SubstateC0WaitPercent(acc, cpuID, coreID, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_core", "cpu_c0_substate_c0_wait_percent"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_core",
			// fields
			map[string]interface{}{
				"cpu_c0_substate_c0_wait_percent": c0WaitPercentExp,
			},
			// tags
			map[string]string{
				"cpu_id":     strconv.Itoa(cpuID),
				"core_id":    strconv.Itoa(coreID),
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCurrentPackagePowerConsumption(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting current package power consumption metric.
		mFetcher.On("GetCurrentPackagePowerConsumptionWatts", packageID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCurrentPackagePower(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID %v", packageCurrentPowerConsumption, packageID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		packageID := 0
		currPower := float64(30.1999)
		currPowerRounded := float64(30.2)

		mFetcher := &fetcherMock{}

		// mock getting current package power consumption metric.
		mFetcher.On("GetCurrentPackagePowerConsumptionWatts", packageID).Return(currPower, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCurrentPackagePower(acc, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_package", "current_power_consumption_watts"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"current_power_consumption_watts": currPowerRounded,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCurrentDramPowerConsumption(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting current dram power consumption metric.
		mFetcher.On("GetCurrentDramPowerConsumptionWatts", packageID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCurrentDramPower(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID %v", packageCurrentDramPowerConsumption, packageID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		packageID := 0
		currPower := float64(30.8235)
		currPowerRounded := float64(30.82)

		mFetcher := &fetcherMock{}

		// mock getting current dram power consumption metric.
		mFetcher.On("GetCurrentDramPowerConsumptionWatts", packageID).Return(currPower, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCurrentDramPower(acc, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_package", "current_dram_power_consumption_watts"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"current_dram_power_consumption_watts": currPowerRounded,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddThermalDesignPower(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting package thermal design power metric.
		mFetcher.On("GetPackageThermalDesignPowerWatts", packageID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addThermalDesignPower(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID %v", packageThermalDesignPower, packageID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Rounded", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		packageID := 0
		tdp := float64(80.1999)
		tdpRounded := float64(80.2)

		mFetcher := &fetcherMock{}

		// mock getting package thermal design power metric.
		mFetcher.On("GetPackageThermalDesignPowerWatts", packageID).Return(tdp, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addThermalDesignPower(acc, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasFloatField("powerstat_package", "thermal_design_power_watts"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"thermal_design_power_watts": tdpRounded,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddCPUBaseFrequency(t *testing.T) {
	t.Run("FailedToGetMetric", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		packageID := 0

		mFetcher := &fetcherMock{}

		// mock getting CPU base frequency metric.
		mFetcher.On("GetCPUBaseFrequency", packageID).Return(uint64(0), errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUBaseFrequency(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID %v", packageCPUBaseFrequency, packageID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		packageID := 0
		baseFreq := uint64(700)

		mFetcher := &fetcherMock{}

		// mock getting CPU base frequency metric.
		mFetcher.On("GetCPUBaseFrequency", packageID).Return(baseFreq, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addCPUBaseFrequency(acc, packageID)

		require.Len(t, acc.GetTelegrafMetrics(), 1)
		require.True(t, acc.HasUIntField("powerstat_package", "cpu_base_frequency_mhz"))
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"cpu_base_frequency_mhz": baseFreq,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddUncoreFrequency(t *testing.T) {
	packageID, dieID := 1, 0

	t.Run("FailedToGetDieIDs", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting die IDs for package ID.
		mFetcher.On("GetPackageDieIDs", packageID).Return(nil, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addUncoreFrequency(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(
			t,
			acc.FirstError(),
			fmt.Sprintf("failed to get die IDs for package ID %v", packageID),
		)
	})

	t.Run("FailedToGetInitialLimits", func(t *testing.T) {
		currMin := 500.0
		currMax := 2500.0
		curr := 1000.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting die IDs for package ID.
		mFetcher.On("GetPackageDieIDs", packageID).Return([]int{dieID}, nil).Once()

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(800.0, nil).Once()

		// mock getting initial maximum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMax", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(currMin, nil).Once()

		// mock getting custom maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", packageID, dieID).Return(currMax, nil).Once()

		// mock getting current uncore frequency value.
		mFetcher.On("GetCurrentUncoreFrequency", packageID, dieID).Return(curr, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addUncoreFrequency(acc, packageID)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(
			t,
			acc.FirstError(),
			fmt.Sprintf("failed to get initial uncore frequency limits for package ID %v and die ID %v", packageID, dieID),
		)
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": currMin,
				"uncore_frequency_limit_mhz_max": currMax,
				"uncore_frequency_mhz_cur":       uint64(curr),
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
				"type":       "current",
				"die":        strconv.Itoa(dieID),
			},
		)
		mFetcher.AssertExpectations(t)
	})

	t.Run("FailedToGetCurrentValues", func(t *testing.T) {
		initMin := 300.0
		initMax := 1200.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting die IDs for package ID.
		mFetcher.On("GetPackageDieIDs", packageID).Return([]int{dieID}, nil).Once()

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(initMin, nil).Once()

		// mock getting initial maximum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMax", packageID, dieID).Return(initMax, nil).Once()

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(500.0, nil).Once()

		// mock getting custom maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", packageID, dieID).Return(1300.0, nil).Once()

		// mock getting current uncore frequency value.
		mFetcher.On("GetCurrentUncoreFrequency", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addUncoreFrequency(acc, packageID)

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(
			t,
			acc.FirstError(),
			fmt.Sprintf("failed to get current uncore frequency values for package ID %v and die ID %v", packageID, dieID),
		)
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": initMin,
				"uncore_frequency_limit_mhz_max": initMax,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
				"type":       "initial",
				"die":        strconv.Itoa(dieID),
			},
		)
		mFetcher.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		initMin := 300.0
		initMax := 1200.0
		currMin := 500.0
		currMax := 2500.0
		curr := 1000.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting die IDs for package ID.
		mFetcher.On("GetPackageDieIDs", packageID).Return([]int{dieID}, nil).Once()

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(initMin, nil).Once()

		// mock getting initial maximum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMax", packageID, dieID).Return(initMax, nil).Once()

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(currMin, nil).Once()

		// mock getting custom maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", packageID, dieID).Return(currMax, nil).Once()

		// mock getting current uncore frequency value.
		mFetcher.On("GetCurrentUncoreFrequency", packageID, dieID).Return(curr, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addUncoreFrequency(acc, packageID)

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 2)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": initMin,
				"uncore_frequency_limit_mhz_max": initMax,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
				"type":       "initial",
				"die":        strconv.Itoa(dieID),
			},
		)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": currMin,
				"uncore_frequency_limit_mhz_max": currMax,
				"uncore_frequency_mhz_cur":       uint64(curr),
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
				"type":       "current",
				"die":        strconv.Itoa(dieID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddUncoreFrequencyInitialLimits(t *testing.T) {
	packageID, dieID := 0, 0

	t.Run("WithModuleNotInitializedError", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		uncoreFreqErr := &ptel.ModuleNotInitializedError{Name: "uncore_frequency"}
		mError := fmt.Errorf("mock error: %w", uncoreFreqErr)
		mFetcher := &fetcherMock{}

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(0.0, mError).Twice()

		p := &PowerStat{
			fetcher: mFetcher,

			logOnce: map[string]struct{}{},
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		// First call adds the error to the accumulator and logOnce map.
		p.addUncoreFrequencyInitialLimits(acc, packageID, dieID)

		// Second call detects previous error in logOnce map and skips adding it to the accumulator.
		p.addUncoreFrequencyInitialLimits(acc, packageID, dieID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q initial limits", packageUncoreFrequency))
		require.ErrorContains(t, acc.FirstError(), uncoreFreqErr.Error())

		require.Len(t, p.logOnce, 1)
		require.Contains(t, p.logOnce, fmt.Sprintf("%s_%s_initial", "uncore_frequency", packageUncoreFrequency))
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithError", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,

			logOnce: map[string]struct{}{},
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addUncoreFrequencyInitialLimits(acc, packageID, dieID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(
			t,
			acc.FirstError(),
			fmt.Sprintf("failed to get initial uncore frequency limits for package ID %v and die ID %v", packageID, dieID),
		)
		require.Empty(t, p.logOnce)
		mFetcher.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		initMin := 300.0
		initMax := 1200.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(initMin, nil).Once()

		// mock getting initial maximum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMax", packageID, dieID).Return(initMax, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addUncoreFrequencyInitialLimits(acc, packageID, dieID)

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": initMin,
				"uncore_frequency_limit_mhz_max": initMax,
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
				"type":       "initial",
				"die":        strconv.Itoa(dieID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddUncoreFrequencyCurrentValues(t *testing.T) {
	packageID, dieID := 0, 0

	t.Run("WithModuleNotInitializedError", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		uncoreFreqErr := &ptel.ModuleNotInitializedError{Name: "uncore_frequency"}
		mError := fmt.Errorf("mock error: %w", uncoreFreqErr)
		mFetcher := &fetcherMock{}

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(0.0, mError).Twice()

		p := &PowerStat{
			fetcher: mFetcher,

			logOnce: map[string]struct{}{},
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		// First call adds the error to the accumulator and logOnce map.
		p.addUncoreFrequencyCurrentValues(acc, packageID, dieID)

		// Second call detects previous error in logOnce map and skips adding it to the accumulator.
		p.addUncoreFrequencyCurrentValues(acc, packageID, dieID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q current value and limits", packageUncoreFrequency))
		require.ErrorContains(t, acc.FirstError(), uncoreFreqErr.Error())

		require.Len(t, p.logOnce, 1)
		require.Contains(t, p.logOnce, fmt.Sprintf("%s_%s_current", "uncore_frequency", packageUncoreFrequency))
		mFetcher.AssertExpectations(t)
	})

	t.Run("WithError", func(t *testing.T) {
		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,

			logOnce: map[string]struct{}{},
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addUncoreFrequencyCurrentValues(acc, packageID, dieID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(
			t,
			acc.FirstError(),
			fmt.Sprintf("failed to get current uncore frequency values for package ID %v and die ID %v", packageID, dieID),
		)
		require.Empty(t, p.logOnce)
		mFetcher.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		currMin := 500.0
		currMax := 2500.0
		curr := 1000.0

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(currMin, nil).Once()

		// mock getting custom maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", packageID, dieID).Return(currMax, nil).Once()

		// mock getting current uncore frequency value.
		mFetcher.On("GetCurrentUncoreFrequency", packageID, dieID).Return(curr, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addUncoreFrequencyCurrentValues(acc, packageID, dieID)

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 1)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"uncore_frequency_limit_mhz_min": currMin,
				"uncore_frequency_limit_mhz_max": currMax,
				"uncore_frequency_mhz_cur":       uint64(curr),
			},
			// tags
			map[string]string{
				"package_id": strconv.Itoa(packageID),
				"type":       "current",
				"die":        strconv.Itoa(dieID),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}

func TestGetUncoreFreqInitialLimits(t *testing.T) {
	packageID, dieID := 0, 0

	t.Run("FailsToGetInitialMinLimit", func(t *testing.T) {
		mFetcher := &fetcherMock{}

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		initMin, initMax, err := getUncoreFreqInitialLimits(mFetcher, packageID, dieID)

		require.ErrorContains(t, err, "failed to get initial minimum uncore frequency limit")
		require.Zero(t, initMin)
		require.Zero(t, initMax)
		mFetcher.AssertExpectations(t)
	})

	t.Run("FailsToGetInitialMaxLimit", func(t *testing.T) {
		mFetcher := &fetcherMock{}

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(800.0, nil).Once()

		// mock getting initial maximum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMax", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		initMin, initMax, err := getUncoreFreqInitialLimits(mFetcher, packageID, dieID)

		require.ErrorContains(t, err, "failed to get initial maximum uncore frequency limit")
		require.Zero(t, initMin)
		require.Zero(t, initMax)
		mFetcher.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		initMinExp := 300.0
		initMaxExp := 1500.0

		mFetcher := &fetcherMock{}

		// mock getting initial minimum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMin", packageID, dieID).Return(initMinExp, nil).Once()

		// mock getting initial maximum uncore frequency limit.
		mFetcher.On("GetInitialUncoreFrequencyMax", packageID, dieID).Return(initMaxExp, nil).Once()

		initMin, initMax, err := getUncoreFreqInitialLimits(mFetcher, packageID, dieID)

		require.NoError(t, err)
		require.Equal(t, initMinExp, initMin)
		require.Equal(t, initMaxExp, initMax)
		mFetcher.AssertExpectations(t)
	})
}

func TestGetUncoreFreqCurrentValues(t *testing.T) {
	packageID, dieID := 0, 0

	t.Run("FailsToGetCurrentMinLimit", func(t *testing.T) {
		mFetcher := &fetcherMock{}

		// mock getting current minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		values, err := getUncoreFreqCurrentValues(mFetcher, packageID, dieID)

		require.ErrorContains(t, err, "failed to get current minimum uncore frequency limit")
		require.Equal(t, uncoreFreqValues{}, values)
		mFetcher.AssertExpectations(t)
	})

	t.Run("FailsToGetCurrentMaxLimit", func(t *testing.T) {
		mFetcher := &fetcherMock{}

		// mock getting current minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(1000.0, nil).Once()

		// mock getting current maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		values, err := getUncoreFreqCurrentValues(mFetcher, packageID, dieID)

		require.ErrorContains(t, err, "failed to get current maximum uncore frequency limit")
		require.Equal(t, uncoreFreqValues{}, values)
		mFetcher.AssertExpectations(t)
	})

	t.Run("FailsToGetCurrentValue", func(t *testing.T) {
		mFetcher := &fetcherMock{}

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(1000.0, nil).Once()

		// mock getting custom maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", packageID, dieID).Return(2000.0, nil).Once()

		// mock getting current uncore frequency value.
		mFetcher.On("GetCurrentUncoreFrequency", packageID, dieID).Return(0.0, errors.New("mock error")).Once()

		values, err := getUncoreFreqCurrentValues(mFetcher, packageID, dieID)

		require.ErrorContains(t, err, "failed to get current uncore frequency")
		require.Equal(t, uncoreFreqValues{}, values)
		mFetcher.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		min := 500.0
		max := 1500.0
		current := 750.0

		uncoreFreqValExp := uncoreFreqValues{
			currMin: min,
			currMax: max,
			curr:    current,
		}

		mFetcher := &fetcherMock{}

		// mock getting custom minimum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMin", packageID, dieID).Return(min, nil).Once()

		// mock getting custom maximum uncore frequency limit.
		mFetcher.On("GetCustomizedUncoreFrequencyMax", packageID, dieID).Return(max, nil).Once()

		// mock getting current uncore frequency value.
		mFetcher.On("GetCurrentUncoreFrequency", packageID, dieID).Return(current, nil).Once()

		values, err := getUncoreFreqCurrentValues(mFetcher, packageID, dieID)

		require.NoError(t, err)
		require.Equal(t, uncoreFreqValExp, values)
		mFetcher.AssertExpectations(t)
	})
}

func TestAddMaxTurboFreqLimits(t *testing.T) {
	t.Run("FailedToGetMetricModuleNotInitializedError", func(t *testing.T) {
		packageID := 1

		acc := &testutil.Accumulator{}

		mErr := &ptel.ModuleNotInitializedError{Name: "msr"}
		mFetcher := &fetcherMock{}

		// mock getting max turbo frequency list.
		mFetcher.On("GetMaxTurboFreqList", packageID).Return(nil, mErr).Twice()

		p := &PowerStat{
			fetcher: mFetcher,

			logOnce: map[string]struct{}{},
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		// First call adds the error to the accumulator and key to logOnce map.
		p.addMaxTurboFreqLimits(acc, packageID)

		// Second call detects previous error in logOnce map and skips adding it to the accumulator.
		p.addMaxTurboFreqLimits(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q: %v", packageTurboLimit, mErr))

		require.Len(t, p.logOnce, 1)
		require.Contains(t, p.logOnce, "msr_max_turbo_frequency")

		mFetcher.AssertExpectations(t)
	})

	t.Run("FailedToGetMetric", func(t *testing.T) {
		packageID := 1

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting max turbo frequency list.
		mFetcher.On("GetMaxTurboFreqList", packageID).Return(nil, errors.New("mock error")).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addMaxTurboFreqLimits(acc, packageID)

		require.Empty(t, acc.GetTelegrafMetrics())
		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.FirstError(), fmt.Sprintf("failed to get %q for package ID %v", packageTurboLimit, packageID))
		mFetcher.AssertExpectations(t)
	})

	t.Run("CPUIsHybird", func(t *testing.T) {
		packageID := 1

		maxTurboFreqList := []ptel.MaxTurboFreq{
			{
				Value:       1000,
				ActiveCores: 10,
				Secondary:   true,
			},
			{
				Value:       2000,
				ActiveCores: 20,
			},
		}

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting max turbo frequency list.
		mFetcher.On("GetMaxTurboFreqList", packageID).Return(maxTurboFreqList, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addMaxTurboFreqLimits(acc, packageID)

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 2)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"max_turbo_frequency_mhz": maxTurboFreqList[0].Value,
			},
			// tags
			map[string]string{
				"package_id":   strconv.Itoa(packageID),
				"active_cores": strconv.Itoa(int(maxTurboFreqList[0].ActiveCores)),
				"hybrid":       "secondary",
			},
		)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"max_turbo_frequency_mhz": maxTurboFreqList[1].Value,
			},
			// tags
			map[string]string{
				"package_id":   strconv.Itoa(packageID),
				"active_cores": strconv.Itoa(int(maxTurboFreqList[1].ActiveCores)),
				"hybrid":       "primary",
			},
		)
		mFetcher.AssertExpectations(t)
	})

	t.Run("CPUIsNotHybrid", func(t *testing.T) {
		packageID := 1

		maxTurboFreqList := []ptel.MaxTurboFreq{
			{
				Value:       1000,
				ActiveCores: 10,
			},
			{
				Value:       2000,
				ActiveCores: 20,
			},
		}

		acc := &testutil.Accumulator{}

		mFetcher := &fetcherMock{}

		// mock getting max turbo frequency list.
		mFetcher.On("GetMaxTurboFreqList", packageID).Return(maxTurboFreqList, nil).Once()

		p := &PowerStat{
			fetcher: mFetcher,
		}

		require.Empty(t, acc.GetTelegrafMetrics())

		p.addMaxTurboFreqLimits(acc, packageID)

		require.Empty(t, acc.Errors)
		require.Len(t, acc.GetTelegrafMetrics(), 2)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"max_turbo_frequency_mhz": maxTurboFreqList[0].Value,
			},
			// tags
			map[string]string{
				"package_id":   strconv.Itoa(packageID),
				"active_cores": strconv.Itoa(int(maxTurboFreqList[0].ActiveCores)),
			},
		)
		acc.AssertContainsTaggedFields(
			t,
			// measurement
			"powerstat_package",
			// fields
			map[string]interface{}{
				"max_turbo_frequency_mhz": maxTurboFreqList[1].Value,
			},
			// tags
			map[string]string{
				"package_id":   strconv.Itoa(packageID),
				"active_cores": strconv.Itoa(int(maxTurboFreqList[1].ActiveCores)),
			},
		)
		mFetcher.AssertExpectations(t)
	})
}
