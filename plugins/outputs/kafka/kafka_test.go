package kafka

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/netmonk"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
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

	zooJass, err := filepath.Abs("testdata/zookeeper/zoo_plain_jaas.conf")
	require.NoError(t, err)

	zookeeper := testutil.Container{
		Image:        "wurstmeister/zookeeper",
		ExposedPorts: []string{"2181:2181"},
		BindMounts: map[string]string{
			"/opt/zookeeper_jaas.conf": zooJass,
		},
		Env: map[string]string{
			"SERVER_JVMFLAGS": "-Djava.security.auth.login.config=/opt/zookeeper_jaas.conf",
		},
		Networks:   []string{networkName},
		WaitingFor: wait.ForLog("binding to port"),
		Name:       "telegraf-test-zookeeper",
	}
	err = zookeeper.Start()
	require.NoError(t, err, "failed to start container")
	defer zookeeper.Terminate()

	kafkaJass, err := filepath.Abs("testdata/kafka/kafka_plain_jaas.conf")
	require.NoError(t, err)

	container := testutil.Container{
		Image:        "wurstmeister/kafka:2.12-2.1.1",
		ExposedPorts: []string{"9092:9092"},
		BindMounts: map[string]string{
			"/opt/kafka_plain_jaas.conf": kafkaJass,
		},
		Env: map[string]string{
			"KAFKA_CREATE_TOPICS":                        "Test:1:1",
			"KAFKA_ZOOKEEPER_CONNECT":                    fmt.Sprintf("telegraf-test-zookeeper:%s", zookeeper.Ports["2181"]),
			"KAFKA_SUPER_USER":                           "User:netmonkbroker",
			"KAFKA_LISTENERS":                            "SASL_PLAINTEXT://:9092",
			"KAFKA_ADVERTISED_LISTENERS":                 "SASL_PLAINTEXT://localhost:9092",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":       "SASL_PLAINTEXT:SASL_PLAINTEXT",
			"ALLOW_PLAINTEXT_LISTENER":                   "yes",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":            "true",
			"KAFKA_INTER_BROKER_LISTENER_NAME":           "SASL_PLAINTEXT",
			"KAFKA_SASL_ENABLED_MECHANISMS":              "PLAIN",
			"KAFKA_SASL_MECHANISM_INTER_BROKER_PROTOCOL": "PLAIN",
			"KAFKA_OPTS":                                 "-Djava.security.auth.login.config=/opt/kafka_plain_jaas.conf",
		},
		Name:       "telegraf-test-kafka",
		Networks:   []string{networkName},
		WaitingFor: wait.ForLog("Startup complete"),
	}

	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	brokers := []string{
		fmt.Sprintf("%s:%s", container.Address, container.Ports["9092"]),
	}

	// Prepare httptest endpoint for agent verification
	r := mux.NewRouter()
	r.HandleFunc("/public/controller/server/server-12345/verify", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"client_id" : "agent-12345",
			"message_broker":{
				"type": "kafka",
				"address": ["localhost:9092"]
			},
			"auth":{
				"is_enabled":true,
				"username": "netmonkbroker",
				"password": "netmonkbrokersecret247"
			},
			"sasl":{
				"is_enabled":true,
				"mechanism": "PLAIN"
			},
			"tls":{
				"is_enabled":false,
				"ca":"ca",
				"access":"access",
				"key":"key"
			}
		}`))
	}).Methods("POST")
	httpTestServer := httptest.NewServer(r)
	defer httpTestServer.Close()

	s := &influx.Serializer{}
	require.NoError(t, s.Init())

	k := &Kafka{
		Brokers:      brokers,
		Topic:        "Test",
		Log:          testutil.Logger{},
		serializer:   s,
		producerFunc: sarama.NewSyncProducer,
		Agent: netmonk.Agent{
			NetmonkHost:      httpTestServer.URL,
			NetmonkServerID:  "server-12345",
			NetmonkServerKey: "12345",
		},
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

			s := &influx.Serializer{}
			require.NoError(t, s.Init())
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
