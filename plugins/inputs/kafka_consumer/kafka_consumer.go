package kafka_consumer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const sampleConfig = `
  ## Kafka brokers.
  brokers = ["localhost:9092"]

  ## Topics to consume.
  topics = ["telegraf"]

  ## When set this tag will be added to all metrics with the topic as the value.
  # topic_tag = ""

  ## Optional Client id
  # client_id = "Telegraf"

  ## Set the minimal supported Kafka version.  Setting this enables the use of new
  ## Kafka features and APIs.  Must be 0.10.2.0 or greater.
  ##   ex: version = "1.1.0"
  # version = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## SASL authentication credentials.  These settings should typically be used
  ## with TLS encryption enabled
  # sasl_username = "kafka"
  # sasl_password = "secret"

  ## Optional SASL:
  ## one of: OAUTHBEARER, PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI
  ## (defaults to PLAIN)
  # sasl_mechanism = ""

  ## used if sasl_mechanism is GSSAPI (experimental)
  # sasl_gssapi_service_name = ""
  # ## One of: KRB5_USER_AUTH and KRB5_KEYTAB_AUTH
  # sasl_gssapi_auth_type = "KRB5_USER_AUTH"
  # sasl_gssapi_kerberos_config_path = "/"
  # sasl_gssapi_realm = "realm"
  # sasl_gssapi_key_tab_path = ""
  # sasl_gssapi_disable_pafxfast = false

  ## used if sasl_mechanism is OAUTHBEARER (experimental)
  # sasl_access_token = ""

  ## SASL protocol version.  When connecting to Azure EventHub set to 0.
  # sasl_version = 1

  # Disable Kafka metadata full fetch
  # metadata_full = false

  ## Name of the consumer group.
  # consumer_group = "telegraf_metrics_consumers"

  ## Compression codec represents the various compression codecs recognized by
  ## Kafka in messages.
  ##  0 : None
  ##  1 : Gzip
  ##  2 : Snappy
  ##  3 : LZ4
  ##  4 : ZSTD
   # compression_codec = 0

  ## Initial offset position; one of "oldest" or "newest".
  # offset = "oldest"

  ## Consumer group partition assignment strategy; one of "range", "roundrobin" or "sticky".
  # balance_strategy = "range"

  ## Maximum length of a message to consume, in bytes (default 0/unlimited);
  ## larger messages are dropped
  max_message_len = 1000000

  ## Maximum messages to read from the broker that have not been written by an
  ## output.  For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Maximum amount of time the consumer should take to process messages. If
  ## the debug log prints messages from sarama about 'abandoning subscription
  ## to [topic] because consuming was taking too long', increase this value to
  ## longer than the time taken by the output plugin(s).
  ##
  ## Note that the effective timeout could be between 'max_processing_time' and
  ## '2 * max_processing_time'.
  # max_processing_time = "100ms"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

const (
	defaultMaxUndeliveredMessages = 1000
	defaultMaxProcessingTime      = config.Duration(100 * time.Millisecond)
	defaultConsumerGroup          = "telegraf_metrics_consumers"
	reconnectDelay                = 5 * time.Second
)

type empty struct{}
type semaphore chan empty

type KafkaConsumer struct {
	Brokers                []string        `toml:"brokers"`
	ConsumerGroup          string          `toml:"consumer_group"`
	MaxMessageLen          int             `toml:"max_message_len"`
	MaxUndeliveredMessages int             `toml:"max_undelivered_messages"`
	MaxProcessingTime      config.Duration `toml:"max_processing_time"`
	Offset                 string          `toml:"offset"`
	BalanceStrategy        string          `toml:"balance_strategy"`
	Topics                 []string        `toml:"topics"`
	TopicTag               string          `toml:"topic_tag"`

	kafka.ReadConfig

	Log telegraf.Logger `toml:"-"`

	ConsumerCreator ConsumerGroupCreator `toml:"-"`
	consumer        ConsumerGroup
	config          *sarama.Config

	parser parsers.Parser
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

type ConsumerGroup interface {
	Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error
	Errors() <-chan error
	Close() error
}

type ConsumerGroupCreator interface {
	Create(brokers []string, group string, cfg *sarama.Config) (ConsumerGroup, error)
}

type SaramaCreator struct{}

func (*SaramaCreator) Create(brokers []string, group string, cfg *sarama.Config) (ConsumerGroup, error) {
	return sarama.NewConsumerGroup(brokers, group, cfg)
}

func (k *KafkaConsumer) SampleConfig() string {
	return sampleConfig
}

func (k *KafkaConsumer) Description() string {
	return "Read metrics from Kafka topics"
}

func (k *KafkaConsumer) SetParser(parser parsers.Parser) {
	k.parser = parser
}

func (k *KafkaConsumer) Init() error {
	if k.MaxUndeliveredMessages == 0 {
		k.MaxUndeliveredMessages = defaultMaxUndeliveredMessages
	}
	if time.Duration(k.MaxProcessingTime) == 0 {
		k.MaxProcessingTime = defaultMaxProcessingTime
	}
	if k.ConsumerGroup == "" {
		k.ConsumerGroup = defaultConsumerGroup
	}

	cfg := sarama.NewConfig()

	// Kafka version 0.10.2.0 is required for consumer groups.
	cfg.Version = sarama.V0_10_2_0

	if err := k.SetConfig(cfg); err != nil {
		return err
	}

	switch strings.ToLower(k.Offset) {
	case "oldest", "":
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		return fmt.Errorf("invalid offset %q", k.Offset)
	}

	switch strings.ToLower(k.BalanceStrategy) {
	case "range", "":
		cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	case "roundrobin":
		cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "sticky":
		cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	default:
		return fmt.Errorf("invalid balance strategy %q", k.BalanceStrategy)
	}

	if k.ConsumerCreator == nil {
		k.ConsumerCreator = &SaramaCreator{}
	}

	cfg.Consumer.MaxProcessingTime = time.Duration(k.MaxProcessingTime)

	k.config = cfg
	return nil
}

func (k *KafkaConsumer) Start(acc telegraf.Accumulator) error {
	var err error
	k.consumer, err = k.ConsumerCreator.Create(
		k.Brokers,
		k.ConsumerGroup,
		k.config,
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	k.cancel = cancel

	// Start consumer goroutine
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		for ctx.Err() == nil {
			handler := NewConsumerGroupHandler(acc, k.MaxUndeliveredMessages, k.parser, k.Log)
			handler.MaxMessageLen = k.MaxMessageLen
			handler.TopicTag = k.TopicTag
			err := k.consumer.Consume(ctx, k.Topics, handler)
			if err != nil {
				acc.AddError(err)
				// Ignore returned error as we cannot do anything about it anyway
				//nolint:errcheck,revive
				internal.SleepContext(ctx, reconnectDelay)
			}
		}
		err = k.consumer.Close()
		if err != nil {
			acc.AddError(err)
		}
	}()

	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		for err := range k.consumer.Errors() {
			acc.AddError(err)
		}
	}()

	return nil
}

func (k *KafkaConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (k *KafkaConsumer) Stop() {
	k.cancel()
	k.wg.Wait()
}

// Message is an aggregate type binding the Kafka message and the session so
// that offsets can be updated.
type Message struct {
	message *sarama.ConsumerMessage
	session sarama.ConsumerGroupSession
}

func NewConsumerGroupHandler(acc telegraf.Accumulator, maxUndelivered int, parser parsers.Parser, log telegraf.Logger) *ConsumerGroupHandler {
	handler := &ConsumerGroupHandler{
		acc:         acc.WithTracking(maxUndelivered),
		sem:         make(chan empty, maxUndelivered),
		undelivered: make(map[telegraf.TrackingID]Message, maxUndelivered),
		parser:      parser,
		log:         log,
	}
	return handler
}

// ConsumerGroupHandler is a sarama.ConsumerGroupHandler implementation.
type ConsumerGroupHandler struct {
	MaxMessageLen int
	TopicTag      string

	acc    telegraf.TrackingAccumulator
	sem    semaphore
	parser parsers.Parser
	wg     sync.WaitGroup
	cancel context.CancelFunc

	mu          sync.Mutex
	undelivered map[telegraf.TrackingID]Message

	log telegraf.Logger
}

// Setup is called once when a new session is opened.  It setups up the handler
// and begins processing delivered messages.
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.undelivered = make(map[telegraf.TrackingID]Message)

	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.run(ctx)
	}()
	return nil
}

// Run processes any delivered metrics during the lifetime of the session.
func (h *ConsumerGroupHandler) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case track := <-h.acc.Delivered():
			h.onDelivery(track)
		}
	}
}

func (h *ConsumerGroupHandler) onDelivery(track telegraf.DeliveryInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()

	msg, ok := h.undelivered[track.ID()]
	if !ok {
		h.log.Errorf("Could not mark message delivered: %d", track.ID())
		return
	}

	if track.Delivered() {
		msg.session.MarkMessage(msg.message, "")
	}

	delete(h.undelivered, track.ID())
	<-h.sem
}

// Reserve blocks until there is an available slot for a new message.
func (h *ConsumerGroupHandler) Reserve(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case h.sem <- empty{}:
		return nil
	}
}

func (h *ConsumerGroupHandler) release() {
	<-h.sem
}

// Handle processes a message and if successful saves it to be acknowledged
// after delivery.
func (h *ConsumerGroupHandler) Handle(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	if h.MaxMessageLen != 0 && len(msg.Value) > h.MaxMessageLen {
		session.MarkMessage(msg, "")
		h.release()
		return fmt.Errorf("message exceeds max_message_len (actual %d, max %d)",
			len(msg.Value), h.MaxMessageLen)
	}

	metrics, err := h.parser.Parse(msg.Value)
	if err != nil {
		h.release()
		return err
	}

	if len(h.TopicTag) > 0 {
		for _, metric := range metrics {
			metric.AddTag(h.TopicTag, msg.Topic)
		}
	}

	h.mu.Lock()
	id := h.acc.AddTrackingMetricGroup(metrics)
	h.undelivered[id] = Message{session: session, message: msg}
	h.mu.Unlock()
	return nil
}

// ConsumeClaim is called once each claim in a goroutine and must be
// thread-safe.  Should run until the claim is closed.
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()

	for {
		err := h.Reserve(ctx)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			err := h.Handle(session, msg)
			if err != nil {
				h.acc.AddError(err)
			}
		}
	}
}

// Cleanup stops the internal goroutine and is called after all ConsumeClaim
// functions have completed.
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.cancel()
	h.wg.Wait()
	return nil
}

func init() {
	inputs.Add("kafka_consumer", func() telegraf.Input {
		return &KafkaConsumer{}
	})
}
