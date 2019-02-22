package loggregator_rlp

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/influxdata/telegraf/internal/tls"

	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type LoggregatorRLPInput struct {
	TlsCommonName           string `toml:"tls_common_name"`
	RlpAddress              string `toml:"rlp_address"`
	InternalMetricsInterval string `toml:"internal_metrics_interval"`
	stopRlpConsumer         context.CancelFunc
	envelopeWriter          *EnvelopeWriter

	tls.ClientConfig
}

func NewLoggregatorRLP() *LoggregatorRLPInput {
	return &LoggregatorRLPInput{
		InternalMetricsInterval: "30s",
	}
}

func (_ *LoggregatorRLPInput) Description() string {
	return "Streams metrics from Loggregator's RLP Endpoint"
}

func (_ *LoggregatorRLPInput) SampleConfig() string {
	return `
  ## A string path to the tls ca certificate
  tls_ca = "/path/to/tls_ca_cert.pem"

  ## A string path to the tls server certificate
  tls_cert = "/path/to/tls_cert.pem"

  ## A string path to the tls server private key
  tls_key = "/path/to/tls_cert.key"

  ## Boolean value indicating whether or not to skip SSL verification
  insecure_skip_verify = false
  
## A string server name that the certificate is valid for
  tls_common_name = "foo"
  
  ## A string address of the RLP server to get logs from
  rlp_address = "bar"

  ## A string duration for how frequently to report internal metrics
  internal_metrics_interval = "30s"
`
}

func (l *LoggregatorRLPInput) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (l *LoggregatorRLPInput) Start(acc telegraf.Accumulator) error {
	internalMetricsInterval, err := time.ParseDuration(l.InternalMetricsInterval)
	if err != nil {
		return err
	}
	envelopeWriter := NewEnvelopeWriter(acc, internalMetricsInterval)
	l.envelopeWriter = envelopeWriter

	tlsConfig, err := l.TLSConfig()
	if err != nil {
		return err
	}
	tlsConfig.ServerName = l.TlsCommonName
	rlpConnector := loggregator.NewEnvelopeStreamConnector(
		l.RlpAddress,
		tlsConfig,
		loggregator.WithEnvelopeStreamLogger(log.New(os.Stderr, "RLP: ", log.LstdFlags)),
	)

	ctx, cancel := context.WithCancel(context.Background())
	l.stopRlpConsumer = cancel

	envelopeStream := rlpConnector.Stream(ctx, &loggregator_v2.EgressBatchRequest{
		Selectors: []*loggregator_v2.Selector{
			{
				Message: &loggregator_v2.Selector_Counter{
					Counter: &loggregator_v2.CounterSelector{},
				},
			},
			{
				Message: &loggregator_v2.Selector_Gauge{
					Gauge: &loggregator_v2.GaugeSelector{},
				},
			},
			{
				Message: &loggregator_v2.Selector_Timer{
					Timer: &loggregator_v2.TimerSelector{},
				},
			},
		},
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, envelope := range envelopeStream() {
					l.envelopeWriter.Write(envelope)
				}
			}
		}
	}()

	return nil
}

func (l *LoggregatorRLPInput) Stop() {
	log.Printf("Info: Stopping RLP listener")
	l.stopRlpConsumer()
}

func init() {
	inputs.Add("loggregator_rlp", func() telegraf.Input {
		return NewLoggregatorRLP()
	})
}
