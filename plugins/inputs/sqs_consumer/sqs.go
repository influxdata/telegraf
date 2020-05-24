package sqs_consumer

import (
	"net/http"
	"time"

	internalaws "github.com/influxdata/telegraf/config/aws"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type Sqs struct {
	URL         string `toml:"url"`
	Region      string `toml:"region"`
	AccessKey   string `toml:"access_key"`
	SecretKey   string `toml:"secret_key"`
	RoleARN     string `toml:"role_arn"`
	Profile     string `toml:"profile"`
	Filename    string `toml:"shared_credential_file"`
	Token       string `toml:"token"`
	EndpointURL string `toml:"endpoint_url"`

	Endpoint string `toml:"endpoint"`

	SqsSdk sqsSdk

	parser parsers.Parser
	client http.Client
}

func (s *Sqs) Description() string {
	return "Read messages from AWS-SQS topic"
}

const sampleConfig = `
[[inputs.sqs]]
  # add goyour sqs you want to subscribe here
  url = "http://aws-region.aws.amazon/queue"
  access_key = "ACCESS_KEY"
  secret_key = "SECRET_ACCESS_KEY"
  region = "eu-west-1"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"
`

func (s *Sqs) SampleConfig() string {
	return sampleConfig
}

func (s *Sqs) Init() error {
	s.client = http.Client{
		Timeout: time.Second * 10,
	}

	credentialConfig := &internalaws.CredentialConfig{
		Region:      s.Region,
		AccessKey:   s.AccessKey,
		SecretKey:   s.SecretKey,
		RoleARN:     s.RoleARN,
		Profile:     s.Profile,
		Filename:    s.Filename,
		Token:       s.Token,
		EndpointURL: s.Endpoint,
	}
	cred := credentialConfig.Credentials()
	s.SqsSdk = sqs.New(cred)
	return nil
}

func (s *Sqs) SetParser(parser parsers.Parser) {
	s.parser = parser
}

func (s *Sqs) Gather(acc telegraf.Accumulator) error {
	res, err := s.SqsSdk.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames:        nil,
		MessageAttributeNames: nil,
		QueueUrl:              &s.URL,
	})

	if err != nil {
		return err
	}

	var deleteBatchEntries []*sqs.DeleteMessageBatchRequestEntry
	for _, m := range res.Messages {
		metric, err := s.parser.ParseLine(*m.Body)
		if err != nil {
			acc.AddError(err)
			continue
		}
		acc.AddMetric(metric)
		deleteBatchEntries = append(deleteBatchEntries, &sqs.DeleteMessageBatchRequestEntry{
			Id:            m.MessageId,
			ReceiptHandle: m.ReceiptHandle,
		})
	}
	if len(deleteBatchEntries) > 0 {
		_, err = s.SqsSdk.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
			Entries:  deleteBatchEntries,
			QueueUrl: &s.URL,
		})
	}
	return err
}

type sqsSdk interface {
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessageBatch(input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error)
}

func init() {
	inputs.Add("sqs", func() telegraf.Input { return &Sqs{} })
}
