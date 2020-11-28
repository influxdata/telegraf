package natsconsumer

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/nats-io/nats.go"
)

var (
	defaultMaxUndeliveredMessages = 1000
)

type empty struct{}
type semaphore chan empty

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
	QueueGroup  string   `toml:"queue_group"`
	Subjects    []string `toml:"subjects"`
	Servers     []string `toml:"servers"`
	Secure      bool     `toml:"secure"`
	Username    string   `toml:"username"`
	Password    string   `toml:"password"`
	Credentials string   `toml:"credentials"`

	tls.ClientConfig

	Log telegraf.Logger

	// Client pending limits:
	PendingMessageLimit int `toml:"pending_message_limit"`
	PendingBytesLimit   int `toml:"pending_bytes_limit"`

	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	// Legacy metric buffer support; deprecated in v0.10.3
	MetricBuffer int

	conn *nats.Conn
	subs []*nats.Subscription

	parser parsers.Parser
	// channel for all incoming NATS messages
	in chan *nats.Msg
	// channel for all NATS read errors
	errs   chan error
	acc    telegraf.TrackingAccumulator
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

var sampleConfig = `
  ## urls of NATS servers
  servers = ["nats://localhost:4222"]

  ## subject(s) to consume
  subjects = ["telegraf"]

  ## name a queue group
  queue_group = "telegraf_consumers"

  ## Optional credentials
  # username = ""
  # password = ""

  ## Optional NATS 2.0 and NATS NGS compatible user credentials
  # credentials = "/etc/telegraf/nats.creds"

  ## Use Transport Layer Security
  # secure = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Sets the limits for pending msgs and bytes for each subscription
  ## These shouldn't need to be adjusted except in very high throughput scenarios
  # pending_message_limit = 65536
  # pending_bytes_limit = 67108864

  ## Maximum messages to read from the broker that have not been written by an
  ## output.  For best throughput set based on the number of metrics within
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
	n.acc = acc.WithTracking(n.MaxUndeliveredMessages)

	var connectErr error

	options := []nats.Option{
		nats.MaxReconnects(-1),
		nats.ErrorHandler(n.natsErrHandler),
	}

	// override authentication, if any was specified
	if n.Username != "" && n.Password != "" {
		options = append(options, nats.UserInfo(n.Username, n.Password))
	}

	if n.Credentials != "" {
		options = append(options, nats.UserCredentials(n.Credentials))
	}

	if n.Secure {
		tlsConfig, err := n.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		options = append(options, nats.Secure(tlsConfig))
	}

	if n.conn == nil || n.conn.IsClosed() {
		n.conn, connectErr = nats.Connect(strings.Join(n.Servers, ","), options...)
		if connectErr != nil {
			return connectErr
		}

		// Setup message and error channels
		n.errs = make(chan error)

		n.in = make(chan *nats.Msg, 1000)
		for _, subj := range n.Subjects {
			sub, err := n.conn.QueueSubscribe(subj, n.QueueGroup, func(m *nats.Msg) {
				n.in <- m
			})
			if err != nil {
				return err
			}

			// set the subscription pending limits
			err = sub.SetPendingLimits(n.PendingMessageLimit, n.PendingBytesLimit)
			if err != nil {
				return err
			}

			n.subs = append(n.subs, sub)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	n.cancel = cancel

	// Start the message reader
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		go n.receiver(ctx)
	}()

	n.Log.Infof("Started the NATS consumer service, nats: %v, subjects: %v, queue: %v",
		n.conn.ConnectedUrl(), n.Subjects, n.QueueGroup)

	return nil
}

// receiver() reads all incoming messages from NATS, and parses them into
// telegraf metrics.
func (n *natsConsumer) receiver(ctx context.Context) {
	sem := make(semaphore, n.MaxUndeliveredMessages)

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.acc.Delivered():
			<-sem
		case err := <-n.errs:
			n.Log.Error(err)
		case sem <- empty{}:
			select {
			case <-ctx.Done():
				return
			case err := <-n.errs:
				<-sem
				n.Log.Error(err)
			case <-n.acc.Delivered():
				<-sem
				<-sem
			case msg := <-n.in:
				metrics, err := n.parser.Parse(msg.Data)
				if err != nil {
					n.Log.Errorf("Subject: %s, error: %s", msg.Subject, err.Error())
					<-sem
					continue
				}

				n.acc.AddTrackingMetricGroup(metrics)
			}
		}
	}
}

func (n *natsConsumer) clean() {
	for _, sub := range n.subs {
		if err := sub.Unsubscribe(); err != nil {
			n.Log.Errorf("Error unsubscribing from subject %s in queue %s: %s",
				sub.Subject, sub.Queue, err.Error())
		}
	}

	if n.conn != nil && !n.conn.IsClosed() {
		n.conn.Close()
	}
}

func (n *natsConsumer) Stop() {
	n.cancel()
	n.wg.Wait()
	n.clean()
}

func (n *natsConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("nats_consumer", func() telegraf.Input {
		return &natsConsumer{
			Servers:                []string{"nats://localhost:4222"},
			Secure:                 false,
			Subjects:               []string{"telegraf"},
			QueueGroup:             "telegraf_consumers",
			PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
			PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		}
	})
}
