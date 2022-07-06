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
	defaultMethod                 = "bind"
	defaultHighWaterMark          = 1000
	defaultReceiveBufferSize      = -1
	defaultTcpKeepAlive           = -1
	defaultTcpKeepAliveIdle       = -1
	defaultTcpKeepAliveInterval   = -1
	defaultMaxUndeliveredMessages = 1000
	channelBufferSize             = 1000
)

const (
	bind  string = "bind"
	connect      = "connect"
)

type empty struct{}
type semaphore chan empty

type zmqConsumer struct {
	Method               string   `toml:"method"`
	Endpoints            []string `toml:"endpoints"`
	Subscriptions        []string `toml:"subscriptions"`
	HighWaterMark        int      `toml:"high_water_mark"`
	Affinity             int      `toml:"affinity"`
	ReceiveBufferSize    int      `toml:"receive_buffer_size"`
	TcpKeepAlive         int      `toml:"tcp_keepalive"`
	TcpKeepAliveIdle     int      `toml:"tcp_keepalive_idle"`
	TcpKeepAliveInterval int      `toml:"tcp_keepalive_interval"`

	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	Log telegraf.Logger

	in chan string

	parser parsers.Parser
	acc    telegraf.TrackingAccumulator

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

var sampleConfig = `
  ## ZeroMQ PUB/SUB connection type (either bind or connect, defaults to bind)
  # method = bind

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

  ## Overrides SO_KEEPALIVE socket option (where supported by OS). 
  ## The default value of -1 means to skip any overrides and leave it to 
  ## OS default.
  ## See: http://api.zeromq.org/master:zmq-setsockopt#toc57
  # tcp_keepalive = -1

  ## Overrides TCP_KEEPIDLE (or TCP_KEEPALIVE on some OS) socket option 
  ## (where supported by OS). 
  ## The default value of -1 means to skip any overrides and leave it to 
  ## OS default.
  ## See: http://api.zeromq.org/master:zmq-setsockopt#toc59
  # tcp_keepalive_idle = -1

  ## Overrides TCP_KEEPINTVL socket option (where supported by OS). 
  ## The default value of -1 means to skip any overrides and leave it to 
  ## OS default.
  ## See: http://api.zeromq.org/master:zmq-setsockopt#toc60
  # tcp_keepalive_interval = -1

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
	if len(z.Method) == 0 {
		z.Method = bind;
	}

	if z.Method != bind && z.Method != connect {
		return fmt.Errorf("unknown endpoint connection method: %s", z.Method)
	}

	if len(z.Endpoints) == 0 {
		return fmt.Errorf("missing publisher endpoints")
	}

	if z.HighWaterMark == 0 {
		z.HighWaterMark = defaultHighWaterMark
	}
	if z.ReceiveBufferSize == 0 {
		z.ReceiveBufferSize = defaultReceiveBufferSize
	}
	if z.TcpKeepAlive == 0 {
		z.TcpKeepAlive = defaultTcpKeepAlive
	}
	if z.TcpKeepAliveIdle == 0 {
		z.TcpKeepAliveIdle = defaultTcpKeepAliveIdle
	}
	if z.TcpKeepAliveInterval == 0 {
		z.TcpKeepAliveInterval = defaultTcpKeepAliveInterval
	}
	if z.MaxUndeliveredMessages == 0 {
		z.MaxUndeliveredMessages = defaultMaxUndeliveredMessages
	}

	z.Log.Infof("%sing to endpoints %s", z.Method, z.Endpoints)
	if len(z.Subscriptions) == 0 {
		z.Log.Infof("no subscriptions specified: subscribing to empty string (wildcard)")
		z.Subscriptions = append(z.Subscriptions, "")
	}else{
		z.Log.Infof("subscriptions: %s", z.Subscriptions)
	}

	z.Log.Debugf("hwm: %d",z.HighWaterMark)
	z.Log.Debugf("recv buf: %d",z.ReceiveBufferSize)
	z.Log.Debugf("keep-alive: %d",z.TcpKeepAlive)
	z.Log.Debugf("keep-alive idle: %d", z.TcpKeepAliveIdle)
	z.Log.Debugf("keep-alive interval: %d", z.TcpKeepAliveInterval)
	z.Log.Debugf("max undelivered: %d", z.MaxUndeliveredMessages)

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
	err = socket.SetRcvbuf(z.ReceiveBufferSize)
	if err != nil {
		return nil, err
	}
	// set TCP keep alive
	err = socket.SetTcpKeepalive(z.TcpKeepAlive)
	if err != nil {
		return nil, err
	}
	// set TCP keep alive idle
	err = socket.SetTcpKeepaliveIdle(z.TcpKeepAliveIdle)
	if err != nil {
		return nil, err
	}
	// set TCP keep alive interval
	err = socket.SetTcpKeepaliveIntvl(z.TcpKeepAliveInterval)
	if err != nil {
		return nil, err
	}

	// connect to endpoints
	for _, endpoint := range z.Endpoints {
		if z.Method == bind{
			err = socket.Bind(endpoint)
		} else {
			err = socket.Connect(endpoint)
		}

		if err != nil {
			return nil, err
		}
	}

	// set subscription filters
	for _, filter := range z.Subscriptions {
		err = socket.SetSubscribe(filter)
		if err != nil {
			return nil, err
		}
	}

	return socket, nil
}

func (z *zmqConsumer) cleanup(socket *zmq.Socket) {
	// discard pending messages
	socket.SetLinger(0)
	err := socket.Close()
	if err != nil {
		z.Log.Errorf("Error closing socket: %s", err.Error())
	}
}

// subscriber receives messages from the socket
func (z *zmqConsumer) subscriber(ctx context.Context) {
	// connect to PUB socket(s)
	socket, err := z.connect()
	if err != nil {
		z.Log.Errorf("Error connecting to socket: %s", err.Error())
		return
	}
	defer z.cleanup(socket)

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
