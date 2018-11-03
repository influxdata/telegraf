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

func TestList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	provider := &MgProvider{}
	scmgr, err := provider.Connect()
	require.NoError(t, err)
	defer scmgr.Disconnect()

	services, err := listServices(scmgr, KnownServices)
	require.NoError(t, err)
	require.Len(t, services, 2, "Different number of services")
	require.Equal(t, services[0], KnownServices[0])
	require.Equal(t, services[1], KnownServices[1])
}

func TestEmptyList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	provider := &MgProvider{}
	scmgr, err := provider.Connect()
	require.NoError(t, err)
	defer scmgr.Disconnect()

	services, err := listServices(scmgr, []string{})
	require.NoError(t, err)
	require.Condition(t, func() bool { return len(services) > 20 }, "Too few service")
}

func TestGatherErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ws := &WinServices{InvalidServices, &MgProvider{}}
	require.Len(t, ws.ServiceNames, 3, "Different number of services")
	var acc testutil.Accumulator
	require.NoError(t, ws.Gather(&acc))
	require.Len(t, acc.Errors, 3, "There should be 3 errors after gather")
}
