package nsq

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "nsqio/nsq",
		ExposedPorts: []string{"4150"},
		Entrypoint:   []string{"/nsqd"},
		WaitingFor:   wait.ForListeningPort("4150/tcp"),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	server := []string{fmt.Sprintf("%s:%s", container.Address, container.Port)}
	s, _ := serializers.NewInfluxSerializer()
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
