package mosquitto

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

type MeasureTopic struct {
	topic       string
	measurement string
	field       string
}

type MQTTConsumer struct {
	Servers           []string
	TopicsToSubscribe map[string]MeasureTopic
	Username          string
	Password          string
	ConnectionTimeout internal.Duration `toml:"connection_timeout"`

	Tags []string

	// Legacy metric buffer support
	MetricBuffer int

	ClientID string `toml:"client_id"`
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

// Build and initialize a MQTTConsumer
func NewMQTTConsumer() *MQTTConsumer {
	m := MQTTConsumer{
		TopicsToSubscribe: setupMeasurementTopics(),
		ConnectionTimeout: defaultConnectionTimeout,
	}

	return &m
}

var sampleConfig = `
  ## MQTT broker URLs to be used. The format should be scheme://host:port,
  ## schema can be tcp, ssl, or ws.
  servers = ["tcp://localhost:1883"]

  ## Connection timeout for initial connection in seconds
  connection_timeout = "30s"
  
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

  ## Optional static tags to be added to measurements
  tags = [
	"mosquitto_instance_name: the_instance_name",
	"other_tag_name: other_value",
  ]
`

func (m *MQTTConsumer) SampleConfig() string {
	return sampleConfig
}

func (m *MQTTConsumer) Description() string {
	return "Gather metrics from Mosquitto $SYS topic(s)"
}

func (m *MQTTConsumer) SetParser(parser parsers.Parser) {
}

func (m *MQTTConsumer) Start(acc telegraf.Accumulator) error {
	m.Lock()
	defer m.Unlock()
	m.connected = false

	m.acc = acc

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

		m.connected = false

		return err
	}

	go m.receiver()

	return nil
}

func (m *MQTTConsumer) onConnect(c mqtt.Client) {

	log.Printf("I! MQTT Client Connected")
	if !m.connected {
		i := 0
		topicsNames := make([]string, len(m.TopicsToSubscribe))
		topics := make(map[string]byte)

		for topic := range m.TopicsToSubscribe {
			topics[topic] = byte(0)
			topicsNames[i] = topic
			i++
		}

		log.Printf("I! Subscribing to %d topics", len(m.TopicsToSubscribe))
		subscribeToken := c.SubscribeMultiple(topics, m.recvMessage)
		subscribeToken.Wait()
		if subscribeToken.Error() != nil {
			m.acc.AddError(fmt.Errorf("E! MQTT Subscribe Error\ntopics: %s\nerror: %s",
				strings.Join(topicsNames, ","), subscribeToken.Error()))
		}
		m.connected = true
	}
	return
}

func (m *MQTTConsumer) onConnectionLost(c mqtt.Client, err error) {
	m.connected = false
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
			metric := msg.Payload()

			// add custom defined tags
			tags := make(map[string]string)
			for rows := range m.Tags {
				var kv = strings.Split(m.Tags[rows], ":")
				tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
			// add current topic to tags
			tags["topic"] = topic

			fields := map[string]interface{}{
				m.TopicsToSubscribe[topic].field: metric,
			}

			m.acc.AddFields(m.TopicsToSubscribe[topic].measurement, fields, tags)
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
	opts.SetOnConnectHandler(m.onConnect)
	opts.SetConnectionLostHandler(m.onConnectionLost)

	return opts, nil
}

func init() {
	inputs.Add("mosquitto", func() telegraf.Input {
		return NewMQTTConsumer()
	})
}

func setupMeasurementTopics() map[string]MeasureTopic {

	m := make(map[string]MeasureTopic)

	m["$SYS/broker/bytes/received"] = MeasureTopic{
		topic:       "$SYS/broker/bytes/received",
		measurement: "mosquitto.global.bytes.received",
		field:       "count",
	}
	m["$SYS/broker/bytes/sent"] = MeasureTopic{
		topic:       "$SYS/broker/bytes/sent",
		measurement: "mosquitto.bytes.sent",
		field:       "count",
	}
	m["$SYS/broker/messages/inflight"] = MeasureTopic{
		topic:       "$SYS/broker/messages/inflight",
		measurement: "mosquitto.messages.inflight",
		field:       "count",
	}
	m["$SYS/broker/messages/received"] = MeasureTopic{
		topic:       "$SYS/broker/messages/received",
		measurement: "mosquitto.messages.received",
		field:       "count",
	}
	m["$SYS/broker/messages/sent"] = MeasureTopic{
		topic:       "$SYS/broker/messages/sent",
		measurement: "mosquitto.messages.sent",
		field:       "count",
	}
	m["$SYS/broker/messages/stored"] = MeasureTopic{
		topic:       "$SYS/broker/messages/stored",
		measurement: "mosquitto.messages.stored",
		field:       "count",
	}
	m["$SYS/broker/publish/messages/dropped"] = MeasureTopic{
		topic:       "$SYS/broker/publish/messages/dropped",
		measurement: "mosquitto.publish_messages.dropped",
		field:       "count",
	}
	m["$SYS/broker/publish/messages/received"] = MeasureTopic{
		topic:       "$SYS/broker/publish/messages/received",
		measurement: "mosquitto.publish_messages.received",
		field:       "count",
	}
	m["$SYS/broker/publish/messages/sent"] = MeasureTopic{
		topic:       "$SYS/broker/publish/messages/sent",
		measurement: "mosquitto.publish_messages.sent",
		field:       "count",
	}
	m["$SYS/broker/retained messages/count"] = MeasureTopic{
		topic:       "$SYS/broker/retained messages/count",
		measurement: "mosquitto.retained_messages",
		field:       "count",
	}
	m["$SYS/broker/subscriptions/count"] = MeasureTopic{
		topic:       "$SYS/broker/subscriptions/count",
		measurement: "mosquitto.subscriptions",
		field:       "count",
	}
	m["$SYS/broker/clients/connected"] = MeasureTopic{
		topic:       "$SYS/broker/clients/connected",
		measurement: "mosquitto.clients.connected",
		field:       "count",
	}
	m["$SYS/broker/clients/disconnected"] = MeasureTopic{
		topic:       "$SYS/broker/clients/disconnected",
		measurement: "mosquitto.clients.disconnected",
		field:       "count",
	}
	m["$SYS/broker/clients/expired"] = MeasureTopic{
		topic:       "$SYS/broker/clients/expired",
		measurement: "mosquitto.clients.expired",
		field:       "count",
	}
	m["$SYS/broker/clients/maximum"] = MeasureTopic{
		topic:       "$SYS/broker/clients/maximum",
		measurement: "mosquitto.clients",
		field:       "maximum",
	}
	m["$SYS/broker/clients/total"] = MeasureTopic{
		topic:       "$SYS/broker/clients/total",
		measurement: "mosquitto.clients",
		field:       "total",
	}
	m["$SYS/broker/load/connections/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/connections/1min",
		measurement: "mosquitto.load.connections",
		field:       "average",
	}
	m["$SYS/broker/load/sockets/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/sockets/1min",
		measurement: "mosquitto.load.sockets",
		field:       "average",
	}
	m["$SYS/broker/load/bytes/received/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/bytes/received/1min",
		measurement: "mosquitto.load.bytes.received",
		field:       "average",
	}
	m["$SYS/broker/load/bytes/sent/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/bytes/sent/1min",
		measurement: "mosquitto.load.bytes.sent",
		field:       "average",
	}
	m["$SYS/broker/load/messages/received/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/messages/received/1min",
		measurement: "mosquitto.load.messages.received",
		field:       "average",
	}
	m["$SYS/broker/load/messages/sent/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/messages/sent/1min",
		measurement: "mosquitto.load.messages.sent",
		field:       "average",
	}
	m["$SYS/broker/load/publish/dropped/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/publish/dropped/1min",
		measurement: "mosquitto.load.publish_messages.dropped",
		field:       "average",
	}
	m["$SYS/broker/load/publish/received/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/publish/received/1min",
		measurement: "mosquitto.load.publish_messages.received",
		field:       "average",
	}
	m["$SYS/broker/load/publish/sent/1min"] = MeasureTopic{
		topic:       "$SYS/broker/load/publish/sent/1min",
		measurement: "mosquitto.load.publish_messages.sent",
		field:       "average",
	}
	m["$SYS/broker/heap/current size"] = MeasureTopic{
		topic:       "$SYS/broker/heap/current size",
		measurement: "mosquitto.heap",
		field:       "current",
	}
	m["$SYS/broker/heap/maximum size"] = MeasureTopic{
		topic:       "$SYS/broker/heap/maximum size",
		measurement: "mosquitto.heap",
		field:       "maximum",
	}

	return m
}
