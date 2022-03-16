//go:build linux
// +build linux

package intel_powerstat

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestPrepareData(t *testing.T) {
	sockets := []string{"intel-rapl:0", "intel-rapl:1"}
	rapl, fsMock := getRaplWithMockedFs()
	fsMock.On("getStringsMatchingPatternOnPath", mock.Anything).Return(sockets, nil).Twice()
	rapl.prepareData()
	fsMock.AssertCalled(t, "getStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, len(sockets), len(rapl.getRaplData()))

	// Verify no data is wiped in the next calls
	socketEnergy := 74563813417.0
	socketID := "0"
	rapl.data[socketID].socketEnergy = socketEnergy

	rapl.prepareData()
	fsMock.AssertCalled(t, "getStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, len(sockets), len(rapl.getRaplData()))
	require.Equal(t, socketEnergy, rapl.data[socketID].socketEnergy)

	// Verify data is wiped once there is no RAPL folders
	fsMock.On("getStringsMatchingPatternOnPath", mock.Anything).
		Return(nil, errors.New("missing RAPL")).Once()
	rapl.prepareData()
	fsMock.AssertCalled(t, "getStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, 0, len(rapl.getRaplData()))
}

func TestFindDramFolders(t *testing.T) {
	sockets := []string{"0", "1"}
	raplFolders := []string{"intel-rapl:0:1", "intel-rapl:0:2", "intel-rapl:0:3"}
	rapl, fsMock := getRaplWithMockedFs()

	for _, socketID := range sockets {
		rapl.data[socketID] = &raplData{}
	}

	firstPath := fmt.Sprintf(intelRaplDramNamePartialPath,
		fmt.Sprintf(intelRaplDramPartialPath, intelRaplPath, "0", raplFolders[2]))
	secondPath := fmt.Sprintf(intelRaplDramNamePartialPath,
		fmt.Sprintf(intelRaplDramPartialPath, intelRaplPath, "1", raplFolders[1]))

	fsMock.
		On("getStringsMatchingPatternOnPath", mock.Anything).Return(raplFolders, nil).Twice().
		On("readFile", firstPath).Return([]byte("dram"), nil).Once().
		On("readFile", secondPath).Return([]byte("dram"), nil).Once().
		On("readFile", mock.Anything).Return([]byte("random"), nil)

	rapl.findDramFolders()

	require.Equal(t, len(sockets), len(rapl.dramFolders))
	require.Equal(t, raplFolders[2], rapl.dramFolders["0"])
	require.Equal(t, raplFolders[1], rapl.dramFolders["1"])
	fsMock.AssertNumberOfCalls(t, "readFile", 5)
}

func TestCalculateDataOverflowCases(t *testing.T) {
	socketID := "1"
	rapl, fsMock := getRaplWithMockedFs()

	rapl.data[socketID] = &raplData{}
	rapl.data[socketID].socketEnergy = convertMicroJoulesToJoules(23424123.1)
	rapl.data[socketID].dramEnergy = convertMicroJoulesToJoules(345611233.2)
	rapl.data[socketID].readDate = 54123

	interval := int64(54343)
	convertedInterval := convertNanoSecondsToSeconds(interval - rapl.data[socketID].readDate)

	newEnergy := 3343443.4
	maxEnergy := 234324546456.6
	convertedNewEnergy := convertMicroJoulesToJoules(newEnergy)
	convertedMaxNewEnergy := convertMicroJoulesToJoules(maxEnergy)

	maxDramEnergy := 981230834098.3
	newDramEnergy := 4533311.1
	convertedMaxDramEnergy := convertMicroJoulesToJoules(maxDramEnergy)
	convertedDramEnergy := convertMicroJoulesToJoules(newDramEnergy)

	expectedCurrentEnergy := (convertedMaxNewEnergy - rapl.data[socketID].socketEnergy + convertedNewEnergy) / convertedInterval
	expectedDramCurrentEnergy := (convertedMaxDramEnergy - rapl.data[socketID].dramEnergy + convertedDramEnergy) / convertedInterval

	fsMock.
		On("readFileToFloat64", mock.Anything).Return(newEnergy, int64(12321), nil).Once().
		On("readFileToFloat64", mock.Anything).Return(newDramEnergy, interval, nil).Once().
		On("readFileToFloat64", mock.Anything).Return(maxEnergy, int64(64534), nil).Once().
		On("readFileToFloat64", mock.Anything).Return(maxDramEnergy, int64(98342), nil).Once()

	require.NoError(t, rapl.calculateData(socketID, strings.NewReader(mock.Anything), strings.NewReader(mock.Anything),
		strings.NewReader(mock.Anything), strings.NewReader(mock.Anything)))

	require.Equal(t, expectedCurrentEnergy, rapl.data[socketID].socketCurrentEnergy)
	require.Equal(t, expectedDramCurrentEnergy, rapl.data[socketID].dramCurrentEnergy)
}

func getRaplWithMockedFs() (*raplServiceImpl, *mockFileService) {
	logger := testutil.Logger{Name: "PowerPluginTest"}
	fsMock := &mockFileService{}
	rapl := newRaplServiceWithFs(logger, fsMock)

	return rapl, fsMock
}
