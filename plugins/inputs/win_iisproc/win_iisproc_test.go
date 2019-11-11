package win_iisproc

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var correctAppPooldata = `<appcmd>	<WP WP.NAME="1001" APPPOOL.NAME="AppPool1" />
<WP WP.NAME="1002" APPPOOL.NAME="AppPool2" /></appcmd>`

var correctPSData = `ps:="Id","CPU","PM"
"1000","19.078125","276463616"
"1001","4.515625","155299840
`

type FakeCommandData struct {
	hasError  bool
	data      string
	fileExist bool
}

type FakeCommandManager struct {
	commandData    FakeCommandData
	powershellData FakeCommandData
}

func (c *FakeCommandManager) RunCmdCommand(arg ...string) ([]byte, error) {
	if c.commandData.hasError {
		return nil, errors.New("An error occured")
	}
	return ([]byte(c.commandData.data)), nil

}
func (c *FakeCommandManager) RunPowershellCommand(arg ...string) ([]byte, error) {
	if c.powershellData.hasError {
		return nil, errors.New("powershell error occured")
	}
	return ([]byte(c.powershellData.data)), nil

}

func (c *FakeCommandManager) FileExists(fileName string) bool {
	return c.commandData.fileExist
}

func TestGather(t *testing.T) {
	iisProc := &IISProc{commandMgr: &FakeCommandManager{commandData: FakeCommandData{hasError: false, data: correctAppPooldata, fileExist: true}, powershellData: FakeCommandData{hasError: false, data: correctPSData}}}
	var acc1 testutil.Accumulator
	require.NoError(t, iisProc.Gather(&acc1))
}

func TestDescription(t *testing.T) {
	var desc = "Input plugin to report IIS worker processes per app pool."
	iisProc := &IISProc{commandMgr: &FakeCommandManager{commandData: FakeCommandData{hasError: false, data: correctAppPooldata, fileExist: true}, powershellData: FakeCommandData{hasError: false, data: correctPSData}}}
	var acc1 testutil.Accumulator
	require.NoError(t, iisProc.Gather(&acc1))
	assert.Equal(t, desc, iisProc.Description())

}
func TestRunCommand(t *testing.T) {

	iisProc := &IISProc{commandMgr: &FakeCommandManager{commandData: FakeCommandData{hasError: true, data: correctAppPooldata, fileExist: true}, powershellData: FakeCommandData{hasError: false, data: correctPSData}}}
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	var acc1 testutil.Accumulator
	err := iisProc.Gather(&acc1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "An error occured")
}
func TestPowershellCommand(t *testing.T) {
	iisProc := &IISProc{commandMgr: &FakeCommandManager{commandData: FakeCommandData{hasError: false, data: correctAppPooldata, fileExist: true}, powershellData: FakeCommandData{hasError: true, data: correctPSData}}}
	var acc1 testutil.Accumulator
	err := iisProc.Gather(&acc1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "powershell error occured")
}

func TestIISIsInstalled(t *testing.T) {
	iisProc := &IISProc{commandMgr: &FakeCommandManager{commandData: FakeCommandData{hasError: false, data: correctAppPooldata, fileExist: false}, powershellData: FakeCommandData{hasError: true, data: correctPSData}}}
	var acc1 testutil.Accumulator
	err := iisProc.Gather(&acc1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "appcmd.exe doest not exists.You need to turn on Web Management Tools")
}
