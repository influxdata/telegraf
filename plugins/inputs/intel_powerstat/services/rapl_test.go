// +build linux

package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/data"
	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/mocks"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrepareData(t *testing.T) {
	sockets := []string{"intel-rapl:0", "intel-rapl:1"}
	rapl, fsMock := getRaplWithMockedFs()
	fsMock.On("GetStringsMatchingPatternOnPath", mock.Anything).Return(sockets, nil).Twice()
	rapl.prepareData()
	fsMock.AssertCalled(t, "GetStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, len(sockets), len(rapl.GetSocketIDs()))

	// Verify no data is wiped in the next calls
	socketEnergy := 74563813417.0
	socketID := "0"
	rapl.data[socketID].SocketEnergy = socketEnergy

	rapl.prepareData()
	fsMock.AssertCalled(t, "GetStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, len(sockets), len(rapl.GetSocketIDs()))
	require.Equal(t, socketEnergy, rapl.data[socketID].SocketEnergy)

	// Verify data is wiped once there is no RAPL folders
	fsMock.On("GetStringsMatchingPatternOnPath", mock.Anything).
		Return(nil, errors.New("missing RAPL")).Once()
	rapl.prepareData()
	fsMock.AssertCalled(t, "GetStringsMatchingPatternOnPath", mock.Anything)
	require.Equal(t, 0, len(rapl.GetSocketIDs()))
}

func TestFindDramFolders(t *testing.T) {
	sockets := []string{"0", "1"}
	raplFolders := []string{"intel-rapl:0:1", "intel-rapl:0:2", "intel-rapl:0:3"}
	rapl, fsMock := getRaplWithMockedFs()

	for _, socketID := range sockets {
		rapl.data[socketID] = &data.RaplData{}
	}

	firstPath := fmt.Sprintf(IntelRaplDramNamePartialPath,
		fmt.Sprintf(IntelRaplDramPartialPath, IntelRaplPath, "0", raplFolders[2]))
	secondPath := fmt.Sprintf(IntelRaplDramNamePartialPath,
		fmt.Sprintf(IntelRaplDramPartialPath, IntelRaplPath, "1", raplFolders[1]))

	fsMock.
		On("GetStringsMatchingPatternOnPath", mock.Anything).Return(raplFolders, nil).Twice().
		On("ReadFile", firstPath).Return([]byte("dram"), nil).Once().
		On("ReadFile", secondPath).Return([]byte("dram"), nil).Once().
		On("ReadFile", mock.Anything).Return([]byte("random"), nil)

	rapl.findDramFolders()

	dramFolders := rapl.getDramFolders()
	require.Equal(t, len(sockets), len(dramFolders))
	require.Equal(t, raplFolders[2], dramFolders["0"])
	require.Equal(t, raplFolders[1], dramFolders["1"])
	fsMock.AssertNumberOfCalls(t, "ReadFile", 5)
}

func getRaplWithMockedFs() (*RaplServiceImpl, *mocks.FileService) {
	logger := testutil.Logger{Name: "PowerPluginTest"}
	fsMock := &mocks.FileService{}
	rapl := NewRaplServiceWithFs(logger, fsMock)

	return rapl, fsMock
}
