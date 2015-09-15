package amqp

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var url = "amqp://" + testutil.GetLocalHost() + ":5672/"
	q := &AMQP{
		URL:      url,
		Exchange: "telegraf_test",
	}

	// Verify that we can connect to the Kafka broker
	err := q.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the kafka broker
	err = q.Write(testutil.MockBatchPoints())
	require.NoError(t, err)
}
