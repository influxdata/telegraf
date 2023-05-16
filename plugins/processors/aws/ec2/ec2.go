//go:generate ../../../../tools/readme_config_includer/generator
package ec2

import (
	"context"
	_ "embed"
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
	"github.com/coocood/freecache"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type AwsEc2Processor struct {
	ImdsTags         []string        `toml:"imds_tags"`
	EC2Tags          []string        `toml:"ec2_tags"`
	Timeout          config.Duration `toml:"timeout"`
	CacheTTL         config.Duration `toml:"cache_ttl"`
	Ordered          bool            `toml:"ordered"`
	MaxParallelCalls int             `toml:"max_parallel_calls"`
	Log              telegraf.Logger `toml:"-"`
	TagCacheSize     int             `toml:"tag_cache_size"`
	LogCacheStats    bool            `toml:"log_cache_stats"`

	tagCache *freecache.Cache

	imdsClient          *imds.Client
	imdsTagsMap         map[string]struct{}
	ec2Client           *ec2.Client
	parallel            parallel.Parallel
	instanceID          string
	cancelCleanupWorker context.CancelFunc
}

const (
	DefaultMaxOrderedQueueSize = 10_000
	DefaultMaxParallelCalls    = 10
	DefaultTimeout             = 10 * time.Second
	DefaultCacheTTL            = 0 * time.Hour
	DefaultCacheSize           = 1000
	DefaultLogCacheStats       = false
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

func (*AwsEc2Processor) SampleConfig() string {
	return sampleConfig
}

func (r *AwsEc2Processor) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	r.parallel.Enqueue(metric)
	return nil
}

func (r *AwsEc2Processor) logCacheStatistics(ctx context.Context) {
	if r.tagCache == nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.Log.Debugf("cache: size=%d hit=%d miss=%d full=%d\n",
				r.tagCache.EntryCount(),
				r.tagCache.HitCount(),
				r.tagCache.MissCount(),
				r.tagCache.EvacuateCount(),
			)
			r.tagCache.ResetStatistics()
		}
	}
}

func (r *AwsEc2Processor) Init() error {
	r.Log.Debug("Initializing AWS EC2 Processor")
	if len(r.EC2Tags) == 0 && len(r.ImdsTags) == 0 {
		return errors.New("no tags specified in configuration")
	}

	for _, tag := range r.ImdsTags {
		if len(tag) == 0 || !isImdsTagAllowed(tag) {
			return fmt.Errorf("not allowed metadata tag specified in configuration: %s", tag)
		}
		r.imdsTagsMap[tag] = struct{}{}
	}
	if len(r.imdsTagsMap) == 0 && len(r.EC2Tags) == 0 {
		return errors.New("no allowed metadata tags specified in configuration")
	}

	return nil
}

func (r *AwsEc2Processor) Start(acc telegraf.Accumulator) error {
	r.tagCache = freecache.NewCache(r.TagCacheSize)
	if r.LogCacheStats {
		ctx, cancel := context.WithCancel(context.Background())
		r.cancelCleanupWorker = cancel
		go r.logCacheStatistics(ctx)
	}

	r.Log.Debugf("cache: size=%d\n", r.TagCacheSize)
	if r.CacheTTL > 0 {
		r.Log.Debugf("cache timeout: seconds=%d\n", int(time.Duration(r.CacheTTL).Seconds()))
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

		// Check if instance is allowed to call DescribeTags.
		_, err = r.ec2Client.DescribeTags(ctx, &ec2.DescribeTagsInput{
			DryRun: aws.Bool(true),
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

	if r.Ordered {
		r.parallel = parallel.NewOrdered(acc, r.asyncAdd, DefaultMaxOrderedQueueSize, r.MaxParallelCalls)
	} else {
		r.parallel = parallel.NewUnordered(acc, r.asyncAdd, r.MaxParallelCalls)
	}

	return nil
}

func (r *AwsEc2Processor) Stop() {
	if r.parallel != nil {
		r.parallel.Stop()
	}
	r.cancelCleanupWorker()
}

func (r *AwsEc2Processor) LookupIMDSTags(metric telegraf.Metric) telegraf.Metric {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Timeout))
	defer cancel()

	var tagsNotFound []string

	for tag := range r.imdsTagsMap {
		val, err := r.tagCache.Get([]byte(tag))
		if err != nil {
			tagsNotFound = append(tagsNotFound, tag)
		} else {
			metric.AddTag(tag, string(val))
		}
	}

	if len(tagsNotFound) == 0 {
		return metric
	}

	iido, err := r.imdsClient.GetInstanceIdentityDocument(
		ctx,
		&imds.GetInstanceIdentityDocumentInput{},
	)

	if err != nil {
		r.Log.Errorf("Error when calling GetInstanceIdentityDocument: %v", err)
		return metric
	}

	for _, tag := range tagsNotFound {
		if v := getTagFromInstanceIdentityDocument(iido, tag); v != "" {
			metric.AddTag(tag, v)
			expiration := int(time.Duration(r.CacheTTL).Seconds())
			err = r.tagCache.Set([]byte(tag), []byte(v), expiration)
			if err != nil {
				r.Log.Errorf("Error when setting IMDS tag cache value: %v", err)
			}
		}
	}

	return metric
}

func (r *AwsEc2Processor) LookupEC2Tags(metric telegraf.Metric) telegraf.Metric {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Timeout))
	defer cancel()

	var tagsNotFound []string

	for _, tag := range r.EC2Tags {
		val, err := r.tagCache.Get([]byte(tag))
		if err != nil {
			tagsNotFound = append(tagsNotFound, tag)
		} else {
			metric.AddTag(tag, string(val))
		}
	}

	if len(tagsNotFound) == 0 {
		return metric
	}

	dto, err := r.ec2Client.DescribeTags(ctx, &ec2.DescribeTagsInput{
		Filters: createFilterFromTags(r.instanceID, r.EC2Tags),
	})

	if err != nil {
		r.Log.Errorf("Error during EC2 DescribeTags: %v", err)
		return metric
	}

	for _, tag := range r.EC2Tags {
		if v := getTagFromDescribeTags(dto, tag); v != "" {
			metric.AddTag(tag, v)
			expiration := int(time.Duration(r.CacheTTL).Seconds())
			err = r.tagCache.Set([]byte(tag), []byte(v), expiration)
			if err != nil {
				r.Log.Errorf("Error when setting EC2Tags tag cache value: %v", err)
			}
		}
	}

	return metric
}

func (r *AwsEc2Processor) asyncAdd(metric telegraf.Metric) []telegraf.Metric {
	// Add IMDS Instance Identity Document tags.
	if len(r.imdsTagsMap) > 0 {
		metric = r.LookupIMDSTags(metric)
	}

	// Add EC2 instance tags.
	if len(r.EC2Tags) > 0 {
		metric = r.LookupEC2Tags(metric)
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
		TagCacheSize:     DefaultCacheSize,
		Timeout:          config.Duration(DefaultTimeout),
		CacheTTL:         config.Duration(DefaultCacheTTL),
		imdsTagsMap:      make(map[string]struct{}),
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
