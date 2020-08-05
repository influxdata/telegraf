package kafka_avro_consumer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/avro"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/linkedin/goavro"
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

  ## Optional SASL Config
  # sasl_username = "kafka"
  # sasl_password = "secret"

  ## Name of the consumer group.
  # consumer_group = "telegraf_metrics_consumers"

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

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Avro SSL client auth
  # schema_registry_certificate = "/tmp/client.crt"
  # schema_registry_key = "/tmp/client.key"
  # schema_registry_ca = "/tmp/ca.pem"
	# schema_registry = "https://my-kafka-schema-registry.example.com/subjects/LmaEventSchema/versions/1"
	## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
  ##  Envelpe fields passed as metadata field
  # meta_pass = ["source", "payload_type", "datacenter"]
`

const (
	defaultMaxUndeliveredMessages = 1000
	defaultMaxMessageLen          = 1000000
	defaultConsumerGroup          = "telegraf_metrics_consumers"
)

type empty struct{}
type semaphore chan empty

type KafkaConsumer struct {
	Brokers                []string `toml:"brokers"`
	ClientID               string   `toml:"client_id"`
	ConsumerGroup          string   `toml:"consumer_group"`
	MaxMessageLen          int      `toml:"max_message_len"`
	MaxUndeliveredMessages int      `toml:"max_undelivered_messages"`
	Offset                 string   `toml:"offset"`
	BalanceStrategy        string   `toml:"balance_strategy"`
	Topics                 []string `toml:"topics"`
	TopicTag               string   `toml:"topic_tag"`
	Version                string   `toml:"version"`
	SASLPassword           string   `toml:"sasl_password"`
	SASLUsername           string   `toml:"sasl_username"`

	// Avro Schema
	avroSchema string
	avroCodec  *goavro.Codec

	// confluent Avro Registry
	AvroMagicByteRequired bool `toml:"avro_magic_byte_required"`

	// Avro schema registry
	SchemaRegistry string   `toml:"schema_registry"`
	MetaPass       []string `toml:"meta_pass"`

	// Additional avro fields
	Loglevel      string `toml:"loglevel"`
	Logtype       string `toml:"type"`
	Servicelevel  string `toml:"servicelevel"`
	PayloadFormat string `toml:"payload_format"`
	PayloadType   string `toml:"payload_type"`
	Source        string `toml:"source"`
	DataCenter    string `toml:"datacenter"`

	tls.ClientConfig

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
	Create(brokers []string, group string, config *sarama.Config) (ConsumerGroup, error)
}

type SaramaCreator struct{}

func (*SaramaCreator) Create(brokers []string, group string, config *sarama.Config) (ConsumerGroup, error) {
	return sarama.NewConsumerGroup(brokers, group, config)
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
	if k.ConsumerGroup == "" {
		k.ConsumerGroup = defaultConsumerGroup
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Kafka version 0.10.2.0 is required for consumer groups.
	config.Version = sarama.V0_10_2_0

	if k.Version != "" {
		version, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return err
		}

		config.Version = version
	}

	tlsConfig, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	if k.SASLUsername != "" && k.SASLPassword != "" {
		config.Net.SASL.User = k.SASLUsername
		config.Net.SASL.Password = k.SASLPassword
		config.Net.SASL.Enable = true
	}

	if k.ClientID != "" {
		config.ClientID = k.ClientID
	} else {
		config.ClientID = "Telegraf"
	}

	switch strings.ToLower(k.Offset) {
	case "oldest", "":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		return fmt.Errorf("invalid offset %q", k.Offset)
	}

	switch strings.ToLower(k.BalanceStrategy) {
	case "range", "":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	default:
		return fmt.Errorf("invalid balance strategy %q", k.BalanceStrategy)
	}

	if k.ConsumerCreator == nil {
		k.ConsumerCreator = &SaramaCreator{}
	}

	k.config = config
	return nil
}

func (k *KafkaConsumer) Start(acc telegraf.Accumulator) error {
	var err error
	k.avroSchema, _ = avro.GetSchema(k.SchemaRegistry)
	codec, err := goavro.NewCodec(k.avroSchema)
	if err != nil {
		fmt.Println(err)
	}
	k.avroCodec = codec

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
			handler := NewConsumerGroupHandler(acc, k.MaxUndeliveredMessages, k.parser, k)
			handler.MaxMessageLen = k.MaxMessageLen
			handler.TopicTag = k.TopicTag
			err := k.consumer.Consume(ctx, k.Topics, handler)
			if err != nil {
				acc.AddError(err)
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

func (k *KafkaConsumer) Gather(acc telegraf.Accumulator) error {
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

func NewConsumerGroupHandler(acc telegraf.Accumulator, maxUndelivered int, parser parsers.Parser, k *KafkaConsumer) *ConsumerGroupHandler {
	handler := &ConsumerGroupHandler{
		acc:         acc.WithTracking(maxUndelivered),
		sem:         make(chan empty, maxUndelivered),
		undelivered: make(map[telegraf.TrackingID]Message, maxUndelivered),
		parser:      parser,
		k:           k,
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

	k *KafkaConsumer

	mu          sync.Mutex
	undelivered map[telegraf.TrackingID]Message
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
func (h *ConsumerGroupHandler) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
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
		log.Printf("E! [inputs.kafka_avro_consumer] Could not mark message delivered: %d", track.ID())
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
	var payload string
	metaFields := map[string]string{}
	for _, v := range h.k.MetaPass {
		metaFields[v] = ""
	}
	var binaryMsg []byte

	if h.k.AvroMagicByteRequired {
		pureMsg := []byte(msg.Value)
		binaryMsg = pureMsg[5:]
	} else {
		binaryMsg = msg.Value
	}

	native, _, err := h.k.avroCodec.NativeFromBinary(binaryMsg)
	if err != nil {
		h.acc.AddError(fmt.Errorf("Message Avro Decode Error\nmessage: %s\nerror: %s",
			string(msg.Value), err.Error()))
	}
	if native != nil {
		if avroPayload, ok := native.(map[string]interface{})["payload"]; ok {
			if val, ok := avroPayload.(map[string]interface{})["string"]; ok {
				payload = val.(string)
			}
		}
	}
	metrics, err := h.parser.Parse([]byte(payload))
	if err != nil {
		h.release()
		return err
	}

	for _, metric := range metrics {
		for k, v := range native.(map[string]interface{}) {
			if v != nil && k != "payload" {
				if _, ok := metaFields[k]; ok {
					envelopeField := v.(map[string]interface{})["string"]
					if envelopeField != nil {
						metric.AddTag(fmt.Sprintf("meta_%s", k), envelopeField.(string))
					}
				}
			}
		}
		h.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		if len(h.TopicTag) > 0 {
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
			return nil
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
	inputs.Add("kafka_avro_consumer", func() telegraf.Input {
		return &KafkaConsumer{}
	})
}
