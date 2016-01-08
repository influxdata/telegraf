package mqtt

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var url = testutil.GetLocalHost() + ":1883"
	m := &MQTT{
		Servers: []string{url},
	}

	// Verify that we can connect to the MQTT broker
	err := m.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the mqtt broker
	err = m.Write(testutil.MockBatchPoints().Points())
	require.NoError(t, err)
}
