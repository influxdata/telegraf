package stackdriver

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"google.golang.org/api/iterator"
	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs" // Imports the Stackdriver Monitoring client package.
	"github.com/influxdata/telegraf/selfstat"
)

const (
	defaultRateLimit = 14
)

var (
	defaultCacheTTL = config.Duration(1 * time.Hour)
	defaultWindow   = config.Duration(1 * time.Minute)
	defaultDelay    = config.Duration(5 * time.Minute)
)

type (
	// Stackdriver is the Google Stackdriver config info.
	Stackdriver struct {
		Project                         string                `toml:"project"`
		RateLimit                       int                   `toml:"rate_limit"`
		Window                          config.Duration       `toml:"window"`
		Delay                           config.Duration       `toml:"delay"`
		CacheTTL                        config.Duration       `toml:"cache_ttl"`
		MetricTypePrefixInclude         []string              `toml:"metric_type_prefix_include"`
		MetricTypePrefixExclude         []string              `toml:"metric_type_prefix_exclude"`
		GatherRawDistributionBuckets    bool                  `toml:"gather_raw_distribution_buckets"`
		DistributionAggregationAligners []string              `toml:"distribution_aggregation_aligners"`
		Filter                          *ListTimeSeriesFilter `toml:"filter"`

		Log telegraf.Logger

		client              metricClient
		timeSeriesConfCache *timeSeriesConfCache
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
	timeSeriesConfCache struct {
		TTL             time.Duration
		Generated       time.Time
		TimeSeriesConfs []*timeSeriesConf
	}

	// Internal structure which holds our configuration for a particular GCP time
	// series.
	timeSeriesConf struct {
		// The influx measurement name this time series maps to
		measurement string
		// The prefix to use before any influx field names that we'll write for
		// this time series. (Or, if we only decide to write one field name, this
		// field just holds the value of the field name.)
		fieldKey string
		// The GCP API request that we'll use to fetch data for this time series.
		listTimeSeriesRequest *monitoringpb.ListTimeSeriesRequest
	}

	// stackdriverMetricClient is a metric client for stackdriver
	stackdriverMetricClient struct {
		log  telegraf.Logger
		conn *monitoring.MetricClient

		listMetricDescriptorsCalls selfstat.Stat
		listTimeSeriesCalls        selfstat.Stat
	}

	// metricClient is convenient for testing
	metricClient interface {
		ListMetricDescriptors(ctx context.Context, req *monitoringpb.ListMetricDescriptorsRequest) (<-chan *metricpb.MetricDescriptor, error)
		ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) (<-chan *monitoringpb.TimeSeries, error)
		Close() error
	}

	lockedSeriesGrouper struct {
		sync.Mutex
		*metric.SeriesGrouper
	}
)

func (g *lockedSeriesGrouper) Add(
	measurement string,
	tags map[string]string,
	tm time.Time,
	field string,
	fieldValue interface{},
) error {
	g.Lock()
	defer g.Unlock()
	return g.SeriesGrouper.Add(measurement, tags, tm, field, fieldValue)
}

// ListMetricDescriptors implements metricClient interface
func (smc *stackdriverMetricClient) ListMetricDescriptors(
	ctx context.Context,
	req *monitoringpb.ListMetricDescriptorsRequest,
) (<-chan *metricpb.MetricDescriptor, error) {
	mdChan := make(chan *metricpb.MetricDescriptor, 1000)

	go func() {
		smc.log.Debugf("List metric descriptor request filter: %s", req.Filter)
		defer close(mdChan)

		// Iterate over metric descriptors and send them to buffered channel
		mdResp := smc.conn.ListMetricDescriptors(ctx, req)
		smc.listMetricDescriptorsCalls.Incr(1)
		for {
			mdDesc, mdErr := mdResp.Next()
			if mdErr != nil {
				if mdErr != iterator.Done {
					smc.log.Errorf("Failed iterating metric descriptor responses: %q: %v", req.String(), mdErr)
				}
				break
			}
			mdChan <- mdDesc
		}
	}()

	return mdChan, nil
}

// ListTimeSeries implements metricClient interface
func (smc *stackdriverMetricClient) ListTimeSeries(
	ctx context.Context,
	req *monitoringpb.ListTimeSeriesRequest,
) (<-chan *monitoringpb.TimeSeries, error) {
	tsChan := make(chan *monitoringpb.TimeSeries, 1000)

	go func() {
		smc.log.Debugf("List time series request filter: %s", req.Filter)
		defer close(tsChan)

		// Iterate over timeseries and send them to buffered channel
		tsResp := smc.conn.ListTimeSeries(ctx, req)
		smc.listTimeSeriesCalls.Incr(1)
		for {
			tsDesc, tsErr := tsResp.Next()
			if tsErr != nil {
				if tsErr != iterator.Done {
					smc.log.Errorf("Failed iterating time series responses: %q: %v", req.String(), tsErr)
				}
				break
			}
			tsChan <- tsDesc
		}
	}()

	return tsChan, nil
}

// Close implements metricClient interface
func (smc *stackdriverMetricClient) Close() error {
	return smc.conn.Close()
}

// Gather implements telegraf.Input interface
func (s *Stackdriver) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	if s.RateLimit == 0 {
		s.RateLimit = defaultRateLimit
	}

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

	grouper := &lockedSeriesGrouper{
		SeriesGrouper: metric.NewSeriesGrouper(),
	}

	var wg sync.WaitGroup
	wg.Add(len(tsConfs))
	for _, tsConf := range tsConfs {
		<-lmtr.C
		go func(tsConf *timeSeriesConf) {
			defer wg.Done()
			acc.AddError(s.gatherTimeSeries(ctx, grouper, tsConf))
		}(tsConf)
	}
	wg.Wait()

	for _, groupedMetric := range grouper.Metrics() {
		acc.AddMetric(groupedMetric)
	}

	return nil
}

// Returns the start and end time for the next collection.
func (s *Stackdriver) updateWindow(prevEnd time.Time) (time.Time, time.Time) {
	var start time.Time
	if time.Duration(s.Window) != 0 {
		start = time.Now().Add(-time.Duration(s.Delay)).Add(-time.Duration(s.Window))
	} else if prevEnd.IsZero() {
		start = time.Now().Add(-time.Duration(s.Delay)).Add(-time.Duration(defaultWindow))
	} else {
		start = prevEnd
	}
	end := time.Now().Add(-time.Duration(s.Delay))
	return start, end
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
func (s *Stackdriver) newTimeSeriesConf(
	metricType string, startTime, endTime time.Time,
) *timeSeriesConf {
	filter := s.newListTimeSeriesFilter(metricType)
	interval := &monitoringpb.TimeInterval{
		EndTime:   &timestamppb.Timestamp{Seconds: endTime.Unix()},
		StartTime: &timestamppb.Timestamp{Seconds: startTime.Unix()},
	}
	tsReq := &monitoringpb.ListTimeSeriesRequest{
		Name:     fmt.Sprintf("projects/%s", s.Project),
		Filter:   filter,
		Interval: interval,
	}
	cfg := &timeSeriesConf{
		measurement:           metricType,
		fieldKey:              "value",
		listTimeSeriesRequest: tsReq,
	}

	// GCP metric types have at least one slash, but we'll be defensive anyway.
	slashIdx := strings.LastIndex(metricType, "/")
	if slashIdx > 0 {
		cfg.measurement = metricType[:slashIdx]
		cfg.fieldKey = metricType[slashIdx+1:]
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
		AlignmentPeriod:  &durationpb.Duration{Seconds: 60},
		PerSeriesAligner: aligner,
	}
	t.fieldKey = t.fieldKey + "_" + strings.ToLower(alignerStr)
	t.listTimeSeriesRequest.Aggregation = agg
}

// IsValid checks timeseriesconf cache validity
func (c *timeSeriesConfCache) IsValid() bool {
	return c.TimeSeriesConfs != nil && time.Since(c.Generated) < c.TTL
}

func (s *Stackdriver) initializeStackdriverClient(ctx context.Context) error {
	if s.client == nil {
		client, err := monitoring.NewMetricClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create stackdriver monitoring client: %v", err)
		}

		tags := map[string]string{
			"project_id": s.Project,
		}
		listMetricDescriptorsCalls := selfstat.Register(
			"stackdriver", "list_metric_descriptors_calls", tags)
		listTimeSeriesCalls := selfstat.Register(
			"stackdriver", "list_timeseries_calls", tags)

		s.client = &stackdriverMetricClient{
			log:                        s.Log,
			conn:                       client,
			listMetricDescriptorsCalls: listMetricDescriptorsCalls,
			listTimeSeriesCalls:        listTimeSeriesCalls,
		}
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
		metricTypeFilters[i] = fmt.Sprintf(`metric.type = starts_with(%q)`, metricTypePrefix)
	}
	return metricTypeFilters
}

// Generate a list of timeSeriesConfig structs by making a ListMetricDescriptors
// API request and filtering the result against our configuration.
func (s *Stackdriver) generatetimeSeriesConfs(
	ctx context.Context, startTime, endTime time.Time,
) ([]*timeSeriesConf, error) {
	if s.timeSeriesConfCache != nil && s.timeSeriesConfCache.IsValid() {
		// Update interval for timeseries requests in timeseries cache
		interval := &monitoringpb.TimeInterval{
			EndTime:   &timestamppb.Timestamp{Seconds: endTime.Unix()},
			StartTime: &timestamppb.Timestamp{Seconds: startTime.Unix()},
		}
		for _, timeSeriesConf := range s.timeSeriesConfCache.TimeSeriesConfs {
			timeSeriesConf.listTimeSeriesRequest.Interval = interval
		}
		return s.timeSeriesConfCache.TimeSeriesConfs, nil
	}

	ret := []*timeSeriesConf{}
	req := &monitoringpb.ListMetricDescriptorsRequest{
		Name: fmt.Sprintf("projects/%s", s.Project),
	}

	filters := s.newListMetricDescriptorsFilters()
	if len(filters) == 0 {
		filters = []string{""}
	}

	for _, filter := range filters {
		// Add filter for list metric descriptors if
		// includeMetricTypePrefixes is specified,
		// this is more efficient than iterating over
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
				if s.GatherRawDistributionBuckets {
					tsConf := s.newTimeSeriesConf(metricType, startTime, endTime)
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

	s.timeSeriesConfCache = &timeSeriesConfCache{
		TimeSeriesConfs: ret,
		Generated:       time.Now(),
		TTL:             time.Duration(s.CacheTTL),
	}

	return ret, nil
}

// Do the work to gather an individual time series. Runs inside a
// timeseries-specific goroutine.
func (s *Stackdriver) gatherTimeSeries(
	ctx context.Context, grouper *lockedSeriesGrouper, tsConf *timeSeriesConf,
) error {
	tsReq := tsConf.listTimeSeriesRequest

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

			if tsDesc.ValueType == metricpb.MetricDescriptor_DISTRIBUTION {
				dist := p.Value.GetDistributionValue()
				if err := s.addDistribution(dist, tags, ts, grouper, tsConf); err != nil {
					return err
				}
			} else {
				var value interface{}

				// Types that are valid to be assigned to Value
				// See: https://godoc.org/google.golang.org/genproto/googleapis/monitoring/v3#TypedValue
				switch tsDesc.ValueType {
				case metricpb.MetricDescriptor_BOOL:
					value = p.Value.GetBoolValue()
				case metricpb.MetricDescriptor_INT64:
					value = p.Value.GetInt64Value()
				case metricpb.MetricDescriptor_DOUBLE:
					value = p.Value.GetDoubleValue()
				case metricpb.MetricDescriptor_STRING:
					value = p.Value.GetStringValue()
				}

				if err := grouper.Add(tsConf.measurement, tags, ts, tsConf.fieldKey, value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// AddDistribution adds metrics from a distribution value type.
func (s *Stackdriver) addDistribution(dist *distributionpb.Distribution, tags map[string]string, ts time.Time,
	grouper *lockedSeriesGrouper, tsConf *timeSeriesConf,
) error {
	field := tsConf.fieldKey
	name := tsConf.measurement

	if err := grouper.Add(name, tags, ts, field+"_count", dist.Count); err != nil {
		return err
	}
	if err := grouper.Add(name, tags, ts, field+"_mean", dist.Mean); err != nil {
		return err
	}
	if err := grouper.Add(name, tags, ts, field+"_sum_of_squared_deviation", dist.SumOfSquaredDeviation); err != nil {
		return err
	}

	if dist.Range != nil {
		if err := grouper.Add(name, tags, ts, field+"_range_min", dist.Range.Min); err != nil {
			return err
		}
		if err := grouper.Add(name, tags, ts, field+"_range_max", dist.Range.Max); err != nil {
			return err
		}
	}

	linearBuckets := dist.BucketOptions.GetLinearBuckets()
	exponentialBuckets := dist.BucketOptions.GetExponentialBuckets()
	explicitBuckets := dist.BucketOptions.GetExplicitBuckets()

	var numBuckets int32
	if linearBuckets != nil {
		numBuckets = linearBuckets.NumFiniteBuckets + 2
	} else if exponentialBuckets != nil {
		numBuckets = exponentialBuckets.NumFiniteBuckets + 2
	} else {
		numBuckets = int32(len(explicitBuckets.Bounds)) + 1
	}

	var i int32
	var count int64
	for i = 0; i < numBuckets; i++ {
		// The last bucket is the overflow bucket, and includes all values
		// greater than the previous bound.
		if i == numBuckets-1 {
			tags["lt"] = "+Inf"
		} else {
			var upperBound float64
			if linearBuckets != nil {
				upperBound = linearBuckets.Offset + (linearBuckets.Width * float64(i))
			} else if exponentialBuckets != nil {
				width := math.Pow(exponentialBuckets.GrowthFactor, float64(i))
				upperBound = exponentialBuckets.Scale * width
			} else if explicitBuckets != nil {
				upperBound = explicitBuckets.Bounds[i]
			}
			tags["lt"] = strconv.FormatFloat(upperBound, 'f', -1, 64)
		}

		// Add to the cumulative count; trailing buckets with value 0 are
		// omitted from the response.
		if i < int32(len(dist.BucketCounts)) {
			count += dist.BucketCounts[i]
		}
		if err := grouper.Add(name, tags, ts, field+"_bucket", count); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	f := func() telegraf.Input {
		return &Stackdriver{
			CacheTTL:                        defaultCacheTTL,
			RateLimit:                       defaultRateLimit,
			Delay:                           defaultDelay,
			GatherRawDistributionBuckets:    true,
			DistributionAggregationAligners: []string{},
		}
	}

	inputs.Add("stackdriver", f)
}
