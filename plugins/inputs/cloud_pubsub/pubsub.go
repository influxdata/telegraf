package cloud_pubsub

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"encoding/base64"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"log"
	"time"
)

type empty struct{}
type semaphore chan empty

const defaultMaxUndeliveredMessages = 1000
const defaultRetryDelaySeconds = 5

type PubSub struct {
	sync.Mutex

	CredentialsFile string `toml:"credentials_file"`
	Project         string `toml:"project"`
	Subscription    string `toml:"subscription"`

	// Subscription ReceiveSettings
	MaxExtension           internal.Duration `toml:"max_extension"`
	MaxOutstandingMessages int               `toml:"max_outstanding_messages"`
	MaxOutstandingBytes    int               `toml:"max_outstanding_bytes"`
	MaxReceiverGoRoutines  int               `toml:"max_receiver_go_routines"`

	// Agent settings
	MaxMessageLen            int `toml:"max_message_len"`
	MaxUndeliveredMessages   int `toml:"max_undelivered_messages"`
	RetryReceiveDelaySeconds int `toml:"retry_delay_seconds"`

	Base64Data bool `toml:"base64_data"`

	sub     subscription
	stubSub func() subscription

	cancel context.CancelFunc

	parser parsers.Parser
	wg     *sync.WaitGroup
	acc    telegraf.TrackingAccumulator

	undelivered map[telegraf.TrackingID]message
	sem         semaphore
}

func (ps *PubSub) Description() string {
	return "Read metrics from Google PubSub"
}

func (ps *PubSub) SampleConfig() string {
	return fmt.Sprintf(sampleConfig, defaultMaxUndeliveredMessages)
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
	if ps.Subscription == "" {
		return fmt.Errorf(`"subscription" is required`)
	}

	if ps.Project == "" {
		return fmt.Errorf(`"project" is required`)
	}

	ps.sem = make(semaphore, ps.MaxUndeliveredMessages)
	ps.acc = ac.WithTracking(ps.MaxUndeliveredMessages)

	// Create top-level context with cancel that will be called on Stop().
	ctx, cancel := context.WithCancel(context.Background())
	ps.cancel = cancel

	if ps.stubSub != nil {
		ps.sub = ps.stubSub()
	} else {
		subRef, err := ps.getGCPSubscription(ps.Subscription)
		if err != nil {
			return fmt.Errorf("unable to create subscription handle: %v", err)
		}
		ps.sub = subRef
	}

	ps.wg = &sync.WaitGroup{}
	// Start goroutine to handle delivery notifications from accumulator.
	ps.wg.Add(1)
	go func() {
		defer ps.wg.Done()
		ps.waitForDelivery(ctx)
	}()

	// Start goroutine for subscription receiver.
	ps.wg.Add(1)
	go func() {
		defer ps.wg.Done()
		ps.receiveWithRetry(ctx)
	}()

	return nil
}

// Stop ensures the PubSub subscriptions receivers are stopped by
// canceling the context and waits for goroutines to finish.
func (ps *PubSub) Stop() {
	ps.cancel()
	ps.wg.Wait()
}

// startReceiver is called within a goroutine and manages keeping a
// subscription.Receive() up and running while the plugin has not been stopped.
func (ps *PubSub) receiveWithRetry(parentCtx context.Context) {
	err := ps.startReceiver(parentCtx)

	for err != nil && parentCtx.Err() == nil {
		log.Printf("E! [inputs.cloud_pubsub] Receiver for subscription %s exited with error: %v", ps.sub.ID(), err)

		delay := defaultRetryDelaySeconds
		if ps.RetryReceiveDelaySeconds > 0 {
			delay = ps.RetryReceiveDelaySeconds
		}

		log.Printf("I! [inputs.cloud_pubsub] Waiting %d seconds before attempting to restart receiver...", delay)
		time.Sleep(time.Duration(delay) * time.Second)

		err = ps.startReceiver(parentCtx)
	}
}

func (ps *PubSub) startReceiver(parentCtx context.Context) error {
	log.Printf("I! [inputs.cloud_pubsub] Starting receiver for subscription %s...", ps.sub.ID())
	cctx, ccancel := context.WithCancel(parentCtx)
	err := ps.sub.Receive(cctx, func(ctx context.Context, msg message) {
		if err := ps.onMessage(ctx, msg); err != nil {
			ps.acc.AddError(fmt.Errorf("unable to add message from subscription %s: %v", ps.sub.ID(), err))
		}
	})
	if err != nil {
		ps.acc.AddError(fmt.Errorf("receiver for subscription %s exited: %v", ps.sub.ID(), err))
	} else {
		log.Printf("I! [inputs.cloud_pubsub] subscription pull ended (no error, most likely stopped)")
	}
	ccancel()
	return err
}

// onMessage handles parsing and adding a received message to the accumulator.
func (ps *PubSub) onMessage(ctx context.Context, msg message) error {
	if ps.MaxMessageLen > 0 && len(msg.Data()) > ps.MaxMessageLen {
		msg.Ack()
		return fmt.Errorf("message longer than max_message_len (%d > %d)", len(msg.Data()), ps.MaxMessageLen)
	}

	var data []byte
	if ps.Base64Data {
		strData, err := base64.StdEncoding.DecodeString(string(msg.Data()))
		if err != nil {
			return fmt.Errorf("unable to base64 decode message: %v", err)
		}
		data = []byte(strData)
	} else {
		data = msg.Data()
	}

	metrics, err := ps.parser.Parse(data)
	if err != nil {
		msg.Ack()
		return err
	}

	if len(metrics) == 0 {
		msg.Ack()
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case ps.sem <- empty{}:
		break
	}

	ps.Lock()
	defer ps.Unlock()

	id := ps.acc.AddTrackingMetricGroup(metrics)
	if ps.undelivered == nil {
		ps.undelivered = make(map[telegraf.TrackingID]message)
	}
	ps.undelivered[id] = msg

	return nil
}

func (ps *PubSub) waitForDelivery(parentCtx context.Context) {
	for {
		select {
		case <-parentCtx.Done():
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
	ps.Lock()
	defer ps.Unlock()

	msg, ok := ps.undelivered[id]
	if !ok {
		return nil
	}
	delete(ps.undelivered, id)
	return msg
}

func (ps *PubSub) getPubSubClient() (*pubsub.Client, error) {
	var credsOpt option.ClientOption
	if ps.CredentialsFile != "" {
		credsOpt = option.WithCredentialsFile(ps.CredentialsFile)
	} else {
		creds, err := google.FindDefaultCredentials(context.Background(), pubsub.ScopeCloudPlatform)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to find GCP Application Default Credentials: %v."+
					"Either set ADC or provide CredentialsFile config", err)
		}
		credsOpt = option.WithCredentials(creds)
	}
	client, err := pubsub.NewClient(
		context.Background(),
		ps.Project,
		credsOpt,
		option.WithScopes(pubsub.ScopeCloudPlatform),
		option.WithUserAgent(internal.ProductToken()),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to generate PubSub client: %v", err)
	}
	return client, nil
}

func (ps *PubSub) getGCPSubscription(subId string) (subscription, error) {
	client, err := ps.getPubSubClient()
	if err != nil {
		return nil, err
	}
	s := client.Subscription(subId)
	s.ReceiveSettings = pubsub.ReceiveSettings{
		NumGoroutines:          ps.MaxReceiverGoRoutines,
		MaxExtension:           ps.MaxExtension.Duration,
		MaxOutstandingMessages: ps.MaxOutstandingMessages,
		MaxOutstandingBytes:    ps.MaxOutstandingBytes,
	}
	return &gcpSubscription{s}, nil
}

func init() {
	inputs.Add("cloud_pubsub", func() telegraf.Input {
		ps := &PubSub{
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		}
		return ps
	})
}

const sampleConfig = `
  ## Required. Name of Google Cloud Platform (GCP) Project that owns
  ## the given PubSub subscription.
  project = "my-project"

  ## Required. Name of PubSub subscription to ingest metrics from.
  subscription = "my-subscription"

  ## Required. Data format to consume.
  ## Each data format has its own unique set of configuration options.
  ## Read more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Optional. Filepath for GCP credentials JSON file to authorize calls to
  ## PubSub APIs. If not set explicitly, Telegraf will attempt to use
  ## Application Default Credentials, which is preferred.
  # credentials_file = "path/to/my/creds.json"

  ## Optional. Number of seconds to wait before attempting to restart the 
  ## PubSub subscription receiver after an unexpected error. 
  ## If the streaming pull for a PubSub Subscription fails (receiver),
  ## the agent attempts to restart receiving messages after this many seconds.
  # retry_delay_seconds = 5

  ## Optional. Maximum byte length of a message to consume.
  ## Larger messages are dropped with an error. If less than 0 or unspecified,
  ## treated as no limit.
  # max_message_len = 1000000

  ## Optional. Maximum messages to read from PubSub that have not been written
  ## to an output. Defaults to %d.
  ## For best throughput set based on the number of metrics within
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

  ## Optional. Maximum number of seconds for which a PubSub subscription
  ## should auto-extend the PubSub ACK deadline for each message. If less than
  ## 0, auto-extension is disabled.
  # max_extension = 0

  ## Optional. Maximum number of unprocessed messages in PubSub
  ## (unacknowledged but not yet expired in PubSub).
  ## A value of 0 is treated as the default PubSub value.
  ## Negative values will be treated as unlimited.
  # max_outstanding_messages = 0

  ## Optional. Maximum size in bytes of unprocessed messages in PubSub
  ## (unacknowledged but not yet expired in PubSub).
  ## A value of 0 is treated as the default PubSub value.
  ## Negative values will be treated as unlimited.
  # max_outstanding_bytes = 0

  ## Optional. Max number of goroutines a PubSub Subscription receiver can spawn
  ## to pull messages from PubSub concurrently. This limit applies to each
  ## subscription separately and is treated as the PubSub default if less than
  ## 1. Note this setting does not limit the number of messages that can be
  ## processed concurrently (use "max_outstanding_messages" instead).
  # max_receiver_go_routines = 0

  ## Optional. If true, Telegraf will attempt to base64 decode the 
  ## PubSub message data before parsing
  # base64_data = false
`
