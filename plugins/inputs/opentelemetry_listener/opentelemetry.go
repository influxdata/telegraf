package opentelemetry_listener

import (
	"net"
	"sync"

	"github.com/influxdata/influxdb-observability/otel2influx"
	otlpcollectorlogs "github.com/influxdata/influxdb-observability/otlp/collector/logs/v1"
	otlpcollectormetrics "github.com/influxdata/influxdb-observability/otlp/collector/metrics/v1"
	otlpcollectortrace "github.com/influxdata/influxdb-observability/otlp/collector/trace/v1"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"google.golang.org/grpc"
)

type OpenTelemetryListener struct {
	ServiceAddress string `toml:"service_address"`

	grpcServer    *grpc.Server
	grpcServerErr error
	listener      net.Listener

	wg sync.WaitGroup

	Log telegraf.Logger `toml:"-"`
}

const sampleConfig = `
  ## Address and port to host OpenTelemetry listener on
  service_address = ":4317"
`

func (o *OpenTelemetryListener) SampleConfig() string {
	return sampleConfig
}

func (o *OpenTelemetryListener) Description() string {
	return "Accept OpenTelemetry traces, metrics, and logs over gRPC and HTTP"
}

func (o *OpenTelemetryListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (o *OpenTelemetryListener) Start(accumulator telegraf.Accumulator) error {
	var err error
	o.listener, err = net.Listen("tcp", o.ServiceAddress)
	if err != nil {
		return err
	}

	influxConverter := otel2influx.NewOpenTelemetryToInfluxConverter(&otelLogger{o.Log})
	influxWriter := &writer{accumulator}
	o.grpcServer = grpc.NewServer()
	otlpcollectortrace.RegisterTraceServiceServer(o.grpcServer, newTraceService(influxConverter, influxWriter))
	otlpcollectormetrics.RegisterMetricsServiceServer(o.grpcServer, newMetricsService(influxConverter, influxWriter))
	otlpcollectorlogs.RegisterLogsServiceServer(o.grpcServer, newLogsService(influxConverter, influxWriter))

	o.wg.Add(1)
	go func() {
		o.grpcServerErr = o.grpcServer.Serve(o.listener)
		o.wg.Done()
	}()

	return nil
}

func (o *OpenTelemetryListener) Stop() {
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
	inputs.Add("opentelemetry_listener", func() telegraf.Input {
		return &OpenTelemetryListener{
		}
	})
}
