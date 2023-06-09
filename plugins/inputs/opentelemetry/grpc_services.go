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
	exporter *otel2influx.OtelTracesToLineProtocol
}

var _ ptraceotlp.GRPCServer = (*traceService)(nil)

func newTraceService(logger common.Logger, writer *writeToAccumulator, spanDimensions []string) (*traceService, error) {
	expConfig := otel2influx.DefaultOtelTracesToLineProtocolConfig()
	expConfig.Logger = logger
	expConfig.Writer = writer
	expConfig.SpanDimensions = spanDimensions
	exp, err := otel2influx.NewOtelTracesToLineProtocol(expConfig)
	if err != nil {
		return nil, err
	}
	return &traceService{
		exporter: exp,
	}, nil
}

func (s *traceService) Export(ctx context.Context, req ptraceotlp.ExportRequest) (ptraceotlp.ExportResponse, error) {
	err := s.exporter.WriteTraces(ctx, req.Traces())
	return ptraceotlp.NewExportResponse(), err
}

type metricsService struct {
	pmetricotlp.UnimplementedGRPCServer
	exporter *otel2influx.OtelMetricsToLineProtocol
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

	expConfig := otel2influx.DefaultOtelMetricsToLineProtocolConfig()
	expConfig.Logger = logger
	expConfig.Writer = writer
	expConfig.Schema = ms
	exp, err := otel2influx.NewOtelMetricsToLineProtocol(expConfig)
	if err != nil {
		return nil, err
	}
	return &metricsService{
		exporter: exp,
	}, nil
}

func (s *metricsService) Export(ctx context.Context, req pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	err := s.exporter.WriteMetrics(ctx, req.Metrics())
	return pmetricotlp.NewExportResponse(), err
}

type logsService struct {
	plogotlp.UnimplementedGRPCServer
	converter *otel2influx.OtelLogsToLineProtocol
}

var _ plogotlp.GRPCServer = (*logsService)(nil)

func newLogsService(logger common.Logger, writer *writeToAccumulator) (*logsService, error) {
	expConfig := otel2influx.DefaultOtelLogsToLineProtocolConfig()
	expConfig.Logger = logger
	expConfig.Writer = writer
	exp, err := otel2influx.NewOtelLogsToLineProtocol(expConfig)
	if err != nil {
		return nil, err
	}
	return &logsService{
		converter: exp,
	}, nil
}

func (s *logsService) Export(ctx context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	err := s.converter.WriteLogs(ctx, req.Logs())
	return plogotlp.NewExportResponse(), err
}
