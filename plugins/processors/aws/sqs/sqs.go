package sqs

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/aws/utils"
)

type (
	SQS struct {
		Region    string `toml:"region"`
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
		RoleARN   string `toml:"role_arn"`
		Profile   string `toml:"profile"`
		Filename  string `toml:"shared_credential_file"`
		Token     string `toml:"token"`

		CacheTTL    internal.Duration `toml:"cache_ttl"`
		MetricNames []string          `toml:"metric_names"`
		Id          string            `toml:"id"`

		client SQSClient
	}

	SQSClient interface {
	}

	CachingSQSClient struct {
		client  SQSClient
		ttl     time.Duration
		fetched time.Time
	}
)

var sampleConfig = `
  ## Amazon Region
  region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #profile = ""
  #shared_credential_file = ""

  ## Specify the TTL for metadata lookups
  #cache_ttl = "1h"

  ## Specify the metric names to annotate with SQS metadata
  ## By default is configured for "cloudwatch_aws_sqs", the default output from the Cloudwatch input plugin
  #metric_names = [ "cloudwatch_aws_sqs" ]

  ## Specify the metric tag which contains the SQS queue name
  ## By default is configured for "queue_name", the default from Cloudwatch input plugin when using the QueueName dimension
  #id = "queue_name"
`

func (s *SQS) SampleConfig() string {
	return sampleConfig
}

func (s *SQS) Description() string {
	return "Annotate metrics with AWS SQS metadata"
}

func (s *SQS) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if s.client == nil {
		s.initSqsClient()
	}
	for _, metric := range in {
		if utils.IsSelected(metric, s.MetricNames) {
			s.annotate(metric)
		}
	}
	return in
}

func init() {
	processors.Add("aws_metadata_sqs", func() telegraf.Processor {
		return &SQS{
			CacheTTL: internal.Duration{Duration: time.Duration(1 * time.Hour)},
			MetricNames: []string{
				"cloudwatch_aws_sqs",
			},
			Id: "queue_name",
		}
	})
}

func (s *SQS) annotate(metric telegraf.Metric) {
}

func (s *SQS) initSqsClient() error {
	credentialConfig := &internalaws.CredentialConfig{
		Region:    s.Region,
		AccessKey: s.AccessKey,
		SecretKey: s.SecretKey,
		RoleARN:   s.RoleARN,
		Profile:   s.Profile,
		Filename:  s.Filename,
		Token:     s.Token,
	}
	configProvider := credentialConfig.Credentials()
	s.client = &CachingSQSClient{
		client: sqs.New(configProvider),
		ttl:    s.CacheTTL.Duration,
	}
	return nil
}
