//go:generate ../../../tools/readme_config_includer/generator
package mqtt

import (
	// Blank import to support go:embed compile directive
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/mqtt"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

var pluginNameRe = regexp.MustCompile(`({{.*\B)\.PluginName(\b[^}]*}})`)
var hostnameRe = regexp.MustCompile(`({{.*\B)\.Hostname(\b[^}]*}})`)

type message struct {
	topic   string
	payload []byte
}

type MQTT struct {
	Topic           string          `toml:"topic"`
	Layout          string          `toml:"layout"`
	HomieDeviceName string          `toml:"homie_device_name"`
	HomieNodeID     string          `toml:"homie_node_id"`
	Log             telegraf.Logger `toml:"-"`
	mqtt.MqttConfig

	client     mqtt.Client
	serializer telegraf.Serializer
	template   *template.Template

	homieDeviceNameGenerator *template.Template
	homieNodeIDGenerator     *template.Template
	homieSeen                map[string]map[string]bool

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

	// Prepare the topic
	topic := hostnameRe.ReplaceAllString(m.Topic, `$1.Tag "host"$2`)
	topic = pluginNameRe.ReplaceAllString(topic, `$1.Name$2`)

	tmpl, err := template.New("topic_name").Funcs(sprig.TxtFuncMap()).Parse(topic)
	if err != nil {
		return fmt.Errorf("creating topic template failed: %w", err)
	}
	for _, p := range strings.Split(topic, "/") {
		if strings.ContainsAny(p, "#+") {
			return fmt.Errorf("found forbidden character %s in the topic name %s", p, topic)
		}
	}
	m.template = tmpl

	switch m.Layout {
	case "":
		m.Layout = "non-batch"
	case "non-batch", "batch", "field":
	case "homie-v4":
		if m.HomieDeviceName == "" {
			return errors.New("missing 'homie_device_name' option")
		}

		m.HomieDeviceName = pluginNameRe.ReplaceAllString(m.HomieDeviceName, `$1.Name$2`)
		m.homieDeviceNameGenerator, err = template.New("topic_name").Funcs(sprig.TxtFuncMap()).Parse(m.HomieDeviceName)
		if err != nil {
			return fmt.Errorf("creating device name generator failed: %w", err)
		}

		if m.HomieNodeID == "" {
			return errors.New("missing 'homie_node_id' option")
		}

		m.HomieNodeID = pluginNameRe.ReplaceAllString(m.HomieNodeID, `$1.Name$2`)
		m.homieNodeIDGenerator, err = template.New("topic_name").Funcs(sprig.TxtFuncMap()).Parse(m.HomieNodeID)
		if err != nil {
			return fmt.Errorf("creating node ID name generator failed: %w", err)
		}
	default:
		return fmt.Errorf("invalid layout %q", m.Layout)
	}

	m.MqttConfig.ClientTrace = m.MqttConfig.ClientTrace || m.Log.Level().Includes(telegraf.Trace)

	return nil
}

func (m *MQTT) Connect() error {
	m.Lock()
	defer m.Unlock()

	m.homieSeen = make(map[string]map[string]bool)

	client, err := mqtt.NewClient(&m.MqttConfig)
	if err != nil {
		return err
	}
	m.client = client

	_, err = m.client.Connect()
	return err
}

func (m *MQTT) SetSerializer(serializer telegraf.Serializer) {
	m.serializer = serializer
}

func (m *MQTT) Close() error {
	// Unregister devices if Homie layout was used. Usually we should do this
	// using a "will" message, but this can only be done at connect time where,
	// due to the dynamic nature of Telegraf messages, we do not know the topics
	// to issue that "will" yet.
	if len(m.homieSeen) > 0 {
		for topic := range m.homieSeen {
			//nolint:errcheck // We will ignore potential errors as we cannot do anything here
			m.client.Publish(topic+"/$state", []byte("lost"))
		}
		// Give the messages some time to settle
		time.Sleep(100 * time.Millisecond)
	}
	return m.client.Close()
}

func (m *MQTT) Write(metrics []telegraf.Metric) error {
	m.Lock()
	defer m.Unlock()
	if len(metrics) == 0 {
		return nil
	}

	// Group the metrics to topics and serialize them
	var topicMessages []message
	switch m.Layout {
	case "batch":
		topicMessages = m.collectBatch(metrics)
	case "non-batch":
		topicMessages = m.collectNonBatch(metrics)
	case "field":
		topicMessages = m.collectField(metrics)
	case "homie-v4":
		topicMessages = m.collectHomieV4(metrics)
	default:
		return fmt.Errorf("unknown layout %q", m.Layout)
	}

	for _, msg := range topicMessages {
		if err := m.client.Publish(msg.topic, msg.payload); err != nil {
			// We do receive a timeout error if the remote broker is down,
			// so let's retry the metrics in this case and drop them otherwise.
			if errors.Is(err, internal.ErrTimeout) {
				return fmt.Errorf("could not publish message to MQTT server: %w", err)
			}
			m.Log.Warnf("Could not publish message to MQTT server: %v", err)
		}
	}

	return nil
}

func (m *MQTT) collectNonBatch(metrics []telegraf.Metric) []message {
	collection := make([]message, 0, len(metrics))
	for _, metric := range metrics {
		topic, err := m.generateTopic(metric)
		if err != nil {
			m.Log.Warnf("Generating topic name failed: %v", err)
			m.Log.Debugf("metric was: %v", metric)
			continue
		}

		buf, err := m.serializer.Serialize(metric)
		if err != nil {
			m.Log.Warnf("Could not serialize metric for topic %q: %v", topic, err)
			m.Log.Debugf("metric was: %v", metric)
			continue
		}
		collection = append(collection, message{topic, buf})
	}

	return collection
}

func (m *MQTT) collectBatch(metrics []telegraf.Metric) []message {
	metricsCollection := make(map[string][]telegraf.Metric)
	for _, metric := range metrics {
		topic, err := m.generateTopic(metric)
		if err != nil {
			m.Log.Warnf("Generating topic name failed: %v", err)
			m.Log.Debugf("metric was: %v", metric)
			continue
		}
		metricsCollection[topic] = append(metricsCollection[topic], metric)
	}

	collection := make([]message, 0, len(metricsCollection))
	for topic, ms := range metricsCollection {
		buf, err := m.serializer.SerializeBatch(ms)
		if err != nil {
			m.Log.Warnf("Could not serialize metric batch for topic %q: %v", topic, err)
			continue
		}
		collection = append(collection, message{topic, buf})
	}
	return collection
}

func (m *MQTT) collectField(metrics []telegraf.Metric) []message {
	var collection []message
	for _, metric := range metrics {
		topic, err := m.generateTopic(metric)
		if err != nil {
			m.Log.Warnf("Generating topic name failed: %v", err)
			m.Log.Debugf("metric was: %v", metric)
			continue
		}

		for n, v := range metric.Fields() {
			buf, err := internal.ToString(v)
			if err != nil {
				m.Log.Warnf("Could not serialize metric for topic %q field %q: %v", topic, n, err)
				m.Log.Debugf("metric was: %v", metric)
				continue
			}
			collection = append(collection, message{topic + "/" + n, []byte(buf)})
		}
	}

	return collection
}

func (m *MQTT) collectHomieV4(metrics []telegraf.Metric) []message {
	var collection []message
	for _, metric := range metrics {
		topic, err := m.generateTopic(metric)
		if err != nil {
			m.Log.Warnf("Generating topic name failed: %v", err)
			m.Log.Debugf("metric was: %v", metric)
			continue
		}

		msgs, nodeID, err := m.collectHomieDeviceMessages(topic, metric)
		if err != nil {
			m.Log.Warn(err.Error())
			m.Log.Debugf("metric was: %v", metric)
			continue
		}
		path := topic + "/" + nodeID
		collection = append(collection, msgs...)

		for _, tag := range metric.TagList() {
			propID := normalizeID(tag.Key)
			collection = append(collection,
				message{path + "/" + propID, []byte(tag.Value)},
				message{path + "/" + propID + "/$name", []byte(tag.Key)},
				message{path + "/" + propID + "/$datatype", []byte("string")},
			)
		}

		for _, field := range metric.FieldList() {
			v, dt, err := convertType(field.Value)
			if err != nil {
				m.Log.Warnf("Could not serialize metric for topic %q field %q: %v", topic, field.Key, err)
				m.Log.Debugf("metric was: %v", metric)
				continue
			}
			propID := normalizeID(field.Key)
			collection = append(collection,
				message{path + "/" + propID, []byte(v)},
				message{path + "/" + propID + "/$name", []byte(field.Key)},
				message{path + "/" + propID + "/$datatype", []byte(dt)},
			)
		}
	}

	return collection
}

func (m *MQTT) generateTopic(metric telegraf.Metric) (string, error) {
	var b strings.Builder
	err := m.template.Execute(&b, metric)
	if err != nil {
		return "", err
	}
	var ts []string
	for _, p := range strings.Split(b.String(), "/") {
		if p != "" {
			ts = append(ts, p)
		}
	}
	topic := strings.Join(ts, "/")

	// This is to keep backward compatibility with previous behaviour where the plugin name was always present
	if topic == "" {
		return metric.Name(), nil
	}
	if strings.HasPrefix(b.String(), "/") {
		topic = "/" + topic
	}
	return topic, nil
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
