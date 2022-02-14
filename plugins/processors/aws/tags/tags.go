package tags

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	internalAWS "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	"github.com/influxdata/telegraf/plugins/processors"
)

type AwsTagsProcessor struct {
	Tags             []string        `toml:"tags"`
	Timeout          config.Duration `toml:"timeout"`
	MaxCacheAge      config.Duration `toml:"max_cache_age"`
	Ordered          bool            `toml:"ordered"`
	MaxParallelCalls int             `toml:"max_parallel_calls"`

	internalAWS.CredentialConfig

	Log telegraf.Logger `toml:"-"`

	tagCache  *TagCache
	ec2Client *ec2.Client
	parallel  parallel.Parallel
}

const sampleConfig = `
  ## Amazon Region
  region = "eu-central-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Web identity provider credentials via STS if role_arn and web_identity_token_file are specified
  ## 2) Assumed credentials via STS if role_arn is specified
  ## 3) explicit credentials from 'access_key' and 'secret_key'
  ## 4) shared profile from 'profile'
  ## 5) environment variables
  ## 6) shared credentials file
  ## 7) EC2 Instance Profile
  # access_key = ""
  # secret_key = ""
  # token = ""
  # role_arn = ""
  # web_identity_token_file = ""
  # role_session_name = ""
  # profile = ""
  # shared_credential_file = ""

  ## EC2 instance tags retrieved with DescribeTags action.
  ## In case tag is empty upon retrieval it's omitted when tagging metrics.
  ## Note that in order for this to work, role attached to EC2 instance or AWS
  ## credentials available from the environment must have a policy attached, that
  ## allows ec2:DescribeTags.
  ##
  ## For more information see:
  ## https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTags.html
  tags = []

  ## Timeout for http requests made against aws ec2 metadata endpoint.
  timeout = "10s"

  ## Maximum age of cached data after which it will be refreshed. Please note that refreshing
  ## the data takes an additional metrics run since it is done asynchronously to not delay metrics.
  ## e.g. the first request after the age expired will still use the cached data but the refresh
  ## will start in the background.
  ## It is recommended to keep this value high to avoid charges for API requests.
  max_cache_age = "1h"

  ## ordered controls whether or not the metrics need to stay in the same order
  ## this plugin received them in. If false, this plugin will change the order
  ## with requests hitting cached results moving through immediately and not
  ## waiting on slower lookups. This may cause issues for you if you are
  ## depending on the order of metrics staying the same. If so, set this to true.
  ## Keeping the metrics ordered may be slightly slower.
  ordered = false

  ## max_parallel_calls is the maximum number of AWS API calls to be in flight
  ## at the same time.
  max_parallel_calls = 10
`

const (
	DefaultMaxOrderedQueueSize = 10_000
	DefaultMaxParallelCalls    = 10
	DefaultTimeout             = 10 * time.Second
)

func (r *AwsTagsProcessor) SampleConfig() string {
	return sampleConfig
}

func (r *AwsTagsProcessor) Description() string {
	return "Attach AWS EC2 metadata to metrics"
}

func (r *AwsTagsProcessor) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	r.parallel.Enqueue(metric)
	return nil
}

func (r *AwsTagsProcessor) Init() error {
	r.Log.Debug("Initializing AWS Tags Processor")
	if len(r.Tags) == 0 {
		return errors.New("no tags specified in configuration")
	}
	r.tagCache = NewTagCache(time.Duration(r.MaxCacheAge), time.Duration(r.Timeout), r.MaxParallelCalls, nil, r.Log)
	return nil
}

func (r *AwsTagsProcessor) Start(acc telegraf.Accumulator) error {
	ctx := context.Background()
	cfg, err := r.CredentialConfig.Credentials()
	if err != nil {
		return fmt.Errorf("failed loading default AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	_, err = ec2Client.DescribeTags(ctx, &ec2.DescribeTagsInput{
		DryRun: true,
	})
	var ae smithy.APIError
	if errors.As(err, &ae) {
		if ae.ErrorCode() != "DryRunOperation" {
			return fmt.Errorf("instance doesn't have permissions to call DescribeTags: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error calling DescribeTags: %w", err)
	}

	r.tagCache.ec2Client = ec2Client

	if r.Ordered {
		r.parallel = parallel.NewOrdered(acc, r.asyncAdd, DefaultMaxOrderedQueueSize, r.MaxParallelCalls)
	} else {
		r.parallel = parallel.NewUnordered(acc, r.asyncAdd, r.MaxParallelCalls)
	}

	return nil
}

func (r *AwsTagsProcessor) Stop() error {
	if r.parallel == nil {
		return errors.New("trying to stop un-started AWS EC2 Processor")
	}
	r.parallel.Stop()
	return nil
}

func (r *AwsTagsProcessor) asyncAdd(metric telegraf.Metric) []telegraf.Metric {
	instanceId := getInstanceId(metric)
	if instanceId == "" {
		r.Log.Errorf("unable to determine instance id, unknown metric name %s", metric.Name())
		return []telegraf.Metric{metric}
	}

	tags := r.tagCache.Get(instanceId)
	tagKeys := tags.Keys()
	if len(tagKeys) == 0 {
		r.Log.Warnf("received empty list of tags for %s", instanceId)
	}

	for _, tag := range tagKeys {
		if r.wantedTag(tag) {
			value := tags.Value(tag)
			if value != "" {
				metric.AddTag(tag, tags.Value(tag))
			}
		}
	}

	return []telegraf.Metric{metric}
}

func (r *AwsTagsProcessor) wantedTag(tag string) bool {
	for _, wantedTag := range r.Tags {
		if wantedTag == tag {
			return true
		}
	}
	return false
}

func getInstanceId(metric telegraf.Metric) string {
	// TODO: add other metric names
	switch metric.Name() {
	case "cloudwatch_aws_nat_gateway":
		return metric.Tags()["nat_gateway_id"]
	default:
		return ""
	}
}

func init() {
	processors.AddStreaming("aws_tags", func() telegraf.StreamingProcessor {
		return newAwsTagsProcessor()
	})
}

func newAwsTagsProcessor() *AwsTagsProcessor {
	return &AwsTagsProcessor{
		MaxParallelCalls: DefaultMaxParallelCalls,
		Timeout:          config.Duration(DefaultTimeout),
	}
}
