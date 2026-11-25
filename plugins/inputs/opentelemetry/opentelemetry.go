//go:generate ../../../tools/readme_config_includer/generator
package opentelemetry

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/influxdata/influxdb-observability/otel2influx"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	pprofileotlp "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const clientCertTagPrefix = "tls_client_"

type OpenTelemetry struct {
	ServiceAddress      string          `toml:"service_address"`
	SpanDimensions      []string        `toml:"span_dimensions"`
	LogRecordDimensions []string        `toml:"log_record_dimensions"`
	ProfileDimensions   []string        `toml:"profile_dimensions"`
	MetricsSchema       string          `toml:"metrics_schema"`
	MaxMsgSize          config.Size     `toml:"max_msg_size"`
	Timeout             config.Duration `toml:"timeout"`
	ClientCertTags      []string        `toml:"tls_client_certificate_tags"`
	Log                 telegraf.Logger `toml:"-"`
	tls.ServerConfig

	listener   net.Listener // overridden in tests
	grpcServer *grpc.Server

	wg sync.WaitGroup
}

func (*OpenTelemetry) SampleConfig() string {
	return sampleConfig
}

func (o *OpenTelemetry) Init() error {
	if o.ServiceAddress == "" {
		o.ServiceAddress = "0.0.0.0:4317"
	}
	switch o.MetricsSchema {
	case "": // Set default
		o.MetricsSchema = "prometheus-v1"
	case "prometheus-v1", "prometheus-v2": // Valid values
	default:
		return fmt.Errorf("invalid metric schema %q", o.MetricsSchema)
	}

	// Validate the requested certificate attributes
	if err := tls.ValidateTagNames(o.ClientCertTags); err != nil {
		return fmt.Errorf("invalid client certificate tags: %w", err)
	}
	if len(o.ClientCertTags) > 0 && len(o.TLSAllowedCACerts) == 0 {
		return errors.New("'tls_client_certificate_tags' requires 'tls_allowed_cacerts' to be set as client certificates are only requested for mutual TLS")
	}

	return nil
}

func (o *OpenTelemetry) Start(acc telegraf.Accumulator) error {
	var grpcOptions []grpc.ServerOption
	if tlsConfig, err := o.ServerConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		grpcOptions = append(grpcOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	if o.Timeout > 0 {
		grpcOptions = append(grpcOptions, grpc.ConnectionTimeout(time.Duration(o.Timeout)))
	}
	if o.MaxMsgSize > 0 {
		grpcOptions = append(grpcOptions, grpc.MaxRecvMsgSize(int(o.MaxMsgSize)))
	}

	logger := &otelLogger{o.Log}
	influxWriter := &writeToAccumulator{accumulator: acc, clientCertTags: o.ClientCertTags}
	o.grpcServer = grpc.NewServer(grpcOptions...)

	traceSvc, err := newTraceService(logger, influxWriter, o.SpanDimensions)
	if err != nil {
		return err
	}
	ptraceotlp.RegisterGRPCServer(o.grpcServer, traceSvc)

	metricsSvc, err := newMetricsService(logger, influxWriter, o.MetricsSchema)
	if err != nil {
		return err
	}
	pmetricotlp.RegisterGRPCServer(o.grpcServer, metricsSvc)

	logsSvc, err := newLogsService(logger, influxWriter, o.LogRecordDimensions)
	if err != nil {
		return err
	}
	plogotlp.RegisterGRPCServer(o.grpcServer, logsSvc)

	profileSvc, err := newProfileService(acc, o.Log, o.ProfileDimensions)
	if err != nil {
		return err
	}
	pprofileotlp.RegisterProfilesServiceServer(o.grpcServer, profileSvc)

	profileSvc.clientCertTags = o.ClientCertTags

	o.listener, err = net.Listen("tcp", o.ServiceAddress)
	if err != nil {
		return err
	}

	o.wg.Go(func() {
		if err := o.grpcServer.Serve(o.listener); err != nil {
			acc.AddError(fmt.Errorf("failed to stop OpenTelemetry gRPC service: %w", err))
		}
	})

	return nil
}

func (*OpenTelemetry) Gather(telegraf.Accumulator) error {
	return nil
}

func (o *OpenTelemetry) Stop() {
	if o.grpcServer != nil {
		o.grpcServer.Stop()
	}
	o.listener = nil

	o.wg.Wait()
}

// clientCertTags returns the requested tags, prefixed with "tls_client_",
// derived from the certificate the client presented during the (mutual) TLS
// handshake. It returns nil if there is no such certificate, e.g. for
// plain-text or server-only TLS connections. The names must be validated
// beforehand, see Init().
func clientCertTags(ctx context.Context, names []string) map[string]string {
	if len(names) == 0 {
		return nil
	}
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}
	info, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || len(info.State.PeerCertificates) == 0 {
		return nil
	}
	return tls.CertificateTags(info.State.PeerCertificates[0], names, clientCertTagPrefix)
}

// removeClientCertTags removes all tags the plugin might derive from the client
// certificate for the given names. This must be done for all names, not only
// the ones with a value in the certificate, to prevent the sender from
// providing the tags itself.
func removeClientCertTags(tags map[string]string, names []string) {
	for _, name := range names {
		delete(tags, clientCertTagPrefix+name)
	}
}

func init() {
	inputs.Add("opentelemetry", func() telegraf.Input {
		return &OpenTelemetry{
			SpanDimensions:      otel2influx.DefaultOtelTracesToLineProtocolConfig().SpanDimensions,
			LogRecordDimensions: otel2influx.DefaultOtelLogsToLineProtocolConfig().LogRecordDimensions,
			Timeout:             config.Duration(5 * time.Second),
		}
	})
}
