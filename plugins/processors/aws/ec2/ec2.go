package ec2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/reverse_dns/parallel"
)

type AwsEc2Processor struct {
	Tags             []string        `toml:"tags"`
	Timeout          config.Duration `toml:"timeout"`
	Ordered          bool            `toml:"ordered"`
	MaxParallelCalls int             `toml:"max_parallel_calls"`

	Log        telegraf.Logger     `toml:"-"`
	tags       map[string]struct{} `toml:"-"`
	imdsClient *imds.Client
	parallel   parallel.Parallel
}

const sampleConfig = `
  ## Tags to attach to metrics. Available tags:
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
  tags = []

  ## Timeout for http requests made by against aws ec2 metadata endpoint.
  timeout = "10s"

  ## ordered controls whether or not the metrics need to stay in the same order
  ## this plugin received them in. If false, this plugin will change the order
  ## with requests hitting cached results moving through immediately and not
  ## waiting on slower lookups. This may cause issues for you if you are
  ## depending on the order of metrics staying the same. If so, set this to true.
  ## Keeping the metrics ordered may be slightly slower.
  ordered = false
`

const (
	DefaultMaxOrderedQueueSize = 10_000
	DefaultMaxParallelCalls    = 10_000
	DefaultTimeout             = 10 * time.Second
)

var allowedTags = map[string]struct{}{
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

func (r *AwsEc2Processor) Add(metric telegraf.Metric, acc telegraf.Accumulator) error {
	r.parallel.Enqueue(metric)
	return nil
}

func (r *AwsEc2Processor) Init() error {
	r.Log.Debug("Initializing AWS EC2 Processor")

	for _, tag := range r.Tags {
		if len(tag) > 0 && isTagAllowed(tag) {
			r.tags[tag] = struct{}{}
		} else {
			return fmt.Errorf(
				"Not allowed metadata tag specified in configuration: %s", tag,
			)
		}
	}
	if len(r.tags) == 0 {
		return errors.New("No allowed metadata tags specified in configuration")
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return err
	}
	r.imdsClient = imds.NewFromConfig(cfg)

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

	o, err := r.imdsClient.GetInstanceIdentityDocument(
		ctx,
		&imds.GetInstanceIdentityDocumentInput{},
	)
	if err != nil {
		r.Log.Errorf("Error in AWS EC2 Processor: %v", err)
		return []telegraf.Metric{metric}
	}

	for tag := range r.tags {
		metric.AddTag(tag, getTag(o, tag))
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
		tags:             make(map[string]struct{}),
	}
}

func getTag(o *imds.GetInstanceIdentityDocumentOutput, tag string) string {
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

func isTagAllowed(tag string) bool {
	_, ok := allowedTags[tag]
	return ok
}
