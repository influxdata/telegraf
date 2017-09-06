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

	topic := "Test_"

	metric := testutil.TestMetric(1)
	metricTagName := "tag1"
	metricTagValue := metric.Tags()[metricTagName]
	metricName := metric.Name()

	var testcases = []topicSuffixTestpair{
		{TopicSuffix{Method: "measurement"},
			topic + metricName},
		{TopicSuffix{Method: "tags", Keys: []string{metricTagName}},
			topic + metricTagValue},
		{TopicSuffix{Method: "tags", Keys: []string{metricTagName, metricTagName, metricTagName}, Separator: "___"},
			topic + metricTagValue + "___" + metricTagValue + "___" + metricTagValue},
		// This ensures backward compatibility
		{TopicSuffix{},
			topic},
	}

	for _, testcase := range testcases {
		topicSuffix := testcase.topicSuffix
		expectedTopic := testcase.expectedTopic
		k := &Kafka{
			Topic:       topic,
			TopicSuffix: topicSuffix,
		}

		topic := k.GetTopicName(metric)
		require.Equal(t, expectedTopic, topic)
	}
}

func TestValidateTopicSuffixMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	err := ValidateTopicSuffixMethod("invalid_topic_suffix_method")
	require.Error(t, err, "Topic suffix method used should be invalid.")

	for _, method := range ValidTopicSuffixMethods {
		err := ValidateTopicSuffixMethod(method)
		require.NoError(t, err, "Topic suffix method used should be valid.")
	}
}
