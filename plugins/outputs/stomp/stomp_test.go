package stomp

import (
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/plugins/serializers"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "61613"
	container := testutil.Container{
		Image:        "rmohr/activemq",
		ExposedPorts: []string{servicePort},
		WaitingFor:   wait.ForListeningPort(nat.Port(servicePort)),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()
	var url = fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])
	s, err := serializers.NewJSONSerializer(10*time.Second, "yyy-dd-mmThh:mm:ss")
	require.NoError(t, err)
	st := &STOMP{
		Host:          url,
		QueueName:     "test_queue",
		HeartBeatSend: 0,
		HeartBeatRec:  0,
		serialize:     s,
	}
	err = st.Connect()
	require.NoError(t, err)

	err = st.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
