package opentelemetry

import (
	"fmt"
	"net"
	"sync"
	"time"

	otlpcollectorlogs "github.com/influxdata/influxdb-observability/otlp/collector/logs/v1"
	otlpcollectormetrics "github.com/influxdata/influxdb-observability/otlp/collector/metrics/v1"
	otlpcollectortrace "github.com/influxdata/influxdb-observability/otlp/collector/trace/v1"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"google.golang.org/grpc"
)

type OpenTelemetry struct {
	ServiceAddress string          `toml:"service_address"`
	Timeout        config.Duration `toml:"timeout"`

	MetricsSchema string `toml:"metrics_schema"`

	Log telegraf.Logger `toml:"-"`

	grpcServer *grpc.Server

	wg sync.WaitGroup
}

const sampleConfig = `
  ## Override the OpenTelemetry gRPC service address:port
  # service_address = "0.0.0.0:4317"

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
	return "Receive OpenTelemetry traces, metrics, and logs over gRPC"
}

func (o *OpenTelemetry) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (o *OpenTelemetry) Start(accumulator telegraf.Accumulator) error {
	listener, err := net.Listen("tcp", o.ServiceAddress)
	if err != nil {
		return err
	}

	logger := &otelLogger{o.Log}
	influxWriter := &writeToAccumulator{accumulator}
	o.grpcServer = grpc.NewServer()

	otlpcollectortrace.RegisterTraceServiceServer(o.grpcServer, newTraceService(logger, influxWriter))
	ms, err := newMetricsService(logger, influxWriter, o.MetricsSchema)
	if err != nil {
		return err
	}
	otlpcollectormetrics.RegisterMetricsServiceServer(o.grpcServer, ms)
	otlpcollectorlogs.RegisterLogsServiceServer(o.grpcServer, newLogsService(logger, influxWriter))

	o.wg.Add(1)
	go func() {
		if err := o.grpcServer.Serve(listener); err != nil {
			accumulator.AddError(fmt.Errorf("failed to stop OpenTelemetry gRPC service: %w", err))
		}
		o.wg.Done()
	}()

	return nil
}

func (o *OpenTelemetry) Stop() {
	if o.grpcServer != nil {
		o.grpcServer.Stop()
	}

	o.wg.Wait()
}

func init() {
	inputs.Add("opentelemetry", func() telegraf.Input {
		return &OpenTelemetry{
			ServiceAddress: "0.0.0.0:4317",
			Timeout:        config.Duration(5 * time.Second),
			MetricsSchema:  "prometheus-v1",
		}
	})
}
