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

type mockCloudWatchClient struct{}

func (m *mockCloudWatchClient) ListMetrics(params *cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error) {
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

func (m *mockCloudWatchClient) GetMetricStatistics(params *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
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

func TestGather(t *testing.T) {
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
	c.client = &mockCloudWatchClient{}

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
