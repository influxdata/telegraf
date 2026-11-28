package opentelemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/otel2influx"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	pprofileotlp "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/influxdata/telegraf"
)

func (o *OpenTelemetry) startGRPCService(acc telegraf.Accumulator, tlsCfg *tls.Config) {
	var grpcOptions []grpc.ServerOption
	if tlsCfg != nil {
		grpcOptions = append(grpcOptions, grpc.Creds(credentials.NewTLS(tlsCfg)))
	}
	if o.Timeout > 0 {
		grpcOptions = append(grpcOptions, grpc.ConnectionTimeout(time.Duration(o.Timeout)))
	}
	if o.MaxMsgSize > 0 {
		grpcOptions = append(grpcOptions, grpc.MaxRecvMsgSize(int(o.MaxMsgSize)))
	}

	o.grpcServer = grpc.NewServer(grpcOptions...)
	ptraceotlp.RegisterGRPCServer(o.grpcServer, o.srvs.traces)
	pmetricotlp.RegisterGRPCServer(o.grpcServer, o.srvs.metrics)
	plogotlp.RegisterGRPCServer(o.grpcServer, o.srvs.logs)
	pprofileotlp.RegisterProfilesServiceServer(o.grpcServer, o.srvs.profiles)

	o.wg.Go(func() {
		if err := o.grpcServer.Serve(o.listener); err != nil {
			acc.AddError(fmt.Errorf("serving OpenTelemetry gRPC service failed: %w", err))
		}
	})
}

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

// Export processes and exports the trace data received in the request.
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

// Export processes and exports the metrics data received in the request.
func (s *metricsService) Export(ctx context.Context, req pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	err := s.exporter.WriteMetrics(ctx, req.Metrics())
	return pmetricotlp.NewExportResponse(), err
}

type logsService struct {
	plogotlp.UnimplementedGRPCServer
	converter *otel2influx.OtelLogsToLineProtocol
}

var _ plogotlp.GRPCServer = (*logsService)(nil)

func newLogsService(logger common.Logger, writer *writeToAccumulator, logRecordDimensions []string) (*logsService, error) {
	expConfig := otel2influx.DefaultOtelLogsToLineProtocolConfig()
	expConfig.Logger = logger
	expConfig.Writer = writer
	expConfig.LogRecordDimensions = logRecordDimensions
	exp, err := otel2influx.NewOtelLogsToLineProtocol(expConfig)
	if err != nil {
		return nil, err
	}
	return &logsService{
		converter: exp,
	}, nil
}

// Export processes and exports the logs data received in the request.
func (s *logsService) Export(ctx context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	err := s.converter.WriteLogs(ctx, req.Logs())
	return plogotlp.NewExportResponse(), err
}
