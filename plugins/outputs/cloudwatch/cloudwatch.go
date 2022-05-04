package cloudwatch

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type CloudWatch struct {
	Namespace             string `toml:"namespace"` // CloudWatch Metrics Namespace
	HighResolutionMetrics bool   `toml:"high_resolution_metrics"`
	svc                   *cloudwatch.Client

	WriteStatistics bool `toml:"write_statistics"`

	Log telegraf.Logger `toml:"-"`

	internalaws.CredentialConfig
}

type statisticType int

const (
	statisticTypeNone statisticType = iota
	statisticTypeMax
	statisticTypeMin
	statisticTypeSum
	statisticTypeCount
)

type cloudwatchField interface {
	addValue(sType statisticType, value float64)
	buildDatum() []types.MetricDatum
}

type statisticField struct {
	metricName        string
	fieldName         string
	tags              map[string]string
	values            map[statisticType]float64
	timestamp         time.Time
	storageResolution int64
}

func (f *statisticField) addValue(sType statisticType, value float64) {
	if sType != statisticTypeNone {
		f.values[sType] = value
	}
}

func (f *statisticField) buildDatum() []types.MetricDatum {
	var datums []types.MetricDatum

	if f.hasAllFields() {
		// If we have all required fields, we build datum with StatisticValues
		min := f.values[statisticTypeMin]
		max := f.values[statisticTypeMax]
		sum := f.values[statisticTypeSum]
		count := f.values[statisticTypeCount]

		datum := types.MetricDatum{
			MetricName: aws.String(strings.Join([]string{f.metricName, f.fieldName}, "_")),
			Dimensions: BuildDimensions(f.tags),
			Timestamp:  aws.Time(f.timestamp),
			StatisticValues: &types.StatisticSet{
				Minimum:     aws.Float64(min),
				Maximum:     aws.Float64(max),
				Sum:         aws.Float64(sum),
				SampleCount: aws.Float64(count),
			},
			StorageResolution: aws.Int32(int32(f.storageResolution)),
		}

		datums = append(datums, datum)
	} else {
		// If we don't have all required fields, we build each field as independent datum
		for sType, value := range f.values {
			datum := types.MetricDatum{
				Value:      aws.Float64(value),
				Dimensions: BuildDimensions(f.tags),
				Timestamp:  aws.Time(f.timestamp),
			}

			switch sType {
			case statisticTypeMin:
				datum.MetricName = aws.String(strings.Join([]string{f.metricName, f.fieldName, "min"}, "_"))
			case statisticTypeMax:
				datum.MetricName = aws.String(strings.Join([]string{f.metricName, f.fieldName, "max"}, "_"))
			case statisticTypeSum:
				datum.MetricName = aws.String(strings.Join([]string{f.metricName, f.fieldName, "sum"}, "_"))
			case statisticTypeCount:
				datum.MetricName = aws.String(strings.Join([]string{f.metricName, f.fieldName, "count"}, "_"))
			default:
				// should not be here
				continue
			}

			datums = append(datums, datum)
		}
	}

	return datums
}

func (f *statisticField) hasAllFields() bool {
	_, hasMin := f.values[statisticTypeMin]
	_, hasMax := f.values[statisticTypeMax]
	_, hasSum := f.values[statisticTypeSum]
	_, hasCount := f.values[statisticTypeCount]

	return hasMin && hasMax && hasSum && hasCount
}

type valueField struct {
	metricName        string
	fieldName         string
	tags              map[string]string
	value             float64
	timestamp         time.Time
	storageResolution int64
}

func (f *valueField) addValue(sType statisticType, value float64) {
	if sType == statisticTypeNone {
		f.value = value
	}
}

func (f *valueField) buildDatum() []types.MetricDatum {
	return []types.MetricDatum{
		{
			MetricName:        aws.String(strings.Join([]string{f.metricName, f.fieldName}, "_")),
			Value:             aws.Float64(f.value),
			Dimensions:        BuildDimensions(f.tags),
			Timestamp:         aws.Time(f.timestamp),
			StorageResolution: aws.Int32(int32(f.storageResolution)),
		},
	}
}

func (c *CloudWatch) Connect() error {
	cfg, err := c.CredentialConfig.Credentials()
	if err != nil {
		return err
	}

	c.svc = cloudwatch.NewFromConfig(cfg)
	return nil
}

func (c *CloudWatch) Close() error {
	return nil
}

func (c *CloudWatch) Write(metrics []telegraf.Metric) error {
	var datums []types.MetricDatum
	for _, m := range metrics {
		d := BuildMetricDatum(c.WriteStatistics, c.HighResolutionMetrics, m)
		datums = append(datums, d...)
	}

	const maxDatumsPerCall = 20 // PutMetricData only supports up to 20 data metrics per call

	for _, partition := range PartitionDatums(maxDatumsPerCall, datums) {
		err := c.WriteToCloudWatch(partition)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CloudWatch) WriteToCloudWatch(datums []types.MetricDatum) error {
	params := &cloudwatch.PutMetricDataInput{
		MetricData: datums,
		Namespace:  aws.String(c.Namespace),
	}

	_, err := c.svc.PutMetricData(context.Background(), params)

	if err != nil {
		c.Log.Errorf("Unable to write to CloudWatch : %+v", err.Error())
	}

	return err
}

// PartitionDatums partitions the MetricDatums into smaller slices of a max size so that are under the limit
// for the AWS API calls.
func PartitionDatums(size int, datums []types.MetricDatum) [][]types.MetricDatum {
	numberOfPartitions := len(datums) / size
	if len(datums)%size != 0 {
		numberOfPartitions++
	}

	partitions := make([][]types.MetricDatum, numberOfPartitions)

	for i := 0; i < numberOfPartitions; i++ {
		start := size * i
		end := size * (i + 1)
		if end > len(datums) {
			end = len(datums)
		}

		partitions[i] = datums[start:end]
	}

	return partitions
}

// BuildMetricDatum makes a MetricDatum from telegraf.Metric. It would check if all required fields of
// cloudwatch.StatisticSet are available. If so, it would build MetricDatum from statistic values.
// Otherwise, fields would still been built independently.
func BuildMetricDatum(buildStatistic bool, highResolutionMetrics bool, point telegraf.Metric) []types.MetricDatum {
	fields := make(map[string]cloudwatchField)
	tags := point.Tags()
	storageResolution := int64(60)
	if highResolutionMetrics {
		storageResolution = 1
	}

	for k, v := range point.Fields() {
		val, ok := convert(v)
		if !ok {
			// Only fields with values that can be converted to float64 (and within CloudWatch boundary) are supported.
			// Non-supported fields are skipped.
			continue
		}

		sType, fieldName := getStatisticType(k)

		// If statistic metric is not enabled or non-statistic type, just take current field as a value field.
		if !buildStatistic || sType == statisticTypeNone {
			fields[k] = &valueField{
				metricName:        point.Name(),
				fieldName:         k,
				tags:              tags,
				timestamp:         point.Time(),
				value:             val,
				storageResolution: storageResolution,
			}
			continue
		}

		// Otherwise, it shall be a statistic field.
		if _, ok := fields[fieldName]; !ok {
			// Hit an uncached field, create statisticField for first time
			fields[fieldName] = &statisticField{
				metricName: point.Name(),
				fieldName:  fieldName,
				tags:       tags,
				timestamp:  point.Time(),
				values: map[statisticType]float64{
					sType: val,
				},
				storageResolution: storageResolution,
			}
		} else {
			// Add new statistic value to this field
			fields[fieldName].addValue(sType, val)
		}
	}

	var datums []types.MetricDatum
	for _, f := range fields {
		d := f.buildDatum()
		datums = append(datums, d...)
	}

	return datums
}

// BuildDimensions makes a list of Dimensions by using a Point's tags. CloudWatch supports up to
// 10 dimensions per metric, so we only keep up to the first 10 alphabetically.
// This always includes the "host" tag if it exists.
func BuildDimensions(mTags map[string]string) []types.Dimension {
	const maxDimensions = 10
	dimensions := make([]types.Dimension, 0, maxDimensions)

	// This is pretty ugly, but we always want to include the "host" tag if it exists.
	if host, ok := mTags["host"]; ok {
		dimensions = append(dimensions, types.Dimension{
			Name:  aws.String("host"),
			Value: aws.String(host),
		})
	}

	var keys []string
	for k := range mTags {
		if k != "host" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		if len(dimensions) >= maxDimensions {
			break
		}

		value := mTags[k]
		if value == "" {
			continue
		}

		dimensions = append(dimensions, types.Dimension{
			Name:  aws.String(k),
			Value: aws.String(mTags[k]),
		})
	}

	return dimensions
}

func getStatisticType(name string) (sType statisticType, fieldName string) {
	switch {
	case strings.HasSuffix(name, "_max"):
		sType = statisticTypeMax
		fieldName = strings.TrimSuffix(name, "_max")
	case strings.HasSuffix(name, "_min"):
		sType = statisticTypeMin
		fieldName = strings.TrimSuffix(name, "_min")
	case strings.HasSuffix(name, "_sum"):
		sType = statisticTypeSum
		fieldName = strings.TrimSuffix(name, "_sum")
	case strings.HasSuffix(name, "_count"):
		sType = statisticTypeCount
		fieldName = strings.TrimSuffix(name, "_count")
	default:
		sType = statisticTypeNone
		fieldName = name
	}

	return sType, fieldName
}

func convert(v interface{}) (value float64, ok bool) {
	ok = true

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
		ok = false
		return value, ok
	}

	// Do CloudWatch boundary checking
	// Constraints at: http://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricDatum.html
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

	return value, ok
}

func init() {
	outputs.Add("cloudwatch", func() telegraf.Output {
		return &CloudWatch{}
	})
}
