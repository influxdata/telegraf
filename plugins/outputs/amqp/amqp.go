package amqp

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/streadway/amqp"
)

type client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	headers amqp.Table
}

type AMQP struct {
	// AMQP brokers to send metrics to
	URL string
	// AMQP exchange
	Exchange string
	// AMQP Auth method
	AuthMethod string
	// Routing Key Tag
	RoutingTag string `toml:"routing_tag"`
	// InfluxDB database
	Database string
	// InfluxDB retention policy
	RetentionPolicy string
	// InfluxDB precision (DEPRECATED)
	Precision string
	// Connection timeout
	Timeout internal.Duration
	// Delivery Mode controls if a published message is persistent
	// Valid options are "transient" and "persistent". default: "transient"
	DeliveryMode string

	tls.ClientConfig

	sync.Mutex
	c *client

	deliveryMode uint8
	serializer   serializers.Serializer
}

type externalAuth struct{}

func (a *externalAuth) Mechanism() string {
	return "EXTERNAL"
}
func (a *externalAuth) Response() string {
	return fmt.Sprintf("\000")
}

const (
	DefaultAuthMethod      = "PLAIN"
	DefaultRetentionPolicy = "default"
	DefaultDatabase        = "telegraf"
)

var sampleConfig = `
  ## AMQP url
  url = "amqp://localhost:5672/influxdb"
  ## AMQP exchange
  exchange = "telegraf"
  ## Auth method. PLAIN and EXTERNAL are supported
  ## Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
  ## described here: https://www.rabbitmq.com/plugins.html
  # auth_method = "PLAIN"
  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, its value will be used as the routing key
  routing_tag = "host"
  ## Delivery Mode controls if a published message is persistent
  ## Valid options are "transient" and "persistent". default: "transient"
  delivery_mode = "transient"

  ## InfluxDB retention policy
  # retention_policy = "default"
  ## InfluxDB database
  # database = "telegraf"

  ## Write timeout, formatted as a string.  If not provided, will default
  ## to 5s. 0s means no timeout (not recommended).
  # timeout = "5s"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (a *AMQP) SetSerializer(serializer serializers.Serializer) {
	a.serializer = serializer
}

func (q *AMQP) Connect() error {
	switch q.DeliveryMode {
	case "transient":
		q.deliveryMode = amqp.Transient
		break
	case "persistent":
		q.deliveryMode = amqp.Persistent
		break
	default:
		q.deliveryMode = amqp.Transient
		break
	}

	headers := amqp.Table{
		"database":         q.Database,
		"retention_policy": q.RetentionPolicy,
	}

	var connection *amqp.Connection
	// make new tls config
	tls, err := q.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	// parse auth method
	var sasl []amqp.Authentication // nil by default

	if strings.ToUpper(q.AuthMethod) == "EXTERNAL" {
		sasl = []amqp.Authentication{&externalAuth{}}
	}

	amqpConf := amqp.Config{
		TLSClientConfig: tls,
		SASL:            sasl, // if nil, it will be PLAIN
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, q.Timeout.Duration)
		},
	}

	connection, err = amqp.DialConfig(q.URL, amqpConf)
	if err != nil {
		return err
	}

	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("Failed to open a channel: %s", err)
	}

	err = channel.ExchangeDeclare(
		q.Exchange, // name
		"topic",    // type
		true,       // durable
		false,      // delete when unused
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare an exchange: %s", err)
	}

	q.setClient(&client{
		conn:    connection,
		channel: channel,
		headers: headers,
	})

	go func() {
		err := <-connection.NotifyClose(make(chan *amqp.Error))
		if err == nil {
			return
		}

		q.setClient(nil)

		log.Printf("I! Closing: %s", err)
		log.Printf("I! Trying to reconnect")
		for err := q.Connect(); err != nil; err = q.Connect() {
			log.Println("E! ", err.Error())
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

func (q *AMQP) Close() error {
	c := q.getClient()
	if c == nil {
		return nil
	}

	err := c.conn.Close()
	if err != nil && err != amqp.ErrClosed {
		log.Printf("E! Error closing AMQP connection: %s", err)
		return err
	}
	return nil
}

func (q *AMQP) SampleConfig() string {
	return sampleConfig
}

func (q *AMQP) Description() string {
	return "Configuration for the AMQP server to send metrics to"
}

func (q *AMQP) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	c := q.getClient()
	if c == nil {
		return fmt.Errorf("connection is not open")
	}

	outbuf := make(map[string][]byte)

	for _, metric := range metrics {
		var key string
		if q.RoutingTag != "" {
			if h, ok := metric.Tags()[q.RoutingTag]; ok {
				key = h
			}
		}

		buf, err := q.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		outbuf[key] = append(outbuf[key], buf...)
	}

	for key, buf := range outbuf {
		// Note that since the channel is not in confirm mode, the absence of
		// an error does not indicate successful delivery.
		err := c.channel.Publish(
			q.Exchange, // exchange
			key,        // routing key
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				Headers:      c.headers,
				ContentType:  "text/plain",
				Body:         buf,
				DeliveryMode: q.deliveryMode,
			})
		if err != nil {
			return fmt.Errorf("Failed to send AMQP message: %s", err)
		}
	}
	return nil
}

func (q *AMQP) getClient() *client {
	q.Lock()
	defer q.Unlock()
	return q.c
}

func (q *AMQP) setClient(c *client) {
	q.Lock()
	q.c = c
	q.Unlock()
}

func init() {
	outputs.Add("amqp", func() telegraf.Output {
		return &AMQP{
			AuthMethod:      DefaultAuthMethod,
			Database:        DefaultDatabase,
			RetentionPolicy: DefaultRetentionPolicy,
			Timeout:         internal.Duration{Duration: time.Second * 5},
		}
	})
}
