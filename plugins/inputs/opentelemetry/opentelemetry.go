//go:generate ../../../tools/readme_config_includer/generator
package opentelemetry

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb-observability/otel2influx"
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

type services struct {
	traces   *traceService
	metrics  *metricsService
	logs     *logsService
	profiles *profileService

	// Paths of the signal endpoints for the HTTP transport
	tracesPath   string
	metricsPath  string
	logsPath     string
	profilesPath string
}

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

	endpoint *url.URL
	srvs     services

	listener   net.Listener
	grpcServer *grpc.Server
	httpServer *http.Server

	wg sync.WaitGroup
}

func (*OpenTelemetry) SampleConfig() string {
	return sampleConfig
}

func (o *OpenTelemetry) Init() error {
	if o.ServiceAddress == "" {
		o.ServiceAddress = "grpc://0.0.0.0:4317"
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

	if !strings.Contains(o.ServiceAddress, "://") {
		o.ServiceAddress = "grpc://" + o.ServiceAddress
	}

	u, err := url.Parse(o.ServiceAddress)
	if err != nil {
		return fmt.Errorf("parsing service address failed: %w", err)
	}

	if u.User != nil {
		return errors.New("credentials are not supported in service address")
	}
	if u.Fragment != "" || u.RawQuery != "" {
		return errors.New("query and fragment are not supported")
	}

	port := u.Port()
	switch u.Scheme {
	case "grpc":
		if port == "" {
			port = "4317"
		}
		if u.Path != "" && u.Path != "/" {
			return errors.New("paths are not supported for gRPC")
		}
	case "https":
		if o.TLSCert == "" || o.TLSKey == "" {
			return errors.New("missing TLS settings for https")
		}
		fallthrough
	case "http":
		if port == "" {
			port = "4318"
		}
		if o.MaxMsgSize <= 0 {
			o.MaxMsgSize = config.Size(4 * 1024 * 1024) // 4Mb
		}

		// Construct the paths of the signal endpoints
		basePath := u.EscapedPath()
		if o.srvs.tracesPath, err = url.JoinPath("/", basePath, "v1", "traces"); err != nil {
			return fmt.Errorf("constructing traces path failed: %w", err)
		}
		if o.srvs.metricsPath, err = url.JoinPath("/", basePath, "v1", "metrics"); err != nil {
			return fmt.Errorf("constructing metrics path failed: %w", err)
		}
		if o.srvs.logsPath, err = url.JoinPath("/", basePath, "v1", "logs"); err != nil {
			return fmt.Errorf("constructing logs path failed: %w", err)
		}
		if o.srvs.profilesPath, err = url.JoinPath("/", basePath, "v1development", "profiles"); err != nil {
			return fmt.Errorf("constructing profiles path failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported scheme %q for service address", u.Scheme)
	}
	o.endpoint = &url.URL{
		Scheme: u.Scheme,
		Host:   net.JoinHostPort(u.Hostname(), port),
	}

	return nil
}

func (o *OpenTelemetry) Start(acc telegraf.Accumulator) error {
	tlsConfig, err := o.ServerConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("configuring TLS failed: %w", err)
	}

	// Setup the services shared by all transports
	logger := &otelLogger{o.Log}
	influxWriter := &writeToAccumulator{accumulator: acc, clientCertTags: o.ClientCertTags}

	if o.srvs.traces, err = newTraceService(logger, influxWriter, o.SpanDimensions); err != nil {
		return fmt.Errorf("creating traces service failed: %w", err)
	}
	if o.srvs.metrics, err = newMetricsService(logger, influxWriter, o.MetricsSchema); err != nil {
		return fmt.Errorf("creating metrics service failed: %w", err)
	}
	if o.srvs.logs, err = newLogsService(logger, influxWriter, o.LogRecordDimensions); err != nil {
		return fmt.Errorf("creating logs service failed: %w", err)
	}
	if o.srvs.profiles, err = newProfileService(acc, o.Log, o.ProfileDimensions); err != nil {
		return fmt.Errorf("creating profiles service failed: %w", err)
	}
	o.srvs.profiles.clientCertTags = o.ClientCertTags

	o.listener, err = net.Listen("tcp", o.endpoint.Host)
	if err != nil {
		return fmt.Errorf("listening to %q failed: %w", o.endpoint.Host, err)
	}

	switch o.endpoint.Scheme {
	case "grpc":
		o.startGRPCService(acc, tlsConfig)
	case "http", "https":
		o.startHTTPService(acc, tlsConfig)
	default:
		if err := o.listener.Close(); err != nil {
			o.Log.Errorf("Closing listener failed: %v", err)
		}
		return fmt.Errorf("unsupported scheme %q for service address", o.endpoint.Scheme)
	}

	return nil
}

func (*OpenTelemetry) Gather(telegraf.Accumulator) error {
	return nil
}

func (o *OpenTelemetry) Stop() {
	if o.grpcServer != nil {
		o.grpcServer.Stop()
	}
	if o.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := o.httpServer.Shutdown(ctx); err != nil {
			o.Log.Errorf("Shutting down HTTP service failed: %v", err)
		}
		cancel()
	}
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
