// +build linux

package intel_powerstat

import (
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/data"
	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/mocks"
	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/services"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInitPlugin(t *testing.T) {
	cores := []string{"cpu0", "cpu1", "cpu2", "cpu3"}
	power, fsMock, _, _ := getPowerWithMockedServices()

	fsMock.On("GetCPUInfoStats", mock.Anything).
		Return(nil, errors.New("error getting cpu stats")).Once()
	require.Error(t, power.Init())

	fsMock.On("GetCPUInfoStats", mock.Anything).
		Return(make(map[string]*data.CPUInfo), nil).Once()
	require.Error(t, power.Init())

	fsMock.On("GetCPUInfoStats", mock.Anything).
		Return(map[string]*data.CPUInfo{"0": {
			VendorID:  "GenuineIntel",
			CPUFamily: "test",
		}}, nil).Once()
	require.Error(t, power.Init())

	fsMock.On("GetStringsMatchingPatternOnPath", mock.Anything).
		Return(cores, nil).Once().
		On("GetCPUInfoStats", mock.Anything).
		Return(map[string]*data.CPUInfo{"0": {
			VendorID:  "GenuineIntel",
			CPUFamily: "6",
		}}, nil)
	// Verify MSR service initialization.
	power.coreCPUFrequency = true
	require.NoError(t, power.Init())
	fsMock.AssertCalled(t, "GetStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, len(cores), len(power.msr.GetCpuIDs()))

	fsMock.On("GetStringsMatchingPatternOnPath", mock.Anything).
		Return(nil, errors.New("error during GetStringsMatchingPatternOnPath")).Once()

	// In case of an error when fetching cpu cores plugin should proceed with execution.
	require.NoError(t, power.Init())
	fsMock.AssertCalled(t, "GetStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, 0, len(power.msr.GetCpuIDs()))
}

func TestParseCPUMetricsConfig(t *testing.T) {
	power, _, _, _ := getPowerWithMockedServices()
	disableCoreMetrics(power)

	power.CPUMetrics = []string{
		"cpu_frequency", "cpu_c1_state_residency", "cpu_c6_state_residency", "cpu_busy_cycles", "cpu_temperature",
		"cpu_busy_frequency",
	}
	power.ParseCPUMetricsConfig()
	verifyCoreMetrics(t, power, true)
	disableCoreMetrics(power)
	verifyCoreMetrics(t, power, false)

	power.CPUMetrics = []string{}
	power.ParseCPUMetricsConfig()
	require.Equal(t, false, power.perCoreMetrics)

	power.CPUMetrics = []string{"cpu_c6_state_residency", "#@$sdkjdfsdf3@", "1pu_c1_state_residency"}
	power.ParseCPUMetricsConfig()
	require.Equal(t, false, power.c1CoreStateResidency)
	require.Equal(t, true, power.c6CoreStateResidency)
	disableCoreMetrics(power)
	verifyCoreMetrics(t, power, false)

	power.CPUMetrics = []string{"#@$sdkjdfsdf3@", "1pu_c1_state_residency", "123"}
	power.ParseCPUMetricsConfig()
	verifyCoreMetrics(t, power, false)
}

func verifyCoreMetrics(t *testing.T, power *PowerStat, enabled bool) {
	require.Equal(t, enabled, power.coreCPUFrequency)
	require.Equal(t, enabled, power.c1CoreStateResidency)
	require.Equal(t, enabled, power.c6CoreStateResidency)
	require.Equal(t, enabled, power.processorBusyCycles)
	require.Equal(t, enabled, power.processorBusyFrequency)
	require.Equal(t, enabled, power.coreCPUTemperature)
}

func TestGather(t *testing.T) {
	var acc testutil.Accumulator
	packageIDs := []string{"0", "1"}
	coreIDs := []string{"0", "1", "2", "3"}
	power, _, raplMock, msrMock := getPowerWithMockedServices()
	prepareCPUInfo(power, coreIDs, packageIDs)
	enableCoreMetrics(power)
	power.skipFirstIteration = false

	raplMock.On("InitializeRaplData", mock.Anything).
		On("GetSocketIDs").Return(packageIDs).
		On("RetrieveAndCalculateData", mock.Anything).Return(nil).Times(len(packageIDs)).
		On("GetConstraintMaxPower", mock.Anything).Return(546783852.3, nil).
		On("GetCurrentPackagePowerConsumption", mock.Anything).Return(13213852.2).
		On("GetCurrentDramPowerConsumption", mock.Anything).Return(784552.0)
	msrMock.On("GetCpuIDs").Return(coreIDs).
		On("OpenAndReadMsr", mock.Anything).Return(nil).
		On("GetThrottleTemperature", mock.Anything).Return(uint64(434643)).
		On("GetTemperature", mock.Anything).Return(uint64(1231541)).
		On("GetTimestampDelta", mock.Anything).Return(uint64(7633345)).
		On("GetC6Delta", mock.Anything).Return(uint64(12634345)).
		On("GetC3Delta", mock.Anything).Return(uint64(978956)).
		On("GetC7Delta", mock.Anything).Return(uint64(1235222)).
		On("GetMperfDelta", mock.Anything).Return(uint64(98457123)).
		On("GetAperfDelta", mock.Anything).Return(uint64(14313123)).
		On("GetReadDate", mock.Anything).Return(int64(323221)).
		On("SetReadDate", mock.Anything, mock.Anything).
		On("RetrieveCPUFrequencyForCore", mock.Anything).Return(1200000.2, int64(0), nil)

	power.perCoreMetrics = true

	require.NoError(t, power.Gather(&acc))
	// Number of global metrics   : 3
	// Number of per core metrics : 6
	require.Equal(t, 3*len(packageIDs)+6*len(coreIDs), len(acc.GetTelegrafMetrics()))
}

func TestAddGlobalMetricsNegative(t *testing.T) {
	var acc testutil.Accumulator
	sockets := []string{"0", "1"}
	power, _, raplMock, _ := getPowerWithMockedServices()
	power.skipFirstIteration = false
	raplMock.On("InitializeRaplData", mock.Anything).Once().
		On("GetSocketIDs").Return(sockets).Once().
		On("RetrieveAndCalculateData", mock.Anything).Return(errors.New("error while calculating data")).Times(len(sockets))

	power.addGlobalMetrics(&acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
	raplMock.AssertNumberOfCalls(t, "RetrieveAndCalculateData", len(sockets))

	raplMock.On("InitializeRaplData", mock.Anything).Once().
		On("GetSocketIDs").Return(make([]string, 0)).Once()

	power.addGlobalMetrics(&acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
	raplMock.AssertNotCalled(t, "RetrieveAndCalculateData")

	raplMock.On("InitializeRaplData", mock.Anything).Once().
		On("GetSocketIDs").Return(sockets).
		On("RetrieveAndCalculateData", mock.Anything).Return(nil).Once().
		On("RetrieveAndCalculateData", mock.Anything).Return(errors.New("error while calculating data")).Once().
		On("GetConstraintMaxPower", mock.Anything).Return(12313851.5, nil).Twice().
		On("GetCurrentPackagePowerConsumption", mock.Anything).Return(13213852.2).Once().
		On("GetCurrentDramPowerConsumption", mock.Anything).Return(784552.0).Once()

	power.addGlobalMetrics(&acc)
	require.Equal(t, 3, len(acc.GetTelegrafMetrics()))
}

func TestAddGlobalMetricsPositive(t *testing.T) {
	var acc testutil.Accumulator
	sockets := []string{"0", "1"}
	maxPower := 546783852.9
	socketCurrentEnergy := 3644574.4
	dramCurrentEnergy := 124234872.5
	power, _, raplMock, _ := getPowerWithMockedServices()
	power.skipFirstIteration = false

	raplMock.On("InitializeRaplData", mock.Anything).
		On("GetSocketIDs").Return(sockets).
		On("RetrieveAndCalculateData", mock.Anything).Return(nil).Times(len(sockets)).
		On("GetConstraintMaxPower", mock.Anything).Return(maxPower, nil).Twice().
		On("GetCurrentPackagePowerConsumption", mock.Anything).Return(socketCurrentEnergy).
		On("GetCurrentDramPowerConsumption", mock.Anything).Return(dramCurrentEnergy)

	power.addGlobalMetrics(&acc)
	require.Equal(t, 6, len(acc.GetTelegrafMetrics()))

	expectedResults := getGlobalMetrics(maxPower, socketCurrentEnergy, dramCurrentEnergy)
	for _, test := range expectedResults {
		acc.AssertContainsTaggedFields(t, "powerstat_package", test.fields, test.tags)
	}
}

func TestAddMetricsForSingleCoreNegative(t *testing.T) {
	var wg sync.WaitGroup
	var acc testutil.Accumulator
	core := "0"
	power, _, _, msrMock := getPowerWithMockedServices()

	msrMock.On("OpenAndReadMsr", core).Return(errors.New("error reading MSR file")).Once()

	// Skip generating metric for CPU frequency.
	power.coreCPUFrequency = false

	wg.Add(1)
	power.addMetricsForSingleCore(core, &acc, &wg)
	wg.Wait()

	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddCPUFrequencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	frequency := 1200000.2
	power, _, _, msrMock := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)

	msrMock.On("RetrieveCPUFrequencyForCore", mock.Anything).
		Return(float64(0), int64(0), errors.New("error on reading file")).Once()

	power.addCPUFrequencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))

	msrMock.On("RetrieveCPUFrequencyForCore", mock.Anything).Return(frequency, int64(0), nil).Once()

	power.addCPUFrequencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedFrequency := services.RoundFloatToNearestTwoDecimalPlaces(services.ConvertKiloHertzToMegaHertz(frequency))
	expectedMetric := getPowerCoreMetric("cpu_frequency", "MHz", expectedFrequency, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)
}

func TestAddCoreCPUTemperatureMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, _, _, msrMock := getPowerWithMockedServices()
	throttleTemp := uint64(343453434)
	temp := uint64(312312)
	expectedTemp := throttleTemp - temp
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)

	msrMock.On("GetThrottleTemperature", mock.Anything).Return(throttleTemp).Once().
		On("GetTemperature", mock.Anything).Return(temp).Once()
	power.addCoreCPUTemperatureMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_temperature", "celsius_degrees", expectedTemp, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)
}

func TestAddC6StateResidencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, _, _, msrMock := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	tscDelta := uint64(2342341123)
	c6Delta := uint64(213233)
	expectedC6 := services.RoundFloatToNearestTwoDecimalPlaces(percentageMultiplier *
		float64(c6Delta) / float64(tscDelta))

	msrMock.On("GetTimestampDelta", mock.Anything).Return(tscDelta).Twice().
		On("GetC6Delta", mock.Anything).Return(c6Delta).Once()
	power.addC6StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_c6_state_residency", "percentage", expectedC6, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	acc.ClearMetrics()
	msrMock.On("GetTimestampDelta", mock.Anything).Return(uint64(0)).Once().
		On("GetC6Delta", mock.Anything).Return(c6Delta).Once()
	power.addC6StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddProcessorBusyCyclesMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, _, _, msrMock := getPowerWithMockedServices()
	mperfDelta := uint64(1233234)
	tscDelta := uint64(656434)
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	expectedBusyCycles := services.RoundFloatToNearestTwoDecimalPlaces(percentageMultiplier * float64(mperfDelta) /
		float64(tscDelta))

	msrMock.On("GetTimestampDelta", mock.Anything).Return(tscDelta).Twice().
		On("GetMperfDelta", mock.Anything).Return(mperfDelta).Once()
	power.addProcessorBusyCyclesMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_busy_cycles", "percentage", expectedBusyCycles, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	acc.ClearMetrics()
	msrMock.On("GetTimestampDelta", mock.Anything).Return(uint64(0)).Once().
		On("GetMperfDelta", mock.Anything).Return(mperfDelta).Once()
	power.addProcessorBusyCyclesMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddProcessorBusyFrequencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, _, _, msrMock := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	power.skipFirstIteration = false

	msrMock.
		On("GetTimestampDelta", mock.Anything).Return(uint64(7633345)).
		On("GetMperfDelta", mock.Anything).Return(uint64(98457123)).Twice().
		On("GetAperfDelta", mock.Anything).Return(uint64(14313123)).
		On("GetReadDate", mock.Anything).Return(int64(323221)).
		On("SetReadDate", mock.Anything, mock.Anything)
	power.addProcessorBusyFrequencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	acc.ClearMetrics()
	msrMock.
		On("GetMperfDelta", mock.Anything).Return(uint64(0))
	power.addProcessorBusyFrequencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddC1StateResidencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, _, _, msrMock := getPowerWithMockedServices()
	tscDelta := uint64(43512313)
	mperfDelta := uint64(112323)
	c3Delta := uint64(23233)
	c6Delta := uint64(2998043)
	c7Delta := uint64(3434323)
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	c1 := tscDelta - mperfDelta - c3Delta - c6Delta - c7Delta
	expectedC1 := services.RoundFloatToNearestTwoDecimalPlaces(percentageMultiplier * float64(c1) / float64(tscDelta))

	msrMock.
		On("GetTimestampDelta", mock.Anything).Return(tscDelta).Once().
		On("GetC6Delta", mock.Anything).Return(c6Delta).
		On("GetC3Delta", mock.Anything).Return(c3Delta).
		On("GetC7Delta", mock.Anything).Return(c7Delta).
		On("GetMperfDelta", mock.Anything).Return(mperfDelta)
	power.addC1StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_c1_state_residency", "percentage", expectedC1, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	acc.ClearMetrics()
	msrMock.
		On("GetTimestampDelta", mock.Anything).Return(uint64(0)).Once()
	power.addC1StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddThermalDesignPowerMetric(t *testing.T) {
	var acc testutil.Accumulator
	sockets := []string{"0"}
	maxPower := 195720672.1
	power, _, raplMock, _ := getPowerWithMockedServices()

	raplMock.On("GetConstraintMaxPower", mock.Anything).
		Return(float64(0), errors.New("GetConstraintMaxPower error")).Once().
		On("GetConstraintMaxPower", mock.Anything).Return(maxPower, nil).Once()

	power.addThermalDesignPowerMetric(sockets[0], &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))

	power.addThermalDesignPowerMetric(sockets[0], &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedTDP := services.RoundFloatToNearestTwoDecimalPlaces(services.ConvertMicroWattToWatt(maxPower))
	expectedMetric := getPowerGlobalMetric("thermal_design_power", "Watt", expectedTDP, sockets[0])
	acc.AssertContainsTaggedFields(t, "powerstat_package", expectedMetric.fields, expectedMetric.tags)
}

func getGlobalMetrics(maxPower float64, socketCurrentEnergy float64, dramCurrentEnergy float64) []struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	return []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		getPowerGlobalMetric("thermal_design_power", "Watt", services.RoundFloatToNearestTwoDecimalPlaces(services.ConvertMicroWattToWatt(maxPower)), "0"),
		getPowerGlobalMetric("thermal_design_power", "Watt", services.RoundFloatToNearestTwoDecimalPlaces(services.ConvertMicroWattToWatt(maxPower)), "1"),
		getPowerGlobalMetric("current_power_consumption", "Watt", services.RoundFloatToNearestTwoDecimalPlaces(socketCurrentEnergy), "0"),
		getPowerGlobalMetric("current_power_consumption", "Watt", services.RoundFloatToNearestTwoDecimalPlaces(socketCurrentEnergy), "1"),
		getPowerGlobalMetric("current_dram_power_consumption", "Watt", services.RoundFloatToNearestTwoDecimalPlaces(dramCurrentEnergy), "0"),
		getPowerGlobalMetric("current_dram_power_consumption", "Watt", services.RoundFloatToNearestTwoDecimalPlaces(dramCurrentEnergy), "1"),
	}
}

func getPowerCoreMetric(name string, unit string, value interface{}, coreID string, packageID string, cpuID string) struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	return getPowerMetric(name, unit, value, map[string]string{"package_id": packageID, "core_id": coreID, "cpu_id": cpuID})
}

func getPowerGlobalMetric(name string, unit string, value interface{}, socketID string) struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	return getPowerMetric(name, unit, value, map[string]string{"package_id": socketID})
}

func getPowerMetric(name string, unit string, value interface{}, tags map[string]string) struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	return struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		map[string]interface{}{
			"name":  name,
			"unit":  unit,
			"value": value,
		},
		tags,
	}
}

func prepareCPUInfoForSingleCPU(power *PowerStat, cpuID string, coreID string, packageID string) {
	power.cpuInfo = make(map[string]*data.CPUInfo)
	power.cpuInfo[cpuID] = &data.CPUInfo{
		PhysicalID: packageID,
		CoreID:     coreID,
		CPUID:      cpuID,
	}
}

func prepareCPUInfo(power *PowerStat, coreIDs []string, packageIDs []string) {
	power.cpuInfo = make(map[string]*data.CPUInfo)
	currentCPU := 0
	for _, packageID := range packageIDs {
		for _, coreID := range coreIDs {
			cpuID := strconv.Itoa(currentCPU)
			power.cpuInfo[cpuID] = &data.CPUInfo{
				PhysicalID: packageID,
				CPUID:      cpuID,
				CoreID:     coreID,
			}
			currentCPU++
		}
	}
}

func enableCoreMetrics(power *PowerStat) {
	power.perCoreMetrics = true
	power.c1CoreStateResidency = true
	power.c6CoreStateResidency = true
	power.coreCPUTemperature = true
	power.processorBusyFrequency = true
	power.coreCPUFrequency = true
	power.processorBusyCycles = true
}

func disableCoreMetrics(power *PowerStat) {
	power.perCoreMetrics = false
	power.c1CoreStateResidency = false
	power.c6CoreStateResidency = false
	power.coreCPUTemperature = false
	power.processorBusyFrequency = false
	power.coreCPUFrequency = false
	power.processorBusyCycles = false
}

func getPowerWithMockedServices() (*PowerStat, *mocks.FileService, *mocks.RaplService, *mocks.MsrService) {
	fsMock := &mocks.FileService{}
	msrMock := &mocks.MsrService{}
	raplMock := &mocks.RaplService{}
	logger := testutil.Logger{Name: "PowerPluginTest"}
	p := NewPowerStat(fsMock)
	p.Log = logger
	p.fs = fsMock
	p.rapl = raplMock
	p.msr = msrMock

	return p, fsMock, raplMock, msrMock
}
