package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/aws/utils"
)

type (
	EC2 struct {
		Region    string `toml:"region"`
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
		RoleARN   string `toml:"role_arn"`
		Profile   string `toml:"profile"`
		Filename  string `toml:"shared_credential_file"`
		Token     string `toml:"token"`

		CacheTTL     internal.Duration `toml:"cache_ttl"`
		MetricNames  []string          `toml:"metric_names"`
		Id           string            `toml:"id"`
		InstanceType bool              `toml:"instance_type"`
		AmiId        bool              `toml:"ami_id"`
		Tags         []string          `toml:"tags"`

		client EC2Client
	}

	EC2Client interface {
		DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	}

	CachingEC2Client struct {
		client  EC2Client
		ttl     time.Duration
		fetched time.Time
		data    map[string]*ec2.DescribeInstancesOutput
	}
)

func (e *CachingEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	id := *input.InstanceIds[0]
	if e.data == nil {
		e.data = map[string]*ec2.DescribeInstancesOutput{}
	}
	if e.fetched.IsZero() {
		e.fetched = time.Now()
	}
	if time.Since(e.fetched) >= e.ttl {
		e.data = map[string]*ec2.DescribeInstancesOutput{}
	}
	if _, ok := e.data[id]; !ok {
		response, err := e.client.DescribeInstances(input)
		if err != nil {
			return nil, err
		}
		e.data[id] = response
	}
	return e.data[id], nil
}

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

  ## Specify the metric names to annotate with EC2 metadata
  ## By default is configured for "cloudwatch_aws_ec2", the default output from the Cloudwatch input plugin
  #metric_names = [ "cloudwatch_aws_ec2" ]

  ## Specify the metric tag which contains the EC2 Instance ID
  ## By default is configured for "instance_id", the default from Cloudwatch input plugin when using the InstanceId dimension
  #id = "instance_id"

  ## Enable annotating metrics with the EC2 Instance Type
  #instance_type = true

  ## Enable annotating metrics with the AMI ID
  #ami_id = true

  ## Specify the EC2 Tags to append as metric tags
  #tags = [ "Name" ]
`

func (e *EC2) SampleConfig() string {
	return sampleConfig
}

func (e *EC2) Description() string {
	return "Annotate metrics with AWS EC2 metadata"
}

func (e *EC2) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if e.client == nil {
		e.initEc2Client()
	}
	for _, metric := range in {
		if utils.IsSelected(metric, e.MetricNames) {
			e.annotate(metric)
		}
	}
	return in
}

func init() {
	processors.Add("aws_metadata_ec2", func() telegraf.Processor {
		return &EC2{
			CacheTTL: internal.Duration{Duration: time.Duration(1 * time.Hour)},
			MetricNames: []string{
				"cloudwatch_aws_ec2",
			},
			Id:           "instance_id",
			InstanceType: true,
			AmiId:        true,
			Tags:         []string{"Name"},
		}
	})
}

func (e *EC2) annotate(metric telegraf.Metric) {
	e.annotateWithInstanceMetadata(metric)
	e.annotateWithTags(metric)
}

func (e *EC2) annotateWithInstanceMetadata(metric telegraf.Metric) {
	instance, err := e.getInstanceForMetric(metric)
	if err != nil {
		log.Printf("E! %s", err)
		return
	}
	if e.InstanceType {
		metric.AddTag("instance_type", *instance.InstanceType)
	}
	if e.AmiId {
		metric.AddTag("ami_id", *instance.ImageId)
	}
}

func (e *EC2) annotateWithTags(metric telegraf.Metric) {
	instance, err := e.getInstanceForMetric(metric)
	if err != nil {
		log.Printf("E! %s", err)
		return
	}
	for _, tag := range e.Tags {
		for _, it := range instance.Tags {
			if tag == *it.Key {
				metric.AddTag(tag, *it.Value)
				break
			}
		}
	}
}

func (e *EC2) getInstanceForMetric(metric telegraf.Metric) (*ec2.Instance, error) {
	id := metric.Tags()[e.Id]
	output, err := e.client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(id),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("Instance %s not found", id)
	}
	return output.Reservations[0].Instances[0], nil
}

func (e *EC2) initEc2Client() error {
	credentialConfig := &internalaws.CredentialConfig{
		Region:    e.Region,
		AccessKey: e.AccessKey,
		SecretKey: e.SecretKey,
		RoleARN:   e.RoleARN,
		Profile:   e.Profile,
		Filename:  e.Filename,
		Token:     e.Token,
	}
	configProvider := credentialConfig.Credentials()
	// e.client = ec2.New(configProvider)
	e.client = &CachingEC2Client{
		client: ec2.New(configProvider),
		ttl:    e.CacheTTL.Duration,
	}
	return nil
}
