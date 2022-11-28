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

type mqttv5Client struct {
	*MQTT
	client *mqttv5auto.ConnectionManager
}

func newMQTTv5Client(cfg *MQTT) *mqttv5Client {
	return &mqttv5Client{MQTT: cfg}
}

func (m *mqttv5Client) Connect() error {
	opts := mqttv5auto.ClientConfig{}
	if m.ClientID != "" {
		opts.ClientID = m.ClientID
	} else {
		opts.ClientID = "Telegraf-Output-" + internal.RandomString(5)
	}

	user := m.Username
	pass := m.Password
	opts.SetUsernamePassword(user, []byte(pass))

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
		Topic:   topic,
		QoS:     byte(m.QoS),
		Retain:  m.Retain,
		Payload: body,
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *mqttv5Client) Close() error {
	return m.client.Disconnect(context.Background())
}
