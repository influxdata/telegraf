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

	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
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
	maxInt = int64(^uint(0) >> 1)

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

// Write the metrics to OTLP destination
func (o *OpenTelemetry) Write(metrics []telegraf.Metric) error {
	batch := metrics
	samples := []*metricspb.ResourceMetrics{}
	currentTs := time.Now().UnixNano()
	for _, m := range batch {
		for _, f := range m.FieldList() {
			sample, point := protoTimeseries(o.resourceTags, m, f)
			labels := protoStringLabels(m.TagList())
			ts := m.Time().UnixNano()

			switch m.Type() {
			case telegraf.Counter:
				switch v := f.Value.(type) {
				case uint64:
					val := maxInt
					if v <= uint64(maxInt) {
						val = int64(v)
					}
					point.Data = &metricspb.Metric_IntSum{
						IntSum: monotonicIntegerPoint(labels, ts, currentTs, val),
					}
				case int64:
					point.Data = &metricspb.Metric_IntSum{
						IntSum: monotonicIntegerPoint(labels, ts, currentTs, v),
					}
				case float64:
					point.Data = &metricspb.Metric_DoubleSum{
						DoubleSum: monotonicDoublePoint(labels, ts, currentTs, v),
					}
				case bool:
					val := int64(0)
					if v {
						val = 1
					}
					point.Data = &metricspb.Metric_IntSum{
						IntSum: monotonicIntegerPoint(labels, ts, currentTs, val),
					}
				default:
					o.Log.Errorf("get type failed: unsupported telegraf value type %v\n", f.Value)
					continue
				}
			case telegraf.Gauge, telegraf.Untyped:
				switch v := f.Value.(type) {
				case uint64:
					val := maxInt
					if v <= uint64(maxInt) {
						val = int64(v)
					}
					point.Data = &metricspb.Metric_IntGauge{
						IntGauge: intGauge(labels, ts, val),
					}
				case int64:
					point.Data = &metricspb.Metric_IntGauge{
						IntGauge: intGauge(labels, ts, v),
					}
				case float64:
					point.Data = &metricspb.Metric_DoubleGauge{
						DoubleGauge: doubleGauge(labels, ts, v),
					}
				case bool:
					val := int64(0)
					if v {
						val = 1
					}
					point.Data = &metricspb.Metric_IntGauge{
						IntGauge: intGauge(labels, ts, val),
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

	if err := o.client.store(samples); err != nil {
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
