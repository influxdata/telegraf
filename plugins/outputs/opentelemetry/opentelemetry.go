package opentelemetry

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/influx2otel"
	otlpcollectormetrics "github.com/influxdata/influxdb-observability/otlp/collector/metrics/v1"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"google.golang.org/grpc"
)

type OpenTelemetry struct {
	ServiceAddress string `toml:"service_address"`
	MetricsSchema  string `toml:"metrics_schema"`

	Log telegraf.Logger `toml:"-"`

	metricsConverter     *influx2otel.LineProtocolToOtelMetrics
	grpcClientConn       *grpc.ClientConn
	metricsServiceClient otlpcollectormetrics.MetricsServiceClient
}

func newOpenTelemetry() *OpenTelemetry {
	return &OpenTelemetry{
		ServiceAddress: "localhost:4317",
		MetricsSchema:  "prometheus-v1",
	}
}

const sampleConfig = `
  ## Override the OpenTelemetry gRPC service address:port
  # service_address = "localhost:4317"

  ## Override the default request timeout
  # timeout = "5s"

  ## Select a schema for metrics: prometheus-v1 or prometheus-v2
  ## For more information about the alternatives, read the Prometheus input
  ## plugin notes.
  # metrics_schema = "prometheus-v1"
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

	grpcClientConn, err := grpc.Dial(o.ServiceAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}

	metricsServiceClient := otlpcollectormetrics.NewMetricsServiceClient(grpcClientConn)

	o.metricsConverter = metricsConverter
	o.grpcClientConn = grpcClientConn
	o.metricsServiceClient = metricsServiceClient

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

	_, err := o.metricsServiceClient.Export(context.Background(), req)
	return err
}

func init() {
	outputs.Add("opentelemetry", func() telegraf.Output {
		return newOpenTelemetry()
	})
}
