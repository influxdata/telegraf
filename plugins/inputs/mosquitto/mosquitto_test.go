package mosquitto

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

const (
	MosquittoBrokerHost = "192.168.77.77"
	MosquittoBrokerPort = "1883"

	NumClientsConnected    = 2
	NumClientsDisconnected = 3
	NumClientsMaximum      = 999
	NumLoadConnections     = 15
	NumLoadSockets         = 32

	TestInstanceName = "test_instance_name"
)

var measurementTopics = setupMeasurementTopics()

func newTestMQTTConsumer() (*MQTTConsumer, chan mqtt.Message) {
	in := make(chan mqtt.Message, 100)
	n := NewMQTTConsumer()

	// initialize for tests
	n.Servers = []string{MosquittoBrokerHost + ":" + MosquittoBrokerPort}
	n.in = in
	n.done = make(chan struct{})
	n.connected = true
	n.Tags = []string{"mosquitto_instance_name:" + TestInstanceName}

	return n, in
}

// Test that default client has random ID
func TestRandomClientID(t *testing.T) {
	m1 := &MQTTConsumer{
		Servers: []string{MosquittoBrokerHost + ":" + MosquittoBrokerPort}}
	opts, err := m1.createOpts()
	assert.NoError(t, err)

	m2 := &MQTTConsumer{
		Servers: []string{MosquittoBrokerHost + ":" + MosquittoBrokerPort}}
	opts2, err2 := m2.createOpts()
	assert.NoError(t, err2)

	assert.NotEqual(t, opts.ClientID, opts2.ClientID)
}

// Test that default client has random ID
func TestClientID(t *testing.T) {
	m1 := &MQTTConsumer{
		Servers:  []string{MosquittoBrokerHost + ":" + MosquittoBrokerPort},
		ClientID: "telegraf-test",
	}
	opts, err := m1.createOpts()
	assert.NoError(t, err)

	m2 := &MQTTConsumer{
		Servers:  []string{MosquittoBrokerHost + ":" + MosquittoBrokerPort},
		ClientID: "telegraf-test",
	}
	opts2, err2 := m2.createOpts()
	assert.NoError(t, err2)

	assert.Equal(t, "telegraf-test", opts2.ClientID)
	assert.Equal(t, "telegraf-test", opts.ClientID)
}

// Send a message a get the accumulator
func SendMessage(t *testing.T, msg mqtt.Message) *testutil.Accumulator {
	n, in := newTestMQTTConsumer()
	acc := testutil.Accumulator{}
	n.acc = &acc

	defer close(n.done)

	go n.receiver()
	in <- msg
	acc.Wait(1)

	n.Gather(&acc)

	return &acc
}

// Test gathering metrics for clients connected
func TestGatherClientsConnected(t *testing.T) {

	topic := "$SYS/broker/clients/connected"
	mt := measurementTopics[topic]

	clientsConnectedMsg := mqttMsgFromInt(topic, Float64bytes(float64(NumClientsConnected)))
	acc := SendMessage(t, clientsConnectedMsg)

	expectedFields := map[string]interface{}{
		mt.field: Float64bytes(float64(NumClientsConnected)),
	}
	expectedTags := map[string]string{
		"mosquitto_instance_name": TestInstanceName,
		"topic":                   topic,
	}
	acc.AssertContainsTaggedFields(t, mt.measurement, expectedFields, expectedTags)
}

// Test gathering metrics for clients disconnected
func TestGatherClientsDisconnected(t *testing.T) {

	topic := "$SYS/broker/clients/disconnected"
	mt := measurementTopics[topic]
	clientsDisconnectedMsg := mqttMsgFromInt(topic, Float64bytes(float64(NumClientsDisconnected)))

	acc := SendMessage(t, clientsDisconnectedMsg)

	// accJSON, _ := json.Marshal(acc)
	// fmt.Println(string(accJSON))

	expectedFields := map[string]interface{}{
		mt.field: Float64bytes(float64(NumClientsDisconnected)),
	}
	expectedTags := map[string]string{
		"mosquitto_instance_name": TestInstanceName,
		"topic":                   topic,
	}
	acc.AssertContainsTaggedFields(t, mt.measurement, expectedFields, expectedTags)
}

// Test gathering metrics for clients maximum
func TestGatherClientsMaximum(t *testing.T) {

	topic := "$SYS/broker/clients/maximum"
	mt := measurementTopics[topic]
	clientsMaximumMsg := mqttMsgFromInt(topic, Float64bytes(float64(NumClientsMaximum)))

	acc := SendMessage(t, clientsMaximumMsg)

	// accJSON, _ := json.Marshal(acc)
	// fmt.Println(string(accJSON))

	expectedFields := map[string]interface{}{
		mt.field: Float64bytes(float64(NumClientsMaximum)),
	}
	expectedTags := map[string]string{
		"mosquitto_instance_name": TestInstanceName,
		"topic":                   topic,
	}
	acc.AssertContainsTaggedFields(t, mt.measurement, expectedFields, expectedTags)
}

// Test gathering metrics for connections
func TestGatherConnectionsRate(t *testing.T) {

	topic := "$SYS/broker/load/connections/1min"
	mt := measurementTopics[topic]
	connectionsRateMsg := mqttMsgFromInt(topic, Float64bytes(float64(NumLoadConnections)))

	acc := SendMessage(t, connectionsRateMsg)

	// accJSON, _ := json.Marshal(acc)
	// fmt.Println(string(accJSON))

	expectedFields := map[string]interface{}{
		mt.field: Float64bytes(float64(NumLoadConnections)),
	}
	expectedTags := map[string]string{
		"mosquitto_instance_name": TestInstanceName,
		"topic":                   topic,
	}
	acc.AssertContainsTaggedFields(t, mt.measurement, expectedFields, expectedTags)
}

// Test gathering metrics for sockets
func TestGatherSocketsRate(t *testing.T) {

	topic := "$SYS/broker/load/sockets/1min"
	mt := measurementTopics[topic]
	socketsRateMsg := mqttMsgFromInt(topic, Float64bytes(float64(NumLoadSockets)))

	acc := SendMessage(t, socketsRateMsg)

	accJSON, _ := json.Marshal(acc)
	fmt.Println(string(accJSON))

	expectedFields := map[string]interface{}{
		mt.field: Float64bytes(float64(NumLoadSockets)),
	}
	expectedTags := map[string]string{
		"mosquitto_instance_name": TestInstanceName,
		"topic":                   topic,
	}
	acc.AssertContainsTaggedFields(t, mt.measurement, expectedFields, expectedTags)
}

func mqttMsg(topic string, val string) mqtt.Message {
	return &message{
		topic:   topic,
		payload: []byte(val),
	}
}

func mqttMsgFromInt(topic string, val []byte) mqtt.Message {
	return &message{
		topic:   topic,
		payload: val,
	}
}

func Float64bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
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
