package cloudwatch

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cwClient "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/limiter"
	internalMetric "github.com/influxdata/telegraf/metric"
	internalProxy "github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	StatisticAverage     = "Average"
	StatisticMaximum     = "Maximum"
	StatisticMinimum     = "Minimum"
	StatisticSum         = "Sum"
	StatisticSampleCount = "SampleCount"
)

// CloudWatch contains the configuration and cache for the cloudwatch plugin.
type CloudWatch struct {
	StatisticExclude []string        `toml:"statistic_exclude"`
	StatisticInclude []string        `toml:"statistic_include"`
	Timeout          config.Duration `toml:"timeout"`

	internalProxy.HTTPProxy

	Period         config.Duration `toml:"period"`
	Delay          config.Duration `toml:"delay"`
	Namespace      string          `toml:"namespace"`
	Namespaces     []string        `toml:"namespaces"`
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

	internalaws.CredentialConfig
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
	Name         string `toml:"name"`
	Value        string `toml:"value"`
	valueMatcher filter.Filter
}

// metricCache caches metrics, their filters, and generated queries.
type metricCache struct {
	ttl     time.Duration
	built   time.Time
	metrics []filteredMetric
	queries map[string][]types.MetricDataQuery
}

type cloudwatchClient interface {
	ListMetrics(context.Context, *cwClient.ListMetricsInput, ...func(*cwClient.Options)) (*cwClient.ListMetricsOutput, error)
	GetMetricData(context.Context, *cwClient.GetMetricDataInput, ...func(*cwClient.Options)) (*cwClient.GetMetricDataOutput, error)
}

// SampleConfig returns the default configuration of the Cloudwatch input plugin.
func (c *CloudWatch) SampleConfig() string {
	return `
  ## Amazon Region
  region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Web identity provider credentials via STS if role_arn and web_identity_token_file are specified
  ## 2) Assumed credentials via STS if role_arn is specified
  ## 3) explicit credentials from 'access_key' and 'secret_key'
  ## 4) shared profile from 'profile'
  ## 5) environment variables
  ## 6) shared credentials file
  ## 7) EC2 Instance Profile
  # access_key = ""
  # secret_key = ""
  # token = ""
  # role_arn = ""
  # web_identity_token_file = ""
  # role_session_name = ""
  # profile = ""
  # shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Set http_proxy (telegraf uses the system wide proxy settings if it's is not set)
  # http_proxy_url = "http://localhost:8888"

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

  ## Metric Statistic Namespaces (required)
  namespaces = ["AWS/ELB"]
  # A single metric statistic namespace that will be appended to namespaces on startup
  # namespace = "AWS/ELB"

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
  #  ## 'value' has wildcard / 'glob' matching support such as 'p-*'.
  #  [[inputs.cloudwatch.metrics.dimensions]]
  #    name = "LoadBalancerName"
  #    value = "p-example"
`
}

// Description returns a one-sentence description on the Cloudwatch input plugin.
func (c *CloudWatch) Description() string {
	return "Pull Metric Statistics from Amazon CloudWatch"
}

func (c *CloudWatch) Init() error {
	if len(c.Namespace) != 0 {
		c.Namespaces = append(c.Namespaces, c.Namespace)
	}

	err := c.initializeCloudWatch()
	if err != nil {
		return err
	}

	// Set config level filter (won't change throughout life of plugin).
	c.statFilter, err = filter.NewIncludeExcludeFilter(c.StatisticInclude, c.StatisticExclude)
	if err != nil {
		return err
	}

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval".
func (c *CloudWatch) Gather(acc telegraf.Accumulator) error {
	filteredMetrics, err := getFilteredMetrics(c)
	if err != nil {
		return err
	}

	c.updateWindow(time.Now())

	// Get all of the possible queries so we can send groups of 100.
	queries := c.getDataQueries(filteredMetrics)
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

	results := map[string][]types.MetricDataResult{}

	for namespace, namespacedQueries := range queries {
		// 500 is the maximum number of metric data queries a `GetMetricData` request can contain.
		batchSize := 500
		var batches [][]types.MetricDataQuery

		for batchSize < len(namespacedQueries) {
			namespacedQueries, batches = namespacedQueries[batchSize:], append(batches, namespacedQueries[0:batchSize:batchSize])
		}
		batches = append(batches, namespacedQueries)

		for i := range batches {
			wg.Add(1)
			<-lmtr.C
			go func(n string, inm []types.MetricDataQuery) {
				defer wg.Done()
				result, err := c.gatherMetrics(c.getDataInputs(inm))
				if err != nil {
					acc.AddError(err)
					return
				}

				rLock.Lock()
				results[n] = append(results[n], result...)
				rLock.Unlock()
			}(namespace, batches[i])
		}
	}

	wg.Wait()

	return c.aggregateMetrics(acc, results)
}

func (c *CloudWatch) initializeCloudWatch() error {
	proxy, err := c.HTTPProxy.Proxy()
	if err != nil {
		return err
	}

	cfg, err := c.CredentialConfig.Credentials()
	if err != nil {
		return err
	}
	c.client = cwClient.NewFromConfig(cfg, func(options *cwClient.Options) {
		// Disable logging
		options.ClientLogMode = 0

		options.HTTPClient = &http.Client{
			// use values from DefaultTransport
			Transport: &http.Transport{
				Proxy: proxy,
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
		}
	})

	// Initialize regex matchers for each Dimension value.
	for _, m := range c.Metrics {
		for _, dimension := range m.Dimensions {
			matcher, err := filter.NewIncludeExcludeFilter([]string{dimension.Value}, nil)
			if err != nil {
				return err
			}

			dimension.valueMatcher = matcher
		}
	}

	return nil
}

type filteredMetric struct {
	metrics    []types.Metric
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
			metrics := []types.Metric{}
			if !hasWildcard(m.Dimensions) {
				dimensions := make([]types.Dimension, len(m.Dimensions))
				for k, d := range m.Dimensions {
					dimensions[k] = types.Dimension{
						Name:  aws.String(d.Name),
						Value: aws.String(d.Value),
					}
				}
				for _, name := range m.MetricNames {
					for _, namespace := range c.Namespaces {
						metrics = append(metrics, types.Metric{
							Namespace:  aws.String(namespace),
							MetricName: aws.String(name),
							Dimensions: dimensions,
						})
					}
				}
			} else {
				allMetrics, err := c.fetchNamespaceMetrics()
				if err != nil {
					return nil, err
				}
				for _, name := range m.MetricNames {
					for _, metric := range allMetrics {
						if isSelected(name, metric, m.Dimensions) {
							for _, namespace := range c.Namespaces {
								metrics = append(metrics, types.Metric{
									Namespace:  aws.String(namespace),
									MetricName: aws.String(name),
									Dimensions: metric.Dimensions,
								})
							}
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
func (c *CloudWatch) fetchNamespaceMetrics() ([]types.Metric, error) {
	metrics := []types.Metric{}

	var token *string

	params := &cwClient.ListMetricsInput{
		Dimensions: []types.DimensionFilter{},
		NextToken:  token,
		MetricName: nil,
	}
	if c.RecentlyActive == "PT3H" {
		params.RecentlyActive = types.RecentlyActivePt3h
	}

	for _, namespace := range c.Namespaces {
		params.Namespace = aws.String(namespace)
		for {
			resp, err := c.client.ListMetrics(context.Background(), params)
			if err != nil {
				return nil, fmt.Errorf("failed to list metrics with params per namespace: %v", err)
			}

			metrics = append(metrics, resp.Metrics...)
			if resp.NextToken == nil {
				break
			}

			params.NextToken = resp.NextToken
		}
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
func (c *CloudWatch) getDataQueries(filteredMetrics []filteredMetric) map[string][]types.MetricDataQuery {
	if c.metricCache != nil && c.metricCache.queries != nil && c.metricCache.isValid() {
		return c.metricCache.queries
	}

	c.queryDimensions = map[string]*map[string]string{}

	dataQueries := map[string][]types.MetricDataQuery{}
	for i, filtered := range filteredMetrics {
		for j, metric := range filtered.metrics {
			id := strconv.Itoa(j) + "_" + strconv.Itoa(i)
			dimension := ctod(metric.Dimensions)
			if filtered.statFilter.Match("average") {
				c.queryDimensions["average_"+id] = dimension
				dataQueries[*metric.Namespace] = append(dataQueries[*metric.Namespace], types.MetricDataQuery{
					Id:    aws.String("average_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_average")),
					MetricStat: &types.MetricStat{
						Metric: &filtered.metrics[j],
						Period: aws.Int32(int32(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(StatisticAverage),
					},
				})
			}
			if filtered.statFilter.Match("maximum") {
				c.queryDimensions["maximum_"+id] = dimension
				dataQueries[*metric.Namespace] = append(dataQueries[*metric.Namespace], types.MetricDataQuery{
					Id:    aws.String("maximum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_maximum")),
					MetricStat: &types.MetricStat{
						Metric: &filtered.metrics[j],
						Period: aws.Int32(int32(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(StatisticMaximum),
					},
				})
			}
			if filtered.statFilter.Match("minimum") {
				c.queryDimensions["minimum_"+id] = dimension
				dataQueries[*metric.Namespace] = append(dataQueries[*metric.Namespace], types.MetricDataQuery{
					Id:    aws.String("minimum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_minimum")),
					MetricStat: &types.MetricStat{
						Metric: &filtered.metrics[j],
						Period: aws.Int32(int32(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(StatisticMinimum),
					},
				})
			}
			if filtered.statFilter.Match("sum") {
				c.queryDimensions["sum_"+id] = dimension
				dataQueries[*metric.Namespace] = append(dataQueries[*metric.Namespace], types.MetricDataQuery{
					Id:    aws.String("sum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_sum")),
					MetricStat: &types.MetricStat{
						Metric: &filtered.metrics[j],
						Period: aws.Int32(int32(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(StatisticSum),
					},
				})
			}
			if filtered.statFilter.Match("sample_count") {
				c.queryDimensions["sample_count_"+id] = dimension
				dataQueries[*metric.Namespace] = append(dataQueries[*metric.Namespace], types.MetricDataQuery{
					Id:    aws.String("sample_count_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_sample_count")),
					MetricStat: &types.MetricStat{
						Metric: &filtered.metrics[j],
						Period: aws.Int32(int32(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(StatisticSampleCount),
					},
				})
			}
		}
	}

	if len(dataQueries) == 0 {
		c.Log.Debug("no metrics found to collect")
		return nil
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

	return dataQueries
}

// gatherMetrics gets metric data from Cloudwatch.
func (c *CloudWatch) gatherMetrics(
	params *cwClient.GetMetricDataInput,
) ([]types.MetricDataResult, error) {
	results := []types.MetricDataResult{}

	for {
		resp, err := c.client.GetMetricData(context.Background(), params)
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
	metricDataResults map[string][]types.MetricDataResult,
) error {
	var (
		grouper = internalMetric.NewSeriesGrouper()
	)

	for namespace, results := range metricDataResults {
		namespace = sanitizeMeasurement(namespace)

		for _, result := range results {
			tags := map[string]string{}

			if dimensions, ok := c.queryDimensions[*result.Id]; ok {
				tags = *dimensions
			}
			tags["region"] = c.Region

			for i := range result.Values {
				if err := grouper.Add(namespace, tags, result.Timestamps[i], *result.Label, result.Values[i]); err != nil {
					acc.AddError(err)
				}
			}
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

// ctod converts cloudwatch dimensions to regular dimensions.
func ctod(cDimensions []types.Dimension) *map[string]string {
	dimensions := map[string]string{}
	for i := range cDimensions {
		dimensions[snakeCase(*cDimensions[i].Name)] = *cDimensions[i].Value
	}
	return &dimensions
}

func (c *CloudWatch) getDataInputs(dataQueries []types.MetricDataQuery) *cwClient.GetMetricDataInput {
	return &cwClient.GetMetricDataInput{
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
		if d.Value == "" || strings.ContainsAny(d.Value, "*?[") {
			return true
		}
	}
	return false
}

func isSelected(name string, metric types.Metric, dimensions []*Dimension) bool {
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
				if d.Value == "" || d.valueMatcher.Match(*d2.Value) {
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
