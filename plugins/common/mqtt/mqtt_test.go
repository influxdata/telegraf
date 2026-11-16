package mqtt

import (
	"testing"

	mqttv5 "github.com/eclipse/paho.golang/paho"
	"github.com/stretchr/testify/require"
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

func TestWillMessageV311(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *MqttConfig
		enabled  bool
		topic    string
		payload  string
		qos      byte
		retained bool
	}{
		{
			name: "no will",
			cfg: &MqttConfig{
				Servers:     []string{"tcp://localhost:1883"},
				WillPayload: "ignored",
				WillQoS:     1,
				WillRetain:  true,
			},
		},
		{
			name: "will set",
			cfg: &MqttConfig{
				Servers:     []string{"tcp://localhost:1883"},
				WillTopic:   "telegraf/status",
				WillPayload: "offline",
				WillQoS:     1,
				WillRetain:  true,
			},
			enabled:  true,
			topic:    "telegraf/status",
			payload:  "offline",
			qos:      1,
			retained: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMQTTv311Client(tt.cfg)
			require.NoError(t, err)

			options := client.client.OptionsReader()
			require.Equal(t, tt.enabled, options.WillEnabled())
			if !tt.enabled {
				return
			}
			require.Equal(t, tt.topic, options.WillTopic())
			require.Equal(t, []byte(tt.payload), options.WillPayload())
			require.Equal(t, tt.qos, options.WillQos())
			require.Equal(t, tt.retained, options.WillRetained())
		})
	}
}

func TestWillMessageV5(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *MqttConfig
		expected *mqttv5.WillMessage
	}{
		{
			name: "no will",
			cfg: &MqttConfig{
				Servers:     []string{"tcp://localhost:1883"},
				Protocol:    "5",
				WillPayload: "ignored",
				WillQoS:     1,
				WillRetain:  true,
			},
		},
		{
			name: "will set",
			cfg: &MqttConfig{
				Servers:     []string{"tcp://localhost:1883"},
				Protocol:    "5",
				WillTopic:   "telegraf/status",
				WillPayload: "offline",
				WillQoS:     2,
				WillRetain:  true,
			},
			expected: &mqttv5.WillMessage{
				Topic:   "telegraf/status",
				Payload: []byte("offline"),
				QoS:     2,
				Retain:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMQTTv5Client(tt.cfg)
			require.NoError(t, err)
			require.Equal(t, tt.expected, client.options.WillMessage)
		})
	}
}

func TestWillMessageInvalid(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		qos      int
		expected string
	}{
		{
			name:     "qos too high",
			topic:    "telegraf/status",
			qos:      3,
			expected: "invalid will QoS value 3",
		},
		{
			name:     "qos negative",
			topic:    "telegraf/status",
			qos:      -1,
			expected: "invalid will QoS value -1",
		},
		{
			name:     "multi-level wildcard",
			topic:    "telegraf/#",
			expected: "must not contain wildcards",
		},
		{
			name:     "single-level wildcard",
			topic:    "telegraf/+/status",
			expected: "must not contain wildcards",
		},
		{
			name:  "valid",
			topic: "telegraf/status",
			qos:   2,
		},
		{
			name: "qos ignored without topic",
			qos:  5,
		},
	}

	for _, tt := range tests {
		for _, protocol := range []string{"3.1.1", "5"} {
			t.Run(tt.name+"/"+protocol, func(t *testing.T) {
				cfg := &MqttConfig{
					Servers:   []string{"tcp://localhost:1883"},
					Protocol:  protocol,
					WillTopic: tt.topic,
					WillQoS:   tt.qos,
				}
				client, err := NewClient(cfg)
				if tt.expected != "" {
					require.ErrorContains(t, err, tt.expected)
					return
				}
				require.NoError(t, err)
				require.NotNil(t, client)
			})
		}
	}
}
