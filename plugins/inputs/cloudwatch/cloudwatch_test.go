package cloudwatch

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/influxdata/telegraf/testutil"
	"testing"
	"time"
)

type ListMetricsMock struct{}

func (lmm *ListMetricsMock) ListMetrics(params *cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error) {
	switch *params.Namespace {
	case "AWS/EC2":
		return &cloudwatch.ListMetricsOutput{
			Metrics: []*cloudwatch.Metric{
				&cloudwatch.Metric{
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("InstanceId"),
							Value: aws.String("i-a6c5c510"),
						},
					},
					MetricName: aws.String("DiskReadBytes"),
					Namespace:  aws.String("AWS/EC2"),
				},
				&cloudwatch.Metric{
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("InstanceId"),
							Value: aws.String("i-10a2c094"),
						},
					},
					MetricName: aws.String("DiskWriteBytes"),
					Namespace:  aws.String("AWS/EC2"),
				},
			},
		}, nil
	case "AWS/DynamoDB":
		return &cloudwatch.ListMetricsOutput{
			Metrics: []*cloudwatch.Metric{
				&cloudwatch.Metric{
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("Operation"),
							Value: aws.String("GetItem"),
						},
						&cloudwatch.Dimension{
							Name:  aws.String("TableName"),
							Value: aws.String("foo"),
						},
					},
					MetricName: aws.String("ThrottledRequests"),
					Namespace:  aws.String("AWS/DynamoDB"),
				},
				&cloudwatch.Metric{
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("Operation"),
							Value: aws.String("PutItem"),
						},
						&cloudwatch.Dimension{
							Name:  aws.String("TableName"),
							Value: aws.String("foo"),
						},
					},
					MetricName: aws.String("ThrottledRequests"),
					Namespace:  aws.String("AWS/DynamoDB"),
				},
			},
		}, nil
	default:
		return nil, errors.New("No such namespace")
	}
}

func TestListMetrics(t *testing.T) {
	var lmm ListMetricsMock

	metrics, err := listMetrics(&lmm, []string{"AWS/EC2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 2 {
		t.Fatal("Expected 2 AWS/EC2 metrics, found ", len(metrics))
	}

	allMetrics, err := listMetrics(&lmm, []string{"AWS/EC2", "AWS/DynamoDB"})
	if err != nil {
		t.Fatal(err)
	}
	if len(allMetrics) != 4 {
		t.Fatal("Expected 4 AWS/EC2 and AWS/DynamoDB metrics, found ", len(allMetrics))
	}
}

type GetMetricStatisticsMock struct{}

func (gmsm *GetMetricStatisticsMock) GetMetricStatistics(input *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
	timestamp := time.Unix(1459201201, 0)
	avg := float64(1.6295248026217786)
	maximum := float64(8.0)
	minimum := float64(0.5)
	sum := float64(10939.0)
	sampleCount := float64(6713.0)
	return &cloudwatch.GetMetricStatisticsOutput{
		Datapoints: []*cloudwatch.Datapoint{
			&cloudwatch.Datapoint{
				Timestamp:   &timestamp,
				Average:     &avg,
				Maximum:     &maximum,
				Minimum:     &minimum,
				Sum:         &sum,
				SampleCount: &sampleCount,
				Unit:        aws.String("Count"),
			},
		},
		Label: aws.String("ConsumedReadCapacityUnits"),
	}, nil
}

type GetMetricStatisticsMockPartial struct{}

func (gmsm *GetMetricStatisticsMockPartial) GetMetricStatistics(input *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
	timestamp := time.Unix(1459201201, 0)
	avg := float64(1.6295248026217786)
	return &cloudwatch.GetMetricStatisticsOutput{
		Datapoints: []*cloudwatch.Datapoint{
			&cloudwatch.Datapoint{
				Timestamp: &timestamp,
				Average:   &avg,
				Unit:      aws.String("Count"),
			},
		},
		Label: aws.String("ConsumedReadCapacityUnits"),
	}, nil
}

func TestGatherMetric(t *testing.T) {
	var svc GetMetricStatisticsMock
	var acc testutil.Accumulator
	metric := &cloudwatch.Metric{
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("TableName"),
				Value: aws.String("foo"),
			},
		},
		MetricName: aws.String("ConsumedReadCapacityUnits"),
		Namespace:  aws.String("AWS/DynamoDB"),
	}
	bogusTime := time.Now()
	err := gatherMetric(&svc, "us-east-1", &acc, metric, bogusTime, bogusTime, 300)
	if err != nil {
		t.Fatal(err)
	}
	acc.AssertContainsTaggedFields(t, "cloudwatch_aws_dynamo_db",
		map[string]interface{}{"consumed_read_capacity_units_average": 1.6295248026217786,
			"consumed_read_capacity_units_maximum":      8.0,
			"consumed_read_capacity_units_minimum":      0.5,
			"consumed_read_capacity_units_sum":          10939.0,
			"consumed_read_capacity_units_sample_count": 6713.0,
		},
		map[string]string{
			"region":     "us-east-1",
			"table_name": "foo",
		})
}

// Test that it still works with only some statistics
func TestGatherMetricPartial(t *testing.T) {
	var svc GetMetricStatisticsMockPartial
	var acc testutil.Accumulator
	metric := &cloudwatch.Metric{
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("TableName"),
				Value: aws.String("foo"),
			},
		},
		MetricName: aws.String("ConsumedReadCapacityUnits"),
		Namespace:  aws.String("AWS/DynamoDB"),
	}
	bogusTime := time.Now()
	err := gatherMetric(&svc, "us-east-1", &acc, metric, bogusTime, bogusTime, 300)
	if err != nil {
		t.Fatal(err)
	}
	acc.AssertContainsTaggedFields(t, "cloudwatch_aws_dynamo_db",
		map[string]interface{}{"consumed_read_capacity_units_average": 1.6295248026217786},
		map[string]string{
			"region":     "us-east-1",
			"table_name": "foo",
		})
}

func TestGetTags(t *testing.T) {
	tags := getTags("us-east-1", []*cloudwatch.Dimension{
		&cloudwatch.Dimension{
			Name:  aws.String("Operation"),
			Value: aws.String("GetItem"),
		},
		&cloudwatch.Dimension{
			Name:  aws.String("TableName"),
			Value: aws.String("foo"),
		},
	})
	if len(tags) != 3 {
		t.Fatal("Expected 3 tags, not ", len(tags))
	}
	if tags["region"] != "us-east-1" {
		t.Fatal("Expected the tag `region` with value `us-east-1`")
	}
	if tags["operation"] != "GetItem" {
		t.Fatal("Expected the tag `operation` with value `GetItem`")
	}
	if tags["table_name"] != "foo" {
		t.Fatal("Expected the tag `table_name` with value `foo`")
	}
}

func TestFormatField(t *testing.T) {
	if s := formatField("DiskWriteBytes", "average"); s != "disk_write_bytes_average" {
		t.Fatal("Field should not be ", s)
	}
}

func TestFormatMeasurement(t *testing.T) {
	if s := formatMeasurement("AWS/EC2"); s != "cloudwatch_aws_ec2" {
		t.Fatal("Measurement should not be ", s)
	}
	if s := formatMeasurement("AWS/DynamoDB"); s != "cloudwatch_aws_dynamo_db" {
		t.Fatal("Measurement should not be ", s)
	}
}
