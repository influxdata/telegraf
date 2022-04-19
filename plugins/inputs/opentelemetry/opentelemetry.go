package opentelemetry

import (
	"fmt"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"net"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type OpenTelemetry struct {
	ServiceAddress string `toml:"service_address"`
	MetricsSchema  string `toml:"metrics_schema"`

	tls.ServerConfig
	Timeout config.Duration `toml:"timeout"`

	Log telegraf.Logger `toml:"-"`

	listener   net.Listener // overridden in tests
	grpcServer *grpc.Server

	wg sync.WaitGroup
}

const sampleConfig = `
  ## Override the default (0.0.0.0:4317) destination OpenTelemetry gRPC service
  ## address:port
  # service_address = "0.0.0.0:4317"

  ## Override the default (5s) new connection timeout
  # timeout = "5s"

  ## Override the default (prometheus-v1) metrics schema.
  ## Supports: "prometheus-v1", "prometheus-v2"
  ## For more information about the alternatives, read the Prometheus input
  ## plugin notes.
  # metrics_schema = "prometheus-v1"

  ## Optional TLS Config.
  ## For advanced options: https://github.com/influxdata/telegraf/blob/v1.18.3/docs/TLS.md
  ##
  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections.
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]
  ## Add service certificate and key.
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
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
	var grpcOptions []grpc.ServerOption
	if tlsConfig, err := o.ServerConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		grpcOptions = append(grpcOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	if o.Timeout > 0 {
		grpcOptions = append(grpcOptions, grpc.ConnectionTimeout(time.Duration(o.Timeout)))
	}

	logger := &otelLogger{o.Log}
	influxWriter := &writeToAccumulator{accumulator}
	o.grpcServer = grpc.NewServer(grpcOptions...)

	ptraceotlp.RegisterServer(o.grpcServer, newTraceService(logger, influxWriter))
	ms, err := newMetricsService(logger, influxWriter, o.MetricsSchema)
	if err != nil {
		return err
	}
	pmetricotlp.RegisterServer(o.grpcServer, ms)
	plogotlp.RegisterServer(o.grpcServer, newLogsService(logger, influxWriter))

	if o.listener == nil {
		o.listener, err = net.Listen("tcp", o.ServiceAddress)
		if err != nil {
			return err
		}
	}

	o.wg.Add(1)
	go func() {
		if err := o.grpcServer.Serve(o.listener); err != nil {
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
			MetricsSchema:  "prometheus-v1",
			Timeout:        config.Duration(5 * time.Second),
		}
	})
}
