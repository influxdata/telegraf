package cloudwatch

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type (
	CloudWatch struct {
		Region      string `toml:"region"`
		AccessKey   string `toml:"access_key"`
		SecretKey   string `toml:"secret_key"`
		RoleARN     string `toml:"role_arn"`
		Profile     string `toml:"profile"`
		Filename    string `toml:"shared_credential_file"`
		Token       string `toml:"token"`
		EndpointURL string `toml:"endpoint_url"`

		Period      internal.Duration `toml:"period"`
		Delay       internal.Duration `toml:"delay"`
		Namespace   string            `toml:"namespace"`
		Metrics     []*Metric         `toml:"metrics"`
		CacheTTL    internal.Duration `toml:"cache_ttl"`
		RateLimit   int               `toml:"ratelimit"`
		client      cloudwatchClient
		metricCache *MetricCache
	}

	Metric struct {
		MetricNames []string     `toml:"names"`
		Dimensions  []*Dimension `toml:"dimensions"`
	}

	Dimension struct {
		Name  string `toml:"name"`
		Value string `toml:"value"`
	}

	MetricCache struct {
		TTL     time.Duration
		Fetched time.Time
		Metrics []*cloudwatch.Metric
	}

	cloudwatchClient interface {
		ListMetrics(*cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error)
		GetMetricStatistics(*cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error)
	}
)

func (c *CloudWatch) SampleConfig() string {
	return `
  ## Amazon Region
  region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #profile = ""
  #shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  # The minimum period for Cloudwatch metrics is 1 minute (60s). However not all
  # metrics are made available to the 1 minute period. Some are collected at
  # 3 minute, 5 minute, or larger intervals. See https://aws.amazon.com/cloudwatch/faqs/#monitoring.
  # Note that if a period is configured that is smaller than the minimum for a
  # particular metric, that metric will not be returned by the Cloudwatch API
  # and will not be collected by Telegraf.
  #
  ## Requested CloudWatch aggregation Period (required - must be a multiple of 60s)
  period = "5m"

  ## Collection Delay (required - must account for metrics availability via CloudWatch API)
  delay = "5m"

  ## Recommended: use metric 'interval' that is a multiple of 'period' to avoid
  ## gaps or overlap in pulled data
  interval = "5m"

  ## Configure the TTL for the internal cache of metrics.
  ## Defaults to 1 hr if not specified
  #cache_ttl = "10m"

  ## Metric Statistic Namespace (required)
  namespace = "AWS/ELB"

  ## Maximum requests per second. Note that the global default AWS rate limit is
  ## 400 reqs/sec, so if you define multiple namespaces, these should add up to a
  ## maximum of 400. Optional - default value is 200.
  ## See http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html
  ratelimit = 200

  ## Metrics to Pull (optional)
  ## Defaults to all Metrics in Namespace if nothing is provided
  ## Refreshes Namespace available metrics every 1h
  #[[inputs.cloudwatch.metrics]]
  #  names = ["Latency", "RequestCount"]
  #
  #  ## Dimension filters for Metric.  These are optional however all dimensions
  #  ## defined for the metric names must be specified in order to retrieve
  #  ## the metric statistics.
  #  [[inputs.cloudwatch.metrics.dimensions]]
  #    name = "LoadBalancerName"
  #    value = "p-example"
`
}

func (c *CloudWatch) Description() string {
	return "Pull Metric Statistics from Amazon CloudWatch"
}

func SelectMetrics(c *CloudWatch) ([]*cloudwatch.Metric, error) {
	var metrics []*cloudwatch.Metric

	// check for provided metric filter
	if c.Metrics != nil {
		metrics = []*cloudwatch.Metric{}
		for _, m := range c.Metrics {
			if !hasWilcard(m.Dimensions) {
				dimensions := make([]*cloudwatch.Dimension, len(m.Dimensions))
				for k, d := range m.Dimensions {
					dimensions[k] = &cloudwatch.Dimension{
						Name:  aws.String(d.Name),
						Value: aws.String(d.Value),
					}
				}
				for _, name := range m.MetricNames {
					metrics = append(metrics, &cloudwatch.Metric{
						Namespace:  aws.String(c.Namespace),
						MetricName: aws.String(name),
						Dimensions: dimensions,
					})
				}
			} else {
				allMetrics, err := c.fetchNamespaceMetrics()
				if err != nil {
					return nil, err
				}
				for _, name := range m.MetricNames {
					for _, metric := range allMetrics {
						if isSelected(name, metric, m.Dimensions) {
							metrics = append(metrics, &cloudwatch.Metric{
								Namespace:  aws.String(c.Namespace),
								MetricName: aws.String(name),
								Dimensions: metric.Dimensions,
							})
						}
					}
				}
			}
		}
	} else {
		var err error
		metrics, err = c.fetchNamespaceMetrics()
		if err != nil {
			return nil, err
		}
	}
	return metrics, nil
}

func (c *CloudWatch) Gather(acc telegraf.Accumulator) error {
	if c.client == nil {
		c.initializeCloudWatch()
	}

	metrics, err := SelectMetrics(c)
	if err != nil {
		return err
	}

	now := time.Now()

	// limit concurrency or we can easily exhaust user connection limit
	// see cloudwatch API request limits:
	// http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_limits.html
	lmtr := limiter.NewRateLimiter(c.RateLimit, time.Second)
	defer lmtr.Stop()
	var wg sync.WaitGroup
	wg.Add(len(metrics))
	for _, m := range metrics {
		<-lmtr.C
		go func(inm *cloudwatch.Metric) {
			defer wg.Done()
			acc.AddError(c.gatherMetric(acc, inm, now))
		}(m)
	}
	wg.Wait()

	return nil
}

func init() {
	inputs.Add("cloudwatch", func() telegraf.Input {
		ttl, _ := time.ParseDuration("1hr")
		return &CloudWatch{
			CacheTTL:  internal.Duration{Duration: ttl},
			RateLimit: 200,
		}
	})
}

/*
 * Initialize CloudWatch client
 */
func (c *CloudWatch) initializeCloudWatch() error {
	credentialConfig := &internalaws.CredentialConfig{
		Region:      c.Region,
		AccessKey:   c.AccessKey,
		SecretKey:   c.SecretKey,
		RoleARN:     c.RoleARN,
		Profile:     c.Profile,
		Filename:    c.Filename,
		Token:       c.Token,
		EndpointURL: c.EndpointURL,
	}
	configProvider := credentialConfig.Credentials()

	c.client = cloudwatch.New(configProvider)
	return nil
}

/*
 * Fetch available metrics for given CloudWatch Namespace
 */
func (c *CloudWatch) fetchNamespaceMetrics() ([]*cloudwatch.Metric, error) {
	if c.metricCache != nil && c.metricCache.IsValid() {
		return c.metricCache.Metrics, nil
	}

	metrics := []*cloudwatch.Metric{}

	var token *string
	for more := true; more; {
		params := &cloudwatch.ListMetricsInput{
			Namespace:  aws.String(c.Namespace),
			Dimensions: []*cloudwatch.DimensionFilter{},
			NextToken:  token,
			MetricName: nil,
		}

		resp, err := c.client.ListMetrics(params)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, resp.Metrics...)

		token = resp.NextToken
		more = token != nil
	}

	c.metricCache = &MetricCache{
		Metrics: metrics,
		Fetched: time.Now(),
		TTL:     c.CacheTTL.Duration,
	}

	return metrics, nil
}

/*
 * Gather given Metric and emit any error
 */
func (c *CloudWatch) gatherMetric(
	acc telegraf.Accumulator,
	metric *cloudwatch.Metric,
	now time.Time,
) error {
	params := c.getStatisticsInput(metric, now)
	resp, err := c.client.GetMetricStatistics(params)
	if err != nil {
		return err
	}

	for _, point := range resp.Datapoints {
		tags := map[string]string{
			"region": c.Region,
			"unit":   snakeCase(*point.Unit),
		}

		for _, d := range metric.Dimensions {
			tags[snakeCase(*d.Name)] = *d.Value
		}

		// record field for each statistic
		fields := map[string]interface{}{}

		if point.Average != nil {
			fields[formatField(*metric.MetricName, cloudwatch.StatisticAverage)] = *point.Average
		}
		if point.Maximum != nil {
			fields[formatField(*metric.MetricName, cloudwatch.StatisticMaximum)] = *point.Maximum
		}
		if point.Minimum != nil {
			fields[formatField(*metric.MetricName, cloudwatch.StatisticMinimum)] = *point.Minimum
		}
		if point.SampleCount != nil {
			fields[formatField(*metric.MetricName, cloudwatch.StatisticSampleCount)] = *point.SampleCount
		}
		if point.Sum != nil {
			fields[formatField(*metric.MetricName, cloudwatch.StatisticSum)] = *point.Sum
		}

		acc.AddFields(formatMeasurement(c.Namespace), fields, tags, *point.Timestamp)
	}

	return nil
}

/*
 * Formatting helpers
 */
func formatField(metricName string, statistic string) string {
	return fmt.Sprintf("%s_%s", snakeCase(metricName), snakeCase(statistic))
}

func formatMeasurement(namespace string) string {
	namespace = strings.Replace(namespace, "/", "_", -1)
	namespace = snakeCase(namespace)
	return fmt.Sprintf("cloudwatch_%s", namespace)
}

func snakeCase(s string) string {
	s = internal.SnakeCase(s)
	s = strings.Replace(s, "__", "_", -1)
	return s
}

/*
 * Map Metric to *cloudwatch.GetMetricStatisticsInput for given timeframe
 */
func (c *CloudWatch) getStatisticsInput(metric *cloudwatch.Metric, now time.Time) *cloudwatch.GetMetricStatisticsInput {
	end := now.Add(-c.Delay.Duration)

	input := &cloudwatch.GetMetricStatisticsInput{
		StartTime:  aws.Time(end.Add(-c.Period.Duration)),
		EndTime:    aws.Time(end),
		MetricName: metric.MetricName,
		Namespace:  metric.Namespace,
		Period:     aws.Int64(int64(c.Period.Duration.Seconds())),
		Dimensions: metric.Dimensions,
		Statistics: []*string{
			aws.String(cloudwatch.StatisticAverage),
			aws.String(cloudwatch.StatisticMaximum),
			aws.String(cloudwatch.StatisticMinimum),
			aws.String(cloudwatch.StatisticSum),
			aws.String(cloudwatch.StatisticSampleCount)},
	}
	return input
}

/*
 * Check Metric Cache validity
 */
func (c *MetricCache) IsValid() bool {
	return c.Metrics != nil && time.Since(c.Fetched) < c.TTL
}

func hasWilcard(dimensions []*Dimension) bool {
	for _, d := range dimensions {
		if d.Value == "" || d.Value == "*" {
			return true
		}
	}
	return false
}

func isSelected(name string, metric *cloudwatch.Metric, dimensions []*Dimension) bool {
	if name != *metric.MetricName {
		return false
	}
	if len(metric.Dimensions) != len(dimensions) {
		return false
	}
	for _, d := range dimensions {
		selected := false
		for _, d2 := range metric.Dimensions {
			if d.Name == *d2.Name {
				if d.Value == "" || d.Value == "*" || d.Value == *d2.Value {
					selected = true
				}
			}
		}
		if !selected {
			return false
		}
	}
	return true
}
