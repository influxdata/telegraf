package amqp

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/streadway/amqp"
)

const (
	DefaultURL             = "amqp://localhost:5672/influxdb"
	DefaultAuthMethod      = "PLAIN"
	DefaultExchangeType    = "topic"
	DefaultRetentionPolicy = "default"
	DefaultDatabase        = "telegraf"
)

type externalAuth struct{}

func (a *externalAuth) Mechanism() string {
	return "EXTERNAL"
}

func (a *externalAuth) Response() string {
	return fmt.Sprintf("\000")
}

type AMQP struct {
	URL                string            `toml:"url"` // deprecated in 1.7; use brokers
	Brokers            []string          `toml:"brokers"`
	Exchange           string            `toml:"exchange"`
	ExchangeType       string            `toml:"exchange_type"`
	ExchangePassive    bool              `toml:"exchange_passive"`
	ExchangeDurability string            `toml:"exchange_durability"`
	ExchangeArguments  map[string]string `toml:"exchange_arguments"`
	Username           string            `toml:"username"`
	Password           string            `toml:"password"`
	MaxMessages        int               `toml:"max_messages"`
	AuthMethod         string            `toml:"auth_method"`
	RoutingTag         string            `toml:"routing_tag"`
	RoutingKey         string            `toml:"routing_key"`
	DeliveryMode       string            `toml:"delivery_mode"`
	Database           string            `toml:"database"`         // deprecated in 1.7; use headers
	RetentionPolicy    string            `toml:"retention_policy"` // deprecated in 1.7; use headers
	Precision          string            `toml:"precision"`        // deprecated; has no effect
	Headers            map[string]string `toml:"headers"`
	Timeout            internal.Duration `toml:"timeout"`
	UseBatchFormat     bool              `toml:"use_batch_format"`
	tls.ClientConfig

	serializer   serializers.Serializer
	connect      func(*ClientConfig) (Client, error)
	client       Client
	config       *ClientConfig
	sentMessages int
}

type Client interface {
	Publish(key string, body []byte) error
	Close() error
}

var sampleConfig = `
  ## Broker to publish to.
  ##   deprecated in 1.7; use the brokers option
  # url = "amqp://localhost:5672/influxdb"

  ## Brokers to publish to.  If multiple brokers are specified a random broker
  ## will be selected anytime a connection is established.  This can be
  ## helpful for load balancing when not using a dedicated load balancer.
  brokers = ["amqp://localhost:5672/influxdb"]

  ## Maximum messages to send over a connection.  Once this is reached, the
  ## connection is closed and a new connection is made.  This can be helpful for
  ## load balancing when not using a dedicated load balancer.
  # max_messages = 0

  ## Exchange to declare and publish to.
  exchange = "telegraf"

  ## Exchange type; common types are "direct", "fanout", "topic", "header", "x-consistent-hash".
  # exchange_type = "topic"

  ## If true, exchange will be passively declared.
  # exchange_declare_passive = false

  ## If true, exchange will be created as a durable exchange.
  # exchange_durable = true

  ## Additional exchange arguments.
  # exchange_arguments = { }
  # exchange_arguments = {"hash_propery" = "timestamp"}

  ## Authentication credentials for the PLAIN auth_method.
  # username = ""
  # password = ""

  ## Auth method. PLAIN and EXTERNAL are supported
  ## Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
  ## described here: https://www.rabbitmq.com/plugins.html
  # auth_method = "PLAIN"

  ## Metric tag to use as a routing key.
  ##   ie, if this tag exists, its value will be used as the routing key
  # routing_tag = "host"

  ## Static routing key.  Used when no routing_tag is set or as a fallback
  ## when the tag specified in routing tag is not found.
  # routing_key = ""
  # routing_key = "telegraf"

  ## Delivery Mode controls if a published message is persistent.
  ##   One of "transient" or "persistent".
  # delivery_mode = "transient"

  ## InfluxDB database added as a message header.
  ##   deprecated in 1.7; use the headers option
  # database = "telegraf"

  ## InfluxDB retention policy added as a message header
  ##   deprecated in 1.7; use the headers option
  # retention_policy = "default"

  ## Static headers added to each published message.
  # headers = { }
  # headers = {"database" = "telegraf", "retention_policy" = "default"}

  ## Connection timeout.  If not provided, will default to 5s.  0s means no
  ## timeout (not recommended).
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## If true use batch serialization format instead of line based delimiting.
  ## Only applies to data formats which are not line based such as JSON.
  ## Recommended to set to true.
  # use_batch_format = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
`

func (q *AMQP) SampleConfig() string {
	return sampleConfig
}

func (q *AMQP) Description() string {
	return "Publishes metrics to an AMQP broker"
}

func (q *AMQP) SetSerializer(serializer serializers.Serializer) {
	q.serializer = serializer
}

func (q *AMQP) Connect() error {
	if q.config == nil {
		config, err := q.makeClientConfig()
		if err != nil {
			return err
		}
		q.config = config
	}

	client, err := q.connect(q.config)
	if err != nil {
		return err
	}
	q.client = client

	return nil
}

func (q *AMQP) Close() error {
	if q.client != nil {
		return q.client.Close()
	}
	return nil
}

func (q *AMQP) routingKey(metric telegraf.Metric) string {
	if q.RoutingTag != "" {
		key, ok := metric.GetTag(q.RoutingTag)
		if ok {
			return key
		}
	}
	return q.RoutingKey
}

func (q *AMQP) Write(metrics []telegraf.Metric) error {
	batches := make(map[string][]telegraf.Metric)
	if q.ExchangeType == "direct" || q.ExchangeType == "header" {
		// Since the routing_key is ignored for these exchange types send as a
		// single batch.
		batches[""] = metrics
	} else {
		for _, metric := range metrics {
			routingKey := q.routingKey(metric)
			if _, ok := batches[routingKey]; !ok {
				batches[routingKey] = make([]telegraf.Metric, 0)
			}

			batches[routingKey] = append(batches[routingKey], metric)
		}
	}

	first := true
	for key, metrics := range batches {
		body, err := q.serialize(metrics)
		if err != nil {
			return err
		}

		err = q.publish(key, body)
		if err != nil {
			// If this is the first attempt to publish and the connection is
			// closed, try to reconnect and retry once.
			if aerr, ok := err.(*amqp.Error); first && ok && aerr == amqp.ErrClosed {
				first = false
				q.client = nil
				err := q.publish(key, body)
				if err != nil {
					return err
				}
			} else {
				q.client = nil
				return err
			}
		}
		first = false
	}

	if q.sentMessages >= q.MaxMessages && q.MaxMessages > 0 {
		log.Printf("D! Output [amqp] sent MaxMessages; closing connection")
		q.client = nil
	}

	return nil
}

func (q *AMQP) publish(key string, body []byte) error {
	if q.client == nil {
		client, err := q.connect(q.config)
		if err != nil {
			return err
		}
		q.sentMessages = 0
		q.client = client
	}

	err := q.client.Publish(key, body)
	if err != nil {
		return err
	}
	q.sentMessages++
	return nil
}

func (q *AMQP) serialize(metrics []telegraf.Metric) ([]byte, error) {
	if q.UseBatchFormat {
		return q.serializer.SerializeBatch(metrics)
	} else {
		var buf bytes.Buffer
		for _, metric := range metrics {
			octets, err := q.serializer.Serialize(metric)
			if err != nil {
				return nil, err
			}
			_, err = buf.Write(octets)
			if err != nil {
				return nil, err
			}
		}
		body := buf.Bytes()
		return body, nil
	}
}

func (q *AMQP) makeClientConfig() (*ClientConfig, error) {
	config := &ClientConfig{
		exchange:        q.Exchange,
		exchangeType:    q.ExchangeType,
		exchangePassive: q.ExchangePassive,
		timeout:         q.Timeout.Duration,
	}

	switch q.ExchangeDurability {
	case "transient":
		config.exchangeDurable = false
	default:
		config.exchangeDurable = true
	}

	config.brokers = q.Brokers
	if len(config.brokers) == 0 {
		config.brokers = []string{q.URL}
	}

	switch q.DeliveryMode {
	case "transient":
		config.deliveryMode = amqp.Transient
	case "persistent":
		config.deliveryMode = amqp.Persistent
	default:
		config.deliveryMode = amqp.Transient
	}

	if len(q.Headers) > 0 {
		config.headers = make(amqp.Table, len(q.Headers))
		for k, v := range q.Headers {
			config.headers[k] = v
		}
	} else {
		// Copy deprecated fields into message header
		config.headers = amqp.Table{
			"database":         q.Database,
			"retention_policy": q.RetentionPolicy,
		}
	}

	if len(q.ExchangeArguments) > 0 {
		config.exchangeArguments = make(amqp.Table, len(q.ExchangeArguments))
		for k, v := range q.ExchangeArguments {
			config.exchangeArguments[k] = v
		}
	}

	tlsConfig, err := q.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	config.tlsConfig = tlsConfig

	var auth []amqp.Authentication
	if strings.ToUpper(q.AuthMethod) == "EXTERNAL" {
		auth = []amqp.Authentication{&externalAuth{}}
	} else if q.Username != "" || q.Password != "" {
		auth = []amqp.Authentication{
			&amqp.PlainAuth{
				Username: q.Username,
				Password: q.Password,
			},
		}
	}
	config.auth = auth

	return config, nil
}

func connect(config *ClientConfig) (Client, error) {
	return Connect(config)
}

func init() {
	outputs.Add("amqp", func() telegraf.Output {
		return &AMQP{
			URL:             DefaultURL,
			ExchangeType:    DefaultExchangeType,
			AuthMethod:      DefaultAuthMethod,
			Database:        DefaultDatabase,
			RetentionPolicy: DefaultRetentionPolicy,
			Timeout:         internal.Duration{Duration: time.Second * 5},
			connect:         connect,
		}
	})
}
