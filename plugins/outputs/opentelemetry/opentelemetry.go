package opentelemetry

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/pkg/errors"

	metricsService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc/metadata"
)

// OpenTelemetry is the OpenTelemetry Protocol config info.
type OpenTelemetry struct {
	Endpoint    string            `toml:"endpoint"`
	Timeout     config.Duration   `toml:"timeout"`
	Compression string            `toml:"compression"`
	Headers     map[string]string `toml:"headers"`
	Attributes  map[string]string `toml:"attributes"`
	Log         telegraf.Logger   `toml:"-"`
	tls.ClientConfig

	client       *client
	resourceTags []*telegraf.Tag
}

const (
	// maxInt is the max int64 value.
	maxInt = int(^uint(0) >> 1)

	defaultEndpoint            = "http://localhost:4317"
	defaultTimeout             = time.Second * 10
	defaultCompression         = "gzip"
	instrumentationLibraryName = "telegraf"
)

var sampleConfig = `
  ## OpenTelemetry endpoint
  # endpoint = "http://localhost:4317"

  ## Timeout when sending data over grpc
  # timeout = "10s"

  ## Compression used to send data, supports: "gzip", "none"
  # compression = "gzip"

  ## Optional TLS Config for use on gRPC connections.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  # Additional resource attributes
  [outputs.opentelemetry.attributes]
    "service.name" = "demo"

  # Additional grpc metadata
  [outputs.opentelemetry.headers]
    key1 = "value1"

`

func (o *OpenTelemetry) Init() error {
	if o.Endpoint == "" {
		o.Endpoint = defaultEndpoint
	}
	endpoint, err := url.Parse(o.Endpoint)
	if err != nil {
		return errors.Wrap(err, "invalid endpoint configured")
	}

	if o.Timeout < config.Duration(time.Second) {
		o.Timeout = config.Duration(defaultTimeout)
	}

	if o.Compression == "" {
		o.Compression = defaultCompression
	}

	for k, v := range o.Attributes {
		o.resourceTags = append(o.resourceTags, &telegraf.Tag{Key: k, Value: v})
	}

	if o.Headers == nil {
		o.Headers = make(map[string]string, 1)
	}

	o.Headers["telemetry-reporting-agent"] = fmt.Sprintf(
		"%s/%s",
		instrumentationLibraryName,
		internal.Version(),
	)

	tlsConfig, err := o.TLSConfig()
	if err != nil {
		return errors.Wrap(err, "invalid tls configuration")
	}

	if o.client == nil {
		o.client = &client{
			logger:     o.Log,
			url:        endpoint,
			timeout:    time.Duration(o.Timeout),
			tlsConfig:  tlsConfig,
			headers:    metadata.New(o.Headers),
			compressor: o.Compression,
		}
	}
	return nil
}

// Connect initiates the primary connection to the OpenTelemetry endpoint.
func (o *OpenTelemetry) Connect() error {
	ctx := context.Background()
	if err := o.client.ping(ctx); err != nil {
		_ = o.client.close()
		return err
	}

	return nil
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
	batch := metrics
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
				default:
					o.Log.Errorf("get type failed: unsupported telegraf value type %v\n", f.Value)
					continue
				}
			default:
				o.Log.Errorf("get type failed: unsupported telegraf metric kind %v\n", m.Type())
				continue
			}
			samples = append(samples, sample)
		}
	}

	if err := o.client.store(&metricsService.ExportMetricsServiceRequest{
		ResourceMetrics: samples,
	}); err != nil {
		return errors.Wrap(err, "unable to write to endpoint")
	}
	return nil
}

// Close will terminate the session to the backend, returning error if an issue arises.
func (o *OpenTelemetry) Close() error {
	return o.client.close()
}

// SampleConfig returns the formatted sample configuration for the plugin.
func (o *OpenTelemetry) SampleConfig() string {
	return sampleConfig
}

// Description returns the human-readable function definition of the plugin.
func (o *OpenTelemetry) Description() string {
	return "Configuration for OpenTelemetry to send metrics to"
}

func init() {
	outputs.Add("opentelemetry", func() telegraf.Output {
		return &OpenTelemetry{}
	})
}
