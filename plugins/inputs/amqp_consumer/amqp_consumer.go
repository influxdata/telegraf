package amqpConsumer

import (
	"fmt"
	"log"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/streadway/amqp"
)

// AMQPConsumer is the top level struct for this plugin
type AMQPConsumer struct {
	Username string
	Password string
	AMQPHost string
	AMQPPort string
	Queue    string
	Prefetch int

	sync.Mutex

	parser parsers.Parser
	conn   *amqp.Connection
	ch     *amqp.Channel
	q      amqp.Queue
}

// Description satisfies the telegraf.ServiceInput interface
func (rmq *AMQPConsumer) Description() string {
	return "AMQP consumer plugin"
}

// SampleConfig satisfies the telegraf.ServiceInput interface
func (rmq *AMQPConsumer) SampleConfig() string {
	return `
  # The following options form a connection string to amqp:
  # amqp://{username}:{password}@{amqp_host}:{amqp_port}
  username = "guest"
  password = "guest"
  amqp_host = "localhost"
  amqp_port = "5672"
  # name of the queue to consume from
  queue = "task_queue"
	prefetch = 1000

  data_format = "influx"
`
}

// SetParser satisfies the telegraf.ServiceInput interface
func (rmq *AMQPConsumer) SetParser(parser parsers.Parser) {
	rmq.parser = parser
}

// Gather satisfies the telegraf.ServiceInput interface
// All gathering is done in the Start function
func (rmq *AMQPConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Start satisfies the telegraf.ServiceInput interface
func (rmq *AMQPConsumer) Start(acc telegraf.Accumulator) error {

	// Create queue connection and assign it to AMQPConsumer
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%v:%v@%v:%v", rmq.Username, rmq.Password, rmq.AMQPHost, rmq.AMQPPort))
	if err != nil {
		return fmt.Errorf("%v: Failed to connect to AMQP", err)
	}
	rmq.conn = conn

	// Create channel and assign it to AMQPConsumer
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("%v: Failed to open a channel", err)
	}
	rmq.ch = ch

	// Declare a queue and assign it to AMQPConsumer
	q, err := ch.QueueDeclare(rmq.Queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%v: Failed to declare a queue", err)
	}
	rmq.q = q

	// Declare QoS on queue
	err = ch.Qos(rmq.Prefetch, 0, false)
	if err != nil {
		return fmt.Errorf("%v: failed to set Qos", err)
	}

	// Register the AMQP parser as a consumer of the queue
	// And start the lister passing in the Accumulator
	msgs := rmq.registerConsumer()
	go rmq.listen(msgs, acc)

	// Log that service has started
	log.Println("Starting AMQP service...")
	return nil
}

// registerConsumer registers the consumer with the AMQP broker
func (rmq *AMQPConsumer) registerConsumer() <-chan amqp.Delivery {
	messages, err := rmq.ch.Consume(rmq.Queue, "", false, false, false, false, nil)
	if err != nil {
		panic(fmt.Errorf("%v: failed establishing connection to queue", err))
	}
	return messages
}

// listen(s) for new messages coming in from AMQP and pawns them off to handleMessage
func (rmq *AMQPConsumer) listen(msgs <-chan amqp.Delivery, acc telegraf.Accumulator) {
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
func (rmq *AMQPConsumer) Stop() {
	rmq.Lock()
	defer rmq.Unlock()
	rmq.conn.Close()
	rmq.ch.Close()
}

func init() {
	inputs.Add("amqp_consumer", func() telegraf.Input {
		return &AMQPConsumer{}
	})
}
