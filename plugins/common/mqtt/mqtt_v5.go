package mqtt

import (
	"context"
	"fmt"
	"net/url"
	"time"

	mqttv5auto "github.com/eclipse/paho.golang/autopaho"
	mqttv5 "github.com/eclipse/paho.golang/paho"
	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

type mqttv5Client struct {
	client     *mqttv5auto.ConnectionManager
	options    mqttv5auto.ClientConfig
	timeout    time.Duration
	qos        int
	retain     bool
	properties *mqttv5.PublishProperties
}

func NewMQTTv5Client(cfg *MqttConfig) (*mqttv5Client, error) {
	opts := mqttv5auto.ClientConfig{
		KeepAlive:      uint16(cfg.KeepAlive),
		OnConnectError: cfg.OnConnectionLost,
	}
	opts.SetConnectPacketConfigurator(func(c *mqttv5.Connect) *mqttv5.Connect {
		c.CleanStart = cfg.PersistentSession
		return c
	})

	if time.Duration(cfg.ConnectionTimeout) >= 1*time.Second {
		opts.ConnectTimeout = time.Duration(cfg.ConnectionTimeout)
	}

	if cfg.ClientID != "" {
		opts.ClientID = cfg.ClientID
	} else {
		id, err := internal.RandomString(5)
		if err != nil {
			return nil, fmt.Errorf("generating random client ID failed: %w", err)
		}
		opts.ClientID = "Telegraf-Output-" + id
	}

	user, err := cfg.Username.Get()
	if err != nil {
		return nil, fmt.Errorf("getting username failed: %w", err)
	}
	pass, err := cfg.Password.Get()
	if err != nil {
		config.ReleaseSecret(user)
		return nil, fmt.Errorf("getting password failed: %w", err)
	}
	opts.SetUsernamePassword(string(user), pass)
	config.ReleaseSecret(user)
	config.ReleaseSecret(pass)

	tlsCfg, err := cfg.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if tlsCfg != nil {
		opts.TlsCfg = tlsCfg
	}

	brokers := make([]*url.URL, 0)
	servers, err := parseServers(cfg.Servers)
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if tlsCfg != nil {
			server.Scheme = "tls"
		}
		brokers = append(brokers, server)
	}
	opts.BrokerUrls = brokers

	// Build the v5 specific publish properties if they are present in the config.
	// These should not change during the lifecycle of the client.
	var properties *mqttv5.PublishProperties
	if cfg.PublishPropertiesV5 != nil {
		properties = &mqttv5.PublishProperties{
			ContentType:   cfg.PublishPropertiesV5.ContentType,
			ResponseTopic: cfg.PublishPropertiesV5.ResponseTopic,
			TopicAlias:    cfg.PublishPropertiesV5.TopicAlias,
		}

		messageExpiry := time.Duration(cfg.PublishPropertiesV5.MessageExpiry)
		if expirySeconds := uint32(messageExpiry.Seconds()); expirySeconds > 0 {
			properties.MessageExpiry = &expirySeconds
		}

		properties.User = make([]mqttv5.UserProperty, 0, len(cfg.PublishPropertiesV5.UserProperties))
		for k, v := range cfg.PublishPropertiesV5.UserProperties {
			properties.User.Add(k, v)
		}
	}

	return &mqttv5Client{
		options:    opts,
		timeout:    time.Duration(cfg.Timeout),
		qos:        cfg.QoS,
		retain:     cfg.Retain,
		properties: properties,
	}, nil
}

func (m *mqttv5Client) Connect() (bool, error) {
	client, err := mqttv5auto.NewConnection(context.Background(), m.options)
	if err != nil {
		return false, err
	}
	m.client = client
	return false, client.AwaitConnection(context.Background())
}

func (m *mqttv5Client) Publish(topic string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	_, err := m.client.Publish(ctx, &mqttv5.Publish{
		Topic:      topic,
		QoS:        byte(m.qos),
		Retain:     m.retain,
		Payload:    body,
		Properties: m.properties,
	})

	return err
}

func (m *mqttv5Client) SubscribeMultiple(filters map[string]byte, callback paho.MessageHandler) error {
	_, _ = filters, callback
	panic("not implemented")
}

func (m *mqttv5Client) AddRoute(topic string, callback paho.MessageHandler) {
	_, _ = topic, callback
	panic("not implemented")
}

func (m *mqttv5Client) Close() error {
	return m.client.Disconnect(context.Background())
}
