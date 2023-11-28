package nats

import (
	"fmt"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteNATSCoreIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "4222/tcp"
	container := testutil.Container{
		Image:        "nats",
		Name:         "nats-core",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForAll(wait.ForLog("Server is ready")),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	server := []string{fmt.Sprintf("nats://%s:%s", container.Address, container.Ports[servicePort])}
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	n := &NATS{
		Servers:    server,
		Name:       "telegraf",
		Subject:    "telegraf",
		serializer: serializer,
	}
	// Verify that we can connect to the NATS daemon
	err = n.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the NATS daemon
	err = n.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestConnectAndWriteNATSJetstreamIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "4222/tcp"
	container := testutil.Container{
		Image:        "nats",
		Name:         "nats-js",
		Cmd:          []string{"--js"},
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForAll(wait.ForLog("Server is ready")),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	server := []string{fmt.Sprintf("nats://%s:%s", container.Address, container.Ports[servicePort])}
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	n := &NATS{
		Servers: server,
		Name:    "telegraf",
		Subject: "telegraf",
		Jetstream: &JetstreamConfig{
			AutoCreateStream: true,
			StreamConfig: jetstream.StreamConfig{
				Name:     "my-telegraf-stream",
				Subjects: []string{"telegraf"},
			},
		},
		serializer: serializer,
	}
	// Verify that we can connect to the NATS daemon
	err = n.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the NATS daemon
	err = n.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
