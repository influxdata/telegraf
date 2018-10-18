package mqtt_consumer

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/eclipse/paho.mqtt.golang"
)

// 30 Seconds is the default used by paho.mqtt.golang
var defaultConnectionTimeout = internal.Duration{Duration: 30 * time.Second}

type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
)

type MQTTConsumer struct {
	Servers           []string
	Topics            []string
	Username          string
	Password          string
	QoS               int               `toml:"qos"`
	ConnectionTimeout internal.Duration `toml:"connection_timeout"`

	parser parsers.Parser

	// Legacy metric buffer support
	MetricBuffer int

	PersistentSession bool
	ClientID          string `toml:"client_id"`
	tls.ClientConfig

	client     mqtt.Client
	acc        telegraf.Accumulator
	state      ConnectionState
	subscribed bool
}

var sampleConfig = `
  ## MQTT broker URLs to be used. The format should be scheme://host:port,
  ## schema can be tcp, ssl, or ws.
  servers = ["tcp://localhost:1883"]

  ## QoS policy for messages
  ##   0 = at most once
  ##   1 = at least once
  ##   2 = exactly once
  ##
  ## When using a QoS of 1 or 2, you should enable persistent_session to allow
  ## resuming unacknowledged messages.
  qos = 0

  ## Connection timeout for initial connection in seconds
  connection_timeout = "30s"

  ## Topics to subscribe to
  topics = [
    "telegraf/host01/cpu",
    "telegraf/+/mem",
    "sensors/#",
  ]

  # if true, messages that can't be delivered while the subscriber is offline
  # will be delivered when it comes back (such as on service restart).
  # NOTE: if true, client_id MUST be set
  persistent_session = false
  # If empty, a random client ID will be generated.
  client_id = ""

  ## username and password to connect MQTT server.
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

func (m *MQTTConsumer) Start(acc telegraf.Accumulator) error {
	m.state = Disconnected

	if m.PersistentSession && m.ClientID == "" {
		return errors.New("persistent_session requires client_id")
	}

	m.acc = acc
	if m.QoS > 2 || m.QoS < 0 {
		return fmt.Errorf("qos value must be 0, 1, or 2: %d", m.QoS)
	}

	if m.ConnectionTimeout.Duration < 1*time.Second {
		return fmt.Errorf("connection_timeout must be greater than 1s: %s", m.ConnectionTimeout.Duration)
	}

	opts, err := m.createOpts()
	if err != nil {
		return err
	}

	m.client = mqtt.NewClient(opts)
	m.state = Connecting
	m.connect()

	return nil
}

func (m *MQTTConsumer) connect() error {
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()
		m.state = Disconnected
		return err
	}

	log.Printf("I! [inputs.mqtt_consumer]: connected %v", m.Servers)
	m.state = Connected

	// Only subscribe on first connection when using persistent sessions.  On
	// subsequent connections the subscriptions should be stored in the
	// session, but the proper way to do this is to check the connection
	// response to ensure a session was found.
	if !m.PersistentSession || !m.subscribed {
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
		m.subscribed = true
	}

	return nil
}

func (m *MQTTConsumer) onConnectionLost(c mqtt.Client, err error) {
	m.acc.AddError(fmt.Errorf("connection lost: %v", err))
	log.Printf("D! [inputs.mqtt_consumer]: disconnected %v", m.Servers)
	m.state = Disconnected
	return
}

func (m *MQTTConsumer) recvMessage(c mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	metrics, err := m.parser.Parse(msg.Payload())
	if err != nil {
		m.acc.AddError(err)
	}

	for _, metric := range metrics {
		tags := metric.Tags()
		tags["topic"] = topic
		m.acc.AddFields(metric.Name(), metric.Fields(), tags, metric.Time())
	}
}

func (m *MQTTConsumer) Stop() {
	if m.state == Connected {
		log.Printf("D! [inputs.mqtt_consumer]: disconnecting %v", m.Servers)
		m.client.Disconnect(200)
		log.Printf("D! [inputs.mqtt_consumer]: disconnected %v", m.Servers)
		m.state = Disconnected
	}
}

func (m *MQTTConsumer) Gather(acc telegraf.Accumulator) error {
	if m.state == Disconnected {
		m.state = Connecting
		log.Printf("D! [inputs.mqtt_consumer]: connecting %v", m.Servers)
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
		return opts, fmt.Errorf("could not get host infomations")
	}

	for _, server := range m.Servers {
		// Preserve support for host:port style servers; deprecated in Telegraf 1.4.4
		if !strings.Contains(server, "://") {
			log.Printf("W! [inputs.mqtt_consumer] server %q should be updated to use `scheme://host:port` format", server)
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

func init() {
	inputs.Add("mqtt_consumer", func() telegraf.Input {
		return &MQTTConsumer{
			ConnectionTimeout: defaultConnectionTimeout,
			state:             Disconnected,
		}
	})
}
