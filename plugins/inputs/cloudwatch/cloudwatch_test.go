package cloudwatch

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cwClient "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/testutil"
)

type mockGatherCloudWatchClient struct{}

func (m *mockGatherCloudWatchClient) ListMetrics(_ context.Context, params *cwClient.ListMetricsInput, _ ...func(*cwClient.Options)) (*cwClient.ListMetricsOutput, error) {
	return &cwClient.ListMetricsOutput{
		Metrics: []types.Metric{
			{
				Namespace:  params.Namespace,
				MetricName: aws.String("Latency"),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("LoadBalancerName"),
						Value: aws.String("p-example"),
					},
				},
			},
		},
	}, nil
}

func (m *mockGatherCloudWatchClient) GetMetricData(_ context.Context, params *cwClient.GetMetricDataInput, _ ...func(*cwClient.Options)) (*cwClient.GetMetricDataOutput, error) {
	return &cwClient.GetMetricDataOutput{
		MetricDataResults: []types.MetricDataResult{
			{
				Id:         aws.String("minimum_0_0"),
				Label:      aws.String("latency_minimum"),
				StatusCode: types.StatusCodeComplete,
				Timestamps: []time.Time{
					*params.EndTime,
				},
				Values: []float64{0.1},
			},
			{
				Id:         aws.String("maximum_0_0"),
				Label:      aws.String("latency_maximum"),
				StatusCode: types.StatusCodeComplete,
				Timestamps: []time.Time{
					*params.EndTime,
				},
				Values: []float64{0.3},
			},
			{
				Id:         aws.String("average_0_0"),
				Label:      aws.String("latency_average"),
				StatusCode: types.StatusCodeComplete,
				Timestamps: []time.Time{
					*params.EndTime,
				},
				Values: []float64{0.2},
			},
			{
				Id:         aws.String("sum_0_0"),
				Label:      aws.String("latency_sum"),
				StatusCode: types.StatusCodeComplete,
				Timestamps: []time.Time{
					*params.EndTime,
				},
				Values: []float64{123},
			},
			{
				Id:         aws.String("sample_count_0_0"),
				Label:      aws.String("latency_sample_count"),
				StatusCode: types.StatusCodeComplete,
				Timestamps: []time.Time{
					*params.EndTime,
				},
				Values: []float64{100},
			},
		},
	}, nil
}

func TestSnakeCase(t *testing.T) {
	require.Equal(t, "cluster_name", snakeCase("Cluster Name"))
	require.Equal(t, "broker_id", snakeCase("Broker ID"))
}

func TestGather(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := config.Duration(duration)
	c := &CloudWatch{
		CredentialConfig: internalaws.CredentialConfig{
			Region: "us-east-1",
		},
		Namespace: "AWS/ELB",
		Delay:     internalDuration,
		Period:    internalDuration,
		RateLimit: 200,
	}

	var acc testutil.Accumulator

	require.NoError(t, c.Init())
	c.client = &mockGatherCloudWatchClient{}
	require.NoError(t, acc.GatherError(c.Gather))

	fields := map[string]interface{}{}
	fields["latency_minimum"] = 0.1
	fields["latency_maximum"] = 0.3
	fields["latency_average"] = 0.2
	fields["latency_sum"] = 123.0
	fields["latency_sample_count"] = 100.0

	tags := map[string]string{}
	tags["region"] = "us-east-1"
	tags["load_balancer_name"] = "p-example"

	require.True(t, acc.HasMeasurement("cloudwatch_aws_elb"))
	acc.AssertContainsTaggedFields(t, "cloudwatch_aws_elb", fields, tags)
}

func TestGather_MultipleNamespaces(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := config.Duration(duration)
	c := &CloudWatch{
		Namespaces: []string{"AWS/ELB", "AWS/EC2"},
		Delay:      internalDuration,
		Period:     internalDuration,
		RateLimit:  200,
	}

	var acc testutil.Accumulator

	require.NoError(t, c.Init())
	c.client = &mockGatherCloudWatchClient{}
	require.NoError(t, acc.GatherError(c.Gather))

	require.True(t, acc.HasMeasurement("cloudwatch_aws_elb"))
	require.True(t, acc.HasMeasurement("cloudwatch_aws_ec2"))
}

type mockSelectMetricsCloudWatchClient struct{}

func (m *mockSelectMetricsCloudWatchClient) ListMetrics(_ context.Context, params *cwClient.ListMetricsInput, _ ...func(*cwClient.Options)) (*cwClient.ListMetricsOutput, error) {
	metrics := []types.Metric{}
	// 4 metrics are available
	metricNames := []string{"Latency", "RequestCount", "HealthyHostCount", "UnHealthyHostCount"}
	// for 3 ELBs
	loadBalancers := []string{"lb-1", "lb-2", "lb-3"}
	// in 2 AZs
	availabilityZones := []string{"us-east-1a", "us-east-1b"}
	for _, m := range metricNames {
		for _, lb := range loadBalancers {
			// For each metric/ELB pair, we get an aggregate value across all AZs.
			metrics = append(metrics, types.Metric{
				Namespace:  aws.String("AWS/ELB"),
				MetricName: aws.String(m),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("LoadBalancerName"),
						Value: aws.String(lb),
					},
				},
			})
			for _, az := range availabilityZones {
				// We get a metric for each metric/ELB/AZ triplet.
				metrics = append(metrics, types.Metric{
					Namespace:  aws.String("AWS/ELB"),
					MetricName: aws.String(m),
					Dimensions: []types.Dimension{
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

	result := &cwClient.ListMetricsOutput{
		Metrics: metrics,
	}
	return result, nil
}

func (m *mockSelectMetricsCloudWatchClient) GetMetricData(_ context.Context, params *cwClient.GetMetricDataInput, _ ...func(*cwClient.Options)) (*cwClient.GetMetricDataOutput, error) {
	return nil, nil
}

func TestSelectMetrics(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := config.Duration(duration)
	c := &CloudWatch{
		CredentialConfig: internalaws.CredentialConfig{
			Region: "us-east-1",
		},
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
						Value: "lb*",
					},
					{
						Name:  "AvailabilityZone",
						Value: "us-east*",
					},
				},
			},
		},
	}
	require.NoError(t, c.Init())
	c.client = &mockSelectMetricsCloudWatchClient{}
	filtered, err := getFilteredMetrics(c)
	// We've asked for 2 (out of 4) metrics, over all 3 load balancers in all 2
	// AZs. We should get 12 metrics.
	require.Equal(t, 12, len(filtered[0].metrics))
	require.NoError(t, err)
}

func TestGenerateStatisticsInputParams(t *testing.T) {
	d := types.Dimension{
		Name:  aws.String("LoadBalancerName"),
		Value: aws.String("p-example"),
	}

	namespace := "AWS/ELB"
	m := types.Metric{
		MetricName: aws.String("Latency"),
		Dimensions: []types.Dimension{d},
		Namespace:  aws.String(namespace),
	}

	duration, _ := time.ParseDuration("1m")
	internalDuration := config.Duration(duration)

	c := &CloudWatch{
		Namespaces: []string{namespace},
		Delay:      internalDuration,
		Period:     internalDuration,
	}

	require.NoError(t, c.initializeCloudWatch())

	now := time.Now()

	c.updateWindow(now)

	statFilter, _ := filter.NewIncludeExcludeFilter(nil, nil)
	queries := c.getDataQueries([]filteredMetric{{metrics: []types.Metric{m}, statFilter: statFilter}})
	params := c.getDataInputs(queries[namespace])

	require.EqualValues(t, *params.EndTime, now.Add(-time.Duration(c.Delay)))
	require.EqualValues(t, *params.StartTime, now.Add(-time.Duration(c.Period)).Add(-time.Duration(c.Delay)))
	require.Len(t, params.MetricDataQueries, 5)
	require.Len(t, params.MetricDataQueries[0].MetricStat.Metric.Dimensions, 1)
	require.EqualValues(t, *params.MetricDataQueries[0].MetricStat.Period, 60)
}

func TestGenerateStatisticsInputParamsFiltered(t *testing.T) {
	d := types.Dimension{
		Name:  aws.String("LoadBalancerName"),
		Value: aws.String("p-example"),
	}

	namespace := "AWS/ELB"
	m := types.Metric{
		MetricName: aws.String("Latency"),
		Dimensions: []types.Dimension{d},
		Namespace:  aws.String(namespace),
	}

	duration, _ := time.ParseDuration("1m")
	internalDuration := config.Duration(duration)

	c := &CloudWatch{
		Namespaces: []string{namespace},
		Delay:      internalDuration,
		Period:     internalDuration,
	}

	require.NoError(t, c.initializeCloudWatch())

	now := time.Now()

	c.updateWindow(now)

	statFilter, _ := filter.NewIncludeExcludeFilter([]string{"average", "sample_count"}, nil)
	queries := c.getDataQueries([]filteredMetric{{metrics: []types.Metric{m}, statFilter: statFilter}})
	params := c.getDataInputs(queries[namespace])

	require.EqualValues(t, *params.EndTime, now.Add(-time.Duration(c.Delay)))
	require.EqualValues(t, *params.StartTime, now.Add(-time.Duration(c.Period)).Add(-time.Duration(c.Delay)))
	require.Len(t, params.MetricDataQueries, 2)
	require.Len(t, params.MetricDataQueries[0].MetricStat.Metric.Dimensions, 1)
	require.EqualValues(t, *params.MetricDataQueries[0].MetricStat.Period, 60)
}

func TestMetricsCacheTimeout(t *testing.T) {
	cache := &metricCache{
		metrics: []filteredMetric{},
		built:   time.Now(),
		ttl:     time.Minute,
	}

	require.True(t, cache.isValid())
	cache.built = time.Now().Add(-time.Minute)
	require.False(t, cache.isValid())
}

func TestUpdateWindow(t *testing.T) {
	duration, _ := time.ParseDuration("1m")
	internalDuration := config.Duration(duration)

	c := &CloudWatch{
		Namespace: "AWS/ELB",
		Delay:     internalDuration,
		Period:    internalDuration,
	}

	now := time.Now()

	require.True(t, c.windowEnd.IsZero())
	require.True(t, c.windowStart.IsZero())

	c.updateWindow(now)

	newStartTime := c.windowEnd

	// initial window just has a single period
	require.EqualValues(t, c.windowEnd, now.Add(-time.Duration(c.Delay)))
	require.EqualValues(t, c.windowStart, now.Add(-time.Duration(c.Delay)).Add(-time.Duration(c.Period)))

	now = time.Now()
	c.updateWindow(now)

	// subsequent window uses previous end time as start time
	require.EqualValues(t, c.windowEnd, now.Add(-time.Duration(c.Delay)))
	require.EqualValues(t, c.windowStart, newStartTime)
}

func TestProxyFunction(t *testing.T) {
	c := &CloudWatch{
		HTTPProxy: proxy.HTTPProxy{HTTPProxyURL: "http://www.penguins.com"},
	}

	proxyFunction, err := c.HTTPProxy.Proxy()
	require.NoError(t, err)

	proxyResult, err := proxyFunction(&http.Request{})
	require.NoError(t, err)
	require.Equal(t, "www.penguins.com", proxyResult.Host)
}

func TestCombineNamespaces(t *testing.T) {
	c := &CloudWatch{Namespace: "AWS/ELB", Namespaces: []string{"AWS/EC2", "AWS/Billing"}}

	require.NoError(t, c.Init())
	require.Equal(t, []string{"AWS/EC2", "AWS/Billing", "AWS/ELB"}, c.Namespaces)
}
