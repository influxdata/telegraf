package kafka

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
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
		map[string]any{
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
		map[string]any{
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
		metric.New(
			"cpu",
			map[string]string{
				"topic": "xyzzy",
			},
			map[string]any{
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

func TestHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected []sarama.RecordHeader
	}{
		{
			name: "none",
		},
		{
			name:    "static string",
			headers: map[string]string{"agent": "telegraf"},
			expected: []sarama.RecordHeader{
				{
					Key:   []byte("agent"),
					Value: []byte("telegraf"),
				},
			},
		},
		{
			name:    "metric name header",
			headers: map[string]string{"metric": "{{ .Name }}"},
			expected: []sarama.RecordHeader{
				{
					Key:   []byte("metric"),
					Value: []byte("cpu"),
				},
			},
		},
		{
			name: "complex",
			headers: map[string]string{
				"source": `{{ .Tag "source" }}:{{ .Tag "topic"}}`,
				"device": `{{ .Name }}-{{ .Field "id" }}`,
			},
			expected: []sarama.RecordHeader{
				{
					Key:   []byte("source"),
					Value: []byte("server:xyzzy"),
				},
				{
					Key:   []byte("device"),
					Value: []byte("cpu-3254345daab4"),
				},
			},
		},
	}

	// Define an input metric for writing
	input := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{
				"topic":  "xyzzy",
				"source": "server",
			},
			map[string]any{
				"id":    "3254345daab4",
				"value": 42.0,
				"hours": 255,
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
				Brokers:      []string{"127.0.0.1"},
				Topic:        "telegraf",
				Headers:      tt.headers,
				Log:          testutil.Logger{},
				producerFunc: newMockProducer,
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

			require.ElementsMatch(t, tt.expected, message.Headers)
		})
	}
}

func TestWritePartialFailure(t *testing.T) {
	producer := &mockProducer{
		sendError: func(msgs []*sarama.ProducerMessage) error {
			return sarama.ProducerErrors{{Msg: msgs[1], Err: sarama.ErrUnknownTopicOrPartition}}
		},
	}
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin := &Kafka{
		TopicTag:   "topic",
		Log:        testutil.Logger{},
		producer:   producer,
		serializer: serializer,
	}
	metrics := []telegraf.Metric{
		metric.New("cpu", map[string]string{"topic": "existing"}, map[string]any{"value": 1}, time.Unix(1, 0)),
		metric.New("cpu", map[string]string{"topic": "missing"}, map[string]any{"value": 2}, time.Unix(1, 0)),
		metric.New("cpu", map[string]string{"topic": "existing"}, map[string]any{"value": 3}, time.Unix(1, 0)),
	}
	err := plugin.Write(metrics)
	var partial *internal.PartialWriteError
	require.ErrorAs(t, err, &partial)
	require.Equal(t, []int{0, 2}, partial.MetricsAccept)
	require.Empty(t, partial.MetricsReject)
}

func TestWriteDeliveryResults(t *testing.T) {
	tests := []struct {
		name           string
		failures       map[int]error
		batchError     error
		invalidMetrics []int
		wrap           bool
		accept         []int
		reject         []int
	}{
		{
			name: "success",
		},
		{
			name:     "unknown topic retry",
			failures: map[int]error{1: sarama.ErrUnknownTopicOrPartition},
			accept:   []int{0, 2},
		},
		{
			name:     "wrapped errors",
			failures: map[int]error{1: fmt.Errorf("wrapped: %w", sarama.ErrUnknownTopicOrPartition)},
			wrap:     true,
			accept:   []int{0, 2},
		},
		{
			name:     "configuration then transient",
			failures: map[int]error{0: sarama.ConfigurationError("oversized message"), 1: sarama.ErrNotEnoughReplicas},
			accept:   []int{2},
			reject:   []int{0},
		},
		{
			name:     "transient then configuration",
			failures: map[int]error{0: sarama.ErrNotEnoughReplicas, 1: sarama.ConfigurationError("unsupported headers")},
			accept:   []int{2},
			reject:   []int{1},
		},
		{
			name:     "wrapped configuration error",
			failures: map[int]error{1: fmt.Errorf("wrapped: %w", sarama.ConfigurationError("oversized message"))},
			wrap:     true,
			accept:   []int{0, 2},
			reject:   []int{1},
		},
		{
			name:       "unclassified batch configuration error",
			batchError: sarama.ConfigurationError("batch delivery status unknown"),
		},
		{
			name:           "serialization gap before configuration error",
			invalidMetrics: []int{1},
			failures:       map[int]error{1: sarama.ConfigurationError("oversized message")},
			accept:         []int{0},
			reject:         []int{1, 2},
		},
		{
			name:     "size then transient",
			failures: map[int]error{0: sarama.ErrMessageSizeTooLarge, 1: sarama.ErrNotEnoughReplicas},
			accept:   []int{2},
			reject:   []int{0},
		},
		{
			name:     "transient then size",
			failures: map[int]error{0: sarama.ErrNotEnoughReplicas, 1: sarama.ErrMessageSizeTooLarge},
			accept:   []int{2},
			reject:   []int{1},
		},
		{
			name:     "timestamp then transient",
			failures: map[int]error{0: sarama.ErrInvalidTimestamp, 1: sarama.ErrNotEnoughReplicas},
			accept:   []int{2},
			reject:   []int{0},
		},
		{
			name:     "all transient",
			failures: map[int]error{0: sarama.ErrNotEnoughReplicas, 1: sarama.ErrNotEnoughReplicas, 2: sarama.ErrNotEnoughReplicas},
		},
		{
			name:       "unclassified error",
			batchError: errors.New("connection lost"),
		},
		{
			name:       "empty producer errors",
			batchError: sarama.ProducerErrors{},
		},
		{
			name:           "serialization then success",
			invalidMetrics: []int{0},
			accept:         []int{1, 2},
			reject:         []int{0},
		},
		{
			name:           "serialization then partial",
			invalidMetrics: []int{0},
			failures:       map[int]error{0: sarama.ErrUnknownTopicOrPartition},
			accept:         []int{2},
			reject:         []int{0},
		},
		{
			name:           "serialization then unclassified",
			invalidMetrics: []int{0},
			batchError:     errors.New("connection lost"),
			reject:         []int{0},
		},
		{
			name:           "serialization gap before retry",
			invalidMetrics: []int{1},
			failures:       map[int]error{1: sarama.ErrUnknownTopicOrPartition},
			accept:         []int{0},
			reject:         []int{1},
		},
		{
			name:           "serialization gap before rejection",
			invalidMetrics: []int{1},
			failures:       map[int]error{1: sarama.ErrMessageSizeTooLarge},
			accept:         []int{0},
			reject:         []int{1, 2},
		},
		{
			name:           "serialization gaps around delivery",
			invalidMetrics: []int{0, 2},
			accept:         []int{1},
			reject:         []int{0, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := &mockProducer{sendError: func(msgs []*sarama.ProducerMessage) error {
				if tt.batchError != nil {
					return tt.batchError
				}
				var failures sarama.ProducerErrors
				// Deliberately return failures out of message order.
				for i, msg := range slices.Backward(msgs) {
					if err, failed := tt.failures[i]; failed {
						failures = append(failures, &sarama.ProducerError{Msg: msg, Err: err})
					}
				}
				if len(failures) == 0 {
					return nil
				}
				if tt.wrap {
					return fmt.Errorf("send failed: %w", failures)
				}
				return failures
			}}
			serializer := &influx.Serializer{}
			require.NoError(t, serializer.Init())
			plugin := &Kafka{
				Topic:      "test",
				Log:        testutil.Logger{},
				producer:   producer,
				serializer: serializer,
			}
			require.NoError(t, plugin.Init())
			metrics := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2), testutil.TestMetric(3)}
			for _, index := range tt.invalidMetrics {
				metrics[index].SetName("")
			}
			err := plugin.Write(metrics)
			if len(tt.failures) == 0 && tt.batchError == nil && len(tt.invalidMetrics) == 0 {
				require.NoError(t, err)
				return
			}
			var partial *internal.PartialWriteError
			require.ErrorAs(t, err, &partial)
			require.Equal(t, tt.accept, partial.MetricsAccept)
			require.Equal(t, tt.reject, partial.MetricsReject)
			require.Len(t, partial.MetricsRejectErrors, len(tt.reject))
			require.NotEmpty(t, partial.Error())
		})
	}
}

func TestWriteNoMessages(t *testing.T) {
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin := &Kafka{Log: testutil.Logger{}, serializer: serializer}
	// No producer is needed when the batch has no serializable metrics.
	require.NoError(t, plugin.Write(nil))
	m := testutil.TestMetric(1)
	m.SetName("")
	err := plugin.Write([]telegraf.Metric{m})
	var partial *internal.PartialWriteError
	require.ErrorAs(t, err, &partial)
	require.Equal(t, []int{0}, partial.MetricsReject)
	require.Empty(t, partial.MetricsAccept)
}

func TestWriteUnknownTopicWithBroker(t *testing.T) {
	broker := sarama.NewMockBroker(t, 0)
	defer broker.Close()
	broker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(broker.Addr(), broker.BrokerID()).
			SetLeader("existing", 0, broker.BrokerID()),
		"ProduceRequest": sarama.NewMockProduceResponse(t),
	})
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V0_11_0_0
	cfg.ApiVersionsRequest = false
	cfg.Producer.Return.Successes = true
	cfg.Producer.Retry.Max = 0
	cfg.Metadata.Retry.Max = 0
	cfg.Net.DialTimeout = 2 * time.Second
	cfg.Net.ReadTimeout = 2 * time.Second
	cfg.Net.WriteTimeout = 2 * time.Second
	producer, err := sarama.NewSyncProducer([]string{broker.Addr()}, cfg)
	require.NoError(t, err)
	defer func() { require.NoError(t, producer.Close()) }()

	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin := &Kafka{
		TopicTag:        "topic",
		ExcludeTopicTag: true,
		Log:             testutil.Logger{},
		producer:        producer,
		serializer:      serializer,
	}
	buffer, err := models.NewBuffer("kafka", t.Name(), "", 10, "memory", "", false)
	require.NoError(t, err)
	defer func() { require.NoError(t, buffer.Close()) }()
	var delivered []telegraf.DeliveryInfo
	for _, topic := range []string{"missing", "existing", "existing"} {
		m, _ := metric.WithTracking(
			metric.New("cpu", map[string]string{"topic": topic}, map[string]any{"value": 1}, time.Unix(1, 0)),
			func(info telegraf.DeliveryInfo) { delivered = append(delivered, info) },
		)
		require.Zero(t, buffer.Add(m))
	}

	// A mixed batch removes successful metrics without changing the
	// buffered topic tags, even when serialization excludes that tag.
	tx := buffer.BeginTransaction(2)
	var partial *internal.PartialWriteError
	require.ErrorAs(t, plugin.Write(tx.Batch), &partial)
	require.Equal(t, []int{1}, partial.MetricsAccept, "delivery result: %v", partial)
	for _, m := range tx.Batch {
		require.True(t, m.HasTag("topic"))
	}
	tx.Accept, tx.Reject = partial.MetricsAccept, partial.MetricsReject
	buffer.EndTransaction(tx)

	// The following flush must reach the newer valid metric too.
	tx = buffer.BeginTransaction(2)
	err = plugin.Write(tx.Batch)
	require.ErrorAs(t, err, &partial)
	require.Equal(t, []int{1}, partial.MetricsAccept)
	require.Empty(t, partial.MetricsReject)
	tx.Accept, tx.Reject = partial.MetricsAccept, partial.MetricsReject
	buffer.EndTransaction(tx)
	require.Equal(t, 1, buffer.Len())
	require.Len(t, delivered, 2)
	for _, info := range delivered {
		require.True(t, info.Delivered())
	}

	// A temporarily missing topic can recover without losing the
	// retained metric or acknowledging it before delivery.
	broker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(broker.Addr(), broker.BrokerID()).
			SetLeader("existing", 0, broker.BrokerID()).
			SetLeader("missing", 0, broker.BrokerID()),
		"ProduceRequest": sarama.NewMockProduceResponse(t),
	})
	tx = buffer.BeginTransaction(2)
	require.NoError(t, plugin.Write(tx.Batch))
	tx.AcceptAll()
	buffer.EndTransaction(tx)
	require.Zero(t, buffer.Len())
	require.Len(t, delivered, 3)
	require.True(t, delivered[2].Delivered())
}

type mockProducer struct {
	sent      []*sarama.ProducerMessage
	sendError func([]*sarama.ProducerMessage) error
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
	if p.sendError != nil {
		return p.sendError(msgs)
	}
	return nil
}

func (*mockProducer) Close() error {
	return nil
}
