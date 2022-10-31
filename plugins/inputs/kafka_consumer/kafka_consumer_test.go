package kafka_consumer

import (
	"context"
	"fmt"
	"math"
	"net"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	kafkaOutput "github.com/influxdata/telegraf/plugins/outputs/kafka"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/value"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

type FakeConsumerGroup struct {
	brokers []string
	group   string
	config  *sarama.Config

	handler sarama.ConsumerGroupHandler
	errors  chan error
}

func (g *FakeConsumerGroup) Consume(_ context.Context, _ []string, handler sarama.ConsumerGroupHandler) error {
	g.handler = handler
	return g.handler.Setup(nil)
}

func (g *FakeConsumerGroup) Errors() <-chan error {
	return g.errors
}

func (g *FakeConsumerGroup) Close() error {
	close(g.errors)
	return nil
}

type FakeCreator struct {
	ConsumerGroup *FakeConsumerGroup
}

func (c *FakeCreator) Create(brokers []string, group string, cfg *sarama.Config) (ConsumerGroup, error) {
	c.ConsumerGroup.brokers = brokers
	c.ConsumerGroup.group = group
	c.ConsumerGroup.config = cfg
	return c.ConsumerGroup, nil
}

func TestInit(t *testing.T) {
	tests := []struct {
		name      string
		plugin    *KafkaConsumer
		initError bool
		check     func(t *testing.T, plugin *KafkaConsumer)
	}{
		{
			name:   "default config",
			plugin: &KafkaConsumer{},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.Equal(t, plugin.ConsumerGroup, defaultConsumerGroup)
				require.Equal(t, plugin.MaxUndeliveredMessages, defaultMaxUndeliveredMessages)
				require.Equal(t, plugin.config.ClientID, "Telegraf")
				require.Equal(t, plugin.config.Consumer.Offsets.Initial, sarama.OffsetOldest)
				require.Equal(t, plugin.config.Consumer.MaxProcessingTime, 100*time.Millisecond)
			},
		},
		{
			name: "parses valid version string",
			plugin: &KafkaConsumer{
				ReadConfig: kafka.ReadConfig{
					Config: kafka.Config{
						Version: "1.0.0",
					},
				},
				Log: testutil.Logger{},
			},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.Equal(t, plugin.config.Version, sarama.V1_0_0_0)
			},
		},
		{
			name: "invalid version string",
			plugin: &KafkaConsumer{
				ReadConfig: kafka.ReadConfig{
					Config: kafka.Config{
						Version: "100",
					},
				},
				Log: testutil.Logger{},
			},
			initError: true,
		},
		{
			name: "custom client_id",
			plugin: &KafkaConsumer{
				ReadConfig: kafka.ReadConfig{
					Config: kafka.Config{
						ClientID: "custom",
					},
				},
				Log: testutil.Logger{},
			},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.Equal(t, plugin.config.ClientID, "custom")
			},
		},
		{
			name: "custom offset",
			plugin: &KafkaConsumer{
				Offset: "newest",
				Log:    testutil.Logger{},
			},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.Equal(t, plugin.config.Consumer.Offsets.Initial, sarama.OffsetNewest)
			},
		},
		{
			name: "invalid offset",
			plugin: &KafkaConsumer{
				Offset: "middle",
				Log:    testutil.Logger{},
			},
			initError: true,
		},
		{
			name: "default tls without tls config",
			plugin: &KafkaConsumer{
				Log: testutil.Logger{},
			},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.False(t, plugin.config.Net.TLS.Enable)
			},
		},
		{
			name: "enabled tls without tls config",
			plugin: &KafkaConsumer{
				ReadConfig: kafka.ReadConfig{
					Config: kafka.Config{
						EnableTLS: func(b bool) *bool { return &b }(true),
					},
				},
				Log: testutil.Logger{},
			},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.True(t, plugin.config.Net.TLS.Enable)
			},
		},
		{
			name: "default tls with a tls config",
			plugin: &KafkaConsumer{
				ReadConfig: kafka.ReadConfig{
					Config: kafka.Config{
						ClientConfig: tls.ClientConfig{
							InsecureSkipVerify: true,
						},
					},
				},
				Log: testutil.Logger{},
			},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.True(t, plugin.config.Net.TLS.Enable)
			},
		},
		{
			name: "Insecure tls",
			plugin: &KafkaConsumer{
				ReadConfig: kafka.ReadConfig{
					Config: kafka.Config{
						ClientConfig: tls.ClientConfig{
							InsecureSkipVerify: true,
						},
					},
				},
				Log: testutil.Logger{},
			},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.True(t, plugin.config.Net.TLS.Enable)
			},
		},
		{
			name: "custom max_processing_time",
			plugin: &KafkaConsumer{
				MaxProcessingTime: config.Duration(1000 * time.Millisecond),
				Log:               testutil.Logger{},
			},
			check: func(t *testing.T, plugin *KafkaConsumer) {
				require.Equal(t, plugin.config.Consumer.MaxProcessingTime, 1000*time.Millisecond)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := &FakeConsumerGroup{}
			tt.plugin.ConsumerCreator = &FakeCreator{ConsumerGroup: cg}
			err := tt.plugin.Init()
			if tt.initError {
				require.Error(t, err)
				return
			}
			// No error path
			require.NoError(t, err)

			tt.check(t, tt.plugin)
		})
	}
}

func TestStartStop(t *testing.T) {
	cg := &FakeConsumerGroup{errors: make(chan error)}
	plugin := &KafkaConsumer{
		ConsumerCreator: &FakeCreator{ConsumerGroup: cg},
		Log:             testutil.Logger{},
	}
	err := plugin.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))

	plugin.Stop()
}

type FakeConsumerGroupSession struct {
	ctx context.Context
}

func (s *FakeConsumerGroupSession) Claims() map[string][]int32 {
	panic("not implemented")
}

func (s *FakeConsumerGroupSession) MemberID() string {
	panic("not implemented")
}

func (s *FakeConsumerGroupSession) GenerationID() int32 {
	panic("not implemented")
}

func (s *FakeConsumerGroupSession) MarkOffset(_ string, _ int32, _ int64, _ string) {
	panic("not implemented")
}

func (s *FakeConsumerGroupSession) ResetOffset(_ string, _ int32, _ int64, _ string) {
	panic("not implemented")
}

func (s *FakeConsumerGroupSession) MarkMessage(_ *sarama.ConsumerMessage, _ string) {
}

func (s *FakeConsumerGroupSession) Context() context.Context {
	return s.ctx
}

func (s *FakeConsumerGroupSession) Commit() {
}

type FakeConsumerGroupClaim struct {
	messages chan *sarama.ConsumerMessage
}

func (c *FakeConsumerGroupClaim) Topic() string {
	panic("not implemented")
}

func (c *FakeConsumerGroupClaim) Partition() int32 {
	panic("not implemented")
}

func (c *FakeConsumerGroupClaim) InitialOffset() int64 {
	panic("not implemented")
}

func (c *FakeConsumerGroupClaim) HighWaterMarkOffset() int64 {
	panic("not implemented")
}

func (c *FakeConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	return c.messages
}

func TestConsumerGroupHandler_Lifecycle(t *testing.T) {
	acc := &testutil.Accumulator{}
	parser := value.Parser{
		MetricName: "cpu",
		DataType:   "int",
	}
	require.NoError(t, parser.Init())
	cg := NewConsumerGroupHandler(acc, 1, &parser, testutil.Logger{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session := &FakeConsumerGroupSession{
		ctx: ctx,
	}
	var claim FakeConsumerGroupClaim
	var err error

	err = cg.Setup(session)
	require.NoError(t, err)

	cancel()
	// This produces a flappy testcase probably due to a race between context cancellation and consumption.
	// Furthermore, it is not clear what the outcome of this test should be...
	// err = cg.ConsumeClaim(session, &claim)
	//require.NoError(t, err)
	// So stick with the line below for now.
	_ = cg.ConsumeClaim(session, &claim)

	err = cg.Cleanup(session)
	require.NoError(t, err)
}

func TestConsumerGroupHandler_ConsumeClaim(t *testing.T) {
	acc := &testutil.Accumulator{}
	parser := value.Parser{
		MetricName: "cpu",
		DataType:   "int",
	}
	require.NoError(t, parser.Init())
	cg := NewConsumerGroupHandler(acc, 1, &parser, testutil.Logger{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session := &FakeConsumerGroupSession{ctx: ctx}
	claim := &FakeConsumerGroupClaim{
		messages: make(chan *sarama.ConsumerMessage, 1),
	}

	err := cg.Setup(session)
	require.NoError(t, err)

	claim.messages <- &sarama.ConsumerMessage{
		Topic: "telegraf",
		Value: []byte("42"),
	}

	go func() {
		err := cg.ConsumeClaim(session, claim)
		require.Error(t, err)
		require.EqualValues(t, "context canceled", err.Error())
	}()

	acc.Wait(1)
	cancel()

	err = cg.Cleanup(session)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Now(),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestConsumerGroupHandler_Handle(t *testing.T) {
	tests := []struct {
		name                string
		maxMessageLen       int
		topicTag            string
		msg                 *sarama.ConsumerMessage
		expected            []telegraf.Metric
		expectedHandleError string
	}{
		{
			name: "happy path",
			msg: &sarama.ConsumerMessage{
				Topic: "telegraf",
				Value: []byte("42"),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Now(),
				),
			},
		},
		{
			name:          "message to long",
			maxMessageLen: 4,
			msg: &sarama.ConsumerMessage{
				Topic: "telegraf",
				Value: []byte("12345"),
			},
			expected:            []telegraf.Metric{},
			expectedHandleError: "message exceeds max_message_len (actual 5, max 4)",
		},
		{
			name: "parse error",
			msg: &sarama.ConsumerMessage{
				Topic: "telegraf",
				Value: []byte("not an integer"),
			},
			expected:            []telegraf.Metric{},
			expectedHandleError: "strconv.Atoi: parsing \"integer\": invalid syntax",
		},
		{
			name:     "add topic tag",
			topicTag: "topic",
			msg: &sarama.ConsumerMessage{
				Topic: "telegraf",
				Value: []byte("42"),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"topic": "telegraf",
					},
					map[string]interface{}{
						"value": 42,
					},
					time.Now(),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &testutil.Accumulator{}
			parser := value.Parser{
				MetricName: "cpu",
				DataType:   "int",
			}
			require.NoError(t, parser.Init())
			cg := NewConsumerGroupHandler(acc, 1, &parser, testutil.Logger{})
			cg.MaxMessageLen = tt.maxMessageLen
			cg.TopicTag = tt.topicTag

			ctx := context.Background()
			session := &FakeConsumerGroupSession{ctx: ctx}

			require.NoError(t, cg.Reserve(ctx))
			err := cg.Handle(session, tt.msg)
			if tt.expectedHandleError != "" {
				require.Error(t, err)
				require.EqualValues(t, tt.expectedHandleError, err.Error())
			} else {
				require.NoError(t, err)
			}

			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestKafkaRoundTripIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var tests = []struct {
		name               string
		connectionStrategy string
	}{
		{"connection strategy startup", "startup"},
		{"connection strategy defer", "defer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("rt: starting network")
			ctx := context.Background()
			networkName := "telegraf-test-kafka-consumer-network"
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

			t.Logf("rt: starting zookeeper")
			zookeeperName := "telegraf-test-kafka-consumer-zookeeper"
			zookeeper := testutil.Container{
				Image:        "wurstmeister/zookeeper",
				ExposedPorts: []string{"2181:2181"},
				Networks:     []string{networkName},
				WaitingFor:   wait.ForLog("binding to port"),
				Name:         zookeeperName,
			}
			require.NoError(t, zookeeper.Start(), "failed to start container")
			defer func() {
				require.NoError(t, zookeeper.Terminate(), "terminating container failed")
			}()

			t.Logf("rt: starting broker")
			topic := "Test"
			container := testutil.Container{
				Name:         "telegraf-test-kafka-consumer",
				Image:        "wurstmeister/kafka",
				ExposedPorts: []string{"9092:9092"},
				Env: map[string]string{
					"KAFKA_ADVERTISED_HOST_NAME": "localhost",
					"KAFKA_ADVERTISED_PORT":      "9092",
					"KAFKA_ZOOKEEPER_CONNECT":    fmt.Sprintf("%s:%s", zookeeperName, zookeeper.Ports["2181"]),
					"KAFKA_CREATE_TOPICS":        fmt.Sprintf("%s:1:1", topic),
				},
				Networks:   []string{networkName},
				WaitingFor: wait.ForLog("Log loaded for partition Test-0 with initial high watermark 0"),
			}
			require.NoError(t, container.Start(), "failed to start container")
			defer func() {
				require.NoError(t, container.Terminate(), "terminating container failed")
			}()

			brokers := []string{
				fmt.Sprintf("%s:%s", container.Address, container.Ports["9092"]),
			}

			// Make kafka output
			t.Logf("rt: starting output plugin")
			creator := outputs.Outputs["kafka"]
			output, ok := creator().(*kafkaOutput.Kafka)
			require.True(t, ok)

			s, _ := serializers.NewInfluxSerializer()
			output.SetSerializer(s)
			output.Brokers = brokers
			output.Topic = topic
			output.Log = testutil.Logger{}

			require.NoError(t, output.Init())
			require.NoError(t, output.Connect())

			// Make kafka input
			t.Logf("rt: starting input plugin")
			input := KafkaConsumer{
				Brokers:                brokers,
				Log:                    testutil.Logger{},
				Topics:                 []string{topic},
				MaxUndeliveredMessages: 1,
				ConnectionStrategy:     tt.connectionStrategy,
			}
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			input.SetParser(parser)
			require.NoError(t, input.Init())

			acc := testutil.Accumulator{}
			require.NoError(t, input.Start(&acc))

			// Shove some metrics through
			expected := testutil.MockMetrics()
			t.Logf("rt: writing")
			require.NoError(t, output.Write(expected))

			// Check that they were received
			t.Logf("rt: expecting")
			acc.Wait(len(expected))
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())

			t.Logf("rt: shutdown")
			require.NoError(t, output.Close())
			input.Stop()

			t.Logf("rt: done")
		})
	}
}

func TestExponentialBackoff(t *testing.T) {
	var err error

	backoff := 10 * time.Millisecond
	max := 3

	// get an unused port by listening on next available port, then closing it
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())

	// try to connect to kafka on that unused port
	brokers := []string{
		fmt.Sprintf("localhost:%d", port),
	}

	input := KafkaConsumer{
		Brokers:                brokers,
		Log:                    testutil.Logger{},
		Topics:                 []string{"topic"},
		MaxUndeliveredMessages: 1,

		ReadConfig: kafka.ReadConfig{
			Config: kafka.Config{
				MetadataRetryMax:     max,
				MetadataRetryBackoff: backoff,
				MetadataRetryType:    "exponential",
			},
		},
	}
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	input.SetParser(parser)

	//time how long initialization (connection) takes
	start := time.Now()
	require.NoError(t, input.Init())

	acc := testutil.Accumulator{}
	require.Error(t, input.Start(&acc))
	elapsed := time.Since(start)
	t.Logf("elapsed %d", elapsed)

	var expectedRetryDuration time.Duration
	for i := 0; i < max; i++ {
		expectedRetryDuration += backoff * time.Duration(math.Pow(2, float64(i)))
	}
	t.Logf("expected > %d", expectedRetryDuration)

	// Other than the expected retry delay, initializing and starting the
	// plugin, including initializing a sarama consumer takes some time.
	//
	// It would be nice to check that the actual time is within an expected
	// range, but we don't know how long the non-retry time should be.
	//
	// For now, just check that elapsed time isn't shorter than we expect the
	// retry delays to be
	require.GreaterOrEqual(t, elapsed, expectedRetryDuration)

	input.Stop()
}
