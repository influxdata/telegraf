package kafka

import (
	"testing"

	"github.com/koksan83/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	brokers := []string{testutil.GetLocalHost() + ":9092"}
	k := &Kafka{
		Brokers: brokers,
		Topic:   "Test",
	}

	// Verify that we can connect to the Kafka broker
	err := k.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the kafka broker
	err = k.Write(testutil.MockBatchPoints())
	require.NoError(t, err)
}
