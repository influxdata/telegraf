package cloudfoundry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"code.cloudfoundry.org/go-loggregator/v8"
	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// init registers the input plugin
func init() {
	inputs.Add("cloudfoundry", func() telegraf.Input {
		return &Cloudfoundry{}
	})

}

var usage = `
  ## HTTP gateway URL to the cloudfoundry reverse log proxy gateway
  gateway_address = "https://log-stream.your-cloudfoundry-system-domain"

  ## API URL to the cloudfoundry API endpoint for your platform
  api_address = "https://api.your-cloudfoundry-system-domain"

  ## Username and password for user authentication
  # username = ""
  # password = ""

  ## Client ID and secret for client authentication
  # client_id = ""
  # client_secret ""

  ## Skip verification of TLS certificates (insecure!)
  # insecure_skip_verify = false

  ## retry_interval sets the delay between reconnecting failed stream
  retry_interval = "1s"

  ## Source ID is the GUID of the application or component stream
  ## to connect and collect metrics from
  ##
  ## If unset (default) metrics from ALL platform components will be
  ## collected.
  ##
  ## If you do not have UAA client_id/secret with the "doppler.firehose" or
  ## "logs.admin" scope you MUST set a source_id.
  ##
  source_id = ""

  ## All instances with the same shard_id will receive an exclusive
  ## subset of the data. Use this to avoid duplicating metric collection
  ## when running multiple instances with the same source_id
  # shard_id = "telegraf"

  ## Limit which types of metrics to collect (default: all)
  # types = ["counter", "timer", "gauge", "event", "log"]
`

const (
	Counter = "counter"
	Timer   = "timer"
	Gauge   = "gauge"
	Event   = "event"
	Log     = "log"
)

var (
	validMetricTypes = []string{Counter, Timer, Gauge, Event, Log}
)

type Cloudfoundry struct {
	ShardID       string            `toml:"shard_id"`
	SourceID      string            `toml:"source_id"`
	RetryInterval internal.Duration `toml:"retry_interval"`
	Types         []string          `toml:"types"`
	ClientConfig
	NewClient ClientFunc

	Log      telegraf.Logger
	errs     chan error
	ctx      context.Context
	shutdown context.CancelFunc
	acc      telegraf.Accumulator
	wg       sync.WaitGroup
}

func (_ *Cloudfoundry) Description() string {
	return "Consume metrics and logs from cloudfoundry platform"
}

// SampleConfig returns default configuration example
func (_ *Cloudfoundry) SampleConfig() string {
	return usage
}

// Gather is no-op for service input plugin, see metricWriter
func (s *Cloudfoundry) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Init validates configuration and sets up the client
func (s *Cloudfoundry) Init() error {
	// validate config
	if s.GatewayAddress == "" {
		return fmt.Errorf("must provide a valid gateway_address")
	}
	if s.APIAddress == "" {
		return fmt.Errorf("must provide a valid api_address")
	}
	if (s.Username == "" || s.Password == "") && (s.ClientID == "" || s.ClientSecret == "") {
		return fmt.Errorf("must provide either username/password or client_id/client_secret authentication")
	}
	for _, t := range s.Types {
		isValid := false
		for _, validType := range validMetricTypes {
			if t == validType {
				isValid = true
			}
		}
		if !isValid {
			return fmt.Errorf("invalid metric type '%s' must be one of %v", t, validMetricTypes)
		}
	}
	// create a client
	if s.NewClient == nil {
		s.NewClient = NewClient
	}
	return nil
}

// Start configures client and starts
func (s *Cloudfoundry) Start(acc telegraf.Accumulator) error {
	if s.ShardID == "" {
		s.ShardID = "telegraf"
	}
	if len(s.Types) < 1 {
		s.Types = validMetricTypes
	}
	if s.RetryInterval.Duration < 1 {
		s.RetryInterval.Duration = time.Second * 1
	}
	s.acc = acc
	s.errs = make(chan error)
	s.ctx, s.shutdown = context.WithCancel(context.Background())

	s.wg.Add(2)
	go s.connectStream()
	go s.logStreamErrors()

	return nil
}

// Stop shutsdown all streams
func (s *Cloudfoundry) Stop() {
	s.shutdown()
	s.wg.Wait()
}

// connectStream maintains a connection to the RLP gateway and sends event
// envelopes to the input chan
func (s *Cloudfoundry) connectStream() {
	defer s.wg.Done()

	delay := time.Second * 0
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(delay): // avoid hammering API on failure
			client := s.NewClient(s.ClientConfig, s.errs)
			req := s.newBatchRequest()
			stream := client.Stream(s.ctx, req)
			s.writeEnvelopes(stream)
			delay = s.RetryInterval.Duration
		}
	}
}

// writeEnvelopes reads each event envelope from stream and writes it to acc
func (s *Cloudfoundry) writeEnvelopes(stream loggregator.EnvelopeStream) {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			batch := stream()
			if batch == nil {
				return
			}
			for _, env := range batch {
				if env == nil {
					continue
				}
				select {
				case <-s.ctx.Done():
					return
				default:
					s.writeEnvelope(env)
				}
			}
		}
	}
}

// writeEnvelope converts the envelope to telegraf metric and adds to acc
func (s *Cloudfoundry) writeEnvelope(env *loggregator_v2.Envelope) {
	m, err := NewMetric(env)
	if err != nil {
		s.acc.AddError(err)
		return
	}
	s.acc.AddMetric(m)
}

// newBatchRequest returns a stream configuration for a given sourceID
func (s *Cloudfoundry) newBatchRequest() *loggregator_v2.EgressBatchRequest {
	req := &loggregator_v2.EgressBatchRequest{
		ShardId: s.ShardID,
	}
	for _, t := range s.Types {
		switch t {
		case Log:
			req.Selectors = append(req.Selectors, &loggregator_v2.Selector{
				SourceId: s.SourceID,
				Message: &loggregator_v2.Selector_Log{
					Log: &loggregator_v2.LogSelector{},
				},
			})
		case Counter:
			req.Selectors = append(req.Selectors, &loggregator_v2.Selector{
				SourceId: s.SourceID,
				Message: &loggregator_v2.Selector_Counter{
					Counter: &loggregator_v2.CounterSelector{},
				},
			})
		case Gauge:
			req.Selectors = append(req.Selectors, &loggregator_v2.Selector{
				SourceId: s.SourceID,
				Message: &loggregator_v2.Selector_Gauge{
					Gauge: &loggregator_v2.GaugeSelector{},
				},
			})
		case Timer:
			req.Selectors = append(req.Selectors, &loggregator_v2.Selector{
				SourceId: s.SourceID,
				Message: &loggregator_v2.Selector_Timer{
					Timer: &loggregator_v2.TimerSelector{},
				},
			})
		case Event:
			req.Selectors = append(req.Selectors, &loggregator_v2.Selector{
				SourceId: s.SourceID,
				Message: &loggregator_v2.Selector_Event{
					Event: &loggregator_v2.EventSelector{},
				},
			})
		}
	}
	return req
}

// logStreamErrors writes debug log messages reported from rlp client
func (s *Cloudfoundry) logStreamErrors() {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case err := <-s.errs:
			s.Log.Debugf("rlp: %s", err)
		}
	}
}
