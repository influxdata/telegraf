package mqtt_consumer

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

const (
	testMsg         = "cpu_load_short,host=server01 value=23422.0 1422568543702900257"
	testMsgGraphite = "cpu.load.short.graphite 23422 1454780029"
	testMsgJSON     = "{\"a\": 5, \"b\": {\"c\": 6}}\n"
	invalidMsg      = "cpu_load_short,host=server01 1422568543702900257"
	metricBuffer    = 5
)

func newTestMQTTConsumer() (*MQTTConsumer, chan mqtt.Message) {
	in := make(chan mqtt.Message, metricBuffer)
	n := &MQTTConsumer{
		Topics:       []string{"telegraf"},
		Servers:      []string{"localhost:1883"},
		MetricBuffer: metricBuffer,
		in:           in,
		done:         make(chan struct{}),
		metricC:      make(chan telegraf.Metric, metricBuffer),
		topicC:       make(chan string, metricBuffer),
	}
	return n, in
}

// Test that the parser parses NATS messages into metrics
func TestRunParser(t *testing.T) {
	n, in := newTestMQTTConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- mqttMsg(testMsg)
	time.Sleep(time.Millisecond)

	if a := len(n.metricC); a != 1 {
		t.Errorf("got %v, expected %v", a, 1)
	}
}

// Test that the parser ignores invalid messages
func TestRunParserInvalidMsg(t *testing.T) {
	n, in := newTestMQTTConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- mqttMsg(invalidMsg)
	time.Sleep(time.Millisecond)

	if a := len(n.metricC); a != 0 {
		t.Errorf("got %v, expected %v", a, 0)
	}
}

// Test that metrics are dropped when we hit the buffer limit
func TestRunParserRespectsBuffer(t *testing.T) {
	n, in := newTestMQTTConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	for i := 0; i < metricBuffer+1; i++ {
		in <- mqttMsg(testMsg)
	}
	time.Sleep(time.Millisecond)

	if a := len(n.metricC); a != metricBuffer {
		t.Errorf("got %v, expected %v", a, metricBuffer)
	}
}

// Test that the parser parses line format messages into metrics
func TestRunParserAndGather(t *testing.T) {
	n, in := newTestMQTTConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- mqttMsg(testMsg)
	time.Sleep(time.Millisecond)

	acc := testutil.Accumulator{}
	n.Gather(&acc)

	if a := len(acc.Metrics); a != 1 {
		t.Errorf("got %v, expected %v", a, 1)
	}
	acc.AssertContainsFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(23422)})
}

// Test that the parser parses graphite format messages into metrics
func TestRunParserAndGatherGraphite(t *testing.T) {
	n, in := newTestMQTTConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewGraphiteParser("_", []string{}, nil)
	go n.receiver()
	in <- mqttMsg(testMsgGraphite)
	time.Sleep(time.Millisecond)

	acc := testutil.Accumulator{}
	n.Gather(&acc)

	if a := len(acc.Metrics); a != 1 {
		t.Errorf("got %v, expected %v", a, 1)
	}
	acc.AssertContainsFields(t, "cpu_load_short_graphite",
		map[string]interface{}{"value": float64(23422)})
}

// Test that the parser parses json format messages into metrics
func TestRunParserAndGatherJSON(t *testing.T) {
	n, in := newTestMQTTConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewJSONParser("nats_json_test", []string{}, nil)
	go n.receiver()
	in <- mqttMsg(testMsgJSON)
	time.Sleep(time.Millisecond)

	acc := testutil.Accumulator{}
	n.Gather(&acc)

	if a := len(acc.Metrics); a != 1 {
		t.Errorf("got %v, expected %v", a, 1)
	}
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
