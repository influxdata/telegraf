package sflow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/sflow/parser/decoder"
)

const sampleConfig = `
  ## URL to listen on
  # service_address = "udp://:6343"
  # service_address = "udp4://:6343"
  # service_address = "udp6://:6343"

  ## Maximum socket buffer size (in bytes when no unit specified).
  ## For stream sockets, once the buffer fills up, the sender will start backing up.
  ## For datagram sockets, once the buffer fills up, metrics will start dropping.
  ## Defaults to the OS default.
  # read_buffer_size = "64KiB"
`

const (
	maxPacketSize = 64 * 1024
)

type SFlow struct {
	ServiceAddress         string        `toml:"service_address"`
	ReadBufferSize         internal.Size `toml:"read_buffer_size"`
	MaxFlowsPerSample      uint32        `toml:"max_flows_per_sample"`
	MaxCountersPerSample   uint32        `toml:"max_counters_per_sample"`
	MaxSamplesPerPacket    uint32        `toml:"max_samples_per_packet"`
	MaxSampleLength        uint32        `toml:"max_sample_length"`
	MaxFlowHeaderLength    uint32        `toml:"max_flow_header_length"`
	MaxCounterHeaderLength uint32        `toml:"max_counter_header_length"`

	Log telegraf.Logger `toml:"-"`

	addr        net.Addr
	decoder     *decoder.DecodeContext
	decoderOpts decoder.Directive
	closer      io.Closer
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// Description answers a description of this input plugin
func (s *SFlow) Description() string {
	return "SFlow V5 Protocol Listener"
}

// SampleConfig answers a sample configuration
func (s *SFlow) SampleConfig() string {
	return sampleConfig
}

func (s *SFlow) Init() error {
	s.decoder = decoder.NewDecodeContext(false)
	s.decoderOpts = V5Format(s.getSflowConfig())
	return nil
}

func (s *SFlow) getSflowConfig() V5FormatOptions {
	sflowConfig := NewDefaultV5FormatOptions()
	if s.MaxFlowsPerSample > 0 {
		sflowConfig.MaxFlowsPerSample = s.MaxFlowsPerSample
	}
	if s.MaxCountersPerSample > 0 {
		sflowConfig.MaxCountersPerSample = s.MaxCountersPerSample
	}
	if s.MaxSamplesPerPacket > 0 {
		sflowConfig.MaxSamplesPerPacket = s.MaxSamplesPerPacket
	}
	if s.MaxSampleLength > 0 {
		sflowConfig.MaxSampleLength = s.MaxSampleLength
	}
	if s.MaxFlowHeaderLength > 0 {
		sflowConfig.MaxFlowHeaderLength = s.MaxFlowHeaderLength
	}
	if s.MaxCounterHeaderLength > 0 {
		sflowConfig.MaxCounterHeaderLength = s.MaxCounterHeaderLength
	}
	return sflowConfig
}

// Start starts this sFlow listener listening on the configured network for sFlow packets
func (s *SFlow) Start(acc telegraf.Accumulator) error {
	u, err := url.Parse(s.ServiceAddress)
	if err != nil {
		return err
	}

	conn, err := listenUDP(u.Scheme, u.Host)
	if err != nil {
		return err
	}
	s.closer = conn
	s.addr = conn.LocalAddr()

	if s.ReadBufferSize.Size > 0 {
		conn.SetReadBuffer(int(s.ReadBufferSize.Size))
	}

	s.Log.Infof("Listening on %s://%s", s.addr.Network(), s.addr.String())

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.read(acc, conn)
	}()

	return nil
}

// Gather is a NOOP for sFlow as it receives, asynchronously, sFlow network packets
func (s *SFlow) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (s *SFlow) Stop() {
	if s.closer != nil {
		s.closer.Close()
	}
	s.wg.Wait()
}

func (s *SFlow) Address() net.Addr {
	return s.addr
}

func (s *SFlow) read(acc telegraf.Accumulator, conn net.PacketConn) {
	buf := make([]byte, maxPacketSize)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				acc.AddError(err)
			}
			break
		}
		s.process(acc, buf[:n])
	}
}

func (s *SFlow) process(acc telegraf.Accumulator, buf []byte) {
	if err := s.decoder.Decode(s.decoderOpts, bytes.NewBuffer(buf)); err != nil {
		acc.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))
	}

	for _, m := range s.decoder.GetMetrics() {
		acc.AddMetric(m)
	}
}

func listenUDP(network string, address string) (*net.UDPConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return net.ListenUDP(network, addr)
	default:
		return nil, fmt.Errorf("unsupported network type: %s", network)
	}
}

// init registers this SFlow input plug in with the Telegraf framework
func init() {
	inputs.Add("sflow", func() telegraf.Input {
		return &SFlow{}
	})
}
