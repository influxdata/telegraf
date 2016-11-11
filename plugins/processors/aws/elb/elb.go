package elb

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/aws/utils"
)

type (
	ELB struct {
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
		Tags        []string          `toml:"tags"`

		client ELBClient
	}

	ELBClient interface {
		DescribeTags(input *elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error)
	}

	CachingELBClient struct {
		client  ELBClient
		ttl     time.Duration
		fetched time.Time
		data    map[string]*elb.DescribeTagsOutput
	}
)

func (e *CachingELBClient) DescribeTags(input *elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {
	id := *input.LoadBalancerNames[0]
	if e.data == nil {
		e.data = map[string]*elb.DescribeTagsOutput{}
	}
	if e.fetched.IsZero() {
		e.fetched = time.Now()
	}
	if time.Since(e.fetched) >= e.ttl {
		e.data = map[string]*elb.DescribeTagsOutput{}
	}
	if _, ok := e.data[id]; !ok {
		response, err := e.client.DescribeTags(input)
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

  ## Specify the metric names to annotate with ELB metadata
  ## By default is configured for "cloudwatch_aws_elb", the default output from the Cloudwatch input plugin
  #metric_names = [ "cloudwatch_aws_elb" ]

  ## Specify the metric tag which contains the ELB Name
  ## By default is configured for "load_balancer_name", the default from Cloudwatch input plugin when using the LoadBalancerName dimension
  #id = "load_balancer_name"

  ## Specify the ELB Tags to append as metric tags
  #tags = [ "Name" ]
`

func (e *ELB) SampleConfig() string {
	return sampleConfig
}

func (e *ELB) Description() string {
	return "Annotate metrics with AWS ELB metadata"
}

func (e *ELB) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if e.client == nil {
		e.initElbClient()
	}
	for _, metric := range in {
		if utils.IsSelected(metric, e.MetricNames) {
			e.annotate(metric)
		}
	}
	return in
}

func init() {
	processors.Add("aws_metadata_elb", func() telegraf.Processor {
		return &ELB{
			CacheTTL: internal.Duration{Duration: time.Duration(1 * time.Hour)},
			MetricNames: []string{
				"cloudwatch_aws_elb",
			},
			Id:   "load_balancer_name",
			Tags: []string{"Name"},
		}
	})
}

func (e *ELB) annotate(metric telegraf.Metric) {
	e.annotateWithTags(metric)
}

func (e *ELB) annotateWithTags(metric telegraf.Metric) {
	tags, err := e.getTagsForLoadBalancer(metric)
	if err != nil {
		log.Printf("E! %s", err)
		return
	}
	for _, tag := range e.Tags {
		for _, it := range tags {
			if tag == *it.Key {
				metric.AddTag(tag, *it.Value)
				break
			}
		}
	}
}

func (e *ELB) getTagsForLoadBalancer(metric telegraf.Metric) ([]*elb.Tag, error) {
	name := metric.Tags()[e.Id]
	output, err := e.client.DescribeTags(&elb.DescribeTagsInput{
		LoadBalancerNames: []*string{
			aws.String(name),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(output.TagDescriptions) == 0 {
		return nil, fmt.Errorf("ELB %s not found", name)
	}
	return output.TagDescriptions[0].Tags, nil
}

func (e *ELB) initElbClient() error {
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
	e.client = &CachingELBClient{
		client: elb.New(configProvider),
		ttl:    e.CacheTTL.Duration,
	}
	return nil
}
