// +build windows

//these tests must be run under administrator account
package win_services

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var InvalidServices = []string{"XYZ1@", "ZYZ@", "SDF_@#"}
var KnownServices = []string{"LanmanServer", "TermService"}

func TestListIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	provider := &MgProvider{}
	scmgr, err := provider.Connect()
	require.NoError(t, err)
	defer scmgr.Disconnect()

	winServices := &WinServices{
		ServiceNames: KnownServices,
	}
	winServices.Init()
	services, err := winServices.listServices(scmgr)
	require.NoError(t, err)
	require.Len(t, services, 2, "Different number of services")
	require.Equal(t, services[0], KnownServices[0])
	require.Equal(t, services[1], KnownServices[1])
}

func TestEmptyListIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	provider := &MgProvider{}
	scmgr, err := provider.Connect()
	require.NoError(t, err)
	defer scmgr.Disconnect()

	winServices := &WinServices{
		ServiceNames: []string{},
	}
	winServices.Init()
	services, err := winServices.listServices(scmgr)
	require.NoError(t, err)
	require.Condition(t, func() bool { return len(services) > 20 }, "Too few service")
}

func TestGatherErrorsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ws := &WinServices{
		Log:          testutil.Logger{},
		ServiceNames: InvalidServices,
		mgrProvider:  &MgProvider{},
	}
	ws.Init()
	require.Len(t, ws.ServiceNames, 3, "Different number of services")
	var acc testutil.Accumulator
	require.NoError(t, ws.Gather(&acc))
	require.Len(t, acc.Errors, 3, "There should be 3 errors after gather")
}
