package cloud_pubsub

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"encoding/base64"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

const sampleConfig = `
  ## Required. Name of Google Cloud Platform (GCP) Project that owns
  ## the given PubSub topic.
  project = "my-project"

  ## Required. Name of PubSub topic to publish metrics to.
  topic = "my-topic"

  ## Required. Data format to consume.
  ## Each data format has its own unique set of configuration options.
  ## Read more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Optional. Filepath for GCP credentials JSON file to authorize calls to
  ## PubSub APIs. If not set explicitly, Telegraf will attempt to use
  ## Application Default Credentials, which is preferred.
  # credentials_file = "path/to/my/creds.json"

  ## Optional. If true, will send all metrics per write in one PubSub message.
  # send_batched = true

  ## The following publish_* parameters specifically configures batching
  ## requests made to the GCP Cloud PubSub API via the PubSub Golang library. Read
  ## more here: https://godoc.org/cloud.google.com/go/pubsub#PublishSettings

  ## Optional. Send a request to PubSub (i.e. actually publish a batch)
  ## when it has this many PubSub messages. If send_batched is true,
  ## this is ignored and treated as if it were 1.
  # publish_count_threshold = 1000

  ## Optional. Send a request to PubSub (i.e. actually publish a batch)
  ## when it has this many PubSub messages. If send_batched is true,
  ## this is ignored and treated as if it were 1
  # publish_byte_threshold = 1000000

  ## Optional. Specifically configures requests made to the PubSub API.
  # publish_num_go_routines = 2

  ## Optional. Specifies a timeout for requests to the PubSub API.
  # publish_timeout = "30s"

  ## Optional. If true, published PubSub message data will be base64-encoded.
  # base64_data = false

  ## Optional. PubSub attributes to add to metrics.
  # [[inputs.pubsub.attributes]]
  #   my_attr = "tag_value"
`

type PubSub struct {
	CredentialsFile string            `toml:"credentials_file"`
	Project         string            `toml:"project"`
	Topic           string            `toml:"topic"`
	Attributes      map[string]string `toml:"attributes"`

	SendBatched           bool              `toml:"send_batched"`
	PublishCountThreshold int               `toml:"publish_count_threshold"`
	PublishByteThreshold  int               `toml:"publish_byte_threshold"`
	PublishNumGoroutines  int               `toml:"publish_num_go_routines"`
	PublishTimeout        internal.Duration `toml:"publish_timeout"`
	Base64Data            bool              `toml:"base64_data"`

	t topic
	c *pubsub.Client

	stubTopic func(id string) topic

	serializer     serializers.Serializer
	publishResults []publishResult
}

func (ps *PubSub) Description() string {
	return "Publish Telegraf metrics to a Google Cloud PubSub topic"
}

func (ps *PubSub) SampleConfig() string {
	return sampleConfig
}

func (ps *PubSub) SetSerializer(serializer serializers.Serializer) {
	ps.serializer = serializer
}

func (ps *PubSub) Connect() error {
	if ps.Topic == "" {
		return fmt.Errorf(`"topic" is required`)
	}

	if ps.Project == "" {
		return fmt.Errorf(`"project" is required`)
	}

	if ps.stubTopic == nil {
		return ps.initPubSubClient()
	} else {
		return nil
	}
}

func (ps *PubSub) Close() error {
	if ps.t != nil {
		ps.t.Stop()
	}
	return nil
}

func (ps *PubSub) Write(metrics []telegraf.Metric) error {
	ps.refreshTopic()

	// Serialize metrics and package into appropriate PubSub messages
	msgs, err := ps.toMessages(metrics)
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(context.Background())

	// Publish all messages - each call to Publish returns a future.
	ps.publishResults = make([]publishResult, len(msgs))
	for i, m := range msgs {
		ps.publishResults[i] = ps.t.Publish(cctx, m)
	}

	// topic.Stop() forces all published messages to be sent, even
	// if PubSub batch limits have not been reached.
	go ps.t.Stop()

	return ps.waitForResults(cctx, cancel)
}

func (ps *PubSub) initPubSubClient() error {
	var credsOpt option.ClientOption
	if ps.CredentialsFile != "" {
		credsOpt = option.WithCredentialsFile(ps.CredentialsFile)
	} else {
		creds, err := google.FindDefaultCredentials(context.Background(), pubsub.ScopeCloudPlatform)
		if err != nil {
			return fmt.Errorf(
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
		return fmt.Errorf("unable to generate PubSub client: %v", err)
	}
	ps.c = client
	return nil
}

func (ps *PubSub) refreshTopic() {
	if ps.stubTopic != nil {
		ps.t = ps.stubTopic(ps.Topic)
	} else {
		t := ps.c.Topic(ps.Topic)
		ps.t = &topicWrapper{t}
	}
	ps.t.SetPublishSettings(ps.publishSettings())
}

func (ps *PubSub) publishSettings() pubsub.PublishSettings {
	settings := pubsub.PublishSettings{}
	if ps.PublishNumGoroutines > 0 {
		settings.NumGoroutines = ps.PublishNumGoroutines
	}

	if ps.PublishTimeout.Duration > 0 {
		settings.CountThreshold = 1
	}

	if ps.SendBatched {
		settings.CountThreshold = 1
	} else if ps.PublishCountThreshold > 0 {
		settings.CountThreshold = ps.PublishCountThreshold
	}

	if ps.PublishByteThreshold > 0 {
		settings.ByteThreshold = ps.PublishByteThreshold
	}

	return settings
}

func (ps *PubSub) toMessages(metrics []telegraf.Metric) ([]*pubsub.Message, error) {
	if ps.SendBatched {
		b, err := ps.serializer.SerializeBatch(metrics)
		if err != nil {
			return nil, err
		}

		if ps.Base64Data {
			encoded := base64.StdEncoding.EncodeToString(b)
			b = []byte(encoded)
		}

		msg := &pubsub.Message{Data: b}
		if ps.Attributes != nil {
			msg.Attributes = ps.Attributes
		}
		return []*pubsub.Message{msg}, nil
	}

	msgs := make([]*pubsub.Message, len(metrics))
	for i, m := range metrics {
		b, err := ps.serializer.Serialize(m)
		if err != nil {
			return nil, err
		}

		if ps.Base64Data {
			encoded := base64.StdEncoding.EncodeToString(b)
			b = []byte(encoded)
		}

		msgs[i] = &pubsub.Message{
			Data: b,
		}
		if ps.Attributes != nil {
			msgs[i].Attributes = ps.Attributes
		}
	}

	return msgs, nil
}

func (ps *PubSub) waitForResults(ctx context.Context, cancel context.CancelFunc) error {
	var pErr error
	var setErr sync.Once
	var wg sync.WaitGroup

	for _, pr := range ps.publishResults {
		wg.Add(1)

		go func(r publishResult) {
			defer wg.Done()
			// Wait on each future
			_, err := r.Get(ctx)
			if err != nil {
				setErr.Do(func() {
					pErr = err
					cancel()
				})
			}
		}(pr)
	}

	wg.Wait()
	return pErr
}

func init() {
	outputs.Add("cloud_pubsub", func() telegraf.Output {
		return &PubSub{}
	})
}
