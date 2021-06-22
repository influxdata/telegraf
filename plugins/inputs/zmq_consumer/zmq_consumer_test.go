package zmq_consumer

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers/value"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		plugin := &zmqConsumer{
			Endpoints: []string{""},
		}

		err := plugin.Init()

		require.NoError(t, err)

		require.Equal(t, plugin.HighWaterMark, defaultHighWaterMark)
		require.Equal(t, plugin.Affinity, 0)
		require.Equal(t, plugin.BufferSize, 0)
		require.Equal(t, plugin.MaxUndeliveredMessages, defaultMaxUndeliveredMessages)
		require.Len(t, plugin.Subscriptions, 1)
		require.Contains(t, plugin.Subscriptions, defaultSubscription)
	})

	t.Run("invalid endpoints", func(t *testing.T) {
		plugin := &zmqConsumer{}

		err := plugin.Init()

		require.Error(t, err)
	})
}

func TestStartStop(t *testing.T) {
	parser := value.NewValueParser("temp", "int", "", nil)
	plugin := &zmqConsumer{
		Endpoints:     []string{"tcp://localhost:6001"},
		Subscriptions: []string{""},
		Log:           testutil.Logger{},
		parser:        parser,
	}

	err := plugin.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	plugin.Stop()
}
