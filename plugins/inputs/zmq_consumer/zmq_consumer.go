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

const (
	defaultMaxUndeliveredMessages = 1000
	socketBufferSize              = 10
)

type empty struct{}
type semaphore chan empty

type zmqConsumer struct {
	Endpoints     []string `toml:"endpoints"`
	Subscriptions []string `toml:"subscriptions"`

	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	Log telegraf.Logger

	in chan string

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

	ctx, cancel := context.WithCancel(context.Background())
	z.cancel = cancel

	z.in = make(chan string, socketBufferSize)

	// start the message subscriber
	z.wg.Add(1)
	go func() {
		defer z.wg.Done()
		go z.subscriber(ctx)
	}()

	// start the message receiver
	z.wg.Add(1)
	go func() {
		defer z.wg.Done()
		go z.receiver(ctx)
	}()

	return nil
}

func (z *zmqConsumer) Stop() {
	z.cancel()
	z.wg.Wait()
}

// connect subscribes to the publisher socket(s)
func (z *zmqConsumer) connect() (*zmq.Socket, error) {
	// create the socket
	socket, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return nil, err
	}

	// set subscription filters
	for _, filter := range z.Subscriptions {
		err = socket.SetSubscribe(filter)
		if err != nil {
			return nil, err
		}
	}
	// connect to endpoints
	for _, endpoint := range z.Endpoints {
		err = socket.Connect(endpoint)
		if err != nil {
			return nil, err
		}
	}

	return socket, nil
}

// subscriber receives messages from the socket
func (z *zmqConsumer) subscriber(ctx context.Context) {
	// connect to PUB socket(s)
	socket, err := z.connect()
	if err != nil {
		z.Log.Errorf("Error connecting to socket: %s", err.Error())
		return
	}
	defer socket.Close()

	for {
		select {
		case <-ctx.Done():
			close(z.in)
			return
		default:
			// non-blocking receive
			msg, err := socket.Recv(zmq.DONTWAIT)
			if err != nil {
				if zmq.AsErrno(err) != zmq.Errno(syscall.EAGAIN) {
					z.Log.Errorf("Error receiving from socket: %v", err)
				}
				continue
			}
			z.in <- msg
		}
	}
}

// receiver reads all published messages from the socket and converts them to metrics
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
			case msg := <-z.in:
				metrics, err := z.parser.Parse([]byte(msg))
				if err != nil {
					z.Log.Errorf("Error parsing message: %s", err.Error())
					<-sem
					continue
				}
				z.acc.AddTrackingMetricGroup(metrics)
			}
		}
	}
}

func init() {
	inputs.Add("zmq_consumer", func() telegraf.Input {
		return &zmqConsumer{}
	})
}
