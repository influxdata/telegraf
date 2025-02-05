package eh

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

/*
** Integration test (requires an Event Hubs instance)
 */

func TestInit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	connstring := "Endpoint=sb://telegraf.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=+"

	// Configure the plugin to target the newly created hub
	e := &EventHubs{
		Hub:              &EventHub{},
		ConnectionString: connstring,
		Timeout:          config.Duration(time.Second * 5),
	}

	err := e.Init()
	// Verify that we can connect to Event Hubs
	require.NoError(t, err)
	e.Close()
}
