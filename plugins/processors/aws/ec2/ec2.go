package ec2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	"github.com/influxdata/telegraf/plugins/processors"
)

type AwsEc2Processor struct {
	ImdsTags         []string        `toml:"imds_tags"`
	EC2Tags          []string        `toml:"ec2_tags"`
	Timeout          config.Duration `toml:"timeout"`
	Ordered          bool            `toml:"ordered"`
	MaxParallelCalls int             `toml:"max_parallel_calls"`

	Log        telegraf.Logger     `toml:"-"`
	imdsClient *imds.Client        `toml:"-"`
	imdsTags   map[string]struct{} `toml:"-"`
	ec2Client  *ec2.Client         `toml:"-"`
	parallel   parallel.Parallel   `toml:"-"`
	instanceID string              `toml:"-"`
}

const sampleConfig = `
  ## Instance identity document tags to attach to metrics.
  ## For more information see:
  ## https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
  ##
  ## Available tags:
  ## * accountId
  ## * architecture
  ## * availabilityZone
  ## * billingProducts
  ## * imageId
  ## * instanceId
  ## * instanceType
  ## * kernelId
  ## * pendingTime
  ## * privateIp
  ## * ramdiskId
  ## * region
  ## * version
  imds_tags = []

  ## EC2 instance tags retrieved with DescribeTags action.
  ## In case tag is empty upon retrieval it's omitted when tagging metrics.
  ## Note that in order for this to work, role attached to EC2 instance or AWS
  ## credentials available from the environment must have a policy attached, that
  ## allows ec2:DescribeTags.
  ##
  ## For more information see:
  ## https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTags.html
  ec2_tags = []

  ## Timeout for http requests made by against aws ec2 metadata endpoint.
  timeout = "10s"

  ## ordered controls whether or not the metrics need to stay in the same order
  ## this plugin received them in. If false, this plugin will change the order
  ## with requests hitting cached results moving through immediately and not
  ## waiting on slower lookups. This may cause issues for you if you are
  ## depending on the order of metrics staying the same. If so, set this to true.
  ## Keeping the metrics ordered may be slightly slower.
  ordered = false

  ## max_parallel_calls is the maximum number of AWS API calls to be in flight
  ## at the same time.
  ## It's probably best to keep this number fairly low.
  max_parallel_calls = 10
`

const (
	DefaultMaxOrderedQueueSize = 10_000
	DefaultMaxParallelCalls    = 10
	DefaultTimeout             = 10 * time.Second
)

var allowedImdsTags = map[string]struct{}{
	"accountId":        {},
	"architecture":     {},
	"availabilityZone": {},
	"billingProducts":  {},
	"imageId":          {},
	"instanceId":       {},
	"instanceType":     {},
	"kernelId":         {},
	"pendingTime":      {},
	"privateIp":        {},
	"ramdiskId":        {},
	"region":           {},
	"version":          {},
}

func (r *AwsEc2Processor) SampleConfig() string {
	return sampleConfig
}

func (r *AwsEc2Processor) Description() string {
	return "Attach AWS EC2 metadata to metrics"
}

func (r *AwsEc2Processor) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	r.parallel.Enqueue(metric)
	return nil
}

func (r *AwsEc2Processor) Init() error {
	r.Log.Debug("Initializing AWS EC2 Processor")
	if len(r.EC2Tags) == 0 && len(r.ImdsTags) == 0 {
		return errors.New("no tags specified in configuration")
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed loading default AWS config: %w", err)
	}
	r.imdsClient = imds.NewFromConfig(cfg)

	iido, err := r.imdsClient.GetInstanceIdentityDocument(
		ctx,
		&imds.GetInstanceIdentityDocumentInput{},
	)
	if err != nil {
		return fmt.Errorf("failed getting instance identity document: %w", err)
	}

	r.instanceID = iido.InstanceID

	if len(r.EC2Tags) > 0 {
		// Add region to AWS config when creating EC2 service client since it's required.
		cfg.Region = iido.Region

		r.ec2Client = ec2.NewFromConfig(cfg)

		// Chceck if instance is allowed to call DescribeTags.
		_, err = r.ec2Client.DescribeTags(ctx, &ec2.DescribeTagsInput{
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
	}

	for _, tag := range r.ImdsTags {
		if len(tag) > 0 && isImdsTagAllowed(tag) {
			r.imdsTags[tag] = struct{}{}
		} else {
			return fmt.Errorf("not allowed metadata tag specified in configuration: %s", tag)
		}
	}
	if len(r.imdsTags) == 0 && len(r.EC2Tags) == 0 {
		return errors.New("no allowed metadata tags specified in configuration")
	}

	return nil
}

func (r *AwsEc2Processor) Start(acc telegraf.Accumulator) error {
	if r.Ordered {
		r.parallel = parallel.NewOrdered(acc, r.asyncAdd, DefaultMaxOrderedQueueSize, r.MaxParallelCalls)
	} else {
		r.parallel = parallel.NewUnordered(acc, r.asyncAdd, r.MaxParallelCalls)
	}

	return nil
}

func (r *AwsEc2Processor) Stop() error {
	if r.parallel == nil {
		return errors.New("Trying to stop unstarted AWS EC2 Processor")
	}
	r.parallel.Stop()
	return nil
}

func (r *AwsEc2Processor) asyncAdd(metric telegraf.Metric) []telegraf.Metric {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Timeout))
	defer cancel()

	// Add IMDS Instance Identity Document tags.
	if len(r.imdsTags) > 0 {
		iido, err := r.imdsClient.GetInstanceIdentityDocument(
			ctx,
			&imds.GetInstanceIdentityDocumentInput{},
		)
		if err != nil {
			r.Log.Errorf("Error when calling GetInstanceIdentityDocument: %v", err)
			return []telegraf.Metric{metric}
		}

		for tag := range r.imdsTags {
			if v := getTagFromInstanceIdentityDocument(iido, tag); v != "" {
				metric.AddTag(tag, v)
			}
		}
	}

	// Add EC2 instance tags.
	if len(r.EC2Tags) > 0 {
		dto, err := r.ec2Client.DescribeTags(ctx, &ec2.DescribeTagsInput{
			Filters: createFilterFromTags(r.instanceID, r.EC2Tags),
		})
		if err != nil {
			r.Log.Errorf("Error during EC2 DescribeTags: %v", err)
			return []telegraf.Metric{metric}
		}

		for _, tag := range r.EC2Tags {
			if v := getTagFromDescribeTags(dto, tag); v != "" {
				metric.AddTag(tag, v)
			}
		}
	}

	return []telegraf.Metric{metric}
}

func init() {
	processors.AddStreaming("aws_ec2", func() telegraf.StreamingProcessor {
		return newAwsEc2Processor()
	})
}

func newAwsEc2Processor() *AwsEc2Processor {
	return &AwsEc2Processor{
		MaxParallelCalls: DefaultMaxParallelCalls,
		Timeout:          config.Duration(DefaultTimeout),
		imdsTags:         make(map[string]struct{}),
	}
}

func createFilterFromTags(instanceID string, tagNames []string) []types.Filter {
	return []types.Filter{
		{
			Name:   aws.String("resource-id"),
			Values: []string{instanceID},
		},
		{
			Name:   aws.String("key"),
			Values: tagNames,
		},
	}
}

func getTagFromDescribeTags(o *ec2.DescribeTagsOutput, tag string) string {
	for _, t := range o.Tags {
		if *t.Key == tag {
			return *t.Value
		}
	}
	return ""
}

func getTagFromInstanceIdentityDocument(o *imds.GetInstanceIdentityDocumentOutput, tag string) string {
	switch tag {
	case "accountId":
		return o.AccountID
	case "architecture":
		return o.Architecture
	case "availabilityZone":
		return o.AvailabilityZone
	case "billingProducts":
		return strings.Join(o.BillingProducts, ",")
	case "imageId":
		return o.ImageID
	case "instanceId":
		return o.InstanceID
	case "instanceType":
		return o.InstanceType
	case "kernelId":
		return o.KernelID
	case "pendingTime":
		return o.PendingTime.String()
	case "privateIp":
		return o.PrivateIP
	case "ramdiskId":
		return o.RamdiskID
	case "region":
		return o.Region
	case "version":
		return o.Version
	default:
		return ""
	}
}

func isImdsTagAllowed(tag string) bool {
	_, ok := allowedImdsTags[tag]
	return ok
}
