package mqtt

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

// Test that default client has random ID
func TestRandomClientID(t *testing.T) {
	var err error

	cfg := &MqttConfig{
		Servers: []string{"tcp://localhost:1883"},
	}

	client1, err := NewMQTTv311Client(cfg)
	require.NoError(t, err)

	client2, err := NewMQTTv311Client(cfg)
	require.NoError(t, err)

	options1 := client1.client.OptionsReader()
	options2 := client2.client.OptionsReader()
	require.NotEqual(t, options1.ClientID(), options2.ClientID())
}

func TestConnectV5UnreachableBroker(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())

	client, err := NewMQTTv5Client(&MqttConfig{
		Servers:           []string{"tcp://" + addr},
		ConnectionTimeout: config.Duration(time.Second),
	})
	require.NoError(t, err)

	_, err = client.Connect()
	require.ErrorContains(t, err, "failed to connect to tcp://"+addr)
}
