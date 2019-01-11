package pulsar

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

	server := "pulsar://" + testutil.GetLocalHost() + ":6650"
	s, _ := serializers.NewInfluxSerializer()
	p := &Pulsar{
		URL: server,
		Producer: &ProducerOpts{
			Topic: "telegraf",
		},

		serializer: s,
	}

	// Verify that we can connect to Pulsar cluster
	err := p.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to Pulsar cluster
	err = p.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
