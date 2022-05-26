package mqtt

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/stretchr/testify/require"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "1883"
	container := testutil.Container{
		Image:        "ncarlier/mqtt",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForListeningPort(nat.Port(servicePort)),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	var url = fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])
	s, _ := serializers.NewInfluxSerializer()
	m := &MQTT{
		Servers:    []string{url},
		serializer: s,
		KeepAlive:  30,
	}

	// Verify that we can connect to the MQTT broker
	err = m.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the mqtt broker
	err = m.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
