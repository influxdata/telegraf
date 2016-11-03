package cloudwatch

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

type (
	basicCloudWatchClient struct{}
	regexCloudWatchClient struct{}
)

func (*basicCloudWatchClient) ListMetrics(params *cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error) {
	metric := &cloudwatch.Metric{
		Namespace:  params.Namespace,
		MetricName: aws.String("Latency"),
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("LoadBalancerName"),
				Value: aws.String("p-example"),
			},
		},
	}

	result := &cloudwatch.ListMetricsOutput{
		Metrics: []*cloudwatch.Metric{metric},
	}
	return result, nil
}

func (*basicCloudWatchClient) GetMetricStatistics(params *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
	dataPoint := &cloudwatch.Datapoint{
		Timestamp:   params.EndTime,
		Minimum:     aws.Float64(0.1),
		Maximum:     aws.Float64(0.3),
		Average:     aws.Float64(0.2),
		Sum:         aws.Float64(123),
		SampleCount: aws.Float64(100),
		Unit:        aws.String("Seconds"),
	}
	result := &cloudwatch.GetMetricStatisticsOutput{
		Label:      aws.String("Latency"),
		Datapoints: []*cloudwatch.Datapoint{dataPoint},
	}
	return result, nil
}

func tblMetric(params *cloudwatch.ListMetricsInput, d1 string, d2 string) *cloudwatch.Metric {
	return &cloudwatch.Metric{
		Namespace:  params.Namespace,
		MetricName: aws.String("ConsumedReadCapacityUnits"),
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("TableName"),
				Value: aws.String(d1),
			},
			&cloudwatch.Dimension{
				Name:  aws.String("IndexName"),
				Value: aws.String(d2),
			},
		},
	}
}

func (*regexCloudWatchClient) ListMetrics(params *cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error) {
	metric1 := tblMetric(params, "foo-table1", "ix-foo-t1")
	metric2 := tblMetric(params, "foo-table2", "ix-foo-t2")
	metric3 := tblMetric(params, "bar-table1", "ix-bar-t1")
	metric4 := tblMetric(params, "bar-table2", "ix-bar-t2")

	result := &cloudwatch.ListMetricsOutput{
		Metrics: []*cloudwatch.Metric{metric1, metric2, metric3, metric4},
	}
	return result, nil
}

func (*regexCloudWatchClient) GetMetricStatistics(params *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
	dataPoint := &cloudwatch.Datapoint{
		Timestamp:   params.EndTime,
		Minimum:     aws.Float64(0),
		Maximum:     aws.Float64(10),
		Average:     aws.Float64(4),
		Sum:         aws.Float64(40),
		SampleCount: aws.Float64(10),
		Unit:        aws.String("Units"),
	}
	result := &cloudwatch.GetMetricStatisticsOutput{
		Label:      aws.String("ConsumedReadCapacityUnits"),
		Datapoints: []*cloudwatch.Datapoint{dataPoint},
	}
	return result, nil
}

func TestBasicGather(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}
	c := &CloudWatch{
		Region:    "us-east-1",
		Namespace: "AWS/ELB",
		Delay:     internalDuration,
		Period:    internalDuration,
		RateLimit: 10,
	}

	var acc testutil.Accumulator
	c.client = &basicCloudWatchClient{}

	c.Gather(&acc)

	fields := map[string]interface{}{}
	fields["latency_minimum"] = 0.1
	fields["latency_maximum"] = 0.3
	fields["latency_average"] = 0.2
	fields["latency_sum"] = 123.0
	fields["latency_sample_count"] = 100.0

	tags := map[string]string{}
	tags["unit"] = "seconds"
	tags["region"] = "us-east-1"
	tags["load_balancer_name"] = "p-example"

	assert.True(t, acc.HasMeasurement("cloudwatch_aws_elb"))
	acc.AssertContainsTaggedFields(t, "cloudwatch_aws_elb", fields, tags)

}

func TestSingleDimensionRegexGather(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}
	c := &CloudWatch{
		Region:    "us-east-1",
		Namespace: "AWS/DynamoDB",
		Delay:     internalDuration,
		Period:    internalDuration,
		RateLimit: 10,
		Metrics: []*Metric{
			&Metric{
				MetricNames: []string{"ConsumedReadCapacityUnits"},
				Dimensions: []*Dimension{
					&Dimension{Name: "TableName", Value: "foo*"},
					&Dimension{Name: "IndexName", Value: ""},
				},
			},
		},
	}

	var acc testutil.Accumulator
	acc.SetDebug(true)
	c.client = &regexCloudWatchClient{}

	c.Gather(&acc)

	fields := map[string]interface{}{}
	fields["consumed_read_capacity_units_minimum"] = 0.
	fields["consumed_read_capacity_units_maximum"] = 10.
	fields["consumed_read_capacity_units_average"] = 4.
	fields["consumed_read_capacity_units_sum"] = 40.
	fields["consumed_read_capacity_units_sample_count"] = 10.

	tags := map[string]string{}
	tags["unit"] = "units"
	tags["region"] = "us-east-1"
	tags["table_name"] = "foo-table1"
	tags["index_name"] = "ix-foo-t1"

	assert.True(t, acc.HasMeasurement("cloudwatch_aws_dynamo_db"))
	acc.AssertContainsTaggedFields(t, "cloudwatch_aws_dynamo_db", fields, tags)

	tags["table_name"] = "foo-table2"
	tags["index_name"] = "ix-foo-t2"
	acc.AssertContainsTaggedFields(t, "cloudwatch_aws_dynamo_db", fields, tags)
}

func TestMultiDimensionRegexGather(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}
	c := &CloudWatch{
		Region:    "us-east-1",
		Namespace: "AWS/DynamoDB",
		Delay:     internalDuration,
		Period:    internalDuration,
		RateLimit: 10,
		Metrics: []*Metric{
			&Metric{
				MetricNames: []string{"ConsumedReadCapacityUnits"},
				Dimensions: []*Dimension{
					&Dimension{Name: "TableName", Value: "foo*"},
					&Dimension{Name: "IndexName", Value: "*t2"},
				},
			},
		},
	}

	var acc testutil.Accumulator
	c.client = &regexCloudWatchClient{}

	c.Gather(&acc)

	fields := map[string]interface{}{}
	fields["consumed_read_capacity_units_minimum"] = 0.
	fields["consumed_read_capacity_units_maximum"] = 10.
	fields["consumed_read_capacity_units_average"] = 4.
	fields["consumed_read_capacity_units_sum"] = 40.
	fields["consumed_read_capacity_units_sample_count"] = 10.

	tags := map[string]string{}
	tags["unit"] = "units"
	tags["region"] = "us-east-1"
	tags["table_name"] = "foo-table2"
	tags["index_name"] = "ix-foo-t2"

	assert.True(t, acc.HasMeasurement("cloudwatch_aws_dynamo_db"))
	acc.AssertContainsTaggedFields(t, "cloudwatch_aws_dynamo_db", fields, tags)
	assert.Equal(t, 1, acc.CountTaggedMeasurements("cloudwatch_aws_dynamo_db", tags))
}

func TestDimensionGather(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}
	c := &CloudWatch{
		Region:    "us-east-1",
		Namespace: "AWS/DynamoDB",
		Delay:     internalDuration,
		Period:    internalDuration,
		RateLimit: 10,
		Metrics: []*Metric{
			&Metric{
				MetricNames: []string{"ConsumedReadCapacityUnits"},
				Dimensions: []*Dimension{
					&Dimension{Name: "TableName", Value: "foo-table1"},
					&Dimension{Name: "IndexName", Value: "ix-foo-t1"},
				},
			},
		},
	}

	var acc testutil.Accumulator
	c.client = &regexCloudWatchClient{}

	c.Gather(&acc)

	fields := map[string]interface{}{}
	fields["consumed_read_capacity_units_minimum"] = 0.
	fields["consumed_read_capacity_units_maximum"] = 10.
	fields["consumed_read_capacity_units_average"] = 4.
	fields["consumed_read_capacity_units_sum"] = 40.
	fields["consumed_read_capacity_units_sample_count"] = 10.

	tags := map[string]string{}
	tags["unit"] = "units"
	tags["region"] = "us-east-1"
	tags["table_name"] = "foo-table1"
	tags["index_name"] = "ix-foo-t1"

	assert.True(t, acc.HasMeasurement("cloudwatch_aws_dynamo_db"))
	acc.AssertContainsTaggedFields(t, "cloudwatch_aws_dynamo_db", fields, tags)
	assert.Equal(t, 1, acc.CountTaggedMeasurements("cloudwatch_aws_dynamo_db", tags))
}

func TestGenerateStatisticsInputParams(t *testing.T) {
	d := &cloudwatch.Dimension{
		Name:  aws.String("LoadBalancerName"),
		Value: aws.String("p-example"),
	}

	m := &cloudwatch.Metric{
		MetricName: aws.String("Latency"),
		Dimensions: []*cloudwatch.Dimension{d},
	}

	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}

	c := &CloudWatch{
		Namespace: "AWS/ELB",
		Delay:     internalDuration,
		Period:    internalDuration,
	}

	c.initializeCloudWatch()

	now := time.Now()

	params := c.getStatisticsInput(m, now)

	assert.EqualValues(t, *params.EndTime, now.Add(-c.Delay.Duration))
	assert.EqualValues(t, *params.StartTime, now.Add(-c.Period.Duration).Add(-c.Delay.Duration))
	assert.Len(t, params.Dimensions, 1)
	assert.Len(t, params.Statistics, 5)
	assert.EqualValues(t, *params.Period, 60)
}

func TestMetricsCacheTimeout(t *testing.T) {
	ttl, _ := time.ParseDuration("5ms")
	cache := &MetricCache{
		Metrics: []*cloudwatch.Metric{},
		Fetched: time.Now(),
		TTL:     ttl,
	}

	assert.True(t, cache.IsValid())
	time.Sleep(ttl)
	assert.False(t, cache.IsValid())
}
