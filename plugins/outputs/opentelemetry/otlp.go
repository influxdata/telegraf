package opentelemetry

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	metricsService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc/metadata"
)

// OpenTelemetry is the OpenTelemetry Protocol config info.
type OpenTelemetry struct {
	Endpoint   string            `toml:"endpoint"`
	Timeout    string            `toml:"timeout"`
	Headers    map[string]string `toml:"headers"`
	Attributes map[string]string `toml:"attributes"`

	Namespace string
	Log       telegraf.Logger `toml:"-"`

	client       *Client
	resourceTags []*telegraf.Tag
	grpcTimeout  time.Duration
}

const (
	// maxInt is the max int64 value.
	maxInt = int(^uint(0) >> 1)

	defaultEndpoint            = "http://localhost:4317"
	defaultTimeout             = time.Second * 60
	instrumentationLibraryName = "Telegraf"
)

var sampleConfig = `
  ## OpenTelemetry endpoint
  # endpoint = "http://localhost:4317"

  ## Timeout used when sending data over grpc
  # timeout = "10s"

  # Additional resource attributes
  [outputs.opentelemetry.attributes]
  	"service.name" = "demo"

  # Additional grpc metadata
  [outputs.opentelemetry.headers]
    key1 = "value1"

`

// Connect initiates the primary connection to the OpenTelemetry endpoint.
func (o *OpenTelemetry) Connect() error {
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

func monotonicIntegerPoint(labels []*commonpb.StringKeyValue, start, end int64, value int64) *metricpb.IntSum {
	integer := &metricpb.IntDataPoint{
		Labels:            labels,
		StartTimeUnixNano: uint64(start),
		TimeUnixNano:      uint64(end),
		Value:             value,
	}
	return &metricpb.IntSum{
		IsMonotonic:            true,
		AggregationTemporality: metricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
		DataPoints:             []*metricpb.IntDataPoint{integer},
	}
}

func monotonicDoublePoint(labels []*commonpb.StringKeyValue, start, end int64, value float64) *metricpb.DoubleSum {
	double := &metricpb.DoubleDataPoint{
		Labels:            labels,
		StartTimeUnixNano: uint64(time.Duration(start) * time.Nanosecond),
		TimeUnixNano:      uint64(time.Duration(end) * time.Nanosecond),
		Value:             value,
	}
	return &metricpb.DoubleSum{
		IsMonotonic:            true,
		AggregationTemporality: metricpb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
		DataPoints:             []*metricpb.DoubleDataPoint{double},
	}
}

func protoLabel(tag *telegraf.Tag) *commonpb.KeyValue {
	return &commonpb.KeyValue{
		Key: tag.Key,
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: tag.Value,
			},
		},
	}
}

func protoStringLabel(tag *telegraf.Tag) *commonpb.StringKeyValue {
	return &commonpb.StringKeyValue{
		Key:   tag.Key,
		Value: tag.Value,
	}
}

func protoResourceAttributes(tags []*telegraf.Tag) []*commonpb.KeyValue {
	ret := make([]*commonpb.KeyValue, len(tags))
	for i := range tags {
		ret[i] = protoLabel(tags[i])
	}
	return ret
}

func protoStringLabels(tags []*telegraf.Tag) []*commonpb.StringKeyValue {
	ret := make([]*commonpb.StringKeyValue, len(tags))
	for i := range tags {
		ret[i] = protoStringLabel(tags[i])
	}
	return ret
}

func (o *OpenTelemetry) protoTimeseries(m telegraf.Metric, f *telegraf.Field) (*metricpb.ResourceMetrics, *metricpb.Metric) {
	metric := &metricpb.Metric{
		Name:        fmt.Sprintf("%s.%s", m.Name(), f.Key),
		Description: "", // TODO
		Unit:        "", // TODO
	}
	return &metricpb.ResourceMetrics{
		Resource: &resourcepb.Resource{
			Attributes: protoResourceAttributes(o.resourceTags),
		},
		InstrumentationLibraryMetrics: []*metricpb.InstrumentationLibraryMetrics{
			{
				InstrumentationLibrary: &commonpb.InstrumentationLibrary{
					Name:    instrumentationLibraryName,
					Version: internal.Version(),
				},
				Metrics: []*metricpb.Metric{metric},
			},
		},
	}, metric
}

func intGauge(labels []*commonpb.StringKeyValue, ts int64, value int64) *metricpb.IntGauge {
	integer := &metricpb.IntDataPoint{
		Labels:       labels,
		TimeUnixNano: uint64(ts),
		Value:        value,
	}
	return &metricpb.IntGauge{
		DataPoints: []*metricpb.IntDataPoint{integer},
	}
}

func doubleGauge(labels []*commonpb.StringKeyValue, ts int64, value float64) *metricpb.DoubleGauge {
	double := &metricpb.DoubleDataPoint{
		Labels:       labels,
		TimeUnixNano: uint64(ts),
		Value:        value,
	}
	return &metricpb.DoubleGauge{
		DataPoints: []*metricpb.DoubleDataPoint{double},
	}
}

// Write the metrics to OTLP destination
func (o *OpenTelemetry) Write(metrics []telegraf.Metric) error {
	batch := sorted(metrics)
	samples := []*metricpb.ResourceMetrics{}
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
					if v <= uint64(maxInt) {
						point.Data = &metricpb.Metric_IntSum{
							IntSum: monotonicIntegerPoint(labels, ts, currentTs, int64(v)),
						}
					} else {
						point.Data = &metricpb.Metric_IntSum{
							IntSum: monotonicIntegerPoint(labels, ts, currentTs, int64(maxInt)),
						}
					}
				case int64:
					point.Data = &metricpb.Metric_IntSum{
						IntSum: monotonicIntegerPoint(labels, ts, currentTs, v),
					}
				case float64:
					point.Data = &metricpb.Metric_DoubleSum{
						DoubleSum: monotonicDoublePoint(labels, ts, currentTs, v),
					}
				case bool:
					if v {
						point.Data = &metricpb.Metric_IntSum{
							IntSum: monotonicIntegerPoint(labels, ts, currentTs, 1),
						}
					} else {
						point.Data = &metricpb.Metric_IntSum{
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
					if v <= uint64(maxInt) {
						point.Data = &metricpb.Metric_IntGauge{
							IntGauge: intGauge(labels, ts, int64(v)),
						}
					} else {
						point.Data = &metricpb.Metric_IntGauge{
							IntGauge: intGauge(labels, ts, int64(maxInt)),
						}
					}
				case int64:
					point.Data = &metricpb.Metric_IntGauge{
						IntGauge: intGauge(labels, ts, v),
					}
				case float64:
					point.Data = &metricpb.Metric_DoubleGauge{
						DoubleGauge: doubleGauge(labels, ts, v),
					}
				case bool:
					if v {
						point.Data = &metricpb.Metric_IntGauge{
							IntGauge: intGauge(labels, ts, 1),
						}
					} else {
						point.Data = &metricpb.Metric_IntGauge{
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
		return fmt.Errorf("unable to write to endpoint: %s", err)
	}
	return nil
}

// Close will terminate the session to the backend, returning error if an issue arises.
func (o *OpenTelemetry) Close() error {
	return o.client.Close()
}

// SampleConfig returns the formatted sample configuration for the plugin.
func (o *OpenTelemetry) SampleConfig() string {
	return sampleConfig
}

// Description returns the human-readable function definition of the plugin.
func (o *OpenTelemetry) Description() string {
	return "Configuration for OpenTelemetry to send metrics to"
}

func newOTLP() *OpenTelemetry {
	return &OpenTelemetry{}
}

func init() {
	outputs.Add("opentelemetry", func() telegraf.Output {
		return newOTLP()
	})
}
