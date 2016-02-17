package amqp

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

	var url = "amqp://" + testutil.GetLocalHost() + ":5672/"
	s, _ := serializers.NewInfluxSerializer()
	q := &AMQP{
		URL:        url,
		Exchange:   "telegraf_test",
		serializer: s,
	}

	// Verify that we can connect to the AMQP broker
	err := q.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the amqp broker
	err = q.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
