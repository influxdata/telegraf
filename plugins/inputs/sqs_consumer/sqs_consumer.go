package sqs_consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type empty struct{}
type semaphore chan empty

const (
	defaultMaxNumberOfMessages     = 10
	defaultWaitTimeSeconds         = 20
	defaultDeleteBatchSize         = 0
	defaultDeleteBatchFlushSeconds = 30
	defaultMaxUndeliveredMessages  = 1000
)

type SQSConsumer struct {
	Region      string `toml:"region"`
	AccessKey   string `toml:"access_key"`
	SecretKey   string `toml:"secret_key"`
	RoleARN     string `toml:"role_arn"`
	Profile     string `toml:"profile"`
	Filename    string `toml:"shared_credential_file"`
	Token       string `toml:"token"`
	EndpointURL string `toml:"endpoint_url"`

	QueueName           string `toml:"queue_name"`
	QueueOwnerAccountID string `toml:"queue_owner_acount_id"`
	QueueURL            string `toml:"queue_url"`

	MaxNumberOfMessages     int `toml:"max_number_of_messages"`
	VisibilityTimeout       int `toml:"visibility_timeout"`
	WaitTimeSeconds         int `toml:"wait_time_seconds"`
	DeleteBatchSize         int `toml:"delete_batch_size"`
	DeleteBatchFlushSeconds int `toml:"delete_batch_flush_seconds"`

	MaxMessageLen          int `toml:"max_message_len"`
	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	Log telegraf.Logger `toml:"-"`

	sqs *sqs.SQS

	parser parsers.Parser

	cancel context.CancelFunc
	wg     *sync.WaitGroup
	acc    telegraf.TrackingAccumulator

	in     chan *sqs.Message
	delete chan *sqs.Message

	deliveries map[telegraf.TrackingID]*sqs.Message
	deletes    map[string]string
}

// TODO: Add internal metrics similar too https://github.com/influxdata/telegraf/blob/master/plugins/inputs/influxdb_listener/influxdb_listener.go#L48

type ticker struct {
	period time.Duration
	ticker time.Ticker
}

func createTicker(period time.Duration) *ticker {
	return &ticker{period, *time.NewTicker(period)}
}

func (t *ticker) resetTicker() {
	t.ticker.Stop()
	t.ticker = *time.NewTicker(t.period)
}

func (s *SQSConsumer) Description() string {
	return "Read metrics from AWS SQS"
}

func (s *SQSConsumer) SampleConfig() string {
	return fmt.Sprintf(sampleConfig, defaultMaxNumberOfMessages, defaultWaitTimeSeconds, defaultDeleteBatchSize, defaultDeleteBatchFlushSeconds, defaultMaxUndeliveredMessages)
}

// Gather does nothing for this service input.
func (s *SQSConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

// SetParser satisfies the ParserInput interface.
func (s *SQSConsumer) SetParser(parser parsers.Parser) {
	s.parser = parser
}

// Start satisfies the telegraf.ServiceInput interface.
func (s *SQSConsumer) Start(acc telegraf.Accumulator) error {
	credentialConfig := &internalaws.CredentialConfig{
		Region:      s.Region,
		AccessKey:   s.AccessKey,
		SecretKey:   s.SecretKey,
		RoleARN:     s.RoleARN,
		Profile:     s.Profile,
		Filename:    s.Filename,
		Token:       s.Token,
		EndpointURL: s.EndpointURL,
	}
	configProvider := credentialConfig.Credentials()

	svc := sqs.New(configProvider)
	s.sqs = svc

	if s.QueueURL == "" && s.QueueName == "" {
		return fmt.Errorf(`either "queue_name" or "queue_url" is required`)
	}

	if s.QueueURL == "" {
		resp, err := s.sqs.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName:              aws.String(s.QueueName),
			QueueOwnerAWSAccountId: aws.String(s.QueueOwnerAccountID),
		})

		if err != nil {
			return fmt.Errorf("error finding SQS queue: %v", err)
		}

		s.QueueURL = *resp.QueueUrl
	}

	s.acc = acc.WithTracking(s.MaxUndeliveredMessages)
	s.in = make(chan *sqs.Message)

	// Create top-level context with cancel that will be called on Stop().
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.wg = &sync.WaitGroup{}

	// Start goroutine to handle delivery notifications from accumulator.
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.waitForDelivery(ctx)
	}()

	// Start goroutine to handle batch message deletes.
	if s.DeleteBatchSize > 1 {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleDeletes(ctx)
		}()
	}

	// Start goroutine for queue consumer.
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.receive(ctx)
	}()

	return nil
}

func (s *SQSConsumer) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *SQSConsumer) waitForDelivery(ctx context.Context) {
	sem := make(semaphore, s.MaxUndeliveredMessages)
	s.deliveries = make(map[telegraf.TrackingID]*sqs.Message, s.MaxUndeliveredMessages)

	for {
		select {
		case <-ctx.Done():
			return
		case track := <-s.acc.Delivered():
			if s.onDelivery(track) {
				<-sem
			}
		case sem <- empty{}:
			select {
			case <-ctx.Done():
				return
			case track := <-s.acc.Delivered():
				if s.onDelivery(track) {
					<-sem
					<-sem
				}
			case msg := <-s.in:
				metrics, err := s.createMetrics(msg)
				if err != nil {
					s.acc.AddError(fmt.Errorf("error parsing message from queue %s: %v", s.QueueURL, err))
					break
				}

				id := s.acc.AddTrackingMetricGroup(metrics)
				s.deliveries[id] = msg
			}
		}
	}
}

func (s *SQSConsumer) receive(ctx context.Context) {
	s.Log.Infof("Starting receiver for queue %s...", s.QueueURL)

	input := sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.QueueURL),
		MaxNumberOfMessages: aws.Int64(int64(s.MaxNumberOfMessages)),
		WaitTimeSeconds:     aws.Int64(int64(s.WaitTimeSeconds)),
	}

	if v := s.VisibilityTimeout; v != 0 {
		input.VisibilityTimeout = aws.Int64(int64(v))
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.Log.Debugf("recieving messages from %s", s.QueueURL)
			resp, err := s.sqs.ReceiveMessage(&input)
			if err != nil {
				s.acc.AddError(fmt.Errorf("receiver for queue %s recieved an error: %v", s.QueueURL, err))
				// TODO: handle errors, probably sleep on retriable errors (timeout, etc), and bail out on other (auth error, wrong queue)
			}

			if len(resp.Messages) == 0 {
				break
				// TODO: add customized wait time on erro or empty reply
			}

			for _, msg := range resp.Messages {
				select {
				case <-ctx.Done():
					return
				default:
					s.in <- msg
				}
			}
		}
	}
}

func (s *SQSConsumer) handleDeletes(ctx context.Context) {
	s.delete = make(chan *sqs.Message)
	s.deletes = make(map[string]string, s.DeleteBatchSize)

	s.Log.Debugf("starting %s queue delete buffer loop", s.QueueURL)

	batchTicker := createTicker(time.Duration(s.DeleteBatchFlushSeconds) * time.Second)
	defer batchTicker.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if v := len(s.deletes); v > 0 {
				s.deleteBatch()
			}
			return
		case <-batchTicker.ticker.C:
			if v := len(s.deletes); v > 0 {
				s.deleteBatch()
			}
		case msg := <-s.delete:
			s.deletes[*msg.MessageId] = *msg.ReceiptHandle

			s.Log.Debugf("%s queue delete buffer fullness: %d / %d messages", s.QueueURL, len(s.deletes), s.DeleteBatchSize)

			if v := len(s.deletes); v == s.DeleteBatchSize {
				batchTicker.resetTicker()

				s.deleteBatch()
			}
		}
	}
}

func (s *SQSConsumer) onDelivery(track telegraf.DeliveryInfo) bool {
	msg, ok := s.deliveries[track.ID()]
	if !ok {
		return false
	}

	if track.Delivered() {
		err := s.deleteMessage(msg)
		if err != nil {
			s.acc.AddError(fmt.Errorf("failed deleting message from queue %s: %v", s.QueueURL, err))
		}
	}

	delete(s.deliveries, track.ID())
	return true
}

func (s *SQSConsumer) createMetrics(msg *sqs.Message) ([]telegraf.Metric, error) {
	if s.MaxMessageLen > 0 && len(*msg.Body) > s.MaxMessageLen {
		// TODO: delete message from queue or allow it to be redelivered and fall in dead letter queue?
		return nil, fmt.Errorf("message longer than max_message_len (%d > %d)", len(*msg.Body), s.MaxMessageLen)
	}

	metrics, err := s.parser.Parse([]byte(*msg.Body))
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s *SQSConsumer) deleteMessage(msg *sqs.Message) error {
	var err error

	if s.DeleteBatchSize <= 1 {
		input := sqs.DeleteMessageInput{
			QueueUrl:      aws.String(s.QueueURL),
			ReceiptHandle: msg.ReceiptHandle,
		}

		s.Log.Debugf("deleting message from %s", s.QueueURL)
		_, err = s.sqs.DeleteMessage(&input)
		if err != nil {
			return err
		}
	} else {
		s.delete <- msg
	}

	return nil
}

func (s *SQSConsumer) deleteBatch() {
	s.Log.Debugf("processign deletion of %d messages from queue %s...", len(s.deletes), s.QueueURL)

	var entries []*sqs.DeleteMessageBatchRequestEntry

	for i, r := range s.deletes {
		entries = append(entries, &sqs.DeleteMessageBatchRequestEntry{
			Id:            aws.String(i),
			ReceiptHandle: aws.String(r),
		})

		delete(s.deletes, i)
	}

	_, err := s.sqs.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(s.QueueURL),
		Entries:  entries,
	})
	if err != nil {
		s.acc.AddError(fmt.Errorf("failed deleting messages from queue %s: %v", s.QueueURL, err))
	}
}

func init() {
	inputs.Add("sqs_consumer", func() telegraf.Input {
		return &SQSConsumer{
			MaxUndeliveredMessages:  defaultMaxUndeliveredMessages,
			MaxNumberOfMessages:     defaultMaxNumberOfMessages,
			WaitTimeSeconds:         defaultWaitTimeSeconds,
			DeleteBatchSize:         defaultDeleteBatchSize,
			DeleteBatchFlushSeconds: defaultDeleteBatchFlushSeconds,
		}
	})
}

const sampleConfig = `
  ## Required. Amazon REGION of SQS endpoint.
  region = "ap-southeast-2"

  ## Optional. Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  # access_key = ""
  # secret_key = ""
  # token = ""
  # role_arn = ""
  # profile = ""
  # shared_credential_file = ""

  ## Optional. Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Optional.
  # queue_name = ""

  ## Optional.
  # queue_owner_acount_id = ""

  ## Optional. Required if queue_name and queue_owner_acount_id were not provided
  queue_url = ""

  ## Optional. The maximum number of messages to return. Defaults to %d
  # max_number_of_messages = 10

  ## Optional. The duration (in seconds) that the received messages are hidden
  ## from subsequent retrieve requests. If not set defaults to queue settings.
  # visibility_timeout = 30

  ## Optional. The duration (in seconds) for which the call waits for a message
  ## to arrive in the queue before returning. If set to higher then 0, enables long polling
  ## 0 enables short polling. Defaults to %d
  # wait_time_seconds = 20

  ## Optional. When > 1 messages will be deleted from a queue in batches of provided size.
  ## Defaults to %d
  # delete_batch_size = 10

  ## Optional. If batch delete is enabled - flush messages
  ## Set this equal to or lower then your visibility timeout. Defaults to %d
  # delete_batch_flush_seconds = 30

  ## Optional. Maximum byte length of a message to consume.
  ## Larger messages are dropped with an error. If less than 0 or unspecified,
  ## treated as no limit.
  # max_message_len = 1000000

  ## Optional. Maximum messages to read from the queue that have not been written by an
  ## output. Defaults to %d.
  ## For best throughput set based on the number of metrics within
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
