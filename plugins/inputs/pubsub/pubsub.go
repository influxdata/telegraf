package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"log"
	"sync"
)

type empty struct{}
type semaphore chan empty

const defaultMaxUndeliveredMessages = 1000

type PubSub struct {
	CredentialsFile string `toml:"credentials_file"`
	Project         string `toml:"google_project"`
	Subscription    string `toml:"subscription"`

	// Subscription ReceiveSettings
	MaxExtension           internal.Duration `toml:"max_extension"`
	MaxOutstandingMessages int               `toml:"max_outstanding_messages"`
	MaxOutstandingBytes    int               `toml:"max_outstanding_bytes"`
	MaxReceiverGoRoutines  int               `toml:"max_receiver_go_routines"`

	// Agent settings
	MaxMessageLen          int `toml:"max_message_len"`
	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	SubscriptionTag string `toml:"subscription_tag"`

	sub    subscription
	client *pubsub.Client
	cancel context.CancelFunc
	parser parsers.Parser
	wg     *sync.WaitGroup
	acc    telegraf.TrackingAccumulator

	mu sync.Mutex

	undelivered map[telegraf.TrackingID]message
	sem         semaphore
}

func (ps *PubSub) Description() string {
	return "Read metrics from Google PubSub"
}

func (ps *PubSub) SampleConfig() string {
	return sampleConfig
}

// Gather does nothing for this service input.
func (ps *PubSub) Gather(acc telegraf.Accumulator) error {
	return nil
}

// SetParser implements ParserInput interface.
func (ps *PubSub) SetParser(parser parsers.Parser) {
	ps.parser = parser
}

// Start initializes the plugin and processing messages from Google PubSub.
// Two goroutines are started - one pulling for the subscription, one
// receiving delivery notifications from the accumulator.
func (ps *PubSub) Start(ac telegraf.Accumulator) error {
	cctx, cancel := context.WithCancel(context.Background())
	ps.cancel = cancel
	if err := ps.initPubSubHandles(cctx); err != nil {
		return err
	}

	ps.wg = &sync.WaitGroup{}
	ps.acc = ac.WithTracking(ps.MaxUndeliveredMessages)
	ps.sem = make(semaphore, ps.MaxUndeliveredMessages)

	// Start receiver in new goroutine for each subscription.
	ps.wg.Add(1)
	go func() {
		defer ps.wg.Done()
		ps.subReceive(cctx)
	}()

	// Start goroutine to handle delivery notifications from accumulator.
	ps.wg.Add(1)
	go func() {
		defer ps.wg.Done()
		ps.receiveDelivered(cctx)
	}()

	log.Printf("Started the PubSub service! Listening to subscriptions %+v", ps.Subscription)
	return nil
}

// Stop ensures the PubSub subscriptions receivers are stopped by
// canceling the context and waits for goroutines to finish.
func (ps *PubSub) Stop() {
	ps.cancel()
	ps.wg.Wait()
}

func (ps *PubSub) subReceive(cctx context.Context) {
	err := ps.sub.Receive(cctx, func(ctx context.Context, msg message) {
		if err := ps.onMessage(ctx, msg); err != nil {
			ps.acc.AddError(fmt.Errorf("unable to add message from subscription %s: %v", ps.sub.ID(), err))
		}
	})
	ps.acc.AddError(fmt.Errorf("receiver for subscription %s exited: %v", ps.sub.ID(), err))
}

// onMessage handles parsing and adding a received message to the accumulator.
func (ps *PubSub) onMessage(ctx context.Context, msg message) error {
	if ps.MaxMessageLen > 0 && len(msg.Data()) > ps.MaxMessageLen {
		msg.Ack()
		return fmt.Errorf("message longer than max_message_len (%d > %d)", len(msg.Data()), ps.MaxMessageLen)
	}

	metrics, err := ps.parser.Parse(msg.Data())
	if err != nil {
		msg.Ack()
		return err
	}

	if len(metrics) == 0 {
		msg.Ack()
		return nil
	}

	if len(ps.SubscriptionTag) > 0 {
		for _, m := range metrics {
			m.AddTag(ps.SubscriptionTag, ps.sub.ID())
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case ps.sem <- empty{}:
		break
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	id := ps.acc.AddTrackingMetricGroup(metrics)
	if ps.undelivered == nil {
		ps.undelivered = make(map[telegraf.TrackingID]message)
	}
	ps.undelivered[id] = msg

	return nil
}

func (ps *PubSub) receiveDelivered(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case info := <-ps.acc.Delivered():
			<-ps.sem
			msg := ps.removeDelivered(info.ID())

			if msg != nil {
				msg.Ack()
			}
		}
	}
}

func (ps *PubSub) removeDelivered(id telegraf.TrackingID) message {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	msg, ok := ps.undelivered[id]
	if !ok {
		return nil
	}
	delete(ps.undelivered, id)
	return msg
}

func (ps *PubSub) initPubSubHandles(ctx context.Context) error {
	var credsOpt option.ClientOption
	if ps.CredentialsFile != "" {
		credsOpt = option.WithCredentialsFile(ps.CredentialsFile)
	} else {
		creds, err := google.FindDefaultCredentials(ctx, pubsub.ScopeCloudPlatform)
		if err != nil {
			return fmt.Errorf("credentials JSON file not provided and unable to find GCP Application Default Credentials: %v", err)
		}
		credsOpt = option.WithCredentials(creds)
	}

	client, err := pubsub.NewClient(
		ctx,
		ps.Project,
		credsOpt,
		option.WithScopes(pubsub.ScopeCloudPlatform),
		option.WithUserAgent(internal.ProductToken()),
	)
	if err != nil {
		return fmt.Errorf("unable to generate PubSub client: %v", err)
	}

	ps.client = client
	s := ps.client.Subscription(ps.Subscription)
	s.ReceiveSettings = pubsub.ReceiveSettings{
		NumGoroutines:          ps.MaxReceiverGoRoutines,
		MaxExtension:           ps.MaxExtension.Duration,
		MaxOutstandingMessages: ps.MaxOutstandingMessages,
		MaxOutstandingBytes:    ps.MaxOutstandingBytes,
	}

	ps.sub = &subWrapper{s}
	return nil
}

func init() {
	inputs.Add("pubsub", func() telegraf.Input {
		return &PubSub{
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		}
	})
}

const sampleConfig = `
  ## Name of Google Cloud Platform (GCP) Project owning PubSub subscriptions
  google_project = "my-project"

  ## Name of PubSub subscriptions
  subscriptions = ["sub-one, sub-two"]

  ## Optional filepath for GCP credentials JSON file to authorize calls to 
  ## PubSub APIs. If not set explicitly, Telegraf will attempt to use 
  ## Application Default Credentials, which is preferred. 
  # credentials_file = "path/to/my/creds.json"

  ## Maximum byte length of a message to consume. Larger messages are dropped.
  ## If less than 0 or unspecified, defaults to unlimited length.
  max_message_len = 1000000

  ## Maximum messages to read from PubSub that have not been written to an
  ## output.  For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message contains 10 metrics and the output 
  ## metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## The following are optional Subscription ReceiveSettings in PubSub.
  ## Read more about these values:
  ## https://godoc.org/cloud.google.com/go/pubsub#ReceiveSettings
  
  ## Maximum number of seconds for which a PubSub subscription
  ## should auto-extend the PubSub ACK deadline for each message. If less than
  ## 0, auto-extension is disabled.
  # max_extension = 0

  ## Maximum number of unprocessed messages in PubSub (unacknowledged but not 
  ## yet expired in PubSub). A value of 0 is treated as the default 
  ## PubSub value. Negative values will be treated as unlimited.
  # max_outstanding_messages = 0

  ## Maximum size in bytes of unprocessed messages in PubSub 
  ## (unacknowledged but not yet expired in PubSub). 
  ## A value of 0 is treated as the default PubSub value.
  ## Negative values will be treated as unlimited.
  # max_outstanding_bytes = 0

  ## Max number of goroutines a PubSub Subscription receiver can spawn 
  ## to pull messages from PubSub concurrently. This limit applies to each 
  ## subscription separately and is treated as the PubSub default if less than 1.
  ## Note this setting does not limit the number of messages that can be 
  ## processed concurrently (use "max_outstanding_messages" instead)
  # max_receiver_go_routines = 0

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options.
  ## Read more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`
