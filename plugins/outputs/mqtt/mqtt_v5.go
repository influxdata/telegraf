package mqtt

import (
	"context"
	"fmt"
	"net/url"
	"time"

	mqttv5auto "github.com/eclipse/paho.golang/autopaho"
	mqttv5 "github.com/eclipse/paho.golang/paho"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

// mqtt v5-specific publish properties.
// See https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901109
type mqttv5PublishProperties struct {
	ContentType    string            `toml:"content_type"`
	ResponseTopic  string            `toml:"response_topic"`
	MessageExpiry  config.Duration   `toml:"message_expiry"`
	TopicAlias     *uint16           `toml:"topic_alias"`
	UserProperties map[string]string `toml:"user_properties"`
}

type mqttv5Client struct {
	*MQTT
	client            *mqttv5auto.ConnectionManager
	publishProperties *mqttv5.PublishProperties
}

func newMQTTv5Client(cfg *MQTT) *mqttv5Client {
	return &mqttv5Client{
		MQTT:              cfg,
		publishProperties: buildPublishProperties(cfg),
	}
}

// Build the v5 specific publish properties if they are present in the
// config.
// These should not change during the lifecycle of the client.
func buildPublishProperties(cfg *MQTT) *mqttv5.PublishProperties {
	if cfg.V5PublishProperties == nil {
		return nil
	}

	publishProperties := &mqttv5.PublishProperties{
		ContentType:   cfg.V5PublishProperties.ContentType,
		ResponseTopic: cfg.V5PublishProperties.ResponseTopic,
		TopicAlias:    cfg.V5PublishProperties.TopicAlias,
		User:          make([]mqttv5.UserProperty, 0, len(cfg.V5PublishProperties.UserProperties)),
	}

	messageExpiry := time.Duration(cfg.V5PublishProperties.MessageExpiry)
	if expirySeconds := uint32(messageExpiry.Seconds()); expirySeconds > 0 {
		publishProperties.MessageExpiry = &expirySeconds
	}

	for k, v := range cfg.V5PublishProperties.UserProperties {
		publishProperties.User.Add(k, v)
	}

	return publishProperties
}

func (m *mqttv5Client) Connect() error {
	opts := mqttv5auto.ClientConfig{}
	if m.ClientID != "" {
		opts.ClientID = m.ClientID
	} else {
		randomString, err := internal.RandomString(5)
		if err != nil {
			return fmt.Errorf("generating random string for client ID failed: %w", err)
		}
		opts.ClientID = "Telegraf-Output-" + randomString
	}

	user, err := m.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	pass, err := m.Password.Get()
	if err != nil {
		config.ReleaseSecret(user)
		return fmt.Errorf("getting password failed: %w", err)
	}
	opts.SetUsernamePassword(string(user), pass)
	config.ReleaseSecret(user)
	config.ReleaseSecret(pass)

	tlsCfg, err := m.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if tlsCfg != nil {
		opts.TlsCfg = tlsCfg
	}

	if len(m.Servers) == 0 {
		return fmt.Errorf("could not get host informations")
	}

	brokers := make([]*url.URL, 0)
	servers, err := parseServers(m.Servers)
	if err != nil {
		return err
	}

	for _, server := range servers {
		if tlsCfg != nil {
			server.Scheme = "tls"
		}
		brokers = append(brokers, server)
		m.MQTT.Log.Debugf("registered mqtt broker: %s", server.String())
	}
	opts.BrokerUrls = brokers

	opts.KeepAlive = uint16(m.KeepAlive)
	if m.Timeout < config.Duration(time.Second) {
		m.Timeout = config.Duration(5 * time.Second)
	}

	client, err := mqttv5auto.NewConnection(context.Background(), opts)
	if err != nil {
		return err
	}
	m.client = client
	return client.AwaitConnection(context.Background())
}

func (m *mqttv5Client) Publish(topic string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.Timeout))
	defer cancel()
	_, err := m.client.Publish(ctx, &mqttv5.Publish{
		Topic:      topic,
		QoS:        byte(m.QoS),
		Retain:     m.Retain,
		Payload:    body,
		Properties: m.publishProperties,
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *mqttv5Client) Close() error {
	return m.client.Disconnect(context.Background())
}
