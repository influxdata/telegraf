package stomp

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

// TestiConnectAndWrite ...
func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var url = testutil.GetLocalHost() + ":61613"
	s, err := serializers.NewJsonSerializer(10)
	require.NoError(t, err)

	st := &STOMP{
		Host:      url,
		Username:  "",
		Password:  "",
		QueueName: "test_queue",
		SSL:       false,
		serialize: s,
	}
	err = st.Connect()
	require.NoError(t, err)

	err = st.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
