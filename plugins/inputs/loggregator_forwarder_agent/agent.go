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

type LoggregatorForwarderAgentInput struct {
	Port                    uint16 `toml:"port"`
	InternalMetricsInterval string `toml:"internal_metrics_interval"`

	tls.ServerConfig

	envelopeWriter          *EnvelopeWriter
	grpcServer              *grpc.Server
}

func NewLoggregator() *LoggregatorForwarderAgentInput {
	return &LoggregatorForwarderAgentInput{
		InternalMetricsInterval: "30s",
	}
}

func (_ *LoggregatorForwarderAgentInput) Description() string {
	return "Read metrics from a Loggregator Forwarder Agent"
}

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

func (l *LoggregatorForwarderAgentInput) Gather(_ telegraf.Accumulator) error {
	return nil
}

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

func (l *LoggregatorForwarderAgentInput) Stop() {
	log.Printf("Info: Stopping GRPC server")
	l.grpcServer.Stop()
	l.envelopeWriter.Stop()
}

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
