package kafka_consumer

import (
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/koksan83/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadsMetricsFromKafka(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	var zkPeers, brokerPeers []string

	zkPeers = []string{testutil.GetLocalHost() + ":2181"}
	brokerPeers = []string{testutil.GetLocalHost() + ":9092"}

	k := &Kafka{
		ConsumerGroupName: "telegraf_test_consumers",
		Topic:             fmt.Sprintf("telegraf_test_topic_%d", time.Now().Unix()),
		ZookeeperPeers:    zkPeers,
	}

	msg := "cpu_load_short,direction=in,host=server01,region=us-west value=23422.0 1422568543702900257"
	producer, err := sarama.NewSyncProducer(brokerPeers, nil)
	require.NoError(t, err)
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{Topic: k.Topic, Value: sarama.StringEncoder(msg)})
	producer.Close()

	var acc testutil.Accumulator

	// Sanity check
	assert.Equal(t, 0, len(acc.Points), "there should not be any points")

	err = k.Gather(&acc)
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
