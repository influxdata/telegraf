package gcp_stackdriver

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

type GCPStackdriver struct {
	Project                         string
	RateLimit                       int
	LookbackSeconds                 int64
	IncludeMetricTypePrefixes       []string
	ExcludeMetricTypePrefixes       []string
	IncludeTagPrefixes              []string
	ExcludeTagPrefixes              []string
	ScrapeDistributionBuckets       bool
	DistributionAggregationAligners []string

	client *monitoring.MetricClient
	ctx    context.Context
}

func init() {
	f := func() telegraf.Input {
		return &GCPStackdriver{
			RateLimit:                       14,
			LookbackSeconds:                 600,
			ScrapeDistributionBuckets:       true,
			DistributionAggregationAligners: []string{},
		}
	}

	inputs.Add("gcp_stackdriver", f)
}

func (s *GCPStackdriver) Description() string {
	return "Plugin that scrapes Google's v3 monitoring API."
}

func (s *GCPStackdriver) SampleConfig() string {
	return `
  ## GCP Project
  project = "projects/erudite-bloom-151019"

  ## API rate limit. On a default project, it seems that a single user can make
  ## ~14 requests per second. This might be configurable. Each API request can
  ## fetch every time series for a single metric type, though -- this is plenty
  ## fast for scraping all the builtin metric types (and even a handful of
  ## custom ones) every 60s.
  # rateLimit = 14

  ## Every query to stackdriver queries for data points that have timestamp t
  ## such that: (now() - lookbackSeconds) <= t < now(). Note that influx will
  ## de-dedupe points that are pulled twice, so it's best to be safe here, just
  ## in case it takes GCP awhile to get around to recording the data you seek.
  # lookbackSeconds = 600

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
`
}

func (s *GCPStackdriver) Gather(acc telegraf.Accumulator) error {
	log.Printf("Hello! Scraping from %s", s.Project)

	err := s.initializeStackdriverClient()
	if err != nil {
		log.Printf("Failed to create stackdriver monitoring client: %v", err)
		return err
	}

	timeSeriesRequests, err := s.generatetimeSeriesConfs()
	if err != nil {
		log.Printf("Failed to get metrics: %s\n", err)
		return err
	}

	s.scrapeAllTimeSeries(acc, timeSeriesRequests)
	if err != nil {
		return err
	}

	return nil
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
	listTimeSeriesRequest monitoringpb.ListTimeSeriesRequest
}

// Create and initialize a timeSeriesConf for a given GCP metric type with
// defaults taken from the gcp_stackdriver plugin configuration.
func (s *GCPStackdriver) newTimeSeriesConf(metricType string) *timeSeriesConf {
	endSec := time.Now().Unix()
	startSec := endSec - s.LookbackSeconds
	endTs := &googlepbts.Timestamp{Seconds: endSec}
	startTs := &googlepbts.Timestamp{Seconds: startSec}

	filter := fmt.Sprintf("metric.type = \"%s\"", metricType)
	interval := &monitoringpb.TimeInterval{
		EndTime:   endTs,
		StartTime: startTs}
	tsReq := monitoringpb.ListTimeSeriesRequest{
		Name:     s.Project,
		Filter:   filter,
		Interval: interval}

	cfg := &timeSeriesConf{
		measurement:           metricType,
		fieldPrefix:           "value",
		listTimeSeriesRequest: tsReq,
	}

	// GCP metric types have at least one slash, but we'll be defensive anyway.
	slashIdx := strings.LastIndex(metricType, "/")
	if slashIdx != -1 {
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
func (t *timeSeriesConf) initForAggregate(alignerStr string) *timeSeriesConf {
	alignerInt := monitoringpb.Aggregation_Aligner_value[alignerStr]
	aligner := monitoringpb.Aggregation_Aligner(alignerInt)
	agg := &monitoringpb.Aggregation{
		AlignmentPeriod:  &googlepbduration.Duration{Seconds: 60},
		PerSeriesAligner: aligner,
	}
	t.fieldPrefix = t.fieldPrefix + "_" + strings.ToLower(alignerStr)
	t.listTimeSeriesRequest.Aggregation = agg

	return t
}

// Change this configuration to query a distribution type. Time series of type
// distribution will generate a lot of fields (one for each bucket).
func (t *timeSeriesConf) initForDistribution() *timeSeriesConf {
	t.fieldPrefix = t.fieldPrefix + "_"

	return t
}

func (s *GCPStackdriver) initializeStackdriverClient() error {
	if s.client == nil {
		s.ctx = context.Background()

		// Creates a client.
		client, err := monitoring.NewMetricClient(s.ctx)
		if err != nil {
			return err
		}

		s.client = client
	}

	return nil
}

func includeExcludeHelper(key string, includes []string, excludes []string) bool {
	if includes != nil {
		for _, includeStr := range includes {
			if strings.Index(key, includeStr) == 0 {
				return true
			}
		}
	} else if excludes != nil {
		found := false
		for _, excludeStr := range excludes {
			if strings.Index(key, excludeStr) == 0 {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	} else {
		return true
	}

	return false
}

// Test whether a particular GCP metric type should be scraped by this plugin
// by checking the plugin name against the configuration's
// "includeMetricTypePrefixes" and "excludeMetricTypePrefixes"
func (s *GCPStackdriver) includeMetricType(metricType string) bool {
	k := metricType
	inc := s.IncludeMetricTypePrefixes
	exc := s.ExcludeMetricTypePrefixes

	return includeExcludeHelper(k, inc, exc)
}

func (s *GCPStackdriver) includeTag(tagKey string) bool {
	k := tagKey
	inc := s.IncludeTagPrefixes
	exc := s.ExcludeTagPrefixes

	return includeExcludeHelper(k, inc, exc)
}

// Generate a list of timeSeriesConfig structs by making a ListMetricDescriptors
// API request and filtering the result against our configuration.
func (s *GCPStackdriver) generatetimeSeriesConfs() ([]timeSeriesConf, error) {
	ret := []timeSeriesConf{}
	req := &monitoringpb.ListMetricDescriptorsRequest{Name: s.Project}
	resp := s.client.ListMetricDescriptors(s.ctx, req)

	for {
		metricDescriptor, err := resp.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		metricType := metricDescriptor.Type
		valueType := metricDescriptor.ValueType

		if s.includeMetricType(metricType) {
			if valueType == metricpb.MetricDescriptor_DISTRIBUTION {
				if s.ScrapeDistributionBuckets {
					tsConf := s.newTimeSeriesConf(metricType)
					tsConf = tsConf.initForDistribution()
					ret = append(ret, *tsConf)
				}
				for _, alignerStr := range s.DistributionAggregationAligners {
					tsConf := s.newTimeSeriesConf(metricType)
					tsConf = tsConf.initForAggregate(alignerStr)
					ret = append(ret, *tsConf)
				}
			} else {
				ret = append(ret, *s.newTimeSeriesConf(metricType))
			}
		}
	}

	return ret, nil
}

// Do the work! Create a (rate-limited) goroutine for each timeSeriesConf. The
// rate limiting is super-necessary: GCP will trout-slap us if we talk to it
// too quickly.
func (s *GCPStackdriver) scrapeAllTimeSeries(
	acc telegraf.Accumulator,
	types []timeSeriesConf) error {

	wg := &sync.WaitGroup{}
	lmtr := limiter.NewRateLimiter(s.RateLimit, time.Second)

	defer lmtr.Stop()
	wg.Add(len(types))
	for _, timeSeriesRequest := range types {
		go func(tsc timeSeriesConf) {
			<-lmtr.C
			defer wg.Done()
			acc.AddError(s.scrapeTimeSeries(acc, tsc))
		}(timeSeriesRequest)
	}

	wg.Wait()
	return nil
}

// Do the work to scrape an individual time series. Runs inside a
// timeseries-specific goroutine.
func (s *GCPStackdriver) scrapeTimeSeries(acc telegraf.Accumulator,
	tsConf timeSeriesConf) error {

	tsReq := tsConf.listTimeSeriesRequest
	measurement := tsConf.measurement
	fieldPrefix := tsConf.fieldPrefix
	tsResp := s.client.ListTimeSeries(s.ctx, &tsReq)

	for {
		tsDesc, tsErr := tsResp.Next()
		if tsErr == iterator.Done {
			break
		}
		if tsErr != nil {
			log.Printf("Request %s failure: %s\n", tsReq.String(), tsErr)
			return tsErr
		}

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
				fields = s.scrapeDistribution(acc, measurement, fieldPrefix, val)
			} else {
				var field interface{}

				switch tsDesc.ValueType {
				case metricpb.MetricDescriptor_BOOL:
					field = p.Value.GetBoolValue()
				case metricpb.MetricDescriptor_INT64:
					field = p.Value.GetInt64Value()
				case metricpb.MetricDescriptor_DOUBLE:
					field = p.Value.GetDoubleValue()
				case metricpb.MetricDescriptor_STRING:
					field = p.Value.GetStringValue()
				default:
					// TODO: AddError here? Is it really an error?
					// TODO: Maybe default to `field = p.Value.GetValue()`
					continue
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
func (s *GCPStackdriver) scrapeDistribution(acc telegraf.Accumulator,
	metricType string,
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
