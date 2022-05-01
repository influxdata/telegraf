package kafka

import (
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

type topicSuffixTestpair struct {
	topicSuffix   TopicSuffix
	expectedTopic string
}

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	brokers := []string{testutil.GetLocalHost() + ":9092"}
	s, _ := serializers.NewInfluxSerializer()
	k := &Kafka{
		Brokers:      brokers,
		Topic:        "Test",
		serializer:   s,
		producerFunc: sarama.NewSyncProducer,
	}

	// Verify that we can connect to the Kafka broker
	err := k.Init()
	require.NoError(t, err)
	err = k.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the kafka broker
	err = k.Write(testutil.MockMetrics())
	require.NoError(t, err)
	k.Close()
}

func TestTopicSuffixesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	topic := "Test"

	m := testutil.TestMetric(1)
	metricTagName := "tag1"
	metricTagValue := m.Tags()[metricTagName]
	metricName := m.Name()

	var testcases = []topicSuffixTestpair{
		// This ensures empty separator is okay
		{TopicSuffix{Method: "measurement"},
			topic + metricName},
		{TopicSuffix{Method: "measurement", Separator: "sep"},
			topic + "sep" + metricName},
		{TopicSuffix{Method: "tags", Keys: []string{metricTagName}, Separator: "_"},
			topic + "_" + metricTagValue},
		{TopicSuffix{Method: "tags", Keys: []string{metricTagName, metricTagName, metricTagName}, Separator: "___"},
			topic + "___" + metricTagValue + "___" + metricTagValue + "___" + metricTagValue},
		{TopicSuffix{Method: "tags", Keys: []string{metricTagName, metricTagName, metricTagName}},
			topic + metricTagValue + metricTagValue + metricTagValue},
		// This ensures non-existing tags are ignored
		{TopicSuffix{Method: "tags", Keys: []string{"non_existing_tag", "non_existing_tag"}, Separator: "___"},
			topic},
		{TopicSuffix{Method: "tags", Keys: []string{metricTagName, "non_existing_tag"}, Separator: "___"},
			topic + "___" + metricTagValue},
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

		_, topic := k.GetTopicName(m)
		require.Equal(t, expectedTopic, topic)
	}
}

func TestValidateTopicSuffixMethodIntegration(t *testing.T) {
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

func TestRoutingKey(t *testing.T) {
	tests := []struct {
		name   string
		kafka  *Kafka
		metric telegraf.Metric
		check  func(t *testing.T, routingKey string)
	}{
		{
			name: "static routing key",
			kafka: &Kafka{
				RoutingKey: "static",
			},
			metric: func() telegraf.Metric {
				m := metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				)
				return m
			}(),
			check: func(t *testing.T, routingKey string) {
				require.Equal(t, "static", routingKey)
			},
		},
		{
			name: "random routing key",
			kafka: &Kafka{
				RoutingKey: "random",
			},
			metric: func() telegraf.Metric {
				m := metric.New(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				)
				return m
			}(),
			check: func(t *testing.T, routingKey string) {
				require.Equal(t, 36, len(routingKey))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := tt.kafka.routingKey(tt.metric)
			require.NoError(t, err)
			tt.check(t, key)
		})
	}
}

type MockProducer struct {
	sent []*sarama.ProducerMessage
}

func (p *MockProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	p.sent = append(p.sent, msg)
	return 0, 0, nil
}

func (p *MockProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	p.sent = append(p.sent, msgs...)
	return nil
}

func (p *MockProducer) Close() error {
	return nil
}

func NewMockProducer(_ []string, _ *sarama.Config) (sarama.SyncProducer, error) {
	return &MockProducer{}, nil
}

func TestTopicTag(t *testing.T) {
	tests := []struct {
		name   string
		plugin *Kafka
		input  []telegraf.Metric
		topic  string
		value  string
	}{
		{
			name: "static topic",
			plugin: &Kafka{
				Brokers:      []string{"127.0.0.1"},
				Topic:        "telegraf",
				producerFunc: NewMockProducer,
			},
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			topic: "telegraf",
			value: "cpu time_idle=42 0\n",
		},
		{
			name: "topic tag overrides static topic",
			plugin: &Kafka{
				Brokers:      []string{"127.0.0.1"},
				Topic:        "telegraf",
				TopicTag:     "topic",
				producerFunc: NewMockProducer,
			},
			input: []telegraf.Metric{
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
			},
			topic: "xyzzy",
			value: "cpu,topic=xyzzy time_idle=42 0\n",
		},
		{
			name: "missing topic tag falls back to  static topic",
			plugin: &Kafka{
				Brokers:      []string{"127.0.0.1"},
				Topic:        "telegraf",
				TopicTag:     "topic",
				producerFunc: NewMockProducer,
			},
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
			topic: "telegraf",
			value: "cpu time_idle=42 0\n",
		},
		{
			name: "exclude topic tag removes tag",
			plugin: &Kafka{
				Brokers:         []string{"127.0.0.1"},
				Topic:           "telegraf",
				TopicTag:        "topic",
				ExcludeTopicTag: true,
				producerFunc:    NewMockProducer,
			},
			input: []telegraf.Metric{
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
			},
			topic: "xyzzy",
			value: "cpu time_idle=42 0\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := serializers.NewInfluxSerializer()
			require.NoError(t, err)
			tt.plugin.SetSerializer(s)

			err = tt.plugin.Connect()
			require.NoError(t, err)

			producer := &MockProducer{}
			tt.plugin.producer = producer

			err = tt.plugin.Write(tt.input)
			require.NoError(t, err)

			require.Equal(t, tt.topic, producer.sent[0].Topic)

			encoded, err := producer.sent[0].Value.Encode()
			require.NoError(t, err)
			require.Equal(t, tt.value, string(encoded))
		})
	}
}
