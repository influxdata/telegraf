// +build linux

package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/mocks"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCalculateMsrOffsets(t *testing.T) {
	offsetConstNumber := 8
	msr, _ := getMsrServiceWithMockedFs()
	require.Equal(t, offsetConstNumber, len(msr.getMsrOffsets()))
}

func TestReadDataFromMsrPositive(t *testing.T) {
	firstValue := uint64(1000000)
	secondValue := uint64(5000000)
	delta := secondValue - firstValue
	cpuCores := []string{"cpu0", "cpu1"}
	msr, fsMock := getMsrServiceWithMockedFs()
	prepareTestData(fsMock, cpuCores, msr, t)
	cores := trimCPUFromCores(cpuCores)

	methodCallNumberForFirstValue := len(msr.getMsrOffsets()) * len(cores)
	methodCallNumberForSecondValue := methodCallNumberForFirstValue * 2

	fsMock.On("ReadFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(firstValue, nil).Times(methodCallNumberForFirstValue)
	for _, core := range cores {
		require.NoError(t, msr.readDataFromMsr(core, nil))
	}
	fsMock.AssertNumberOfCalls(t, "ReadFileAtOffsetToUint64", methodCallNumberForFirstValue)
	verifyCPUCoresData(cores, t, msr, firstValue, false, 0)

	fsMock.On("ReadFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(secondValue, nil).Times(methodCallNumberForFirstValue)
	for _, core := range cores {
		require.NoError(t, msr.readDataFromMsr(core, nil))
	}
	fsMock.AssertNumberOfCalls(t, "ReadFileAtOffsetToUint64", methodCallNumberForSecondValue)
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

	methodCallNumberPerCore := len(msr.getMsrOffsets())

	// Normal execution for first core.
	fsMock.On("ReadFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(firstValue, nil).Times(methodCallNumberPerCore).
		// Fail to read file for second core.
		On("ReadFileAtOffsetToUint64", mock.Anything, mock.Anything).
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

	fsMock.On("ReadFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(zero, errors.New("error reading file")).Once()
	require.Error(t, msr.readValueFromFileAtOffset(ctx, testChannel, nil, 0))

	fsMock.On("ReadFileAtOffsetToUint64", mock.Anything, mock.Anything).
		Return(zero, nil).Once()
	require.Equal(t, nil, msr.readValueFromFileAtOffset(ctx, testChannel, nil, 0))
	require.Equal(t, zero, <-testChannel)
}

func prepareTestData(fsMock *mocks.FileService, cores []string, msr *MsrServiceImpl, t *testing.T) {
	// Prepare MSR offsets and CPUCoresData for test.
	fsMock.On("GetStringsMatchingPatternOnPath", mock.Anything).
		Return(cores, nil).Once()
	msr.calculateMsrOffsets()
	require.NoError(t, msr.setCPUCores())
	fsMock.AssertCalled(t, "GetStringsMatchingPatternOnPath", mock.Anything)
}

func verifyCPUCoresData(cores []string, t *testing.T, msr *MsrServiceImpl, expectedValue uint64, verifyDelta bool, delta uint64) {
	for _, core := range cores {
		require.Equal(t, expectedValue, msr.cpuCoresData[core].C3)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].C6)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].C7)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].Mperf)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].Aperf)
		require.Equal(t, expectedValue, msr.cpuCoresData[core].Tsc)
		require.Equal(t, (expectedValue>>16)&0xFF, msr.cpuCoresData[core].ThrottleTemp)
		require.Equal(t, (expectedValue>>16)&0x7F, msr.cpuCoresData[core].Temp)

		if verifyDelta {
			require.Equal(t, delta, msr.cpuCoresData[core].C3Delta)
			require.Equal(t, delta, msr.cpuCoresData[core].C6Delta)
			require.Equal(t, delta, msr.cpuCoresData[core].C7Delta)
			require.Equal(t, delta, msr.cpuCoresData[core].MperfDelta)
			require.Equal(t, delta, msr.cpuCoresData[core].AperfDelta)
			require.Equal(t, delta, msr.cpuCoresData[core].TscDelta)
		}
	}
}

func getMsrServiceWithMockedFs() (*MsrServiceImpl, *mocks.FileService) {
	cores := []string{"cpu0", "cpu1", "cpu2", "cpu3"}
	logger := testutil.Logger{Name: "PowerPluginTest"}
	fsMock := &mocks.FileService{}
	fsMock.On("GetStringsMatchingPatternOnPath", mock.Anything).
		Return(cores, nil).Once()
	msr := NewMsrServiceWithFs(logger, fsMock)

	return msr, fsMock
}
