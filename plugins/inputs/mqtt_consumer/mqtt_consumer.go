//go:generate ../../../tools/readme_config_includer/generator
package mqtt_consumer

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

var (
	// 30 Seconds is the default used by paho.mqtt.golang
	defaultConnectionTimeout      = config.Duration(30 * time.Second)
	defaultMaxUndeliveredMessages = 1000
)

type ConnectionState int
type empty struct{}
type semaphore chan empty

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
)

type Client interface {
	Connect() mqtt.Token
	SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token
	AddRoute(topic string, callback mqtt.MessageHandler)
	Disconnect(quiesce uint)
}
type ClientFactory func(o *mqtt.ClientOptions) Client
type TopicParsingConfig struct {
	Topic       string            `toml:"topic"`
	Measurement string            `toml:"measurement"`
	Tags        string            `toml:"tags"`
	Fields      string            `toml:"fields"`
	FieldTypes  map[string]string `toml:"types"`
	// cached split of user given information
	MeasurementIndex int
	SplitTags        []string
	SplitFields      []string
	SplitTopic       []string
}
type MQTTConsumer struct {
	Servers                []string             `toml:"servers"`
	Topics                 []string             `toml:"topics"`
	TopicTag               *string              `toml:"topic_tag"`
	TopicParsing           []TopicParsingConfig `toml:"topic_parsing"`
	Username               config.Secret        `toml:"username"`
	Password               config.Secret        `toml:"password"`
	QoS                    int                  `toml:"qos"`
	ConnectionTimeout      config.Duration      `toml:"connection_timeout"`
	MaxUndeliveredMessages int                  `toml:"max_undelivered_messages"`
	parser                 parsers.Parser

	MetricBuffer      int `toml:"metric_buffer" deprecated:"0.10.3;2.0.0;option is ignored"`
	PersistentSession bool
	ClientID          string `toml:"client_id"`

	tls.ClientConfig

	Log           telegraf.Logger
	clientFactory ClientFactory
	client        Client
	opts          *mqtt.ClientOptions
	acc           telegraf.TrackingAccumulator
	state         ConnectionState
	sem           semaphore
	messages      map[telegraf.TrackingID]bool
	messagesMutex sync.Mutex
	topicTagParse string
	ctx           context.Context
	cancel        context.CancelFunc
	payloadSize   selfstat.Stat
	messagesRecv  selfstat.Stat
}

func (*MQTTConsumer) SampleConfig() string {
	return sampleConfig
}

func (m *MQTTConsumer) SetParser(parser parsers.Parser) {
	m.parser = parser
}
func (m *MQTTConsumer) Init() error {
	m.state = Disconnected
	if m.PersistentSession && m.ClientID == "" {
		return errors.New("persistent_session requires client_id")
	}
	if m.QoS > 2 || m.QoS < 0 {
		return fmt.Errorf("qos value must be 0, 1, or 2: %d", m.QoS)
	}
	if time.Duration(m.ConnectionTimeout) < 1*time.Second {
		return fmt.Errorf("connection_timeout must be greater than 1s: %s", time.Duration(m.ConnectionTimeout))
	}
	m.topicTagParse = "topic"
	if m.TopicTag != nil {
		m.topicTagParse = *m.TopicTag
	}
	opts, err := m.createOpts()
	if err != nil {
		return err
	}
	m.opts = opts
	m.messages = map[telegraf.TrackingID]bool{}

	for i, p := range m.TopicParsing {
		splitMeasurement := strings.Split(p.Measurement, "/")
		for j := range splitMeasurement {
			if splitMeasurement[j] != "_" && splitMeasurement[j] != "" {
				m.TopicParsing[i].MeasurementIndex = j
				break
			}
		}
		m.TopicParsing[i].SplitTags = strings.Split(p.Tags, "/")
		m.TopicParsing[i].SplitFields = strings.Split(p.Fields, "/")
		m.TopicParsing[i].SplitTopic = strings.Split(p.Topic, "/")

		if len(splitMeasurement) != len(m.TopicParsing[i].SplitTopic) && len(splitMeasurement) != 1 {
			return fmt.Errorf("config error topic parsing: measurement length does not equal topic length")
		}

		if len(m.TopicParsing[i].SplitFields) != len(m.TopicParsing[i].SplitTopic) && p.Fields != "" {
			return fmt.Errorf("config error topic parsing: fields length does not equal topic length")
		}

		if len(m.TopicParsing[i].SplitTags) != len(m.TopicParsing[i].SplitTopic) && p.Tags != "" {
			return fmt.Errorf("config error topic parsing: tags length does not equal topic length")
		}
	}

	m.payloadSize = selfstat.Register("mqtt_consumer", "payload_size", map[string]string{})
	m.messagesRecv = selfstat.Register("mqtt_consumer", "messages_received", map[string]string{})
	return nil
}
func (m *MQTTConsumer) Start(acc telegraf.Accumulator) error {
	m.state = Disconnected
	m.acc = acc.WithTracking(m.MaxUndeliveredMessages)
	m.sem = make(semaphore, m.MaxUndeliveredMessages)
	m.ctx, m.cancel = context.WithCancel(context.Background())
	return m.connect()
}
func (m *MQTTConsumer) connect() error {
	m.state = Connecting
	m.client = m.clientFactory(m.opts)
	// AddRoute sets up the function for handling messages.  These need to be
	// added in case we find a persistent session containing subscriptions so we
	// know where to dispatch persisted and new messages to.  In the alternate
	// case that we need to create the subscriptions these will be replaced.
	for _, topic := range m.Topics {
		m.client.AddRoute(topic, m.recvMessage)
	}
	token := m.client.Connect()
	if token.Wait() && token.Error() != nil {
		err := token.Error()
		m.state = Disconnected
		return err
	}
	m.Log.Infof("Connected %v", m.Servers)
	m.state = Connected
	// Persistent sessions should skip subscription if a session is present, as
	// the subscriptions are stored by the server.
	type sessionPresent interface {
		SessionPresent() bool
	}
	if t, ok := token.(sessionPresent); ok && t.SessionPresent() {
		m.Log.Debugf("Session found %v", m.Servers)
		return nil
	}
	topics := make(map[string]byte)
	for _, topic := range m.Topics {
		topics[topic] = byte(m.QoS)
	}
	subscribeToken := m.client.SubscribeMultiple(topics, m.recvMessage)
	subscribeToken.Wait()
	if subscribeToken.Error() != nil {
		m.acc.AddError(fmt.Errorf("subscription error: topics %q: %w", strings.Join(m.Topics[:], ","), subscribeToken.Error()))
	}
	return nil
}
func (m *MQTTConsumer) onConnectionLost(_ mqtt.Client, err error) {
	// Should already be disconnected, but make doubly sure
	m.client.Disconnect(5)
	m.acc.AddError(fmt.Errorf("connection lost: %w", err))
	m.Log.Debugf("Disconnected %v", m.Servers)
	m.state = Disconnected
}
func (m *MQTTConsumer) recvMessage(_ mqtt.Client, msg mqtt.Message) {
	for {
		// Drain anything that's been delivered
		select {
		case track := <-m.acc.Delivered():
			m.onDelivered(track)
			continue
		default:
		}

		// Wait for room to accumulate metric, but make delivery progress if possible
		// (Note that select will randomly pick a case if both are available)
		select {
		case track := <-m.acc.Delivered():
			m.onDelivered(track)
		case m.sem <- empty{}:
			err := m.onMessage(m.acc, msg)
			if err != nil {
				m.acc.AddError(err)
				<-m.sem
			}
			return
		}
	}
}

// compareTopics is used to support the mqtt wild card `+` which allows for one topic of any value
func compareTopics(expected []string, incoming []string) bool {
	if len(expected) != len(incoming) {
		return false
	}

	for i, expected := range expected {
		if incoming[i] != expected && expected != "+" {
			return false
		}
	}

	return true
}

func (m *MQTTConsumer) onDelivered(track telegraf.DeliveryInfo) {
	<-m.sem
	m.messagesMutex.Lock()
	_, ok := m.messages[track.ID()]
	if ok {
		// No ack, MQTT does not support durable handling
		delete(m.messages, track.ID())
	}
	m.messagesMutex.Unlock()
}

func (m *MQTTConsumer) onMessage(acc telegraf.TrackingAccumulator, msg mqtt.Message) error {
	payloadBytes := len(msg.Payload())
	m.payloadSize.Incr(int64(payloadBytes))
	m.messagesRecv.Incr(1)

	metrics, err := m.parser.Parse(msg.Payload())
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		if m.topicTagParse != "" {
			metric.AddTag(m.topicTagParse, msg.Topic())
		}
		for _, p := range m.TopicParsing {
			values := strings.Split(msg.Topic(), "/")
			if !compareTopics(p.SplitTopic, values) {
				continue
			}

			if p.Measurement != "" {
				metric.SetName(values[p.MeasurementIndex])
			}
			if p.Tags != "" {
				err := parseMetric(p.SplitTags, values, p.FieldTypes, true, metric)
				if err != nil {
					return err
				}
			}
			if p.Fields != "" {
				err := parseMetric(p.SplitFields, values, p.FieldTypes, false, metric)
				if err != nil {
					return err
				}
			}
		}
	}
	id := acc.AddTrackingMetricGroup(metrics)
	m.messagesMutex.Lock()
	m.messages[id] = true
	m.messagesMutex.Unlock()
	return nil
}
func (m *MQTTConsumer) Stop() {
	if m.state == Connected {
		m.Log.Debugf("Disconnecting %v", m.Servers)
		m.client.Disconnect(200)
		m.Log.Debugf("Disconnected %v", m.Servers)
		m.state = Disconnected
	}
	m.cancel()
}
func (m *MQTTConsumer) Gather(_ telegraf.Accumulator) error {
	if m.state == Disconnected {
		m.Log.Debugf("Connecting %v", m.Servers)
		return m.connect()
	}
	return nil
}
func (m *MQTTConsumer) createOpts() (*mqtt.ClientOptions, error) {
	opts := mqtt.NewClientOptions()
	opts.ConnectTimeout = time.Duration(m.ConnectionTimeout)
	if m.ClientID == "" {
		randomString, err := internal.RandomString(5)
		if err != nil {
			return nil, fmt.Errorf("generating random string for client ID failed: %w", err)
		}
		opts.SetClientID("Telegraf-Consumer-" + randomString)
	} else {
		opts.SetClientID(m.ClientID)
	}
	tlsCfg, err := m.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		opts.SetTLSConfig(tlsCfg)
	}
	if !m.Username.Empty() {
		user, err := m.Username.Get()
		if err != nil {
			return nil, fmt.Errorf("getting username failed: %w", err)
		}
		opts.SetUsername(string(user))
		config.ReleaseSecret(user)
	}

	if !m.Password.Empty() {
		password, err := m.Password.Get()
		if err != nil {
			return nil, fmt.Errorf("getting password failed: %w", err)
		}
		opts.SetPassword(string(password))
		config.ReleaseSecret(password)
	}
	if len(m.Servers) == 0 {
		return opts, fmt.Errorf("could not get host information")
	}
	for _, server := range m.Servers {
		// Preserve support for host:port style servers; deprecated in Telegraf 1.4.4
		if !strings.Contains(server, "://") {
			m.Log.Warnf("Server %q should be updated to use `scheme://host:port` format", server)
			if tlsCfg == nil {
				server = "tcp://" + server
			} else {
				server = "ssl://" + server
			}
		}
		opts.AddBroker(server)
	}
	opts.SetAutoReconnect(false)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetCleanSession(!m.PersistentSession)
	opts.SetConnectionLostHandler(m.onConnectionLost)
	return opts, nil
}

// parseFields gets multiple fields from the topic based on the user configuration (TopicParsing.Fields)
func parseMetric(keys []string, values []string, types map[string]string, isTag bool, metric telegraf.Metric) error {
	for i, k := range keys {
		if k == "_" || k == "" {
			continue
		}

		if isTag {
			metric.AddTag(k, values[i])
		} else {
			newType, err := typeConvert(types, values[i], k)
			if err != nil {
				return err
			}
			metric.AddField(k, newType)
		}
	}
	return nil
}

func typeConvert(types map[string]string, topicValue string, key string) (interface{}, error) {
	var newType interface{}
	var err error
	// If the user configured inputs.mqtt_consumer.topic.types, check for the desired type
	if desiredType, ok := types[key]; ok {
		switch desiredType {
		case "uint":
			newType, err = strconv.ParseUint(topicValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to convert field %q to type uint: %w", topicValue, err)
			}
		case "int":
			newType, err = strconv.ParseInt(topicValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to convert field %q to type int: %w", topicValue, err)
			}
		case "float":
			newType, err = strconv.ParseFloat(topicValue, 64)
			if err != nil {
				return nil, fmt.Errorf("unable to convert field %q to type float: %w", topicValue, err)
			}
		default:
			return nil, fmt.Errorf("converting to the type %s is not supported: use int, uint, or float", desiredType)
		}
	} else {
		newType = topicValue
	}

	return newType, nil
}

func New(factory ClientFactory) *MQTTConsumer {
	return &MQTTConsumer{
		Servers:                []string{"tcp://127.0.0.1:1883"},
		ConnectionTimeout:      defaultConnectionTimeout,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		clientFactory:          factory,
		state:                  Disconnected,
	}
}
func init() {
	inputs.Add("mqtt_consumer", func() telegraf.Input {
		return New(func(o *mqtt.ClientOptions) Client {
			return mqtt.NewClient(o)
		})
	})
}
