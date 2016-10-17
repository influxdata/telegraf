package amqp

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/streadway/amqp"
)

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
	// InfluxDB precision
	Precision string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	channel *amqp.Channel
	sync.Mutex
	headers amqp.Table

	serializer serializers.Serializer
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
	DefaultPrecision       = "s"
)

var sampleConfig = `
  ## AMQP url
  url = "amqp://localhost:5672/influxdb"
  ## AMQP exchange
  exchange = "telegraf"
  ## Auth method. PLAIN and EXTERNAL are supported
  # auth_method = "PLAIN"
  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, it's value will be used as the routing key
  routing_tag = "host"

  ## InfluxDB retention policy
  # retention_policy = "default"
  ## InfluxDB database
  # database = "telegraf"
  ## InfluxDB precision
  # precision = "s"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (a *AMQP) SetSerializer(serializer serializers.Serializer) {
	a.serializer = serializer
}

func (q *AMQP) Connect() error {
	q.Lock()
	defer q.Unlock()

	q.headers = amqp.Table{
		"precision":        q.Precision,
		"database":         q.Database,
		"retention_policy": q.RetentionPolicy,
	}

	var connection *amqp.Connection
	// make new tls config
	tls, err := internal.GetTLSConfig(
		q.SSLCert, q.SSLKey, q.SSLCA, q.InsecureSkipVerify)
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
	q.channel = channel
	go func() {
		log.Printf("I! Closing: %s", <-connection.NotifyClose(make(chan *amqp.Error)))
		log.Printf("I! Trying to reconnect")
		for err := q.Connect(); err != nil; err = q.Connect() {
			log.Println("E! ", err.Error())
			time.Sleep(10 * time.Second)
		}

	}()
	return nil
}

func (q *AMQP) Close() error {
	return q.channel.Close()
}

func (q *AMQP) SampleConfig() string {
	return sampleConfig
}

func (q *AMQP) Description() string {
	return "Configuration for the AMQP server to send metrics to"
}

func (q *AMQP) Write(metrics []telegraf.Metric) error {
	q.Lock()
	defer q.Unlock()
	if len(metrics) == 0 {
		return nil
	}
	var outbuf = make(map[string][][]byte)

	for _, metric := range metrics {
		var key string
		if q.RoutingTag != "" {
			if h, ok := metric.Tags()[q.RoutingTag]; ok {
				key = h
			}
		}

		values, err := q.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		for _, value := range values {
			outbuf[key] = append(outbuf[key], []byte(value))
		}
	}

	for key, buf := range outbuf {
		err := q.channel.Publish(
			q.Exchange, // exchange
			key,        // routing key
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				Headers:     q.headers,
				ContentType: "text/plain",
				Body:        bytes.Join(buf, []byte("\n")),
			})
		if err != nil {
			return fmt.Errorf("FAILED to send amqp message: %s", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("amqp", func() telegraf.Output {
		return &AMQP{
			AuthMethod:      DefaultAuthMethod,
			Database:        DefaultDatabase,
			Precision:       DefaultPrecision,
			RetentionPolicy: DefaultRetentionPolicy,
		}
	})
}
