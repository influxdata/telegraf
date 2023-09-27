//go:generate ../../../tools/readme_config_includer/generator
package cloudwatch

import (
	"context"
	_ "embed"
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
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/limiter"
	internalMetric "github.com/influxdata/telegraf/metric"
	internalaws "github.com/influxdata/telegraf/plugins/common/aws"
	internalProxy "github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// CloudWatch contains the configuration and cache for the cloudwatch plugin.
type CloudWatch struct {
	StatisticExclude []string        `toml:"statistic_exclude"`
	StatisticInclude []string        `toml:"statistic_include"`
	Timeout          config.Duration `toml:"timeout"`

	internalProxy.HTTPProxy

	Period                config.Duration `toml:"period"`
	Delay                 config.Duration `toml:"delay"`
	Namespace             string          `toml:"namespace" deprecated:"1.25.0;use 'namespaces' instead"`
	Namespaces            []string        `toml:"namespaces"`
	Metrics               []*Metric       `toml:"metrics"`
	CacheTTL              config.Duration `toml:"cache_ttl"`
	RateLimit             int             `toml:"ratelimit"`
	RecentlyActive        string          `toml:"recently_active"`
	BatchSize             int             `toml:"batch_size"`
	IncludeLinkedAccounts bool            `toml:"include_linked_accounts"`

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

func (*CloudWatch) SampleConfig() string {
	return sampleConfig
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
		var batches [][]types.MetricDataQuery

		for c.BatchSize < len(namespacedQueries) {
			namespacedQueries, batches = namespacedQueries[c.BatchSize:], append(batches, namespacedQueries[0:c.BatchSize:c.BatchSize])
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

	awsCreds, err := c.CredentialConfig.Credentials()
	if err != nil {
		return err
	}

	var customResolver cwClient.EndpointResolver
	if c.CredentialConfig.EndpointURL != "" && c.CredentialConfig.Region != "" {
		customResolver = cwClient.EndpointResolverFromURL(c.CredentialConfig.EndpointURL)
	}

	c.client = cwClient.NewFromConfig(awsCreds, func(options *cwClient.Options) {
		if customResolver != nil {
			options.EndpointResolver = customResolver
		}

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
	accounts   []string
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
			var accounts []string
			if !hasWildcard(m.Dimensions) {
				dimensions := make([]types.Dimension, 0, len(m.Dimensions))
				for _, d := range m.Dimensions {
					dimensions = append(dimensions, types.Dimension{
						Name:  aws.String(d.Name),
						Value: aws.String(d.Value),
					})
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
				allMetrics, allAccounts := c.fetchNamespaceMetrics()

				for _, name := range m.MetricNames {
					for i, metric := range allMetrics {
						if isSelected(name, metric, m.Dimensions) {
							for _, namespace := range c.Namespaces {
								metrics = append(metrics, types.Metric{
									Namespace:  aws.String(namespace),
									MetricName: aws.String(name),
									Dimensions: metric.Dimensions,
								})
							}
							if c.IncludeLinkedAccounts {
								accounts = append(accounts, allAccounts[i])
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
				accounts:   accounts,
			})
		}
	} else {
		metrics, accounts := c.fetchNamespaceMetrics()
		fMetrics = []filteredMetric{
			{
				metrics:    metrics,
				statFilter: c.statFilter,
				accounts:   accounts,
			},
		}
	}
	c.metricCache = &metricCache{
		metrics: fMetrics,
		built:   time.Now(),
		ttl:     time.Duration(c.CacheTTL),
	}
	return fMetrics, nil
}

// fetchNamespaceMetrics retrieves available metrics for a given CloudWatch namespace.
func (c *CloudWatch) fetchNamespaceMetrics() ([]types.Metric, []string) {
	metrics := []types.Metric{}
	var accounts []string
	for _, namespace := range c.Namespaces {
		params := &cwClient.ListMetricsInput{
			Dimensions:            []types.DimensionFilter{},
			Namespace:             aws.String(namespace),
			IncludeLinkedAccounts: c.IncludeLinkedAccounts,
		}
		if c.RecentlyActive == "PT3H" {
			params.RecentlyActive = types.RecentlyActivePt3h
		}

		for {
			resp, err := c.client.ListMetrics(context.Background(), params)
			if err != nil {
				c.Log.Errorf("failed to list metrics with namespace %s: %v", namespace, err)
				// skip problem namespace on error and continue to next namespace
				break
			}
			metrics = append(metrics, resp.Metrics...)
			accounts = append(accounts, resp.OwningAccounts...)

			if resp.NextToken == nil {
				break
			}
			params.NextToken = resp.NextToken
		}
	}
	return metrics, accounts
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
			var accountID *string
			if c.IncludeLinkedAccounts {
				accountID = aws.String(filtered.accounts[j])
				(*dimension)["account"] = filtered.accounts[j]
			}

			statisticTypes := map[string]string{
				"average":      "Average",
				"maximum":      "Maximum",
				"minimum":      "Minimum",
				"sum":          "Sum",
				"sample_count": "SampleCount",
			}

			for statisticType, statistic := range statisticTypes {
				if !filtered.statFilter.Match(statisticType) {
					continue
				}
				queryID := statisticType + "_" + id
				c.queryDimensions[queryID] = dimension
				dataQueries[*metric.Namespace] = append(dataQueries[*metric.Namespace], types.MetricDataQuery{
					Id:        aws.String(queryID),
					AccountId: accountID,
					Label:     aws.String(snakeCase(*metric.MetricName + "_" + statisticType)),
					MetricStat: &types.MetricStat{
						Metric: &filtered.metrics[j],
						Period: aws.Int32(int32(time.Duration(c.Period).Seconds())),
						Stat:   aws.String(statistic),
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
			return nil, fmt.Errorf("failed to get metric data: %w", err)
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
				grouper.Add(namespace, tags, result.Timestamps[i], *result.Label, result.Values[i])
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
		BatchSize: 500,
	}
}

func sanitizeMeasurement(namespace string) string {
	namespace = strings.ReplaceAll(namespace, "/", "_")
	namespace = snakeCase(namespace)
	return "cloudwatch_" + namespace
}

func snakeCase(s string) string {
	s = internal.SnakeCase(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "__", "_")
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
