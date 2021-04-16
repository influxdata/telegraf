package otlp

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	// TODO: this should be simplified
	// Imports the OTLP client from the prometheus sidecar

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	metricsService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	otlpcommonpb "go.opentelemetry.io/proto/otlp/common/v1"
	otlpmetricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	otlpresourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc/metadata"
)

// OTLP is the OpenTelemetry Protocol config info.
type OTLP struct {
	Endpoint   string            `toml:"endpoint"`
	Timeout    string            `toml:"timeout"`
	Headers    map[string]string `toml:"headers"`
	Attributes map[string]string `toml:"attributes"`

	Namespace string
	Log       telegraf.Logger

	client       *Client
	resourceTags []*telegraf.Tag
	grpcTimeout  time.Duration
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

	defaultEndpoint            = "http://localhost:4317"
	defaultTimeout             = time.Second * 60
	instrumentationLibraryName = "Telegraf"
)

var sampleConfig = `
  ## OpenTelemetry endpoint
  endpoint = "localhost:4317"

  ## Timeout used when sending data over grpc
  timeout = 10s

  # Additional resource attributes
  [outputs.otlp.attributes]
  	"service.name" = "demo"

  # Additional grpc metadata
  [outputs.otlp.headers]
    key1 = "value1"

`

// Connect initiates the primary connection to the OTLP endpoint.
func (o *OTLP) Connect() error {
	if o.Endpoint == "" {
		o.Endpoint = defaultEndpoint
	}
	endpoint, err := url.Parse(o.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint configured")
	}

	if o.Timeout == "" {
		o.grpcTimeout = defaultTimeout
	} else {
		o.grpcTimeout, err = time.ParseDuration(o.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout configured")
		}
	}

	for k, v := range o.Attributes {
		o.resourceTags = append(o.resourceTags, &telegraf.Tag{Key: k, Value: v})
	}

	if o.Headers == nil {
		o.Headers = make(map[string]string, 1)
	}

	o.Headers["telemetry-reporting-agent"] = fmt.Sprint(
		"telegraf/",
		internal.Version(),
	)

	if o.client == nil {
		ctx := context.Background()
		o.client = NewClient(ClientConfig{
			URL:     endpoint,
			Headers: metadata.New(o.Headers),
			Timeout: o.grpcTimeout,
		})
		if err := o.client.Selftest(ctx); err != nil {
			_ = o.client.Close()
			return err
		}
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

// getNanos converts a millisecond timestamp into a OTLP nanosecond timestamp.
func getNanos(t int64) uint64 {
	return uint64(time.Duration(t) * time.Nanosecond)
}

func monotonicIntegerPoint(labels []*otlpcommonpb.StringKeyValue, start, end int64, value int64) *otlpmetricpb.IntSum {
	integer := &otlpmetricpb.IntDataPoint{
		Labels:            labels,
		StartTimeUnixNano: uint64(start),
		TimeUnixNano:      uint64(end),
		Value:             value,
	}
	return &otlpmetricpb.IntSum{
		IsMonotonic:            true,
		AggregationTemporality: otlpmetricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
		DataPoints:             []*otlpmetricpb.IntDataPoint{integer},
	}
}

func monotonicDoublePoint(labels []*otlpcommonpb.StringKeyValue, start, end int64, value float64) *otlpmetricpb.DoubleSum {
	double := &otlpmetricpb.DoubleDataPoint{
		Labels:            labels,
		StartTimeUnixNano: getNanos(start),
		TimeUnixNano:      getNanos(end),
		Value:             value,
	}
	return &otlpmetricpb.DoubleSum{
		IsMonotonic:            true,
		AggregationTemporality: otlpmetricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
		DataPoints:             []*otlpmetricpb.DoubleDataPoint{double},
	}
}

func protoLabel(tag *telegraf.Tag) *otlpcommonpb.KeyValue {
	return &otlpcommonpb.KeyValue{
		Key: tag.Key,
		Value: &otlpcommonpb.AnyValue{
			Value: &otlpcommonpb.AnyValue_StringValue{
				StringValue: tag.Value,
			},
		},
	}
}

func protoStringLabel(tag *telegraf.Tag) *otlpcommonpb.StringKeyValue {
	return &otlpcommonpb.StringKeyValue{
		Key:   tag.Key,
		Value: tag.Value,
	}
}

func protoResourceAttributes(tags []*telegraf.Tag) []*otlpcommonpb.KeyValue {
	ret := make([]*otlpcommonpb.KeyValue, len(tags))
	for i := range tags {
		ret[i] = protoLabel(tags[i])
	}
	return ret
}

func protoStringLabels(tags []*telegraf.Tag) []*otlpcommonpb.StringKeyValue {
	ret := make([]*otlpcommonpb.StringKeyValue, len(tags))
	for i := range tags {
		ret[i] = protoStringLabel(tags[i])
	}
	return ret
}

func (o *OTLP) protoTimeseries(m telegraf.Metric, f *telegraf.Field) (*otlpmetricpb.ResourceMetrics, *otlpmetricpb.Metric) {
	metric := &otlpmetricpb.Metric{
		Name:        fmt.Sprintf("%s.%s", m.Name(), f.Key),
		Description: "", // TODO
		Unit:        "", // TODO
	}
	return &otlpmetricpb.ResourceMetrics{
		Resource: &otlpresourcepb.Resource{
			Attributes: protoResourceAttributes(o.resourceTags),
		},
		InstrumentationLibraryMetrics: []*otlpmetricpb.InstrumentationLibraryMetrics{
			{
				InstrumentationLibrary: &otlpcommonpb.InstrumentationLibrary{
					Name:    instrumentationLibraryName,
					Version: internal.Version(),
				},
				Metrics: []*otlpmetricpb.Metric{metric},
			},
		},
	}, metric
}

func intGauge(labels []*otlpcommonpb.StringKeyValue, ts int64, value int64) *otlpmetricpb.IntGauge {
	integer := &otlpmetricpb.IntDataPoint{
		Labels:       labels,
		TimeUnixNano: uint64(ts),
		Value:        value,
	}
	return &otlpmetricpb.IntGauge{
		DataPoints: []*otlpmetricpb.IntDataPoint{integer},
	}
}

func doubleGauge(labels []*otlpcommonpb.StringKeyValue, ts int64, value float64) *otlpmetricpb.DoubleGauge {
	double := &otlpmetricpb.DoubleDataPoint{
		Labels:       labels,
		TimeUnixNano: uint64(ts),
		Value:        value,
	}
	return &otlpmetricpb.DoubleGauge{
		DataPoints: []*otlpmetricpb.DoubleDataPoint{double},
	}
}

// Write the metrics to OTLP destination
func (o *OTLP) Write(metrics []telegraf.Metric) error {
	batch := sorted(metrics)
	samples := []*otlpmetricpb.ResourceMetrics{}
	currentTs := time.Now().UnixNano()
	for _, m := range batch {
		for _, f := range m.FieldList() {
			sample, point := o.protoTimeseries(m, f)

			labels := protoStringLabels(m.TagList())
			ts := m.Time().UnixNano()

			switch m.Type() {
			case telegraf.Counter:
				switch v := f.Value.(type) {
				case uint64:
					if v <= uint64(MaxInt) {
						point.Data = &otlpmetricpb.Metric_IntSum{
							IntSum: monotonicIntegerPoint(labels, ts, currentTs, int64(v)),
						}
					} else {
						point.Data = &otlpmetricpb.Metric_IntSum{
							IntSum: monotonicIntegerPoint(labels, ts, currentTs, int64(MaxInt)),
						}
					}
				case int64:
					point.Data = &otlpmetricpb.Metric_IntSum{
						IntSum: monotonicIntegerPoint(labels, ts, currentTs, v),
					}
				case float64:
					point.Data = &otlpmetricpb.Metric_DoubleSum{
						DoubleSum: monotonicDoublePoint(labels, ts, currentTs, v),
					}
				case bool:
					if v {
						point.Data = &otlpmetricpb.Metric_IntSum{
							IntSum: monotonicIntegerPoint(labels, ts, currentTs, 1),
						}
					} else {
						point.Data = &otlpmetricpb.Metric_IntSum{
							IntSum: monotonicIntegerPoint(labels, ts, currentTs, 0),
						}
					}
				case string:
					o.Log.Error("get type failed: unsupported telegraf value type string")
					continue
				default:
					o.Log.Errorf("get type failed: unsupported telegraf value type %v\n", f.Value)
					continue
				}
			case telegraf.Gauge, telegraf.Untyped:
				switch v := f.Value.(type) {
				case uint64:
					if v <= uint64(MaxInt) {
						point.Data = &otlpmetricpb.Metric_IntGauge{
							IntGauge: intGauge(labels, ts, int64(v)),
						}
					} else {
						point.Data = &otlpmetricpb.Metric_IntGauge{
							IntGauge: intGauge(labels, ts, int64(MaxInt)),
						}
					}
				case int64:
					point.Data = &otlpmetricpb.Metric_IntGauge{
						IntGauge: intGauge(labels, ts, v),
					}
				case float64:
					point.Data = &otlpmetricpb.Metric_DoubleGauge{
						DoubleGauge: doubleGauge(labels, ts, v),
					}
				case bool:
					if v {
						point.Data = &otlpmetricpb.Metric_IntGauge{
							IntGauge: intGauge(labels, ts, 1),
						}
					} else {
						point.Data = &otlpmetricpb.Metric_IntGauge{
							IntGauge: intGauge(labels, ts, 0),
						}
					}
				case string:
					o.Log.Error("get type failed: unsupported telegraf value type string")
					continue
				default:
					o.Log.Errorf("get type failed: unsupported telegraf value type %v\n", f.Value)
					continue
				}
			// TODO: add support for histogram & summary
			case telegraf.Histogram, telegraf.Summary:
				fallthrough
			default:
				o.Log.Errorf("get type failed: unsupported telegraf metric kind %v\n", m.Type())
				continue
			}
			samples = append(samples, sample)
		}
	}

	if err := o.client.Store(&metricsService.ExportMetricsServiceRequest{
		ResourceMetrics: samples,
	}); err != nil {
		o.Log.Errorf("unable to write to endpoint: %s", err)
		return err
	}
	return nil
}

// Close will terminate the session to the backend, returning error if an issue arises.
func (o *OTLP) Close() error {
	return o.client.Close()
}

// SampleConfig returns the formatted sample configuration for the plugin.
func (o *OTLP) SampleConfig() string {
	return sampleConfig
}

// Description returns the human-readable function definition of the plugin.
func (o *OTLP) Description() string {
	return "Configuration for OTLP to send metrics to"
}

func newOTLP() *OTLP {
	return &OTLP{}
}

func init() {
	outputs.Add("otlp", func() telegraf.Output {
		return newOTLP()
	})
}
