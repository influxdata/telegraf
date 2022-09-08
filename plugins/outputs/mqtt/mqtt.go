//go:generate ../../../tools/readme_config_includer/generator
package mqtt

import (
	// Blank import to support go:embed compile directive
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

const (
	defaultKeepAlive = 30
)

type MQTT struct {
	Servers     []string `toml:"servers"`
	Protocol    string   `toml:"protocol"`
	Username    string   `toml:"username"`
	Password    string   `toml:"password"`
	Database    string
	Timeout     config.Duration `toml:"timeout"`
	TopicPrefix string          `toml:"topic_prefix"`
	QoS         int             `toml:"qos"`
	ClientID    string          `toml:"client_id"`
	tls.ClientConfig
	BatchMessage bool            `toml:"batch"`
	Retain       bool            `toml:"retain"`
	KeepAlive    int64           `toml:"keep_alive"`
	Log          telegraf.Logger `toml:"-"`

	client     Client
	serializer serializers.Serializer

	sync.Mutex
}

// Client is a protocol neutral MQTT client for connecting,
// disconnecting, and publishing data to a topic.
// The protocol specific clients must implement this interface
type Client interface {
	Connect() error
	Publish(topic string, data []byte) error
	Close() error
}

func (*MQTT) SampleConfig() string {
	return sampleConfig
}

func (m *MQTT) Connect() error {
	m.Lock()
	defer m.Unlock()
	if m.QoS > 2 || m.QoS < 0 {
		return fmt.Errorf("MQTT Output, invalid QoS value: %d", m.QoS)
	}

	switch m.Protocol {
	case "", "3.1.1":
		m.client = newMQTTv311Client(m)
	case "5":
		m.client = newMQTTv5Client(m)
	default:
		return fmt.Errorf("unsuported protocol %q: must be \"3.1.1\" or \"5\"", m.Protocol)
	}

	return m.client.Connect()
}

func (m *MQTT) SetSerializer(serializer serializers.Serializer) {
	m.serializer = serializer
}

func (m *MQTT) Close() error {
	return m.client.Close()
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

			err = m.client.Publish(topic, buf)
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
		err = m.client.Publish(key, buf)
		if err != nil {
			return fmt.Errorf("could not write to MQTT server, %s", err)
		}
	}

	return nil
}

func parseServers(servers []string) ([]*url.URL, error) {
	urls := make([]*url.URL, 0, len(servers))
	for _, svr := range servers {
		if !strings.Contains(svr, "://") {
			urls = append(urls, &url.URL{Scheme: "tcp", Host: svr})
		} else {
			u, err := url.Parse(svr)
			if err != nil {
				return nil, err
			}
			urls = append(urls, u)
		}
	}
	return urls, nil
}

func init() {
	outputs.Add("mqtt", func() telegraf.Output {
		return &MQTT{
			KeepAlive: defaultKeepAlive,
		}
	})
}
