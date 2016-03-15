package zabbix

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	host := testutil.GetLocalHost()

	z := &Zabbix{
		Host:    host,
		Port:    10051,
		Hosttag: "host",
	}

	err := z.Connect()
	require.NoError(t, err)

	err = z.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
