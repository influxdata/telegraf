package poller

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal/config"

	// needing to load the plugins
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	// needing to load the outputs
	_ "github.com/influxdata/telegraf/plugins/outputs/all"

	"github.com/streadway/amqp"
)

func AMQPcreation() {
	// Connect
	connection, _ := amqp.Dial("amqp://guest:guest@127.0.0.1:5673/")
	// Channel
	channel, _ := connection.Channel()

	// SPLIT_BY_TAGS_EXCHANGE
	channel.ExchangeDeclare(
		"SPLIT_BY_TAGS_EXCHANGE", // name
		"headers",                // type
		true,                     // durable
		false,                    // delete when unused
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	// ENTRANCE
	channel.ExchangeDeclare(
		"ENTRANCE", // name
		"fanout",   // type
		true,       // durable
		false,      // delete when unused
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	channel.ExchangeBind(
		"SPLIT_BY_TAGS_EXCHANGE",
		"",
		"ENTRANCE",
		false,
		nil,
	)
	//REQUEUE_AFTER_WAIT_EXCHANGE
	channel.ExchangeDeclare(
		"REQUEUE_AFTER_WAIT_EXCHANGE",
		"fanout", // type
		true,     // durable
		false,    // delete when unused
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	channel.ExchangeBind(
		"SPLIT_BY_TAGS_EXCHANGE",
		"",
		"REQUEUE_AFTER_WAIT_EXCHANGE",
		false,
		nil,
	)
	//SPLIT_BY_INTERVAL_EXCHANGE
	channel.ExchangeDeclare(
		"SPLIT_BY_INTERVAL_EXCHANGE", // name
		"headers",                    // type
		true,                         // durable
		false,                        // delete when unused
		false,                        // internal
		false,                        // no-wait
		nil,                          // arguments
	)
	// QUEUE wait_queue_5s
	args := amqp.Table{}
	args["x-dead-letter-exchange"] = "REQUEUE_AFTER_WAIT_EXCHANGE"
	args["x-message-ttl"] = int64(5000)
	channel.QueueDeclare(
		"wait_queue_5s",
		true,
		false,
		false,
		false,
		args,
	)
	// Purge queue
	channel.QueuePurge("wait_queue_5s", false)
	args = amqp.Table{}
	args["interval"] = "5000"
	args["x-match"] = "all"
	channel.QueueBind(
		"wait_queue_5s",
		"",
		"SPLIT_BY_INTERVAL_EXCHANGE",
		false,
		args,
	)
	// QUEUE label1
	args = amqp.Table{}
	args["x-dead-letter-exchange"] = "SPLIT_BY_INTERVAL_EXCHANGE"
	channel.QueueDeclare(
		"label1",
		true,
		false,
		false,
		false,
		args,
	)
	// Purge queue
	channel.QueuePurge("label1", false)
	args = amqp.Table{}
	args["label1"] = "value1"
	args["x-match"] = "all"
	channel.QueueBind(
		"label1",
		"",
		"SPLIT_BY_TAGS_EXCHANGE",
		false,
		args,
	)
	// Create tasks
	msg := amqp.Publishing{
		Headers: amqp.Table{
			"interval": "5000",
			"label1":   "value1",
		},
		Body: []byte("[[inputs.mem]]"),
	}
	channel.Publish(
		"ENTRANCE",
		"",
		false,
		false,
		msg,
	)
}

func TestPoller_Reconnection(t *testing.T) {
	// Remove old file
	os.Remove("/tmp/metrics.out")
	// Create conf
	c := config.NewConfig()
	err := c.LoadConfig("../internal/config/testdata/telegraf-poller-bad.toml")
	assert.NoError(t, err)
	p, _ := NewPoller(c)
	// Connect output
	p.Connect()
	// Prepare shutdown
	shutdown := make(chan struct{})
	go func(sdchan chan struct{}) {
		log.Println("Waiting 10s before shuting down poller")
		time.Sleep(time.Duration(5) * time.Second)
		sdchan <- struct{}{}
		close(sdchan)
	}(shutdown)
	err = p.Run(shutdown)
	assert.NoError(t, err)
	p.Config.Outputs[0].Output.Close()
	// Check output
	f, err := os.Open("/tmp/metrics.out")
	assert.NoError(t, err)
	//if err != nil {
	finfo, _ := f.Stat()
	assert.True(t, finfo.Size() == int64(0))
	f.Close()
	//}

}

func TestPoller_Polling(t *testing.T) {
	// Remove old file
	os.Remove("/tmp/metrics.out")
	// Create AMQP flow
	AMQPcreation()
	// Create conf
	c := config.NewConfig()
	err := c.LoadConfig("../internal/config/testdata/telegraf-poller.toml")
	assert.NoError(t, err)
	p, _ := NewPoller(c)
	// Connect output
	p.Connect()
	// Prepare shutdown
	shutdown := make(chan struct{})
	go func(sdchan chan struct{}) {
		log.Println("Waiting 10s before shuting down poller")
		time.Sleep(time.Duration(10) * time.Second)
		//	time.Sleep(time.Duration(5) * time.Second)
		sdchan <- struct{}{}
		//p.Config.Outputs[0].Output.Close()
		//	p.Close()
		close(sdchan)
	}(shutdown)
	err = p.Run(shutdown)
	assert.NoError(t, err)
	p.Config.Outputs[0].Output.Close()
	// Check output
	f, _ := os.Open("/tmp/metrics.out")
	finfo, _ := f.Stat()
	assert.True(t, finfo.Size() > int64(0))
	f.Close()
}
