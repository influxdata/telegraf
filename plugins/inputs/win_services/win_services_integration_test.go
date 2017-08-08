// +build windows

//these tests must be run under administrator account
package win_services

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc/mgr"
	"testing"
)

var InvalidServices = []string{"XYZ1@", "ZYZ@", "SDF_@#"}
var KnownServices = []string{"LanmanServer", "TermService"}

func TestList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	services, err := listServices(&MgProvider{}, KnownServices)
	require.NoError(t, err)
	assert.Len(t, services, 2, "Different number of services")
	assert.Equal(t, services[0].ServiceName, KnownServices[0])
	assert.Nil(t, services[0].Error)
	assert.Equal(t, services[1].ServiceName, KnownServices[1])
	assert.Nil(t, services[1].Error)
}

func TestEmptyList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	services, err := listServices(&MgProvider{}, []string{})
	require.NoError(t, err)
	assert.Condition(t, func() bool { return len(services) > 20 }, "Too few service")
}

func TestListEr(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	services, err := listServices(&MgProvider{}, InvalidServices)
	require.NoError(t, err)
	assert.Len(t, services, 3, "Different number of services")
	for i := 0; i < 3; i++ {
		assert.Equal(t, services[i].ServiceName, InvalidServices[i])
		assert.NotNil(t, services[i].Error)
	}
}

func TestGather(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ws := &WinServices{KnownServices, &MgProvider{}}
	assert.Len(t, ws.ServiceNames, 2, "Different number of services")
	var acc testutil.Accumulator
	require.NoError(t, ws.Gather(&acc))
	assert.Len(t, acc.Errors, 0, "There should be no errors after gather")

	for i := 0; i < 2; i++ {
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		si := getServiceInfo(KnownServices[i])
		fields["state"] = int(si.State)
		fields["startup_mode"] = int(si.StartUpMode)
		tags["service_name"] = si.ServiceName
		tags["display_name"] = si.DisplayName
		acc.AssertContainsTaggedFields(t, "win_services", fields, tags)
	}
}

func TestGatherErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ws := &WinServices{InvalidServices, &MgProvider{}}
	assert.Len(t, ws.ServiceNames, 3, "Different number of services")
	var acc testutil.Accumulator
	require.NoError(t, ws.Gather(&acc))
	assert.Len(t, acc.Errors, 3, "There should be 3 errors after gather")
}

func getServiceInfo(srvName string) *ServiceInfo {

	scmgr, err := mgr.Connect()
	if err != nil {
		return nil
	}
	defer scmgr.Disconnect()

	srv, err := scmgr.OpenService(srvName)
	if err != nil {
		return nil
	}
	var si ServiceInfo
	si.ServiceName = srvName
	srvStatus, err := srv.Query()
	if err == nil {
		si.State = int(srvStatus.State)
	} else {
		si.Error = err
	}

	srvCfg, err := srv.Config()
	if err == nil {
		si.DisplayName = srvCfg.DisplayName
		si.StartUpMode = int(srvCfg.StartType)
	} else {
		si.Error = err
	}
	srv.Close()
	return &si
}
