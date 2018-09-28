package mqtt_consumer

import (
	"fmt"
	"log"
	"strings"
	"sync"
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

	sync.Mutex
	client mqtt.Client
	// channel of all incoming raw mqtt messages
	in   chan mqtt.Message
	done chan struct{}

	// keep the accumulator internally:
	acc telegraf.Accumulator

	connected bool
}

var sampleConfig = `
  ## MQTT broker URLs to be used. The format should be scheme://host:port,
  ## schema can be tcp, ssl, or ws.
  servers = ["tcp://localhost:1883"]

  ## MQTT QoS, must be 0, 1, or 2
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
	m.Lock()
	defer m.Unlock()
	m.connected = false

	if m.PersistentSession && m.ClientID == "" {
		return fmt.Errorf("ERROR MQTT Consumer: When using persistent_session" +
			" = true, you MUST also set client_id")
	}

	m.acc = acc
	if m.QoS > 2 || m.QoS < 0 {
		return fmt.Errorf("MQTT Consumer, invalid QoS value: %d", m.QoS)
	}

	if m.ConnectionTimeout.Duration < 1*time.Second {
		return fmt.Errorf("MQTT Consumer, invalid connection_timeout value: %s", m.ConnectionTimeout.Duration)
	}

	opts, err := m.createOpts()
	if err != nil {
		return err
	}

	m.client = mqtt.NewClient(opts)
	m.in = make(chan mqtt.Message, 1000)
	m.done = make(chan struct{})

	m.connect()

	return nil
}

func (m *MQTTConsumer) connect() error {
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()
		log.Printf("D! MQTT Consumer, connection error - %v", err)

		return err
	}

	go m.receiver()

	return nil
}

func (m *MQTTConsumer) onConnect(c mqtt.Client) {
	log.Printf("I! MQTT Client Connected")
	if !m.PersistentSession || !m.connected {
		topics := make(map[string]byte)
		for _, topic := range m.Topics {
			topics[topic] = byte(m.QoS)
		}
		subscribeToken := c.SubscribeMultiple(topics, m.recvMessage)
		subscribeToken.Wait()
		if subscribeToken.Error() != nil {
			m.acc.AddError(fmt.Errorf("E! MQTT Subscribe Error\ntopics: %s\nerror: %s",
				strings.Join(m.Topics[:], ","), subscribeToken.Error()))
		}
		m.connected = true
	}
	return
}

func (m *MQTTConsumer) onConnectionLost(c mqtt.Client, err error) {
	m.acc.AddError(fmt.Errorf("E! MQTT Connection lost\nerror: %s\nMQTT Client will try to reconnect", err.Error()))
	return
}

// receiver() reads all incoming messages from the consumer, and parses them into
// influxdb metric points.
func (m *MQTTConsumer) receiver() {
	for {
		select {
		case <-m.done:
			return
		case msg := <-m.in:
			topic := msg.Topic()
			metrics, err := m.parser.Parse(msg.Payload())
			if err != nil {
				m.acc.AddError(fmt.Errorf("E! MQTT Parse Error\nmessage: %s\nerror: %s",
					string(msg.Payload()), err.Error()))
			}

			for _, metric := range metrics {
				tags := metric.Tags()
				tags["topic"] = topic
				m.acc.AddFields(metric.Name(), metric.Fields(), tags, metric.Time())
			}
		}
	}
}

func (m *MQTTConsumer) recvMessage(_ mqtt.Client, msg mqtt.Message) {
	m.in <- msg
}

func (m *MQTTConsumer) Stop() {
	m.Lock()
	defer m.Unlock()

	if m.connected {
		close(m.done)
		m.client.Disconnect(200)
		m.connected = false
	}
}

func (m *MQTTConsumer) Gather(acc telegraf.Accumulator) error {
	if !m.connected {
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
			log.Printf("W! mqtt_consumer server %q should be updated to use `scheme://host:port` format", server)
			if tlsCfg == nil {
				server = "tcp://" + server
			} else {
				server = "ssl://" + server
			}
		}

		opts.AddBroker(server)
	}
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetCleanSession(!m.PersistentSession)
	opts.SetOnConnectHandler(m.onConnect)
	opts.SetConnectionLostHandler(m.onConnectionLost)

	return opts, nil
}

func init() {
	inputs.Add("mqtt_consumer", func() telegraf.Input {
		return &MQTTConsumer{
			ConnectionTimeout: defaultConnectionTimeout,
		}
	})
}
