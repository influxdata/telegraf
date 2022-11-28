package mqtt

import (
	"fmt"
	"time"

	// Library that supports v3.1.1
	mqttv3 "github.com/eclipse/paho.mqtt.golang"
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
		opts.SetClientID("Telegraf-Output-" + internal.RandomString(5))
	}

	tlsCfg, err := m.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	opts.SetTLSConfig(tlsCfg)

	user := m.Username
	if user != "" {
		opts.SetUsername(user)
	}
	password := m.Password
	if password != "" {
		opts.SetPassword(password)
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
