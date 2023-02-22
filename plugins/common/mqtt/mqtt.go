package mqtt

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

// mqtt v5-specific publish properties.
// See https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901109
type PublishProperties struct {
	ContentType    string            `toml:"content_type"`
	ResponseTopic  string            `toml:"response_topic"`
	MessageExpiry  config.Duration   `toml:"message_expiry"`
	TopicAlias     *uint16           `toml:"topic_alias"`
	UserProperties map[string]string `toml:"user_properties"`
}

type MqttConfig struct {
	Servers             []string           `toml:"servers"`
	Protocol            string             `toml:"protocol"`
	Username            config.Secret      `toml:"username"`
	Password            config.Secret      `toml:"password"`
	Timeout             config.Duration    `toml:"timeout"`
	ConnectionTimeout   config.Duration    `toml:"connection_timeout"`
	QoS                 int                `toml:"qos"`
	ClientID            string             `toml:"client_id"`
	Retain              bool               `toml:"retain"`
	KeepAlive           int64              `toml:"keep_alive"`
	PersistentSession   bool               `toml:"persistent_session"`
	PublishPropertiesV5 *PublishProperties `toml:"v5"`

	tls.ClientConfig

	AutoReconnect    bool        `toml:"-"`
	OnConnectionLost func(error) `toml:"-"`
}

// Client is a protocol neutral MQTT client for connecting,
// disconnecting, and publishing data to a topic.
// The protocol specific clients must implement this interface
type Client interface {
	Connect() (bool, error)
	Publish(topic string, data []byte) error
	SubscribeMultiple(filters map[string]byte, callback paho.MessageHandler) error
	AddRoute(topic string, callback paho.MessageHandler)
	Close() error
}

func NewClient(cfg *MqttConfig) (Client, error) {
	if len(cfg.Servers) == 0 {
		return nil, errors.New("no servers specified")
	}

	if cfg.PersistentSession && cfg.ClientID == "" {
		return nil, errors.New("persistent_session requires client_id")
	}

	if cfg.QoS > 2 || cfg.QoS < 0 {
		return nil, fmt.Errorf("invalid QoS value %d; must be 0, 1 or 2", cfg.QoS)
	}

	switch cfg.Protocol {
	case "", "3.1.1":
		return NewMQTTv311Client(cfg)
	case "5":
		return NewMQTTv5Client(cfg)
	}
	return nil, fmt.Errorf("unsuported protocol %q: must be \"3.1.1\" or \"5\"", cfg.Protocol)
}

func parseServers(servers []string) ([]*url.URL, error) {
	urls := make([]*url.URL, 0, len(servers))
	for _, svr := range servers {
		// Preserve support for host:port style servers; deprecated in Telegraf 1.4.4
		if !strings.Contains(svr, "://") {
			urls = append(urls, &url.URL{Scheme: "tcp", Host: svr})
			continue
		}

		u, err := url.Parse(svr)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}
