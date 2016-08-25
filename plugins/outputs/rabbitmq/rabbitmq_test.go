package rabbitmq

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s, _ := serializers.NewInfluxSerializer()
	rmq := &RabbitMQ{
		Username:     "guest",
		Password:     "guest",
		RabbitmqHost: fmt.Sprintf("%v", testutil.GetLocalHost()),
		RabbitmqPort: "5672",
		Queue:        "test_queue",
		serializer:   s,
	}

	// Verify that we can connect to the RabbitMQ broker
	err := rmq.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the RabbitMQ broker
	err = rmq.Write(testutil.MockMetrics())
	require.NoError(t, err)

	// Verify that we can close the RabbitMQ connection
	err = rmq.Close()
	require.NoError(t, err)
}
