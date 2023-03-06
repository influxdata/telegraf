package mqtt

import (
	"fmt"
	"time"

	mqttv3 "github.com/eclipse/paho.mqtt.golang" // Library that supports v3.1.1

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

type mqttv311Client struct {
	client  mqttv3.Client
	timeout time.Duration
	qos     int
	retain  bool
}

func NewMQTTv311Client(cfg *MqttConfig) (*mqttv311Client, error) {
	opts := mqttv3.NewClientOptions()
	opts.KeepAlive = cfg.KeepAlive
	opts.WriteTimeout = time.Duration(cfg.Timeout)
	if time.Duration(cfg.ConnectionTimeout) >= 1*time.Second {
		opts.ConnectTimeout = time.Duration(cfg.ConnectionTimeout)
	}
	opts.SetCleanSession(!cfg.PersistentSession)
	if cfg.OnConnectionLost != nil {
		onConnectionLost := func(_ mqttv3.Client, err error) {
			cfg.OnConnectionLost(err)
		}
		opts.SetConnectionLostHandler(onConnectionLost)
	}
	opts.SetAutoReconnect(cfg.AutoReconnect)

	if cfg.ClientID != "" {
		opts.SetClientID(cfg.ClientID)
	} else {
		id, err := internal.RandomString(5)
		if err != nil {
			return nil, fmt.Errorf("generating random client ID failed: %w", err)
		}
		opts.SetClientID("Telegraf-Output-" + id)
	}

	tlsCfg, err := cfg.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	opts.SetTLSConfig(tlsCfg)

	if !cfg.Username.Empty() {
		user, err := cfg.Username.Get()
		if err != nil {
			return nil, fmt.Errorf("getting username failed: %w", err)
		}
		opts.SetUsername(string(user))
		config.ReleaseSecret(user)
	}
	if !cfg.Password.Empty() {
		password, err := cfg.Password.Get()
		if err != nil {
			return nil, fmt.Errorf("getting password failed: %w", err)
		}
		opts.SetPassword(string(password))
		config.ReleaseSecret(password)
	}

	servers, err := parseServers(cfg.Servers)
	if err != nil {
		return nil, err
	}
	for _, server := range servers {
		if tlsCfg != nil {
			server.Scheme = "tls"
		}
		broker := server.String()
		opts.AddBroker(broker)
	}

	return &mqttv311Client{
		client:  mqttv3.NewClient(opts),
		timeout: time.Duration(cfg.Timeout),
		qos:     cfg.QoS,
		retain:  cfg.Retain,
	}, nil
}

func (m *mqttv311Client) Connect() (bool, error) {
	token := m.client.Connect()

	if token.Wait() && token.Error() != nil {
		return false, token.Error()
	}

	// Persistent sessions should skip subscription if a session is present, as
	// the subscriptions are stored by the server.
	type sessionPresent interface {
		SessionPresent() bool
	}
	if t, ok := token.(sessionPresent); ok {
		return t.SessionPresent(), nil
	}

	return false, nil
}

func (m *mqttv311Client) Publish(topic string, body []byte) error {
	token := m.client.Publish(topic, byte(m.qos), m.retain, body)
	if !token.WaitTimeout(m.timeout) {
		return internal.ErrTimeout
	}
	return token.Error()
}

func (m *mqttv311Client) SubscribeMultiple(filters map[string]byte, callback mqttv3.MessageHandler) error {
	token := m.client.SubscribeMultiple(filters, callback)
	token.Wait()
	return token.Error()
}

func (m *mqttv311Client) AddRoute(topic string, callback mqttv3.MessageHandler) {
	m.client.AddRoute(topic, callback)
}

func (m *mqttv311Client) Close() error {
	if m.client.IsConnected() {
		m.client.Disconnect(100)
	}
	return nil
}
