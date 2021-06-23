package zmq_consumer

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/value"
	"github.com/influxdata/telegraf/testutil"
	zmq "github.com/pebbe/zmq4"
	"github.com/stretchr/testify/require"
)

func runPublisher(ctx context.Context, msg string) (string, error) {
	// create a PUB socket
	socket, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return "", err
	}

	err = socket.Bind("tcp://*:*")
	if err != nil {
		return "", err
	}

	endpoint, err := socket.GetLastEndpoint()
	if err != nil {
		return "", err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				socket.SetLinger(0)
				socket.Close()
				socket.Unbind(endpoint)
				socket = nil
				return
			default:
				if socket != nil {
					socket.Send(msg, 0)
					time.Sleep(50 * time.Millisecond)
				}
			}
		}
	}()

	return endpoint, nil
}

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

func TestReceiver(t *testing.T) {
	temp := 81

	ctx, cancel := context.WithCancel(context.Background())

	endpoint, err := runPublisher(ctx, strconv.Itoa(temp))
	require.NoError(t, err)

	parser := value.NewValueParser("temp", "int", "", nil)
	plugin := &zmqConsumer{
		Endpoints:              []string{endpoint},
		Subscriptions:          []string{""},
		MaxUndeliveredMessages: 1000,
		Log:                    testutil.Logger{},
		parser:                 parser,
	}

	err = plugin.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	acc.Wait(1)

	cancel()
	plugin.Stop()

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"temp",
			map[string]string{},
			map[string]interface{}{
				"value": temp,
			},
			time.Now(),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
