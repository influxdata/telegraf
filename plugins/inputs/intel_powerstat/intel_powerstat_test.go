//go:build linux
// +build linux

package intel_powerstat

import (
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestInitPlugin(t *testing.T) {
	cores := []string{"cpu0", "cpu1", "cpu2", "cpu3"}
	power, fsMock, _, _ := getPowerWithMockedServices()

	fsMock.On("getCPUInfoStats", mock.Anything).
		Return(nil, errors.New("error getting cpu stats")).Once()
	require.Error(t, power.Init())

	fsMock.On("getCPUInfoStats", mock.Anything).
		Return(make(map[string]*cpuInfo), nil).Once()
	require.Error(t, power.Init())

	fsMock.On("getCPUInfoStats", mock.Anything).
		Return(map[string]*cpuInfo{"0": {
			vendorID:  "GenuineIntel",
			cpuFamily: "test",
		}}, nil).Once()
	require.Error(t, power.Init())

	fsMock.On("getStringsMatchingPatternOnPath", mock.Anything).
		Return(cores, nil).Once().
		On("getCPUInfoStats", mock.Anything).
		Return(map[string]*cpuInfo{"0": {
			vendorID:  "GenuineIntel",
			cpuFamily: "6",
		}}, nil)
	// Verify MSR service initialization.
	power.cpuFrequency = true
	require.NoError(t, power.Init())
	fsMock.AssertCalled(t, "getStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, len(cores), len(power.msr.getCPUCoresData()))

	fsMock.On("getStringsMatchingPatternOnPath", mock.Anything).
		Return(nil, errors.New("error during getStringsMatchingPatternOnPath")).Once()

	// In case of an error when fetching cpu cores plugin should proceed with execution.
	require.NoError(t, power.Init())
	fsMock.AssertCalled(t, "getStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, 0, len(power.msr.getCPUCoresData()))
}

func TestParseCPUMetricsConfig(t *testing.T) {
	power, _, _, _ := getPowerWithMockedServices()
	disableCoreMetrics(power)

	power.CPUMetrics = []string{
		"cpu_frequency", "cpu_c1_state_residency", "cpu_c6_state_residency", "cpu_busy_cycles", "cpu_temperature",
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

	power, _, raplMock, msrMock := getPowerWithMockedServices()
	prepareCPUInfo(power, coreIDs, packageIDs)
	enableCoreMetrics(power)
	power.skipFirstIteration = false

	raplMock.On("initializeRaplData", mock.Anything).
		On("getRaplData").Return(raplDataMap).
		On("retrieveAndCalculateData", mock.Anything).Return(nil).Times(len(raplDataMap)).
		On("getConstraintMaxPowerWatts", mock.Anything).Return(546783852.3, nil)
	msrMock.On("getCPUCoresData").Return(preparedCPUData).
		On("openAndReadMsr", mock.Anything).Return(nil).
		On("retrieveCPUFrequencyForCore", mock.Anything).Return(1200000.2, nil)

	require.NoError(t, power.Gather(&acc))
	// Number of global metrics   : 3
	// Number of per core metrics : 6
	require.Equal(t, 3*len(packageIDs)+6*len(coreIDs), len(acc.GetTelegrafMetrics()))
}

func TestAddGlobalMetricsNegative(t *testing.T) {
	var acc testutil.Accumulator
	socketCurrentEnergy := 13213852.2
	dramCurrentEnergy := 784552.0
	raplDataMap := prepareRaplDataMap([]string{"0", "1"}, socketCurrentEnergy, dramCurrentEnergy)
	power, _, raplMock, _ := getPowerWithMockedServices()
	power.skipFirstIteration = false
	raplMock.On("initializeRaplData", mock.Anything).Once().
		On("getRaplData").Return(raplDataMap).Once().
		On("retrieveAndCalculateData", mock.Anything).Return(errors.New("error while calculating data")).Times(len(raplDataMap))

	power.addGlobalMetrics(&acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
	raplMock.AssertNumberOfCalls(t, "retrieveAndCalculateData", len(raplDataMap))

	raplMock.On("initializeRaplData", mock.Anything).Once().
		On("getRaplData").Return(make(map[string]*raplData)).Once()

	power.addGlobalMetrics(&acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
	raplMock.AssertNotCalled(t, "retrieveAndCalculateData")

	raplMock.On("initializeRaplData", mock.Anything).Once().
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
	power, _, raplMock, _ := getPowerWithMockedServices()
	power.skipFirstIteration = false

	raplMock.On("initializeRaplData", mock.Anything).
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
	power, _, _, msrMock := getPowerWithMockedServices()

	msrMock.On("openAndReadMsr", core).Return(errors.New("error reading MSR file")).Once()

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
	power, _, _, msrMock := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)

	msrMock.On("retrieveCPUFrequencyForCore", mock.Anything).
		Return(float64(0), errors.New("error on reading file")).Once()

	power.addCPUFrequencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))

	msrMock.On("retrieveCPUFrequencyForCore", mock.Anything).Return(frequency, nil).Once()

	power.addCPUFrequencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedFrequency := roundFloatToNearestTwoDecimalPlaces(frequency)
	expectedMetric := getPowerCoreMetric("cpu_frequency_mhz", expectedFrequency, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)
}

func TestAddCoreCPUTemperatureMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, _, _, msrMock := getPowerWithMockedServices()
	preparedData := getPreparedCPUData([]string{cpuID})
	expectedTemp := preparedData[cpuID].throttleTemp - preparedData[cpuID].temp
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)

	msrMock.On("getCPUCoresData").Return(preparedData).Once()
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
	power, _, _, msrMock := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})
	expectedC6 := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier *
		float64(preparedData[cpuID].c6Delta) / float64(preparedData[cpuID].timeStampCounterDelta))

	msrMock.On("getCPUCoresData").Return(preparedData).Twice()
	power.addCPUC6StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_c6_state_residency_percent", expectedC6, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	acc.ClearMetrics()
	preparedData[cpuID].timeStampCounterDelta = 0

	power.addCPUC6StateResidencyMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddProcessorBusyCyclesMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, _, _, msrMock := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})
	expectedBusyCycles := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier * float64(preparedData[cpuID].mperfDelta) /
		float64(preparedData[cpuID].timeStampCounterDelta))

	msrMock.On("getCPUCoresData").Return(preparedData).Twice()
	power.addCPUBusyCyclesMetric(cpuID, &acc)
	require.Equal(t, 1, len(acc.GetTelegrafMetrics()))

	expectedMetric := getPowerCoreMetric("cpu_busy_cycles_percent", expectedBusyCycles, coreID, packageID, cpuID)
	acc.AssertContainsTaggedFields(t, "powerstat_core", expectedMetric.fields, expectedMetric.tags)

	acc.ClearMetrics()
	preparedData[cpuID].timeStampCounterDelta = 0
	power.addCPUBusyCyclesMetric(cpuID, &acc)
	require.Equal(t, 0, len(acc.GetTelegrafMetrics()))
}

func TestAddProcessorBusyFrequencyMetric(t *testing.T) {
	var acc testutil.Accumulator
	cpuID := "0"
	coreID := "2"
	packageID := "1"
	power, _, _, msrMock := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})
	power.skipFirstIteration = false

	msrMock.On("getCPUCoresData").Return(preparedData).Twice()
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
	power, _, _, msrMock := getPowerWithMockedServices()
	prepareCPUInfoForSingleCPU(power, cpuID, coreID, packageID)
	preparedData := getPreparedCPUData([]string{cpuID})
	c1 := preparedData[cpuID].timeStampCounterDelta - preparedData[cpuID].mperfDelta - preparedData[cpuID].c3Delta -
		preparedData[cpuID].c6Delta - preparedData[cpuID].c7Delta
	expectedC1 := roundFloatToNearestTwoDecimalPlaces(percentageMultiplier * float64(c1) / float64(preparedData[cpuID].timeStampCounterDelta))

	msrMock.On("getCPUCoresData").Return(preparedData).Twice()

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
	power, _, raplMock, _ := getPowerWithMockedServices()

	raplMock.On("getConstraintMaxPowerWatts", mock.Anything).
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
	power.cpuC1StateResidency = true
	power.cpuC6StateResidency = true
	power.cpuTemperature = true
	power.cpuBusyFrequency = true
	power.cpuFrequency = true
	power.cpuBusyCycles = true
}

func disableCoreMetrics(power *PowerStat) {
	power.cpuC1StateResidency = false
	power.cpuC6StateResidency = false
	power.cpuTemperature = false
	power.cpuBusyFrequency = false
	power.cpuFrequency = false
	power.cpuBusyCycles = false
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

func getPowerWithMockedServices() (*PowerStat, *mockFileService, *mockRaplService, *mockMsrService) {
	fsMock := &mockFileService{}
	msrMock := &mockMsrService{}
	raplMock := &mockRaplService{}
	logger := testutil.Logger{Name: "PowerPluginTest"}
	p := newPowerStat(fsMock)
	p.Log = logger
	p.fs = fsMock
	p.rapl = raplMock
	p.msr = msrMock

	return p, fsMock, raplMock, msrMock
}
