package nats

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "4222"
	container := testutil.Container{
		Image:        "nats",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForLog("Server is ready"),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	server := []string{fmt.Sprintf("nats://%s:%s", container.Address, container.Ports[servicePort])}
	s := serializers.NewInfluxSerializer()
	n := &NATS{
		Servers:    server,
		Name:       "telegraf",
		Subject:    "telegraf",
		serializer: s,
	}

	// Verify that we can connect to the NATS daemon
	err = n.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the NATS daemon
	err = n.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
