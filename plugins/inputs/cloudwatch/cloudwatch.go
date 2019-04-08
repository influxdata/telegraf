package cloudwatch

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type (
	// CloudWatch contains the configuration and cache for the cloudwatch plugin.
	CloudWatch struct {
		Region           string   `toml:"region"`
		AccessKey        string   `toml:"access_key"`
		SecretKey        string   `toml:"secret_key"`
		RoleARN          string   `toml:"role_arn"`
		Profile          string   `toml:"profile"`
		CredentialPath   string   `toml:"shared_credential_file"`
		Token            string   `toml:"token"`
		EndpointURL      string   `toml:"endpoint_url"`
		StatisticExclude []string `toml:"statistic_exclude"`
		StatisticInclude []string `toml:"statistic_include"`

		Period      internal.Duration `toml:"period"`
		Delay       internal.Duration `toml:"delay"`
		Namespace   string            `toml:"namespace"`
		Metrics     []*Metric         `toml:"metrics"`
		CacheTTL    internal.Duration `toml:"cache_ttl"`
		RateLimit   int               `toml:"ratelimit"`
		client      cloudwatchClient
		metricCache *MetricCache
		queryCache  []*cloudwatch.MetricDataQuery
		queries     []queryData
		windowStart time.Time
		windowEnd   time.Time
	}

	// Metric defines a simplified Cloudwatch metric.
	Metric struct {
		StatisticExclude *[]string    `toml:"statistic_exclude"`
		StatisticInclude *[]string    `toml:"statistic_include"`
		MetricNames      []string     `toml:"names"`
		Dimensions       []*Dimension `toml:"dimensions"`
	}

	// Dimension defines a simplified Cloudwatch dimension (provides metric filtering).
	Dimension struct {
		Name  string `toml:"name"`
		Value string `toml:"value"`
	}

	// MetricCache caches automatically fetched metrics.
	MetricCache struct {
		TTL     time.Duration
		Fetched time.Time
		Metrics []*cloudwatch.Metric
	}

	cloudwatchClient interface {
		ListMetrics(*cloudwatch.ListMetricsInput) (*cloudwatch.ListMetricsOutput, error)
		GetMetricData(*cloudwatch.GetMetricDataInput) (*cloudwatch.GetMetricDataOutput, error)
	}
)

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

  ## Configure the TTL for the internal cache of metrics.
  ## Defaults to 1 hr if not specified
  # cache_ttl = "1h"

  ## Metric Statistic Namespace (required)
  namespace = "AWS/ELB"

  ## Maximum requests per second. Note that the global default AWS rate limit is
  ## 400 reqs/sec, so if you define multiple namespaces, these should add up to a
  ## maximum of 400. Default value is 200.
  ## See http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html
  # ratelimit = 200

  ## Namespace-wide statistic filters. These allow fewer queries to be made to
  ## cloudwatch.
  # statistic_exclude = [ "average", "sum", minimum", "maximum", sample_count" ]
  # statistic_include = []

  ## Metrics to Pull
  ## Defaults to all Metrics in Namespace if nothing is provided
  ## Refreshes Namespace available metrics every 1h
  #[[inputs.cloudwatch.metrics]]
  #  names = ["Latency", "RequestCount"]
  #
  #  ## Statistic filters for Metric.  These allow for retrieving specific
  #  ## statistics for an individual metric.
  #  # statistic_exclude = [ "average", "sum", minimum", "maximum", sample_count" ]
  #  # statistic_include = []
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

type filteredMetric struct {
	metrics    []*cloudwatch.Metric
	statFilter filter.Filter
}

// selectMetrics returns metrics specified in the config file or metrics listed from Cloudwatch.
func selectMetrics(c *CloudWatch) ([]filteredMetric, error) {
	fMetrics := []filteredMetric{}
	var metrics []*cloudwatch.Metric

	// check for provided metric filter
	if c.Metrics != nil {
		for _, m := range c.Metrics {
			metrics = []*cloudwatch.Metric{}
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
		var err error
		metrics, err = c.fetchNamespaceMetrics()
		if err != nil {
			return nil, err
		}

		// use config level filters
		statFilter, err := filter.NewIncludeExcludeFilter(c.StatisticInclude, c.StatisticExclude)
		if err != nil {
			return nil, err
		}

		fMetrics = []filteredMetric{{
			metrics:    metrics,
			statFilter: statFilter,
		}}
	}
	return fMetrics, nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval".
func (c *CloudWatch) Gather(acc telegraf.Accumulator) error {
	if c.client == nil {
		c.initializeCloudWatch()
	}

	metrics, err := selectMetrics(c)
	if err != nil {
		return err
	}

	err = c.updateWindow(time.Now())
	if err != nil {
		return err
	}

	// Get all of the possible queries so we can send groups of 100.
	// note: these are cached using metricCache's specs (when only namespace is defined)
	queries, err := c.getDataQueries(metrics)
	if err != nil {
		return err
	}

	// Limit concurrency or we can easily exhaust user connection limit.
	// See cloudwatch API request limits:
	// http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_limits.html
	lmtr := limiter.NewRateLimiter(c.RateLimit, time.Second)
	defer lmtr.Stop()
	wg := sync.WaitGroup{}
	rLock := sync.Mutex{}
	// create master list of results so aggregation works the best way.
	results := []*cloudwatch.MetricDataResult{}

	var aggregateResults = func(inm []*cloudwatch.MetricDataQuery) {
		defer wg.Done()
		result, err := c.gatherMetrics(c.getDataInputs(inm))
		if err != nil {
			acc.AddError(err)
			return
		}

		rLock.Lock()
		results = append(results, result...)
		rLock.Unlock()
	}

	// 100 is the maximum number of queries a `GetMetricData` request can contain.
	batchSize := 100
	var batches [][]*cloudwatch.MetricDataQuery

	for batchSize < len(queries) {
		queries, batches = queries[batchSize:], append(batches, queries[0:batchSize:batchSize])
	}
	batches = append(batches, queries)

	for i := range batches {
		wg.Add(1)
		<-lmtr.C
		go aggregateResults(batches[i])
	}

	wg.Wait()

	return c.aggregateMetrics(acc, results)
}

func (c *CloudWatch) updateWindow(relativeTo time.Time) error {
	windowEnd := relativeTo.Add(-c.Delay.Duration)

	if c.windowEnd.IsZero() {
		// this is the first run, no window info, so just get a single period
		c.windowStart = windowEnd.Add(-c.Period.Duration)
	} else {
		// subsequent window, start where last window left off
		c.windowStart = c.windowEnd
	}

	c.windowEnd = windowEnd

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

func (c *CloudWatch) initializeCloudWatch() error {
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

	cfg := &aws.Config{}
	loglevel := aws.LogOff
	c.client = cloudwatch.New(configProvider, cfg.WithLogLevel(loglevel))
	return nil
}

// fetchNamespaceMetrics retrieves available metrics for a given CloudWatch namespace.
func (c *CloudWatch) fetchNamespaceMetrics() ([]*cloudwatch.Metric, error) {
	if c.metricCache != nil && c.metricCache.isValid() {
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
	namespace := sanitizeMeasurement(c.Namespace)

	for _, query := range c.queries {
		tags := map[string]string{
			"region": c.Region,
		}

		type metric struct {
			fields map[string]interface{}
			tags   map[string]string
		}
		results := map[time.Time]metric{}
		for _, result := range metricDataResults {
			if nameMatch(query.id, *result.Id) {
				for _, dimension := range query.metric.Dimensions {
					tags[snakeCase(*dimension.Name)] = *dimension.Value
				}

				for j := range result.Values {
					if _, ok := results[*result.Timestamps[j]]; !ok {
						results[*result.Timestamps[j]] = metric{fields: map[string]interface{}{*result.Label: *result.Values[j]}, tags: tags}
					}
					results[*result.Timestamps[j]].fields[*result.Label] = *result.Values[j]
				}
			}
		}

		for t, m := range results {
			acc.AddFields(namespace, m.fields, m.tags, t)
		}
	}

	return nil
}

func nameMatch(name, id string) bool {
	if strings.TrimPrefix(id, "average_") == name {
		return true
	}
	if strings.TrimPrefix(id, "maximum_") == name {
		return true
	}
	if strings.TrimPrefix(id, "minimum_") == name {
		return true
	}
	if strings.TrimPrefix(id, "sum_") == name {
		return true
	}
	if strings.TrimPrefix(id, "sample_count_") == name {
		return true
	}
	return false
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

type queryData struct {
	metric *cloudwatch.Metric
	id     string
}

// getDataQueries gets all of the possible queries so we can maximize the request payload.
func (c *CloudWatch) getDataQueries(filteredMetrics []filteredMetric) ([]*cloudwatch.MetricDataQuery, error) {
	if c.queryCache != nil && c.metricCache != nil && c.metricCache.isValid() {
		return c.queryCache, nil
	}

	c.queries = []queryData{}

	dataQueries := []*cloudwatch.MetricDataQuery{}
	for i, filtered := range filteredMetrics {
		for j, metric := range filtered.metrics {
			id := strconv.Itoa(j) + "_" + strconv.Itoa(i)
			c.queries = append(c.queries, queryData{metric: metric, id: id})

			if filtered.statFilter.Match("average") {
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("average_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_average")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(c.Period.Duration.Seconds())),
						Stat:   aws.String(cloudwatch.StatisticAverage),
					},
				})
			}
			if filtered.statFilter.Match("maximum") {
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("maximum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_maximum")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(c.Period.Duration.Seconds())),
						Stat:   aws.String(cloudwatch.StatisticMaximum),
					},
				})
			}
			if filtered.statFilter.Match("minimum") {
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("minimum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_minimum")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(c.Period.Duration.Seconds())),
						Stat:   aws.String(cloudwatch.StatisticMinimum),
					},
				})
			}
			if filtered.statFilter.Match("sum") {
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("sum_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_sum")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(c.Period.Duration.Seconds())),
						Stat:   aws.String(cloudwatch.StatisticSum),
					},
				})
			}
			if filtered.statFilter.Match("sample_count") {
				dataQueries = append(dataQueries, &cloudwatch.MetricDataQuery{
					Id:    aws.String("sample_count_" + id),
					Label: aws.String(snakeCase(*metric.MetricName + "_sample_count")),
					MetricStat: &cloudwatch.MetricStat{
						Metric: metric,
						Period: aws.Int64(int64(c.Period.Duration.Seconds())),
						Stat:   aws.String(cloudwatch.StatisticSampleCount),
					},
				})
			}
		}
	}

	if len(dataQueries) == 0 {
		return nil, errors.New("no metrics found to collect")
	}

	c.queryCache = dataQueries
	return dataQueries, nil
}

func (c *CloudWatch) getDataInputs(dataQueries []*cloudwatch.MetricDataQuery) *cloudwatch.GetMetricDataInput {
	return &cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(c.windowStart),
		EndTime:           aws.Time(c.windowEnd),
		MetricDataQueries: dataQueries,
	}
}

// isValid checks the validity of the metric cache.
func (c *MetricCache) isValid() bool {
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
