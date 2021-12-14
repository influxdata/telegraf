package kafka_consumer

import (
	"context"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/parsers/value"
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
	err = plugin.Start(&acc)
	require.NoError(t, err)

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
	parser := value.NewValueParser("cpu", "int", "", nil)
	cg := NewConsumerGroupHandler(acc, 1, parser, testutil.Logger{})

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
	//nolint:errcheck
	cg.ConsumeClaim(session, &claim)

	err = cg.Cleanup(session)
	require.NoError(t, err)
}

func TestConsumerGroupHandler_ConsumeClaim(t *testing.T) {
	acc := &testutil.Accumulator{}
	parser := value.NewValueParser("cpu", "int", "", nil)
	cg := NewConsumerGroupHandler(acc, 1, parser, testutil.Logger{})

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
			parser := value.NewValueParser("cpu", "int", "", nil)
			cg := NewConsumerGroupHandler(acc, 1, parser, testutil.Logger{})
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
