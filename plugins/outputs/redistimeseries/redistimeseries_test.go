package redistimeseries

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/testutil"
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

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	servicePort := "6379"
	container := testutil.Container{
		Image:        "redislabs/redistimeseries",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForListeningPort(nat.Port(servicePort)),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	redis := &RedisTimeSeries{
		Address: fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort]),
	}
	// Verify that we can connect to the RedisTimeSeries server
	require.NoError(t, redis.Connect())
	// Verify that we can successfully write data to the RedisTimeSeries server
	require.NoError(t, redis.Write(testutil.MockMetrics()))
}
