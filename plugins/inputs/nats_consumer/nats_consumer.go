//go:generate ../../../tools/readme_config_includer/generator
package nats_consumer

import (
	"context"
	_ "embed"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var (
	once                          sync.Once
	defaultMaxUndeliveredMessages = 1000
)

type NatsConsumer struct {
	QueueGroup             string          `toml:"queue_group"`
	Subjects               []string        `toml:"subjects"`
	Servers                []string        `toml:"servers"`
	Secure                 bool            `toml:"secure"`
	Username               string          `toml:"username"`
	Password               string          `toml:"password"`
	Credentials            string          `toml:"credentials"`
	NkeySeed               string          `toml:"nkey_seed"`
	JsSubjects             []string        `toml:"jetstream_subjects"`
	JsStream               string          `toml:"jetstream_stream"`
	PendingMessageLimit    int             `toml:"pending_message_limit"`
	PendingBytesLimit      int             `toml:"pending_bytes_limit"`
	MaxUndeliveredMessages int             `toml:"max_undelivered_messages"`
	Log                    telegraf.Logger `toml:"-"`
	tls.ClientConfig

	conn   *nats.Conn
	jsConn nats.JetStreamContext
	subs   []*nats.Subscription
	jsSubs []*nats.Subscription

	parser telegraf.Parser
	// channel for all incoming NATS messages
	in          chan *nats.Msg
	undelivered map[telegraf.TrackingID]*nats.Msg
	// channel for all NATS read errors
	errs   chan error
	acc    telegraf.TrackingAccumulator
	sem    semaphore
	wg     sync.WaitGroup
	cancel context.CancelFunc
	sync.Mutex
}

type (
	empty     struct{}
	semaphore chan empty
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

func (*NatsConsumer) SampleConfig() string {
	return sampleConfig
}

func (n *NatsConsumer) SetParser(parser telegraf.Parser) {
	n.parser = parser
}

// Start the nats consumer. Caller must call *NatsConsumer.Stop() to clean up.
func (n *NatsConsumer) Start(acc telegraf.Accumulator) error {
	n.sem = make(semaphore, n.MaxUndeliveredMessages)
	n.acc = acc.WithTracking(n.MaxUndeliveredMessages)
	n.undelivered = make(map[telegraf.TrackingID]*nats.Msg, n.MaxUndeliveredMessages)

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

	if n.NkeySeed != "" {
		opt, err := nats.NkeyOptionFromSeed(n.NkeySeed)
		if err != nil {
			return err
		}
		options = append(options, opt)
	}

	if n.Secure {
		tlsConfig, err := n.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		options = append(options, nats.Secure(tlsConfig))
	}

	if n.conn == nil || n.conn.IsClosed() {
		var connectErr error
		n.conn, connectErr = nats.Connect(strings.Join(n.Servers, ","), options...)
		if connectErr != nil {
			return connectErr
		}

		// Setup message and error channels
		n.errs = make(chan error)

		n.in = make(chan *nats.Msg, n.PendingMessageLimit)
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

		if len(n.JsSubjects) > 0 {
			var connErr error
			subOptions := []nats.SubOpt{
				nats.ManualAck(),
			}
			if n.JsStream != "" {
				subOptions = append(subOptions, nats.BindStream(n.JsStream))
			}
			n.jsConn, connErr = n.conn.JetStream(nats.PublishAsyncMaxPending(256))
			if connErr != nil {
				return connErr
			}

			if n.jsConn != nil {
				for _, jsSub := range n.JsSubjects {
					sub, err := n.jsConn.QueueSubscribe(jsSub, n.QueueGroup, func(m *nats.Msg) {
						n.in <- m
					}, subOptions...)
					if err != nil {
						return err
					}

					// set the subscription pending limits
					err = sub.SetPendingLimits(n.PendingMessageLimit, n.PendingBytesLimit)
					if err != nil {
						return err
					}

					n.jsSubs = append(n.jsSubs, sub)
				}
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	n.cancel = cancel

	// Start goroutine to handle delivery notifications from accumulator.
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.waitForDelivery(ctx)
	}()

	// Start the message reader
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		go n.receiver(ctx)
	}()

	n.Log.Infof("Started the NATS consumer service, nats: %v, subjects: %v, jssubjects: %v, queue: %v",
		n.conn.ConnectedUrl(), n.Subjects, n.JsSubjects, n.QueueGroup)

	return nil
}

func (*NatsConsumer) Gather(telegraf.Accumulator) error {
	return nil
}

func (n *NatsConsumer) Stop() {
	n.cancel()
	n.wg.Wait()
	n.clean()
}

func (n *NatsConsumer) natsErrHandler(c *nats.Conn, s *nats.Subscription, e error) {
	select {
	case n.errs <- natsError{conn: c, sub: s, err: e}:
	default:
		return
	}
}

// receiver() reads all incoming messages from NATS, and parses them into
// telegraf metrics.
func (n *NatsConsumer) receiver(ctx context.Context) {
	for {
		// Acquire a semaphore to block consumption if the number of undelivered messages
		// reached it's limit
		select {
		case <-ctx.Done():
			return
		case n.sem <- empty{}:
		}

		// Consume messages and errors
		select {
		case <-ctx.Done():
			return
		case err := <-n.errs:
			n.Log.Error(err)
		case msg := <-n.in:
			jetstreamMsg := slices.Contains(n.jsSubs, msg.Sub)

			if jetstreamMsg {
				if err := msg.InProgress(); err != nil {
					n.Log.Warnf("Failed to mark JetStream message as in progress on subject %s: %v", msg.Subject, err)
				}
			}

			// Parse the metric and add it to the accumulator
			metrics, err := n.parser.Parse(msg.Data)
			if err != nil {
				n.acc.AddError(fmt.Errorf("failed to handle message on subject %s: %w", msg.Subject, err))
			}
			if len(metrics) == 0 {
				once.Do(func() {
					n.Log.Debug(internal.NoMetricsCreatedMsg)
				})
				<-n.sem
				if jetstreamMsg {
					if err := msg.Ack(); err != nil {
						n.acc.AddError(fmt.Errorf("failed to acknowledge JetStream message on subject %s: %w", msg.Subject, err))
					}
				}
			} else {
				for _, m := range metrics {
					m.AddTag("subject", msg.Subject)
				}
				id := n.acc.AddTrackingMetricGroup(metrics)

				// Make sure we manually acknowledge the messages later on delivery to Telegraf output(s)
				if jetstreamMsg {
					n.Lock()
					n.undelivered[id] = msg
					n.Unlock()
				}
			}
		}
	}
}

func (n *NatsConsumer) waitForDelivery(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case track := <-n.acc.Delivered():
			// Get the tracked metric if any. Please remember, only Jetstream messages support a manual ACK
			n.Lock()
			msg, ok := n.undelivered[track.ID()]
			delete(n.undelivered, track.ID())
			n.Unlock()

			if !ok {
				<-n.sem
				continue
			}
			if track.Delivered() {
				if err := msg.Ack(); err != nil {
					n.Log.Errorf("Failed to acknowledge JetStream message on subject %s: %v", msg.Subject, err)
				}
			} else {
				if err := msg.Term(); err != nil {
					n.Log.Errorf("Failed to terminate JetStream message on subject %s: %v", msg.Subject, err)
				}
			}
			<-n.sem
		}
	}
}

func (n *NatsConsumer) clean() {
	for _, sub := range n.subs {
		if err := sub.Unsubscribe(); err != nil {
			n.Log.Errorf("Error unsubscribing from subject %s in queue %s: %s",
				sub.Subject, sub.Queue, err)
		}
	}

	for _, sub := range n.jsSubs {
		if err := sub.Unsubscribe(); err != nil {
			n.Log.Errorf("Error unsubscribing from subject %s in queue %s: %s",
				sub.Subject, sub.Queue, err)
		}
	}

	if n.conn != nil && !n.conn.IsClosed() {
		n.conn.Close()
	}
}

func init() {
	inputs.Add("nats_consumer", func() telegraf.Input {
		return &NatsConsumer{
			Servers:                []string{"nats://localhost:4222"},
			Subjects:               []string{"telegraf"},
			QueueGroup:             "telegraf_consumers",
			PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
			PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		}
	})
}
