package mqtt_consumer

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/eclipse/paho.mqtt.golang"
)

type MQTTConsumer struct {
	Servers  []string
	Topics   []string
	Username string
	Password string
	QoS      int `toml:"qos"`

	parser parsers.Parser

	// Legacy metric buffer support
	MetricBuffer int

	PersistentSession bool
	ClientID          string `toml:"client_id"`

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	sync.Mutex
	client mqtt.Client
	// channel of all incoming raw mqtt messages
	in   chan mqtt.Message
	done chan struct{}

	// keep the accumulator internally:
	acc telegraf.Accumulator

	started bool
}

var sampleConfig = `
  servers = ["localhost:1883"]
  ## MQTT QoS, must be 0, 1, or 2
  qos = 0

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

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
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
	m.started = false

	if m.PersistentSession && m.ClientID == "" {
		return fmt.Errorf("ERROR MQTT Consumer: When using persistent_session" +
			" = true, you MUST also set client_id")
	}

	m.acc = acc
	if m.QoS > 2 || m.QoS < 0 {
		return fmt.Errorf("MQTT Consumer, invalid QoS value: %d", m.QoS)
	}

	opts, err := m.createOpts()
	if err != nil {
		return err
	}

	m.client = mqtt.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	m.in = make(chan mqtt.Message, 1000)
	m.done = make(chan struct{})

	go m.receiver()

	return nil
}
func (m *MQTTConsumer) onConnect(c mqtt.Client) {
	log.Printf("I! MQTT Client Connected")
	if !m.PersistentSession || !m.started {
		topics := make(map[string]byte)
		for _, topic := range m.Topics {
			topics[topic] = byte(m.QoS)
		}
		subscribeToken := c.SubscribeMultiple(topics, m.recvMessage)
		subscribeToken.Wait()
		if subscribeToken.Error() != nil {
			log.Printf("E! MQTT Subscribe Error\ntopics: %s\nerror: %s",
				strings.Join(m.Topics[:], ","), subscribeToken.Error())
		}
		m.started = true
	}
	return
}

func (m *MQTTConsumer) onConnectionLost(c mqtt.Client, err error) {
	log.Printf("E! MQTT Connection lost\nerror: %s\nMQTT Client will try to reconnect", err.Error())
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
				log.Printf("E! MQTT Parse Error\nmessage: %s\nerror: %s",
					string(msg.Payload()), err.Error())
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
	close(m.done)
	m.client.Disconnect(200)
	m.started = false
}

func (m *MQTTConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (m *MQTTConsumer) createOpts() (*mqtt.ClientOptions, error) {
	opts := mqtt.NewClientOptions()

	if m.ClientID == "" {
		opts.SetClientID("Telegraf-Consumer-" + internal.RandomString(5))
	} else {
		opts.SetClientID(m.ClientID)
	}

	tlsCfg, err := internal.GetTLSConfig(
		m.SSLCert, m.SSLKey, m.SSLCA, m.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	scheme := "tcp"
	if tlsCfg != nil {
		scheme = "ssl"
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
	for _, host := range m.Servers {
		server := fmt.Sprintf("%s://%s", scheme, host)

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
		return &MQTTConsumer{}
	})
}
