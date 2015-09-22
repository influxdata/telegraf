package kafka_consumer

import (
	"strings"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/koksan83/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMsg = "cpu_load_short,direction=in,host=server01,region=us-west value=23422.0 1422568543702900257"

func TestReadFromKafkaBatchesMsgsOnBatchSize(t *testing.T) {
	halt := make(chan bool, 1)
	metricChan := make(chan []byte, 1)
	kafkaChan := make(chan *sarama.ConsumerMessage, 10)
	for i := 0; i < 10; i++ {
		kafkaChan <- saramaMsg(testMsg)
	}

	expectedBatch := strings.Repeat(testMsg+"\n", 9) + testMsg
	readFromKafka(kafkaChan, metricChan, 10, func(msg *sarama.ConsumerMessage) error {
		batch := <-metricChan
		assert.Equal(t, expectedBatch, string(batch))

		halt <- true

		return nil
	}, halt)
}

func TestReadFromKafkaBatchesMsgsOnTimeout(t *testing.T) {
	halt := make(chan bool, 1)
	metricChan := make(chan []byte, 1)
	kafkaChan := make(chan *sarama.ConsumerMessage, 10)
	for i := 0; i < 3; i++ {
		kafkaChan <- saramaMsg(testMsg)
	}

	expectedBatch := strings.Repeat(testMsg+"\n", 2) + testMsg
	readFromKafka(kafkaChan, metricChan, 10, func(msg *sarama.ConsumerMessage) error {
		batch := <-metricChan
		assert.Equal(t, expectedBatch, string(batch))

		halt <- true

		return nil
	}, halt)
}

func TestEmitMetricsSendMetricsToAcc(t *testing.T) {
	k := &Kafka{}
	var acc testutil.Accumulator
	testChan := make(chan []byte, 1)
	testChan <- []byte(testMsg)

	err := emitMetrics(k, &acc, testChan)
	require.NoError(t, err)

	assert.Equal(t, 1, len(acc.Points), "there should be a single point")

	point := acc.Points[0]
	assert.Equal(t, "cpu_load_short", point.Measurement)
	assert.Equal(t, map[string]interface{}{"value": 23422.0}, point.Values)
	assert.Equal(t, map[string]string{
		"host":      "server01",
		"direction": "in",
		"region":    "us-west",
	}, point.Tags)

	assert.Equal(t, time.Unix(0, 1422568543702900257), point.Time)
}

func TestEmitMetricsTimesOut(t *testing.T) {
	k := &Kafka{}
	var acc testutil.Accumulator
	testChan := make(chan []byte)

	err := emitMetrics(k, &acc, testChan)
	require.NoError(t, err)

	assert.Equal(t, 0, len(acc.Points), "there should not be a any points")
}

func saramaMsg(val string) *sarama.ConsumerMessage {
	return &sarama.ConsumerMessage{
		Key:       nil,
		Value:     []byte(val),
		Offset:    0,
		Partition: 0,
	}
}
