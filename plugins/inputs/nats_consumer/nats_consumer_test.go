package natsconsumer

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/nats-io/nats"
)

const (
	testMsg         = "cpu_load_short,host=server01 value=23422.0 1422568543702900257"
	testMsgGraphite = "cpu.load.short.graphite 23422 1454780029"
	testMsgJSON     = "{\"a\": 5, \"b\": {\"c\": 6}}\n"
	invalidMsg      = "cpu_load_short,host=server01 1422568543702900257"
	metricBuffer    = 5
)

func newTestNatsConsumer() (*natsConsumer, chan *nats.Msg) {
	in := make(chan *nats.Msg, metricBuffer)
	n := &natsConsumer{
		QueueGroup:   "test",
		Subjects:     []string{"telegraf"},
		Servers:      []string{"nats://localhost:4222"},
		Secure:       false,
		MetricBuffer: metricBuffer,
		in:           in,
		errs:         make(chan error, metricBuffer),
		done:         make(chan struct{}),
		metricC:      make(chan telegraf.Metric, metricBuffer),
	}
	return n, in
}

// Test that the parser parses NATS messages into metrics
func TestRunParser(t *testing.T) {
	n, in := newTestNatsConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- natsMsg(testMsg)
	time.Sleep(time.Millisecond)

	if a := len(n.metricC); a != 1 {
		t.Errorf("got %v, expected %v", a, 1)
	}
}

// Test that the parser ignores invalid messages
func TestRunParserInvalidMsg(t *testing.T) {
	n, in := newTestNatsConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- natsMsg(invalidMsg)
	time.Sleep(time.Millisecond)

	if a := len(n.metricC); a != 0 {
		t.Errorf("got %v, expected %v", a, 0)
	}
}

// Test that metrics are dropped when we hit the buffer limit
func TestRunParserRespectsBuffer(t *testing.T) {
	n, in := newTestNatsConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	for i := 0; i < metricBuffer+1; i++ {
		in <- natsMsg(testMsg)
	}
	time.Sleep(time.Millisecond)

	if a := len(n.metricC); a != metricBuffer {
		t.Errorf("got %v, expected %v", a, metricBuffer)
	}
}

// Test that the parser parses line format messages into metrics
func TestRunParserAndGather(t *testing.T) {
	n, in := newTestNatsConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewInfluxParser()
	go n.receiver()
	in <- natsMsg(testMsg)
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
	n, in := newTestNatsConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewGraphiteParser("_", []string{}, nil)
	go n.receiver()
	in <- natsMsg(testMsgGraphite)
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
	n, in := newTestNatsConsumer()
	defer close(n.done)

	n.parser, _ = parsers.NewJSONParser("nats_json_test", []string{}, nil)
	go n.receiver()
	in <- natsMsg(testMsgJSON)
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

func natsMsg(val string) *nats.Msg {
	return &nats.Msg{
		Subject: "telegraf",
		Data:    []byte(val),
	}
}
