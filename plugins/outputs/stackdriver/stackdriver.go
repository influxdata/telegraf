package stackdriver

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"path"
	"sort"
	"strings"

	monitoring "cloud.google.com/go/monitoring/apiv3" // Imports the Stackdriver Monitoring client package.
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// Stackdriver is the Google Stackdriver config info.
type Stackdriver struct {
	Project        string
	Namespace      string
	ResourceType   string            `toml:"resource_type"`
	ResourceLabels map[string]string `toml:"resource_labels"`

	client *monitoring.MetricClient
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

	// StartTime for cumulative metrics.
	StartTime = int64(1)
	// MaxInt is the max int64 value.
	MaxInt = int(^uint(0) >> 1)

	errStringPointsOutOfOrder  = "One or more of the points specified had an older end time than the most recent point"
	errStringPointsTooOld      = "Data points cannot be written more than 24h in the past"
	errStringPointsTooFrequent = "One or more points were written more frequently than the maximum sampling period configured for the metric"
)

var sampleConfig = `
  ## GCP Project
  project = "erudite-bloom-151019"

  ## The namespace for the metric descriptor
  namespace = "telegraf"

  ## Custom resource type
  # resource_type = "generic_node"

  ## Additonal resource labels
  # [outputs.stackdriver.resource_labels]
  #   node_id = "$HOSTNAME"
  #   namespace = "myapp"
  #   location = "eu-north0"
`

// Connect initiates the primary connection to the GCP project.
func (s *Stackdriver) Connect() error {
	if s.Project == "" {
		return fmt.Errorf("Project is a required field for stackdriver output")
	}

	if s.Namespace == "" {
		return fmt.Errorf("Namespace is a required field for stackdriver output")
	}

	if s.ResourceType == "" {
		s.ResourceType = "global"
	}

	if s.ResourceLabels == nil {
		s.ResourceLabels = make(map[string]string, 1)
	}

	s.ResourceLabels["project_id"] = s.Project

	if s.client == nil {
		ctx := context.Background()
		client, err := monitoring.NewMetricClient(ctx)
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

// Write the metrics to Google Cloud Stackdriver.
func (s *Stackdriver) Write(metrics []telegraf.Metric) error {
	ctx := context.Background()

	batch := sorted(metrics)
	buckets := make(timeSeriesBuckets)
	for _, m := range batch {
		for _, f := range m.FieldList() {
			value, err := getStackdriverTypedValue(f.Value)
			if err != nil {
				log.Printf("E! [outputs.stackdriver] get type failed: %s", err)
				continue
			}

			if value == nil {
				continue
			}

			metricKind, err := getStackdriverMetricKind(m.Type())
			if err != nil {
				log.Printf("E! [outputs.stackdriver] get metric failed: %s", err)
				continue
			}

			timeInterval, err := getStackdriverTimeInterval(metricKind, StartTime, m.Time().Unix())
			if err != nil {
				log.Printf("E! [outputs.stackdriver] get time interval failed: %s", err)
				continue
			}

			// Prepare an individual data point.
			dataPoint := &monitoringpb.Point{
				Interval: timeInterval,
				Value:    value,
			}

			// Prepare time series.
			timeSeries := &monitoringpb.TimeSeries{
				Metric: &metricpb.Metric{
					Type:   path.Join("custom.googleapis.com", s.Namespace, m.Name(), f.Key),
					Labels: getStackdriverLabels(m.TagList()),
				},
				MetricKind: metricKind,
				Resource: &monitoredrespb.MonitoredResource{
					Type:   s.ResourceType,
					Labels: s.ResourceLabels,
				},
				Points: []*monitoringpb.Point{
					dataPoint,
				},
			}

			buckets.Add(m, f, timeSeries)
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
			Name:       monitoring.MetricProjectPath(s.Project),
			TimeSeries: timeSeries,
		}

		// Create the time series in Stackdriver.
		err := s.client.CreateTimeSeries(ctx, timeSeriesRequest)
		if err != nil {
			if strings.Contains(err.Error(), errStringPointsOutOfOrder) ||
				strings.Contains(err.Error(), errStringPointsTooOld) ||
				strings.Contains(err.Error(), errStringPointsTooFrequent) {
				log.Printf("D! [outputs.stackdriver] unable to write to Stackdriver: %s", err)
				return nil
			}
			log.Printf("E! [outputs.stackdriver] unable to write to Stackdriver: %s", err)
			return err
		}
	}

	return nil
}

func getStackdriverTimeInterval(
	m metricpb.MetricDescriptor_MetricKind,
	start int64,
	end int64,
) (*monitoringpb.TimeInterval, error) {
	switch m {
	case metricpb.MetricDescriptor_GAUGE:
		return &monitoringpb.TimeInterval{
			EndTime: &googlepb.Timestamp{
				Seconds: end,
			},
		}, nil
	case metricpb.MetricDescriptor_CUMULATIVE:
		return &monitoringpb.TimeInterval{
			StartTime: &googlepb.Timestamp{
				Seconds: start,
			},
			EndTime: &googlepb.Timestamp{
				Seconds: end,
			},
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
		return metricpb.MetricDescriptor_METRIC_KIND_UNSPECIFIED, fmt.Errorf("unsupported telegraf value type")
	}
}

func getStackdriverTypedValue(value interface{}) (*monitoringpb.TypedValue, error) {
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
				Int64Value: int64(v),
			},
		}, nil
	case float64:
		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_DoubleValue{
				DoubleValue: float64(v),
			},
		}, nil
	case bool:
		return &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_BoolValue{
				BoolValue: bool(v),
			},
		}, nil
	case string:
		// String value types are not available for custom metrics
		return nil, nil
	default:
		return nil, fmt.Errorf("value type \"%T\" not supported for stackdriver custom metrics", v)
	}
}

func getStackdriverLabels(tags []*telegraf.Tag) map[string]string {
	labels := make(map[string]string)
	for _, t := range tags {
		labels[t.Key] = t.Value
	}
	for k, v := range labels {
		if len(k) > QuotaStringLengthForLabelKey {
			log.Printf(
				"W! [outputs.stackdriver] removing tag [%s] key exceeds string length for label key [%d]",
				k,
				QuotaStringLengthForLabelKey,
			)
			delete(labels, k)
			continue
		}
		if len(v) > QuotaStringLengthForLabelValue {
			log.Printf(
				"W! [outputs.stackdriver] removing tag [%s] value exceeds string length for label value [%d]",
				k,
				QuotaStringLengthForLabelValue,
			)
			delete(labels, k)
			continue
		}
	}
	if len(labels) > QuotaLabelsPerMetricDescriptor {
		excess := len(labels) - QuotaLabelsPerMetricDescriptor
		log.Printf(
			"W! [outputs.stackdriver] tag count [%d] exceeds quota for stackdriver labels [%d] removing [%d] random tags",
			len(labels),
			QuotaLabelsPerMetricDescriptor,
			excess,
		)
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

// SampleConfig returns the formatted sample configuration for the plugin.
func (s *Stackdriver) SampleConfig() string {
	return sampleConfig
}

// Description returns the human-readable function definition of the plugin.
func (s *Stackdriver) Description() string {
	return "Configuration for Google Cloud Stackdriver to send metrics to"
}

func newStackdriver() *Stackdriver {
	return &Stackdriver{}
}

func init() {
	outputs.Add("stackdriver", func() telegraf.Output {
		return newStackdriver()
	})
}
