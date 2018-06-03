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

	Exchange           string            `toml:"exchange"`
	ExchangeType       string            `toml:"exchange_type"`
	ExchangeDurability string            `toml:"exchange_durability"`
	ExchangePassive    bool              `toml:"exchange_passive"`
	ExchangeArguments  map[string]string `toml:"exchange_arguments"`

	// AMQP Auth method
	AuthMethod string
	// Routing Key (static)
	RoutingKey string `toml:"routing_key"`
	// Routing Key from Tag
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
	DefaultAuthMethod = "PLAIN"

	DefaultExchangeType       = "topic"
	DefaultExchangeDurability = "durable"

	DefaultRetentionPolicy = "default"
	DefaultDatabase        = "telegraf"
)

var sampleConfig = `
  ## AMQP url
  url = "amqp://localhost:5672/influxdb"

  ## Exchange to declare and publish to.
  exchange = "telegraf"

  ## Exchange type; common types are "direct", "fanout", "topic", "header", "x-consistent-hash".
  # exchange_type = "topic"

  ## If true, exchange will be passively declared.
  # exchange_passive = false

  ## Exchange durability can be either "transient" or "durable".
  # exchange_durability = "durable"

  ## Additional exchange arguments.
  # exchange_args = { }
  # exchange_args = {"hash_propery" = "timestamp"}

  ## Auth method. PLAIN and EXTERNAL are supported
  ## Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
  ## described here: https://www.rabbitmq.com/plugins.html
  # auth_method = "PLAIN"
  ## Topic routing key
  # routing_key = ""
  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, its value will be used as the routing key
  ##  and override routing_key config even if defined
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

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
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

	var exchangeDurable = true
	switch q.ExchangeDurability {
	case "transient":
		exchangeDurable = false
	default:
		exchangeDurable = true
	}

	exchangeArgs := make(amqp.Table, len(q.ExchangeArguments))
	for k, v := range q.ExchangeArguments {
		exchangeArgs[k] = v
	}

	err = declareExchange(
		channel,
		q.Exchange,
		q.ExchangeType,
		q.ExchangePassive,
		exchangeDurable,
		exchangeArgs)
	if err != nil {
		return err
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

func declareExchange(
	channel *amqp.Channel,
	exchangeName string,
	exchangeType string,
	exchangePassive bool,
	exchangeDurable bool,
	exchangeArguments amqp.Table,
) error {
	var err error
	if exchangePassive {
		err = channel.ExchangeDeclarePassive(
			exchangeName,
			exchangeType,
			exchangeDurable,
			false, // delete when unused
			false, // internal
			false, // no-wait
			exchangeArguments,
		)
	} else {
		err = channel.ExchangeDeclare(
			exchangeName,
			exchangeType,
			exchangeDurable,
			false, // delete when unused
			false, // internal
			false, // no-wait
			exchangeArguments,
		)
	}
	if err != nil {
		return fmt.Errorf("error declaring exchange: %v", err)
	}
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
		if q.RoutingKey != "" {
			key = q.RoutingKey
		}
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
			AuthMethod:         DefaultAuthMethod,
			ExchangeType:       DefaultExchangeType,
			ExchangeDurability: DefaultExchangeDurability,
			Database:           DefaultDatabase,
			RetentionPolicy:    DefaultRetentionPolicy,
			Timeout:            internal.Duration{Duration: time.Second * 5},
		}
	})
}
