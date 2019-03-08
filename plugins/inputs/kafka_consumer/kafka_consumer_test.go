package kafka_consumer

import (
	"context"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

const (
	testMsg         = "cpu_load_short,host=server01 value=23422.0 1422568543702900257\n"
	testMsgGraphite = "cpu.load.short.graphite 23422 1454780029"
	testMsgJSON     = "{\"a\": 5, \"b\": {\"c\": 6}}\n"
	invalidMsg      = "cpu_load_short,host=server01 1422568543702900257\n"
)

type TestConsumer struct {
	errors   chan error
	messages chan *sarama.ConsumerMessage
}

func (c *TestConsumer) Errors() <-chan error {
	return c.errors
}

func (c *TestConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return c.messages
}

func (c *TestConsumer) MarkOffset(msg *sarama.ConsumerMessage, metadata string) {
}

func (c *TestConsumer) Close() error {
	return nil
}

func (c *TestConsumer) Inject(msg *sarama.ConsumerMessage) {
	c.messages <- msg
}

func newTestKafka() (*Kafka, *TestConsumer) {
	consumer := &TestConsumer{
		errors:   make(chan error),
		messages: make(chan *sarama.ConsumerMessage, 1000),
	}
	k := Kafka{
		cluster:                consumer,
		ConsumerGroup:          "test",
		Topics:                 []string{"telegraf"},
		Brokers:                []string{"localhost:9092"},
		Offset:                 "oldest",
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		doNotCommitMsgs:        true,
		messages:               make(map[telegraf.TrackingID]*sarama.ConsumerMessage),
	}
	return &k, consumer
}

func newTestKafkaWithTopicTag() (*Kafka, *TestConsumer) {
	consumer := &TestConsumer{
		errors:   make(chan error),
		messages: make(chan *sarama.ConsumerMessage, 1000),
	}
	k := Kafka{
		cluster:                consumer,
		ConsumerGroup:          "test",
		Topics:                 []string{"telegraf"},
		Brokers:                []string{"localhost:9092"},
		Offset:                 "oldest",
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		doNotCommitMsgs:        true,
		messages:               make(map[telegraf.TrackingID]*sarama.ConsumerMessage),
		TopicTag:               "topic",
	}
	return &k, consumer
}

// Test that the parser parses kafka messages into points
func TestRunParser(t *testing.T) {
	k, consumer := newTestKafka()
	acc := testutil.Accumulator{}
	ctx := context.Background()

	k.parser, _ = parsers.NewInfluxParser()
	go k.receiver(ctx, &acc)
	consumer.Inject(saramaMsg(testMsg))
	acc.Wait(1)

	assert.Equal(t, acc.NFields(), 1)
}

// Test that the parser parses kafka messages into points
// and adds the topic tag
func TestRunParserWithTopic(t *testing.T) {
	k, consumer := newTestKafkaWithTopicTag()
	acc := testutil.Accumulator{}
	ctx := context.Background()

	k.parser, _ = parsers.NewInfluxParser()
	go k.receiver(ctx, &acc)
	consumer.Inject(saramaMsgWithTopic(testMsg, "test_topic"))
	acc.Wait(1)

	assert.Equal(t, acc.NFields(), 1)
	assert.True(t, acc.HasTag("cpu_load_short", "topic"))
}

// Test that the parser ignores invalid messages
func TestRunParserInvalidMsg(t *testing.T) {
	k, consumer := newTestKafka()
	acc := testutil.Accumulator{}
	ctx := context.Background()

	k.parser, _ = parsers.NewInfluxParser()
	go k.receiver(ctx, &acc)
	consumer.Inject(saramaMsg(invalidMsg))
	acc.WaitError(1)

	assert.Equal(t, acc.NFields(), 0)
}

// Test that overlong messages are dropped
func TestDropOverlongMsg(t *testing.T) {
	const maxMessageLen = 64 * 1024
	k, consumer := newTestKafka()
	k.MaxMessageLen = maxMessageLen
	acc := testutil.Accumulator{}
	ctx := context.Background()
	overlongMsg := strings.Repeat("v", maxMessageLen+1)

	go k.receiver(ctx, &acc)
	consumer.Inject(saramaMsg(overlongMsg))
	acc.WaitError(1)

	assert.Equal(t, acc.NFields(), 0)
}

// Test that the parser parses kafka messages into points
func TestRunParserAndGather(t *testing.T) {
	k, consumer := newTestKafka()
	acc := testutil.Accumulator{}
	ctx := context.Background()

	k.parser, _ = parsers.NewInfluxParser()
	go k.receiver(ctx, &acc)
	consumer.Inject(saramaMsg(testMsg))
	acc.Wait(1)

	acc.GatherError(k.Gather)

	assert.Equal(t, acc.NFields(), 1)
	acc.AssertContainsFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(23422)})
}

// Test that the parser parses kafka messages into points
func TestRunParserAndGatherGraphite(t *testing.T) {
	k, consumer := newTestKafka()
	acc := testutil.Accumulator{}
	ctx := context.Background()

	k.parser, _ = parsers.NewGraphiteParser("_", []string{}, nil)
	go k.receiver(ctx, &acc)
	consumer.Inject(saramaMsg(testMsgGraphite))
	acc.Wait(1)

	acc.GatherError(k.Gather)

	assert.Equal(t, acc.NFields(), 1)
	acc.AssertContainsFields(t, "cpu_load_short_graphite",
		map[string]interface{}{"value": float64(23422)})
}

// Test that the parser parses kafka messages into points
func TestRunParserAndGatherJSON(t *testing.T) {
	k, consumer := newTestKafka()
	acc := testutil.Accumulator{}
	ctx := context.Background()

	k.parser, _ = parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "kafka_json_test",
	})
	go k.receiver(ctx, &acc)
	consumer.Inject(saramaMsg(testMsgJSON))
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

func saramaMsgWithTopic(val string, topic string) *sarama.ConsumerMessage {
	return &sarama.ConsumerMessage{
		Key:       nil,
		Value:     []byte(val),
		Offset:    0,
		Partition: 0,
		Topic:     topic,
	}
}
