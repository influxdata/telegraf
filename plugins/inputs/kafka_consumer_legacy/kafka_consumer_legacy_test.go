package kafka_consumer_legacy

import (
	"strings"
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
)

const (
	testMsg         = "cpu_load_short,host=server01 value=23422.0 1422568543702900257\n"
	testMsgGraphite = "cpu.load.short.graphite 23422 1454780029"
	testMsgJSON     = "{\"a\": 5, \"b\": {\"c\": 6}}\n"
	invalidMsg      = "cpu_load_short,host=server01 1422568543702900257\n"
)

func newTestKafka() (*Kafka, chan *sarama.ConsumerMessage) {
	in := make(chan *sarama.ConsumerMessage, 1000)
	k := Kafka{
		ConsumerGroup:   "test",
		Topics:          []string{"telegraf"},
		ZookeeperPeers:  []string{"localhost:2181"},
		Offset:          "oldest",
		in:              in,
		doNotCommitMsgs: true,
		errs:            make(chan error, 1000),
		done:            make(chan struct{}),
	}
	return &k, in
}

// Test that the parser parses kafka messages into points
func TestRunParser(t *testing.T) {
	k, in := newTestKafka()
	acc := testutil.Accumulator{}
	k.acc = &acc
	defer close(k.done)

	k.parser, _ = parsers.NewInfluxParser()
	go k.receiver()
	in <- saramaMsg(testMsg)
	acc.Wait(1)

	assert.Equal(t, acc.NFields(), 1)
}

// Test that the parser ignores invalid messages
func TestRunParserInvalidMsg(t *testing.T) {
	k, in := newTestKafka()
	acc := testutil.Accumulator{}
	k.acc = &acc
	defer close(k.done)

	k.parser, _ = parsers.NewInfluxParser()
	go k.receiver()
	in <- saramaMsg(invalidMsg)
	acc.WaitError(1)

	assert.Equal(t, acc.NFields(), 0)
}

// Test that overlong messages are dropped
func TestDropOverlongMsg(t *testing.T) {
	const maxMessageLen = 64 * 1024
	k, in := newTestKafka()
	k.MaxMessageLen = maxMessageLen
	acc := testutil.Accumulator{}
	k.acc = &acc
	defer close(k.done)
	overlongMsg := strings.Repeat("v", maxMessageLen+1)

	go k.receiver()
	in <- saramaMsg(overlongMsg)
	acc.WaitError(1)

	assert.Equal(t, acc.NFields(), 0)
}

// Test that the parser parses kafka messages into points
func TestRunParserAndGather(t *testing.T) {
	k, in := newTestKafka()
	acc := testutil.Accumulator{}
	k.acc = &acc
	defer close(k.done)

	k.parser, _ = parsers.NewInfluxParser()
	go k.receiver()
	in <- saramaMsg(testMsg)
	acc.Wait(1)

	acc.GatherError(k.Gather)

	assert.Equal(t, acc.NFields(), 1)
	acc.AssertContainsFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(23422)})
}

// Test that the parser parses kafka messages into points
func TestRunParserAndGatherGraphite(t *testing.T) {
	k, in := newTestKafka()
	acc := testutil.Accumulator{}
	k.acc = &acc
	defer close(k.done)

	k.parser, _ = parsers.NewGraphiteParser("_", []string{}, nil)
	go k.receiver()
	in <- saramaMsg(testMsgGraphite)
	acc.Wait(1)

	acc.GatherError(k.Gather)

	assert.Equal(t, acc.NFields(), 1)
	acc.AssertContainsFields(t, "cpu_load_short_graphite",
		map[string]interface{}{"value": float64(23422)})
}

// Test that the parser parses kafka messages into points
func TestRunParserAndGatherJSON(t *testing.T) {
	k, in := newTestKafka()
	acc := testutil.Accumulator{}
	k.acc = &acc
	defer close(k.done)

	k.parser, _ = parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "kafka_json_test",
	})
	go k.receiver()
	in <- saramaMsg(testMsgJSON)
	acc.Wait(1)

	acc.GatherError(k.Gather)

	assert.Equal(t, acc.NFields(), 2)
	acc.AssertContainsFields(t, "kafka_json_test",
		map[string]interface{}{
			"a":   float64(5),
			"b_c": float64(6),
		})
}

func saramaMsg(val string) *sarama.ConsumerMessage {
	return &sarama.ConsumerMessage{
		Key:       nil,
		Value:     []byte(val),
		Offset:    0,
		Partition: 0,
	}
}
