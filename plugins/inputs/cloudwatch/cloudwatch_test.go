package cloudwatch

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGatherCloudWatchClient struct{}

func (m *mockGatherCloudWatchClient) ListMetrics(params *cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error) {
	return &cloudwatch.ListMetricsOutput{
		Metrics: []*cloudwatch.Metric{
			{
				Namespace:  params.Namespace,
				MetricName: aws.String("Latency"),
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  aws.String("LoadBalancerName"),
						Value: aws.String("p-example"),
					},
				},
			},
		},
	}, nil
}

func (m *mockGatherCloudWatchClient) GetMetricData(params *cloudwatch.GetMetricDataInput) (*cloudwatch.GetMetricDataOutput, error) {
	return &cloudwatch.GetMetricDataOutput{
		MetricDataResults: []*cloudwatch.MetricDataResult{
			{
				Id:         aws.String("minimum_latency_p_example"),
				Label:      aws.String("latency_minimum"),
				StatusCode: aws.String("completed"),
				Timestamps: []*time.Time{
					params.EndTime,
				},
				Values: []*float64{
					aws.Float64(0.1),
				},
			},
			{
				Id:         aws.String("maximum_latency_p_example"),
				Label:      aws.String("latency_maximum"),
				StatusCode: aws.String("completed"),
				Timestamps: []*time.Time{
					params.EndTime,
				},
				Values: []*float64{
					aws.Float64(0.3),
				},
			},
			{
				Id:         aws.String("average_latency_p_example"),
				Label:      aws.String("latency_average"),
				StatusCode: aws.String("completed"),
				Timestamps: []*time.Time{
					params.EndTime,
				},
				Values: []*float64{
					aws.Float64(0.2),
				},
			},
			{
				Id:         aws.String("sum_latency_p_example"),
				Label:      aws.String("latency_sum"),
				StatusCode: aws.String("completed"),
				Timestamps: []*time.Time{
					params.EndTime,
				},
				Values: []*float64{
					aws.Float64(123),
				},
			},
			{
				Id:         aws.String("sample_count_latency_p_example"),
				Label:      aws.String("latency_sample_count"),
				StatusCode: aws.String("completed"),
				Timestamps: []*time.Time{
					params.EndTime,
				},
				Values: []*float64{
					aws.Float64(100),
				},
			},
		},
	}, nil
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

	assert.NoError(t, acc.GatherError(c.Gather))

	fields := map[string]interface{}{}
	fields["latency_minimum"] = 0.1
	fields["latency_maximum"] = 0.3
	fields["latency_average"] = 0.2
	fields["latency_sum"] = 123.0
	fields["latency_sample_count"] = 100.0

	tags := map[string]string{}
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

func (m *mockSelectMetricsCloudWatchClient) GetMetricData(params *cloudwatch.GetMetricDataInput) (*cloudwatch.GetMetricDataOutput, error) {
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
	filtered, err := SelectMetrics(c)
	// We've asked for 2 (out of 4) metrics, over all 3 load balancers in all 2
	// AZs. We should get 12 metrics.
	assert.Equal(t, 12, len(aggregateFiltered(filtered)))
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

	statFilter, _ := filter.NewIncludeExcludeFilter(nil, nil)
	queries, _ := c.getDataQueries([]filteredMetric{{metrics: []*cloudwatch.Metric{m}, statFilter: statFilter}})
	params := c.getDataInputs(queries)

	assert.EqualValues(t, *params.EndTime, now.Add(-c.Delay.Duration))
	assert.EqualValues(t, *params.StartTime, now.Add(-c.Period.Duration).Add(-c.Delay.Duration))
	require.Len(t, params.MetricDataQueries, 5)
	assert.Len(t, params.MetricDataQueries[0].MetricStat.Metric.Dimensions, 1)
	assert.EqualValues(t, *params.MetricDataQueries[0].MetricStat.Period, 60)
}

func TestGenerateStatisticsInputParamsFiltered(t *testing.T) {
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

	statFilter, _ := filter.NewIncludeExcludeFilter([]string{"average", "sample_count"}, nil)
	queries, _ := c.getDataQueries([]filteredMetric{{metrics: []*cloudwatch.Metric{m}, statFilter: statFilter}})
	params := c.getDataInputs(queries)

	assert.EqualValues(t, *params.EndTime, now.Add(-c.Delay.Duration))
	assert.EqualValues(t, *params.StartTime, now.Add(-c.Period.Duration).Add(-c.Delay.Duration))
	require.Len(t, params.MetricDataQueries, 2)
	assert.Len(t, params.MetricDataQueries[0].MetricStat.Metric.Dimensions, 1)
	assert.EqualValues(t, *params.MetricDataQueries[0].MetricStat.Period, 60)
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

var testStrings = []string{
	"average_integration_latency_snake_case1/snake_case2/{ver}/snake_case3/snake_case4/snake_case5",
	"average_request_count_targetgroup/snake_case1/snake_case2/snake_case3/snake_case4",
	"average_returned_records_count_snake_case_2018_02_05t19:05:07_516_get_records",
	"average_cluster_status.snake_case1",
	"average_throttles_snake_case1:snake_case2",
	"average_processed_bytes_net/snake_case1/snake_case2",
	"average_execution_time_arn:aws:states:snake_case1:01234567890:snake_case2:snake_case3",
}

var serviceIDRegex = regexp.MustCompile("[^a-zA-Z0-9_]+")

func BenchmarkRegex(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, name := range testStrings {
			serviceIDRegex.ReplaceAllString(name, "_")
		}
	}
}

func TestEquals(t *testing.T) {
	regexSlice := make([]string, len(testStrings))
	for i, name := range testStrings {
		regexSlice[i] = serviceIDRegex.ReplaceAllString(name, "_")
	}

	stringSlice := make([]string, len(testStrings))
	for i, name := range testStrings {
		stringSlice[i] = validID.ReplaceAllString(name, "_")
	}

	if len(regexSlice) != len(stringSlice) {
		t.Errorf("Lengths do not match")
		t.FailNow()
	}
	for i := range regexSlice {
		if regexSlice[i] != stringSlice[i] {
			t.Errorf("Expected '%s'; Got '%s'", regexSlice[i], stringSlice[i])
			t.FailNow()
		}
	}
}

var replacer = strings.NewReplacer("/", "_", "-", "_", ".", "_", " ", "_", ":", "_", "{", "", "}", "", "%", "")

func BenchmarkStrings(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, name := range testStrings {
			replacer.Replace(name)
		}
	}
}
