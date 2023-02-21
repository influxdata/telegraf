//go:generate ../../../tools/readme_config_includer/generator
package mqtt

import (
	// Blank import to support go:embed compile directive
	_ "embed"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/mqtt"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

//go:embed sample.conf
var sampleConfig string

type MQTT struct {
	TopicPrefix  string          `toml:"topic_prefix" deprecated:"1.25.0;use 'topic' instead"`
	Topic        string          `toml:"topic"`
	BatchMessage bool            `toml:"batch"`
	Log          telegraf.Logger `toml:"-"`
	mqtt.MqttConfig

	client     mqtt.Client
	serializer serializers.Serializer
	generator  *TopicNameGenerator

	sync.Mutex
}

func (*MQTT) SampleConfig() string {
	return sampleConfig
}

func (m *MQTT) Init() error {
	if len(m.Servers) == 0 {
		return errors.New("no servers specified")
	}

	if m.PersistentSession && m.ClientID == "" {
		return errors.New("persistent_session requires client_id")
	}

	if m.QoS > 2 || m.QoS < 0 {
		return fmt.Errorf("qos value must be 0, 1, or 2: %d", m.QoS)
	}

	var err error
	m.generator, err = NewTopicNameGenerator(m.TopicPrefix, m.Topic)
	return err
}

func (m *MQTT) Connect() error {
	m.Lock()
	defer m.Unlock()

	client, err := mqtt.NewClient(&m.MqttConfig)
	if err != nil {
		return err
	}
	m.client = client

	_, err = m.client.Connect()
	return err
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
		topic, err := m.generator.Generate(hostname, metric)
		if err != nil {
			return fmt.Errorf("topic name couldn't be generated due to error: %w", err)
		}

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

func init() {
	outputs.Add("mqtt", func() telegraf.Output {
		return &MQTT{
			MqttConfig: mqtt.MqttConfig{
				KeepAlive:     30,
				Timeout:       config.Duration(5 * time.Second),
				AutoReconnect: true,
			},
		}
	})
}
