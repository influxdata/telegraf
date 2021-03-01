package redistimeseries

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	address := testutil.GetLocalHost() + ":6379"
	redis := &RedisTimeSeries{
		Address: address,
	}

	// Verify that we can connect to the RedisTimeSeries server
	err := redis.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the RedisTimeSeries server
	err = redis.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
