package cloudwatch

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type (
	CloudWatch struct {
		Region      string            `toml:"region"`
		Period      internal.Duration `toml:"period"`
		Delay       internal.Duration `toml:"delay"`
		Namespace   string            `toml:"namespace"`
		Metrics     []*Metric         `toml:"metrics"`
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
  region = 'us-east-1'

  ## Requested CloudWatch aggregation Period (required - must be a multiple of 60s)
  period = '1m'

  ## Collection Delay (required - must account for metrics availability via CloudWatch API)
  delay = '1m'

  ## Recomended: use metric 'interval' that is a multiple of 'period' to avoid 
  ## gaps or overlap in pulled data
  interval = '1m'

  ## Metric Statistic Namespace (required)
  namespace = 'AWS/ELB'

  ## Metrics to Pull (optional)
  ## Defaults to all Metrics in Namespace if nothing is provided
  ## Refreshes Namespace available metrics every 1h
  #[[inputs.cloudwatch.metrics]]
  #  names = ['Latency', 'RequestCount']
  #	
  #  ## Dimension filters for Metric (optional)
  #  [[inputs.cloudwatch.metrics.dimensions]]
  #    name = 'LoadBalancerName'
  #    value = 'p-example'
`
}

func (c *CloudWatch) Description() string {
	return "Pull Metric Statistics from Amazon CloudWatch"
}

func (c *CloudWatch) Gather(acc telegraf.Accumulator) error {
	if c.client == nil {
		c.initializeCloudWatch()
	}

	var metrics []*cloudwatch.Metric

	// check for provided metric filter
	if c.Metrics != nil {
		metrics = []*cloudwatch.Metric{}
		for _, m := range c.Metrics {
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
		}
	} else {
		var err error
		metrics, err = c.fetchNamespaceMetrics()
		if err != nil {
			return err
		}
	}

	metricCount := len(metrics)
	var errChan = make(chan error, metricCount)

	now := time.Now()

	// limit concurrency or we can easily exhaust user connection limit
	semaphore := make(chan byte, 64)

	for _, m := range metrics {
		semaphore <- 0x1
		go c.gatherMetric(acc, m, now, semaphore, errChan)
	}

	for i := 1; i <= metricCount; i++ {
		err := <-errChan
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	inputs.Add("cloudwatch", func() telegraf.Input {
		return &CloudWatch{}
	})
}

/*
 * Initialize CloudWatch client
 */
func (c *CloudWatch) initializeCloudWatch() error {
	config := &aws.Config{
		Region: aws.String(c.Region),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
			}),
	}

	c.client = cloudwatch.New(session.New(config))
	return nil
}

/*
 * Fetch available metrics for given CloudWatch Namespace
 */
func (c *CloudWatch) fetchNamespaceMetrics() (metrics []*cloudwatch.Metric, err error) {
	if c.metricCache != nil && c.metricCache.IsValid() {
		metrics = c.metricCache.Metrics
		return
	}

	metrics = []*cloudwatch.Metric{}

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

	cacheTTL, _ := time.ParseDuration("1hr")
	c.metricCache = &MetricCache{
		Metrics: metrics,
		Fetched: time.Now(),
		TTL:     cacheTTL,
	}

	return
}

/*
 * Gather given Metric and emit any error
 */
func (c *CloudWatch) gatherMetric(acc telegraf.Accumulator, metric *cloudwatch.Metric, now time.Time, semaphore chan byte, errChan chan error) {
	params := c.getStatisticsInput(metric, now)
	resp, err := c.client.GetMetricStatistics(params)
	if err != nil {
		errChan <- err
		<-semaphore
		return
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

	errChan <- nil
	<-semaphore
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
