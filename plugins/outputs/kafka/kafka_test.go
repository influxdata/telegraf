package kafka

import (
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	kafkaContainer, err := kafkacontainer.Run(t.Context(), "confluentinc/confluent-local:7.5.0")
	require.NoError(t, err)
	defer kafkaContainer.Terminate(t.Context()) //nolint:errcheck // ignored

	brokers, err := kafkaContainer.Brokers(t.Context())
	require.NoError(t, err)

	// Setup the plugin
	plugin := &Kafka{
		Brokers:      brokers,
		Topic:        "Test",
		Log:          testutil.Logger{},
		producerFunc: sarama.NewSyncProducer,
	}

	// Setup the metric serializer
	s := &influx.Serializer{}
	require.NoError(t, s.Init())
	plugin.SetSerializer(s)

	// Verify that we can connect to the Kafka broker
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Verify that we can successfully write data to the kafka broker
	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}

func TestTopicSuffixes(t *testing.T) {
	topic := "Test"

	m := testutil.TestMetric(1)
	metricTagName := "tag1"
	metricTagValue := m.Tags()[metricTagName]
	metricName := m.Name()

	var tests = []struct {
		suffix   TopicSuffix
		expected string
	}{
		// This ensures empty separator is okay
		{
			TopicSuffix{Method: "measurement"},
			topic + metricName,
		},
		{
			TopicSuffix{Method: "measurement", Separator: "sep"},
			topic + "sep" + metricName,
		},
		{
			TopicSuffix{Method: "tags", Keys: []string{metricTagName}, Separator: "_"},
			topic + "_" + metricTagValue,
		},
		{
			TopicSuffix{Method: "tags", Keys: []string{metricTagName, metricTagName, metricTagName}, Separator: "___"},
			topic + "___" + metricTagValue + "___" + metricTagValue + "___" + metricTagValue,
		},
		{
			TopicSuffix{Method: "tags", Keys: []string{metricTagName, metricTagName, metricTagName}},
			topic + metricTagValue + metricTagValue + metricTagValue,
		},
		{
			// Ensure non-existing tags are ignored
			TopicSuffix{Method: "tags", Keys: []string{"non_existing_tag", "non_existing_tag"}, Separator: "___"},
			topic,
		},
		{
			TopicSuffix{Method: "tags", Keys: []string{metricTagName, "non_existing_tag"}, Separator: "___"},
			topic + "___" + metricTagValue,
		},
		{
			// Ensure backward compatibility
			TopicSuffix{},
			topic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			topicSuffix := tt.suffix
			expectedTopic := tt.expected
			k := &Kafka{
				Topic:       topic,
				TopicSuffix: topicSuffix,
				Log:         testutil.Logger{},
			}

			_, topic := k.getTopicName(m)
			require.Equal(t, expectedTopic, topic)
		})
	}
}

func TestValidTopicSuffixMethod(t *testing.T) {
	for _, method := range []string{"", "measurement", "tags"} {
		name := method
		if method == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			plugin := &Kafka{
				TopicSuffix: TopicSuffix{
					Method: method,
				},
				Log: testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
		})
	}
}

func TestInvalidTopicSuffixMethod(t *testing.T) {
	plugin := &Kafka{
		TopicSuffix: TopicSuffix{
			Method: "invalid_topic_suffix_method",
		},
		Log: testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "unknown topic suffix method provided")
}

func TestRoutingKeyStatic(t *testing.T) {
	plugin := &Kafka{
		RoutingKey: "static",
		Log:        testutil.Logger{},
	}

	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)

	key, err := plugin.routingKey(m)
	require.NoError(t, err)
	require.Equal(t, "static", key)
}

func TestRoutingKeyRandom(t *testing.T) {
	plugin := &Kafka{
		RoutingKey: "random",
		Log:        testutil.Logger{},
	}

	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)

	key, err := plugin.routingKey(m)
	require.NoError(t, err)
	require.Len(t, key, 36)
}

func TestTopicTag(t *testing.T) {
	tests := []struct {
		name            string
		topicTag        string
		excludeTopicTag bool
		expectedTopic   string
		expectedContent string
	}{
		{
			name:            "static topic",
			expectedTopic:   "telegraf",
			expectedContent: "cpu,topic=xyzzy time_idle=42 0\n",
		},
		{
			name:            "topic tag overrides static topic",
			topicTag:        "topic",
			expectedTopic:   "xyzzy",
			expectedContent: "cpu,topic=xyzzy time_idle=42 0\n",
		},
		{
			name:            "missing topic tag falls back to  static topic",
			topicTag:        "non-existant",
			expectedTopic:   "telegraf",
			expectedContent: "cpu,topic=xyzzy time_idle=42 0\n",
		},
		{
			name:            "exclude topic tag removes tag",
			topicTag:        "topic",
			excludeTopicTag: true,
			expectedTopic:   "xyzzy",
			expectedContent: "cpu time_idle=42 0\n",
		},
	}

	// Define an input metric for writing
	input := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"topic": "xyzzy",
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the serializer
			s := &influx.Serializer{}
			require.NoError(t, s.Init())

			// Setup the plugin under test
			plugin := &Kafka{
				Brokers:         []string{"127.0.0.1"},
				Topic:           "telegraf",
				TopicTag:        tt.topicTag,
				ExcludeTopicTag: tt.excludeTopicTag,
				Log:             testutil.Logger{},
				producerFunc:    newMockProducer,
			}
			plugin.SetSerializer(s)
			require.NoError(t, plugin.Init())

			// Connect and write a metric
			require.NoError(t, plugin.Connect())
			require.NoError(t, plugin.Write(input))

			// Check the content that would be sent by the producer
			producer, ok := plugin.producer.(*mockProducer)
			require.True(t, ok, "invalid producer type")

			producer.Lock()
			message := producer.sent[0]
			producer.Unlock()

			require.Equal(t, tt.expectedTopic, message.Topic)
			encoded, err := message.Value.Encode()
			require.NoError(t, err)
			require.Equal(t, tt.expectedContent, string(encoded))
		})
	}
}

type mockProducer struct {
	sent []*sarama.ProducerMessage
	sarama.SyncProducer
	sync.Mutex
}

func newMockProducer(_ []string, _ *sarama.Config) (sarama.SyncProducer, error) {
	return &mockProducer{}, nil
}

func (p *mockProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	p.Lock()
	defer p.Unlock()
	p.sent = append(p.sent, msg)
	return 0, 0, nil
}

func (p *mockProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	p.Lock()
	defer p.Unlock()
	p.sent = append(p.sent, msgs...)
	return nil
}

func (*mockProducer) Close() error {
	return nil
}
