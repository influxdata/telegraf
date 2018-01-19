package stan

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := []string{"nats://" + testutil.GetLocalHost() + ":4222"}
	s, _ := serializers.NewInfluxSerializer()
	n := &STAN{
		Servers:    server,
		ClusterID:  "test-cluster",
		ClientID:   "telegraf-test-client",
		Subject:    "telegraf",
		serializer: s,
	}

	// Verify that we can connect to the NATS Streaming daemon
	err := n.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the NATS Streaming daemon
	err = n.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
