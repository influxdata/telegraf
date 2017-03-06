package amqp_consumer

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

// AMQPConsumer is the top level struct for this plugin
type AMQPConsumer struct {
	URL string
	// AMQP exchange
	Exchange string
	// Queue Name
	Queue string
	// Binding Key
	BindingKey string `toml:"binding_key"`

	// Controls how many messages the server will try to keep on the network
	// for consumers before receiving delivery acks.
	PrefetchCount int

	// AMQP Auth method
	AuthMethod string
	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	parser parsers.Parser
	conn   *amqp.Connection
	wg     *sync.WaitGroup
}

type externalAuth struct{}

func (a *externalAuth) Mechanism() string {
	return "EXTERNAL"
}
func (a *externalAuth) Response() string {
	return fmt.Sprintf("\000")
}

const (
	DefaultAuthMethod    = "PLAIN"
	DefaultPrefetchCount = 50
)

func (a *AMQPConsumer) SampleConfig() string {
	return `
  ## AMQP url
  url = "amqp://localhost:5672/influxdb"
  ## AMQP exchange
  exchange = "telegraf"
  ## AMQP queue name
  queue = "telegraf"
  ## Binding Key
  binding_key = "#"

  ## Maximum number of messages server should give to the worker.
  prefetch_count = 50

  ## Auth method. PLAIN and EXTERNAL are supported
  ## Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
  ## described here: https://www.rabbitmq.com/plugins.html
  # auth_method = "PLAIN"

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
}

func (a *AMQPConsumer) Description() string {
	return "AMQP consumer plugin"
}

func (a *AMQPConsumer) SetParser(parser parsers.Parser) {
	a.parser = parser
}

// All gathering is done in the Start function
func (a *AMQPConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (a *AMQPConsumer) createConfig() (*amqp.Config, error) {
	// make new tls config
	tls, err := internal.GetTLSConfig(
		a.SSLCert, a.SSLKey, a.SSLCA, a.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	// parse auth method
	var sasl []amqp.Authentication // nil by default

	if strings.ToUpper(a.AuthMethod) == "EXTERNAL" {
		sasl = []amqp.Authentication{&externalAuth{}}
	}

	config := amqp.Config{
		TLSClientConfig: tls,
		SASL:            sasl, // if nil, it will be PLAIN
	}
	return &config, nil
}

// Start satisfies the telegraf.ServiceInput interface
func (a *AMQPConsumer) Start(acc telegraf.Accumulator) error {
	amqpConf, err := a.createConfig()
	if err != nil {
		return err
	}

	msgs, err := a.connect(amqpConf)
	if err != nil {
		return err
	}

	a.wg = &sync.WaitGroup{}
	a.wg.Add(1)
	go a.process(msgs, acc)

	go func() {
		err := <-a.conn.NotifyClose(make(chan *amqp.Error))
		if err == nil {
			return
		}

		log.Printf("I! AMQP consumer connection closed: %s; trying to reconnect", err)
		for {
			msgs, err := a.connect(amqpConf)
			if err != nil {
				log.Printf("E! AMQP connection failed: %s", err)
				time.Sleep(10 * time.Second)
				continue
			}

			a.wg.Add(1)
			go a.process(msgs, acc)
			break
		}
	}()

	return nil
}

func (a *AMQPConsumer) connect(amqpConf *amqp.Config) (<-chan amqp.Delivery, error) {
	conn, err := amqp.DialConfig(a.URL, *amqpConf)
	if err != nil {
		return nil, err
	}
	a.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open a channel: %s", err)
	}

	err = ch.ExchangeDeclare(
		a.Exchange, // name
		"topic",    // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare an exchange: %s", err)
	}

	q, err := ch.QueueDeclare(
		a.Queue, // queue
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare a queue: %s", err)
	}

	err = ch.QueueBind(
		q.Name,       // queue
		a.BindingKey, // binding-key
		a.Exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to bind a queue: %s", err)
	}

	err = ch.Qos(
		a.PrefetchCount,
		0,     // prefetch-size
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to set QoS: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Failed establishing connection to queue: %s", err)
	}

	log.Println("I! Started AMQP consumer")
	return msgs, err
}

// Read messages from queue and add them to the Accumulator
func (a *AMQPConsumer) process(msgs <-chan amqp.Delivery, acc telegraf.Accumulator) {
	defer a.wg.Done()
	for d := range msgs {
		metrics, err := a.parser.Parse(d.Body)
		if err != nil {
			log.Printf("E! %v: error parsing metric - %v", err, string(d.Body))
		} else {
			for _, m := range metrics {
				acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
			}
		}

		d.Ack(false)
	}
	log.Printf("I! AMQP consumer queue closed")
}

func (a *AMQPConsumer) Stop() {
	err := a.conn.Close()
	if err != nil && err != amqp.ErrClosed {
		log.Printf("E! Error closing AMQP connection: %s", err)
		return
	}
	a.wg.Wait()
	log.Println("I! Stopped AMQP service")
}

func init() {
	inputs.Add("amqp_consumer", func() telegraf.Input {
		return &AMQPConsumer{
			AuthMethod:    DefaultAuthMethod,
			PrefetchCount: DefaultPrefetchCount,
		}
	})
}
