package stackdriver

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/iterator"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/plugins/inputs"

	// Imports the Stackdriver Monitoring client package.
	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepbduration "github.com/golang/protobuf/ptypes/duration"
	googlepbts "github.com/golang/protobuf/ptypes/timestamp"
	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
	description  = "Plugin that scrapes Google's v3 monitoring API."
	sampleConfig = `
## GCP Project (required - must be prefixed with "projects/")
project = "projects/{project_id_or_number}"

## API rate limit. On a default project, it seems that a single user can make
## ~14 requests per second. This might be configurable. Each API request can
## fetch every time series for a single metric type, though -- this is plenty
## fast for scraping all the builtin metric types (and even a handful of
## custom ones) every 60s.
# rateLimit = 14

## Collection Delay Seconds (required - must account for metrics availability via Stackdriver Monitoring API)
# delaySeconds = 60

## The first query to stackdriver queries for data points that have timestamp t
## such that: (now() - delaySeconds - lookbackSeconds) <= t <= (now() - delaySeconds).
## The subsequence queries to stackdriver query for data points that have timestamp t
## such that: lastQueryEndTime <= t <= (now() - delaySeconds).
## Note that influx will de-dedupe points that are pulled twice,
## so it's best to be safe here, just in case it takes GCP awhile
## to get around to recording the data you seek.
# lookbackSeconds = 120

## Metric collection period
interval = "1m"

## Configure the TTL for the internal cache of timeseries requests.
## Defaults to 1 hr if not specified
# cacheTTLSeconds = 3600

## Sets whether or not to scrape all bucket counts for metrics whose value
## type is "distribution". If those ~70 fields per metric
## type are annoying to you, try out the distributionAggregationAligners
## configuration option, wherein you may specifiy a list of aggregate functions
## (e.g., ALIGN_PERCENTILE_99) that might be more useful to you.
# scrapeDistributionBuckets = true

## Excluded GCP metric types. Any string prefix works.
## Only declare either this or includeMetricTypePrefixes
excludeMetricTypePrefixes = [
	"agent",
	"aws",
	"custom"
]

## *Only* include these GCP metric types. Any string prefix works
## Only declare either this or excludeMetricTypePrefixes
# includeMetricTypePrefixes = nil

## Excluded GCP metric and resource tags. Any string prefix works.
## Only declare either this or includeTagPrefixes
excludeTagPrefixes = [
	"pod_id",
]

## *Only* include these GCP metric and resource tags. Any string prefix works
## Only declare either this or excludeTagPrefixes
# includeTagPrefixes = nil

## Declares a list of aggregate functions to be used for metric types whose
## value type is "distribution". These aggregate values are recorded in the
## distribution's measurement *in addition* to the bucket counts. That is to
## say: setting this option is not mutually exclusive with
## scrapeDistributionBuckets.
distributionAggregationAligners = [
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
# [[inputs.stackdriver.filter.resourceLabels]]
#   key = "instance_name"
#   value = 'starts_with("localhost")'

## Declares metric labels to filter GCP metrics
## that match any of them.
#  [[inputs.stackdriver.filter.metricLabels]]
#  	 key = "device_name"
#  	 value = 'one_of("sda", "sdb")'
`
)

type (
	// Stackdriver is the Google Stackdriver config info.
	Stackdriver struct {
		Project                         string
		RateLimit                       int
		LookbackSeconds                 int64
		DelaySeconds                    int64
		CacheTTLSeconds                 int64
		IncludeMetricTypePrefixes       []string
		ExcludeMetricTypePrefixes       []string
		IncludeTagPrefixes              []string
		ExcludeTagPrefixes              []string
		ScrapeDistributionBuckets       bool
		DistributionAggregationAligners []string
		Filter                          *ListTimeSeriesFilter

		client              metricClient
		timeSeriesConfCache *TimeSeriesConfCache
		ctx                 context.Context
		windowStart         time.Time
		windowEnd           time.Time
	}

	// ListTimeSeriesFilter contains resource labels and metric labels
	ListTimeSeriesFilter struct {
		ResourceLabels []*Label
		MetricLabels   []*Label
	}

	// Label contains key and value
	Label struct {
		Key   string
		Value string
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
					log.Printf("D! Request %s failure: %s\n", req.String(), mdErr)
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
					log.Printf("D! Request %s failure: %s\n", req.String(), tsErr)
				}
				break
			}
			tsChan <- tsDesc
		}
	}()

	return tsChan, nil
}

// Close implements metricClient interface
func (c *stackdriverMetricClient) Close() error {
	return c.conn.Close()
}

func init() {
	f := func() telegraf.Input {
		return &Stackdriver{
			RateLimit:                       14,
			LookbackSeconds:                 120,
			DelaySeconds:                    60,
			ScrapeDistributionBuckets:       true,
			DistributionAggregationAligners: []string{},
		}
	}

	inputs.Add("stackdriver", f)
}

// Description implements telegraf.inputs interface
func (s *Stackdriver) Description() string {
	return description
}

// SampleConfig implements telegraf.inputs interface
func (s *Stackdriver) SampleConfig() string {
	return sampleConfig
}

// Gather implements telegraf.inputs interface
func (s *Stackdriver) Gather(acc telegraf.Accumulator) error {
	err := s.initializeStackdriverClient()
	if err != nil {
		log.Printf("E! Failed to create stackdriver monitoring client: %v", err)
		return err
	}

	s.updateWindow()

	tsConfs, err := s.generatetimeSeriesConfs()
	if err != nil {
		log.Printf("E! Failed to get metrics: %s\n", err)
		return err
	}

	var wg sync.WaitGroup
	lmtr := limiter.NewRateLimiter(s.RateLimit, time.Second)
	defer lmtr.Stop()

	wg.Add(len(tsConfs))
	for _, tsConf := range tsConfs {
		<-lmtr.C
		go func(tsConf *timeSeriesConf) {
			defer wg.Done()
			acc.AddError(s.scrapeTimeSeries(acc, tsConf))
		}(tsConf)
	}
	wg.Wait()

	return nil
}

func (s *Stackdriver) updateWindow() {
	windowEnd := time.Now().Add(-time.Duration(s.DelaySeconds) * time.Second)

	if s.windowEnd.IsZero() {
		s.windowStart = windowEnd.Add(-time.Duration(s.LookbackSeconds) * time.Second)
	} else {
		s.windowStart = s.windowEnd
	}
	s.windowEnd = windowEnd
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
	if s.Filter == nil {
		return ""
	}

	functions := []string{
		"starts_with",
		"ends_with",
		"has_substring",
		"one_of",
	}
	filterString := fmt.Sprintf(`metric.type = "%s"`, metricType)

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
func (s *Stackdriver) newTimeSeriesConf(metricType string) *timeSeriesConf {
	filter := s.newListTimeSeriesFilter(metricType)
	interval := &monitoringpb.TimeInterval{
		EndTime:   &googlepbts.Timestamp{Seconds: s.windowEnd.Unix()},
		StartTime: &googlepbts.Timestamp{Seconds: s.windowStart.Unix()},
	}
	tsReq := &monitoringpb.ListTimeSeriesRequest{
		Name:     s.Project,
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

func (s *Stackdriver) initializeStackdriverClient() error {
	if s.client == nil {
		s.ctx = context.Background()

		// Creates a client.
		client, err := monitoring.NewMetricClient(s.ctx)
		if err != nil {
			return err
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
	inc := s.IncludeMetricTypePrefixes
	exc := s.ExcludeMetricTypePrefixes

	return includeExcludeHelper(k, inc, exc)
}

func (s *Stackdriver) includeTag(tagKey string) bool {
	k := tagKey
	inc := s.IncludeTagPrefixes
	exc := s.ExcludeTagPrefixes

	return includeExcludeHelper(k, inc, exc)
}

// Generates filter for list metric descriptors request
func (s *Stackdriver) newListMetricDescriptorsFilters() []string {
	if len(s.IncludeMetricTypePrefixes) == 0 {
		return nil
	}

	metricTypeFilters := make([]string, len(s.IncludeMetricTypePrefixes))
	for i, metricTypePrefix := range s.IncludeMetricTypePrefixes {
		metricTypeFilters[i] = fmt.Sprintf(`metric.type = starts_with("%s")`, metricTypePrefix)
	}
	return metricTypeFilters
}

// Generate a list of timeSeriesConfig structs by making a ListMetricDescriptors
// API request and filtering the result against our configuration.
func (s *Stackdriver) generatetimeSeriesConfs() ([]*timeSeriesConf, error) {
	if s.timeSeriesConfCache != nil && s.timeSeriesConfCache.IsValid() {
		// Update interval for timeseries requests in timeseries cache
		interval := &monitoringpb.TimeInterval{
			EndTime:   &googlepbts.Timestamp{Seconds: s.windowEnd.Unix()},
			StartTime: &googlepbts.Timestamp{Seconds: s.windowStart.Unix()},
		}
		for _, timeSeriesConf := range s.timeSeriesConfCache.TimeSeriesConfs {
			timeSeriesConf.listTimeSeriesRequest.Interval = interval
		}
		return s.timeSeriesConfCache.TimeSeriesConfs, nil
	}

	ret := []*timeSeriesConf{}
	req := &monitoringpb.ListMetricDescriptorsRequest{Name: s.Project}

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
		mdRespChan, err := s.client.ListMetricDescriptors(s.ctx, req)
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
					tsConf := s.newTimeSeriesConf(metricType)
					tsConf.initForDistribution()
					ret = append(ret, tsConf)
				}
				for _, alignerStr := range s.DistributionAggregationAligners {
					tsConf := s.newTimeSeriesConf(metricType)
					tsConf.initForAggregate(alignerStr)
					ret = append(ret, tsConf)
				}
			} else {
				ret = append(ret, s.newTimeSeriesConf(metricType))
			}
		}
	}

	s.timeSeriesConfCache = &TimeSeriesConfCache{
		TimeSeriesConfs: ret,
		Generated:       time.Now(),
		TTL:             time.Duration(s.CacheTTLSeconds) * time.Second,
	}

	return ret, nil
}

// Do the work to scrape an individual time series. Runs inside a
// timeseries-specific goroutine.
func (s *Stackdriver) scrapeTimeSeries(acc telegraf.Accumulator,
	tsConf *timeSeriesConf) error {

	tsReq := tsConf.listTimeSeriesRequest
	measurement := tsConf.measurement
	fieldPrefix := tsConf.fieldPrefix
	tsRespChan, err := s.client.ListTimeSeries(s.ctx, tsReq)
	if err != nil {
		return err
	}

	for tsDesc := range tsRespChan {
		tags := map[string]string{
			"resource_type": tsDesc.Resource.Type,
		}
		for k, v := range tsDesc.Resource.Labels {
			if s.includeTag(k) {
				tags[k] = v
			}
		}
		for k, v := range tsDesc.Metric.Labels {
			if s.includeTag(k) {
				tags[k] = v
			}
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
