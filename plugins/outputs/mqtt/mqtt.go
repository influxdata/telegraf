package mqtt

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	paho "github.com/eclipse/paho.mqtt.golang"
)

var sampleConfig = `
  servers = ["localhost:1883"] # required.

  ## MQTT outputs send metrics to this topic format
  ##    "<topic_prefix>/<hostname>/<pluginname>/"
  ##   ex: prefix/web01.example.com/mem
  topic_prefix = "telegraf"

  ## username and password to connect MQTT server.
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"

  ## Timeout for write operations. default: 5s
  # timeout = "5s"

  ## client ID, if not set a random ID is generated
  # client_id = ""

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

type MQTT struct {
	Servers     []string `toml:"servers"`
	Username    string
	Password    string
	Database    string
	Timeout     internal.Duration
	TopicPrefix string
	QoS         int    `toml:"qos"`
	ClientID    string `toml:"client_id"`

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	client paho.Client
	opts   *paho.ClientOptions

	serializer serializers.Serializer

	sync.Mutex
}

func (m *MQTT) Connect() error {
	var err error
	m.Lock()
	defer m.Unlock()
	if m.QoS > 2 || m.QoS < 0 {
		return fmt.Errorf("MQTT Output, invalid QoS value: %d", m.QoS)
	}

	m.opts, err = m.createOpts()
	if err != nil {
		return err
	}

	m.client = paho.NewClient(m.opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m *MQTT) SetSerializer(serializer serializers.Serializer) {
	m.serializer = serializer
}

func (m *MQTT) Close() error {
	if m.client.IsConnected() {
		m.client.Disconnect(20)
	}
	return nil
}

func (m *MQTT) SampleConfig() string {
	return sampleConfig
}

func (m *MQTT) Description() string {
	return "Configuration for MQTT server to send metrics to"
}

func (m *MQTT) Write(metrics []telegraf.Metric) error {
	m.Lock()
	defer m.Unlock()
	if len(metrics) == 0 {
		return nil
	}
	hostname, ok := metrics[0].Tags()["host"]
	if !ok {
		hostname = ""
	}

	for _, metric := range metrics {
		var t []string
		if m.TopicPrefix != "" {
			t = append(t, m.TopicPrefix)
		}
		if hostname != "" {
			t = append(t, hostname)
		}

		t = append(t, metric.Name())
		topic := strings.Join(t, "/")

		buf, err := m.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("MQTT Could not serialize metric: %s",
				metric.String())
		}

		err = m.publish(topic, buf)
		if err != nil {
			return fmt.Errorf("Could not write to MQTT server, %s", err)
		}
	}

	return nil
}

func (m *MQTT) publish(topic string, body []byte) error {
	token := m.client.Publish(topic, byte(m.QoS), false, body)
	token.WaitTimeout(m.Timeout.Duration)
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (m *MQTT) createOpts() (*paho.ClientOptions, error) {
	opts := paho.NewClientOptions()
	opts.KeepAlive = 0 * time.Second

	if m.Timeout.Duration < time.Second {
		m.Timeout.Duration = 5 * time.Second
	}
	opts.WriteTimeout = m.Timeout.Duration

	if m.ClientID != "" {
		opts.SetClientID(m.ClientID)
	} else {
		opts.SetClientID("Telegraf-Output-" + internal.RandomString(5))
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
	return opts, nil
}

func init() {
	outputs.Add("mqtt", func() telegraf.Output {
		return &MQTT{}
	})
}
