package mqtt_consumer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

var (
	// 30 Seconds is the default used by paho.mqtt.golang
	defaultConnectionTimeout = internal.Duration{Duration: 30 * time.Second}

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

type MQTTConsumer struct {
	Servers                []string          `toml:"servers"`
	Topics                 []string          `toml:"topics"`
	TopicTag               *string           `toml:"topic_tag"`
	Username               string            `toml:"username"`
	Password               string            `toml:"password"`
	QoS                    int               `toml:"qos"`
	ConnectionTimeout      internal.Duration `toml:"connection_timeout"`
	MaxUndeliveredMessages int               `toml:"max_undelivered_messages"`

	parser parsers.Parser

	// Legacy metric buffer support; deprecated in v0.10.3
	MetricBuffer int

	PersistentSession bool
	ClientID          string `toml:"client_id"`
	tls.ClientConfig

	Log telegraf.Logger

	clientFactory ClientFactory
	client        Client
	opts          *mqtt.ClientOptions
	acc           telegraf.TrackingAccumulator
	state         ConnectionState
	sem           semaphore
	messages      map[telegraf.TrackingID]bool
	topicTag      string

	ctx    context.Context
	cancel context.CancelFunc
}

var sampleConfig = `
  ## MQTT broker URLs to be used. The format should be scheme://host:port,
  ## schema can be tcp, ssl, or ws.
  servers = ["tcp://127.0.0.1:1883"]

  ## Topics that will be subscribed to.
  topics = [
    "telegraf/host01/cpu",
    "telegraf/+/mem",
    "sensors/#",
  ]

  ## The message topic will be stored in a tag specified by this value.  If set
  ## to the empty string no topic tag will be created.
  # topic_tag = "topic"

  ## QoS policy for messages
  ##   0 = at most once
  ##   1 = at least once
  ##   2 = exactly once
  ##
  ## When using a QoS of 1 or 2, you should enable persistent_session to allow
  ## resuming unacknowledged messages.
  # qos = 0

  ## Connection timeout for initial connection in seconds
  # connection_timeout = "30s"

  ## Maximum messages to read from the broker that have not been written by an
  ## output.  For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Persistent session disables clearing of the client session on connection.
  ## In order for this option to work you must also set client_id to identify
  ## the client.  To receive messages that arrived while the client is offline,
  ## also set the qos option to 1 or 2 and don't forget to also set the QoS when
  ## publishing.
  # persistent_session = false

  ## If unset, a random client ID will be generated.
  # client_id = ""

  ## Username and password to connect MQTT server.
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (m *MQTTConsumer) SampleConfig() string {
	return sampleConfig
}

func (m *MQTTConsumer) Description() string {
	return "Read metrics from MQTT topic(s)"
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

	if m.ConnectionTimeout.Duration < 1*time.Second {
		return fmt.Errorf("connection_timeout must be greater than 1s: %s", m.ConnectionTimeout.Duration)
	}

	m.topicTag = "topic"
	if m.TopicTag != nil {
		m.topicTag = *m.TopicTag
	}

	opts, err := m.createOpts()
	if err != nil {
		return err
	}

	m.opts = opts

	return nil
}

func (m *MQTTConsumer) Start(acc telegraf.Accumulator) error {
	m.state = Disconnected

	m.acc = acc.WithTracking(m.MaxUndeliveredMessages)
	m.sem = make(semaphore, m.MaxUndeliveredMessages)
	m.ctx, m.cancel = context.WithCancel(context.Background())

	m.client = m.clientFactory(m.opts)

	// AddRoute sets up the function for handling messages.  These need to be
	// added in case we find a persistent session containing subscriptions so we
	// know where to dispatch persisted and new messages to.  In the alternate
	// case that we need to create the subscriptions these will be replaced.
	for _, topic := range m.Topics {
		m.client.AddRoute(topic, m.recvMessage)
	}

	m.state = Connecting
	m.connect()

	return nil
}

func (m *MQTTConsumer) connect() error {
	token := m.client.Connect()
	if token.Wait() && token.Error() != nil {
		err := token.Error()
		m.state = Disconnected
		return err
	}

	m.Log.Infof("Connected %v", m.Servers)
	m.state = Connected
	m.messages = make(map[telegraf.TrackingID]bool)

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
		m.acc.AddError(fmt.Errorf("subscription error: topics: %s: %v",
			strings.Join(m.Topics[:], ","), subscribeToken.Error()))
	}

	return nil
}

func (m *MQTTConsumer) onConnectionLost(c mqtt.Client, err error) {
	m.acc.AddError(fmt.Errorf("connection lost: %v", err))
	m.Log.Debugf("Disconnected %v", m.Servers)
	m.state = Disconnected
	return
}

func (m *MQTTConsumer) recvMessage(c mqtt.Client, msg mqtt.Message) {
	for {
		select {
		case track := <-m.acc.Delivered():
			<-m.sem
			_, ok := m.messages[track.ID()]
			if !ok {
				// Added by a previous connection
				continue
			}
			// No ack, MQTT does not support durable handling
			delete(m.messages, track.ID())
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

func (m *MQTTConsumer) onMessage(acc telegraf.TrackingAccumulator, msg mqtt.Message) error {
	metrics, err := m.parser.Parse(msg.Payload())
	if err != nil {
		return err
	}

	if m.topicTag != "" {
		topic := msg.Topic()
		for _, metric := range metrics {
			metric.AddTag(m.topicTag, topic)
		}
	}

	id := acc.AddTrackingMetricGroup(metrics)
	m.messages[id] = true
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

func (m *MQTTConsumer) Gather(acc telegraf.Accumulator) error {
	if m.state == Disconnected {
		m.state = Connecting
		m.Log.Debugf("Connecting %v", m.Servers)
		m.connect()
	}

	return nil
}

func (m *MQTTConsumer) createOpts() (*mqtt.ClientOptions, error) {
	opts := mqtt.NewClientOptions()

	opts.ConnectTimeout = m.ConnectionTimeout.Duration

	if m.ClientID == "" {
		opts.SetClientID("Telegraf-Consumer-" + internal.RandomString(5))
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

	user := m.Username
	if user != "" {
		opts.SetUsername(user)
	}
	password := m.Password
	if password != "" {
		opts.SetPassword(password)
	}

	if len(m.Servers) == 0 {
		return opts, fmt.Errorf("could not get host informations")
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
