package kafka_consumer

import (
	"testing"
	"time"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/telegraf/testutil"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
)

const (
	testMsg     = "cpu_load_short,host=server01 value=23422.0 1422568543702900257"
	invalidMsg  = "cpu_load_short,host=server01 1422568543702900257"
	pointBuffer = 5
)

func NewTestKafka() (*Kafka, chan *sarama.ConsumerMessage) {
	in := make(chan *sarama.ConsumerMessage, pointBuffer)
	k := Kafka{
		ConsumerGroup:   "test",
		Topics:          []string{"telegraf"},
		ZookeeperPeers:  []string{"localhost:2181"},
		PointBuffer:     pointBuffer,
		Offset:          "oldest",
		in:              in,
		doNotCommitMsgs: true,
		errs:            make(chan *sarama.ConsumerError, pointBuffer),
		done:            make(chan struct{}),
		pointChan:       make(chan models.Point, pointBuffer),
	}
	return &k, in
}

// Test that the parser parses kafka messages into points
func TestRunParser(t *testing.T) {
	k, in := NewTestKafka()
	defer close(k.done)

	go k.parser()
	in <- saramaMsg(testMsg)
	time.Sleep(time.Millisecond)

	assert.Equal(t, len(k.pointChan), 1)
}

// Test that the parser ignores invalid messages
func TestRunParserInvalidMsg(t *testing.T) {
	k, in := NewTestKafka()
	defer close(k.done)

	go k.parser()
	in <- saramaMsg(invalidMsg)
	time.Sleep(time.Millisecond)

	assert.Equal(t, len(k.pointChan), 0)
}

// Test that points are dropped when we hit the buffer limit
func TestRunParserRespectsBuffer(t *testing.T) {
	k, in := NewTestKafka()
	defer close(k.done)

	go k.parser()
	for i := 0; i < pointBuffer+1; i++ {
		in <- saramaMsg(testMsg)
	}
	time.Sleep(time.Millisecond)

	assert.Equal(t, len(k.pointChan), 5)
}

// Test that the parser parses kafka messages into points
func TestRunParserAndGather(t *testing.T) {
	k, in := NewTestKafka()
	defer close(k.done)

	go k.parser()
	in <- saramaMsg(testMsg)
	time.Sleep(time.Millisecond)

	acc := testutil.Accumulator{}
	k.Gather(&acc)

	assert.Equal(t, len(acc.Points), 1)
	acc.AssertContainsFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(23422)})
}

func saramaMsg(val string) *sarama.ConsumerMessage {
	return &sarama.ConsumerMessage{
		Key:       nil,
		Value:     []byte(val),
		Offset:    0,
		Partition: 0,
	}
}
