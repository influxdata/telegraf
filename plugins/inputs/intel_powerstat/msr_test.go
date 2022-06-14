//go:build linux
// +build linux

package intel_powerstat

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestReadDataFromMsrPositive(t *testing.T) {
	firstValue := uint64(1000000)
	secondValue := uint64(5000000)
	delta := secondValue - firstValue
	cpuCores := []string{"cpu0", "cpu1"}
	msr, fsMock := getMsrServiceWithMockedFs()
	prepareTestData(fsMock, cpuCores, msr, t)
	cores := trimCPUFromCores(cpuCores)

	methodCallNumberForFirstValue := len(msr.msrOffsets) * len(cores)
	methodCallNumberForSecondValue := methodCallNumberForFirstValue * 2

	fsMock.On("readFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(firstValue, nil).Times(methodCallNumberForFirstValue)
	for _, core := range cores {
		require.NoError(t, msr.readDataFromMsr(core, nil))
	}
	fsMock.AssertNumberOfCalls(t, "readFileAtOffsetToUint64", methodCallNumberForFirstValue)
	verifyCPUCoresData(cores, t, msr, firstValue, false, 0)

	fsMock.On("readFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(secondValue, nil).Times(methodCallNumberForFirstValue)
	for _, core := range cores {
		require.NoError(t, msr.readDataFromMsr(core, nil))
	}
	fsMock.AssertNumberOfCalls(t, "readFileAtOffsetToUint64", methodCallNumberForSecondValue)
	verifyCPUCoresData(cores, t, msr, secondValue, true, delta)
}

func trimCPUFromCores(cpuCores []string) []string {
	cores := make([]string, 0)
	for _, core := range cpuCores {
		cores = append(cores, strings.TrimPrefix(core, "cpu"))
	}
	return cores
}

func TestReadDataFromMsrNegative(t *testing.T) {
	firstValue := uint64(1000000)
	cpuCores := []string{"cpu0", "cpu1"}
	msr, fsMock := getMsrServiceWithMockedFs()

	prepareTestData(fsMock, cpuCores, msr, t)
	cores := trimCPUFromCores(cpuCores)

	methodCallNumberPerCore := len(msr.msrOffsets)

	// Normal execution for first core.
	fsMock.On("readFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(firstValue, nil).Times(methodCallNumberPerCore).
		// Fail to read file for second core.
		On("readFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(uint64(0), errors.New("error reading file")).Times(methodCallNumberPerCore)

	require.NoError(t, msr.readDataFromMsr(cores[0], nil))
	require.Error(t, msr.readDataFromMsr(cores[1], nil))
}

func TestReadValueFromFileAtOffset(t *testing.T) {
	cores := []string{"cpu0", "cpu1"}
	msr, fsMock := getMsrServiceWithMockedFs()
	ctx := context.Background()
	testChannel := make(chan uint64, 1)
	defer close(testChannel)
	zero := uint64(0)

	prepareTestData(fsMock, cores, msr, t)

	fsMock.On("readFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(zero, errors.New("error reading file")).Once()
	require.Error(t, msr.readValueFromFileAtOffset(ctx, testChannel, nil, 0))

	fsMock.On("readFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(zero, nil).Once()
	require.Equal(t, nil, msr.readValueFromFileAtOffset(ctx, testChannel, nil, 0))
	require.Equal(t, zero, <-testChannel)
}

func TestCreateUncoreFreqPath(t *testing.T) {
	path, err := createUncoreFreqPath("0", "initial", "min", "0")
	expectedPath := "/sys/devices/system/cpu/intel_uncore_frequency/package_00_die_00/initial_min_freq_khz"
	require.Equal(t, nil, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("0", "initial", "max", "0")
	expectedPath = "/sys/devices/system/cpu/intel_uncore_frequency/package_00_die_00/initial_max_freq_khz"
	require.Equal(t, nil, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("0", "current", "min", "0")
	expectedPath = "/sys/devices/system/cpu/intel_uncore_frequency/package_00_die_00/min_freq_khz"
	require.Equal(t, nil, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("0", "current", "max", "0")
	expectedPath = "/sys/devices/system/cpu/intel_uncore_frequency/package_00_die_00/max_freq_khz"
	require.Equal(t, nil, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("9", "current", "max", "0")
	expectedPath = "/sys/devices/system/cpu/intel_uncore_frequency/package_09_die_00/max_freq_khz"
	require.Equal(t, nil, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("99", "current", "max", "0")
	expectedPath = "/sys/devices/system/cpu/intel_uncore_frequency/package_99_die_00/max_freq_khz"
	require.Equal(t, nil, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("0", "current", "max", "9")
	expectedPath = "/sys/devices/system/cpu/intel_uncore_frequency/package_00_die_09/max_freq_khz"
	require.Equal(t, nil, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("0", "current", "max", "99")
	expectedPath = "/sys/devices/system/cpu/intel_uncore_frequency/package_00_die_99/max_freq_khz"
	require.Equal(t, nil, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("0", "foo", "max", "0")
	expectedPath = ""
	expectedError := errors.New("unknown frequency type foo, only 'initial' and 'current' are supported")
	require.Equal(t, expectedError, err)
	require.Equal(t, expectedPath, path)

	path, err = createUncoreFreqPath("0", "current", "bar", "0")
	expectedPath = ""
	expectedError = errors.New("unknown frequency type bar, only 'min' and 'max' are supported")
	require.Equal(t, expectedError, err)
	require.Equal(t, expectedPath, path)
}

func prepareTestData(fsMock *mockFileService, cores []string, msr *msrServiceImpl, t *testing.T) {
	// Prepare MSR offsets and CPUCoresData for test.
	fsMock.On("getStringsMatchingPatternOnPath", mock.Anything).
		Return(cores, nil).Once()
	require.NoError(t, msr.setCPUCores())
	fsMock.AssertCalled(t, "getStringsMatchingPatternOnPath", mock.Anything)
}

func verifyCPUCoresData(cores []string, t *testing.T, msr *msrServiceImpl, expectedValue uint64, verifyDelta bool, delta uint64) {
	for _, core := range cores {
		require.Equal(t, expectedValue, msr.cpuCoresData[core].c3)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].c6)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].c7)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].mperf)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].aperf)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].timeStampCounter)
		require.Equal(t, int64((expectedValue>>16)&0xFF), msr.cpuCoresData[core].throttleTemp)
		require.Equal(t, int64((expectedValue>>16)&0x7F), msr.cpuCoresData[core].temp)

		if verifyDelta {
			require.Equal(t, delta, msr.cpuCoresData[core].c3Delta)
			require.Equal(t, delta, msr.cpuCoresData[core].c6Delta)
			require.Equal(t, delta, msr.cpuCoresData[core].c7Delta)
			require.Equal(t, delta, msr.cpuCoresData[core].mperfDelta)
			require.Equal(t, delta, msr.cpuCoresData[core].aperfDelta)
			require.Equal(t, delta, msr.cpuCoresData[core].timeStampCounterDelta)
		}
	}
}

func getMsrServiceWithMockedFs() (*msrServiceImpl, *mockFileService) {
	cores := []string{"cpu0", "cpu1", "cpu2", "cpu3"}
	logger := testutil.Logger{Name: "PowerPluginTest"}
	fsMock := &mockFileService{}
	fsMock.On("getStringsMatchingPatternOnPath", mock.Anything).
		Return(cores, nil).Once()
	msr := newMsrServiceWithFs(logger, fsMock)

	return msr, fsMock
}
