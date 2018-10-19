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

type mockGatherCloudWatchClient struct{}

func (m *mockGatherCloudWatchClient) ListMetrics(params *cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error) {
	metric := &cloudwatch.Metric{
		Namespace:  params.Namespace,
		MetricName: aws.String("Latency"),
		Dimensions: []*cloudwatch.Dimension{
			{
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

func (m *mockGatherCloudWatchClient) GetMetricStatistics(params *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
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
		RateLimit: 200,
	}

	var acc testutil.Accumulator
	c.client = &mockGatherCloudWatchClient{}

	acc.GatherError(c.Gather)

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

type mockSelectMetricsCloudWatchClient struct{}

func (m *mockSelectMetricsCloudWatchClient) ListMetrics(params *cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error) {
	metrics := []*cloudwatch.Metric{}
	// 4 metrics are available
	metricNames := []string{"Latency", "RequestCount", "HealthyHostCount", "UnHealthyHostCount"}
	// for 3 ELBs
	loadBalancers := []string{"lb-1", "lb-2", "lb-3"}
	// in 2 AZs
	availabilityZones := []string{"us-east-1a", "us-east-1b"}
	for _, m := range metricNames {
		for _, lb := range loadBalancers {
			// For each metric/ELB pair, we get an aggregate value across all AZs.
			metrics = append(metrics, &cloudwatch.Metric{
				Namespace:  aws.String("AWS/ELB"),
				MetricName: aws.String(m),
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  aws.String("LoadBalancerName"),
						Value: aws.String(lb),
					},
				},
			})
			for _, az := range availabilityZones {
				// We get a metric for each metric/ELB/AZ triplet.
				metrics = append(metrics, &cloudwatch.Metric{
					Namespace:  aws.String("AWS/ELB"),
					MetricName: aws.String(m),
					Dimensions: []*cloudwatch.Dimension{
						{
							Name:  aws.String("LoadBalancerName"),
							Value: aws.String(lb),
						},
						{
							Name:  aws.String("AvailabilityZone"),
							Value: aws.String(az),
						},
					},
				})
			}
		}
	}

	result := &cloudwatch.ListMetricsOutput{
		Metrics: metrics,
	}
	return result, nil
}

func (m *mockSelectMetricsCloudWatchClient) GetMetricStatistics(params *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
	return nil, nil
}

func TestSelectMetrics(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}
	c := &CloudWatch{
		Region:    "us-east-1",
		Namespace: "AWS/ELB",
		Delay:     internalDuration,
		Period:    internalDuration,
		RateLimit: 200,
		Metrics: []*Metric{
			{
				MetricNames: []string{"Latency", "RequestCount"},
				Dimensions: []*Dimension{
					{
						Name:  "LoadBalancerName",
						Value: "*",
					},
					{
						Name:  "AvailabilityZone",
						Value: "*",
					},
				},
			},
		},
	}
	c.client = &mockSelectMetricsCloudWatchClient{}
	metrics, err := SelectMetrics(c)
	// We've asked for 2 (out of 4) metrics, over all 3 load balancers in all 2
	// AZs. We should get 12 metrics.
	assert.Equal(t, 12, len(metrics))
	assert.Nil(t, err)
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

	c.updateWindow(now)

	params := c.getStatisticsInput(m)

	assert.EqualValues(t, *params.EndTime, now.Add(-c.Delay.Duration))
	assert.EqualValues(t, *params.StartTime, now.Add(-c.Period.Duration).Add(-c.Delay.Duration))
	assert.Len(t, params.Dimensions, 1)
	assert.Len(t, params.Statistics, 5)
	assert.EqualValues(t, *params.Period, 60)
}

func TestMetricsCacheTimeout(t *testing.T) {
	cache := &MetricCache{
		Metrics: []*cloudwatch.Metric{},
		Fetched: time.Now(),
		TTL:     time.Minute,
	}

	assert.True(t, cache.IsValid())
	cache.Fetched = time.Now().Add(-time.Minute)
	assert.False(t, cache.IsValid())
}

func TestUpdateWindow(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := internal.Duration{
		Duration: duration,
	}

	c := &CloudWatch{
		Namespace: "AWS/ELB",
		Delay:     internalDuration,
		Period:    internalDuration,
	}

	now := time.Now()

	assert.True(t, c.windowEnd.IsZero())
	assert.True(t, c.windowStart.IsZero())

	c.updateWindow(now)

	newStartTime := c.windowEnd

	// initial window just has a single period
	assert.EqualValues(t, c.windowEnd, now.Add(-c.Delay.Duration))
	assert.EqualValues(t, c.windowStart, now.Add(-c.Delay.Duration).Add(-c.Period.Duration))

	now = time.Now()
	c.updateWindow(now)

	// subsequent window uses previous end time as start time
	assert.EqualValues(t, c.windowEnd, now.Add(-c.Delay.Duration))
	assert.EqualValues(t, c.windowStart, newStartTime)
}
