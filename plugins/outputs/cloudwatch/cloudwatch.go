//go:generate ../../../tools/readme_config_includer/generator
package cloudwatch

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/influxdata/telegraf"
	common_aws "github.com/influxdata/telegraf/plugins/common/aws"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	// Cloudwatch only supports up to 1000 data metrics per call according to
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutMetricData.html
	maxBatchSize = 1000
	// Cloudwatch only accepts up to 30 dimensions per metric according to
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html#usingDimensions
	maxDimensions = 30
)

type CloudWatch struct {
	Namespace             string          `toml:"namespace"` // CloudWatch Metrics Namespace
	HighResolutionMetrics bool            `toml:"high_resolution_metrics"`
	WriteStatistics       bool            `toml:"write_statistics"`
	MaxDimensions         int             `toml:"max_dimensions"`
	Log                   telegraf.Logger `toml:"-"`
	common_aws.CredentialConfig
	common_http.HTTPClientConfig

	client     *http.Client
	svc        *cloudwatch.Client
	resolution int64
}

func (*CloudWatch) SampleConfig() string {
	return sampleConfig
}

func (c *CloudWatch) Init() error {
	// Check user settings
	if c.Namespace == "" {
		return errors.New("namespace is required")
	}

	if c.MaxDimensions < 0 || c.MaxDimensions > maxDimensions {
		return fmt.Errorf("number of dimensions has to be between 0 and %d", maxDimensions)
	}

	// Determine the metric resolution
	c.resolution = 60
	if c.HighResolutionMetrics {
		c.resolution = 1
	}

	return nil
}

func (c *CloudWatch) Connect() error {
	cfg, err := c.CredentialConfig.Credentials()
	if err != nil {
		return err
	}

	ctx := context.Background()
	client, err := c.HTTPClientConfig.CreateClient(ctx, c.Log)
	if err != nil {
		return err
	}

	c.client = client
	c.svc = cloudwatch.NewFromConfig(cfg, func(options *cloudwatch.Options) {
		options.HTTPClient = c.client
	})

	return nil
}

func (c *CloudWatch) Close() error {
	if c.client != nil {
		c.client.CloseIdleConnections()
	}

	return nil
}

func (c *CloudWatch) Write(metrics []telegraf.Metric) error {
	datums := make([]types.MetricDatum, 0, len(metrics))
	for _, m := range metrics {
		d := c.buildMetricDatum(m)
		datums = append(datums, d...)
	}

	for _, partition := range partitionDatums(datums, maxBatchSize) {
		params := &cloudwatch.PutMetricDataInput{
			MetricData: partition,
			Namespace:  aws.String(c.Namespace),
		}

		if _, err := c.svc.PutMetricData(context.Background(), params); err != nil {
			return fmt.Errorf("unable to write to CloudWatch: %w", err)
		}
	}

	return nil
}

func (c *CloudWatch) buildMetricDatum(m telegraf.Metric) []types.MetricDatum {
	// Extract the dimensions from tags
	dimensions := c.buildDimensions(m.TagList())

	// Aggregate the metric values into statistics if enabled
	fields := make(map[string]cloudwatchField, len(m.FieldList()))
	for _, f := range m.FieldList() {
		val, ok := convert(f.Value)
		if !ok {
			// Skip over fields that cannot be converted to float64 or did not
			// pass the CloudWatch boundary check
			continue
		}

		// Determine the field name and type of statistic if any
		fieldName := f.Key
		sType := statisticTypeNone
		switch {
		case strings.HasSuffix(f.Key, "_max"):
			sType = statisticTypeMax
			fieldName = strings.TrimSuffix(f.Key, "_max")
		case strings.HasSuffix(f.Key, "_min"):
			sType = statisticTypeMin
			fieldName = strings.TrimSuffix(f.Key, "_min")
		case strings.HasSuffix(f.Key, "_sum"):
			sType = statisticTypeSum
			fieldName = strings.TrimSuffix(f.Key, "_sum")
		case strings.HasSuffix(f.Key, "_count"):
			sType = statisticTypeCount
			fieldName = strings.TrimSuffix(f.Key, "_count")
		}

		if !c.WriteStatistics || sType == statisticTypeNone {
			// The statistic metric is not enabled or non-statistic type, just
			// use the current field
			fields[f.Key] = &valueField{
				measurement: m.Name(),
				name:        f.Key,
				dimensions:  dimensions,
				timestamp:   m.Time(),
				value:       val,
				resolution:  c.resolution,
			}
		} else if _, ok := fields[fieldName]; !ok {
			// Add a new statistic field
			fields[fieldName] = &statisticField{
				measurement: m.Name(),
				name:        fieldName,
				dimensions:  dimensions,
				timestamp:   m.Time(),
				values:      map[statisticType]float64{sType: val},
				resolution:  c.resolution,
			}
		} else {
			// Aggregate
			fields[fieldName].addValue(sType, val)
		}
	}

	// The buildDatum function returns at most one entry per statistic type for
	// each field so allocate the maximum amount of entries
	datums := make([]types.MetricDatum, 0, 4*len(fields))
	for _, f := range fields {
		datums = append(datums, f.buildDatum()...)
	}

	return datums
}

func (c *CloudWatch) buildDimensions(tags []*telegraf.Tag) []types.Dimension {
	dimensions := make([]types.Dimension, 0, c.MaxDimensions)
	if c.MaxDimensions == 0 {
		return dimensions
	}

	// Make sure we add the "host" tag if any
	for _, t := range tags {
		if t.Key != "host" {
			continue
		}
		dimensions = append(dimensions, types.Dimension{
			Name:  aws.String("host"),
			Value: aws.String(t.Value),
		})
		break
	}

	// Add more tags until we reach the maximum
	// NOTE: The tag-list is already sorted so no need to sort it again
	for _, t := range tags {
		if len(dimensions) >= c.MaxDimensions {
			break
		}
		if t.Key == "host" || t.Value == "" {
			continue
		}
		dimensions = append(dimensions, types.Dimension{
			Name:  aws.String(t.Key),
			Value: aws.String(t.Value),
		})
	}

	return dimensions
}

func partitionDatums(datums []types.MetricDatum, batchSize int) [][]types.MetricDatum {
	// Partition all given metrics into batches with the given batch size
	numberOfPartitions := len(datums) / batchSize
	if len(datums)%batchSize != 0 {
		numberOfPartitions++
	}

	partitions := make([][]types.MetricDatum, numberOfPartitions)
	for i := range numberOfPartitions {
		start := batchSize * i
		end := min(batchSize*(i+1), len(datums))

		partitions[i] = datums[start:end]
	}

	return partitions
}

func convert(v interface{}) (float64, bool) {
	var value float64

	switch t := v.(type) {
	case int:
		value = float64(t)
	case int32:
		value = float64(t)
	case int64:
		value = float64(t)
	case uint64:
		value = float64(t)
	case float64:
		value = t
	case bool:
		if t {
			value = 1
		} else {
			value = 0
		}
	case time.Time:
		value = float64(t.Unix())
	default:
		// Skip unsupported type.
		return value, false
	}

	// Do CloudWatch boundary checking according to
	// http://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricDatum.html
	switch {
	case math.IsNaN(value):
		return 0, false
	case math.IsInf(value, 0):
		return 0, false
	case value > 0 && value < float64(8.515920e-109):
		return 0, false
	case value > float64(1.174271e+108):
		return 0, false
	}

	return value, true
}

func init() {
	outputs.Add("cloudwatch", func() telegraf.Output {
		return &CloudWatch{
			MaxDimensions: 10, // for backward compatibility
		}
	})
}
