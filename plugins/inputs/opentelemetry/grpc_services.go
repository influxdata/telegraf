package opentelemetry

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/otel2influx"
	"go.opentelemetry.io/collector/model/otlpgrpc"
	"go.opentelemetry.io/collector/model/pdata"
)

type traceService struct {
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

func (s *traceService) Export(ctx context.Context, req pdata.Traces) (otlpgrpc.TracesResponse, error) {
	err := s.converter.WriteTraces(ctx, req, s.writer)
	return otlpgrpc.NewTracesResponse(), err
}

type metricsService struct {
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

func (s *metricsService) Export(ctx context.Context, req pdata.Metrics) (otlpgrpc.MetricsResponse, error) {
	err := s.converter.WriteMetrics(ctx, req, s.writer)
	return otlpgrpc.NewMetricsResponse(), err
}

type logsService struct {
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

func (s *logsService) Export(ctx context.Context, req pdata.Logs) (otlpgrpc.LogsResponse, error) {
	err := s.converter.WriteLogs(ctx, req, s.writer)
	return otlpgrpc.NewLogsResponse(), err
}
