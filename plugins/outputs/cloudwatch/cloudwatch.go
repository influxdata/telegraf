package cloudwatch

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/plugins/outputs"

	"github.com/meirf/gopart"
)

type CloudWatchOutput struct {
	Region    string // AWS Region
	Namespace string // CloudWatch Metrics Namespace
	svc       *cloudwatch.CloudWatch
}

var sampleConfig = `
  # Amazon REGION
  region = 'us-east-1'

  # Namespace for the CloudWatch MetricDatums
  namespace = 'InfluxData/Telegraf'
`

func (c *CloudWatchOutput) SampleConfig() string {
	return sampleConfig
}

func (c *CloudWatchOutput) Description() string {
	return "Configuration for AWS CloudWatch output."
}

func (c *CloudWatchOutput) Connect() error {
	Config := &aws.Config{
		Region: aws.String(c.Region),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
			}),
	}

	svc := cloudwatch.New(session.New(Config))

	params := &cloudwatch.ListMetricsInput{
		Namespace: aws.String(c.Namespace),
	}

	_, err := svc.ListMetrics(params) // Try a read-only call to test connection.

	if err != nil {
		log.Printf("cloudwatch: Error in ListMetrics API call : %+v \n", err.Error())
	}

	c.svc = svc

	return err
}

func (c *CloudWatchOutput) Close() error {
	return errors.New("Error")
}

func (c *CloudWatchOutput) Write(points []*client.Point) error {
	for _, pt := range points {
		err := c.WriteSinglePoint(pt)
		if err != nil {
			return err
		}
	}

	return nil
}

// Write data for a single point. A point can have many fields and one field
// is equal to one MetricDatum. There is a limit on how many MetricDatums a
// request can have so we process one Point at a time.
func (c *CloudWatchOutput) WriteSinglePoint(point *client.Point) error {
	datums := buildMetricDatum(point)

	const maxDatumsPerCall = 20 // PutMetricData only supports up to 20 data points per call

	for idxRange := range gopart.Partition(len(datums), maxDatumsPerCall) {
		err := c.WriteToCloudWatch(datums[idxRange.Low:idxRange.High])

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CloudWatchOutput) WriteToCloudWatch(datums []*cloudwatch.MetricDatum) error {
	params := &cloudwatch.PutMetricDataInput{
		MetricData: datums,
		Namespace:  aws.String(c.Namespace),
	}

	_, err := c.svc.PutMetricData(params)

	if err != nil {
		log.Printf("CloudWatch: Unable to write to CloudWatch : %+v \n", err.Error())
	}

	return err
}

// Make a MetricDatum for each field in a Point. Only fields with values that can be
// converted to float64 are supported. Non-supported fields are skipped.
func buildMetricDatum(point *client.Point) []*cloudwatch.MetricDatum {
	datums := make([]*cloudwatch.MetricDatum, len(point.Fields()))
	i := 0

	var value float64

	for k, v := range point.Fields() {
		switch t := v.(type) {
		case int:
			value = float64(t)
		case int32:
			value = float64(t)
		case int64:
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
			datums = datums[:len(datums)-1]
			continue
		}

		datums[i] = &cloudwatch.MetricDatum{
			MetricName: aws.String(strings.Join([]string{point.Name(), k}, "_")),
			Value:      aws.Float64(value),
			Dimensions: buildDimensions(point.Tags()),
			Timestamp:  aws.Time(point.Time()),
		}

		i += 1
	}

	return datums
}

// Make a list of Dimensions by using a Point's tags.
func buildDimensions(ptTags map[string]string) []*cloudwatch.Dimension {

	dimensions := make([]*cloudwatch.Dimension, len(ptTags))
	i := 0

	for k, v := range ptTags {
		dimensions[i] = &cloudwatch.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		}

		i += 1
	}

	return dimensions
}

func init() {
	outputs.Add("cloudwatch", func() outputs.Output {
		return &CloudWatchOutput{}
	})
}
