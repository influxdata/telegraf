package sflow

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"text/scanner"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/sflow"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type packetSFlowListener struct {
	net.PacketConn
	*SFlowListener
	resolver
}

func (psl *packetSFlowListener) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := psl.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				psl.AddError(err)
			}
			break
		}
		psl.process(buf[:n])
	}
}

func (psl *packetSFlowListener) process(buf []byte) {
	metrics, err := psl.Parse(buf)
	if err != nil {
		psl.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))

	}
	for _, m := range metrics {
		psl.resolver.resolve(m, func(resolvedM telegraf.Metric) {
			psl.AddMetric(resolvedM)
		})
	}
}

// SFlowListener configuration structure
type SFlowListener struct {
	ServiceAddress    string        `toml:"service_address"`
	ReadBufferSize    internal.Size `toml:"read_buffer_size"`
	SNMPCommunity     string        `toml:"snmp_community"`
	SNMPIfaceResolve  bool          `toml:"snmp_iface_resolve"`
	SNMPIfaceCacheTTL int           `toml:"snmp_iface_cache_ttl"`
	DNSFQDNResolve    bool          `toml:"dns_fqdn_resolve"`
	DNSFQDNCacheTTL   int           `toml:"dns_fqdn_cache_ttl"`

	MaxFlowsPerSample     uint32 `toml:"max_flows_per_sample"`
	MaxCountersPerSample  uint32 `toml:"max_counters_per_sample"`
	MaxSamplesPerPacket   uint32 `toml:"max_samples_per_packet"`
	TagsAsFields          string `toml:"as_fields"`
	DNSMultiNameProcessor string `toml:"dns_multi_name_processor"`

	//dnsTTLTicker   *time.Ticker
	//ifaceTTLTicker *time.Ticker
	nameResolver resolver
	//tlsint.ServerConfig
	parsers.Parser
	telegraf.Accumulator
	io.Closer
}

// Description answers a description of this input plugin
func (sl *SFlowListener) Description() string {
	return "SFlow protocol listener"
}

// SampleConfig answers a sample configuration
func (sl *SFlowListener) SampleConfig() string {
	return `
  ## URL to listen on
  # service_address = "udp://:6343"
  # service_address = "udp4://:6343"
  # service_address = "udp6://:6343"
  
  ## Maximum socket buffer size (in bytes when no unit specified).
  ## For stream sockets, once the buffer fills up, the sender will start backing up.
  ## For datagram sockets, once the buffer fills up, metrics will start dropping.
  ## Defaults to the OS default.
  # read_buffer_size = "64KiB"

  ## Whether interface indexes should be turned into interface names via use of sn,p
  # snmp_iface_resolve = false

  ## The SNMP community string to use for access SNMP on the agents in order to resolve interface names
  # snmp_community = "public"

  ## The length of time the interface names are cached
  # snmp_iface_cache_ttl = 3600

  ## Should IP addresses be resolved to host names through DNS lookup
  # dns_fqdn_resolve = false

  ## The length of time the FWDNs are cached
  # dns_fqdn_cache_ttl = 3600

  ##
  # max_flows_per_sample = 10
  # max_counters_per_sample = 10
  # max_samples_per_packet = 10
`
}

// Gather is a NOP for sFlow as it receives, asynchronously, sFlow network packets
func (sl *SFlowListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Start starts this sFlow listener listening on the configured network for sFlow packets
func (sl *SFlowListener) Start(acc telegraf.Accumulator) error {
	sl.Accumulator = acc
	sl.nameResolver = newAsyncResolver(sl.DNSFQDNResolve, sl.DNSMultiNameProcessor, sl.SNMPIfaceResolve, sl.SNMPCommunity)
	sl.nameResolver.start(time.Duration(sl.DNSFQDNCacheTTL)*time.Second, time.Duration(sl.SNMPIfaceCacheTTL)*time.Second)

	tagsAsFields := mapTagsAsFieldsCommaString(sl.TagsAsFields)
	sflowConfig := sflow.NewDefaultV5FormatOptions()
	sflowConfig.MaxFlowsPerSample = sl.MaxFlowsPerSample
	sflowConfig.MaxCountersPerSample = sl.MaxCountersPerSample
	sflowConfig.MaxSamplesPerPacket = sl.MaxSamplesPerPacket

	parser, err := sflow.NewParser("sflow", make(map[string]string), sflowConfig, tagsAsFields)
	if err != nil {
		return err
	}
	sl.Parser = parser

	spl := strings.SplitN(sl.ServiceAddress, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid service address: %s", sl.ServiceAddress)
	}

	protocol := spl[0]
	addr := spl[1]

	switch protocol {
	case "udp", "udp4", "udp6":
		pc, err := udpListen(protocol, addr)
		if err != nil {
			return err
		}
		if sl.ReadBufferSize.Size > 0 {
			if srb, ok := pc.(setReadBufferer); ok {
				srb.SetReadBuffer(int(sl.ReadBufferSize.Size))
			} else {
				log.Printf("W! Unable to set read buffer on a %s socket", protocol)
			}
		}

		log.Printf("I! [inputs.sflow] Listening on %s://%s", protocol, pc.LocalAddr())

		psl := &packetSFlowListener{
			PacketConn:    pc,
			SFlowListener: sl,
			resolver:      sl.nameResolver,
		}

		sl.Closer = psl
		go psl.listen()
	default:
		return fmt.Errorf("unsupported protocol '%s' in '%s'", protocol, sl.ServiceAddress)
	}

	return nil
}

func mapTagsAsFieldsCommaString(input string) map[string]bool {
	var s scanner.Scanner
	s.Init(strings.NewReader(input))
	s.Whitespace |= 1 << ','
	asFields := make(map[string]bool)
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		asFields[s.TokenText()] = true
	}
	return asFields
}

func udpListen(network string, address string) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return net.ListenUDP(network, addr)
	default:
		return nil, fmt.Errorf("unsupported network type %s", network)
	}
}

// Stop thie SFlowListener
func (sl *SFlowListener) Stop() {
	if sl.Closer != nil {
		sl.Close()
		sl.Closer = nil
	}
	sl.nameResolver.stop()
}

func newSFlowListener() *SFlowListener {
	sflowConfig := sflow.NewDefaultV5FormatOptions()
	parser, _ := sflow.NewParser("sflow", make(map[string]string), sflowConfig, nil) // TODO
	return &SFlowListener{
		Parser: parser,
	}
}

func init() {
	inputs.Add("sflow", func() telegraf.Input { return newSFlowListener() })
}
