package amqp10

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	// Use port 5673 instead of 5672 to not interfere with existing RabbitMQ tests
	brokers := []string{"amqp://" + testutil.GetLocalHost() + ":5673"}
	s, _ := serializers.NewInfluxSerializer()
	k := &AMQP10{
		Brokers:    brokers,
		Topic:      "Test",
		serializer: s,
		Username:   "admin",
		Password:   "admin",
		Timeout:    internal.Duration{Duration: time.Second * 5},
	}

	// Verify that we can connect to the AMQP1.0 broker
	err := k.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the AMQP1.0 broker
	err = k.Write(testutil.MockMetrics())
	require.NoError(t, err)
	k.Close()
}
