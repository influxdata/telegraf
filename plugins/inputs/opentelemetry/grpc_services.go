package opentelemetry

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/otel2influx"
	otlpcollectorlogs "github.com/influxdata/influxdb-observability/otlp/collector/logs/v1"
	otlpcollectormetrics "github.com/influxdata/influxdb-observability/otlp/collector/metrics/v1"
	otlpcollectortrace "github.com/influxdata/influxdb-observability/otlp/collector/trace/v1"
)

type traceService struct {
	otlpcollectortrace.UnimplementedTraceServiceServer

	converter *otel2influx.OtelTracesToLineProtocol
	writer    *writeToAccumulator
}

func newTraceService(logger common.Logger, writer *writeToAccumulator) *traceService {
	converter := otel2influx.NewOtelTracesToLineProtocol(logger)
	return &traceService{
		converter: converter,
		writer:    writer,
	}
}

func (s *traceService) Export(ctx context.Context, req *otlpcollectortrace.ExportTraceServiceRequest) (*otlpcollectortrace.ExportTraceServiceResponse, error) {
	err := s.converter.WriteTraces(ctx, req.ResourceSpans, s.writer)
	if err != nil {
		return nil, err
	}
	return &otlpcollectortrace.ExportTraceServiceResponse{}, nil
}

type metricsService struct {
	otlpcollectormetrics.UnimplementedMetricsServiceServer

	converter *otel2influx.OtelMetricsToLineProtocol
	writer    *writeToAccumulator
}

var metricsSchemata = map[string]common.MetricsSchema{
	"prometheus-v1": common.MetricsSchemaTelegrafPrometheusV1,
	"prometheus-v2": common.MetricsSchemaTelegrafPrometheusV2,
}

func newMetricsService(logger common.Logger, writer *writeToAccumulator, schema string) (*metricsService, error) {
	ms, found := metricsSchemata[schema]
	if !found {
		return nil, fmt.Errorf("schema '%s' not recognized", schema)
	}

	converter, err := otel2influx.NewOtelMetricsToLineProtocol(logger, ms)
	if err != nil {
		return nil, err
	}
	return &metricsService{
		converter: converter,
		writer:    writer,
	}, nil
}

func (s *metricsService) Export(ctx context.Context, req *otlpcollectormetrics.ExportMetricsServiceRequest) (*otlpcollectormetrics.ExportMetricsServiceResponse, error) {
	err := s.converter.WriteMetrics(ctx, req.ResourceMetrics, s.writer)
	if err != nil {
		return nil, err
	}
	return &otlpcollectormetrics.ExportMetricsServiceResponse{}, nil
}

type logsService struct {
	otlpcollectorlogs.UnimplementedLogsServiceServer

	converter *otel2influx.OtelLogsToLineProtocol
	writer    *writeToAccumulator
}

func newLogsService(logger common.Logger, writer *writeToAccumulator) *logsService {
	converter := otel2influx.NewOtelLogsToLineProtocol(logger)
	return &logsService{
		converter: converter,
		writer:    writer,
	}
}

func (s *logsService) Export(ctx context.Context, req *otlpcollectorlogs.ExportLogsServiceRequest) (*otlpcollectorlogs.ExportLogsServiceResponse, error) {
	err := s.converter.WriteLogs(ctx, req.ResourceLogs, s.writer)
	if err != nil {
		return nil, err
	}
	return &otlpcollectorlogs.ExportLogsServiceResponse{}, nil
}
