package redists

import (
	"testing"
	
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	address := testutil.GetLocalHost() + ":6379"
	redis := &RedisTS{
		Addr: address,
	}

	// Verify that we can connect to the RedisTS server
	err := redis.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the RedisTS server
	err = redis.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
