package opentelemetry

import (
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

	grpcServer    *grpc.Server
	grpcServerErr error
	listener      net.Listener

	wg sync.WaitGroup

	Log telegraf.Logger `toml:"-"`
}

const sampleConfig = `
  ## Override the OpenTelemetry gRPC service address:port 
  # service_address = "0.0.0.0:4317"

  ## Override the default request timeout
  # timeout = "1s"

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
	var err error
	o.listener, err = net.Listen("tcp", o.ServiceAddress)
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
		o.grpcServerErr = o.grpcServer.Serve(o.listener)
		o.wg.Done()
	}()

	return nil
}

func (o *OpenTelemetry) Stop() {
	if o.grpcServer != nil {
		o.grpcServer.Stop()
	}
	var listenerErr error
	if o.listener != nil {
		listenerErr = o.listener.Close()
	}

	o.wg.Wait()

	if o.grpcServerErr != nil {
		o.Log.Warn("failed to stop OpenTelemetry gRPC service: %q", o.grpcServerErr)
	}
	if listenerErr != nil {
		o.Log.Warn("failed to stop OpenTelemetry net.Listener: %q", listenerErr)
	}
}

func init() {
	inputs.Add("opentelemetry", func() telegraf.Input {
		return &OpenTelemetry{
			ServiceAddress: "0.0.0.0:4317",
			Timeout:        config.Duration(time.Second),
		}
	})
}
