package rabbitmq

import (
	"fmt"
	"log"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/streadway/amqp"
)

// RabbitMQ is the top level object for this output plugin
type RabbitMQ struct {
	Username     string
	Password     string
	RabbitmqHost string
	RabbitmqPort string
	Queue        string

	sync.Mutex

	serializer serializers.Serializer
	conn       *amqp.Connection
	ch         *amqp.Channel
	q          amqp.Queue
}

func (rmq *RabbitMQ) publish(body string) error {
	err := rmq.ch.Publish("", rmq.q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         []byte(body),
	})
	return err
}

// SetSerializer enables different output formats
func (rmq *RabbitMQ) SetSerializer(serializer serializers.Serializer) {
	rmq.serializer = serializer
}

// Description describes the plugin
func (rmq *RabbitMQ) Description() string {
	return "An output for publishing to RabbitMQ"
}

// SampleConfig defines the configuration options for the plugin
func (rmq *RabbitMQ) SampleConfig() string {
	return `
  # The following options form a connection string to rabbitmq:
  # amqp://{username}:{password}@{rabbitmq_host}:{rabbitmq_port}
  username = "guest"
  password = "guest"
  rabbitmq_host = "localhost"
  rabbitmq_port = "5672"
  # name of the queue to publish to 
  queue = "task_queue"

  data_format = "influx"
`
}

// Connect starts the connection with RabbitMQ
func (rmq *RabbitMQ) Connect() error {
	// Declare and store the connection to RabbitMQ
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%v:%v@%v:%v", rmq.Username, rmq.Password, rmq.RabbitmqHost, rmq.RabbitmqPort))
	if err != nil {
		return fmt.Errorf("%v: failed to connect to RabbitMQ", err)
	}
	rmq.conn = conn

	// Declare and store the sending channel for RabbitMQ
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("%v: Failed to open a channel", err)
	}
	rmq.ch = ch

	// Declare and store the queue
	q, err := ch.QueueDeclare(rmq.Queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%v: Failed to declare a queue", err)
	}
	rmq.q = q

	log.Printf("Connection established to RabbitMQ server at %v:%v", rmq.RabbitmqHost, rmq.RabbitmqPort)
	return nil
}

// Close is called to close the connection
func (rmq *RabbitMQ) Close() error {
	rmq.Lock()
	defer rmq.Unlock()
	rmq.conn.Close()
	rmq.ch.Close()
	log.Println("RabbitMQ connection closed")
	return nil
}

// Write defines how metrics are written
func (rmq *RabbitMQ) Write(metrics []telegraf.Metric) error {
	// If there are no metrics to write do nothing
	if len(metrics) == 0 {
		return nil
	}
	// Iterate over metrics collection
	for _, metric := range metrics {
		// Serialize metric
		points, err := rmq.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("%v: error serializing metric", err)
		}
		// Iterate over the points
		for _, point := range points {
			// Publish each point
			err := rmq.publish(point)
			if err != nil {
				return fmt.Errorf("Failed to write %v, %v", point, err)
			}
		}
	}
	return nil
}

func init() {
	outputs.Add("rabbitmq", func() telegraf.Output {
		return &RabbitMQ{}
	})
}
