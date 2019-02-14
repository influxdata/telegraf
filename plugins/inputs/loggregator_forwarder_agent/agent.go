package loggregator_forwarder_agent

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"context"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"time"

	"github.com/influxdata/telegraf/internal/tls"
)

// LoggregatorForwarderAgentInput configures the Loggregator ingress API
type LoggregatorForwarderAgentInput struct {
	Port                    uint16 `toml:"port"`
	InternalMetricsInterval string `toml:"internal_metrics_interval"`

	tls.ServerConfig

	envelopeWriter *EnvelopeWriter
	grpcServer     *grpc.Server
}

// NewLoggregator creates a default LoggregatorForwarderAgentInput
func NewLoggregator() *LoggregatorForwarderAgentInput {
	return &LoggregatorForwarderAgentInput{
		InternalMetricsInterval: "30s",
	}
}

// Description returns the description of this plugin
func (_ *LoggregatorForwarderAgentInput) Description() string {
	return "Read metrics from a Loggregator Forwarder Agent"
}

// SampleConfig returns a sample configuration for this plugin
func (_ *LoggregatorForwarderAgentInput) SampleConfig() string {
	return `
  ## A uint16 port for the LoggregatorForwarderAgentInput Ingress server to listen on
  port = 13322

  ## A string path to the tls ca certificate
  tls_allowed_cacerts = [ "/path/to/tls_ca_cert.pem" ]

  ## A string path to the tls server certificate
  tls_cert = "/path/to/tls_cert.pem"

  ## A string path to the tls server private key
  tls_key = "/path/to/tls_cert.key"
	
  ## A string duration for how frequently to report internal metrics
  internal_metrics_interval = "30s"
`
}

// Gather is a no-op to conform with the Input interface
func (l *LoggregatorForwarderAgentInput) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Start begins collecting metrics from the Loggregator Forwarder Agent
func (l *LoggregatorForwarderAgentInput) Start(acc telegraf.Accumulator) error {
	internalMetricsInterval, err := time.ParseDuration(l.InternalMetricsInterval)
	if err != nil {
		return err
	}

	envelopeWriter := NewEnvelopeWriter(acc, internalMetricsInterval)
	l.envelopeWriter = envelopeWriter

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", l.Port))
	if err != nil {
		return err
	}
	log.Printf("grpc bound to: %s", lis.Addr())

	if len(l.TLSAllowedCACerts) == 0 || l.TLSCert == "" || l.TLSKey == "" {
		l.grpcServer = grpc.NewServer()
	} else {
		l.grpcServer = grpc.NewServer(l.buildTLSCreds())
	}

	loggregator_v2.RegisterIngressServer(l.grpcServer, l)

	go func() {
		if err := l.grpcServer.Serve(lis); err != nil {
			log.Println(err.Error())
		}
	}()
	return nil
}

func (l *LoggregatorForwarderAgentInput) buildTLSCreds() grpc.ServerOption {
	tlsConfig, err := l.TLSConfig()
	if err != nil {
		panic(err)
	}
	return grpc.Creds(credentials.NewTLS(tlsConfig))
}

// Stop stops collecting metrics from the Loggregator Forwarder Agent
func (l *LoggregatorForwarderAgentInput) Stop() {
	log.Printf("Info: Stopping GRPC server")
	l.grpcServer.Stop()
	l.envelopeWriter.Stop()
}

// Sender allows forwarding from the Loggregator Forwarder Agent via GRPC. Needed to implement Loggregator v2 API.
func (l *LoggregatorForwarderAgentInput) Sender(sender loggregator_v2.Ingress_SenderServer) error {
	for {
		env, err := sender.Recv()
		if err != nil {
			log.Printf("Failed to receive data: %s", err)
			return err
		}

		l.envelopeWriter.Write(env)
	}
}

// BatchSender allows batch forwarding from the Loggregator Forwarder Agent via GRPC. Needed to implement Loggregator v2 API.
func (l *LoggregatorForwarderAgentInput) BatchSender(sender loggregator_v2.Ingress_BatchSenderServer) error {
	for {
		envelopes, err := sender.Recv()
		if err != nil {
			log.Printf("Failed to receive data: %s", err)
			return err
		}

		for _, e := range envelopes.Batch {
			l.envelopeWriter.Write(e)
		}
	}
}

// Send allows forwarding from the Loggregator Forwarder Agent via GRPC. Needed to implement Loggregator v2 API.
func (l *LoggregatorForwarderAgentInput) Send(ctx context.Context, batch *loggregator_v2.EnvelopeBatch) (*loggregator_v2.SendResponse, error) {
	for _, e := range batch.GetBatch() {
		l.envelopeWriter.Write(e)
	}

	return &loggregator_v2.SendResponse{}, nil
}

func init() {
	inputs.Add("loggregator_forwarder_agent", func() telegraf.Input {
		return NewLoggregator()
	})
}
