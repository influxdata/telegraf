package kafka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

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

	ctx := context.Background()
	networkName := "kafka-test-network"
	net, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			Attachable:     true,
			CheckDuplicate: true,
		},
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, net.Remove(ctx), "terminating network failed")
	}()

	zookeeper := testutil.Container{
		Image:        "wurstmeister/zookeeper",
		ExposedPorts: []string{"2181:2181"},
		Networks:     []string{networkName},
		WaitingFor:   wait.ForLog("binding to port"),
		Name:         "telegraf-test-zookeeper",
	}
	err = zookeeper.Start()
	require.NoError(t, err, "failed to start container")
	defer zookeeper.Terminate()

	container := testutil.Container{
		Image:        "wurstmeister/kafka",
		ExposedPorts: []string{"9092:9092"},
		Env: map[string]string{
			"KAFKA_ADVERTISED_HOST_NAME": "localhost",
			"KAFKA_ADVERTISED_PORT":      "9092",
			"KAFKA_ZOOKEEPER_CONNECT":    fmt.Sprintf("telegraf-test-zookeeper:%s", zookeeper.Ports["2181"]),
			"KAFKA_CREATE_TOPICS":        "Test:1:1",
		},
		Networks:   []string{networkName},
		WaitingFor: wait.ForLog("Log loaded for partition Test-0 with initial high watermark 0"),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	brokers := []string{
		fmt.Sprintf("%s:%s", container.Address, container.Ports["9092"]),
	}

	s := serializers.NewInfluxSerializer()
	k := &Kafka{
		Brokers:      brokers,
		Topic:        "Test",
		Log:          testutil.Logger{},
		serializer:   s,
		producerFunc: sarama.NewSyncProducer,
	}

	// Verify that we can connect to the Kafka broker
	err = k.Init()
	require.NoError(t, err)
	err = k.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the kafka broker
	err = k.Write(testutil.MockMetrics())
	require.NoError(t, err)
	err = k.Close()
	require.NoError(t, err)
}

func TestTopicSuffixes(t *testing.T) {
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
			Log:         testutil.Logger{},
		}

		_, topic := k.GetTopicName(m)
		require.Equal(t, expectedTopic, topic)
	}
}

func TestValidateTopicSuffixMethod(t *testing.T) {
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
			tt.kafka.Log = testutil.Logger{}
			key, err := tt.kafka.routingKey(tt.metric)
			require.NoError(t, err)
			tt.check(t, key)
		})
	}
}

type MockProducer struct {
	sent []*sarama.ProducerMessage
	sarama.SyncProducer
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
			tt.plugin.Log = testutil.Logger{}

			s := serializers.NewInfluxSerializer()
			tt.plugin.SetSerializer(s)

			err := tt.plugin.Connect()
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
