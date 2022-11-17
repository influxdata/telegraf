package nsq

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "4150"
	container := testutil.Container{
		Image:        "nsqio/nsq",
		ExposedPorts: []string{servicePort},
		Entrypoint:   []string{"/nsqd"},
		WaitingFor:   wait.ForListeningPort(nat.Port(servicePort)),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	server := []string{fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])}
	s := serializers.NewInfluxSerializer()
	n := &NSQ{
		Server:     server[0],
		Topic:      "telegraf",
		serializer: s,
	}

	// Verify that we can connect to the NSQ daemon
	err = n.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the NSQ daemon
	err = n.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
