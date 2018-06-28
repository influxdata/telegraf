package mqtt_consumer

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"

	"github.com/eclipse/paho.mqtt.golang"
)

const (
	testMsg         = "cpu_load_short,host=server01 value=23422.0 1422568543702900257\n"
	testMsgNeg      = "cpu_load_short,host=server01 value=-23422.0 1422568543702900257\n"
	testMsgGraphite = "cpu.load.short.graphite 23422 1454780029"
	testMsgJSON     = "{\"a\": 5, \"b\": {\"c\": 6}}\n"
	invalidMsg      = "cpu_load_short,host=server01 1422568543702900257\n"
)

func newTestMQTTConsumer() (*MQTTConsumer, chan mqtt.Message) {
	in := make(chan mqtt.Message, 100)
	n := &MQTTConsumer{
		Topics:    []string{"telegraf"},
		Servers:   []string{"localhost:1883"},
		in:        in,
		done:      make(chan struct{}),
		connected: true,
	}

	return n, in
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

func TestRunParser(t *testing.T) {
	n, in := newTestMQTTConsumer()
	acc := testutil.Accumulator{}
	n.acc = &acc
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- mqttMsg(testMsgNeg)
	acc.Wait(1)

	if a := acc.NFields(); a != 1 {
		t.Errorf("got %v, expected %v", a, 1)
	}
}

func TestRunParserNegativeNumber(t *testing.T) {
	n, in := newTestMQTTConsumer()
	acc := testutil.Accumulator{}
	n.acc = &acc
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- mqttMsg(testMsg)
	acc.Wait(1)

	if a := acc.NFields(); a != 1 {
		t.Errorf("got %v, expected %v", a, 1)
	}
}

// Test that the parser ignores invalid messages
func TestRunParserInvalidMsg(t *testing.T) {
	n, in := newTestMQTTConsumer()
	acc := testutil.Accumulator{}
	n.acc = &acc
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- mqttMsg(invalidMsg)
	acc.WaitError(1)

	if a := acc.NFields(); a != 0 {
		t.Errorf("got %v, expected %v", a, 0)
	}
	assert.Contains(t, acc.Errors[0].Error(), "MQTT Parse Error")
}

// Test that the parser parses line format messages into metrics
func TestRunParserAndGather(t *testing.T) {
	n, in := newTestMQTTConsumer()
	acc := testutil.Accumulator{}
	n.acc = &acc

	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- mqttMsg(testMsg)
	acc.Wait(1)

	n.Gather(&acc)

	acc.AssertContainsFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(23422)})
}

// Test that the parser parses graphite format messages into metrics
func TestRunParserAndGatherGraphite(t *testing.T) {
	n, in := newTestMQTTConsumer()
	acc := testutil.Accumulator{}
	n.acc = &acc
	defer close(n.done)

	n.parser, _ = parsers.NewGraphiteParser("_", []string{}, nil)
	go n.receiver()
	in <- mqttMsg(testMsgGraphite)

	n.Gather(&acc)
	acc.Wait(1)

	acc.AssertContainsFields(t, "cpu_load_short_graphite",
		map[string]interface{}{"value": float64(23422)})
}

// Test that the parser parses json format messages into metrics
func TestRunParserAndGatherJSON(t *testing.T) {
	n, in := newTestMQTTConsumer()
	acc := testutil.Accumulator{}
	n.acc = &acc
	defer close(n.done)

	n.parser, _ = parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "nats_json_test",
	})
	go n.receiver()
	in <- mqttMsg(testMsgJSON)

	n.Gather(&acc)

	acc.Wait(1)

	acc.AssertContainsFields(t, "nats_json_test",
		map[string]interface{}{
			"a":   float64(5),
			"b_c": float64(6),
		})
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
