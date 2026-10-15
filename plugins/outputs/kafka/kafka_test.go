package kafka

import (
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs"
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
		producerFunc: newProducer,
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

type mockProducer struct {
	sent []*sarama.ProducerMessage
	sarama.SyncProducer
	sync.Mutex
}

func newMockProducer(_ []string, _ *sarama.Config) (sarama.SyncProducer, io.Closer, error) {
	return &mockProducer{}, nil, nil
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

func TestDeliveryTimeoutAbandonsStuckProducer(t *testing.T) {
	stuck := newBlockingProducer()
	factory := &producerFactory{producers: []sarama.SyncProducer{stuck, &mockProducer{}}}
	plugin := newTimeoutTestPlugin(t, factory, 50*time.Millisecond)

	// Abandon the stuck delivery
	err := plugin.Write(timeoutTestMetrics())
	require.EqualError(t, err, "delivery timed out after 50ms")
	require.Nil(t, plugin.producer)
	require.EqualValues(t, 1, plugin.abandoned.Load())

	// Write using a new producer
	require.NoError(t, plugin.Write(timeoutTestMetrics()))
	require.Equal(t, 2, factory.count())
	replacement, ok := plugin.producer.(*mockProducer)
	require.True(t, ok, "invalid producer type")
	replacement.Lock()
	require.Len(t, replacement.sent, 1)
	replacement.Unlock()

	// Close the client right away and the producer after the delivery returned
	require.Eventually(t, factory.clients[0].closed.Load, time.Second, 10*time.Millisecond)
	require.False(t, stuck.isClosed())
	close(stuck.release)
	require.Eventually(t, func() bool {
		return plugin.abandoned.Load() == 0 && stuck.isClosed()
	}, time.Second, 10*time.Millisecond)
}

func TestDeliveryTimeoutKeepsProducerOnResult(t *testing.T) {
	sendErr := &sarama.ProducerError{Msg: &sarama.ProducerMessage{Topic: "telegraf"}, Err: sarama.ErrOutOfBrokers}
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "success",
		},
		{
			name:     "error",
			err:      sarama.ProducerErrors{sendErr},
			expected: sendErr.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := &erroringProducer{err: tt.err}
			factory := &producerFactory{producers: []sarama.SyncProducer{producer}}
			plugin := newTimeoutTestPlugin(t, factory, time.Minute)

			err := plugin.Write(timeoutTestMetrics())
			if tt.expected == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.expected)
			}
			require.Same(t, producer, plugin.producer)
			require.Zero(t, plugin.abandoned.Load())
			require.Equal(t, 1, factory.count())
		})
	}
}

func TestDeliveryTimeoutRepeated(t *testing.T) {
	stuck := []*blockingProducer{newBlockingProducer(), newBlockingProducer(), newBlockingProducer()}
	factory := &producerFactory{producers: []sarama.SyncProducer{stuck[0], stuck[1], stuck[2], &mockProducer{}}}
	plugin := newTimeoutTestPlugin(t, factory, 10*time.Millisecond)

	// Keep replacing producers while deliveries are stuck
	for range stuck {
		require.EqualError(t, plugin.Write(timeoutTestMetrics()), "delivery timed out after 10ms")
	}
	require.NoError(t, plugin.Write(timeoutTestMetrics()))
	require.Equal(t, 4, factory.count())
	require.EqualValues(t, 3, plugin.abandoned.Load())
	require.Eventually(t, func() bool {
		return factory.clients[0].closed.Load() && factory.clients[1].closed.Load() && factory.clients[2].closed.Load()
	}, time.Second, 10*time.Millisecond)

	for _, p := range stuck {
		close(p.release)
	}
	require.Eventually(t, func() bool {
		return plugin.abandoned.Load() == 0
	}, time.Second, 10*time.Millisecond)
}

func TestDeliveryTimeoutDisabled(t *testing.T) {
	stuck := newBlockingProducer()
	factory := &producerFactory{producers: []sarama.SyncProducer{stuck}}
	plugin := newTimeoutTestPlugin(t, factory, 0)

	// Wait for the delivery without a timeout
	result := make(chan error, 1)
	go func() {
		result <- plugin.Write(timeoutTestMetrics())
	}()
	require.Never(t, func() bool {
		return len(result) > 0
	}, 100*time.Millisecond, 10*time.Millisecond)

	close(stuck.release)
	require.NoError(t, <-result)
	require.Same(t, stuck, plugin.producer)
	require.Zero(t, plugin.abandoned.Load())
}

func TestDeliveryTimeoutValidation(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected string
		warning  bool
	}{
		{
			name:    "default",
			timeout: 5 * time.Minute,
		},
		{
			name:    "disabled",
			timeout: 0,
		},
		{
			name:    "below maximum delivery duration",
			timeout: time.Minute,
			warning: true,
		},
		{
			name:     "negative",
			timeout:  -time.Second,
			expected: "delivery_timeout must not be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &testutil.CaptureLogger{}
			plugin := outputs.Outputs["kafka"]().(*Kafka)
			plugin.Brokers = []string{"127.0.0.1"}
			plugin.Topic = "telegraf"
			plugin.DeliveryTimeout = config.Duration(tt.timeout)
			plugin.Log = logger

			err := plugin.Init()
			if tt.expected != "" {
				require.EqualError(t, err, tt.expected)
				return
			}
			require.NoError(t, err)

			var warned bool
			for _, msg := range logger.Warnings() {
				if strings.Contains(msg, "delivery_timeout 1m0s is below the maximum delivery duration of 4m0.3s") {
					warned = true
				}
			}
			require.Equal(t, tt.warning, warned, "unexpected warnings: %v", logger.Warnings())
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name       string
		blockClose bool
		expected   string
	}{
		{
			name: "success",
		},
		{
			name:       "producer close times out",
			blockClose: true,
			expected:   "closing producer timed out after 10ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := newBlockingProducer()
			producer.blockClose = tt.blockClose
			defer close(producer.release)
			client := &mockClient{}

			plugin := &Kafka{
				producer:     producer,
				client:       client,
				closeTimeout: 10 * time.Millisecond,
				Log:          testutil.Logger{},
			}
			err := plugin.Close()
			if tt.expected == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.expected)
			}
			require.True(t, client.closed.Load())
			require.Nil(t, plugin.producer)
			require.Nil(t, plugin.client)
		})
	}
}

func newTimeoutTestPlugin(t *testing.T, factory *producerFactory, timeout time.Duration) *Kafka {
	t.Helper()

	for range factory.producers {
		factory.clients = append(factory.clients, &mockClient{})
	}

	s := &influx.Serializer{}
	require.NoError(t, s.Init())

	plugin := &Kafka{
		Brokers:         []string{"127.0.0.1"},
		Topic:           "telegraf",
		DeliveryTimeout: config.Duration(timeout),
		Log:             testutil.Logger{},
		producerFunc:    factory.create,
		closeTimeout:    time.Second,
	}
	plugin.SetSerializer(s)
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	return plugin
}

func timeoutTestMetrics() []telegraf.Metric {
	return []telegraf.Metric{
		metric.New("cpu", map[string]string{}, map[string]any{"time_idle": 42.0}, time.Unix(0, 0)),
	}
}

// producerFactory returns the given producers in order, each with a client
type producerFactory struct {
	producers []sarama.SyncProducer
	clients   []*mockClient
	created   int
	sync.Mutex
}

func (f *producerFactory) create(_ []string, _ *sarama.Config) (sarama.SyncProducer, io.Closer, error) {
	f.Lock()
	defer f.Unlock()
	if f.created >= len(f.producers) {
		return nil, nil, errors.New("no more producers")
	}
	p := f.producers[f.created]
	f.created++
	return p, f.clients[f.created-1], nil
}

func (f *producerFactory) count() int {
	f.Lock()
	defer f.Unlock()
	return f.created
}

// blockingProducer blocks deliveries, and optionally Close, until released
type blockingProducer struct {
	sarama.SyncProducer
	release    chan struct{}
	blockClose bool
	closed     atomic.Bool
}

func newBlockingProducer() *blockingProducer {
	return &blockingProducer{release: make(chan struct{})}
}

func (p *blockingProducer) SendMessages([]*sarama.ProducerMessage) error {
	<-p.release
	return nil
}

func (p *blockingProducer) Close() error {
	if p.blockClose {
		<-p.release
	}
	p.closed.Store(true)
	return nil
}

func (p *blockingProducer) isClosed() bool {
	return p.closed.Load()
}

type erroringProducer struct {
	sarama.SyncProducer
	err error
}

func (p *erroringProducer) SendMessages([]*sarama.ProducerMessage) error {
	return p.err
}

func (*erroringProducer) Close() error {
	return nil
}

type mockClient struct {
	closed atomic.Bool
}

func (c *mockClient) Close() error {
	c.closed.Store(true)
	return nil
}
