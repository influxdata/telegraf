//go:build linux

package intel_powerstat

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

type MockServices struct {
	fs   *mockFileService
	msr  *mockMsrService
	rapl *mockRaplService
}

func TestInitPlugin(t *testing.T) {
	cores := []string{"cpu0", "cpu1", "cpu2", "cpu3"}
	power, mockServices := getPowerWithMockedServices()

	mockServices.fs.On("getCPUInfoStats", mock.Anything).
		Return(nil, errors.New("error getting cpu stats")).Once()
	require.Error(t, power.Init())

	mockServices.fs.On("getCPUInfoStats", mock.Anything).
		Return(make(map[string]*cpuInfo), nil).Once()
	require.Error(t, power.Init())

	mockServices.fs.On("getCPUInfoStats", mock.Anything).
		Return(map[string]*cpuInfo{"0": {
			vendorID:  "GenuineIntel",
			cpuFamily: "test",
		}}, nil).Once()
	require.Error(t, power.Init())

	mockServices.fs.On("getStringsMatchingPatternOnPath", mock.Anything).
		Return(cores, nil).Once().
		On("getCPUInfoStats", mock.Anything).
		Return(map[string]*cpuInfo{"0": {
			vendorID:  "GenuineIntel",
			cpuFamily: "6",
		}}, nil)
	// Verify MSR service initialization.
	power.cpuFrequency = true
	require.NoError(t, power.Init())
	mockServices.fs.AssertCalled(t, "getStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, len(cores), len(power.msr.getCPUCoresData()))

	mockServices.fs.On("getStringsMatchingPatternOnPath", mock.Anything).
		Return(nil, errors.New("error during getStringsMatchingPatternOnPath")).Once()

	// In case of an error when fetching cpu cores plugin should proceed with execution.
	require.NoError(t, power.Init())
	mockServices.fs.AssertCalled(t, "getStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, 0, len(power.msr.getCPUCoresData()))
}

func TestParseCPUMetricsConfig(t *testing.T) {
	power, _ := getPowerWithMockedServices()
	disableCoreMetrics(power)

	power.CPUMetrics = []string{
		"cpu_frequency", "cpu_c0_state_residency", "cpu_c1_state_residency", "cpu_c6_state_residency", "cpu_busy_cycles", "cpu_temperature",
		"cpu_busy_frequency",
	}
	power.parseCPUMetricsConfig()
	verifyCoreMetrics(t, power, true)
	disableCoreMetrics(power)
	verifyCoreMetrics(t, power, false)

	power.CPUMetrics = []string{}
	power.parseCPUMetricsConfig()

	power.CPUMetrics = []string{"cpu_c6_state_residency", "#@$sdkjdfsdf3@", "1pu_c1_state_residency"}
	power.parseCPUMetricsConfig()
	require.Equal(t, false, power.cpuC1StateResidency)
	require.Equal(t, true, power.cpuC6StateResidency)
	disableCoreMetrics(power)
	verifyCoreMetrics(t, power, false)

	power.CPUMetrics = []string{"#@$sdkjdfsdf3@", "1pu_c1_state_residency", "123"}
	power.parseCPUMetricsConfig()
	verifyCoreMetrics(t, power, false)
}

func verifyCoreMetrics(t *testing.T, power *PowerStat, enabled bool) {
	require.Equal(t, enabled, power.cpuFrequency)
	require.Equal(t, enabled, power.cpuC1StateResidency)
	require.Equal(t, enabled, power.cpuC6StateResidency)
	require.Equal(t, enabled, power.cpuC0StateResidency)
	require.Equal(t, enabled, power.cpuBusyCycles)
	require.Equal(t, enabled, power.cpuBusyFrequency)
	require.Equal(t, enabled, power.cpuTemperature)
}

func TestGather(t *testing.T) {
	var acc testutil.Accumulator
	packageIDs := []string{"0", "1"}
	coreIDs := []string{"0", "1", "2", "3"}
	socketCurrentEnergy := 13213852.2
	dramCurrentEnergy := 784552.0
	preparedCPUData := getPreparedCPUData(coreIDs)
	raplDataMap := prepareRaplDataMap(packageIDs, socketCurrentEnergy, dramCurrentEnergy)

	power, mockServices := getPowerWithMockedServices()
	prepareCPUInfo(power, coreIDs, packageIDs)
	enableCoreMetrics(power)
	power.skipFirstIteration = false

	mockServices.rapl.On("initializeRaplData", mock.Anything).
		On("getRaplData").Return(raplDataMap).
		On("retrieveAndCalculateData", mock.Anything).Return(nil).Times(len(raplDataMap)).
		On("getConstraintMaxPowerWatts", mock.Anything).Return(546783852.3, nil)
	mockServices.msr.On("getCPUCoresData").Return(preparedCPUData).
		On("isMsrLoaded", mock.Anything).Return(true).
		On("openAndReadMsr", mock.Anything).Return(nil).
		On("retrieveCPUFrequencyForCore", mock.Anything).Return(1200000.2, nil)

	require.NoError(t, power.Gather(&acc))
	// Number of global metrics   : 3
	// Number of per core metrics : 7
	require.Equal(t, 3*len(packageIDs)+7*len(coreIDs), len(acc.GetTelegrafMetrics()))
}

func TestAddGlobalMetricsNegative(t *testing.T) {
	var acc testutil.Accumulator
	socketCurrentEnergy := 13213852.2
	dramCurrentEnergy := 784552.0
	raplDataMap := prepareRaplDataMap([]string{"0", "1"}, socketCurrentEnergy, dramCurrentEnergy)
	power, mockServices := getPowerWithMockedServices()
	power.skipFirstIteration = false
	mockServices.rapl.On("initializeRaplData", mock.Anything).Once().
		On("getRaplData").Return(raplDataMap).Once().
		On("retrieveAndCalculateData", mock.Anything).Return(errors.New("error while calculating data")).Times(len(raplDataMap))

	power.addGlobalMetrics(&acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
	mockServices.rapl.AssertNumberOfCalls(t, "retrieveAndCalculateData", len(raplDataMap))

	mockServices.rapl.On("initializeRaplData", mock.Anything).Once().
		On("getRaplData").Return(make(map[string]*raplData)).Once()

	power.addGlobalMetrics(&acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
	mockServices.rapl.AssertNotCalled(t, "retrieveAndCalculateData")

	mockServices.rapl.On("initializeRaplData", mock.Anything).Once().
		On("getRaplData").Return(raplDataMap).
		On("retrieveAndCalculateData", mock.Anything).Return(nil).Once().
		On("retrieveAndCalculateData", mock.Anything).Return(errors.New("error while calculating data")).Once().
		On("getConstraintMaxPowerWatts", mock.Anything).Return(12313851.5, nil).Twice()

	power.addGlobalMetrics(&acc)
	require.Equal(t, 3, len(acc.GetTelegrafMetrics()))
}

func TestAddGlobalMetricsPositive(t *testing.T) {
	var acc testutil.Accumulator
	socketCurrentEnergy := 3644574.4
	dramCurrentEnergy := 124234872.5
	raplDataMap := prepareRaplDataMap([]string{"0", "1"}, socketCurrentEnergy, dramCurrentEnergy)
	maxPower := 546783852.9
	power, mockServices := getPowerWithMockedServices()
	power.skipFirstIteration = false

	mockServices.rapl.On("initializeRaplData", mock.Anything).
		On("getRaplData").Return(raplDataMap).
		On("retrieveAndCalculateData", mock.Anything).Return(nil).Times(len(raplDataMap)).
		On("getConstraintMaxPowerWatts", mock.Anything).Return(maxPower, nil).Twice().
		On("getCurrentDramPowerConsumption", mock.Anything).Return(dramCurrentEnergy)

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
	power, mockServices := getPowerWithMockedServices()

	mockServices.msr.On("openAndReadMsr", core).Return(errors.New("error reading MSR file")).Once()

	// Skip generating metric for CPU frequency.
	power.cpuFrequency = false

	wg.Add(1)
	power.addMetricsForSingleCore(core, &acc, &wg)
	wg.Wait()

	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddCPUFrequencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "1"
	coreID := "3"
	packageID := "0"
	frequency := 1200000.2
	power, mockServices := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)

	mockServices.msr.On("retrieveCPUFrequencyForCore", mock.Anything).
		Return(float64(0), errors.New("error on reading file")).Once()

	power.addCPUFrequencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))

	mockServices.msr.On("retrieveCPUFrequencyForCore", mock.Anything).Return(frequency, nil).Once()

	power.addCPUFrequencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedFrequency := roundFloatToNearestTwoDecimalPlaces(frequency)
	expectedMetric := getPowerCoreMetric("cpu_frequency_mhz", expectedFrequency, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)
}

func TestReadUncoreFreq(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "0"
	packageID := "0"
	die := "0"
	power, mockServices := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})

	mockServices.msr.On("getCPUCoresData").Return(preparedData)

	mockServices.msr.On("isMsrLoaded").Return(true)

	mockServices.msr.On("readSingleMsr", "0", msrUncorePerfStatusString).Return(uint64(10), nil)

	mockServices.msr.On("retrieveUncoreFrequency", "0", "initial", "min", "0").
		Return(float64(500), nil)
	mockServices.msr.On("retrieveUncoreFrequency", "0", "initial", "max", "0").
		Return(float64(1200), nil)
	mockServices.msr.On("retrieveUncoreFrequency", "0", "current", "min", "0").
		Return(float64(600), nil)
	mockServices.msr.On("retrieveUncoreFrequency", "0", "current", "max", "0").
		Return(float64(1100), nil)

	power.readUncoreFreq("current", packageID, die, &acc)
	power.readUncoreFreq("initial", packageID, die, &acc)

	require.Equal(t, 2, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerUncoreFreqMetric("initial", float64(500), float64(1200), nil, packageID, die)
	acc.AssertContainsTaggedFields(t, "powerstat_package", expectedMetric.fields, expectedMetric.tags)

	expectedMetric = getPowerUncoreFreqMetric("current", float64(600), float64(1100), uint64(1000), packageID, die)
	acc.AssertContainsTaggedFields(t, "powerstat_package", expectedMetric.fields, expectedMetric.tags)
}

func TestAddCoreCPUTemperatureMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, mockServices := getPowerWithMockedServices()
	preparedData := getPreparedCPUData([]string{cpuID})
	expectedTemp := preparedData[cpuID].throttleTemp - preparedData[cpuID].temp
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)

	mockServices.msr.On("getCPUCoresData").Return(preparedData).Once()
	power.addCPUTemperatureMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_temperature_celsius", expectedTemp, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)
}

func TestAddC6StateResidencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, mockServices := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})
	expectedC6 := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier *
		float64(preparedData[cpuID].c6Delta) / float64(preparedData[cpuID].timeStampCounterDelta))

	mockServices.msr.On("getCPUCoresData").Return(preparedData).Twice()
	power.addCPUC6StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_c6_state_residency_percent", expectedC6, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	acc.ClearMetrics()
	preparedData[cpuID].timeStampCounterDelta = 0

	power.addCPUC6StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddC0StateResidencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, mockServices := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})
	expectedBusyCycles := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier * float64(preparedData[cpuID].mperfDelta) /
		float64(preparedData[cpuID].timeStampCounterDelta))

	mockServices.msr.On("getCPUCoresData").Return(preparedData).Twice()
	power.cpuBusyCycles, power.cpuC0StateResidency = true, true
	power.addCPUC0StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 2, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_c0_state_residency_percent", expectedBusyCycles, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	// Deprecated
	expectedMetric = getPowerCoreMetric("cpu_busy_cycles_percent", expectedBusyCycles, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	acc.ClearMetrics()
	preparedData[cpuID].timeStampCounterDelta = 0
	power.addCPUC0StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddProcessorBusyFrequencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, mockServices := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})
	power.skipFirstIteration = false

	mockServices.msr.On("getCPUCoresData").Return(preparedData).Twice()
	power.addCPUBusyFrequencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	acc.ClearMetrics()
	preparedData[cpuID].mperfDelta = 0
	power.addCPUBusyFrequencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddC1StateResidencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, mockServices := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})
	c1 := preparedData[cpuID].timeStampCounterDelta - preparedData[cpuID].mperfDelta - preparedData[cpuID].c3Delta -
		preparedData[cpuID].c6Delta - preparedData[cpuID].c7Delta
	expectedC1 := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier * float64(c1) / float64(preparedData[cpuID].timeStampCounterDelta))

	mockServices.msr.On("getCPUCoresData").Return(preparedData).Twice()

	power.addCPUC1StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_c1_state_residency_percent", expectedC1, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	acc.ClearMetrics()
	preparedData[cpuID].timeStampCounterDelta = 0
	power.addCPUC1StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddThermalDesignPowerMetric(t *testing.T) {
	var acc testutil.Accumulator
	sockets := []string{"0"}
	maxPower := 195720672.1
	power, mockServices := getPowerWithMockedServices()

	mockServices.rapl.On("getConstraintMaxPowerWatts", mock.Anything).
		Return(float64(0), errors.New("getConstraintMaxPowerWatts error")).Once().
		On("getConstraintMaxPowerWatts", mock.Anything).Return(maxPower, nil).Once()

	power.addThermalDesignPowerMetric(sockets[0], &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))

	power.addThermalDesignPowerMetric(sockets[0], &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedTDP := roundFloatToNearestTwoDecimalPlaces(maxPower)
	expectedMetric := getPowerGlobalMetric("thermal_design_power_watts", expectedTDP, sockets[0])
	acc.AssertContainsTaggedFields(t, "powerstat_package", expectedMetric.fields, expectedMetric.tags)
}

func TestCalculateTurboRatioGroup(t *testing.T) {
	coreCounts := uint64(0x0807060504030201)
	msr := uint64(0x0807060504030201)
	turboRatioLimitGroups := make(map[int]uint64)

	calculateTurboRatioGroup(coreCounts, msr, turboRatioLimitGroups)
	require.Equal(t, 8, len(turboRatioLimitGroups))
	require.Equal(t, uint64(100), turboRatioLimitGroups[1])
	require.Equal(t, uint64(200), turboRatioLimitGroups[2])
	require.Equal(t, uint64(300), turboRatioLimitGroups[3])
	require.Equal(t, uint64(400), turboRatioLimitGroups[4])
	require.Equal(t, uint64(500), turboRatioLimitGroups[5])
	require.Equal(t, uint64(600), turboRatioLimitGroups[6])
	require.Equal(t, uint64(700), turboRatioLimitGroups[7])
	require.Equal(t, uint64(800), turboRatioLimitGroups[8])

	coreCounts = uint64(0x100e0c0a08060402)
	calculateTurboRatioGroup(coreCounts, msr, turboRatioLimitGroups)
	require.Equal(t, 16, len(turboRatioLimitGroups))
	require.Equal(t, uint64(100), turboRatioLimitGroups[1])
	require.Equal(t, uint64(100), turboRatioLimitGroups[2])
	require.Equal(t, uint64(200), turboRatioLimitGroups[3])
	require.Equal(t, uint64(200), turboRatioLimitGroups[4])
	require.Equal(t, uint64(300), turboRatioLimitGroups[5])
	require.Equal(t, uint64(300), turboRatioLimitGroups[6])
	require.Equal(t, uint64(400), turboRatioLimitGroups[7])
	require.Equal(t, uint64(400), turboRatioLimitGroups[8])
	require.Equal(t, uint64(500), turboRatioLimitGroups[9])
	require.Equal(t, uint64(500), turboRatioLimitGroups[10])
	require.Equal(t, uint64(600), turboRatioLimitGroups[11])
	require.Equal(t, uint64(600), turboRatioLimitGroups[12])
	require.Equal(t, uint64(700), turboRatioLimitGroups[13])
	require.Equal(t, uint64(700), turboRatioLimitGroups[14])
	require.Equal(t, uint64(800), turboRatioLimitGroups[15])
	require.Equal(t, uint64(800), turboRatioLimitGroups[16])
	coreCounts = uint64(0x1211)
	msr = uint64(0xfffe)
	calculateTurboRatioGroup(coreCounts, msr, turboRatioLimitGroups)
	require.Equal(t, 18, len(turboRatioLimitGroups))
	require.Equal(t, uint64(25400), turboRatioLimitGroups[17])
	require.Equal(t, uint64(25500), turboRatioLimitGroups[18])

	coreCounts = uint64(0x1201)
	msr = uint64(0x0202)
	calculateTurboRatioGroup(coreCounts, msr, turboRatioLimitGroups)
	require.Equal(t, 18, len(turboRatioLimitGroups))
	require.Equal(t, uint64(200), turboRatioLimitGroups[1])
	require.Equal(t, uint64(200), turboRatioLimitGroups[2])
	require.Equal(t, uint64(200), turboRatioLimitGroups[3])
	require.Equal(t, uint64(200), turboRatioLimitGroups[4])
	require.Equal(t, uint64(200), turboRatioLimitGroups[5])
	require.Equal(t, uint64(200), turboRatioLimitGroups[6])
	require.Equal(t, uint64(200), turboRatioLimitGroups[7])
	require.Equal(t, uint64(200), turboRatioLimitGroups[8])
	require.Equal(t, uint64(200), turboRatioLimitGroups[9])
	require.Equal(t, uint64(200), turboRatioLimitGroups[10])
	require.Equal(t, uint64(200), turboRatioLimitGroups[11])
	require.Equal(t, uint64(200), turboRatioLimitGroups[12])
	require.Equal(t, uint64(200), turboRatioLimitGroups[13])
	require.Equal(t, uint64(200), turboRatioLimitGroups[14])
	require.Equal(t, uint64(200), turboRatioLimitGroups[15])
	require.Equal(t, uint64(200), turboRatioLimitGroups[16])
	require.Equal(t, uint64(200), turboRatioLimitGroups[17])
	require.Equal(t, uint64(200), turboRatioLimitGroups[18])

	coreCounts = uint64(0x1211)
	msr = uint64(0xfffe)
	turboRatioLimitGroups = make(map[int]uint64)
	calculateTurboRatioGroup(coreCounts, msr, turboRatioLimitGroups)
	require.Equal(t, 2, len(turboRatioLimitGroups))
	require.Equal(t, uint64(25400), turboRatioLimitGroups[17])
	require.Equal(t, uint64(25500), turboRatioLimitGroups[18])
}

func getPreparedCPUData(cores []string) map[string]*msrData {
	msrDataMap := make(map[string]*msrData)

	for _, core := range cores {
		msrDataMap[core] = &msrData{
			mperf:                 43079,
			aperf:                 82001,
			timeStampCounter:      15514,
			c3:                    52829,
			c6:                    86930,
			c7:                    25340,
			throttleTemp:          88150,
			temp:                  40827,
			mperfDelta:            23515,
			aperfDelta:            33866,
			timeStampCounterDelta: 13686000,
			c3Delta:               20003,
			c6Delta:               44518,
			c7Delta:               20979,
		}
	}

	return msrDataMap
}

func getGlobalMetrics(maxPower float64, socketCurrentEnergy float64, dramCurrentEnergy float64) []struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	return []struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		getPowerGlobalMetric("thermal_design_power_watts", roundFloatToNearestTwoDecimalPlaces(maxPower), "0"),
		getPowerGlobalMetric("thermal_design_power_watts", roundFloatToNearestTwoDecimalPlaces(maxPower), "1"),
		getPowerGlobalMetric("current_power_consumption_watts", roundFloatToNearestTwoDecimalPlaces(socketCurrentEnergy), "0"),
		getPowerGlobalMetric("current_power_consumption_watts", roundFloatToNearestTwoDecimalPlaces(socketCurrentEnergy), "1"),
		getPowerGlobalMetric("current_dram_power_consumption_watts", roundFloatToNearestTwoDecimalPlaces(dramCurrentEnergy), "0"),
		getPowerGlobalMetric("current_dram_power_consumption_watts", roundFloatToNearestTwoDecimalPlaces(dramCurrentEnergy), "1"),
	}
}

func getPowerCoreMetric(name string, value interface{}, coreID string, packageID string, cpuID string) struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	return getPowerMetric(name, value, map[string]string{"package_id": packageID, "core_id": coreID, "cpu_id": cpuID})
}

func getPowerGlobalMetric(name string, value interface{}, socketID string) struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	return getPowerMetric(name, value, map[string]string{"package_id": socketID})
}

func getPowerUncoreFreqMetric(typeFreq string, limitMin interface{}, limitMax interface{}, current interface{}, socketID string, die string) struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	var ret struct {
		fields map[string]interface{}
		tags   map[string]string
	}
	ret.tags = make(map[string]string)
	ret.fields = make(map[string]interface{})
	ret.tags["package_id"] = socketID
	ret.tags["die"] = die
	ret.tags["type"] = typeFreq
	ret.fields["uncore_frequency_limit_mhz_min"] = limitMin
	ret.fields["uncore_frequency_limit_mhz_max"] = limitMax
	if typeFreq == "current" {
		ret.fields["uncore_frequency_mhz_cur"] = current
	}
	return ret
}

func getPowerMetric(name string, value interface{}, tags map[string]string) struct {
	fields map[string]interface{}
	tags   map[string]string
} {
	return struct {
		fields map[string]interface{}
		tags   map[string]string
	}{
		map[string]interface{}{
			name: value,
		},
		tags,
	}
}

func prepareCPUInfoForSingleCPU(power *PowerStat, cpuID string, coreID string, packageID string) {
	power.cpuInfo = make(map[string]*cpuInfo)
	power.cpuInfo[cpuID] = &cpuInfo{
		physicalID: packageID,
		coreID:     coreID,
		cpuID:      cpuID,
	}
}

func prepareCPUInfo(power *PowerStat, coreIDs []string, packageIDs []string) {
	power.cpuInfo = make(map[string]*cpuInfo)
	currentCPU := 0
	for _, packageID := range packageIDs {
		for _, coreID := range coreIDs {
			cpuID := strconv.Itoa(currentCPU)
			power.cpuInfo[cpuID] = &cpuInfo{
				physicalID: packageID,
				cpuID:      cpuID,
				coreID:     coreID,
			}
			currentCPU++
		}
	}
}

func enableCoreMetrics(power *PowerStat) {
	power.cpuC0StateResidency = true
	power.cpuC1StateResidency = true
	power.cpuC6StateResidency = true
	power.cpuTemperature = true
	power.cpuBusyFrequency = true
	power.cpuFrequency = true
	power.cpuBusyCycles = true
}

func disableCoreMetrics(power *PowerStat) {
	power.cpuC0StateResidency = false
	power.cpuC1StateResidency = false
	power.cpuC6StateResidency = false
	power.cpuBusyCycles = false
	power.cpuTemperature = false
	power.cpuBusyFrequency = false
	power.cpuFrequency = false
}

func prepareRaplDataMap(socketIDs []string, socketCurrentEnergy float64, dramCurrentEnergy float64) map[string]*raplData {
	raplDataMap := make(map[string]*raplData, len(socketIDs))
	for _, socketID := range socketIDs {
		raplDataMap[socketID] = &raplData{
			socketCurrentEnergy: socketCurrentEnergy,
			dramCurrentEnergy:   dramCurrentEnergy,
		}
	}

	return raplDataMap
}

func getPowerWithMockedServices() (*PowerStat, *MockServices) {
	var mockServices MockServices
	mockServices.fs = &mockFileService{}
	mockServices.msr = &mockMsrService{}
	mockServices.rapl = &mockRaplService{}
	p := newPowerStat(mockServices.fs)
	p.Log = testutil.Logger{Name: "PowerPluginTest"}
	p.rapl = mockServices.rapl
	p.msr = mockServices.msr
	p.packageCurrentPowerConsumption = true
	p.packageCurrentDramPowerConsumption = true
	p.packageThermalDesignPower = true

	return p, &mockServices
}

func TestGetBusClock(t *testing.T) {
	tests := []struct {
		name                string
		modelCPU            uint64
		socketID            string
		msrFSBFreqValue     uint64
		readSingleMsrErrFSB error
		cpuBusClockValue    float64
	}{
		{
			name:             "Error_withUnknownCPUmodel",
			socketID:         "0",
			modelCPU:         0xFF,
			cpuBusClockValue: 0,
		},
		{
			name:             "OK_withFBS100",
			socketID:         "0",
			modelCPU:         106,
			msrFSBFreqValue:  1,
			cpuBusClockValue: 100.0,
		},
		{
			name:             "OK_withFBS133",
			socketID:         "0",
			modelCPU:         0x1F,
			cpuBusClockValue: 133,
		},
		{
			name:                "Error_withFBSCalculated",
			socketID:            "0",
			modelCPU:            0x37,
			msrFSBFreqValue:     0,
			readSingleMsrErrFSB: errors.New("something is wrong"),
		},
		{
			name:             "OK_withFBSCalculated83.3",
			socketID:         "0",
			modelCPU:         0x37,
			msrFSBFreqValue:  0,
			cpuBusClockValue: 83.3,
		},
		{
			name:             "OK_withFBSCalculated100",
			socketID:         "0",
			modelCPU:         0x37,
			msrFSBFreqValue:  1,
			cpuBusClockValue: 100,
		},
		{
			name:             "OK_withFBSCalculated133.3",
			socketID:         "0",
			modelCPU:         0x37,
			msrFSBFreqValue:  2,
			cpuBusClockValue: 133.3,
		},
		{
			name:             "OK_withFBSCalculated116.7",
			socketID:         "0",
			modelCPU:         0x37,
			msrFSBFreqValue:  3,
			cpuBusClockValue: 116.7,
		},
		{
			name:             "OK_withFBSCalculated80",
			socketID:         "0",
			modelCPU:         0x37,
			msrFSBFreqValue:  4,
			cpuBusClockValue: 80,
		},
		{
			name:             "OK_withFBSCalculatedUnknownFSBFreq",
			socketID:         "0",
			modelCPU:         0x37,
			msrFSBFreqValue:  5,
			cpuBusClockValue: 116.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, mockServices := getPowerWithMockedServices()
			busClockCalculate := []uint64{0x37, 0x4D}
			p.cpuInfo = map[string]*cpuInfo{
				tt.socketID: {cpuID: tt.socketID, physicalID: tt.socketID, model: strconv.FormatUint(tt.modelCPU, 10)},
			}
			if contains(busClockCalculate, tt.modelCPU) {
				mockServices.msr.On("readSingleMsr", mock.Anything, msrFSBFreqString).Return(tt.msrFSBFreqValue, tt.readSingleMsrErrFSB)
			}
			defer mockServices.msr.AssertExpectations(t)

			value := p.getBusClock(tt.socketID)
			require.Equal(t, tt.cpuBusClockValue, value)
		})
	}
}

func TestFillCPUBusClock(t *testing.T) {
	tests := []struct {
		name                       string
		modelCPU                   uint64
		busClockValue              float64
		packageCPUBaseFrequencySet bool
	}{
		{
			name:          "NotSet_0",
			modelCPU:      0xFF,
			busClockValue: 0,
		},
		{
			name:                       "Set_100",
			modelCPU:                   0x2A,
			busClockValue:              100,
			packageCPUBaseFrequencySet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := getPowerWithMockedServices()
			p.packageCPUBaseFrequency = true
			p.cpuInfo = map[string]*cpuInfo{
				"0": {cpuID: "0", physicalID: "0", model: strconv.FormatUint(tt.modelCPU, 10)},
			}

			p.fillCPUBusClock()
			require.Equal(t, tt.busClockValue, p.cpuBusClockValue)
			require.Equal(t, tt.packageCPUBaseFrequencySet, p.packageCPUBaseFrequency)
		})
	}
}

func TestAddCPUBaseFreq(t *testing.T) {
	tests := []struct {
		name                  string
		socketID              string
		readSingleMsrErrRatio error
		msrPlatformInfoValue  uint64
		setupPowerstat        func(t *testing.T)
		clockBusValue         float64
		nonTurboRatio         float64
		metricExpected        bool
	}{
		{
			name:                  "Error_reading_msr",
			socketID:              "0",
			clockBusValue:         100,
			readSingleMsrErrRatio: errors.New("can't read msr"),
			metricExpected:        false,
		},
		{
			name:                 "NoMetric_Ratio_is_0",
			socketID:             "0",
			msrPlatformInfoValue: 0x8008082FF2810000,
			clockBusValue:        100,
			nonTurboRatio:        0,
			metricExpected:       false,
		},
		{
			name:                 "OK_Ratio_is_24",
			socketID:             "0",
			msrPlatformInfoValue: 0x8008082FF2811800,
			clockBusValue:        100,
			nonTurboRatio:        24,
			metricExpected:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator
			p, mockServices := getPowerWithMockedServices()

			p.cpuInfo = map[string]*cpuInfo{
				tt.socketID: {cpuID: tt.socketID, physicalID: tt.socketID},
			}
			p.cpuBusClockValue = tt.clockBusValue

			mockServices.msr.On("readSingleMsr", mock.Anything, msrPlatformInfoString).Return(tt.msrPlatformInfoValue, tt.readSingleMsrErrRatio)
			defer mockServices.msr.AssertExpectations(t)

			p.addCPUBaseFreq(tt.socketID, &acc)
			actual := acc.GetTelegrafMetrics()
			if !tt.metricExpected {
				require.Len(t, actual, 0)
				return
			}

			require.Len(t, actual, 1)
			expected := []telegraf.Metric{
				testutil.MustMetric(
					"powerstat_package",
					map[string]string{
						"package_id": tt.socketID,
					},
					map[string]interface{}{
						"cpu_base_frequency_mhz": uint64(tt.nonTurboRatio * tt.clockBusValue),
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			}
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}
