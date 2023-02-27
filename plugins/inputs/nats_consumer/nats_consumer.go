//go:generate ../../../tools/readme_config_includer/generator
package nats_consumer

import (
	"context"
	_ "embed"
	"github.com/influxdata/telegraf/plugins/common/parsing"
	"github.com/nats-io/nats.go"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

//go:embed sample.conf
var sampleConfig string

var (
	defaultMaxUndeliveredMessages = 1000
)

type empty struct{}
type semaphore chan empty

type SubjectParsingConfig struct {
	parsing.ConfigEntry
	Base string `toml:"subject"`
}

type client interface {
	QueueSubscribe(subj, group string, callback nats.MsgHandler) (*nats.Subscription, error)
	JetStream(opts ...nats.JSOpt) (nats.JetStreamContext, error)
	ConnectedUrl() string
	IsClosed() bool
	Close()
}

// factory methods to make the client testable,
type clientFactory func(url string, opts ...nats.Option) (client, error)
type setPendingLimitsFactory func(subscription *nats.Subscription, msgLimit, bytesLimit int) error

type natsConsumer struct {
	QueueGroup     string                 `toml:"queue_group"`
	Subjects       []string               `toml:"subjects"`
	SubjectTag     *string                `toml:"subject_tag"`
	SubjectParsing []SubjectParsingConfig `toml:"subject_parsing"`
	Servers        []string               `toml:"servers"`
	Secure         bool                   `toml:"secure"`
	Username       string                 `toml:"username"`
	Password       string                 `toml:"password"`
	Credentials    string                 `toml:"credentials"`
	JsSubjects     []string               `toml:"jetstream_subjects"`

	tls.ClientConfig

	Log telegraf.Logger

	// Client pending limits:
	PendingMessageLimit int `toml:"pending_message_limit"`
	PendingBytesLimit   int `toml:"pending_bytes_limit"`

	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`
	MetricBuffer           int `toml:"metric_buffer" deprecated:"0.10.3;2.0.0;option is ignored"`

	clientFactory           clientFactory
	setPendingLimitsFactory setPendingLimitsFactory
	conn                    client
	jsConn                  nats.JetStreamContext
	subs                    []*nats.Subscription
	jsSubs                  []*nats.Subscription

	parser parsers.Parser
	acc    telegraf.TrackingAccumulator

	ctx           context.Context
	cancel        context.CancelFunc
	sem           semaphore
	messages      map[telegraf.TrackingID]bool
	messagesMutex sync.Mutex

	parsingConfig *parsing.Config
}

func (*natsConsumer) SampleConfig() string {
	return sampleConfig
}

func (n *natsConsumer) SetParser(parser parsers.Parser) {
	n.parser = parser
}

func (n *natsConsumer) natsErrHandler(c *nats.Conn, s *nats.Subscription, e error) {
	n.Log.Errorf("%s url:%s id:%s sub:%s queue:%s", e.Error(), c.ConnectedUrl(), c.ConnectedServerId(), s.Subject, s.Queue)
}

func (n *natsConsumer) Init() error {
	n.messages = map[telegraf.TrackingID]bool{}
	n.parsingConfig = parsing.NewConfig(transformConfig(n.SubjectParsing), "subject", ".", "*", n.Log)
	return n.parsingConfig.Init()
}

func transformConfig(subjectParsing []SubjectParsingConfig) []parsing.ConfigEntry {
	ret := make([]parsing.ConfigEntry, len(subjectParsing))
	for i, c := range subjectParsing {
		ret[i] = parsing.ConfigEntry{
			Base:        c.Base,
			Measurement: c.Measurement,
			Tags:        c.Tags,
			Fields:      c.Fields,
			FieldTypes:  c.FieldTypes,
		}
	}
	return ret
}

// Start the nats consumer. Caller must call *natsConsumer.Stop() to clean up.
func (n *natsConsumer) Start(acc telegraf.Accumulator) error {
	n.acc = acc.WithTracking(n.MaxUndeliveredMessages)
	n.ctx, n.cancel = context.WithCancel(context.Background())
	n.sem = make(semaphore, n.MaxUndeliveredMessages)

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
		n.conn, connectErr = n.clientFactory(strings.Join(n.Servers, ","), options...)
		if connectErr != nil {
			return connectErr
		}

		for _, subj := range n.Subjects {
			sub, err := n.conn.QueueSubscribe(subj, n.QueueGroup, n.recMessage)
			if err != nil {
				return err
			}

			// set the subscription pending limits
			err = n.setPendingLimitsFactory(sub, n.PendingMessageLimit, n.PendingBytesLimit)
			if err != nil {
				return err
			}

			n.subs = append(n.subs, sub)
		}

		if len(n.JsSubjects) > 0 {
			var connErr error
			n.jsConn, connErr = n.conn.JetStream(nats.PublishAsyncMaxPending(256))
			if connErr != nil {
				return connErr
			}

			if n.jsConn != nil {
				for _, jsSub := range n.JsSubjects {
					sub, err := n.jsConn.QueueSubscribe(jsSub, n.QueueGroup, n.recMessage)
					if err != nil {
						return err
					}

					// set the subscription pending limits
					err = n.setPendingLimitsFactory(sub, n.PendingMessageLimit, n.PendingBytesLimit)
					if err != nil {
						return err
					}

					n.jsSubs = append(n.jsSubs, sub)
				}
			}
		}
	}

	n.Log.Infof("Started the NATS consumer service, nats: %v, subjects: %v, jssubjects: %v, queue: %v",
		n.conn.ConnectedUrl(), n.Subjects, n.JsSubjects, n.QueueGroup)

	return nil
}

// receiver() reads all incoming messages from NATS, and parses them into
// telegraf metrics.
func (n *natsConsumer) recMessage(msg *nats.Msg) {
	for {
		// Drain anything that's been delivered
		select {
		case track := <-n.acc.Delivered():
			n.onDelivered(track)
			continue
		default:
		}

		// Wait for room to accumulate metric, but make delivery progress if possible
		// (Note that select will randomly pick a case if both are available)
		select {
		case track := <-n.acc.Delivered():
			n.onDelivered(track)
		case n.sem <- empty{}:
			err := n.onMessage(n.acc, msg)
			if err != nil {
				n.acc.AddError(err)
				<-n.sem
			}
			return
		}
	}
}
func (n *natsConsumer) onMessage(acc telegraf.TrackingAccumulator, msg *nats.Msg) error {
	metrics, err := n.parser.Parse(msg.Data)
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if n.SubjectTag != nil && *n.SubjectTag != "" {
			metric.AddTag(*n.SubjectTag, msg.Subject)
		}

		if n.parsingConfig != nil {
			err := n.parsingConfig.Parse(msg.Subject, metric)
			if err != nil {
				return err
			}
		}
	}

	id := acc.AddTrackingMetricGroup(metrics)
	n.messagesMutex.Lock()
	n.messages[id] = true
	n.messagesMutex.Unlock()
	return nil
}

func (n *natsConsumer) clean() {
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

func (n *natsConsumer) Stop() {
	n.cancel()
	n.clean()
}

func (n *natsConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (n *natsConsumer) onDelivered(track telegraf.DeliveryInfo) {
	<-n.sem
	n.messagesMutex.Lock()
	_, ok := n.messages[track.ID()]
	if ok {
		delete(n.messages, track.ID())
	}
	n.messagesMutex.Unlock()
}

func New(factory clientFactory, limitsFactory setPendingLimitsFactory) *natsConsumer {
	return &natsConsumer{
		Servers:                 []string{"nats://localhost:4222"},
		Secure:                  false,
		Subjects:                []string{"telegraf"},
		QueueGroup:              "telegraf_consumers",
		PendingBytesLimit:       nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:     nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages:  defaultMaxUndeliveredMessages,
		clientFactory:           factory,
		setPendingLimitsFactory: limitsFactory,
	}
}

func init() {
	inputs.Add("nats_consumer", func() telegraf.Input {
		return New(func(url string, opts ...nats.Option) (client, error) {
			return nats.Connect(url, opts...)
		}, func(subscription *nats.Subscription, msgLimit, bytesLimit int) error {
			return subscription.SetPendingLimits(msgLimit, bytesLimit)
		})
	})
}
