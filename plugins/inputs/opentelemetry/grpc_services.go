package opentelemetry

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/otel2influx"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

type traceService struct {
	ptraceotlp.UnimplementedGRPCServer
	converter *otel2influx.OtelTracesToLineProtocol
}

var _ ptraceotlp.GRPCServer = (*traceService)(nil)

func newTraceService(logger common.Logger, writer *writeToAccumulator) (*traceService, error) {
	converter, err := otel2influx.NewOtelTracesToLineProtocol(logger, writer)
	if err != nil {
		return nil, err
	}

	return &traceService{
		converter: converter,
	}, nil
}

func (s *traceService) Export(ctx context.Context, req ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	err := s.converter.WriteTraces(ctx, req.Traces())
	return ptraceotlp.NewExportResponse(), err
}

type metricsService struct {
	pmetricotlp.UnimplementedGRPCServer
	converter *otel2influx.OtelMetricsToLineProtocol
}

var _ pmetricotlp.GRPCServer = (*metricsService)(nil)

var metricsSchemata = map[string]common.MetricsSchema{
	"prometheus-v1": common.MetricsSchemaTelegrafPrometheusV1,
	"prometheus-v2": common.MetricsSchemaTelegrafPrometheusV2,
}

func newMetricsService(logger common.Logger, writer *writeToAccumulator, schema string) (*metricsService, error) {
	ms, found := metricsSchemata[schema]
	if !found {
		return nil, fmt.Errorf("schema %q not recognized", schema)
	}

	converter, err := otel2influx.NewOtelMetricsToLineProtocol(logger, writer, ms)
	if err != nil {
		return nil, err
	}
	return &metricsService{
		converter: converter,
	}, nil
}

func (s *metricsService) Export(ctx context.Context, req pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	err := s.converter.WriteMetrics(ctx, req.Metrics())
	return pmetricotlp.NewExportResponse(), err
}

type logsService struct {
	plogotlp.UnimplementedGRPCServer
	converter *otel2influx.OtelLogsToLineProtocol
}

var _ plogotlp.GRPCServer = (*logsService)(nil)

func newLogsService(logger common.Logger, writer *writeToAccumulator) *logsService {
	converter := otel2influx.NewOtelLogsToLineProtocol(logger, writer)
	return &logsService{
		converter: converter,
	}
}

func (s *logsService) Export(ctx context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	err := s.converter.WriteLogs(ctx, req.Logs())
	return plogotlp.NewExportResponse(), err
}
