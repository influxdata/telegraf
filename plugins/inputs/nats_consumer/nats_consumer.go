package natsconsumer

import (
	"fmt"
	"log"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/nats-io/nats"
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

	// Legacy metric buffer support
	MetricBuffer int

	parser parsers.Parser

	sync.Mutex
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
  servers = ["nats://localhost:4222"]
  ## Use Transport Layer Security
  secure = false
  ## subject(s) to consume
  subjects = ["telegraf"]
  ## name a queue group
  queue_group = "telegraf_consumers"

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
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

	opts := nats.DefaultOptions
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

		n.in = make(chan *nats.Msg)
		for _, subj := range n.Subjects {
			sub, err := n.Conn.ChanQueueSubscribe(subj, n.QueueGroup, n.in)
			if err != nil {
				return err
			}
			n.Subs = append(n.Subs, sub)
		}
	}

	n.done = make(chan struct{})

	// Start the message reader
	go n.receiver()
	log.Printf("I! Started the NATS consumer service, nats: %v, subjects: %v, queue: %v\n",
		n.Conn.ConnectedUrl(), n.Subjects, n.QueueGroup)

	return nil
}

// receiver() reads all incoming messages from NATS, and parses them into
// telegraf metrics.
func (n *natsConsumer) receiver() {
	defer n.clean()
	for {
		select {
		case <-n.done:
			return
		case err := <-n.errs:
			log.Printf("E! error reading from %s\n", err.Error())
		case msg := <-n.in:
			metrics, err := n.parser.Parse(msg.Data)
			if err != nil {
				log.Printf("E! subject: %s, error: %s", msg.Subject, err.Error())
			}

			for _, metric := range metrics {
				n.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
			}

		}
	}
}

func (n *natsConsumer) clean() {
	n.Lock()
	defer n.Unlock()
	close(n.in)
	close(n.errs)

	for _, sub := range n.Subs {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("E! Error unsubscribing from subject %s in queue %s: %s\n",
				sub.Subject, sub.Queue, err.Error())
		}
	}

	if n.Conn != nil && !n.Conn.IsClosed() {
		n.Conn.Close()
	}
}

func (n *natsConsumer) Stop() {
	n.Lock()
	close(n.done)
	n.Unlock()
}

func (n *natsConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("nats_consumer", func() telegraf.Input {
		return &natsConsumer{}
	})
}
