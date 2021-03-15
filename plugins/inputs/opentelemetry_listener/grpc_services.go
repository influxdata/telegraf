package opentelemetry_listener

import (
	"context"

	"github.com/influxdata/influxdb-observability/otel2influx"
	otlpcollectorlogs "github.com/influxdata/influxdb-observability/otlp/collector/logs/v1"
	otlpcollectormetrics "github.com/influxdata/influxdb-observability/otlp/collector/metrics/v1"
	otlpcollectortrace "github.com/influxdata/influxdb-observability/otlp/collector/trace/v1"
)

type grpcServiceBase struct {
	influxConverter *otel2influx.OpenTelemetryToInfluxConverter
	influxWriter    *writer
}

type traceService struct {
	otlpcollectortrace.UnimplementedTraceServiceServer
	grpcServiceBase
}

func newTraceService(influxConverter *otel2influx.OpenTelemetryToInfluxConverter, influxWriter *writer) *traceService {
	return &traceService{
		grpcServiceBase: grpcServiceBase{
			influxConverter: influxConverter,
			influxWriter:    influxWriter,
		},
	}
}

func (s *traceService) Export(ctx context.Context, req *otlpcollectortrace.ExportTraceServiceRequest) (*otlpcollectortrace.ExportTraceServiceResponse, error) {
	_ = s.influxConverter.WriteTraces(ctx, req.ResourceSpans, s.influxWriter)
	return &otlpcollectortrace.ExportTraceServiceResponse{}, nil
}

type metricsService struct {
	otlpcollectormetrics.UnimplementedMetricsServiceServer
	grpcServiceBase
}

func newMetricsService(influxConverter *otel2influx.OpenTelemetryToInfluxConverter, influxWriter *writer) *metricsService {
	return &metricsService{
		grpcServiceBase: grpcServiceBase{
			influxConverter: influxConverter,
			influxWriter:    influxWriter,
		},
	}
}

func (s *metricsService) Export(ctx context.Context, req *otlpcollectormetrics.ExportMetricsServiceRequest) (*otlpcollectormetrics.ExportMetricsServiceResponse, error) {
	_ = s.influxConverter.WriteMetrics(ctx, req.ResourceMetrics, s.influxWriter)
	return &otlpcollectormetrics.ExportMetricsServiceResponse{}, nil
}

type logsService struct {
	otlpcollectorlogs.UnimplementedLogsServiceServer
	grpcServiceBase
}

func newLogsService(influxConverter *otel2influx.OpenTelemetryToInfluxConverter, influxWriter *writer) *logsService {
	return &logsService{
		grpcServiceBase: grpcServiceBase{
			influxConverter: influxConverter,
			influxWriter:    influxWriter,
		},
	}
}

func (s *logsService) Export(ctx context.Context, req *otlpcollectorlogs.ExportLogsServiceRequest) (*otlpcollectorlogs.ExportLogsServiceResponse, error) {
	_ = s.influxConverter.WriteLogs(ctx, req.ResourceLogs, s.influxWriter)
	return &otlpcollectorlogs.ExportLogsServiceResponse{}, nil
}
