package cloudwatch

import (
	"errors"
	"log"
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

		Debug bool `toml:"debug"`
	}

	Metric struct {
		StatisticExclude []string     `toml:"statistic_exclude"`
		StatisticInclude []string     `toml:"statistic_include"`
		MetricNames      []string     `toml:"names"`
		Dimensions       []*Dimension `toml:"dimensions"`
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
		GetMetricData(*cloudwatch.GetMetricDataInput) (*cloudwatch.GetMetricDataOutput, error)
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
  # cache_ttl = "10m"

  ## Metric Statistic Namespace (required)
  namespace = "AWS/ELB"

  ## Maximum requests per second. Note that the global default AWS rate limit is
  ## 400 reqs/sec, so if you define multiple namespaces, these should add up to a
  ## maximum of 400. Optional - default value is 200.
  ## See http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html
  ratelimit = 200

  ## Namespace-wide statistic filters (only gets used if no metrics are defined). These
  ## are optional and allow fewer queries to be made to cloudwatch.
  # statistic_exclude = [ "average", "sum", minimum", "maximum", sample_count" ]
  # statistic_include = [ "average", "sum", minimum", "maximum", sample_count" ]

  ## Metrics to Pull (optional)
  ## Defaults to all Metrics in Namespace if nothing is provided
  ## Refreshes Namespace available metrics every 1h
  #[[inputs.cloudwatch.metrics]]
  #  names = ["Latency", "RequestCount"]
  #
  #  ## Statistic filters for Metric.  These are optional and allow for retrieving
  #  ## specific statistics for an individual metric.
  #  # statistic_exclude = [ "average", "sum", minimum", "maximum", sample_count" ]
  #  # statistic_include = [ "average", "sum", minimum", "maximum", sample_count" ]
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

type filteredMetric struct {
	metrics    []*cloudwatch.Metric
	statFilter filter.Filter
}

func SelectMetrics(c *CloudWatch) ([]filteredMetric, error) {
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
			statFilter, err := filter.NewIncludeExcludeFilter(m.StatisticInclude, m.StatisticExclude)
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

func (c *CloudWatch) Gather(acc telegraf.Accumulator) error {
	if c.client == nil {
		c.initializeCloudWatch()
	}

	metrics, err := SelectMetrics(c)
	if err != nil {
		return err
	}

	err = c.updateWindow(time.Now())
	if err != nil {
		return err
	}

	// get all of the possible queries so we can send groups of 100
	// note: these are cached using metricCache's specs (when only namespace is defined)
	queries, err := c.getDataQueries(metrics)
	if err != nil {
		return err
	}

	// limit concurrency or we can easily exhaust user connection limit
	// see cloudwatch API request limits:
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

	// loop through metrics and send groups of 100 at a time to gatherMetrics in order
	// to maximize the request body, thus lowering the total number of requests required.
	// 100 is the maximum number of queries a request can contain.
	groups := len(queries) / 100
	for i := 0; i < groups; i++ {
		wg.Add(1)
		<-lmtr.C
		go aggregateResults(queries[i*100 : (i+1)*100])
	}

	// gather remainder (or initial) group
	<-lmtr.C
	wg.Add(1)
	go aggregateResults(queries[(groups * 100):])

	wg.Wait()

	acc.AddError(c.aggregateMetrics(acc, results))

	return nil
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
		Filename:    c.CredentialPath,
		Token:       c.Token,
		EndpointURL: c.EndpointURL,
	}
	configProvider := credentialConfig.Credentials()

	cfg := &aws.Config{}
	loglevel := aws.LogOff
	if c.Debug {
		loglevel = aws.LogDebug
	}
	c.client = cloudwatch.New(configProvider, cfg.WithLogLevel(loglevel))
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
func (c *CloudWatch) gatherMetrics(
	params *cloudwatch.GetMetricDataInput,
) ([]*cloudwatch.MetricDataResult, error) {
	results := []*cloudwatch.MetricDataResult{}

	for {
		resp, err := c.client.GetMetricData(params)
		if err != nil {
			return nil, errors.New("Failed to get metric data - " + err.Error())
		}

		results = append(results, resp.MetricDataResults...)
		if resp.NextToken == nil {
			break
		}
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
				if len(result.Timestamps) != len(result.Values) {
					log.Println("[inputs.cloudwatch] W! MISMATCHED TIMESTAMP/VALUE LENGTH!!", len(result.Timestamps), len(result.Values))
					continue
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

/*
 * Formatting helpers
 */
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

// get all of the possible queries so we can send groups of 100
func (c *CloudWatch) getDataQueries(filteredMetrics []filteredMetric) ([]*cloudwatch.MetricDataQuery, error) {
	if c.queryCache != nil && c.metricCache != nil && c.metricCache.IsValid() {
		return c.queryCache, nil
	}

	// clear slice
	c.queries = []queryData{}

	dataQueries := []*cloudwatch.MetricDataQuery{}
	for _, filtered := range filteredMetrics {
		for i, metric := range filtered.metrics {
			id := strconv.Itoa(i)
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
		return nil, errors.New("No metrics found to collect")
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
