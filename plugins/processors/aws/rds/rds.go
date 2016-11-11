package rds

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/aws/utils"
)

type (
	RDS struct {
		Region    string `toml:"region"`
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
		RoleARN   string `toml:"role_arn"`
		Profile   string `toml:"profile"`
		Filename  string `toml:"shared_credential_file"`
		Token     string `toml:"token"`

		CacheTTL      internal.Duration `toml:"cache_ttl"`
		MetricNames   []string          `toml:"metric_names"`
		Id            string            `toml:"id"`
		InstanceType  bool              `toml:"instance_type"`
		Engine        bool              `toml:"engine"`
		EngineVersion bool              `toml:"engine_version"`
		Tags          []string          `toml:"tags"`

		client RDSClient
	}

	RDSClient interface {
		DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)
		ListTagsForResource(input *rds.ListTagsForResourceInput) (*rds.ListTagsForResourceOutput, error)
	}

	CachingRDSClient struct {
		client  RDSClient
		ttl     time.Duration
		fetched time.Time

		instanceData map[string]*rds.DescribeDBInstancesOutput
		tagData      map[string]*rds.ListTagsForResourceOutput
	}
)

func (e *CachingRDSClient) DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
	id := *input.DBInstanceIdentifier
	if e.instanceData == nil {
		e.instanceData = map[string]*rds.DescribeDBInstancesOutput{}
	}
	if e.fetched.IsZero() {
		e.fetched = time.Now()
	}
	if time.Since(e.fetched) >= e.ttl {
		e.instanceData = map[string]*rds.DescribeDBInstancesOutput{}
	}
	if _, ok := e.instanceData[id]; !ok {
		response, err := e.client.DescribeDBInstances(input)
		if err != nil {
			return nil, err
		}
		e.instanceData[id] = response
	}
	return e.instanceData[id], nil
}

func (e *CachingRDSClient) ListTagsForResource(input *rds.ListTagsForResourceInput) (*rds.ListTagsForResourceOutput, error) {
	id := *input.ResourceName
	if e.tagData == nil {
		e.tagData = map[string]*rds.ListTagsForResourceOutput{}
	}
	if e.fetched.IsZero() {
		e.fetched = time.Now()
	}
	if time.Since(e.fetched) >= e.ttl {
		e.tagData = map[string]*rds.ListTagsForResourceOutput{}
	}
	if _, ok := e.tagData[id]; !ok {
		response, err := e.client.ListTagsForResource(input)
		if err != nil {
			return nil, err
		}
		e.tagData[id] = response
	}
	return e.tagData[id], nil
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

  ## Specify the metric names to annotate with RDS metadata
  ## By default is configured for "cloudwatch_aws_rds", the default output from the Cloudwatch input plugin
  #metric_names = [ "cloudwatch_aws_rds" ]

  ## Specify the metric tag which contains the RDS DB Instance Identifier
  ## By default is configured for "db_instance_identifier", the default from Cloudwatch input plugin when using the RDS dimension
  #id = "db_instance_identifier"

  ## Enable annotating with RDS DB Instance type
  #instance_type = true
  
  ## Enable annotating with the RDS engine type
  #engine = true
  
  ## Enable annotating with the RDS engine version
  #engine_version = true
  
  ## Specify the RDS Tags to append as metric tags
  #tags = [ "Name" ]
`

func (r *RDS) SampleConfig() string {
	return sampleConfig
}

func (r *RDS) Description() string {
	return "Annotate metrics with AWS RDS metadata"
}

func (r *RDS) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if r.client == nil {
		r.initRdsClient()
	}
	for _, metric := range in {
		if utils.IsSelected(metric, r.MetricNames) {
			r.annotate(metric)
		}
	}
	return in
}

func init() {
	processors.Add("aws_metadata_rds", func() telegraf.Processor {
		return &RDS{
			CacheTTL: internal.Duration{Duration: time.Duration(1 * time.Hour)},
			MetricNames: []string{
				"cloudwatch_aws_rds",
			},
			Id:            "db_instance_identifier",
			Engine:        true,
			EngineVersion: true,
			Tags:          []string{"Name"},
		}
	})
}

func (r *RDS) annotate(metric telegraf.Metric) {
	r.annotateWithInstanceData(metric)
	r.annotateWithTags(metric)
}

func (r *RDS) annotateWithInstanceData(metric telegraf.Metric) {
	name := metric.Tags()[r.Id]
	instance, err := r.getDBInstance(name)
	if err != nil {
		log.Printf("E! %s", err)
		return
	}
	if r.Engine {
		metric.AddTag("engine", *instance.Engine)
	}
	if r.EngineVersion {
		metric.AddTag("engine_version", *instance.EngineVersion)
	}
}

func (r *RDS) annotateWithTags(metric telegraf.Metric) {
	name := metric.Tags()[r.Id]
	instance, err := r.getDBInstance(name)
	if err != nil {
		log.Printf("E! %s", err)
		return
	}
	tags, err := r.client.ListTagsForResource(&rds.ListTagsForResourceInput{
		ResourceName: instance.DBInstanceArn,
	})
	if err != nil {
		log.Printf("E! %s", err)
		return
	}
	for _, tag := range r.Tags {
		for _, it := range tags.TagList {
			if tag == *it.Key {
				metric.AddTag(tag, *it.Value)
				break
			}
		}
	}
}

func (r *RDS) getDBInstance(identifier string) (*rds.DBInstance, error) {
	output, err := r.client.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}
	if len(output.DBInstances) == 0 {
		return nil, fmt.Errorf("DB Instance %s not found", identifier)
	}
	return output.DBInstances[0], nil
}

func (r *RDS) initRdsClient() error {
	credentialConfig := &internalaws.CredentialConfig{
		Region:    r.Region,
		AccessKey: r.AccessKey,
		SecretKey: r.SecretKey,
		RoleARN:   r.RoleARN,
		Profile:   r.Profile,
		Filename:  r.Filename,
		Token:     r.Token,
	}
	configProvider := credentialConfig.Credentials()
	r.client = &CachingRDSClient{
		client: rds.New(configProvider),
		ttl:    r.CacheTTL.Duration,
	}
	return nil
}
