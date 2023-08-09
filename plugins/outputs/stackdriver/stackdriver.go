//go:generate ../../../tools/readme_config_includer/generator
package stackdriver

import (
	"context"
	_ "embed"
	"fmt"
	"hash/fnv"
	"path"
	"sort"
	"strings"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/option"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

// Stackdriver is the Google Stackdriver config info.
type Stackdriver struct {
	Project              string            `toml:"project"`
	Namespace            string            `toml:"namespace"`
	ResourceType         string            `toml:"resource_type"`
	ResourceLabels       map[string]string `toml:"resource_labels"`
	MetricTypePrefix     string            `toml:"metric_type_prefix"`
	MetricNameFormat     string            `toml:"metric_name_format"`
	MetricDataType       string            `toml:"metric_data_type"`
	TagsAsResourceLabels []string          `toml:"tags_as_resource_label"`
	Log                  telegraf.Logger   `toml:"-"`

	client       *monitoring.MetricClient
	counterCache *counterCache
}

const (
	// QuotaLabelsPerMetricDescriptor is the limit
	// to labels (tags) per metric descriptor.
	QuotaLabelsPerMetricDescriptor = 10
	// QuotaStringLengthForLabelKey is the limit
	// to string length for label key.
	QuotaStringLengthForLabelKey = 100
	// QuotaStringLengthForLabelValue is the limit
	// to string length for label value.
	QuotaStringLengthForLabelValue = 1024

	// MaxInt is the max int64 value.
	MaxInt = int(^uint(0) >> 1)

	errStringPointsOutOfOrder  = "one or more of the points specified had an older end time than the most recent point"
	errStringPointsTooOld      = "data points cannot be written more than 24h in the past"
	errStringPointsTooFrequent = "one or more points were written more frequently than the maximum sampling period configured for the metric"
)

func (s *Stackdriver) Init() error {
	if s.MetricTypePrefix == "" {
		s.MetricTypePrefix = "custom.googleapis.com"
	}

	switch s.MetricNameFormat {
	case "":
		s.MetricNameFormat = "path"
	case "path", "official":
	default:
		return fmt.Errorf("unrecognized metric name format: %s", s.MetricNameFormat)
	}

	switch s.MetricDataType {
	case "":
		s.MetricDataType = "source"
	case "source", "double":
	default:
		return fmt.Errorf("unrecognized metric data type: %s", s.MetricDataType)
	}

	return nil
}

func (*Stackdriver) SampleConfig() string {
	return sampleConfig
}

// Connect initiates the primary connection to the GCP project.
func (s *Stackdriver) Connect() error {
	if s.Project == "" {
		return fmt.Errorf("project is a required field for stackdriver output")
	}

	if s.Namespace == "" {
		s.Log.Warn("plugin-level namespace is empty")
	}

	if s.ResourceType == "" {
		s.ResourceType = "global"
	}

	if s.ResourceLabels == nil {
		s.ResourceLabels = make(map[string]string, 1)
	}

	if s.counterCache == nil {
		s.counterCache = NewCounterCache(s.Log)
	}

	s.ResourceLabels["project_id"] = s.Project

	if s.client == nil {
		ctx := context.Background()
		client, err := monitoring.NewMetricClient(ctx, option.WithUserAgent(internal.ProductToken()))
		if err != nil {
			return err
		}
		s.client = client
	}

	return nil
}

// Sorted returns a copy of the metrics in time ascending order.  A copy is
// made to avoid modifying the input metric slice since doing so is not
// allowed.
func sorted(metrics []telegraf.Metric) []telegraf.Metric {
	batch := make([]telegraf.Metric, 0, len(metrics))
	for i := len(metrics) - 1; i >= 0; i-- {
		batch = append(batch, metrics[i])
	}
	sort.Slice(batch, func(i, j int) bool {
		return batch[i].Time().Before(batch[j].Time())
	})
	return batch
}

type timeSeriesBuckets map[uint64][]*monitoringpb.TimeSeries

func (tsb timeSeriesBuckets) Add(m telegraf.Metric, f *telegraf.Field, ts *monitoringpb.TimeSeries) {
	h := fnv.New64a()
	h.Write([]byte(m.Name()))
	h.Write([]byte{'\n'})
	h.Write([]byte(f.Key))
	h.Write([]byte{'\n'})
	for key, value := range m.Tags() {
		h.Write([]byte(key))
		h.Write([]byte{'\n'})
		h.Write([]byte(value))
		h.Write([]byte{'\n'})
	}
	k := h.Sum64()

	s := tsb[k]
	s = append(s, ts)
	tsb[k] = s
}

// Split metrics up by timestamp and send to Google Cloud Stackdriver
func (s *Stackdriver) Write(metrics []telegraf.Metric) error {
	metricBatch := make(map[int64][]telegraf.Metric)
	timestamps := []int64{}
	for _, metric := range sorted(metrics) {
		timestamp := metric.Time().UnixNano()
		if existingSlice, ok := metricBatch[timestamp]; ok {
			metricBatch[timestamp] = append(existingSlice, metric)
		} else {
			metricBatch[timestamp] = []telegraf.Metric{metric}
			timestamps = append(timestamps, timestamp)
		}
	}

	// sort the timestamps we collected
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	s.Log.Debugf("received %d metrics\n", len(metrics))
	s.Log.Debugf("split into %d groups by timestamp\n", len(metricBatch))
	for _, timestamp := range timestamps {
		if err := s.sendBatch(metricBatch[timestamp]); err != nil {
			return err
		}
	}

	return nil
}

// Write the metrics to Google Cloud Stackdriver.
func (s *Stackdriver) sendBatch(batch []telegraf.Metric) error {
	ctx := context.Background()

	buckets := make(timeSeriesBuckets)
	for _, m := range batch {
		for _, f := range m.FieldList() {
			value, err := s.getStackdriverTypedValue(f.Value)
			if err != nil {
				s.Log.Errorf("Get type failed: %q", err)
				continue
			}

			if value == nil {
				continue
			}

			metricKind, err := getStackdriverMetricKind(m.Type())
			if err != nil {
				s.Log.Errorf("Get kind for metric %q (%T) field %q failed: %s", m.Name(), m.Type(), f, err)
				continue
			}

			startTime, endTime := getStackdriverIntervalEndpoints(metricKind, value, m, f, s.counterCache)

			timeInterval, err := getStackdriverTimeInterval(metricKind, startTime, endTime)
			if err != nil {
				s.Log.Errorf("Get time interval failed: %s", err)
				continue
			}

			// Prepare an individual data point.
			dataPoint := &monitoringpb.Point{
				Interval: timeInterval,
				Value:    value,
			}

			// Convert any declared tag to a resource label and remove it from
			// the metric
			resourceLabels := s.ResourceLabels
			for _, tag := range s.TagsAsResourceLabels {
				if val, ok := m.GetTag(tag); ok {
					resourceLabels[tag] = val
					m.RemoveTag(tag)
				}
			}

			// Prepare time series.
			timeSeries := &monitoringpb.TimeSeries{
				Metric: &metricpb.Metric{
					Type:   s.generateMetricName(m, f.Key),
					Labels: s.getStackdriverLabels(m.TagList()),
				},
				MetricKind: metricKind,
				Resource: &monitoredrespb.MonitoredResource{
					Type:   s.ResourceType,
					Labels: resourceLabels,
				},
				Points: []*monitoringpb.Point{
					dataPoint,
				},
			}

			buckets.Add(m, f, timeSeries)

			// If the metric is untyped, it will end with unknown. We will also
			// send another metric with the unknown:counter suffix. Google will
			// do some heuristics to know which one to use for queries. This
			// only occurs when using the official name format.
			if s.MetricNameFormat == "official" && strings.HasSuffix(timeSeries.Metric.Type, "unknown") {
				metricKind := metricpb.MetricDescriptor_CUMULATIVE
				startTime, endTime := getStackdriverIntervalEndpoints(metricKind, value, m, f, s.counterCache)
				timeInterval, err := getStackdriverTimeInterval(metricKind, startTime, endTime)
				if err != nil {
					s.Log.Errorf("Get time interval failed: %s", err)
					continue
				}
				dataPoint := &monitoringpb.Point{
					Interval: timeInterval,
					Value:    value,
				}

				counterTimeSeries := &monitoringpb.TimeSeries{
					Metric: &metricpb.Metric{
						Type:   s.generateMetricName(m, f.Key) + ":counter",
						Labels: s.getStackdriverLabels(m.TagList()),
					},
					MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
					Resource: &monitoredrespb.MonitoredResource{
						Type:   s.ResourceType,
						Labels: resourceLabels,
					},
					Points: []*monitoringpb.Point{
						dataPoint,
					},
				}
				buckets.Add(m, f, counterTimeSeries)
			}
		}
	}

	// process the buckets in order
	keys := make([]uint64, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for len(buckets) != 0 {
		// can send up to 200 time series to stackdriver
		timeSeries := make([]*monitoringpb.TimeSeries, 0, 200)
		for i := 0; i < len(keys) && len(timeSeries) < cap(timeSeries); i++ {
			k := keys[i]
			s := buckets[k]
			timeSeries = append(timeSeries, s[0])
			if len(s) == 1 {
				delete(buckets, k)
				keys = append(keys[:i], keys[i+1:]...)
				i--
				continue
			}

			s = s[1:]
			buckets[k] = s
		}

		// Prepare time series request.
		timeSeriesRequest := &monitoringpb.CreateTimeSeriesRequest{
			Name:       fmt.Sprintf("projects/%s", s.Project),
			TimeSeries: timeSeries,
		}

		// Create the time series in Stackdriver.
		err := s.client.CreateTimeSeries(ctx, timeSeriesRequest)
		if err != nil {
			if strings.Contains(err.Error(), errStringPointsOutOfOrder) ||
				strings.Contains(err.Error(), errStringPointsTooOld) ||
				strings.Contains(err.Error(), errStringPointsTooFrequent) {
				s.Log.Debugf("Unable to write to Stackdriver: %s", err)
				return nil
			}
			s.Log.Errorf("Unable to write to Stackdriver: %s", err)
			return err
		}
	}

	return nil
}

func (s *Stackdriver) generateMetricName(m telegraf.Metric, key string) string {
	if s.MetricNameFormat == "path" {
		return path.Join(s.MetricTypePrefix, s.Namespace, m.Name(), key)
	}

	name := m.Name() + "_" + key
	if s.Namespace != "" {
		name = s.Namespace + "_" + m.Name() + "_" + key
	}

	var kind string
	switch m.Type() {
	case telegraf.Gauge:
		kind = "gauge"
	case telegraf.Untyped:
		kind = "unknown"
	case telegraf.Counter:
		kind = "counter"
	case telegraf.Histogram:
		kind = "histogram"
	default:
		kind = ""
	}

	return path.Join(s.MetricTypePrefix, name, kind)
}

func getStackdriverIntervalEndpoints(
	kind metricpb.MetricDescriptor_MetricKind,
	value *monitoringpb.TypedValue,
	m telegraf.Metric,
	f *telegraf.Field,
	cc *counterCache,
) (*timestamppb.Timestamp, *timestamppb.Timestamp) {
	endTime := timestamppb.New(m.Time())
	var startTime *timestamppb.Timestamp
	if kind == metricpb.MetricDescriptor_CUMULATIVE {
		// Interval starts for stackdriver CUMULATIVE metrics must reset any time
		// the counter resets, so we keep a cache of the start times and last
		// observed values for each counter in the batch.
		startTime = cc.GetStartTime(GetCounterCacheKey(m, f), value, endTime)
	}
	return startTime, endTime
}

func getStackdriverTimeInterval(
	m metricpb.MetricDescriptor_MetricKind,
	startTime *timestamppb.Timestamp,
	endTime *timestamppb.Timestamp,
) (*monitoringpb.TimeInterval, error) {
	switch m {
	case metricpb.MetricDescriptor_GAUGE:
		return &monitoringpb.TimeInterval{
			EndTime: endTime,
		}, nil
	case metricpb.MetricDescriptor_CUMULATIVE:
		return &monitoringpb.TimeInterval{
			StartTime: startTime,
			EndTime:   endTime,
		}, nil
	case metricpb.MetricDescriptor_DELTA, metricpb.MetricDescriptor_METRIC_KIND_UNSPECIFIED:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported metric kind %T", m)
	}
}

func getStackdriverMetricKind(vt telegraf.ValueType) (metricpb.MetricDescriptor_MetricKind, error) {
	switch vt {
	case telegraf.Untyped:
		return metricpb.MetricDescriptor_GAUGE, nil
	case telegraf.Gauge:
		return metricpb.MetricDescriptor_GAUGE, nil
	case telegraf.Counter:
		return metricpb.MetricDescriptor_CUMULATIVE, nil
	case telegraf.Histogram, telegraf.Summary:
		fallthrough
	default:
		return metricpb.MetricDescriptor_METRIC_KIND_UNSPECIFIED, fmt.Errorf("unsupported telegraf value type: %T", vt)
	}
}

func (s *Stackdriver) getStackdriverTypedValue(value interface{}) (*monitoringpb.TypedValue, error) {
	if s.MetricDataType == "double" {
		v, err := internal.ToFloat64(value)
		if err != nil {
			return nil, err
		}

		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_DoubleValue{
				DoubleValue: v,
			},
		}, nil
	}

	switch v := value.(type) {
	case uint64:
		if v <= uint64(MaxInt) {
			return &monitoringpb.TypedValue{
				Value: &monitoringpb.TypedValue_Int64Value{
					Int64Value: int64(v),
				},
			}, nil
		}
		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: int64(MaxInt),
			},
		}, nil
	case int64:
		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: v,
			},
		}, nil
	case float64:
		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_DoubleValue{
				DoubleValue: v,
			},
		}, nil
	case bool:
		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_BoolValue{
				BoolValue: v,
			},
		}, nil
	case string:
		// String value types are not available for custom metrics
		return nil, nil
	default:
		return nil, fmt.Errorf("value type \"%T\" not supported for stackdriver custom metrics", v)
	}
}

func (s *Stackdriver) getStackdriverLabels(tags []*telegraf.Tag) map[string]string {
	labels := make(map[string]string)
	for _, t := range tags {
		labels[t.Key] = t.Value
	}
	for k, v := range labels {
		if len(k) > QuotaStringLengthForLabelKey {
			s.Log.Warnf("Removing tag %q key exceeds string length for label key [%d]", k, QuotaStringLengthForLabelKey)
			delete(labels, k)
			continue
		}
		if len(v) > QuotaStringLengthForLabelValue {
			s.Log.Warnf("Removing tag %q value exceeds string length for label value [%d]", k, QuotaStringLengthForLabelValue)
			delete(labels, k)
			continue
		}
	}
	if len(labels) > QuotaLabelsPerMetricDescriptor {
		excess := len(labels) - QuotaLabelsPerMetricDescriptor
		s.Log.Warnf("Tag count [%d] exceeds quota for stackdriver labels [%d] removing [%d] random tags", len(labels), QuotaLabelsPerMetricDescriptor, excess)
		for k := range labels {
			if excess == 0 {
				break
			}
			excess--
			delete(labels, k)
		}
	}

	return labels
}

// Close will terminate the session to the backend, returning error if an issue arises.
func (s *Stackdriver) Close() error {
	return s.client.Close()
}

func newStackdriver() *Stackdriver {
	return &Stackdriver{}
}

func init() {
	outputs.Add("stackdriver", func() telegraf.Output {
		return newStackdriver()
	})
}
