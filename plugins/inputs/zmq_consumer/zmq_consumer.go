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
	defaultHighWaterMark          = 1000
	defaultMaxUndeliveredMessages = 1000
	defaultSubscription           = ""
	channelBufferSize             = 1000
)

type empty struct{}
type semaphore chan empty

type zmqConsumer struct {
	Endpoints     []string `toml:"endpoints"`
	Subscriptions []string `toml:"subscriptions"`
	HighWaterMark int      `toml:"high_water_mark"`
	Affinity      int      `toml:"affinity"`
	BufferSize    int      `toml:"receive_buffer_size"`

	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	Log telegraf.Logger

	in chan string

	parser parsers.Parser
	acc    telegraf.TrackingAccumulator

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

var sampleConfig = `
  ## ZeroMQ publisher endpoint urls.
  # endpoints = ["tcp://localhost:6001", "tcp://localhost:6002"]

  ## Subscription filters. If not specified the plugin  will subscribe 
  ## to all incoming  messages.
  # subscriptions = ["telegraf"]

  ## High water mark for inbound messages. Sets the ZMQ_RCVHWM option 
  ## on the specified socket. 
  ## The default value is 1000.
  ## See: http://api.zeromq.org/4-1:zmq-setsockopt#toc28
  # high_water_mark = 1000

  ## I/O thread affinity
  ## Affinity determines which threads from the ØMQ I/O thread pool 
  ## associated with the socket's context shall handle newly created 
  ## connections. A value of zero specifies no affinity, meaning that 
  ## work shall be distributed fairly among all ØMQ I/O threads in the 
  ## thread pool. For non-zero values, the lowest bit corresponds to 
  ## thread 1, second lowest bit to thread 2 and so on.
  ## The default value is 0.
  ## See: http://api.zeromq.org/4-1:zmq-setsockopt#toc3
  # affinity = 0
  
  ## Kernel receive buffer size
  ## Sets the underlying kernel receive buffer size for the socket to the 
  ## specified size in bytes. A value of zero means leave the OS default
  ## unchanged.
  ## The default value is 0.
  ## See: http://api.zeromq.org/4-1:zmq-setsockopt#toc27
  # receive_buffer_size = 0

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
	if len(z.Endpoints) == 0 {
		return fmt.Errorf("missing publisher endpoints")
	}
	if len(z.Subscriptions) == 0 {
		z.Subscriptions = append(z.Subscriptions, defaultSubscription)
	}
	if z.HighWaterMark == 0 {
		z.HighWaterMark = defaultHighWaterMark
	}
	if z.MaxUndeliveredMessages == 0 {
		z.MaxUndeliveredMessages = defaultMaxUndeliveredMessages
	}
	return nil
}

// Start the ZeroMQ consumer. Caller must call *zmqConsumer.Stop() to clean up.
func (z *zmqConsumer) Start(acc telegraf.Accumulator) error {
	z.acc = acc.WithTracking(z.MaxUndeliveredMessages)

	ctx, cancel := context.WithCancel(context.Background())
	z.cancel = cancel

	z.in = make(chan string, channelBufferSize)

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

	// set receive high water mark
	err = socket.SetRcvhwm(z.HighWaterMark)
	if err != nil {
		return nil, err
	}
	// set I/O thread affinity
	err = socket.SetAffinity(uint64(z.Affinity))
	if err != nil {
		return nil, err
	}
	// set kernel receive buffer size
	err = socket.SetRcvbuf(z.BufferSize)
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
			case msg, ok := <-z.in:
				if !ok {
					z.in = nil
					continue
				}
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
