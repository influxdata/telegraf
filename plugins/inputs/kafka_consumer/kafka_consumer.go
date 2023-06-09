//go:generate ../../../tools/readme_config_includer/generator
package kafka_consumer

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

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
	TopicRegexps           []string        `toml:"topic_regexps"`
	TopicTag               string          `toml:"topic_tag"`
	ConsumerFetchDefault   config.Size     `toml:"consumer_fetch_default"`
	ConnectionStrategy     string          `toml:"connection_strategy"`

	kafka.ReadConfig

	kafka.Logger

	Log telegraf.Logger `toml:"-"`

	ConsumerCreator ConsumerGroupCreator `toml:"-"`
	consumer        ConsumerGroup
	config          *sarama.Config

	topicClient     sarama.Client
	regexps         []regexp.Regexp
	allWantedTopics []string
	ticker          *time.Ticker
	fingerprint     string

	parser    telegraf.Parser
	topicLock sync.Mutex
	wg        sync.WaitGroup
	cancel    context.CancelFunc
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

func (*KafkaConsumer) SampleConfig() string {
	return sampleConfig
}

func (k *KafkaConsumer) SetParser(parser telegraf.Parser) {
	k.parser = parser
}

func (k *KafkaConsumer) Init() error {
	k.SetLogger()

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

	if err := k.SetConfig(cfg, k.Log); err != nil {
		return fmt.Errorf("SetConfig: %w", err)
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
		cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	case "roundrobin":
		cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	case "sticky":
		cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	default:
		return fmt.Errorf("invalid balance strategy %q", k.BalanceStrategy)
	}

	if k.ConsumerCreator == nil {
		k.ConsumerCreator = &SaramaCreator{}
	}

	cfg.Consumer.MaxProcessingTime = time.Duration(k.MaxProcessingTime)

	if k.ConsumerFetchDefault != 0 {
		cfg.Consumer.Fetch.Default = int32(k.ConsumerFetchDefault)
	}

	switch strings.ToLower(k.ConnectionStrategy) {
	default:
		return fmt.Errorf("invalid connection strategy %q", k.ConnectionStrategy)
	case "defer", "startup", "":
	}

	k.config = cfg

	if len(k.TopicRegexps) == 0 {
		k.allWantedTopics = k.Topics
	} else {
		if err := k.compileTopicRegexps(); err != nil {
			return err
		}
		// We have regexps, so we're going to need a client to ask
		// the broker for topics
		client, err := sarama.NewClient(k.Brokers, k.config)
		if err != nil {
			return err
		}
		k.topicClient = client
	}

	return nil
}

func (k *KafkaConsumer) compileTopicRegexps() error {
	// While we can add new topics matching extant regexps, we can't
	// update that list on the fly.  We compile them once at startup.
	// Changing them is a configuration change and requires a restart.

	k.regexps = make([]regexp.Regexp, 0, len(k.TopicRegexps))
	for _, r := range k.TopicRegexps {
		re, err := regexp.Compile(r)
		if err != nil {
			return fmt.Errorf("regular expression %q did not compile: '%w", r, err)
		}
		k.regexps = append(k.regexps, *re)
	}
	return nil
}

func (k *KafkaConsumer) refreshTopics() error {
	// We have instantiated a new generic Kafka client, so we can ask
	// it for all the topics it knows about.  Then we build
	// regexps from our strings, loop over those, loop over the
	// topics, and if we find a match, add that topic to
	// out topic set, which then we turn back into a list at the end.

	if len(k.regexps) == 0 {
		return nil
	}

	allDiscoveredTopics, err := k.topicClient.Topics()
	if err != nil {
		return err
	}
	k.Log.Debugf("discovered topics: %v", allDiscoveredTopics)

	extantTopicSet := make(map[string]bool, len(allDiscoveredTopics))
	for _, t := range allDiscoveredTopics {
		extantTopicSet[t] = true
	}
	// Even if a topic specified by a literal string (that is, k.Topics)
	// does not appear in the topic list, we want to keep it around, in
	// case it pops back up--it is not guaranteed to be matched by any
	// of our regular expressions.  Therefore, we pretend that it's in
	// extantTopicSet, even if it isn't.
	//
	// Assuming that literally-specified topics are usually in the topics
	// present on the broker, this should not need a resizing (although if
	// you have many topics that you don't care about, it will be too big)
	wantedTopicSet := make(map[string]bool, len(allDiscoveredTopics))
	for _, t := range k.Topics {
		// Get our pre-specified topics
		k.Log.Debugf("adding literally-specified topic %s", t)
		wantedTopicSet[t] = true
	}
	for _, t := range allDiscoveredTopics {
		// Add topics that match regexps
		for _, r := range k.regexps {
			if r.MatchString(t) {
				wantedTopicSet[t] = true
				k.Log.Debugf("adding regexp-matched topic %q", t)
				break
			}
		}
	}
	topicList := make([]string, 0, len(wantedTopicSet))
	for t := range wantedTopicSet {
		topicList = append(topicList, t)
	}
	sort.Strings(topicList)
	fingerprint := strings.Join(topicList, ";")
	if fingerprint != k.fingerprint {
		k.Log.Infof("updating topics: replacing %q with %q", k.allWantedTopics, topicList)
	}
	k.topicLock.Lock()
	k.fingerprint = fingerprint
	k.allWantedTopics = topicList
	k.topicLock.Unlock()
	return nil
}

func (k *KafkaConsumer) create() error {
	var err error
	k.consumer, err = k.ConsumerCreator.Create(
		k.Brokers,
		k.ConsumerGroup,
		k.config,
	)

	return err
}

func (k *KafkaConsumer) startErrorAdder(acc telegraf.Accumulator) {
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		for err := range k.consumer.Errors() {
			acc.AddError(fmt.Errorf("channel: %w", err))
		}
	}()
}

func (k *KafkaConsumer) Start(acc telegraf.Accumulator) error {
	var err error

	// If TopicRegexps is set, add matches to Topics
	if len(k.TopicRegexps) > 0 {
		if err := k.refreshTopics(); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	k.cancel = cancel

	if k.ConnectionStrategy != "defer" {
		err = k.create()
		if err != nil {
			return fmt.Errorf("create consumer: %w", err)
		}
		k.startErrorAdder(acc)
	}

	// Start consumer goroutine
	k.wg.Add(1)
	go func() {
		var err error
		defer k.wg.Done()

		if k.consumer == nil {
			err = k.create()
			if err != nil {
				acc.AddError(fmt.Errorf("create consumer async: %w", err))
				return
			}
		}

		k.startErrorAdder(acc)

		for ctx.Err() == nil {
			handler := NewConsumerGroupHandler(acc, k.MaxUndeliveredMessages, k.parser, k.Log)
			handler.MaxMessageLen = k.MaxMessageLen
			handler.TopicTag = k.TopicTag
			// We need to copy allWantedTopics; the Consume() is
			// long-running and we can easily deadlock if our
			// topic-update-checker fires.
			topics := make([]string, len(k.allWantedTopics))
			k.topicLock.Lock()
			copy(topics, k.allWantedTopics)
			k.topicLock.Unlock()
			err := k.consumer.Consume(ctx, topics, handler)
			if err != nil {
				acc.AddError(fmt.Errorf("consume: %w", err))
				internal.SleepContext(ctx, reconnectDelay) //nolint:errcheck // ignore returned error as we cannot do anything about it anyway
			}
		}
		err = k.consumer.Close()
		if err != nil {
			acc.AddError(fmt.Errorf("close: %w", err))
		}
	}()

	return nil
}

func (k *KafkaConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (k *KafkaConsumer) Stop() {
	if k.ticker != nil {
		k.ticker.Stop()
	}
	// Lock so that a topic refresh cannot start while we are stopping.
	k.topicLock.Lock()
	defer k.topicLock.Unlock()
	if k.topicClient != nil {
		k.topicClient.Close()
	}
	k.cancel()
	k.wg.Wait()
}

// Message is an aggregate type binding the Kafka message and the session so
// that offsets can be updated.
type Message struct {
	message *sarama.ConsumerMessage
	session sarama.ConsumerGroupSession
}

func NewConsumerGroupHandler(acc telegraf.Accumulator, maxUndelivered int, parser telegraf.Parser, log telegraf.Logger) *ConsumerGroupHandler {
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
	parser telegraf.Parser
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
