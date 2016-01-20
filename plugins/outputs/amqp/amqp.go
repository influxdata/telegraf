package amqp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/streadway/amqp"
)

type AMQP struct {
	// AMQP brokers to send metrics to
	URL string
	// AMQP exchange
	Exchange string
	// path to CA file
	SslCa string
	// path to host cert file
	SslCert string
	// path to cert key file
	SslKey string
	// Routing Key Tag
	RoutingTag string `toml:"routing_tag"`
	// InfluxDB database
	Database string
	// InfluxDB retention policy
	RetentionPolicy string
	// InfluxDB precision
	Precision string

	channel *amqp.Channel
	sync.Mutex
	headers amqp.Table
}

const (
	DefaultRetentionPolicy = "default"
	DefaultDatabase        = "telegraf"
	DefaultPrecision       = "s"
)

var sampleConfig = `
  # AMQP url
  url = "amqp://localhost:5672/influxdb"
  # AMQP exchange
  exchange = "telegraf"
  # Telegraf tag to use as a routing key
  #  ie, if this tag exists, it's value will be used as the routing key
  routing_tag = "host"

  # Use ssl
  #ssl_ca = "/etc/telegraf/ca.pem"
  #ssl_cert = "/etc/telegraf/cert.pem"
  #ssl_key = "/etc/telegraf/key.pem"

  # InfluxDB retention policy
  #retention_policy = "default"
  # InfluxDB database
  #database = "telegraf"
  # InfluxDB precision
  #precision = "s"
`

func (q *AMQP) Connect() error {
	q.Lock()
	defer q.Unlock()

	q.headers = amqp.Table{
		"precision":        q.Precision,
		"database":         q.Database,
		"retention_policy": q.RetentionPolicy,
	}

	var connection *amqp.Connection
	var err error
	if q.SslCert != "" && q.SslKey != "" {
		// make new tls config
		cfg := new(tls.Config)
		if q.SslCa != "" {
			// create ca pool
			cfg.RootCAs = x509.NewCertPool()

			// add self-signed cert
			if ca, err := ioutil.ReadFile(q.SslCa); err == nil {
				cfg.RootCAs.AppendCertsFromPEM(ca)
			} else {
				log.Println(err)
			}
		}
		if cert, err := tls.LoadX509KeyPair(q.SslCert, q.SslKey); err == nil {
			cfg.Certificates = append(cfg.Certificates, cert)
		} else {
			log.Println(err)
		}
		connection, err = amqp.DialTLS(q.URL, cfg)

	} else {
		connection, err = amqp.Dial(q.URL)
	}
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
		log.Printf("Closing: %s", <-connection.NotifyClose(make(chan *amqp.Error)))
		log.Printf("Trying to reconnect")
		for err := q.Connect(); err != nil; err = q.Connect() {
			log.Println(err)
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

func (q *AMQP) Write(points []*client.Point) error {
	q.Lock()
	defer q.Unlock()
	if len(points) == 0 {
		return nil
	}
	var outbuf = make(map[string][][]byte)

	for _, p := range points {
		// Combine tags from Point and BatchPoints and grab the resulting
		// line-protocol output string to write to AMQP
		var value, key string
		value = p.String()

		if q.RoutingTag != "" {
			if h, ok := p.Tags()[q.RoutingTag]; ok {
				key = h
			}
		}
		outbuf[key] = append(outbuf[key], []byte(value))

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
	outputs.Add("amqp", func() outputs.Output {
		return &AMQP{
			Database:        DefaultDatabase,
			Precision:       DefaultPrecision,
			RetentionPolicy: DefaultRetentionPolicy,
		}
	})
}
