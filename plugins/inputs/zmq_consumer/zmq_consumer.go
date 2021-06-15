package zmq_consumer

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	zmq "github.com/pebbe/zmq4"
)

type empty struct{}
type semaphore chan empty

type zmqConsumer struct {
	Endpoints     []string `toml:"endpoints"`
	Subscriptions []string `toml:"subscriptions"`

	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	Log telegraf.Logger

	socket *zmq.Socket
	parser parsers.Parser
	acc    telegraf.TrackingAccumulator

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

var sampleConfig = `
  ## ZeroMQ publisher endpoints
  # endpoints = ["tcp://localhost:6060"]

  ## Subscription filters
  # subscriptions = ["telegraf"]

  ## Maximum messages to read from the broker that have not been written by an
  ## output. For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

const (
	defaultMaxUndeliveredMessages = 1000
)

func (z *zmqConsumer) SampleConfig() string {
	return sampleConfig
}

func (z *zmqConsumer) Description() string {
	return "ZeroMQ consumer plugin"
}

func (n *zmqConsumer) SetParser(parser parsers.Parser) {
	n.parser = parser
}

func (z *zmqConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (z *zmqConsumer) Init() error {
	if z.MaxUndeliveredMessages == 0 {
		z.MaxUndeliveredMessages = defaultMaxUndeliveredMessages
	}
	if len(z.Endpoints) == 0 {
		return fmt.Errorf("missing publisher endpoints")
	}
	if len(z.Subscriptions) == 0 {
		return fmt.Errorf("missing subscription filters")
	}
	return nil
}

// Start the ZeroMQ consumer. Caller must call *zmqConsumer.Stop() to clean up.
func (z *zmqConsumer) Start(acc telegraf.Accumulator) error {
	z.acc = acc.WithTracking(z.MaxUndeliveredMessages)

	if z.socket == nil {
		// BEGIN TODO: extract to a separate func
		// create the socket
		socket, err := zmq.NewSocket(zmq.SUB)
		if err != nil {
			return err
		}
		z.socket = socket

		// set subscription filters
		for _, filter := range z.Subscriptions {
			z.socket.SetSubscribe(filter)
		}
		// connect to endpoints
		for _, endpoint := range z.Endpoints {
			z.socket.Connect(endpoint)
		}
		// END TODO: extract to a separate func

		ctx, cancel := context.WithCancel(context.Background())
		z.cancel = cancel

		// start the message reader
		z.wg.Add(1)
		go func() {
			defer z.wg.Done()
			go z.receiver(ctx)
		}()
	}
	return nil
}

// receiver() reads all published messages from the socket
func (z *zmqConsumer) receiver(ctx context.Context) {
	sem := make(semaphore, z.MaxUndeliveredMessages)

	for {
		select {
		case <-ctx.Done():
			return
		case <-z.acc.Delivered():
			<-sem
		case sem <- empty{}:
			select {
			case <-ctx.Done():
				return
			case <-z.acc.Delivered():
				<-sem
				<-sem
			default:
				msg, err := z.socket.Recv(zmq.DONTWAIT)
				if err != nil {
					if zmq.AsErrno(err) != zmq.Errno(syscall.EAGAIN) {
						z.Log.Errorf("Error receiving from socket: %v", err)
						<-sem
						continue
					}
				} else {
					metrics, err := z.parser.Parse([]byte(msg))
					fmt.Println(msg)
					if err != nil {
						z.Log.Errorf("Error parsing message: %v", err)
						<-sem
						continue
					}
					z.acc.AddTrackingMetricGroup(metrics)
				}
			}
		}
	}
}

func (z *zmqConsumer) Stop() {
	z.cancel()
	z.wg.Wait()
	z.socket.Close()
}

func init() {
	inputs.Add("zmq_consumer", func() telegraf.Input {
		return &zmqConsumer{}
	})
}
