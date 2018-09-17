package natsconsumer

import (
	"fmt"
	"log"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	nats "github.com/nats-io/go-nats"
)

type natsError struct {
	conn *nats.Conn
	sub  *nats.Subscription
	err  error
}

func (e natsError) Error() string {
	return fmt.Sprintf("%s url:%s id:%s sub:%s queue:%s",
		e.err.Error(), e.conn.ConnectedUrl(), e.conn.ConnectedServerId(), e.sub.Subject, e.sub.Queue)
}

type natsConsumer struct {
	QueueGroup string
	Subjects   []string
	Servers    []string
	Secure     bool

	// Client pending limits:
	PendingMessageLimit int
	PendingBytesLimit   int

	// Legacy metric buffer support
	MetricBuffer int

	parser parsers.Parser

	sync.Mutex
	wg   sync.WaitGroup
	Conn *nats.Conn
	Subs []*nats.Subscription

	// channel for all incoming NATS messages
	in chan *nats.Msg
	// channel for all NATS read errors
	errs chan error
	done chan struct{}
	acc  telegraf.Accumulator
}

var sampleConfig = `
  ## urls of NATS servers
  # servers = ["nats://localhost:4222"]
  ## Use Transport Layer Security
  # secure = false
  ## subject(s) to consume
  # subjects = ["telegraf"]
  ## name a queue group
  # queue_group = "telegraf_consumers"

  ## Sets the limits for pending msgs and bytes for each subscription
  ## These shouldn't need to be adjusted except in very high throughput scenarios
  # pending_message_limit = 65536
  # pending_bytes_limit = 67108864

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (n *natsConsumer) SampleConfig() string {
	return sampleConfig
}

func (n *natsConsumer) Description() string {
	return "Read metrics from NATS subject(s)"
}

func (n *natsConsumer) SetParser(parser parsers.Parser) {
	n.parser = parser
}

func (n *natsConsumer) natsErrHandler(c *nats.Conn, s *nats.Subscription, e error) {
	select {
	case n.errs <- natsError{conn: c, sub: s, err: e}:
	default:
		return
	}
}

// Start the nats consumer. Caller must call *natsConsumer.Stop() to clean up.
func (n *natsConsumer) Start(acc telegraf.Accumulator) error {
	n.Lock()
	defer n.Unlock()

	n.acc = acc

	var connectErr error

	// set default NATS connection options
	opts := nats.DefaultOptions

	// override max reconnection tries
	opts.MaxReconnect = -1

	// override servers if any were specified
	opts.Servers = n.Servers

	opts.Secure = n.Secure

	if n.Conn == nil || n.Conn.IsClosed() {
		n.Conn, connectErr = opts.Connect()
		if connectErr != nil {
			return connectErr
		}

		// Setup message and error channels
		n.errs = make(chan error)
		n.Conn.SetErrorHandler(n.natsErrHandler)

		n.in = make(chan *nats.Msg, 1000)
		for _, subj := range n.Subjects {
			sub, err := n.Conn.QueueSubscribe(subj, n.QueueGroup, func(m *nats.Msg) {
				n.in <- m
			})
			if err != nil {
				return err
			}
			// ensure that the subscription has been processed by the server
			if err = n.Conn.Flush(); err != nil {
				return err
			}
			// set the subscription pending limits
			if err = sub.SetPendingLimits(n.PendingMessageLimit, n.PendingBytesLimit); err != nil {
				return err
			}
			n.Subs = append(n.Subs, sub)
		}
	}

	n.done = make(chan struct{})

	// Start the message reader
	n.wg.Add(1)
	go n.receiver()
	log.Printf("I! Started the NATS consumer service, nats: %v, subjects: %v, queue: %v\n",
		n.Conn.ConnectedUrl(), n.Subjects, n.QueueGroup)

	return nil
}

// receiver() reads all incoming messages from NATS, and parses them into
// telegraf metrics.
func (n *natsConsumer) receiver() {
	defer n.wg.Done()
	for {
		select {
		case <-n.done:
			return
		case err := <-n.errs:
			n.acc.AddError(fmt.Errorf("E! error reading from %s\n", err.Error()))
		case msg := <-n.in:
			metrics, err := n.parser.Parse(msg.Data)
			if err != nil {
				n.acc.AddError(fmt.Errorf("E! subject: %s, error: %s", msg.Subject, err.Error()))
			}

			for _, metric := range metrics {
				n.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
			}
		}
	}
}

func (n *natsConsumer) clean() {
	for _, sub := range n.Subs {
		if err := sub.Unsubscribe(); err != nil {
			n.acc.AddError(fmt.Errorf("E! Error unsubscribing from subject %s in queue %s: %s\n",
				sub.Subject, sub.Queue, err.Error()))
		}
	}

	if n.Conn != nil && !n.Conn.IsClosed() {
		n.Conn.Close()
	}
}

func (n *natsConsumer) Stop() {
	n.Lock()
	close(n.done)
	n.wg.Wait()
	n.clean()
	n.Unlock()
}

func (n *natsConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("nats_consumer", func() telegraf.Input {
		return &natsConsumer{
			Servers:             []string{"nats://localhost:4222"},
			Secure:              false,
			Subjects:            []string{"telegraf"},
			QueueGroup:          "telegraf_consumers",
			PendingBytesLimit:   nats.DefaultSubPendingBytesLimit,
			PendingMessageLimit: nats.DefaultSubPendingMsgsLimit,
		}
	})
}
