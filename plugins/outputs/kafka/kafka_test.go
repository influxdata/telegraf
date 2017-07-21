package kafka

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type topicSuffixTestpair struct {
	topicSuffix   TopicSuffix
	expectedTopic string
}

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	brokers := []string{testutil.GetLocalHost() + ":9092"}
	s, _ := serializers.NewInfluxSerializer()
	k := &Kafka{
		Brokers:    brokers,
		Topic:      "Test",
		serializer: s,
	}

	// Verify that we can connect to the Kafka broker
	err := k.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the kafka broker
	err = k.Write(testutil.MockMetrics())
	require.NoError(t, err)
	k.Close()
}

func TestTopicSuffixes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	brokers := []string{testutil.GetLocalHost() + ":9092"}
	s, _ := serializers.NewInfluxSerializer()

	topic := "Test_"

	metric := testutil.TestMetric(1)
	metricTagName := "tag1"
	metricTagValue := metric.Tags()[metricTagName]
	metricName := metric.Name()

	var testcases = []topicSuffixTestpair{
		{TopicSuffix{Method: "measurement"},
			topic + metricName},
		{TopicSuffix{Method: "tag", Key: metricTagName},
			topic + metricTagValue},
		{TopicSuffix{Method: "tags", Keys: []string{metricTagName, metricTagName, metricTagName}, KeySeparator: "___"},
			topic + metricTagValue + "___" + metricTagValue + "___" + metricTagValue},
		// This ensures backward compatibility
		{TopicSuffix{},
			topic},
	}

	for _, testcase := range testcases {
		topicSuffix := testcase.topicSuffix
		expectedTopic := testcase.expectedTopic
		k := &Kafka{
			Brokers:     brokers,
			Topic:       topic,
			serializer:  s,
			TopicSuffix: topicSuffix,
		}

		err := k.Connect()
		require.NoError(t, err)

		topic := k.GetTopicName(metric)
		require.Equal(t, expectedTopic, topic)
		k.Close()
	}
}
