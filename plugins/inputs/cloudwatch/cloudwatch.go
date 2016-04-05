package cloudwatch

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/azer/snakecase"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
	"strings"
	"time"
)

type Cloudwatch struct {
	Region     string
	Namespaces []string

	// List of metrics to gather statistics for. Computed the first time Gather() is called.
	metrics []*cloudwatch.Metric
	// Get metrics statistics since last time Gather() was called.
	lastTime *time.Time
}

var sampleConfig = `
  ## AWS region
  region = "us-east-1"
  ## specify namespaces as strings
  namespaces = ["AWS/EC2", "AWS/DynamoDB"]
`

func (r *Cloudwatch) SampleConfig() string {
	return sampleConfig
}

func (r *Cloudwatch) Description() string {
	return "Read metrics from AWS CloudWatch"
}

// Reads stats from AWS CloudWatch. Accumulates stats.
// Returns one of the errors encountered while gathering stats (if any).
func (r *Cloudwatch) Gather(acc telegraf.Accumulator) error {
	svc := cloudwatch.New(session.New(), &aws.Config{Region: aws.String(r.Region)})
	var outerr error
	now := time.Now()
	if r.lastTime == nil {
		r.metrics, outerr = listMetrics(svc, r.Namespaces)
		log.Printf("Found %d cloudwatch metrics", len(r.metrics))
	} else {
		for i := 0; i < len(r.metrics); i++ {
			metric := r.metrics[i]
			minutes := int64(now.Sub(*r.lastTime) / time.Minute)
			if minutes == 0 {
				minutes = 1
			}
			period := minutes * 60
			outerr = gatherMetric(svc, r.Region, acc, metric, *r.lastTime, now, period)
		}
	}
	r.lastTime = &now
	return outerr
}

type listMetricsAPI interface {
	ListMetrics(*cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error)
}

func listMetrics(svc listMetricsAPI, namespaces []string) ([]*cloudwatch.Metric, error) {
	metrics := []*cloudwatch.Metric{}
	for i := 0; i < len(namespaces); i++ {
		namespace := namespaces[i]
		params := &cloudwatch.ListMetricsInput{
			Dimensions: []*cloudwatch.DimensionFilter{},
			MetricName: nil,
			Namespace:  aws.String(namespace),
			NextToken:  nil,
		}
		resp, err := svc.ListMetrics(params)
		if err != nil {
			return []*cloudwatch.Metric{}, err
		}
		metrics = append(metrics, resp.Metrics...)
	}
	return metrics, nil

}

type getMetricStatisticsAPI interface {
	GetMetricStatistics(*cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error)
}

func gatherMetric(svc getMetricStatisticsAPI, region string, acc telegraf.Accumulator, metric *cloudwatch.Metric, startTime time.Time, endTime time.Time, period int64) error {
	r, err := svc.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		Namespace:  metric.Namespace,
		MetricName: metric.MetricName,
		Dimensions: metric.Dimensions,
		Statistics: []*string{
			aws.String("Average"),
			aws.String("Maximum"),
			aws.String("Minimum"),
			aws.String("Sum"),
			aws.String("SampleCount")},
		StartTime: &startTime,
		EndTime:   &endTime,
		Period:    &period,
	})
	if err != nil {
		return err
	}
	if len(r.Datapoints) == 0 {
		return nil
	}
	if len(r.Datapoints) != 1 {
		return errors.New(fmt.Sprintf("Expected one datapoint, received %d", len(r.Datapoints)))
	}
	dp := r.Datapoints[0]
	fields := make(map[string]interface{})
	if dp.Average != nil {
		fields[formatField(*metric.MetricName, "average")] = *dp.Average
	}
	if dp.Maximum != nil {
		fields[formatField(*metric.MetricName, "maximum")] = *dp.Maximum
	}
	if dp.Minimum != nil {
		fields[formatField(*metric.MetricName, "minimum")] = *dp.Minimum
	}
	if dp.Sum != nil {
		fields[formatField(*metric.MetricName, "sum")] = *dp.Sum
	}
	if dp.SampleCount != nil {
		fields[formatField(*metric.MetricName, "sample_count")] = *dp.SampleCount
	}
	tags := getTags(region, metric.Dimensions)
	acc.AddFields(formatMeasurement(*metric.Namespace), fields, tags)
	return nil
}

func getTags(region string, dimensions []*cloudwatch.Dimension) map[string]string {
	tags := make(map[string]string)
	tags["region"] = region
	for _, d := range dimensions {
		tags[snakecase.SnakeCase(*d.Name)] = *d.Value
	}
	return tags
}

func formatField(metricName string, statistic string) string {
	return fmt.Sprintf("%s_%s", snakecase.SnakeCase(metricName), statistic)
}

func formatMeasurement(namespace string) string {
	ns := snakecase.SnakeCase(namespace)
	ns = strings.Replace(ns, "/", "_", -1)
	ns = strings.Replace(ns, "__", "_", -1)
	return fmt.Sprintf("cloudwatch_%s", ns)
}

func init() {
	inputs.Add("cloudwatch", func() telegraf.Input {
		return &Cloudwatch{}
	})
}
