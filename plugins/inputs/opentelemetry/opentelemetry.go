//go:generate ../../../tools/readme_config_includer/generator
package opentelemetry

import (
	_ "embed"
	"fmt"
	"net"
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

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

func (*OpenTelemetry) SampleConfig() string {
	return sampleConfig
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

	traceService, err := newTraceService(logger, influxWriter)
	if err != nil {
		return err
	}
	ptraceotlp.RegisterGRPCServer(o.grpcServer, traceService)
	ms, err := newMetricsService(logger, influxWriter, o.MetricsSchema)
	if err != nil {
		return err
	}
	pmetricotlp.RegisterGRPCServer(o.grpcServer, ms)
	plogotlp.RegisterGRPCServer(o.grpcServer, newLogsService(logger, influxWriter))

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
