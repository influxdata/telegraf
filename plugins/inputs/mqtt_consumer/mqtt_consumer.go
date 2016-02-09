package mqtt_consumer

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

type MQTTConsumer struct {
	Servers      []string
	Topics       []string
	Username     string
	Password     string
	MetricBuffer int
	QoS          int `toml:"qos"`

	parser parsers.Parser

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	sync.Mutex
	client *mqtt.Client
	// channel for all incoming parsed mqtt metrics
	metricC chan telegraf.Metric
	// channel for the topics of all incoming metrics (for tagging metrics)
	topicC chan string
	// channel of all incoming raw mqtt messages
	in   chan mqtt.Message
	done chan struct{}
}

var sampleConfig = `
  servers = ["localhost:1883"]
  ### MQTT QoS, must be 0, 1, or 2
  qos = 0

  ### Topics to subscribe to
  topics = [
    "telegraf/host01/cpu",
    "telegraf/+/mem",
    "sensors/#",
  ]

  ### Maximum number of metrics to buffer between collection intervals
  metric_buffer = 100000

  ### username and password to connect MQTT server.
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"

  ### Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ### Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ### Data format to consume. This can be "json", "influx" or "graphite"
  ### Each data format has it's own unique set of configuration options, read
  ### more about them here:
  ### https://github.com/influxdata/telegraf/blob/master/DATA_FORMATS_INPUT.md
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

func (m *MQTTConsumer) Start() error {
	m.Lock()
	defer m.Unlock()
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

	m.in = make(chan mqtt.Message, m.MetricBuffer)
	m.done = make(chan struct{})
	if m.MetricBuffer == 0 {
		m.MetricBuffer = 100000
	}
	m.metricC = make(chan telegraf.Metric, m.MetricBuffer)
	m.topicC = make(chan string, m.MetricBuffer)

	topics := make(map[string]byte)
	for _, topic := range m.Topics {
		topics[topic] = byte(m.QoS)
	}
	subscribeToken := m.client.SubscribeMultiple(topics, m.recvMessage)
	subscribeToken.Wait()
	if subscribeToken.Error() != nil {
		return subscribeToken.Error()
	}

	go m.receiver()

	return nil
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
				log.Printf("MQTT PARSE ERROR\nmessage: %s\nerror: %s",
					string(msg.Payload()), err.Error())
			}

			for _, metric := range metrics {
				select {
				case m.metricC <- metric:
					m.topicC <- topic
				default:
					log.Printf("MQTT Consumer buffer is full, dropping a metric." +
						" You may want to increase the metric_buffer setting")
				}
			}
		}
	}
}

func (m *MQTTConsumer) recvMessage(_ *mqtt.Client, msg mqtt.Message) {
	m.in <- msg
}

func (m *MQTTConsumer) Stop() {
	m.Lock()
	defer m.Unlock()
	close(m.done)
	m.client.Disconnect(200)
}

func (m *MQTTConsumer) Gather(acc telegraf.Accumulator) error {
	m.Lock()
	defer m.Unlock()
	nmetrics := len(m.metricC)
	for i := 0; i < nmetrics; i++ {
		metric := <-m.metricC
		topic := <-m.topicC
		tags := metric.Tags()
		tags["topic"] = topic
		acc.AddFields(metric.Name(), metric.Fields(), tags, metric.Time())
	}
	return nil
}

func (m *MQTTConsumer) createOpts() (*mqtt.ClientOptions, error) {
	opts := mqtt.NewClientOptions()

	opts.SetClientID("Telegraf-Consumer-" + internal.RandomString(5))

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
	if user == "" {
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
	return opts, nil
}

func init() {
	inputs.Add("mqtt_consumer", func() telegraf.Input {
		return &MQTTConsumer{}
	})
}
