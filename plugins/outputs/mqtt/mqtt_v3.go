package mqtt

import (
	"fmt"
	"time"

	mqttv3 "github.com/eclipse/paho.mqtt.golang" // Library that supports v3.1.1

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

type mqttv311Client struct {
	*MQTT
	client mqttv3.Client
}

func newMQTTv311Client(cfg *MQTT) *mqttv311Client {
	return &mqttv311Client{MQTT: cfg}
}

func (m *mqttv311Client) Connect() error {
	opts := mqttv3.NewClientOptions()
	opts.KeepAlive = m.KeepAlive

	if m.Timeout < config.Duration(time.Second) {
		m.Timeout = config.Duration(5 * time.Second)
	}
	opts.WriteTimeout = time.Duration(m.Timeout)

	if m.ClientID != "" {
		opts.SetClientID(m.ClientID)
	} else {
		randomString, err := internal.RandomString(5)
		if err != nil {
			return fmt.Errorf("generating random string for client ID failed: %w", err)
		}
		opts.SetClientID("Telegraf-Output-" + randomString)
	}

	tlsCfg, err := m.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	opts.SetTLSConfig(tlsCfg)

	if !m.Username.Empty() {
		user, err := m.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		opts.SetUsername(string(user))
		config.ReleaseSecret(user)
	}
	if !m.Password.Empty() {
		password, err := m.Password.Get()
		if err != nil {
			return fmt.Errorf("getting password failed: %w", err)
		}
		opts.SetPassword(string(password))
		config.ReleaseSecret(password)
	}

	if len(m.Servers) == 0 {
		return fmt.Errorf("could not get server informations")
	}

	servers, err := parseServers(m.Servers)
	if err != nil {
		return err
	}
	for _, server := range servers {
		if tlsCfg != nil {
			server.Scheme = "tls"
		}
		broker := server.String()
		opts.AddBroker(broker)
		m.MQTT.Log.Debugf("registered mqtt broker: %v", broker)
	}

	opts.SetAutoReconnect(true)
	m.client = mqttv3.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m *mqttv311Client) Publish(topic string, body []byte) error {
	token := m.client.Publish(topic, byte(m.QoS), m.Retain, body)
	token.WaitTimeout(time.Duration(m.Timeout))
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (m *mqttv311Client) Close() error {
	if m.client.IsConnected() {
		m.client.Disconnect(20)
	}
	return nil
}
