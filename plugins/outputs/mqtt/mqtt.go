package mqtt

import (
	"fmt"
	"strings"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const (
	defaultKeepAlive = 0
)

type MQTT struct {
	Servers     []string `toml:"servers"`
	Username    string
	Password    string
	Database    string
	Timeout     config.Duration
	TopicPrefix string
	QoS         int    `toml:"qos"`
	ClientID    string `toml:"client_id"`
	tls.ClientConfig
	BatchMessage bool            `toml:"batch"`
	Retain       bool            `toml:"retain"`
	KeepAlive    int64           `toml:"keep_alive"`
	Log          telegraf.Logger `toml:"-"`

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

	metricsmap := make(map[string][]telegraf.Metric)

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

		if m.BatchMessage {
			metricsmap[topic] = append(metricsmap[topic], metric)
		} else {
			buf, err := m.serializer.Serialize(metric)
			if err != nil {
				m.Log.Debugf("Could not serialize metric: %v", err)
				continue
			}

			err = m.publish(topic, buf)
			if err != nil {
				return fmt.Errorf("could not write to MQTT server, %s", err)
			}
		}
	}

	for key := range metricsmap {
		buf, err := m.serializer.SerializeBatch(metricsmap[key])

		if err != nil {
			return err
		}
		publisherr := m.publish(key, buf)
		if publisherr != nil {
			return fmt.Errorf("could not write to MQTT server, %s", publisherr)
		}
	}

	return nil
}

func (m *MQTT) publish(topic string, body []byte) error {
	token := m.client.Publish(topic, byte(m.QoS), m.Retain, body)
	token.WaitTimeout(time.Duration(m.Timeout))
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (m *MQTT) createOpts() (*paho.ClientOptions, error) {
	opts := paho.NewClientOptions()
	opts.KeepAlive = m.KeepAlive

	if m.Timeout < config.Duration(time.Second) {
		m.Timeout = config.Duration(5 * time.Second)
	}
	opts.WriteTimeout = time.Duration(m.Timeout)

	if m.ClientID != "" {
		opts.SetClientID(m.ClientID)
	} else {
		opts.SetClientID("Telegraf-Output-" + internal.RandomString(5))
	}

	tlsCfg, err := m.ClientConfig.TLSConfig()
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
		return opts, fmt.Errorf("could not get host informations")
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
		return &MQTT{
			KeepAlive: defaultKeepAlive,
		}
	})
}
