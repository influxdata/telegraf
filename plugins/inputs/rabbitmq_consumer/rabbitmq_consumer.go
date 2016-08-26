package rabbitmqConsumer

import (
	"fmt"
	"log"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/streadway/amqp"
)

// RabbitMQConsumer is the top level struct for this plugin
type RabbitMQConsumer struct {
	Username     string
	Password     string
	RabbitmqHost string
	RabbitmqPort string
	Queue        string

	sync.Mutex

	parser parsers.Parser
	conn   *amqp.Connection
	ch     *amqp.Channel
	q      amqp.Queue
}

// Description satisfies the telegraf.ServiceInput interface
func (rmq *RabbitMQConsumer) Description() string {
	return "RabbitMQ consumer plugin"
}

// SampleConfig satisfies the telegraf.ServiceInput interface
func (rmq *RabbitMQConsumer) SampleConfig() string {
	return `
  # The following options form a connection string to rabbitmq:
  # amqp://{username}:{password}@{rabbitmq_host}:{rabbitmq_port}
  username = "guest"
  password = "guest"
  rabbitmq_host = "localhost"
  rabbitmq_port = "5672"
  # name of the queue to consume from
  queue = "task_queue"

  data_format = "influx"
`
}

// SetParser satisfies the telegraf.ServiceInput interface
func (rmq *RabbitMQConsumer) SetParser(parser parsers.Parser) {
	rmq.parser = parser
}

// Gather satisfies the telegraf.ServiceInput interface
// All gathering is done in the Start function
func (rmq *RabbitMQConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Start satisfies the telegraf.ServiceInput interface
func (rmq *RabbitMQConsumer) Start(acc telegraf.Accumulator) error {

	// Create queue connection and assign it to RabbitMQConsumer
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%v:%v@%v:%v", rmq.Username, rmq.Password, rmq.RabbitmqHost, rmq.RabbitmqPort))
	if err != nil {
		return fmt.Errorf("%v: Failed to connect to RabbitMQ", err)
	}
	rmq.conn = conn

	// Create channel and assign it to RabbitMQConsumer
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("%v: Failed to open a channel", err)
	}
	rmq.ch = ch

	// Declare a queue and assign it to RabbitMQConsumer
	q, err := ch.QueueDeclare(rmq.Queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%v: Failed to declare a queue", err)
	}
	rmq.q = q

	// Declare QoS on queue
	err = ch.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("%v: failed to set Qos", err)
	}

	// Register the RabbitMQ parser as a consumer of the queue
	// And start the lister passing in the Accumulator
	msgs := rmq.registerConsumer()
	go rmq.listen(msgs, acc)

	// Log that service has started
	log.Println("Starting RabbitMQ service...")
	return nil
}

// registerConsumer registers the consumer with the RabbitMQ broker
func (rmq *RabbitMQConsumer) registerConsumer() <-chan amqp.Delivery {
	messages, err := rmq.ch.Consume(rmq.Queue, "", false, false, false, false, nil)
	if err != nil {
		panic(fmt.Errorf("%v: failed establishing connection to queue", err))
	}
	return messages
}

// listen(s) for new messages coming in from RabbitMQ and pawns them off to handleMessage
func (rmq *RabbitMQConsumer) listen(msgs <-chan amqp.Delivery, acc telegraf.Accumulator) {
	for d := range msgs {
		go handleMessage(d, acc, rmq.parser)
	}
}

// handleMessage parses the incoming messages and passes them to the Accumulator
func handleMessage(d amqp.Delivery, acc telegraf.Accumulator, parser parsers.Parser) {
	metric, err := parser.Parse(d.Body)
	if err != nil {
		log.Fatalf("%v: error parsing metric - %v", err, string(d.Body))
	}
	for _, m := range metric {
		acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
	d.Ack(false)
}

// Stop satisfies the telegraf.ServiceInput interface
func (rmq *RabbitMQConsumer) Stop() {
	rmq.Lock()
	defer rmq.Unlock()
	rmq.conn.Close()
	rmq.ch.Close()
}

func init() {
	inputs.Add("rabbitmq_consumer", func() telegraf.Input {
		return &RabbitMQConsumer{}
	})
}
