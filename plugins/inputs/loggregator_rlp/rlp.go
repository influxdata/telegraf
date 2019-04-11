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
	TlsCommonName           string   `toml:"tls_common_name"`
	RlpAddress              string   `toml:"rlp_address"`
	InternalMetricsInterval string   `toml:"internal_metrics_interval"`
	DiodeBufferSize         int      `toml:"diode_buffer_size"`
	EnvelopeTypes           []string `toml:"envelope_types"`
	SourceIdFilters         []string `toml:"source_id_filters"`
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

  ## Size of diode buffer, this is the limit of how many envelopes will be held in memory before dropping new envelopes
  diode_buffer_size = 100000

  ## Envelope type, an array of one or all of ['counter', 'gauge', 'timer']
  envelope_types = ['counter', 'gauge']

  ## Source Id filters allow the filtering of metrics out of the RLP from only the specified source ids. Should be an array of strings.
  source_id_filters = ['router']
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
		Selectors: l.buildSelectors(),
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

func (l *LoggregatorRLPInput) buildSelectors() []*loggregator_v2.Selector {
	var selectors []*loggregator_v2.Selector

	if len(l.EnvelopeTypes) == 0 {
		panic("No envelope_type selectors provided.")
	}

	if len(l.SourceIdFilters) == 0 {
		for _, envelopeType := range l.EnvelopeTypes {
			selectors = append(selectors, buildEnvelopeTypeSelector(envelopeType))
		}

		return selectors
	}

	for _, sourceId := range l.SourceIdFilters {
		for _, envelopeType := range l.EnvelopeTypes {
			selector := buildEnvelopeTypeSelector(envelopeType)
			selector.SourceId = sourceId
			selectors = append(selectors, selector)
		}
	}
	return selectors
}

func buildEnvelopeTypeSelector(envelopeType string) *loggregator_v2.Selector {
	switch envelopeType {
	case "counter":
		return &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Counter{
				Counter: &loggregator_v2.CounterSelector{},
			},
		}
	case "gauge":
		return &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Gauge{
				Gauge: &loggregator_v2.GaugeSelector{},
			},
		}
	case "timer":
		return &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Timer{
				Timer: &loggregator_v2.TimerSelector{},
			},
		}
	default:
		panic("Invalid envelope type " + envelopeType)
	}
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
