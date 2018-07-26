package amqp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/streadway/amqp"
)

type ClientConfig struct {
	brokers           []string
	exchange          string
	exchangeType      string
	exchangePassive   bool
	exchangeDurable   bool
	exchangeArguments amqp.Table
	headers           amqp.Table
	deliveryMode      uint8
	tlsConfig         *tls.Config
	timeout           time.Duration
	auth              []amqp.Authentication
}

type client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *ClientConfig
}

// Connect opens a connection to one of the brokers at random
func Connect(config *ClientConfig) (*client, error) {
	client := &client{
		config: config,
	}

	p := rand.Perm(len(config.brokers))
	for _, n := range p {
		broker := config.brokers[n]
		log.Printf("D! Output [amqp] connecting to %q", broker)
		conn, err := amqp.DialConfig(
			broker, amqp.Config{
				TLSClientConfig: config.tlsConfig,
				SASL:            config.auth, // if nil, it will be PLAIN taken from url
				Dial: func(network, addr string) (net.Conn, error) {
					return net.DialTimeout(network, addr, config.timeout)
				},
			})
		if err == nil {
			client.conn = conn
			log.Printf("D! Output [amqp] connected to %q", broker)
			break
		}
		log.Printf("D! Output [amqp] error connecting to %q", broker)
	}

	if client.conn == nil {
		return nil, errors.New("could not connect to any broker")
	}

	channel, err := client.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("error opening channel: %v", err)
	}
	client.channel = channel

	err = client.DeclareExchange()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *client) DeclareExchange() error {
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
		return fmt.Errorf("error declaring exchange: %v", err)
	}
	return nil
}

func (c *client) Publish(key string, body []byte) error {
	// Note that since the channel is not in confirm mode, the absence of
	// an error does not indicate successful delivery.
	return c.channel.Publish(
		c.config.exchange, // exchange
		key,               // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			Headers:      c.config.headers,
			ContentType:  "text/plain",
			Body:         body,
			DeliveryMode: c.config.deliveryMode,
		})
}

func (c *client) Close() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	if err != nil && err != amqp.ErrClosed {
		return err
	}
	return nil
}
