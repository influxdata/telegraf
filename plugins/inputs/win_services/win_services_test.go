// +build windows

package win_services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

//testData is DD wrapper for unit testing of WinServices
type testData struct {
	//collection that will be returned in ListServices if service array passed into WinServices constructor is empty
	queryServiceList     []string
	mgrConnectError      error
	mgrListServicesError error
	services             []serviceTestInfo
}

type serviceTestInfo struct {
	serviceOpenError   error
	serviceQueryError  error
	serviceConfigError error
	serviceName        string
	displayName        string
	state              int
	startUpMode        int
}

type FakeSvcMgr struct {
	testData testData
}

func (m *FakeSvcMgr) Disconnect() error {
	return nil
}

func (m *FakeSvcMgr) OpenService(name string) (WinService, error) {
	for _, s := range m.testData.services {
		if s.serviceName == name {
			if s.serviceOpenError != nil {
				return nil, s.serviceOpenError
			} else {
				return &FakeWinSvc{s}, nil
			}
		}
	}
	return nil, fmt.Errorf("Cannot find service %s", name)
}

func (m *FakeSvcMgr) ListServices() ([]string, error) {
	if m.testData.mgrListServicesError != nil {
		return nil, m.testData.mgrListServicesError
	} else {
		return m.testData.queryServiceList, nil
	}
}

type FakeMgProvider struct {
	testData testData
}

func (m *FakeMgProvider) Connect() (WinServiceManager, error) {
	if m.testData.mgrConnectError != nil {
		return nil, m.testData.mgrConnectError
	} else {
		return &FakeSvcMgr{m.testData}, nil
	}
}

type FakeWinSvc struct {
	testData serviceTestInfo
}

func (m *FakeWinSvc) Close() error {
	return nil
}
func (m *FakeWinSvc) Config() (mgr.Config, error) {
	if m.testData.serviceConfigError != nil {
		return mgr.Config{}, m.testData.serviceConfigError
	} else {
		return mgr.Config{
			ServiceType:      0,
			StartType:        uint32(m.testData.startUpMode),
			ErrorControl:     0,
			BinaryPathName:   "",
			LoadOrderGroup:   "",
			TagId:            0,
			Dependencies:     nil,
			ServiceStartName: m.testData.serviceName,
			DisplayName:      m.testData.displayName,
			Password:         "",
			Description:      "",
		}, nil
	}
}
func (m *FakeWinSvc) Query() (svc.Status, error) {
	if m.testData.serviceQueryError != nil {
		return svc.Status{}, m.testData.serviceQueryError
	} else {
		return svc.Status{
			State:      svc.State(m.testData.state),
			Accepts:    0,
			CheckPoint: 0,
			WaitHint:   0,
		}, nil
	}
}

var testErrors = []testData{
	{nil, errors.New("Fake mgr connect error"), nil, nil},
	{nil, nil, errors.New("Fake mgr list services error"), nil},
	{[]string{"Fake service 1", "Fake service 2", "Fake service 3"}, nil, nil, []serviceTestInfo{
		{errors.New("Fake srv open error"), nil, nil, "Fake service 1", "", 0, 0},
		{nil, errors.New("Fake srv query error"), nil, "Fake service 2", "", 0, 0},
		{nil, nil, errors.New("Fake srv config error"), "Fake service 3", "", 0, 0},
	}},
	{nil, nil, nil, []serviceTestInfo{
		{errors.New("Fake srv open error"), nil, nil, "Fake service 1", "", 0, 0},
	}},
}

func TestBasicInfo(t *testing.T) {

	winServices := &WinServices{nil, &FakeMgProvider{testErrors[0]}}
	assert.NotEmpty(t, winServices.SampleConfig())
	assert.NotEmpty(t, winServices.Description())
}

func TestMgrErrors(t *testing.T) {
	//mgr.connect error
	winServices := &WinServices{nil, &FakeMgProvider{testErrors[0]}}
	var acc1 testutil.Accumulator
	err := winServices.Gather(&acc1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), testErrors[0].mgrConnectError.Error())

	////mgr.listServices error
	winServices = &WinServices{nil, &FakeMgProvider{testErrors[1]}}
	var acc2 testutil.Accumulator
	err = winServices.Gather(&acc2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), testErrors[1].mgrListServicesError.Error())

	////mgr.listServices error 2
	winServices = &WinServices{[]string{"Fake service 1"}, &FakeMgProvider{testErrors[3]}}
	var acc3 testutil.Accumulator
	err = winServices.Gather(&acc3)
	require.NoError(t, err)
	assert.Len(t, acc3.Errors, 1)

}

func TestServiceErrors(t *testing.T) {
	winServices := &WinServices{nil, &FakeMgProvider{testErrors[2]}}
	var acc1 testutil.Accumulator
	require.NoError(t, winServices.Gather(&acc1))
	assert.Len(t, acc1.Errors, 3)
	//open service error
	assert.Contains(t, acc1.Errors[0].Error(), testErrors[2].services[0].serviceOpenError.Error())
	//query service error
	assert.Contains(t, acc1.Errors[1].Error(), testErrors[2].services[1].serviceQueryError.Error())
	//config service error
	assert.Contains(t, acc1.Errors[2].Error(), testErrors[2].services[2].serviceConfigError.Error())

}

var testSimpleData = []testData{
	{[]string{"Service 1", "Service 2"}, nil, nil, []serviceTestInfo{
		{nil, nil, nil, "Service 1", "Fake service 1", 1, 2},
		{nil, nil, nil, "Service 2", "Fake service 2", 1, 2},
	}},
}

func TestGather2(t *testing.T) {
	winServices := &WinServices{nil, &FakeMgProvider{testSimpleData[0]}}
	var acc1 testutil.Accumulator
	require.NoError(t, winServices.Gather(&acc1))
	assert.Len(t, acc1.Errors, 0, "There should be no errors after gather")

	for _, s := range testSimpleData[0].services {
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		fields["state"] = int(s.state)
		fields["startup_mode"] = int(s.startUpMode)
		tags["service_name"] = s.serviceName
		tags["display_name"] = s.displayName
		acc1.AssertContainsTaggedFields(t, "win_services", fields, tags)
	}

}
