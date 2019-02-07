package stackdriver

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepbduration "github.com/golang/protobuf/ptypes/duration"
	googlepbts "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/plugins/inputs" // Imports the Stackdriver Monitoring client package.
	"google.golang.org/api/iterator"
	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
	description  = "Gather timeseries from Google Cloud Platform v3 monitoring API"
	sampleConfig = `
  ## GCP Project
  project = "erudite-bloom-151019"

  ## Select all timeseries that start with the given metric type.
  metric_type_prefix_include = [
    "custom.googleapis.com/",
    "compute.googleapis.com/",
  ]

  ## Exclude timeseries that start with the given metric type.
  metric_type_prefix_exclude = []

  ## API rate limit. On a default project, it seems that a single user can make
  ## ~14 requests per second. This might be configurable. Each API request can
  ## fetch every time series for a single metric type, though -- this is plenty
  ## fast for scraping all the builtin metric types (and even a handful of
  ## custom ones) every 60s.
  # rate_limit = 14

  ## Collection delay; if set too low metrics may not yet be available.
  # delay = "5m"

  ## The first query to stackdriver queries for data points that have timestamp t
  ## such that: (now() - delaySeconds - lookbackSeconds) <= t <= (now() - delaySeconds).
  ## The subsequence queries to stackdriver query for data points that have timestamp t
  ## such that: lastQueryEndTime <= t <= (now() - delaySeconds).
  ## Note that influx will de-dedupe points that are pulled twice,
  ## so it's best to be safe here, just in case it takes GCP awhile
  ## to get around to recording the data you seek.
  ## Collection window size.
  ##
  ## Along with the delay option, controls the number of points selected on
  ## each gather.  When set, metrics are gathered between:
  ##   now() - delay and now() - delay - window.
  ##
  ## If unset, the window will start at 1m and be set dynamically to span the time
  ## between calls; which will be approximately the length of the interval.
  # window = "1m"

  ## Override data collection interval; recommended to set to 1m or larger.
  interval = "1m"

  ## Configure the TTL for the internal cache of timeseries requests.
  # cache_ttl = "1h"

  ## Sets whether or not to scrape all bucket counts for metrics whose value
  ## type is "distribution". If those ~70 fields per metric
  ## type are annoying to you, try out the distributionAggregationAligners
  ## configuration option, wherein you may specifiy a list of aggregate functions
  ## (e.g., ALIGN_PERCENTILE_99) that might be more useful to you.
  # scrape_distribution_buckets = true

  ## Declares a list of aggregate functions to be used for metric types whose
  ## value type is "distribution". These aggregate values are recorded in the
  ## distribution's measurement *in addition* to the bucket counts. That is to
  ## say: setting this option is not mutually exclusive with
  ## scrapeDistributionBuckets.
  distribution_aggregation_aligners = [
  	"ALIGN_PERCENTILE_99",
  	"ALIGN_PERCENTILE_95",
  	"ALIGN_PERCENTILE_50",
  ]

  ## The filter string consists of logical AND of the
  ## resource labels and metric labels if both of them
  ## are specified. (optional)
  ## See: https://cloud.google.com/monitoring/api/v3/filters
  ## Declares resource labels to filter GCP metrics
  ## that match any of them.
  # [[inputs.stackdriver.filter.resource_labels]]
  #   key = "instance_name"
  #   value = 'starts_with("localhost")'

  ## Declares metric labels to filter GCP metrics
  ## that match any of them.
  #  [[inputs.stackdriver.filter.metric_labels]]
  #  	 key = "device_name"
  #  	 value = 'one_of("sda", "sdb")'
`
)

var (
	defaultWindow = internal.Duration{Duration: 1 * time.Minute}
	defaultDelay  = internal.Duration{Duration: 5 * time.Minute}
)

type (
	// Stackdriver is the Google Stackdriver config info.
	Stackdriver struct {
		Project                         string                `toml:"project"`
		RateLimit                       int                   `toml:"rate_limit"`
		Window                          internal.Duration     `toml:"window"`
		Delay                           internal.Duration     `toml:"delay"`
		CacheTTL                        internal.Duration     `toml:"cache_ttl"`
		MetricTypePrefixInclude         []string              `toml:"metric_type_prefix_include"`
		MetricTypePrefixExclude         []string              `toml:"metric_type_prefix_exclude"`
		ScrapeDistributionBuckets       bool                  `toml:"scrape_distribution_buckets"`
		DistributionAggregationAligners []string              `toml:"distribution_aggregation_aligners"`
		Filter                          *ListTimeSeriesFilter `toml:"filter"`

		client              metricClient
		timeSeriesConfCache *TimeSeriesConfCache
		prevEnd             time.Time
	}

	// ListTimeSeriesFilter contains resource labels and metric labels
	ListTimeSeriesFilter struct {
		ResourceLabels []*Label `json:"resource_labels"`
		MetricLabels   []*Label `json:"metric_labels"`
	}

	// Label contains key and value
	Label struct {
		Key   string `toml:"key"`
		Value string `toml:"value"`
	}

	// TimeSeriesConfCache caches generated timeseries configurations
	TimeSeriesConfCache struct {
		TTL             time.Duration
		Generated       time.Time
		TimeSeriesConfs []*timeSeriesConf
	}

	// stackdriverMetricClient is a metric client for stackdriver
	stackdriverMetricClient struct {
		conn *monitoring.MetricClient
	}

	// metricClient is convenient for testing
	metricClient interface {
		ListMetricDescriptors(ctx context.Context, req *monitoringpb.ListMetricDescriptorsRequest) (<-chan *metricpb.MetricDescriptor, error)
		ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) (<-chan *monitoringpb.TimeSeries, error)
		Close() error
	}
)

// ListMetricDescriptors implements metricClient interface
func (c *stackdriverMetricClient) ListMetricDescriptors(
	ctx context.Context,
	req *monitoringpb.ListMetricDescriptorsRequest,
) (<-chan *metricpb.MetricDescriptor, error) {
	mdChan := make(chan *metricpb.MetricDescriptor, 1000)

	go func() {
		// Channel must be closed for safety
		defer close(mdChan)

		// Iterate over metric descriptors and send them to buffered channel
		mdResp := c.conn.ListMetricDescriptors(ctx, req)
		for {
			mdDesc, mdErr := mdResp.Next()
			if mdErr != nil {
				if mdErr != iterator.Done {
					log.Printf("E! Request %s failure: %s\n", req.String(), mdErr)
				}
				break
			}
			mdChan <- mdDesc
		}
	}()

	return mdChan, nil
}

// ListTimeSeries implements metricClient interface
func (c *stackdriverMetricClient) ListTimeSeries(
	ctx context.Context,
	req *monitoringpb.ListTimeSeriesRequest,
) (<-chan *monitoringpb.TimeSeries, error) {
	tsChan := make(chan *monitoringpb.TimeSeries, 1000)

	go func() {
		// Channel must be closed for safety
		defer close(tsChan)

		// Iterate over timeseries and send them to buffered channel
		tsResp := c.conn.ListTimeSeries(ctx, req)
		for {
			tsDesc, tsErr := tsResp.Next()
			if tsErr != nil {
				if tsErr != iterator.Done {
					log.Printf("E! Request %s failure: %s\n", req.String(), tsErr)
				}
				break
			}
			tsChan <- tsDesc
		}
	}()

	return tsChan, nil
}

// Close implements metricClient interface
func (s *stackdriverMetricClient) Close() error {
	return s.conn.Close()
}

// Description implements telegraf.Input interface
func (s *Stackdriver) Description() string {
	return description
}

// SampleConfig implements telegraf.Input interface
func (s *Stackdriver) SampleConfig() string {
	return sampleConfig
}

// Gather implements telegraf.Input interface
func (s *Stackdriver) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	err := s.initializeStackdriverClient(ctx)
	if err != nil {
		return err
	}

	start, end := s.updateWindow(s.prevEnd)
	s.prevEnd = end

	tsConfs, err := s.generatetimeSeriesConfs(ctx, start, end)
	if err != nil {
		return err
	}

	lmtr := limiter.NewRateLimiter(s.RateLimit, time.Second)
	defer lmtr.Stop()

	var wg sync.WaitGroup
	wg.Add(len(tsConfs))
	for _, tsConf := range tsConfs {
		<-lmtr.C
		go func(tsConf *timeSeriesConf) {
			defer wg.Done()
			acc.AddError(s.scrapeTimeSeries(ctx, acc, tsConf))
		}(tsConf)
	}
	wg.Wait()

	return nil
}

func (s *Stackdriver) updateWindow(prevEnd time.Time) (time.Time, time.Time) {
	var start time.Time
	if s.Window.Duration != 0 {
		start = time.Now().Add(-s.Delay.Duration).Add(-s.Window.Duration)
	} else if prevEnd.IsZero() {
		start = time.Now().Add(-s.Delay.Duration).Add(-defaultWindow.Duration)
	} else {
		start = prevEnd
	}
	end := time.Now().Add(-s.Delay.Duration)
	return start, end
}

// Internal structure which holds our configuration for a particular GCP time
// series.
type timeSeriesConf struct {
	// The influx measurement name this time series maps to
	measurement string
	// The prefix to use before any influx field names that we'll write for
	// this time series. (Or, if we only decide to write one field name, this
	// field just holds the value of the field name.)
	fieldPrefix string
	// The GCP API request that we'll use to fetch data for this time series.
	listTimeSeriesRequest *monitoringpb.ListTimeSeriesRequest
}

// Generate filter string for ListTimeSeriesRequest
func (s *Stackdriver) newListTimeSeriesFilter(metricType string) string {
	functions := []string{
		"starts_with",
		"ends_with",
		"has_substring",
		"one_of",
	}
	filterString := fmt.Sprintf(`metric.type = "%s"`, metricType)
	if s.Filter == nil {
		return filterString
	}

	var valueFmt string
	if len(s.Filter.ResourceLabels) > 0 {
		resourceLabelsFilter := make([]string, len(s.Filter.ResourceLabels))
		for i, resourceLabel := range s.Filter.ResourceLabels {
			// check if resource label value contains function
			if includeExcludeHelper(resourceLabel.Value, functions, nil) {
				valueFmt = `resource.labels.%s = %s`
			} else {
				valueFmt = `resource.labels.%s = "%s"`
			}
			resourceLabelsFilter[i] = fmt.Sprintf(valueFmt, resourceLabel.Key, resourceLabel.Value)
		}
		if len(resourceLabelsFilter) == 1 {
			filterString += fmt.Sprintf(" AND %s", resourceLabelsFilter[0])
		} else {
			filterString += fmt.Sprintf(" AND (%s)", strings.Join(resourceLabelsFilter, " OR "))
		}
	}

	if len(s.Filter.MetricLabels) > 0 {
		metricLabelsFilter := make([]string, len(s.Filter.MetricLabels))
		for i, metricLabel := range s.Filter.MetricLabels {
			// check if metric label value contains function
			if includeExcludeHelper(metricLabel.Value, functions, nil) {
				valueFmt = `metric.labels.%s = %s`
			} else {
				valueFmt = `metric.labels.%s = "%s"`
			}
			metricLabelsFilter[i] = fmt.Sprintf(valueFmt, metricLabel.Key, metricLabel.Value)
		}
		if len(metricLabelsFilter) == 1 {
			filterString += fmt.Sprintf(" AND %s", metricLabelsFilter[0])
		} else {
			filterString += fmt.Sprintf(" AND (%s)", strings.Join(metricLabelsFilter, " OR "))
		}
	}

	return filterString
}

// Create and initialize a timeSeriesConf for a given GCP metric type with
// defaults taken from the gcp_stackdriver plugin configuration.
func (s *Stackdriver) newTimeSeriesConf(metricType string, startTime, endTime time.Time) *timeSeriesConf {
	filter := s.newListTimeSeriesFilter(metricType)
	interval := &monitoringpb.TimeInterval{
		EndTime:   &googlepbts.Timestamp{Seconds: endTime.Unix()},
		StartTime: &googlepbts.Timestamp{Seconds: startTime.Unix()},
	}
	tsReq := &monitoringpb.ListTimeSeriesRequest{
		Name:     monitoring.MetricProjectPath(s.Project),
		Filter:   filter,
		Interval: interval,
	}
	cfg := &timeSeriesConf{
		measurement:           metricType,
		fieldPrefix:           "value",
		listTimeSeriesRequest: tsReq,
	}

	// GCP metric types have at least one slash, but we'll be defensive anyway.
	slashIdx := strings.LastIndex(metricType, "/")
	if slashIdx > 0 {
		cfg.measurement = metricType[:slashIdx]
		cfg.fieldPrefix = metricType[slashIdx+1:]
	}

	return cfg
}

// Change this configuration to query an aggregate by specifying an "aligner".
// In GCP monitoring, "aligning" is aggregation performed *within* a time
// series, to distill a pile of data points down to a single data point for
// some given time period (here, we specify 60s as our time period). This is
// especially useful for scraping GCP "distribution" metric types, whose raw
// data amounts to a ~60 bucket histogram, which is fairly hard to query and
// visualize in the TICK stack.
func (t *timeSeriesConf) initForAggregate(alignerStr string) {
	// Check if alignerStr is valid
	alignerInt, isValid := monitoringpb.Aggregation_Aligner_value[alignerStr]
	if !isValid {
		alignerStr = monitoringpb.Aggregation_Aligner_name[alignerInt]
	}
	aligner := monitoringpb.Aggregation_Aligner(alignerInt)
	agg := &monitoringpb.Aggregation{
		AlignmentPeriod:  &googlepbduration.Duration{Seconds: 60},
		PerSeriesAligner: aligner,
	}
	t.fieldPrefix = t.fieldPrefix + "_" + strings.ToLower(alignerStr) + "_"
	t.listTimeSeriesRequest.Aggregation = agg
}

// Change this configuration to query a distribution type. Time series of type
// distribution will generate a lot of fields (one for each bucket).
func (t *timeSeriesConf) initForDistribution() {
	t.fieldPrefix = t.fieldPrefix + "_"
}

// IsValid checks timeseriesconf cache validity
func (c *TimeSeriesConfCache) IsValid() bool {
	return c.TimeSeriesConfs != nil && time.Since(c.Generated) < c.TTL
}

func (s *Stackdriver) initializeStackdriverClient(ctx context.Context) error {
	if s.client == nil {
		client, err := monitoring.NewMetricClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create stackdriver monitoring client: %v", err)
		}
		s.client = &stackdriverMetricClient{conn: client}
	}

	return nil
}

func includeExcludeHelper(key string, includes []string, excludes []string) bool {
	if len(includes) > 0 {
		for _, includeStr := range includes {
			if strings.HasPrefix(key, includeStr) {
				return true
			}
		}
		return false
	}
	if len(excludes) > 0 {
		for _, excludeStr := range excludes {
			if strings.HasPrefix(key, excludeStr) {
				return false
			}
		}
		return true
	}
	return true
}

// Test whether a particular GCP metric type should be scraped by this plugin
// by checking the plugin name against the configuration's
// "includeMetricTypePrefixes" and "excludeMetricTypePrefixes"
func (s *Stackdriver) includeMetricType(metricType string) bool {
	k := metricType
	inc := s.MetricTypePrefixInclude
	exc := s.MetricTypePrefixExclude

	return includeExcludeHelper(k, inc, exc)
}

// Generates filter for list metric descriptors request
func (s *Stackdriver) newListMetricDescriptorsFilters() []string {
	if len(s.MetricTypePrefixInclude) == 0 {
		return nil
	}

	metricTypeFilters := make([]string, len(s.MetricTypePrefixInclude))
	for i, metricTypePrefix := range s.MetricTypePrefixInclude {
		metricTypeFilters[i] = fmt.Sprintf(`metric.type = starts_with("%s")`, metricTypePrefix)
	}
	return metricTypeFilters
}

// Generate a list of timeSeriesConfig structs by making a ListMetricDescriptors
// API request and filtering the result against our configuration.
func (s *Stackdriver) generatetimeSeriesConfs(ctx context.Context, startTime, endTime time.Time) ([]*timeSeriesConf, error) {
	if s.timeSeriesConfCache != nil && s.timeSeriesConfCache.IsValid() {
		// Update interval for timeseries requests in timeseries cache
		interval := &monitoringpb.TimeInterval{
			EndTime:   &googlepbts.Timestamp{Seconds: endTime.Unix()},
			StartTime: &googlepbts.Timestamp{Seconds: startTime.Unix()},
		}
		for _, timeSeriesConf := range s.timeSeriesConfCache.TimeSeriesConfs {
			timeSeriesConf.listTimeSeriesRequest.Interval = interval
		}
		return s.timeSeriesConfCache.TimeSeriesConfs, nil
	}

	ret := []*timeSeriesConf{}
	req := &monitoringpb.ListMetricDescriptorsRequest{Name: monitoring.MetricProjectPath(s.Project)}

	filters := s.newListMetricDescriptorsFilters()
	if len(filters) == 0 {
		filters = []string{""}
	}

	for _, filter := range filters {
		// Add filter for list metric descriptors if
		// includeMetricTypePrefixes is specified,
		// this is more effecient than iterating over
		// all metric descriptors
		req.Filter = filter
		mdRespChan, err := s.client.ListMetricDescriptors(ctx, req)
		if err != nil {
			return nil, err
		}

		for metricDescriptor := range mdRespChan {
			metricType := metricDescriptor.Type
			valueType := metricDescriptor.ValueType

			if filter == "" && !s.includeMetricType(metricType) {
				continue
			}

			if valueType == metricpb.MetricDescriptor_DISTRIBUTION {
				if s.ScrapeDistributionBuckets {
					tsConf := s.newTimeSeriesConf(metricType, startTime, endTime)
					tsConf.initForDistribution()
					ret = append(ret, tsConf)
				}
				for _, alignerStr := range s.DistributionAggregationAligners {
					tsConf := s.newTimeSeriesConf(metricType, startTime, endTime)
					tsConf.initForAggregate(alignerStr)
					ret = append(ret, tsConf)
				}
			} else {
				ret = append(ret, s.newTimeSeriesConf(metricType, startTime, endTime))
			}
		}
	}

	s.timeSeriesConfCache = &TimeSeriesConfCache{
		TimeSeriesConfs: ret,
		Generated:       time.Now(),
		TTL:             s.CacheTTL.Duration,
	}

	return ret, nil
}

// Do the work to scrape an individual time series. Runs inside a
// timeseries-specific goroutine.
func (s *Stackdriver) scrapeTimeSeries(ctx context.Context, acc telegraf.Accumulator,
	tsConf *timeSeriesConf) error {

	tsReq := tsConf.listTimeSeriesRequest
	measurement := tsConf.measurement
	fieldPrefix := tsConf.fieldPrefix
	tsRespChan, err := s.client.ListTimeSeries(ctx, tsReq)
	if err != nil {
		return err
	}

	for tsDesc := range tsRespChan {
		tags := map[string]string{
			"resource_type": tsDesc.Resource.Type,
		}
		for k, v := range tsDesc.Resource.Labels {
			tags[k] = v
		}
		for k, v := range tsDesc.Metric.Labels {
			tags[k] = v
		}
		for _, p := range tsDesc.Points {
			ts := time.Unix(p.Interval.EndTime.Seconds, 0)

			var fields map[string]interface{}

			if tsDesc.ValueType == metricpb.MetricDescriptor_DISTRIBUTION {
				val := p.Value.GetDistributionValue()
				fields = s.scrapeDistribution(fieldPrefix, val)
			} else {
				var field interface{}

				// Types that are valid to be assigned to Value
				// See: https://godoc.org/google.golang.org/genproto/googleapis/monitoring/v3#TypedValue
				switch tsDesc.ValueType {
				case metricpb.MetricDescriptor_BOOL:
					field = p.Value.GetBoolValue()
				case metricpb.MetricDescriptor_INT64:
					field = p.Value.GetInt64Value()
				case metricpb.MetricDescriptor_DOUBLE:
					field = p.Value.GetDoubleValue()
				case metricpb.MetricDescriptor_STRING:
					field = p.Value.GetStringValue()
				}

				fields = map[string]interface{}{
					fieldPrefix: field,
				}
			}

			acc.AddFields(measurement, fields, tags, ts)
		}
	}

	return nil
}

// Write out fields for a distribution metric. In this function, we interpret
// the width of each bucket and write out a special field name for each bucket
// that encodes the bucket's lower bounds. For example, the bucket named
// "bucket_ge_1.5" includes a count of points in the distribution that were
// greater than the value 1.5. To find the upper bound, simply look at other
// buckets. If the next bucket in the order is, for example, "bucket_ge_2.6",
// that means bucket_ge_1.5 holds points that are strictly less than the value
// 2.6.
func (s *Stackdriver) scrapeDistribution(
	fieldPrefix string,
	metric *distributionpb.Distribution) map[string]interface{} {

	fields := map[string]interface{}{}

	fields[fieldPrefix+"count"] = metric.Count
	fields[fieldPrefix+"mean"] = metric.Mean
	fields[fieldPrefix+"sum_of_squared_deviation"] = metric.SumOfSquaredDeviation

	if metric.Range != nil {
		fields[fieldPrefix+"range_min"] = metric.Range.Min
		fields[fieldPrefix+"range_max"] = metric.Range.Max
	}

	if metric.Count > 0 {
		fields[fieldPrefix+"bucket_underflow"] = metric.BucketCounts[0]
	}

	linearBuckets := metric.BucketOptions.GetLinearBuckets()
	exponentialBuckets := metric.BucketOptions.GetExponentialBuckets()
	explicitBuckets := metric.BucketOptions.GetExplicitBuckets()

	var numBuckets int32
	if linearBuckets != nil {
		numBuckets = linearBuckets.NumFiniteBuckets
	} else if exponentialBuckets != nil {
		numBuckets = exponentialBuckets.NumFiniteBuckets
	} else {
		numBuckets = int32(len(explicitBuckets.Bounds)) + 1
	}

	var i int32
	for i = 1; i < numBuckets; i++ {
		var num float64
		if linearBuckets != nil {
			bucketWidth := linearBuckets.Width * float64(i-1)
			num = linearBuckets.Offset + bucketWidth
		} else if exponentialBuckets != nil {
			width := math.Pow(exponentialBuckets.GrowthFactor, float64(i-1))
			num = exponentialBuckets.Scale * width
		} else if explicitBuckets != nil {
			num = explicitBuckets.Bounds[i-1]
		}

		col := fmt.Sprintf("%sbucket_ge_%.3f", fieldPrefix, num)
		if i < int32(len(metric.BucketCounts)) {
			fields[col] = metric.BucketCounts[i]
		} else {
			fields[col] = 0
		}
	}

	return fields
}

func init() {
	f := func() telegraf.Input {
		return &Stackdriver{
			RateLimit:                       14,
			Delay:                           defaultDelay,
			ScrapeDistributionBuckets:       true,
			DistributionAggregationAligners: []string{},
		}
	}

	inputs.Add("stackdriver", f)
}
