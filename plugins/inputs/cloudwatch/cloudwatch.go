package cloudwatch

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// CloudWatch contains the configuration and cache for the cloudwatch plugin.
type CloudWatch struct {
	Region           string          `toml:"region"`
	AccessKey        string          `toml:"access_key"`
	SecretKey        string          `toml:"secret_key"`
	RoleARN          string          `toml:"role_arn"`
	Profile          string          `toml:"profile"`
	CredentialPath   string          `toml:"shared_credential_file"`
	Token            string          `toml:"token"`
	EndpointURL      string          `toml:"endpoint_url"`
	StatisticExclude []string        `toml:"statistic_exclude"`
	StatisticInclude []string        `toml:"statistic_include"`
	Timeout          config.Duration `toml:"timeout"`

	Period         config.Duration `toml:"period"`
	Delay          config.Duration `toml:"delay"`
	Namespace      string          `toml:"namespace"`
	Metrics        []*Metric       `toml:"metrics"`
	CacheTTL       config.Duration `toml:"cache_ttl"`
	RateLimit      int             `toml:"ratelimit"`
	RecentlyActive string          `toml:"recently_active"`

	Log telegraf.Logger `toml:"-"`

	client          cloudwatchClient
	statFilter      filter.Filter
	metricCache     *metricCache
	queryDimensions map[string]*map[string]string
	windowStart     time.Time
	windowEnd       time.Time
}

// Metric defines a simplified Cloudwatch metric.
type Metric struct {
	StatisticExclude *[]string    `toml:"statistic_exclude"`
	StatisticInclude *[]string    `toml:"statistic_include"`
	MetricNames      []string     `toml:"names"`
	Dimensions       []*Dimension `toml:"dimensions"`
}

// Dimension defines a simplified Cloudwatch dimension (provides metric filtering).
type Dimension struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

// metricCache caches metrics, their filters, and generated queries.
type metricCache struct {
	ttl     time.Duration
	built   time.Time
	metrics []filteredMetric
	queries []*cloudwatch.MetricDataQuery
}

type cloudwatchClient interface {
	ListMetrics(*cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error)
	GetMetricData(*cloudwatch.GetMetricDataInput) (*cloudwatch.GetMetricDataOutput, error)
}

// SampleConfig returns the default configuration of the Cloudwatch input plugin.
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
  # access_key = ""
  # secret_key = ""
  # token = ""
  # role_arn = ""
  # profile = ""
  # shared_credential_file = ""

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

  ## Recommended if "delay" and "period" are both within 3 hours of request time. Invalid values will be ignored.
  ## Recently Active feature will only poll for CloudWatch ListMetrics values that occurred within the last 3 Hours.
  ## If enabled, it will reduce total API usage of the CloudWatch ListMetrics API and require less memory to retain.
  ## Do not enable if "period" or "delay" is longer than 3 hours, as it will not return data more than 3 hours old.
  ## See https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_ListMetrics.html
  #recently_active = "PT3H"

  ## Configure the TTL for the internal cache of metrics.
  # cache_ttl = "1h"

  ## Metric Statistic Namespace (required)
  namespace = "AWS/ELB"

  ## Maximum requests per second. Note that the global default AWS rate limit is
  ## 50 reqs/sec, so if you define multiple namespaces, these should add up to a
  ## maximum of 50.
  ## See http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html
  # ratelimit = 25

  ## Timeout for http requests made by the cloudwatch client.
  # timeout = "5s"

  ## Namespace-wide statistic filters. These allow fewer queries to be made to
  ## cloudwatch.
  # statistic_include = [ "average", "sum", "minimum", "maximum", sample_count" ]
  # statistic_exclude = []

  ## Metrics to Pull
  ## Defaults to all Metrics in Namespace if nothing is provided
  ## Refreshes Namespace available metrics every 1h
  #[[inputs.cloudwatch.metrics]]
  #  names = ["Latency", "RequestCount"]
  #
  #  ## Statistic filters for Metric.  These allow for retrieving specific
  #  ## statistics for an individual metric.
  #  # statistic_include = [ "average", "sum", "minimum", "maximum", sample_count" ]
  #  # statistic_exclude = []
  #
  #  ## Dimension filters for Metric.  All dimensions defined for the metric names
  #  ## must be specified in order to retrieve the metric statistics.
  #  [[inputs.cloudwatch.metrics.dimensions]]
  #    name = "LoadBalancerName"
  #    value = "p-example"
`
}

// Description returns a one-sentence description on the Cloudwatch input plugin.
func (c *CloudWatch) Description() string {
	return "Pull Metric Statistics from Amazon CloudWatch"
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval".
func (c *CloudWatch) Gather(acc telegraf.Accumulator) error {
	if c.statFilter == nil {
		var err error
		// Set config level filter (won't change throughout life of plugin).
		c.statFilter, err = filter.NewIncludeExcludeFilter(c.StatisticInclude, c.StatisticExclude)
		if err != nil {
			return err
		}
	}

	if c.client == nil {
		c.initializeCloudWatch()
	}

	filteredMetrics, err := getFilteredMetrics(c)
	if err != nil {
		return err
	}

	c.updateWindow(time.Now())

	// Get all of the possible queries so we can send groups of 100.
	queries, err := c.getDataQueries(filteredMetrics)
	if err != nil {
		return err
	}

	if len(queries) == 0 {
		return nil
	}

	// Limit concurrency or we can easily exhaust user connection limit.
	// See cloudwatch API request limits:
	// http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_limits.html
	lmtr := limiter.NewRateLimiter(c.RateLimit, time.Second)
	defer lmtr.Stop()
	wg := sync.WaitGroup{}
	rLock := sync.Mutex{}

	results := []*cloudwatch.MetricDataResult{}

	// 500 is the maximum number of metric data queries a `GetMetricData` request can contain.
	batchSize := 500
	var batches [][]*cloudwatch.MetricDataQuery

	for batchSize < len(queries) {
		queries, batches = queries[batchSize:], append(batches, queries[0:batchSize:batchSize])
	}
	batches = append(batches, queries)

	for i := range batches {
		wg.Add(1)
		<-lmtr.C
		go func(inm []*cloudwatch.MetricDataQuery) {
			defer wg.Done()
			result, err := c.gatherMetrics(c.getDataInputs(inm))
			if err != nil {
				acc.AddError(err)
				return
			}

			rLock.Lock()
			results = append(results, result...)
			rLock.Unlock()
		}(batches[i])
	}

	wg.Wait()

	return c.aggregateMetrics(acc, results)
}

func (c *CloudWatch) initializeCloudWatch() {
	credentialConfig := &internalaws.CredentialConfig{
		Region:      c.Region,
		AccessKey:   c.AccessKey,
		SecretKey:   c.SecretKey,
		RoleARN:     c.RoleARN,
		Profile:     c.Profile,
		Filename:    c.CredentialPath,
		Token:       c.Token,
		EndpointURL: c.EndpointURL,
	}
	configProvider := credentialConfig.Credentials()

	cfg := &aws.Config{
		HTTPClient: &http.Client{
			// use values from DefaultTransport
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: time.Duration(c.Timeout),
		},
	}

	loglevel := aws.LogOff
	c.client = cloudwatch.New(configProvider, cfg.WithLogLevel(loglevel))
}

type filteredMetric struct {
	metrics    []*cloudwatch.Metric
	statFilter filter.Filter
}

// getFilteredMetrics returns metrics specified in the config file or metrics listed from Cloudwatch.
func getFilteredMetrics(c *CloudWatch) ([]filteredMetric, error) {
	if c.metricCache != nil && c.metricCache.isValid() {
		return c.metricCache.metrics, nil
	}

	fMetrics := []filteredMetric{}

	// check for provided metric filter
	if c.Metrics != nil {
		for _, m := range c.Metrics {
			metrics := []*cloudwatch.Metric{}
			if !hasWildcard(m.Dimensions) {
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

			if m.StatisticExclude == nil {
				m.StatisticExclude = &c.StatisticExclude
			}
			if m.StatisticInclude == nil {
				m.StatisticInclude = &c.StatisticInclude
			}
			statFilter, err := filter.NewIncludeExcludeFilter(*m.StatisticInclude, *m.StatisticExclude)
			if err != nil {
				return nil, err
			}

			fMetrics = append(fMetrics, filteredMetric{
				metrics:    metrics,
				statFilter: statFilter,
			})
		}
	} else {
		metrics, err := c.fetchNamespaceMetrics()
		if err != nil {
			return nil, err
		}

		fMetrics = []filteredMetric{{
			metrics:    metrics,
			statFilter: c.statFilter,
		}}
	}

	c.metricCache = &metricCache{
		metrics: fMetrics,
		built:   time.Now(),
		ttl:     time.Duration(c.CacheTTL),
	}

	return fMetrics, nil
}

// fetchNamespaceMetrics retrieves available metrics for a given CloudWatch namespace.
func (c *CloudWatch) fetchNamespaceMetrics() ([]*cloudwatch.Metric, error) {
	metrics := []*cloudwatch.Metric{}

	var token *string
	var params *cloudwatch.ListMetricsInput
	var recentlyActive *string = nil

	switch c.RecentlyActive {
	case "PT3H":
		recentlyActive = &c.RecentlyActive
	default:
		recentlyActive = nil
	}
	params = &cloudwatch.ListMetricsInput{
		Namespace:      aws.String(c.Namespace),
		Dimensions:     []*cloudwatch.DimensionFilter{},
		NextToken:      token,
		MetricName:     nil,
		RecentlyActive: recentlyActive,
	}
	for {
		resp, err := c.client.ListMetrics(params)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, resp.Metrics...)
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return metrics, nil
}

func (c *CloudWatch) updateWindow(relativeTo time.Time) {
	windowEnd := relativeTo.Add(-time.Duration(c.Delay))

	if c.windowEnd.IsZero() {
		// this is the first run, no window info, so just get a single period
		c.windowStart = windowEnd.Add(-time.Duration(c.Period))
	} else {
		// subsequent window, start where last window left off
		c.windowStart = c.windowEnd
	}

	c.windowEnd = windowEnd
}

// getDataQueries gets all of the possible queries so we can maximize the request payload.
func (c *CloudWatch) getDataQueries(filteredMetrics []filteredMetric) ([]*cloudwatch.MetricDataQuery, error) {
	if c.metricCache != nil && c.metricCache.queries != nil && c.metricCache.isValid() {
		return c.metricCache.queries, nil
	}

	c.queryDimensions = map[string]*map[string]string{}

	dataQueries := []*cloudwatch.MetricDataQuery{}
	for i, filtered := range filteredMetrics {
		for j, metric := range filtered.metrics {
			id := strconv.Itoa(j) + "_" + strconv.Itoa(i)
			dimension := ctod(metric.Dimensions)
			if filtered.statFilter.Match("average") {
				c.queryDimensions["average_"+id] = dimension
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("average_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_average")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(cloudwatch.StatisticAverage),
					},
				})
			}
			if filtered.statFilter.Match("maximum") {
				c.queryDimensions["maximum_"+id] = dimension
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("maximum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_maximum")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(cloudwatch.StatisticMaximum),
					},
				})
			}
			if filtered.statFilter.Match("minimum") {
				c.queryDimensions["minimum_"+id] = dimension
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("minimum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_minimum")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(cloudwatch.StatisticMinimum),
					},
				})
			}
			if filtered.statFilter.Match("sum") {
				c.queryDimensions["sum_"+id] = dimension
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("sum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_sum")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(cloudwatch.StatisticSum),
					},
				})
			}
			if filtered.statFilter.Match("sample_count") {
				c.queryDimensions["sample_count_"+id] = dimension
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("sample_count_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_sample_count")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(cloudwatch.StatisticSampleCount),
					},
				})
			}
		}
	}

	if len(dataQueries) == 0 {
		c.Log.Debug("no metrics found to collect")
		return nil, nil
	}

	if c.metricCache == nil {
		c.metricCache = &metricCache{
			queries: dataQueries,
			built:   time.Now(),
			ttl:     time.Duration(c.CacheTTL),
		}
	} else {
		c.metricCache.queries = dataQueries
	}

	return dataQueries, nil
}

// gatherMetrics gets metric data from Cloudwatch.
func (c *CloudWatch) gatherMetrics(
	params *cloudwatch.GetMetricDataInput,
) ([]*cloudwatch.MetricDataResult, error) {
	results := []*cloudwatch.MetricDataResult{}

	for {
		resp, err := c.client.GetMetricData(params)
		if err != nil {
			return nil, fmt.Errorf("failed to get metric data: %v", err)
		}

		results = append(results, resp.MetricDataResults...)
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}

	return results, nil
}

func (c *CloudWatch) aggregateMetrics(
	acc telegraf.Accumulator,
	metricDataResults []*cloudwatch.MetricDataResult,
) error {
	var (
		grouper   = metric.NewSeriesGrouper()
		namespace = sanitizeMeasurement(c.Namespace)
	)

	for _, result := range metricDataResults {
		tags := map[string]string{}

		if dimensions, ok := c.queryDimensions[*result.Id]; ok {
			tags = *dimensions
		}
		tags["region"] = c.Region

		for i := range result.Values {
			grouper.Add(namespace, tags, *result.Timestamps[i], *result.Label, *result.Values[i])
		}
	}

	for _, metric := range grouper.Metrics() {
		acc.AddMetric(metric)
	}

	return nil
}

func init() {
	inputs.Add("cloudwatch", func() telegraf.Input {
		return New()
	})
}

// New instance of the cloudwatch plugin
func New() *CloudWatch {
	return &CloudWatch{
		CacheTTL:  config.Duration(time.Hour),
		RateLimit: 25,
		Timeout:   config.Duration(time.Second * 5),
	}
}

func sanitizeMeasurement(namespace string) string {
	namespace = strings.Replace(namespace, "/", "_", -1)
	namespace = snakeCase(namespace)
	return "cloudwatch_" + namespace
}

func snakeCase(s string) string {
	s = internal.SnakeCase(s)
	s = strings.Replace(s, " ", "_", -1)
	s = strings.Replace(s, "__", "_", -1)
	return s
}

type dimension struct {
	name  string
	value string
}

// ctod converts cloudwatch dimensions to regular dimensions.
func ctod(cDimensions []*cloudwatch.Dimension) *map[string]string {
	dimensions := map[string]string{}
	for i := range cDimensions {
		dimensions[snakeCase(*cDimensions[i].Name)] = *cDimensions[i].Value
	}
	return &dimensions
}

func (c *CloudWatch) getDataInputs(dataQueries []*cloudwatch.MetricDataQuery) *cloudwatch.GetMetricDataInput {
	return &cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(c.windowStart),
		EndTime:           aws.Time(c.windowEnd),
		MetricDataQueries: dataQueries,
	}
}

// isValid checks the validity of the metric cache.
func (f *metricCache) isValid() bool {
	return f.metrics != nil && time.Since(f.built) < f.ttl
}

func hasWildcard(dimensions []*Dimension) bool {
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
