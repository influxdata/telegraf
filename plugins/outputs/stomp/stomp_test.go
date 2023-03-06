package stomp

import (
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
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
	defer container.Terminate()
	var url = fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort])
	s, err := serializers.NewJSONSerializer(
		&serializers.Config{
			TimestampUnits:  10 * time.Second,
			TimestampFormat: "yyy-dd-mmThh:mm:ss",
		})
	require.NoError(t, err)
	st := &STOMP{
		Host:          url,
		QueueName:     "test_queue",
		HeartBeatSend: 0,
		HeartBeatRec:  0,
		Log:           testutil.Logger{},
		serialize:     s,
	}
	require.NoError(t, st.Connect())

	require.NoError(t, st.Write(testutil.MockMetrics()))
}
