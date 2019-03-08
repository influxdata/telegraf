package mqtt_consumer

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

const (
	testMsg    = "cpu_load_short,host=server01 value=23422.0 1422568543702900257\n"
	invalidMsg = "cpu_load_short,host=server01 1422568543702900257\n"
)

func newTestMQTTConsumer() *MQTTConsumer {
	n := &MQTTConsumer{
		Topics:  []string{"telegraf"},
		Servers: []string{"localhost:1883"},
	}

	return n
}

// Test that default client has random ID
func TestRandomClientID(t *testing.T) {
	m1 := &MQTTConsumer{
		Servers: []string{"localhost:1883"}}
	opts, err := m1.createOpts()
	assert.NoError(t, err)

	m2 := &MQTTConsumer{
		Servers: []string{"localhost:1883"}}
	opts2, err2 := m2.createOpts()
	assert.NoError(t, err2)

	assert.NotEqual(t, opts.ClientID, opts2.ClientID)
}

// Test that default client has random ID
func TestClientID(t *testing.T) {
	m1 := &MQTTConsumer{
		Servers:  []string{"localhost:1883"},
		ClientID: "telegraf-test",
	}
	opts, err := m1.createOpts()
	assert.NoError(t, err)

	m2 := &MQTTConsumer{
		Servers:  []string{"localhost:1883"},
		ClientID: "telegraf-test",
	}
	opts2, err2 := m2.createOpts()
	assert.NoError(t, err2)

	assert.Equal(t, "telegraf-test", opts2.ClientID)
	assert.Equal(t, "telegraf-test", opts.ClientID)
}

// Test that Start() fails if client ID is not set but persistent is
func TestPersistentClientIDFail(t *testing.T) {
	m1 := &MQTTConsumer{
		Servers:           []string{"localhost:1883"},
		PersistentSession: true,
	}
	acc := testutil.Accumulator{}
	err := m1.Start(&acc)
	assert.Error(t, err)
}

func mqttMsg(val string) mqtt.Message {
	return &message{
		topic:   "telegraf/unit_test",
		payload: []byte(val),
	}
}

// Take the message struct from the paho mqtt client library for returning
// a test message interface.
type message struct {
	duplicate bool
	qos       byte
	retained  bool
	topic     string
	messageID uint16
	payload   []byte
}

func (m *message) Duplicate() bool {
	return m.duplicate
}

func (m *message) Qos() byte {
	return m.qos
}

func (m *message) Retained() bool {
	return m.retained
}

func (m *message) Topic() string {
	return m.topic
}

func (m *message) MessageID() uint16 {
	return m.messageID
}

func (m *message) Payload() []byte {
	return m.payload
}
