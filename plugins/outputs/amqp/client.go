package amqp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/proxy"
)

type ClientConfig struct {
	brokers           []string
	exchange          string
	exchangeType      string
	exchangePassive   bool
	exchangeDurable   bool
	exchangeArguments amqp.Table
	encoding          string
	headers           amqp.Table
	deliveryMode      uint8
	tlsConfig         *tls.Config
	timeout           time.Duration
	auth              []amqp.Authentication
	dialer            *proxy.ProxiedDialer
	log               telegraf.Logger
}

type client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *ClientConfig
}

// newClient opens a connection to one of the brokers at random
func newClient(config *ClientConfig) (*client, error) {
	client := &client{
		config: config,
	}

	p := rand.Perm(len(config.brokers))
	for _, n := range p {
		broker := config.brokers[n]
		config.log.Debugf("Connecting to %q", broker)
		conn, err := amqp.DialConfig(
			broker, amqp.Config{
				TLSClientConfig: config.tlsConfig,
				SASL:            config.auth, // if nil, it will be PLAIN taken from url
				Dial: func(network, addr string) (net.Conn, error) {
					return config.dialer.DialTimeout(network, addr, config.timeout)
				},
			})
		if err == nil {
			client.conn = conn
			config.log.Debugf("Connected to %q", broker)
			break
		}
		config.log.Debugf("Error connecting to %q - %v", broker, err.Error())
	}

	if client.conn == nil {
		return nil, errors.New("could not connect to any broker")
	}

	channel, err := client.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("error opening channel: %w", err)
	}
	client.channel = channel

	err = client.DeclareExchange()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *client) DeclareExchange() error {
	if c.config.exchange == "" {
		return nil
	}

	var err error
	if c.config.exchangePassive {
		err = c.channel.ExchangeDeclarePassive(
			c.config.exchange,
			c.config.exchangeType,
			c.config.exchangeDurable,
			false, // delete when unused
			false, // internal
			false, // no-wait
			c.config.exchangeArguments,
		)
	} else {
		err = c.channel.ExchangeDeclare(
			c.config.exchange,
			c.config.exchangeType,
			c.config.exchangeDurable,
			false, // delete when unused
			false, // internal
			false, // no-wait
			c.config.exchangeArguments,
		)
	}
	if err != nil {
		return fmt.Errorf("error declaring exchange: %w", err)
	}
	return nil
}

func (c *client) Publish(key string, body []byte) error {
	// Note that since the channel is not in confirm mode, the absence of
	// an error does not indicate successful delivery.
	return c.channel.PublishWithContext(
		context.Background(),
		c.config.exchange, // exchange
		key,               // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			Headers:         c.config.headers,
			ContentType:     "text/plain",
			ContentEncoding: c.config.encoding,
			Body:            body,
			DeliveryMode:    c.config.deliveryMode,
		})
}

func (c *client) Close() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	if err != nil && !errors.Is(err, amqp.ErrClosed) {
		return err
	}
	return nil
}
