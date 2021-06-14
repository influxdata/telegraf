package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/influx2otel"
	otlpcollectormetrics "github.com/influxdata/influxdb-observability/otlp/collector/metrics/v1"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type OpenTelemetry struct {
	ServiceAddress string `toml:"service_address"`
	MetricsSchema  string `toml:"metrics_schema"`

	tls.ClientConfig
	Timeout     config.Duration   `toml:"timeout"`
	Compression string            `toml:"compression"`
	Headers     map[string]string `toml:"headers"`
	Attributes  map[string]string `toml:"attributes"`

	Log telegraf.Logger `toml:"-"`

	metricsConverter     *influx2otel.LineProtocolToOtelMetrics
	grpcClientConn       *grpc.ClientConn
	metricsServiceClient otlpcollectormetrics.MetricsServiceClient
	callOptions          []grpc.CallOption
}

const sampleConfig = `
  ## Override the default (localhost:4317) OpenTelemetry gRPC service
  ## address:port
  # service_address = "localhost:4317"

  ## Override the default (5s) request timeout
  # timeout = "5s"

  ## Override the default (prometheus-v1) metrics schema.
  ## Supports: "prometheus-v1", "prometheus-v2"
  ## For more information about the alternatives, read the Prometheus input
  ## plugin notes.
  # metrics_schema = "prometheus-v1"

  ## Optional TLS Config.
  ##
  ## Root certificates for verifying server certificates encoded in PEM format.
  # tls_ca = "/etc/telegraf/ca.pem"
  ## The public and private keypairs for the client encoded in PEM format.
  ## May contain intermediate certificates.
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS, but skip TLS chain and host verification.
  # insecure_skip_verify = false
  ## Send the specified TLS server name via SNI.
  # tls_server_name = "foo.example.com"

  ## Override the default (gzip) compression used to send data.
  ## Supports: "gzip", "none"
  # compression = "gzip"

  ## Additional OpenTelemetry resource attributes
  # [outputs.opentelemetry.attributes]
  # "service.name" = "demo"

  ## Additional gRPC request metadata
  # [outputs.opentelemetry.headers]
  # key1 = "value1"
`

func (o *OpenTelemetry) SampleConfig() string {
	return sampleConfig
}

func (o *OpenTelemetry) Description() string {
	return "Send OpenTelemetry metrics over gRPC"
}

var metricsSchemata = map[string]common.MetricsSchema{
	"prometheus-v1": common.MetricsSchemaTelegrafPrometheusV1,
	"prometheus-v2": common.MetricsSchemaTelegrafPrometheusV2,
}

func (o *OpenTelemetry) Connect() error {
	logger := &otelLogger{o.Log}
	ms, found := metricsSchemata[o.MetricsSchema]
	if !found {
		return fmt.Errorf("schema '%s' not recognized", o.MetricsSchema)
	}

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(logger, ms)
	if err != nil {
		return err
	}

	var grpcTLSDialOption grpc.DialOption
	if tlsConfig, err := o.ClientConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else {
		grpcTLSDialOption = grpc.WithInsecure()
	}

	grpcClientConn, err := grpc.Dial(o.ServiceAddress, grpcTLSDialOption)
	if err != nil {
		return err
	}

	metricsServiceClient := otlpcollectormetrics.NewMetricsServiceClient(grpcClientConn)

	o.metricsConverter = metricsConverter
	o.grpcClientConn = grpcClientConn
	o.metricsServiceClient = metricsServiceClient

	if o.Compression != "" && o.Compression != "none" {
		o.callOptions = append(o.callOptions, grpc.UseCompressor(o.Compression))
	}

	return nil
}

func (o *OpenTelemetry) Close() error {
	return o.grpcClientConn.Close()
}

func (o *OpenTelemetry) Write(metrics []telegraf.Metric) error {
	batch := o.metricsConverter.NewBatch()
	for _, metric := range metrics {
		var vType common.InfluxMetricValueType
		switch metric.Type() {
		case telegraf.Gauge:
			vType = common.InfluxMetricValueTypeGauge
		case telegraf.Untyped:
			vType = common.InfluxMetricValueTypeUntyped
		case telegraf.Counter:
			vType = common.InfluxMetricValueTypeSum
		case telegraf.Histogram:
			vType = common.InfluxMetricValueTypeHistogram
		case telegraf.Summary:
			vType = common.InfluxMetricValueTypeSummary
		default:
			o.Log.Warnf("unrecognized metric type %Q", metric.Type())
			continue
		}
		err := batch.AddPoint(metric.Name(), metric.Tags(), metric.Fields(), metric.Time(), vType)
		if err != nil {
			o.Log.Warnf("failed to add point: %s", err)
			continue
		}
	}
	otlpResourceMetricss := batch.ToProto()
	req := &otlpcollectormetrics.ExportMetricsServiceRequest{
		ResourceMetrics: otlpResourceMetricss,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))
	defer cancel()
	_, err := o.metricsServiceClient.Export(ctx, req, o.callOptions...)
	return err
}

func init() {
	outputs.Add("opentelemetry", func() telegraf.Output {
		return &OpenTelemetry{
			ServiceAddress: "localhost:4317",
			MetricsSchema:  "prometheus-v1",
			Timeout:        config.Duration(5 * time.Second),
			Compression:    "gzip",
		}
	})
}
