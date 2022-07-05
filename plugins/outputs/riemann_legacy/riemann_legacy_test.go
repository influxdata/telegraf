package riemann_legacy

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "5555"
	container := testutil.Container{
		Image:        "rlister/riemann",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForLog("Hyperspace core online"),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	url := fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])

	r := &Riemann{
		URL:       url,
		Transport: "tcp",
		Log:       testutil.Logger{},
	}

	err = r.Connect()
	require.NoError(t, err)

	err = r.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
